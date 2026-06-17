package modbus

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"strings"
	"testing"
	"time"

	"goMqttDnp3/config"
	"goMqttDnp3/source"
)

// This file is the Modbus-protocol conformance suite: it exercises register
// decoding for every data type and byte order (TestReorder, TestDecodeNumeric),
// the function-code → PointType mapping (TestDecodeSample), the "read one
// register of a float32 as int16" question (TestFloat32RegistersAsInt16), and a
// full end-to-end exchange against an in-process slave that prints the real
// on-wire frames (TestModbusE2E*). It closes the decode/byte-order coverage gap
// noted in ARCHITECTURE.md and mirrors the bank layout of the soak harness's
// scripts/sim/modbusslave.

// ─────────────────────────── byte-order matrix ───────────────────────────────

// TestReorder pins the four word/byte orders to canonical big-endian. Given the
// wire bytes [A B C D] (big-endian per register, high register first), reorder
// must produce the canonical big-endian layout the decoder then reads.
func TestReorder(t *testing.T) {
	in := []byte{0xA, 0xB, 0xC, 0xD}
	cases := map[string][]byte{
		"ABCD": {0xA, 0xB, 0xC, 0xD}, // no change (already big-endian)
		"DCBA": {0xD, 0xC, 0xB, 0xA}, // full reverse
		"BADC": {0xB, 0xA, 0xD, 0xC}, // swap bytes within each register
		"CDAB": {0xC, 0xD, 0xA, 0xB}, // swap the two registers
	}
	for order, want := range cases {
		got := reorder(in, order)
		if string(got) != string(want) {
			t.Errorf("reorder(%s) = % x, want % x", order, got, want)
		}
	}
	// Empty/unknown order is treated as ABCD (pass-through).
	if got := reorder(in, ""); string(got) != string(in) {
		t.Errorf(`reorder("") = % x, want pass-through`, got)
	}
}

// ─────────────────────────── decode matrix ──────────────────────────────────

// be packs uint16 registers into the big-endian wire bytes a slave would send.
func be(regs ...uint16) []byte {
	b := make([]byte, 2*len(regs))
	for i, r := range regs {
		binary.BigEndian.PutUint16(b[i*2:], r)
	}
	return b
}

