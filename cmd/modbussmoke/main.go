// Command modbussmoke is a self-contained runtime test for the Modbus source
// adapter. It starts a minimal in-process Modbus/TCP slave, points a
// modbus.Poller at it, and verifies that decoded samples (float32, uint16, coil)
// flow through with the right values. Pure Go — no cgo, no external broker.
//
// Run:  go run ./cmd/modbussmoke   (exit 0 = PASS)
package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"goMqttDnp3/config"
	"goMqttDnp3/modbus"
	"goMqttDnp3/source"
)

// --- minimal Modbus/TCP slave -------------------------------------------------

type slave struct {
	mu    sync.Mutex
	regs  [256]uint16
	coils [256]bool
}

func (s *slave) tick(level *float32, count *uint16, coil *bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bits := math.Float32bits(*level)
	s.regs[0] = uint16(bits >> 16) // HR0 = high word  (ABCD big-endian on the wire)
	s.regs[1] = uint16(bits)       // HR1 = low word
	s.regs[2] = *count             // HR2 = uint16 counter
	s.coils[0] = *coil
	*level += 0.5
	*count++
	*coil = !*coil
}

func (s *slave) serve(ln net.Listener) {
	for {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		go s.handle(c)
	}
}

func (s *slave) handle(c net.Conn) {
	defer c.Close()
	hdr := make([]byte, 7)
	for {
		if _, err := io.ReadFull(c, hdr); err != nil {
			return
		}
		length := binary.BigEndian.Uint16(hdr[4:6])
		if length < 1 {
			return
		}
		pdu := make([]byte, length-1)
		if _, err := io.ReadFull(c, pdu); err != nil {
			return
		}
		fc := pdu[0]
		addr := binary.BigEndian.Uint16(pdu[1:3])
		qty := binary.BigEndian.Uint16(pdu[3:5])

		var resp []byte
		s.mu.Lock()
		switch fc {
		case 0x03, 0x04: // read holding / input registers
			bc := byte(qty * 2)
			resp = make([]byte, 2+int(bc))
			resp[0], resp[1] = fc, bc
			for i := 0; i < int(qty); i++ {
				binary.BigEndian.PutUint16(resp[2+i*2:], s.regs[addr+uint16(i)])
			}
		case 0x01, 0x02: // read coils / discrete inputs
			bc := byte((qty + 7) / 8)
			resp = make([]byte, 2+int(bc))
			resp[0], resp[1] = fc, bc
			for i := 0; i < int(qty); i++ {
				if s.coils[addr+uint16(i)] {
					resp[2+i/8] |= 1 << (uint(i) % 8)
				}
			}
		default:
			resp = []byte{fc | 0x80, 0x01} // illegal function
		}
		s.mu.Unlock()

		out := make([]byte, 7+len(resp))
		copy(out[0:2], hdr[0:2]) // echo transaction id
		binary.BigEndian.PutUint16(out[4:6], uint16(1+len(resp)))
		out[6] = hdr[6] // unit id
		copy(out[7:], resp)
		if _, err := c.Write(out); err != nil {
			return
		}
	}
}

// --- recorder (source.Handler) ------------------------------------------------

type recorder struct {
	mu        sync.Mutex
	connected bool
	byMetric  map[string]float64 // last value seen per "type:idx"
	count     int
}

func (r *recorder) OnSample(s source.Sample) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.count++
	key := fmt.Sprintf("%s:%d", s.PointType, s.Index)
	switch s.PointType {
	case source.PointBinary:
		if s.BoolValue {
			r.byMetric[key] = 1
		} else {
			r.byMetric[key] = 0
		}
	default:
		r.byMetric[key] = s.FloatValue
	}
}

func (r *recorder) OnStatusChange(s source.Status) {
	r.mu.Lock()
	if s.Connected {
		r.connected = true
	}
	r.mu.Unlock()
	log.Printf("status: id=%s connected=%v err=%q rx=%d", s.ID, s.Connected, s.LastError, s.MeasurementsRx)
}

func (r *recorder) OnLog(level, msg string) { log.Printf("[%s] %s", level, msg) }

func main() {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	sl := &slave{}
	level := float32(12.5)
	count := uint16(1000)
	coil := true
	sl.tick(&level, &count, &coil) // seed
	go sl.serve(ln)
	go func() {
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()
		for range t.C {
			sl.tick(&level, &count, &coil)
		}
	}()
	log.Printf("modbus slave on 127.0.0.1:%d", port)

	rec := &recorder{byMetric: make(map[string]float64)}
	p := modbus.New(rec)
	dev := config.ModbusDevice{ID: "plc", Label: "Sim PLC", Host: "127.0.0.1", Port: port,
		UnitID: 1, ScanRateMs: 400, TimeoutMs: 2000, Retries: 1, Enabled: true}
	mappings := []config.SignalMapping{
		{Protocol: "modbus", SourceID: "plc", Function: "holding_register", Address: 0, DataType: "float32", ByteOrder: "ABCD", MetricName: "mb/level", Enabled: true},
		{Protocol: "modbus", SourceID: "plc", Function: "holding_register", Address: 2, DataType: "uint16", MetricName: "mb/count", Enabled: true},
		{Protocol: "modbus", SourceID: "plc", Function: "coil", Address: 0, MetricName: "mb/run", Enabled: true},
	}
	if err := p.AddDevice(dev, mappings); err != nil {
		log.Fatalf("AddDevice: %v", err)
	}
	if err := p.Start(context.Background()); err != nil {
		log.Fatalf("Start: %v", err)
	}
	time.Sleep(3 * time.Second)
	p.Stop()
	_ = ln.Close()

	rec.mu.Lock()
	defer rec.mu.Unlock()
	fmt.Println("\n──────── modbus smoke summary ────────")
	fmt.Printf("connected:    %v\n", rec.connected)
	fmt.Printf("samples:      %d\n", rec.count)
	for k, v := range rec.byMetric {
		fmt.Printf("  %-12s last=%g\n", k, v)
	}
	for _, st := range p.Status() {
		fmt.Printf("final status: id=%s connected=%v rx=%d err=%q\n", st.ID, st.Connected, st.MeasurementsRx, st.LastError)
	}
	fmt.Println("──────────────────────────────────────")

	// Expect: connected, samples for holding_register:0 (float ~12.5+),
	// holding_register:2 (uint ~1000+), coil:0. The modbus poller routes points by
	// function name (so holding@0 and input@0 don't collide), so the recorder keys
	// are "<function>:<address>". float32 decode sanity: level was ≥12.5.
	lvl, okL := rec.byMetric["holding_register:0"]
	cnt, okC := rec.byMetric["holding_register:2"]
	_, okB := rec.byMetric["coil:0"]
	if !rec.connected || !okL || !okC || !okB {
		fmt.Println("FAIL: missing connection or one of the expected points")
		os.Exit(1)
	}
	if lvl < 12.0 || cnt < 1000 {
		fmt.Printf("FAIL: decoded values look wrong (level=%g count=%g)\n", lvl, cnt)
		os.Exit(1)
	}
	fmt.Println("PASS")
}
