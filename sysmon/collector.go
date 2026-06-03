// Package sysmon collects host telemetry (CPU, memory, disk, network,
// temperature, uptime) and renders it as Sparkplug B node metrics.
//
// It is deliberately decoupled from the field-protocol sources: unlike Modbus or
// DNP3 points (which the user maps explicitly), system metrics are zero-config —
// the publisher runs a Collector on a ticker and publishes the result as
// node-level metrics under a "System/" folder. gopsutil on Linux is pure-Go
// (reads /proc and /sys), so this still cross-compiles for the ICR-3232
// (linux/arm/v7) with CGO disabled.
package sysmon

import (
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	"goMqttDnp3/config"
	"goMqttDnp3/sparkplug"
)

// Collector samples the host on demand. It is safe for concurrent use; Collect
// keeps the previous network counters to derive transfer rates between calls.
type Collector struct {
	prefix     string
	mounts     []string
	interfaces map[string]bool // explicit allow-list; empty = all non-loopback
	tempKey    string          // optional sensor-key substring to prefer
	sel        config.SystemMetrics // which metric groups to publish
	disabled   map[string]bool      // canonical suffixes to exclude (individual opt-out)

	mu       sync.Mutex
	prevNet  map[string]net.IOCountersStat
	prevTime time.Time

	warnOnce sync.Map // group name → struct{}, so a missing subsystem logs once
}

// New builds a Collector from the system config, applying defaults.
func New(cfg config.SystemConfig) *Collector {
	prefix := cfg.MetricPrefix
	if prefix == "" {
		prefix = "System/"
	}
	mounts := cfg.Mounts
	if len(mounts) == 0 {
		mounts = []string{"/"}
	}
	var ifaces map[string]bool
	if len(cfg.Interfaces) > 0 {
		ifaces = make(map[string]bool, len(cfg.Interfaces))
		for _, n := range cfg.Interfaces {
			ifaces[n] = true
		}
	}
	disabled := make(map[string]bool, len(cfg.DisabledMetrics))
	for _, d := range cfg.DisabledMetrics {
		disabled[d] = true
	}
	return &Collector{
		prefix:     prefix,
		mounts:     mounts,
		interfaces: ifaces,
		tempKey:    cfg.TempSensorKey,
		sel:        cfg.Metrics.Effective(),
		disabled:   disabled,
		prevNet:    make(map[string]net.IOCountersStat),
	}
}

// Collect samples every subsystem once and returns the metrics at timestamp ts.
// A failing subsystem is skipped (logged once) rather than aborting the batch,
// so a kernel without e.g. thermal sensors still yields CPU/mem/disk/net.
func (c *Collector) Collect(ts uint64) []*sparkplug.Metric {
	c.mu.Lock()
	defer c.mu.Unlock()

	var ms []*sparkplug.Metric
	add := func(name string, v float64) {
		// Honor individual opt-outs: a metric is skipped if its name, or its
		// canonical (placeholder) form for per-mount/iface metrics, is disabled.
		if c.disabled[name] || c.disabled[canonicalSuffix(name)] {
			return
		}
		ms = append(ms, sparkplug.MetricDouble(c.prefix+name, ts, v))
	}
	addStr := func(name, v string) {
		if v == "" || c.disabled[name] {
			return
		}
		ms = append(ms, sparkplug.MetricString(c.prefix+name, ts, v))
	}

	c.collectCPU(add)
	c.collectMemory(add)
	c.collectDisk(add)
	c.collectNetwork(add)
	c.collectTemperature(add)
	c.collectHost(add)
	c.collectVendor(add, addStr) // board sensors + identity (ICR build only)

	return ms
}

func (c *Collector) collectCPU(add func(string, float64)) {
	if c.sel.CPU {
		// interval 0 → percentage since the previous call (i.e. over our tick).
		if pct, err := cpu.Percent(0, false); err == nil && len(pct) > 0 {
			add("CPU/Usage_pct", round2(pct[0]))
		} else if err != nil {
			c.warn("cpu", err)
		}
	}
	if c.sel.Load {
		if la, err := load.Avg(); err == nil && la != nil {
			add("CPU/Load1", round2(la.Load1))
			add("CPU/Load5", round2(la.Load5))
			add("CPU/Load15", round2(la.Load15))
		} else if err != nil {
			c.warn("load", err)
		}
	}
}

func (c *Collector) collectMemory(add func(string, float64)) {
	if c.sel.Memory {
		if vm, err := mem.VirtualMemory(); err == nil && vm != nil {
			add("Memory/Used_pct", round2(vm.UsedPercent))
			add("Memory/Used_MB", bytesToMB(vm.Used))
			add("Memory/Available_MB", bytesToMB(vm.Available))
			add("Memory/Total_MB", bytesToMB(vm.Total))
		} else if err != nil {
			c.warn("memory", err)
		}
	}
	if c.sel.Swap {
		if sw, err := mem.SwapMemory(); err == nil && sw != nil && sw.Total > 0 {
			add("Memory/Swap_Used_pct", round2(sw.UsedPercent))
		}
	}
}