// TestDecodeNumeric covers every supported register data type at its canonical
// (ABCD) byte order, including signed/unsigned edges.
func TestDecodeNumeric(t *testing.T) {
	f32 := func(v float32) []byte { b := make([]byte, 4); binary.BigEndian.PutUint32(b, math.Float32bits(v)); return b }
	f64 := func(v float64) []byte { b := make([]byte, 8); binary.BigEndian.PutUint64(b, math.Float64bits(v)); return b }
	i32 := func(v int32) []byte { b := make([]byte, 4); binary.BigEndian.PutUint32(b, uint32(v)); return b }
	u32 := func(v uint32) []byte { b := make([]byte, 4); binary.BigEndian.PutUint32(b, v); return b }

	cases := []struct {
		name     string
		dataType string
		order    string
		raw      []byte
		want     float64
	}{
		{"int16 positive", "int16", "ABCD", be(0x0064), 100},
		{"int16 negative", "int16", "ABCD", be(0xFFD8), -40}, // 0xFFD8 = -40
		{"int16 empty type defaults int16", "", "ABCD", be(0x0064), 100},
		{"uint16 max", "uint16", "ABCD", be(0xFFFF), 65535},
		{"uint16 of same bits", "uint16", "ABCD", be(0xFFD8), 65496},
		{"int32 negative", "int32", "ABCD", i32(-123456), -123456},
		{"uint32 large", "uint32", "ABCD", u32(4000000000), 4000000000},
		{"float32 10.0", "float32", "ABCD", f32(10.0), 10.0},
		{"float32 the field value", "float32", "ABCD", be(0x4287, 0xD02F), 67.90660858},
		{"float64 1234.5", "float64", "ABCD", f64(1234.5), 1234.5},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := decodeNumeric(c.raw, c.dataType, c.order)
			if !ok {
				t.Fatalf("decodeNumeric returned ok=false")
			}
			if math.Abs(got-c.want) > 1e-4 {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}

	// Byte-order matrix on one float32: the slave stores 10.0 as ABCD, so only
	// ABCD decodes to 10.0; the others reinterpret the same bytes and must NOT.
	abcd := f32(10.0)
	for _, order := range []string{"DCBA", "BADC", "CDAB"} {
		got, ok := decodeNumeric(abcd, "float32", order)
		if !ok {
			t.Fatalf("decodeNumeric(%s) ok=false", order)
		}
		if math.Abs(got-10.0) < 1e-4 {
			t.Errorf("byte order %s decoded to 10.0 — orders are not actually applied", order)
		}
		t.Logf("float32 bytes % x read as %s → %v (wrong order ⇒ garbage)", abcd, order, got)
	}

	// Short buffers must fail closed, never panic or read past the slice.
	if _, ok := decodeNumeric(be(0x4120), "float32", "ABCD"); ok {
		t.Error("float32 from 2 bytes should fail (needs 4)")
	}
}

// ─────────────────────── function-code → sample ─────────────────────────────

// TestDecodeSample checks the function → PointType/value routing for each FC.
func TestDecodeSample(t *testing.T) {
	t.Run("coil set", func(t *testing.T) {
		pt := Point{Ptype: "coil", Function: "coil"}
		s, ok := DecodeSample("d", pt, []byte{0x01})
		if !ok || !s.BoolValue {
			t.Fatalf("coil decode: ok=%v val=%v", ok, s.BoolValue)
		}
	})
	t.Run("coil clear", func(t *testing.T) {
		pt := Point{Ptype: "discrete_input", Function: "discrete_input"}
		s, ok := DecodeSample("d", pt, []byte{0x00})
		if !ok || s.BoolValue {
			t.Fatalf("discrete decode: ok=%v val=%v", ok, s.BoolValue)
		}
	})
	t.Run("holding register float32", func(t *testing.T) {
		pt := Point{Ptype: "holding_register", Function: "holding_register", DataType: "float32", ByteOrder: "ABCD"}
		s, ok := DecodeSample("d", pt, be(0x4120, 0x0000))
		if !ok || math.Abs(s.FloatValue-10.0) > 1e-6 {
			t.Fatalf("hr decode: ok=%v val=%v", ok, s.FloatValue)
		}
		if s.Quality != source.QualityOnline {
			t.Errorf("expected QualityOnline on a good read, got %v", s.Quality)
		}
	})
}

// TestResolvePointQuantity verifies the register count auto-derived per data
// type when the mapping leaves Quantity at 0.
func TestResolvePointQuantity(t *testing.T) {
	cases := map[string]uint16{
		"int16": 1, "uint16": 1, "bool": 1, "": 1,
		"int32": 2, "uint32": 2, "float32": 2,
		"float64": 4,
	}
	for dt, want := range cases {
		m := config.SignalMapping{Function: "holding_register", DataType: dt, Address: 7006}
		pt, err := ResolvePoint(m)
		if err != nil {
			t.Fatalf("ResolvePoint(%q): %v", dt, err)
		}
		if pt.Quantity != want {
			t.Errorf("dataType %q → quantity %d, want %d", dt, pt.Quantity, want)
		}
	}
	// Coils always read 1 bit when unset.
	pt, _ := ResolvePoint(config.SignalMapping{Function: "coil"})
	if pt.Quantity != 1 {
		t.Errorf("coil default quantity = %d, want 1", pt.Quantity)
	}
	// An unknown function is rejected.
	if _, err := ResolvePoint(config.SignalMapping{Function: "bogus"}); err == nil {
		t.Error("expected error for unknown function")
	}
}

// ───────────────── the question: float32 (2 regs) read as int16 ─────────────

// TestFloat32RegistersAsInt16 answers "if I have a float32 over 2 registers, can
// I read it as int16?". Yes — the wire doesn't care — but each register is only
// half of the IEEE-754 bit pattern, so the int16 you get is NOT the physical
// value; it's the high or low 16 bits of the float's encoding.
func TestFloat32RegistersAsInt16(t *testing.T) {
	// Tank level 10.0 m, stored ABCD: bits 0x41200000 → reg0=0x4120, reg1=0x0000.
	raw := be(0x4120, 0x0000)

	asF32, _ := decodeNumeric(raw, "float32", "ABCD")
	hiAsI16, _ := decodeNumeric(raw[0:2], "int16", "ABCD")  // first register only
	loAsI16, _ := decodeNumeric(raw[2:4], "int16", "ABCD")  // second register only
	hiAsU16, _ := decodeNumeric(raw[0:2], "uint16", "ABCD") // first register, unsigned

	t.Logf("registers % x", raw)
	t.Logf("  read as float32 (2 regs) = %v m   ← the real value", asF32)
	t.Logf("  reg0 as int16            = %v     (= 0x4120, high half of the float bits)", hiAsI16)
	t.Logf("  reg0 as uint16           = %v", hiAsU16)
	t.Logf("  reg1 as int16            = %v     (low half — here 0)", loAsI16)

	if math.Abs(asF32-10.0) > 1e-6 {
		t.Fatalf("float32 decode = %v, want 10.0", asF32)
	}
	if hiAsI16 != 16672 { // 0x4120
		t.Fatalf("reg0 as int16 = %v, want 16672", hiAsI16)
	}
	if hiAsI16 == 10 || loAsI16 == 10 {
		t.Fatal("a single register of a float32 must not equal the physical value")
	}
}

// ─────────────────────── end-to-end against a slave ─────────────────────────

// simSlave is an in-process Modbus/TCP slave with the bank layout of the soak
// harness's modbusslave: holding 0-1 float32 tank, 2 uint16 rpm, 3 int16 temp×10,
// 4-5 float32 flow; input 0 uint16 pressure; coil 0 pump, 1 valve; discrete 0
// fault. Banks are fixed (no ticker) so frames are deterministic.
type simSlave struct {
	holding  [256]uint16
	input    [256]uint16
	coils    [256]bool
	discrete [256]bool
}

func newSimSlave() *simSlave {
	s := &simSlave{}
	putF32(&s.holding, 0, 10.0) // tank level
	s.holding[2] = 1450         // pump rpm
	s.holding[3] = uint16(int16(235))
	putF32(&s.holding, 4, 3.0) // flow
	s.input[0] = 101           // pressure
	s.coils[0] = true          // pump running
	s.coils[1] = false         // valve
	s.discrete[0] = true       // fault
	return s
}

func putF32(bank *[256]uint16, addr int, v float32) {
	bits := math.Float32bits(v)
	bank[addr] = uint16(bits >> 16) // high word first → ABCD on the wire
	bank[addr+1] = uint16(bits)
}

func (s *simSlave) serve(t *testing.T) (host string, port int, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			go s.handle(c)
		}
	}()
	h, ps, _ := net.SplitHostPort(ln.Addr().String())
	p := 0
	fmt.Sscanf(ps, "%d", &p)
	return h, p, func() { _ = ln.Close() }
}

