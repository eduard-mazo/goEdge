// Package modbus is a Modbus/TCP Source for the gateway. It polls one or more
// devices on a ticker, decodes raw registers/coils per each mapping's
// data-type/byte-order, and pushes source.Sample values to a source.Handler —
// converting Modbus's pull model into the gateway's push model.
package modbus

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"sync"
	"time"

	mb "github.com/grid-x/modbus"

	"goMqttDnp3/config"
	"goMqttDnp3/source"
)

// point is a resolved read derived from a Modbus SignalMapping.
type point struct {
	ptype     source.PointType // routing point type (binary or analog)
	index     uint16           // routing index == Modbus address
	function  string
	address   uint16
	quantity  uint16
	dataType  string
	byteOrder string
}

// device holds one Modbus/TCP connection and the points to poll on it.
type device struct {
	cfg    config.ModbusDevice
	points []point

	handler *mb.TCPClientHandler
	client  mb.Client
	conn    bool // connection currently established

	mu     sync.RWMutex
	status source.Status
}

// Poller is the Modbus source.Source. Construct with New, register devices with
// AddDevice, then Start/Stop.
type Poller struct {
	h source.Handler

	mu      sync.Mutex
	devices map[string]*device

	cancel  context.CancelFunc
	wg      sync.WaitGroup
	started bool
}

// New creates a Poller that delivers samples to h.
func New(h source.Handler) *Poller {
	return &Poller{h: h, devices: make(map[string]*device)}
}

// AddDevice registers a device and the (Modbus) mappings to poll on it. Mappings
// that aren't Modbus or don't target this device are ignored. Must be called
// before Start.
func (p *Poller) AddDevice(d config.ModbusDevice, mappings []config.SignalMapping) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.started {
		return fmt.Errorf("modbus: cannot AddDevice after Start")
	}
	if _, ok := p.devices[d.ID]; ok {
		return fmt.Errorf("modbus: device %q already added", d.ID)
	}

	dev := &device{
		cfg:    d,
		status: source.Status{ID: d.ID, Label: d.Label, Addr: d.Addr()},
	}
	for _, m := range mappings {
		if !m.IsModbus() || m.Src() != d.ID || !m.Enabled {
			continue
		}
		pt, err := resolvePoint(m)
		if err != nil {
			return fmt.Errorf("modbus device %q: %w", d.ID, err)
		}
		dev.points = append(dev.points, pt)
	}
	p.devices[d.ID] = dev
	return nil
}

func (p *Poller) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.started {
		p.mu.Unlock()
		return fmt.Errorf("modbus: already started")
	}
	p.started = true
	cctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel
	list := make([]*device, 0, len(p.devices))
	for _, d := range p.devices {
		list = append(list, d)
	}
	p.mu.Unlock()

	for _, dev := range list {
		dev.handler = mb.NewTCPClientHandler(dev.cfg.Addr())
		dev.handler.Timeout = msOr(dev.cfg.TimeoutMs, 3000)
		dev.handler.IdleTimeout = 60 * time.Second
		dev.handler.SlaveID = dev.cfg.UnitID
		dev.client = mb.NewClient(dev.handler)
		p.wg.Add(1)
		go p.pollDevice(cctx, dev)
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
	list := make([]*device, 0, len(p.devices))
	for _, d := range p.devices {
		list = append(list, d)
	}
	p.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	p.wg.Wait()
	for _, dev := range list {
		if dev.handler != nil {
			_ = dev.handler.Close()
		}
	}
}

func (p *Poller) Status() []source.Status {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]source.Status, 0, len(p.devices))
	for _, dev := range p.devices {
		dev.mu.RLock()
		out = append(out, dev.status)
		dev.mu.RUnlock()
	}
	return out
}

