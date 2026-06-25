// Package modbusrtu is a Modbus RTU (serial / RS-485) Source for the gateway.
// It is the serial sibling of package modbus (Modbus/TCP): Modbus RTU carries
// the same application PDU as Modbus/TCP, so this package reuses modbus.Point,
// modbus.ResolvePoint, modbus.Read and modbus.DecodeSample for point resolution,
// the read switch and register decoding. Only the transport differs.
//
// The primary deployment target is the Advantech ICR-3232, whose RS-485 port
// enumerates as /dev/ttyS1 (dmesg: "RS485 expansion board detected on ttyS1").
// On that UART the kernel drives the DE/RE direction line in hardware, so the
// port is opened like any serial port; for USB adapters that need software RTS
// toggling, SerialDevice.RS485 applies the Linux TIOCSRS485 ioctl.
//
// RS-485 is a half-duplex multidrop bus: several slaves share one pair of wires
// and only one transaction may be in flight at a time. The Poller therefore
// groups devices by serial port into a "bus" and runs ONE goroutine per bus that
// serializes every transaction across the slaves on that port. Each slave keeps
// its own scan cadence, timeout and retry budget.
package modbusrtu

import (
	"context"
	"fmt"
	"sync"
	"time"

	mb "github.com/grid-x/modbus"
	gxserial "github.com/grid-x/serial"

	"goMqttDnp3/config"
	"goMqttDnp3/modbus"
	"goMqttDnp3/source"
)

// rtuDevice is one Modbus RTU slave (a UnitID) on a serial bus.
type rtuDevice struct {
	cfg    config.SerialDevice
	points []modbus.Point

	// nextPoll is the earliest time this slave is due again. Owned exclusively by
	// the bus goroutine, so it needs no lock.
	nextPoll time.Time

	mu     sync.RWMutex
	status source.Status
}

// bus is one serial port carrying Modbus RTU, shared by one or more slaves
// (multidrop). All access to handler/client/open happens on the single bus
// goroutine, so they need no lock.
type bus struct {
	port    string
	line    config.SerialDevice // line settings (baud/parity/…/RS485) — from the first device on this port
	devices []*rtuDevice

	handler *mb.RTUClientHandler
	client  mb.Client
	open    bool // port currently open

	// busMu serializes transactions on the half-duplex bus: the sweep's reads and
	// control writes from other goroutines — only one transaction may be in flight.
	busMu sync.Mutex
}

// Poller is the Modbus RTU source.Source. Construct with New, register slaves
// with AddDevice, then Start/Stop.
type Poller struct {
	h source.Handler

	mu    sync.Mutex
	buses map[string]*bus // keyed by serial port

	cancel  context.CancelFunc
	wg      sync.WaitGroup
	started bool
}

var _ source.Source = (*Poller)(nil)

// New creates a Poller that delivers samples to h.
func New(h source.Handler) *Poller {
	return &Poller{h: h, buses: make(map[string]*bus)}
}

// AddDevice registers a Modbus RTU slave and the (Modbus RTU) mappings to poll on
// it. Devices sharing a Port join the same half-duplex bus; the bus-level line
// settings (baud/data/parity/stop/RS485) are taken from the first device added
// for that port. Mappings that aren't Modbus RTU or don't target this device are
// ignored. Must be called before Start.
func (p *Poller) AddDevice(d config.SerialDevice, mappings []config.SignalMapping) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.started {
		return fmt.Errorf("modbusrtu: cannot AddDevice after Start")
	}
	if d.Port == "" {
		return fmt.Errorf("modbusrtu: device %q has no serial port", d.ID)
	}
	for _, b := range p.buses {
		for _, ex := range b.devices {
			if ex.cfg.ID == d.ID {
				return fmt.Errorf("modbusrtu: device %q already added", d.ID)
			}
		}
	}

	dev := &rtuDevice{
		cfg:    d,
		status: source.Status{ID: d.ID, Label: d.Label, Addr: d.Addr()},
	}
	for _, m := range mappings {
		if !m.IsModbusRTU() || m.Src() != d.ID || !m.Enabled {
			continue
		}
		pt, err := modbus.ResolvePoint(m)
		if err != nil {
			return fmt.Errorf("modbusrtu device %q: %w", d.ID, err)
		}
		dev.points = append(dev.points, pt)
	}

	b := p.buses[d.Port]
	if b == nil {
		b = &bus{port: d.Port, line: d}
		p.buses[d.Port] = b
	} else if !sameLine(b.line, d) {
		p.logWarn(fmt.Sprintf("device %q shares port %s but with different line settings; using %s",
			d.ID, d.Port, b.line.Addr()))
	}
	b.devices = append(b.devices, dev)
	return nil
}

func (p *Poller) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.started {
		p.mu.Unlock()
		return fmt.Errorf("modbusrtu: already started")
	}
	p.started = true
	cctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	list := make([]*bus, 0, len(p.buses))
	for _, b := range p.buses {
		list = append(list, b)
	}
	p.mu.Unlock()

	for _, b := range list {
		b.handler = newHandler(b.line)
		if b.line.LogFrames {
			b.handler.Logger = modbus.FrameLogger{H: p.h, Prefix: b.line.ID}
		}
		b.client = mb.NewClient(b.handler)
		p.wg.Add(1)
		go p.pollBus(cctx, b)
	}
	return nil
}

