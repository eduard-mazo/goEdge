package modbus

import (
	"encoding/binary"
	"io"
	"math"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"goMqttDnp3/config"
	"goMqttDnp3/source"
)

// recorder is a test source.Handler that captures samples and log lines.
type recorder struct {
	mu      sync.Mutex
	samples []source.Sample
	logs    []string
}

func (r *recorder) OnSample(s source.Sample) {
	r.mu.Lock()
	r.samples = append(r.samples, s)
	r.mu.Unlock()
}
func (r *recorder) OnStatusChange(source.Status) {}
func (r *recorder) OnLog(level, msg string) {
	r.mu.Lock()
	r.logs = append(r.logs, level+": "+msg)
	r.mu.Unlock()
}

// fakeRTUServer accepts RTU-over-TCP connections and replies to every FC03
// request with the canned response carrying registers 7007=17031, 7008=53295 —
// the two holding registers the user's bench tool shows as float 67.906609.
func fakeRTUServer(t *testing.T) (addr string, stop func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	// Response ADU (RTU framing): slave 0x15, FC 0x03, bytecount 0x04,
	// 0x4287 0xD02F, CRC 0x16 0x7F.
	resp := []byte{0x15, 0x03, 0x04, 0x42, 0x87, 0xD0, 0x2F, 0x16, 0x7F}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				buf := make([]byte, 8) // FC03 request ADU is 8 bytes
				for {
					if _, err := io.ReadFull(c, buf); err != nil {
						return
					}
					_, _ = c.Write(resp)
				}
			}(conn)
		}
	}()
	return ln.Addr().String(), func() { _ = ln.Close() }
}

func TestRTUOverTCPDecodesFloat(t *testing.T) {
	addr, stop := fakeRTUServer(t)
	defer stop()

	host, portStr, _ := net.SplitHostPort(addr)
	port := 0
	for _, c := range portStr {
		port = port*10 + int(c-'0')
	}

	rec := &recorder{}
	p := New(rec)
	dev := config.ModbusDevice{
		ID:         "RTU_TCP",
		Host:       host,
		Port:       port,
		Transport:  "rtuovertcp",
		UnitID:     21,
		ScanRateMs: 50,
		TimeoutMs:  500,
		LogFrames:  true,
		Enabled:    true,
	}
	mapping := config.SignalMapping{
		Protocol: "modbus", SourceID: "RTU_TCP", Enabled: true,
		Function: "holding_register", Address: 7006, Quantity: 2,
		DataType: "float32", ByteOrder: "ABCD",
	}
	if err := p.AddDevice(dev, []config.SignalMapping{mapping}); err != nil {
		t.Fatalf("AddDevice: %v", err)
	}

	if err := p.Start(t.Context()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer p.Stop()

	// Wait for the first decoded sample.
	deadline := time.After(3 * time.Second)
	for {
		rec.mu.Lock()
		n := len(rec.samples)
		rec.mu.Unlock()
		if n > 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("no sample received")
		case <-time.After(10 * time.Millisecond):
		}
	}

	rec.mu.Lock()
	got := rec.samples[0].FloatValue
	logs := strings.Join(rec.logs, "\n")
	rec.mu.Unlock()

	const want = 67.906609
	if math.Abs(got-want) > 1e-4 {
		t.Errorf("decoded float = %v, want ≈ %v", got, want)
	}

	// Frame logging must surface the raw send/recv hex in the log (the Registro).
	if !strings.Contains(logs, "send") || !strings.Contains(logs, "recv") {
		t.Errorf("expected raw send/recv frames in log, got:\n%s", logs)
	}
	if !strings.Contains(logs, "15 03") {
		t.Errorf("expected hex frame with slave 0x15 in log, got:\n%s", logs)
	}
}

// Sanity guard: the canned registers really are 67.906609 under ABCD, so a
// failure above points at the transport/wiring, not the fixture.
func TestFixtureFloatSanity(t *testing.T) {
	b := []byte{0x42, 0x87, 0xD0, 0x2F}
	f := math.Float32frombits(binary.BigEndian.Uint32(b))
	if math.Abs(float64(f)-67.906609) > 1e-4 {
		t.Fatalf("fixture float = %v", f)
	}
}
