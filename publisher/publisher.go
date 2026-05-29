// Package publisher orchestrates Modbus polling → Sparkplug B publishing.
package publisher

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"sync/atomic"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"

	"goMqttModbus/config"
	"goMqttModbus/mapping"
	"goMqttModbus/modbus"
	"goMqttModbus/sparkplug"
)

// Status snapshots the publisher's runtime state for the API.
type Status struct {
	Running       bool                     `json:"running"`
	MQTTConnected bool                     `json:"mqttConnected"`
	BdSeq         uint64                   `json:"bdSeq"`
	Devices       map[string]DeviceStatus  `json:"devices"`
	PublishCount  int64                    `json:"publishCount"`
	ErrorCount    int64                    `json:"errorCount"`
	Uptime        string                   `json:"uptime"`
	LastReadings  map[string]float64       `json:"lastReadings"`
}

// DeviceStatus reports per-device Modbus health.
type DeviceStatus struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Addr      string `json:"addr"`
	Connected bool   `json:"connected"`
	Errors    int64  `json:"errors"`
	Reads     int64  `json:"reads"`
}

// Publisher manages the full lifecycle: MQTT connect → NBIRTH → poll → NDATA → NDEATH.
type Publisher struct {
	store   config.AppConfig
	pool    *modbus.Pool
	node    *sparkplug.Node
	client  mqtt.Client

	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running atomic.Bool
	startAt time.Time

	// counters
	publishCount atomic.Int64
	errorCount   atomic.Int64

	// per-device error/read counters
	devMu    sync.RWMutex
	devStats map[string]*devCounter

	// last published value per metric (for deadband)
	lastMu   sync.RWMutex
	lastVal  map[string]float64

	// store-and-forward offline buffer
	bufMu  sync.Mutex
	msgBuf []bufferedMsg
	bufMax int

	// live status push (optional; set before Start)
	OnStatus func(Status)
	OnLog    func(level, msg string)
}

type devCounter struct {
	errors atomic.Int64
	reads  atomic.Int64
	connected atomic.Bool
}

type bufferedMsg struct {
	deviceID string // "" = node metric
	metrics  []*sparkplug.Metric
}

// New creates a Publisher from the current AppConfig.
func New(cfg config.AppConfig) *Publisher {
	p := &Publisher{
		store:    cfg,
		pool:     modbus.NewPool(3000, 2),
		lastVal:  make(map[string]float64),
		devStats: make(map[string]*devCounter),
		bufMax:   500,
	}
	for _, dev := range cfg.Devices {
		p.devStats[dev.ID] = &devCounter{}
	}
	return p
}

// Start connects to MQTT, sends NBIRTH, and launches per-scan-rate goroutines.
// Returns an error if the MQTT connection fails.
func (p *Publisher) Start(ctx context.Context) error {
	if p.running.Swap(true) {
		return fmt.Errorf("publisher already running")
	}

	cfg := p.store
	sp := cfg.Sparkplug
	mq := cfg.MQTT

	p.node = sparkplug.NewNode(sp.GroupID, sp.NodeID)
	opts := p.node.NewClientOptions(mq.Broker, mq.ClientID, mq.Username, mq.Password)
	opts.SetAutoReconnect(false)
	opts.SetOnConnectHandler(func(c mqtt.Client) {
		p.logInfo("MQTT connected to " + mq.Broker)
		p.drainBuffer()
	})
	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		p.logWarn("MQTT connection lost: " + err.Error())
	})

	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()
	if err := token.Error(); err != nil {
		p.running.Store(false)
		return fmt.Errorf("MQTT connect %s: %w", mq.Broker, err)
	}
	p.client = client

	// Build NBIRTH metrics from enabled mappings (type stubs only).
	birthMetrics := p.buildBirthMetrics()
	if err := p.node.PublishNBirth(client, birthMetrics); err != nil {
		p.logWarn("NBIRTH failed: " + err.Error())
	}

	// Subscribe NCMD for rebirth.
	p.node.SubscribeNCMD(client, func() {
		p.logInfo("Rebirth requested, re-publishing NBIRTH")
		p.node.PublishNBirth(client, p.buildBirthMetrics())
		p.publishDBirths()
	})

	// Publish DBIRTHs for devices that appear in mappings.
	p.publishDBirths()

	pCtx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	p.startAt = time.Now()

	// Group mappings by scanRate for goroutine efficiency.
	groups := groupByScanRate(cfg.Mappings)
	for scanRate, sigs := range groups {
		rate := scanRate
		signals := sigs
		p.wg.Add(1)
		go p.scanGroup(pCtx, rate, signals)
	}

	p.logInfo(fmt.Sprintf("Publisher started: %d scan groups, %d signals", len(groups), len(cfg.Mappings)))
	return nil
}

// Stop sends NDEATH and gracefully shuts down all goroutines.
func (p *Publisher) Stop() {
	if !p.running.Load() {
		return
	}
	if p.cancel != nil {
		p.cancel()
	}
	p.wg.Wait()
	if p.client != nil && p.client.IsConnected() {
		if err := p.node.PublishNDeath(p.client); err != nil {
			p.logWarn("NDEATH failed: " + err.Error())
		}
		p.client.Disconnect(500)
	}
	p.pool.CloseAll()
	p.running.Store(false)
	p.logInfo("Publisher stopped")
}