// pollDevice polls one device until ctx is cancelled. The first scan runs
// immediately; subsequent scans run on the device's scan-rate ticker.
func (p *Poller) pollDevice(ctx context.Context, dev *device) {
	defer p.wg.Done()
	ticker := time.NewTicker(msOr(dev.cfg.ScanRateMs, 1000))
	defer ticker.Stop()
	for {
		p.scan(ctx, dev)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// scan reads every configured point once and emits samples. Connection failures
// are recorded in status and retried on the next tick.
func (p *Poller) scan(ctx context.Context, dev *device) {
	if !dev.conn {
		if err := dev.handler.Connect(ctx); err != nil {
			p.setConn(dev, false, "connect: "+err.Error())
			return
		}
		dev.conn = true
		p.setConn(dev, true, "")
	}

	for _, pt := range dev.points {
		if ctx.Err() != nil {
			return
		}
		raw, err := p.read(ctx, dev, pt)
		if err != nil {
			// Treat a read failure as a dropped connection; reconnect next tick.
			_ = dev.handler.Close()
			dev.conn = false
			p.setConn(dev, false, "read: "+err.Error())
			return
		}
		s, ok := decodeSample(dev.cfg.ID, pt, raw)
		if !ok {
			continue
		}
		dev.mu.Lock()
		dev.status.MeasurementsRx++
		dev.status.LastReadAt = time.Now()
		dev.mu.Unlock()
		if p.h != nil {
			p.h.OnSample(s)
		}
	}
}

func (p *Poller) read(ctx context.Context, dev *device, pt point) ([]byte, error) {
	retries := dev.cfg.Retries
	if retries < 0 {
		retries = 0
	}
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		var (
			data []byte
			err  error
		)
		switch pt.function {
		case "coil":
			data, err = dev.client.ReadCoils(ctx, pt.address, pt.quantity)
		case "discrete_input":
			data, err = dev.client.ReadDiscreteInputs(ctx, pt.address, pt.quantity)
		case "input_register":
			data, err = dev.client.ReadInputRegisters(ctx, pt.address, pt.quantity)
		case "holding_register":
			data, err = dev.client.ReadHoldingRegisters(ctx, pt.address, pt.quantity)
		default:
			return nil, fmt.Errorf("unsupported function %q", pt.function)
		}
		if err == nil {
			return data, nil
		}
		lastErr = err
		if attempt < retries {
			select {
			case <-time.After(msOr(dev.cfg.RetryDelayMs, 500)):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	return nil, lastErr
}

func (p *Poller) setConn(dev *device, connected bool, errMsg string) {
	dev.mu.Lock()
	changed := dev.status.Connected != connected
	dev.status.Connected = connected
	dev.status.LastError = errMsg
	snap := dev.status
	dev.mu.Unlock()
	if changed && p.h != nil {
		p.h.OnStatusChange(snap)
	}
}

// resolvePoint derives the routing point type/index and read parameters from a
// Modbus mapping.
func resolvePoint(m config.SignalMapping) (point, error) {
	pt := point{
		// Routing point type is the function itself so that e.g. holding@0 and
		// input@0 (or coil@0 and discrete@0) don't collide on (pointType,index).
		ptype:     source.PointType(m.Function),
		index:     m.Address,
		function:  m.Function,
		address:   m.Address,
		quantity:  m.Quantity,
		dataType:  m.DataType,
		byteOrder: m.ByteOrder,
	}
	switch m.Function {
	case "coil", "discrete_input":
		if pt.quantity == 0 {
			pt.quantity = 1
		}
	case "input_register", "holding_register":
		if pt.quantity == 0 {
			q, err := registersFor(m.DataType)
			if err != nil {
				return point{}, err
			}
			pt.quantity = q
		}
	default:
		return point{}, fmt.Errorf("unsupported function %q (mapping %q)", m.Function, m.MetricName)
	}
	return pt, nil
}

// registersFor returns the number of 16-bit registers a data type spans.
func registersFor(dataType string) (uint16, error) {
	switch dataType {
	case "int16", "uint16", "bool", "":
		return 1, nil
	case "int32", "uint32", "float32":
		return 2, nil
	case "float64":
		return 4, nil
	default:
		return 0, fmt.Errorf("unknown dataType %q", dataType)
	}
}

// decodeSample turns raw Modbus bytes into a source.Sample for the point.
func decodeSample(deviceID string, pt point, raw []byte) (source.Sample, bool) {
	s := source.Sample{
		SourceID:  deviceID,
		PointType: pt.ptype,
		Index:     pt.index,
		Time:      time.Now(),
		Quality:   source.QualityOnline, // a successful read is, by definition, online
		IsEvent:   false,                // Modbus reads are polls, never events
	}
	switch pt.function {
	case "coil", "discrete_input":
		if len(raw) < 1 {
			return source.Sample{}, false
		}
		s.BoolValue = raw[0]&0x01 != 0
	default: // registers
		v, ok := decodeNumeric(raw, pt.dataType, pt.byteOrder)
		if !ok {
			return source.Sample{}, false
		}
		s.FloatValue = v
	}
	return s, true
}

// decodeNumeric decodes register bytes (Modbus wire order is big-endian per
// register) into a float64, honoring the configured word/byte order.
func decodeNumeric(raw []byte, dataType, byteOrder string) (float64, bool) {
	b := reorder(raw, byteOrder)
	switch dataType {
	case "int16", "":
		if len(b) < 2 {
			return 0, false
		}
		return float64(int16(binary.BigEndian.Uint16(b))), true
	case "uint16":
		if len(b) < 2 {
			return 0, false
		}
		return float64(binary.BigEndian.Uint16(b)), true
	case "int32":
		if len(b) < 4 {
			return 0, false
		}
		return float64(int32(binary.BigEndian.Uint32(b))), true
	case "uint32":
		if len(b) < 4 {
			return 0, false
		}
		return float64(binary.BigEndian.Uint32(b)), true
	case "float32":
		if len(b) < 4 {
			return 0, false
		}
		return float64(math.Float32frombits(binary.BigEndian.Uint32(b))), true
	case "float64":
		if len(b) < 8 {
			return 0, false
		}
		return math.Float64frombits(binary.BigEndian.Uint64(b)), true
	default:
		return 0, false
	}
}

// reorder rearranges raw bytes to canonical big-endian according to byteOrder.
//
//	ABCD = big-endian (no change)        DCBA = full reverse
//	BADC = swap bytes within each word   CDAB = swap 16-bit words
func reorder(raw []byte, order string) []byte {
	n := len(raw)
	out := make([]byte, n)
	switch order {
	case "", "ABCD":
		copy(out, raw)
	case "DCBA":
		for i := 0; i < n; i++ {
			out[i] = raw[n-1-i]
		}
	case "BADC":
		for i := 0; i+1 < n; i += 2 {
			out[i], out[i+1] = raw[i+1], raw[i]
		}
		if n%2 == 1 {
			out[n-1] = raw[n-1]
		}
	case "CDAB":
		for i := 0; i+1 < n; i += 2 {
			j := n - 2 - i
			out[i], out[i+1] = raw[j], raw[j+1]
		}
	default:
		copy(out, raw)
	}
	return out
}

func msOr(ms, def int) time.Duration {
	if ms <= 0 {
		ms = def
	}
	return time.Duration(ms) * time.Millisecond
}
