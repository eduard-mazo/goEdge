//go:build dnp3_ffi

// Command dnp3smoke is a runtime smoke test for the opendnp3 master binding.
// It drives the dnp3 package directly (no MQTT/Sparkplug) against a DNP3
// outstation — typically scripts/sim/outstation_sim — and reports whether
// measurements, status changes, and an on-demand integrity poll work.
//
// Build:  go build -tags dnp3_ffi ./cmd/dnp3smoke   (with the opendnp3 cgo flags)
// Run:    dnp3smoke -addr 127.0.0.1:20000 -duration 12s
//
// Exit code 0 only if the outstation connected AND at least one measurement of
// each polled kind was received; non-zero otherwise (so it's CI-usable).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"goMqttDnp3/config"
	"goMqttDnp3/dnp3"
)

type recorder struct {
	mu        sync.Mutex
	byType    map[dnp3.PointType]int
	total     int
	events    int
	static    int
	connected bool
	samples   []string
}

func newRecorder() *recorder {
	return &recorder{byType: make(map[dnp3.PointType]int)}
}

func (r *recorder) OnMeasurement(m dnp3.Measurement) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byType[m.PointType]++
	r.total++
	if m.IsEvent {
		r.events++
	} else {
		r.static++
	}
	if len(r.samples) < 10 {
		kind := "static"
		if m.IsEvent {
			kind = "event"
		}
		r.samples = append(r.samples, fmt.Sprintf("%-20s idx=%-3d %s q=0x%02x %-6s t=%s",
			m.PointType, m.Index, valueStr(m), uint8(m.Quality), kind, m.Time.Format("15:04:05.000")))
	}
}

func (r *recorder) OnStatusChange(s dnp3.OutstationStatus) {
	r.mu.Lock()
	if s.Connected {
		r.connected = true
	}
	r.mu.Unlock()
	log.Printf("status: id=%s connected=%v err=%q rx=%d", s.ID, s.Connected, s.LastError, s.MeasurementsRx)
}

func (r *recorder) OnLog(level, msg string) {
	log.Printf("[lib:%s] %s", level, msg)
}

func valueStr(m dnp3.Measurement) string {
	switch m.PointType {
	case dnp3.PointBinary, dnp3.PointBinaryOutputStatus:
		return fmt.Sprintf("val=%v", m.BoolValue)
	case dnp3.PointDoubleBitBinary:
		return fmt.Sprintf("val=%d", m.DBBValue)
	case dnp3.PointCounter, dnp3.PointFrozenCounter:
		return fmt.Sprintf("val=%d", m.UintValue)
	case dnp3.PointAnalog, dnp3.PointAnalogOutputStatus:
		return fmt.Sprintf("val=%g", m.FloatValue)
	case dnp3.PointOctetString:
		return fmt.Sprintf("val=%x", m.BytesValue)
	}
	return ""
}

func main() {
	addr := flag.String("addr", "127.0.0.1:20000", "outstation host:port")
	duration := flag.Duration("duration", 12*time.Second, "total test duration")
	flag.Parse()

	host, port, err := splitHostPort(*addr)
	if err != nil {
		log.Fatalf("bad -addr %q: %v", *addr, err)
	}

	rec := newRecorder()
	m := dnp3.New(rec)

	o := config.DNP3Outstation{
		ID:                "sim",
		Label:             "Simulator",
		Host:              host,
		Port:              port,
		MasterAddress:     1,
		OutstationAddress: 1024,
		ResponseTimeoutMs: 3000,
		KeepAliveMs:       60000,
		Class1ScanMs:      2000, // poll class-1 events
		StartupIntegrity:  true, // integrity scan on connect (all classes)
		Enabled:           true,
	}
	if err := m.AddOutstation(o); err != nil {
		log.Fatalf("AddOutstation: %v", err)
	}

	log.Printf("connecting to %s (master=1 outstation=1024), running %s", *addr, *duration)
	if err := m.Start(context.Background()); err != nil {
		log.Fatalf("Start: %v", err)
	}

	// Halfway through, fire an on-demand integrity poll.
	half := *duration / 2
	time.Sleep(half)
	log.Printf("triggering on-demand integrity poll")
	if err := m.IntegrityPoll("sim"); err != nil {
		log.Printf("IntegrityPoll error: %v", err)
	}
	time.Sleep(*duration - half)

	status := m.Status()
	m.Stop()

	// Report.
	rec.mu.Lock()
	defer rec.mu.Unlock()

	fmt.Println("\n──────── smoke test summary ────────")
	fmt.Printf("connected:    %v\n", rec.connected)
	fmt.Printf("measurements: %d total (%d event, %d static)\n", rec.total, rec.events, rec.static)
	types := make([]string, 0, len(rec.byType))
	for t := range rec.byType {
		types = append(types, string(t))
	}
	sort.Strings(types)
	for _, t := range types {
		fmt.Printf("  %-22s %d\n", t, rec.byType[dnp3.PointType(t)])
	}
	if len(rec.samples) > 0 {
		fmt.Println("sample measurements:")
		for _, s := range rec.samples {
			fmt.Printf("  %s\n", s)
		}
	}
	for _, s := range status {
		fmt.Printf("final status: id=%s connected=%v rx=%d err=%q\n", s.ID, s.Connected, s.MeasurementsRx, s.LastError)
	}
	fmt.Println("────────────────────────────────────")

	if !rec.connected {
		fmt.Println("FAIL: never connected to outstation")
		os.Exit(1)
	}
	if rec.total == 0 {
		fmt.Println("FAIL: connected but received no measurements")
		os.Exit(1)
	}
	fmt.Println("PASS")
}

func splitHostPort(addr string) (string, int, error) {
	i := strings.LastIndex(addr, ":")
	if i < 0 {
		return "", 0, fmt.Errorf("missing port")
	}
	host := addr[:i]
	port, err := strconv.Atoi(addr[i+1:])
	if err != nil {
		return "", 0, err
	}
	if host == "" {
		host = "127.0.0.1"
	}
	return host, port, nil
}