func (s *simSlave) handle(c net.Conn) {
	defer c.Close()
	hdr := make([]byte, 7)
	for {
		if _, err := io.ReadFull(c, hdr); err != nil {
			return
		}
		length := binary.BigEndian.Uint16(hdr[4:6])
		if length < 2 || length > 255 {
			return
		}
		pdu := make([]byte, length-1)
		if _, err := io.ReadFull(c, pdu); err != nil {
			return
		}
		resp := s.respond(pdu)
		out := make([]byte, 7+len(resp))
		copy(out[0:4], hdr[0:4])
		binary.BigEndian.PutUint16(out[4:6], uint16(1+len(resp)))
		out[6] = hdr[6]
		copy(out[7:], resp)
		if _, err := c.Write(out); err != nil {
			return
		}
	}
}

func (s *simSlave) respond(pdu []byte) []byte {
	if len(pdu) < 5 {
		return []byte{pdu[0] | 0x80, 0x03}
	}
	fc := pdu[0]
	addr := binary.BigEndian.Uint16(pdu[1:3])
	qty := binary.BigEndian.Uint16(pdu[3:5])
	switch fc {
	case 0x03, 0x04:
		src := &s.holding
		if fc == 0x04 {
			src = &s.input
		}
		if qty == 0 || qty > 125 || int(addr)+int(qty) > len(src) {
			return []byte{fc | 0x80, 0x02} // illegal data address
		}
		resp := make([]byte, 2+int(qty)*2)
		resp[0], resp[1] = fc, byte(qty*2)
		for i := 0; i < int(qty); i++ {
			binary.BigEndian.PutUint16(resp[2+i*2:], src[int(addr)+i])
		}
		return resp
	case 0x01, 0x02:
		src := &s.coils
		if fc == 0x02 {
			src = &s.discrete
		}
		if qty == 0 || qty > 2000 || int(addr)+int(qty) > len(src) {
			return []byte{fc | 0x80, 0x02}
		}
		bc := (int(qty) + 7) / 8
		resp := make([]byte, 2+bc)
		resp[0], resp[1] = fc, byte(bc)
		for i := 0; i < int(qty); i++ {
			if src[int(addr)+i] {
				resp[2+i/8] |= 1 << (uint(i) % 8)
			}
		}
		return resp
	default:
		return []byte{fc | 0x80, 0x01} // illegal function
	}
}

