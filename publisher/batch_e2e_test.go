package publisher

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"goMqttDnp3/config"
)

// --- minimal in-process Modbus/TCP slave (mirrors cmd/modbussmoke) ------------

type e2eSlave struct {
	mu    sync.Mutex
	regs  [256]uint16
	coils [256]bool
}

func (s *e2eSlave) serve(ln net.Listener) {
	for {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		go s.handle(c)
	}
}

func (s *e2eSlave) handle(c net.Conn) {
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
			resp = []byte{fc | 0x80, 0x01}
		}
		s.mu.Unlock()

		out := make([]byte, 7+len(resp))
		copy(out[0:2], hdr[0:2])
		binary.BigEndian.PutUint16(out[4:6], uint16(1+len(resp)))
		out[6] = hdr[6]
		copy(out[7:], resp)
		if _, err := c.Write(out); err != nil {
			return
		}
	}
}

// bufferedCount peeks the offline buffer length (test-only).
func (p *Publisher) bufferedCount() int {
	p.bufMu.Lock()
	defer p.bufMu.Unlock()
	return len(p.msgBuf)
}

// e2eConfig builds a 5-mapping single-device config against the slave at port.
// ScanRate is large so only the immediate first scan runs during the test.
func e2eConfig(port, batchMs int) config.AppConfig {
	dev := config.ModbusDevice{
		ID: "plc", Label: "Sim PLC", Host: "127.0.0.1", Port: port,
		UnitID: 1, ScanRateMs: 60000, TimeoutMs: 2000, Retries: 1, Enabled: true,
	}
	m := func(metric, fn string, addr uint16, dtype string) config.SignalMapping {
		return config.SignalMapping{
			Protocol: "modbus", SourceID: "plc", DeviceID: "DEVA",
			Function: fn, Address: addr, DataType: dtype, ByteOrder: "ABCD",
			MetricName: metric, Enabled: true,
		}
	}
	return config.AppConfig{
		// Dead broker: never connects, so the publish path lands in the offline
		// buffer where the test inspects the exact message grouping.
		MQTT: config.MQTTConfig{Broker: "tcp://127.0.0.1:1", ClientID: "e2e", PublishBatchMs: batchMs},
		Sparkplug: config.SparkplugConfig{GroupID: "g", NodeID: "n"},
		ModbusDevices: []config.ModbusDevice{dev},
		Mappings: []config.SignalMapping{
			m("DEVA/level", "holding_register", 0, "float32"),
			m("DEVA/count", "holding_register", 2, "uint16"),
			m("DEVA/temp", "holding_register", 3, "int16"),
			m("DEVA/run", "coil", 0, "bool"),
			m("DEVA/fault", "discrete_input", 0, "bool"),
		},
	}
}

// startSlave spins up the in-process slave with seeded data and returns its port.
func startSlave(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	sl := &e2eSlave{}
	sl.regs[0], sl.regs[1] = 0x4248, 0x0000 // float32 50.0 (ABCD)
	sl.regs[2] = 1234
	var temp int16 = -50
	sl.regs[3] = uint16(temp) //nolint:gosec // pack a negative int16 into the register
	sl.coils[0] = true
	go sl.serve(ln)
	t.Cleanup(func() { ln.Close() })
	return ln.Addr().(*net.TCPAddr).Port
}

// waitBuffered polls until the offline buffer reaches at least want, then a short
// settle so a full scan's worth of messages has landed. Fails on timeout.
func waitBuffered(t *testing.T, p *Publisher, want int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if p.bufferedCount() >= want {
			time.Sleep(100 * time.Millisecond) // settle: catch any straggler
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d buffered message(s), got %d", want, p.bufferedCount())
}

// TestE2E_Batching_GroupsScanIntoOneMessage drives the real Modbus poller →
// mapping → publisher pipeline and asserts a device's whole scan publishes as a
// SINGLE grouped message when batching is on.
func TestE2E_Batching_GroupsScanIntoOneMessage(t *testing.T) {
	port := startSlave(t)
	p := New(e2eConfig(port, 150)) // 150ms coalesce window
	if err := p.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer p.Stop()

	waitBuffered(t, p, 1)
	buf := p.drainTestBuffer()
	if len(buf) != 1 {
		t.Fatalf("batching on: want 1 grouped message for the scan, got %d", len(buf))
	}
	if len(buf[0].metrics) != 5 {
		t.Fatalf("grouped message: want all 5 metrics together, got %d", len(buf[0].metrics))
	}
	if buf[0].deviceID != "DEVA" {
		t.Fatalf("grouped message: want deviceID DEVA, got %q", buf[0].deviceID)
	}
}

// TestE2E_NoBatching_OneMessagePerSignal is the control: with batching off the
// same scan produces one message per signal (today's legacy behavior).
func TestE2E_NoBatching_OneMessagePerSignal(t *testing.T) {
	port := startSlave(t)
	p := New(e2eConfig(port, -1)) // batching disabled (negative = explicit opt-out)
	if err := p.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer p.Stop()

	waitBuffered(t, p, 5)
	buf := p.drainTestBuffer()
	if len(buf) != 5 {
		t.Fatalf("batching off: want 5 separate messages, got %d", len(buf))
	}
	for _, msg := range buf {
		if len(msg.metrics) != 1 {
			t.Fatalf("batching off: each message should carry 1 metric, got %+v", msg)
		}
	}
}