func (c *Collector) collectDisk(add func(string, float64)) {
	if !c.sel.Disk {
		return
	}
	for _, mount := range c.mounts {
		u, err := disk.Usage(mount)
		if err != nil || u == nil {
			c.warn("disk:"+mount, err)
			continue
		}
		key := "Disk/" + sanitizeMount(mount) + "/"
		add(key+"Used_pct", round2(u.UsedPercent))
		add(key+"Free_MB", bytesToMB(u.Free))
	}
}

func (c *Collector) collectNetwork(add func(string, float64)) {
	if !c.sel.Network && !c.sel.NetworkRates {
		return
	}
	counters, err := net.IOCounters(true)
	if err != nil {
		c.warn("network", err)
		return
	}
	now := time.Now()
	elapsed := now.Sub(c.prevTime).Seconds()
	for _, io := range counters {
		if !c.includeIface(io.Name) {
			continue
		}
		key := "Network/" + sanitizeMount(io.Name) + "/"
		if c.sel.Network {
			add(key+"Rx_MB", bytesToMB(io.BytesRecv))
			add(key+"Tx_MB", bytesToMB(io.BytesSent))
		}
		if c.sel.NetworkRates {
			if prev, ok := c.prevNet[io.Name]; ok && elapsed > 0 {
				add(key+"RxRate_kbps", rateKbps(io.BytesRecv, prev.BytesRecv, elapsed))
				add(key+"TxRate_kbps", rateKbps(io.BytesSent, prev.BytesSent, elapsed))
			}
		}
		c.prevNet[io.Name] = io
	}
	c.prevTime = now
}

func (c *Collector) collectTemperature(add func(string, float64)) {
	if !c.sel.Temperature {
		return
	}
	temps, err := host.SensorsTemperatures()
	if err != nil || len(temps) == 0 {
		// Many embedded kernels expose no usable sensor; don't spam — warn once.
		c.warn("temperature", err)
		return
	}
	if t, ok := c.pickTemp(temps); ok {
		add("Temperature/CPU_C", round2(t))
	}
}

func (c *Collector) collectHost(add func(string, float64)) {
	if c.sel.Uptime {
		if up, err := host.Uptime(); err == nil {
			add("Uptime_h", round2(float64(up)/3600.0))
		} else {
			c.warn("uptime", err)
		}
	}
	if c.sel.Processes {
		if info, err := host.Info(); err == nil && info != nil {
			add("Process/Count", float64(info.Procs))
		}
	}
}

// pickTemp chooses the most representative CPU/board temperature: the configured
// key substring if set, otherwise the first cpu/thermal/soc sensor, otherwise the
// first sensor reporting a plausible (>0) value.
func (c *Collector) pickTemp(temps []host.TemperatureStat) (float64, bool) {
	if c.tempKey != "" {
		for _, t := range temps {
			if strings.Contains(strings.ToLower(t.SensorKey), strings.ToLower(c.tempKey)) && t.Temperature > 0 {
				return t.Temperature, true
			}
		}
	}
	for _, pref := range []string{"cpu", "thermal", "soc", "coretemp"} {
		for _, t := range temps {
			if strings.Contains(strings.ToLower(t.SensorKey), pref) && t.Temperature > 0 {
				return t.Temperature, true
			}
		}
	}
	for _, t := range temps {
		if t.Temperature > 0 {
			return t.Temperature, true
		}
	}
	return 0, false
}

func (c *Collector) includeIface(name string) bool {
	if c.interfaces != nil {
		return c.interfaces[name]
	}
	return name != "lo" && !strings.HasPrefix(name, "lo:")
}

func (c *Collector) warn(group string, err error) {
	if _, seen := c.warnOnce.LoadOrStore(group, struct{}{}); seen {
		return
	}
	if err != nil {
		slog.Warn("sysmon: subsystem unavailable", "group", group, "err", err)
	} else {
		slog.Warn("sysmon: subsystem returned no data", "group", group)
	}
}

// --- helpers ---

// canonicalSuffix maps a concrete metric suffix to its placeholder form so a
// single config entry can disable a metric across every mount/interface:
//
//	Disk/root/Free_MB        → Disk/<mount>/Free_MB
//	Network/eth0/RxRate_kbps → Network/<iface>/RxRate_kbps
//
// Non-templated suffixes are returned unchanged.
func canonicalSuffix(name string) string {
	if rest, ok := strings.CutPrefix(name, "Disk/"); ok {
		if i := strings.IndexByte(rest, '/'); i >= 0 {
			return "Disk/<mount>/" + rest[i+1:]
		}
	}
	if rest, ok := strings.CutPrefix(name, "Network/"); ok {
		if i := strings.IndexByte(rest, '/'); i >= 0 {
			return "Network/<iface>/" + rest[i+1:]
		}
	}
	return name
}

func sanitizeMount(s string) string {
	s = strings.Trim(s, "/")
	if s == "" {
		return "root"
	}
	s = strings.ReplaceAll(s, "/", "_")
	return strings.ReplaceAll(s, " ", "_")
}

func rateKbps(cur, prev uint64, elapsed float64) float64 {
	if cur < prev { // counter wrapped/reset
		return 0
	}
	return round2(float64(cur-prev) * 8 / 1000 / elapsed)
}

func bytesToMB(b uint64) float64 {
	return round2(float64(b) / (1024 * 1024))
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