func (p *Poller) Stop() {
	p.mu.Lock()
	if !p.started {
		p.mu.Unlock()
		return
	}
	p.started = false
	cancel := p.cancel
	list := make([]*bus, 0, len(p.buses))
	for _, b := range p.buses {
		list = append(list, b)
	}
	p.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	p.wg.Wait()
	for _, b := range list {
		if b.handler != nil {
			_ = b.handler.Close()
		}
	}
}

func (p *Poller) Status() []source.Status {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]source.Status, 0, len(p.buses))
	for _, b := range p.buses {
		for _, d := range b.devices {
			d.mu.RLock()
			out = append(out, d.status)
			d.mu.RUnlock()
		}
	}
	return out
}

// newHandler builds a grid-x RTU handler from a bus's line settings, applying
// sane Modbus defaults (9600 8N1) and the optional RS-485 ioctl.
func newHandler(line config.SerialDevice) *mb.RTUClientHandler {
	h := mb.NewRTUClientHandler(line.Port)
	h.BaudRate = intOr(line.BaudRate, 9600)
	h.DataBits = intOr(line.DataBits, 8)
	h.StopBits = intOr(line.StopBits, 1)
	h.Parity = parityOr(line.Parity, "N")
	h.Timeout = msOr(line.TimeoutMs, 1000)
	h.IdleTimeout = 60 * time.Second
	if line.RS485.Enabled {
		h.RS485 = gxserial.RS485Config{
			Enabled:            true,
			RtsHighDuringSend:  line.RS485.RtsHighDuringSend,
			RtsHighAfterSend:   line.RS485.RtsHighAfterSend,
			RxDuringTx:         line.RS485.RxDuringTx,
			DelayRtsBeforeSend: usDur(line.RS485.DelayRtsBeforeSendUs),
			DelayRtsAfterSend:  usDur(line.RS485.DelayRtsAfterSendUs),
		}
	}
	return h
}