// Status returns a snapshot of the publisher's current state.
func (p *Publisher) Status() Status {
	s := Status{
		Running:      p.running.Load(),
		PublishCount: p.publishCount.Load(),
		ErrorCount:   p.errorCount.Load(),
		Devices:      make(map[string]DeviceStatus),
		LastReadings: make(map[string]float64),
	}
	if p.client != nil {
		s.MQTTConnected = p.client.IsConnected()
	}
	if p.node != nil {
		s.BdSeq = p.node.Bdseq()
	}
	if !p.startAt.IsZero() {
		s.Uptime = time.Since(p.startAt).Round(time.Second).String()
	}

	p.devMu.RLock()
	for id, dc := range p.devStats {
		label := id
		addr := ""
		for _, dev := range p.store.Devices {
			if dev.ID == id {
				label = dev.Label
				addr = dev.Addr()
				break
			}
		}
		s.Devices[id] = DeviceStatus{
			ID: id, Label: label, Addr: addr,
			Connected: dc.connected.Load(),
			Errors:    dc.errors.Load(),
			Reads:     dc.reads.Load(),
		}
	}
	p.devMu.RUnlock()

	p.lastMu.RLock()
	for k, v := range p.lastVal {
		s.LastReadings[k] = v
	}
	p.lastMu.RUnlock()

	return s
}

// scanGroup polls all signals in the group at the given scan rate.
func (p *Publisher) scanGroup(ctx context.Context, scanRateMs int, sigs []config.SignalMapping) {
	defer p.wg.Done()
	ticker := time.NewTicker(time.Duration(scanRateMs) * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.pollSignals(ctx, sigs)
		}
	}
}

func (p *Publisher) pollSignals(ctx context.Context, sigs []config.SignalMapping) {
	// Group by device for batching (one pass per device connection lock).
	type devGroup struct {
		dev  config.ModbusDevice
		sigs []config.SignalMapping
	}
	devMap := make(map[string]*devGroup)
	for _, sig := range sigs {
		if !sig.Enabled {
			continue
		}
		dg, ok := devMap[sig.ModbusDeviceID]
		if !ok {
			var dev config.ModbusDevice
			for _, d := range p.store.Devices {
				if d.ID == sig.ModbusDeviceID {
					dev = d
					break
				}
			}
			if dev.ID == "" || !dev.Enabled {
				continue
			}
			dg = &devGroup{dev: dev}
			devMap[sig.ModbusDeviceID] = dg
		}
		dg.sigs = append(dg.sigs, sig)
	}

	for _, dg := range devMap {
		p.pollDevice(ctx, dg.dev, dg.sigs)
	}
}

func (p *Publisher) pollDevice(ctx context.Context, dev config.ModbusDevice, sigs []config.SignalMapping) {
	// Gather results per Sparkplug device ID.
	nodeMetrics := make([]*sparkplug.Metric, 0)
	devMetrics := make(map[string][]*sparkplug.Metric)

	for _, sig := range sigs {
		if ctx.Err() != nil {
			return
		}
		raw, err := p.readSignal(ctx, dev, sig)
		if err != nil {
			p.errorCount.Add(1)
			p.devCounter(dev.ID).errors.Add(1)
			p.devCounter(dev.ID).connected.Store(false)
			slog.Warn("read error", "device", dev.ID, "metric", sig.MetricName, "err", err)
			if sig.QualityPolicy == "bad_on_error" {
				ts := sparkplug.NowMs()
				m := sparkplug.MetricDouble(sig.MetricName, ts, 0)
				m.IsNull = true
				p.append(sig.DeviceID, &nodeMetrics, devMetrics, m)
			}
			// last_known: skip → previous value stays
			continue
		}
		p.devCounter(dev.ID).connected.Store(true)
		p.devCounter(dev.ID).reads.Add(1)

		ts := sparkplug.NowMs()
		result, err := mapping.Apply(sig, raw, ts)
		if err != nil {
			slog.Warn("mapping error", "metric", sig.MetricName, "err", err)
			p.errorCount.Add(1)
			continue
		}

		p.lastMu.Lock()
		last, exists := p.lastVal[sig.MetricName]
		if !exists {
			last = math.NaN()
		}
		publish := mapping.CheckDeadband(result.Value, last, sig.Deadband)
		if publish {
			p.lastVal[sig.MetricName] = result.Value
		}
		p.lastMu.Unlock()

		if publish {
			p.append(sig.DeviceID, &nodeMetrics, devMetrics, result.Metric)
		}
	}

	// Publish batched metrics.
	if len(nodeMetrics) > 0 {
		if err := p.publishOrBuffer("", nodeMetrics); err != nil {
			slog.Warn("NDATA publish failed", "err", err)
		}
	}
	for devID, metrics := range devMetrics {
		if err := p.publishOrBuffer(devID, metrics); err != nil {
			slog.Warn("DDATA publish failed", "deviceId", devID, "err", err)
		}
	}
}

