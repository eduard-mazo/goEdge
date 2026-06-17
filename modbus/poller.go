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
	"strings"
	"sync"
	"time"

	mb "github.com/grid-x/modbus"

	"goMqttDnp3/config"
	"goMqttDnp3/source"
)

// Handler is the subset of the grid-x client handler that the gateway drives,
// satisfied by both *mb.TCPClientHandler (Modbus/TCP) and
// *mb.RTUOverTCPClientHandler (RTU over TCP). It lets one device struct hold
// either transport.
type Handler interface {
	Connect(ctx context.Context) error
	Close() error
	SetSlave(slaveID byte)
}

// FrameLogger adapts a source.Handler into a grid-x modbus.Logger. The library
// logs every raw ADU as "modbus: send % x" / "modbus: recv % x" when a Logger is
// set; FrameLogger forwards those lines to the gateway log (the UI "Registro"),
// tagged with the device ID, so serial/TCP traffic can be inspected live. Shared
// by the Modbus/TCP poller and the serial RTU source; enabled per device via
// the LogFrames config flag.
type FrameLogger struct {
	H      source.Handler
	Prefix string // device tag, e.g. "MOD_SIM"
}

// Printf implements mb.Logger.
func (l FrameLogger) Printf(format string, v ...any) {
	if l.H == nil {
		return
	}
	msg := strings.TrimRight(fmt.Sprintf(format, v...), "\n")
	l.H.OnLog("info", l.Prefix+" "+msg)
}

// Point is a resolved read derived from a Modbus SignalMapping. The Modbus
// framing (function codes, register/coil decoding, word/byte order) is identical
// for Modbus/TCP and Modbus RTU, so Point — together with ResolvePoint and
// DecodeSample — is exported and reused by the serial RTU source (package
// modbusrtu). Only the transport differs between the two.
type Point struct {
	Ptype     source.PointType // routing point type (binary or analog)
	Index     uint16           // routing index == Modbus address
	Function  string
	Address   uint16
	Quantity  uint16
	DataType  string
	ByteOrder string
}

// device holds one Modbus/TCP connection and the points to poll on it.
type device struct {
	cfg    config.ModbusDevice
	points []Point

	handler Handler
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
		pt, err := ResolvePoint(m)
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
		dev.handler, dev.client = p.newHandler(dev.cfg)
		p.wg.Add(1)
		go p.pollDevice(cctx, dev)
	}
	return nil
}

// newHandler builds the grid-x handler and client for a device, choosing the
// transport (Modbus/TCP vs RTU-over-TCP) and wiring frame logging when enabled.
func (p *Poller) newHandler(d config.ModbusDevice) (Handler, mb.Client) {
	timeout := msOr(d.TimeoutMs, 3000)
	var fl mb.Logger
	if d.LogFrames {
		fl = FrameLogger{H: p.h, Prefix: d.ID}
	}
	switch d.Transport {
	case "rtuovertcp":
		h := mb.NewRTUOverTCPClientHandler(d.Addr())
		h.Timeout = timeout
		h.IdleTimeout = 60 * time.Second
		h.Logger = fl
		h.SetSlave(d.UnitID)
		return h, mb.NewClient(h)
	default: // "tcp" / "" → Modbus/TCP
		h := mb.NewTCPClientHandler(d.Addr())
		h.Timeout = timeout
		h.IdleTimeout = 60 * time.Second
		h.Logger = fl
		h.SetSlave(d.UnitID)
		return h, mb.NewClient(h)
	}
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
		s, ok := DecodeSample(dev.cfg.ID, pt, raw)
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

func (p *Poller) read(ctx context.Context, dev *device, pt Point) ([]byte, error) {
	retries := dev.cfg.Retries
	if retries < 0 {
		retries = 0
	}
	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		data, err := Read(ctx, dev.client, pt)
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

// ResolvePoint derives the routing point type/index and read parameters from a
// Modbus mapping. Shared by the Modbus/TCP poller and the serial RTU source.
func ResolvePoint(m config.SignalMapping) (Point, error) {
	pt := Point{
		// Routing point type is the function itself so that e.g. holding@0 and
		// input@0 (or coil@0 and discrete@0) don't collide on (pointType,index).
		Ptype:     source.PointType(m.Function),
		Index:     m.Address,
		Function:  m.Function,
		Address:   m.Address,
		Quantity:  m.Quantity,
		DataType:  m.DataType,
		ByteOrder: m.ByteOrder,
	}
	switch m.Function {
	case "coil", "discrete_input":
		if pt.Quantity == 0 {
			pt.Quantity = 1
		}
	case "input_register", "holding_register":
		if pt.Quantity == 0 {
			q, err := registersFor(m.DataType)
			if err != nil {
				return Point{}, err
			}
			pt.Quantity = q
		}
	default:
		return Point{}, fmt.Errorf("unsupported function %q (mapping %q)", m.Function, m.MetricName)
	}
	return pt, nil
}

// Read issues the single Modbus read for pt against client and returns the raw
// response bytes. Transport-agnostic: works with any grid-x mb.Client (TCP or
// RTU). Shared by the Modbus/TCP poller and the serial RTU source.
func Read(ctx context.Context, client mb.Client, pt Point) ([]byte, error) {
	switch pt.Function {
	case "coil":
		return client.ReadCoils(ctx, pt.Address, pt.Quantity)
	case "discrete_input":
		return client.ReadDiscreteInputs(ctx, pt.Address, pt.Quantity)
	case "input_register":
		return client.ReadInputRegisters(ctx, pt.Address, pt.Quantity)
	case "holding_register":
		return client.ReadHoldingRegisters(ctx, pt.Address, pt.Quantity)
	default:
		return nil, fmt.Errorf("unsupported function %q", pt.Function)
	}
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

// DecodeSample turns raw Modbus bytes into a source.Sample for the point.
// Shared by the Modbus/TCP poller and the serial RTU source.
func DecodeSample(deviceID string, pt Point, raw []byte) (source.Sample, bool) {
	s := source.Sample{
		SourceID:  deviceID,
		PointType: pt.Ptype,
		Index:     pt.Index,
		Time:      time.Now(),
		Quality:   source.QualityOnline, // a successful read is, by definition, online
		IsEvent:   false,                // Modbus reads are polls, never events
	}
	switch pt.Function {
	case "coil", "discrete_input":
		if len(raw) < 1 {
			return source.Sample{}, false
		}
		s.BoolValue = raw[0]&0x01 != 0
	default: // registers
		v, ok := decodeNumeric(raw, pt.DataType, pt.ByteOrder)
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