// pollBus polls every slave on one serial bus until ctx is cancelled. Slaves are
// polled in turn (the bus is half-duplex), each on its own scan cadence. The
// first sweep runs immediately.
func (p *Poller) pollBus(ctx context.Context, b *bus) {
	defer p.wg.Done()
	ticker := time.NewTicker(b.baseInterval())
	defer ticker.Stop()
	for {
		p.sweep(ctx, b)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// sweep opens the port if needed, then polls every slave that is due. If every
// slave tried in this sweep fails, the port is dropped so the next sweep reopens
// it — recovering from a yanked/re-enumerated USB adapter.
func (p *Poller) sweep(ctx context.Context, b *bus) {
	b.busMu.Lock()
	defer b.busMu.Unlock()
	if !p.ensureOpen(ctx, b) {
		return
	}
	now := time.Now()
	attempted, ok := 0, 0
	for _, d := range b.devices {
		if ctx.Err() != nil {
			return
		}
		if now.Before(d.nextPoll) {
			continue
		}
		d.nextPoll = now.Add(scanInterval(d.cfg))
		attempted++
		if p.pollDevice(ctx, b, d) {
			ok++
		}
	}
	// If every slave failed the port may be gone; drop it so the next sweep
	// reopens it. Skip on shutdown (ctx cancelled) — the failures are just the
	// in-flight reads unwinding, not a dead bus.
	if attempted > 0 && ok == 0 && ctx.Err() == nil {
		p.closeBus(b, "all slaves failed; reopening port")
	}
}

// ensureOpen makes sure the bus's serial port is open. On failure every slave on
// the bus is marked comm-lost and the next sweep retries.
func (p *Poller) ensureOpen(ctx context.Context, b *bus) bool {
	if b.open {
		return true
	}
	if err := b.handler.Connect(ctx); err != nil {
		for _, d := range b.devices {
			p.setConn(d, false, "open "+b.port+": "+err.Error())
		}
		return false
	}
	b.open = true
	return true
}

func (p *Poller) closeBus(b *bus, reason string) {
	if b.handler != nil {
		_ = b.handler.Close()
	}
	b.open = false
	if reason != "" {
		p.logWarn(fmt.Sprintf("bus %s: %s", b.port, reason))
	}
}

// pollDevice reads every point of one slave and emits samples. Reports whether
// the slave responded. A read failure marks only this slave offline and aborts
// its remaining points — the bus stays open so the other slaves still get polled.
func (p *Poller) pollDevice(ctx context.Context, b *bus, d *rtuDevice) bool {
	b.handler.SetSlave(d.cfg.UnitID)
	// Buffer the scan's samples and emit them as a single end-of-scan burst.
	// Reads are sequential and, on a slow RS-485 bus, a whole scan can span far
	// more than the publisher's coalescing window; emitting inline would scatter
	// the points across windows and defeat batching. A tight burst once the scan
	// finishes lands every point in one window → one message per device per scan.
	batch := make([]source.Sample, 0, len(d.points))
	for _, pt := range d.points {
		if ctx.Err() != nil {
			return false // shutting down: drop the partial scan
		}
		raw, err := p.read(ctx, b, d, pt)
		if err != nil {
			p.emitBurst(batch) // publish whatever was read before the failure
			p.setConn(d, false, "read: "+err.Error())
			return false
		}
		s, ok := modbus.DecodeSample(d.cfg.ID, pt, raw)
		if !ok {
			continue
		}
		d.mu.Lock()
		d.status.MeasurementsRx++
		d.status.LastReadAt = time.Now()
		d.mu.Unlock()
		batch = append(batch, s)
	}
	p.emitBurst(batch)
	p.setConn(d, true, "")
	return true
}

// emitBurst hands a scan's buffered samples to the handler back-to-back, so the
// publisher's coalescing window groups them into one message.
func (p *Poller) emitBurst(batch []source.Sample) {
	if p.h == nil {
		return
	}
	for i := range batch {
		p.h.OnSample(batch[i])
	}
}

func (p *Poller) read(ctx context.Context, b *bus, d *rtuDevice, pt modbus.Point) ([]byte, error) {
	retries := d.cfg.Retries
	if retries < 0 {
		retries = 0
	}
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		raw, err := modbus.Read(ctx, b.client, pt)
		if err == nil {
			return raw, nil
		}
		lastErr = err
		if attempt < retries {
			select {
			case <-time.After(msOr(d.cfg.RetryDelayMs, 200)):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	return nil, lastErr
}

// deviceBus finds the bus and slave for a device ID.
func (p *Poller) deviceBus(id string) (*bus, *rtuDevice, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, b := range p.buses {
		for _, d := range b.devices {
			if d.cfg.ID == id {
				return b, d, nil
			}
		}
	}
	return nil, nil, fmt.Errorf("modbusrtu: unknown device %q", id)
}

// WriteCoil writes a single coil to a serial slave (binary control), serialized
// on its half-duplex bus. on → 0xFF00, off → 0x0000.
func (p *Poller) WriteCoil(deviceID string, addr uint16, on bool) error {
	b, d, err := p.deviceBus(deviceID)
	if err != nil {
		return err
	}
	b.busMu.Lock()
	defer b.busMu.Unlock()
	if !b.open || b.client == nil {
		return fmt.Errorf("modbusrtu device %q: bus not open", deviceID)
	}
	b.handler.SetSlave(d.cfg.UnitID)
	var v uint16
	if on {
		v = 0xFF00
	}
	_, err = b.client.WriteSingleCoil(context.Background(), addr, v)
	return err
}

// WriteRegister writes a single holding register to a serial slave (analog
// control), serialized on its half-duplex bus.
func (p *Poller) WriteRegister(deviceID string, addr uint16, value uint16) error {
	b, d, err := p.deviceBus(deviceID)
	if err != nil {
		return err
	}
	b.busMu.Lock()
	defer b.busMu.Unlock()
	if !b.open || b.client == nil {
		return fmt.Errorf("modbusrtu device %q: bus not open", deviceID)
	}
	b.handler.SetSlave(d.cfg.UnitID)
	_, err = b.client.WriteSingleRegister(context.Background(), addr, value)
	return err
}

func (p *Poller) setConn(d *rtuDevice, connected bool, errMsg string) {
	d.mu.Lock()
	changed := d.status.Connected != connected
	d.status.Connected = connected
	d.status.LastError = errMsg
	snap := d.status
	d.mu.Unlock()
	if changed && p.h != nil {
		p.h.OnStatusChange(snap)
	}
}

func (p *Poller) logWarn(msg string) {
	if p.h != nil {
		p.h.OnLog("warn", "modbusrtu: "+msg)
	}
}

// baseInterval is the bus ticker period: the smallest slave scan rate on the bus
// (so the fastest slave is honored), floored to avoid a hot spin.
func (b *bus) baseInterval() time.Duration {
	base := scanInterval(b.devices[0].cfg)
	for _, d := range b.devices[1:] {
		base = min(base, scanInterval(d.cfg))
	}
	return max(base, 20*time.Millisecond)
}

func scanInterval(d config.SerialDevice) time.Duration { return msOr(d.ScanRateMs, 1000) }

// sameLine reports whether two devices request the same serial line settings.
// Only the line-level fields matter (per-device cadence/timeout/retry are
// independent), so a mismatch is just a warning, not an error.
func sameLine(a, b config.SerialDevice) bool {
	return a.BaudRate == b.BaudRate &&
		a.DataBits == b.DataBits &&
		a.StopBits == b.StopBits &&
		a.Parity == b.Parity &&
		a.RS485 == b.RS485
}

func intOr(v, def int) int {
	if v <= 0 {
		return def
	}
	return v
}

func parityOr(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func msOr(ms, def int) time.Duration {
	if ms <= 0 {
		ms = def
	}
	return time.Duration(ms) * time.Millisecond
}

func usDur(us int) time.Duration {
	if us <= 0 {
		return 0
	}
	return time.Duration(us) * time.Microsecond
}