func (p *Publisher) append(deviceID string, nodeMetrics *[]*sparkplug.Metric, devMetrics map[string][]*sparkplug.Metric, m *sparkplug.Metric) {
	if deviceID == "" {
		*nodeMetrics = append(*nodeMetrics, m)
	} else {
		devMetrics[deviceID] = append(devMetrics[deviceID], m)
	}
}

func (p *Publisher) readSignal(ctx context.Context, dev config.ModbusDevice, sig config.SignalMapping) ([]byte, error) {
	var fn modbus.ReadFunc
	switch sig.Function {
	case "coil":
		fn = modbus.ReadCoils
	case "discrete_input":
		fn = modbus.ReadDiscreteInputs
	case "input_register":
		fn = modbus.ReadInputRegisters
	case "holding_register":
		fn = modbus.ReadHoldingRegisters
	default:
		return nil, fmt.Errorf("unknown function %q", sig.Function)
	}

	timeoutMs := dev.TimeoutMs
	if timeoutMs <= 0 {
		timeoutMs = 3000
	}
	retries := dev.Retries
	if retries < 0 {
		retries = 2
	}
	retryDelay := dev.RetryDelayMs
	if retryDelay <= 0 {
		retryDelay = 500
	}

	return p.pool.Read(ctx, dev.Addr(), sig.UnitID, fn, sig.Address, sig.Quantity, timeoutMs, retries, retryDelay)
}

func (p *Publisher) publishOrBuffer(deviceID string, metrics []*sparkplug.Metric) error {
	if p.client == nil || !p.client.IsConnected() {
		p.enqueue(deviceID, metrics)
		return nil
	}
	p.publishCount.Add(1)
	if deviceID == "" {
		return p.node.PublishNData(p.client, metrics)
	}
	return p.node.PublishDData(p.client, deviceID, metrics)
}

func (p *Publisher) enqueue(deviceID string, metrics []*sparkplug.Metric) {
	p.bufMu.Lock()
	defer p.bufMu.Unlock()
	if len(p.msgBuf) >= p.bufMax {
		// Drop oldest (head) to stay within limit.
		p.msgBuf = p.msgBuf[1:]
	}
	// Copy metric slice to avoid mutation.
	cp := make([]*sparkplug.Metric, len(metrics))
	copy(cp, metrics)
	p.msgBuf = append(p.msgBuf, bufferedMsg{deviceID: deviceID, metrics: cp})
}

func (p *Publisher) drainBuffer() {
	p.bufMu.Lock()
	msgs := make([]bufferedMsg, len(p.msgBuf))
	copy(msgs, p.msgBuf)
	p.msgBuf = nil
	p.bufMu.Unlock()

	if len(msgs) == 0 {
		return
	}
	slog.Info("draining offline buffer", "count", len(msgs))
	for i := range msgs {
		_ = p.publishOrBuffer(msgs[i].deviceID, msgs[i].metrics)
	}
}

func (p *Publisher) buildBirthMetrics() []*sparkplug.Metric {
	ts := sparkplug.NowMs()
	var metrics []*sparkplug.Metric
	for _, sig := range p.store.Mappings {
		if !sig.Enabled || sig.DeviceID != "" {
			continue
		}
		metrics = append(metrics, sparkplug.MetricDouble(sig.MetricName, ts, 0))
	}
	return metrics
}

func (p *Publisher) publishDBirths() {
	ts := sparkplug.NowMs()
	// Collect metrics per device ID.
	devMetrics := make(map[string][]*sparkplug.Metric)
	for _, sig := range p.store.Mappings {
		if !sig.Enabled || sig.DeviceID == "" {
			continue
		}
		devMetrics[sig.DeviceID] = append(devMetrics[sig.DeviceID], sparkplug.MetricDouble(sig.MetricName, ts, 0))
	}
	for devID, metrics := range devMetrics {
		if err := p.node.PublishDBirth(p.client, devID, metrics); err != nil {
			p.logWarn(fmt.Sprintf("DBIRTH %s failed: %v", devID, err))
		}
	}
}

func (p *Publisher) devCounter(id string) *devCounter {
	p.devMu.Lock()
	dc := p.devStats[id]
	if dc == nil {
		dc = &devCounter{}
		p.devStats[id] = dc
	}
	p.devMu.Unlock()
	return dc
}

func (p *Publisher) logInfo(msg string) {
	slog.Info(msg)
	if p.OnLog != nil {
		p.OnLog("info", msg)
	}
}

func (p *Publisher) logWarn(msg string) {
	slog.Warn(msg)
	if p.OnLog != nil {
		p.OnLog("warn", msg)
	}
}

// groupByScanRate groups mappings by scan rate, filtering disabled.
func groupByScanRate(sigs []config.SignalMapping) map[int][]config.SignalMapping {
	groups := make(map[int][]config.SignalMapping)
	for _, sig := range sigs {
		if !sig.Enabled {
			continue
		}
		rate := sig.ScanRateMs
		if rate < 100 {
			rate = 1000
		}
		groups[rate] = append(groups[rate], sig)
	}
	return groups
}