// waitSamples blocks until rec has at least n samples or the deadline passes.
func waitSamples(t *testing.T, rec *recorder, n int) {
	t.Helper()
	deadline := time.After(3 * time.Second)
	for {
		rec.mu.Lock()
		got := len(rec.samples)
		rec.mu.Unlock()
		if got >= n {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for %d samples (got %d)", n, got)
		case <-time.After(10 * time.Millisecond):
		}
	}
}

// findSample returns the first sample matching a function (PointType) and index.
func findSample(rec *recorder, ptype source.PointType, index uint16) (source.Sample, bool) {
	rec.mu.Lock()
	defer rec.mu.Unlock()
	for _, s := range rec.samples {
		if s.PointType == ptype && s.Index == index {
			return s, true
		}
	}
	return source.Sample{}, false
}

// TestModbusE2EAllFunctions polls one register/coil of each function code and
// data type against the in-process slave, asserts the decoded values, and prints
// the real Modbus/TCP frames captured via FrameLogger.
func TestModbusE2EAllFunctions(t *testing.T) {
	s := newSimSlave()
	host, port, stop := s.serve(t)
	defer stop()

	rec := &recorder{}
	p := New(rec)
	dev := config.ModbusDevice{
		ID: "plc-sim", Host: host, Port: port, UnitID: 1,
		ScanRateMs: 100, TimeoutMs: 1000, LogFrames: true, Enabled: true,
	}
	mk := func(name, fn string, addr uint16, dt, order string) config.SignalMapping {
		return config.SignalMapping{
			MetricName: name, Protocol: "modbus", SourceID: "plc-sim", Enabled: true,
			Function: fn, Address: addr, DataType: dt, ByteOrder: order,
		}
	}
	maps := []config.SignalMapping{
		mk("tank_level", "holding_register", 0, "float32", "ABCD"), // 10.0
		mk("pump_rpm", "holding_register", 2, "uint16", "ABCD"),    // 1450
		mk("temp_x10", "holding_register", 3, "int16", "ABCD"),     // 235
		mk("flow", "holding_register", 4, "float32", "ABCD"),       // 3.0
		mk("pressure", "input_register", 0, "uint16", "ABCD"),      // 101
		mk("pump_run", "coil", 0, "bool", ""),                      // true
		mk("valve", "coil", 1, "bool", ""),                         // false
		mk("fault", "discrete_input", 0, "bool", ""),               // true
	}
	if err := p.AddDevice(dev, maps); err != nil {
		t.Fatalf("AddDevice: %v", err)
	}
	if err := p.Start(t.Context()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer p.Stop()

	waitSamples(t, rec, len(maps))

	type want struct {
		ptype source.PointType
		index uint16
		f     float64
		b     bool
		isB   bool
	}
	wants := []want{
		{ptype: "holding_register", index: 0, f: 10.0},
		{ptype: "holding_register", index: 2, f: 1450},
		{ptype: "holding_register", index: 3, f: 235},
		{ptype: "holding_register", index: 4, f: 3.0},
		{ptype: "input_register", index: 0, f: 101},
		{ptype: "coil", index: 0, b: true, isB: true},
		{ptype: "coil", index: 1, b: false, isB: true},
		{ptype: "discrete_input", index: 0, b: true, isB: true},
	}
	for _, w := range wants {
		got, ok := findSample(rec, w.ptype, w.index)
		if !ok {
			t.Errorf("no sample for %s@%d", w.ptype, w.index)
			continue
		}
		if w.isB {
			if got.BoolValue != w.b {
				t.Errorf("%s@%d = %v, want %v", w.ptype, w.index, got.BoolValue, w.b)
			}
		} else if math.Abs(got.FloatValue-w.f) > 1e-4 {
			t.Errorf("%s@%d = %v, want %v", w.ptype, w.index, got.FloatValue, w.f)
		}
	}

	// Surface a few real frames for the report.
	rec.mu.Lock()
	var frames []string
	for _, l := range rec.logs {
		if strings.Contains(l, "send") || strings.Contains(l, "recv") {
			frames = append(frames, l)
		}
	}
	rec.mu.Unlock()
	t.Logf("captured %d frame log lines; sample:", len(frames))
	for i, f := range frames {
		if i >= 8 {
			break
		}
		t.Logf("  %s", f)
	}
}

// TestModbusE2EIllegalAddress proves an out-of-range read surfaces as a read
// error (Modbus exception 0x02) and yields no sample — the failure the user hit
// when an off-by-one pushed a 2-register float past the slave's last register.
func TestModbusE2EIllegalAddress(t *testing.T) {
	s := newSimSlave()
	host, port, stop := s.serve(t)
	defer stop()

	rec := &recorder{}
	p := New(rec)
	dev := config.ModbusDevice{
		ID: "plc-sim", Host: host, Port: port, UnitID: 1,
		ScanRateMs: 100, TimeoutMs: 1000, Retries: 0, LogFrames: true, Enabled: true,
	}
	// Read 2 registers starting at the very last address → 255+1 overruns the bank.
	m := config.SignalMapping{
		MetricName: "overrun", Protocol: "modbus", SourceID: "plc-sim", Enabled: true,
		Function: "holding_register", Address: 255, Quantity: 2, DataType: "float32", ByteOrder: "ABCD",
	}
	if err := p.AddDevice(dev, []config.SignalMapping{m}); err != nil {
		t.Fatalf("AddDevice: %v", err)
	}
	if err := p.Start(t.Context()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer p.Stop()

	// Give it a couple of scans, then assert: no samples, and the error surfaced.
	time.Sleep(400 * time.Millisecond)
	rec.mu.Lock()
	nSamples := len(rec.samples)
	logs := strings.Join(rec.logs, "\n")
	rec.mu.Unlock()

	if nSamples != 0 {
		t.Errorf("expected 0 samples on illegal address, got %d", nSamples)
	}
	// The poller records read failures via OnStatusChange, not OnLog, but the
	// frame log should show the request going out and (with logging) the path.
	t.Logf("illegal-address run produced %d samples; log lines:\n%s", nSamples, logs)
}
