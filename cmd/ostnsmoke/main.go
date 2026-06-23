//go:build dnp3_ffi

// Command ostnsmoke is a runtime smoke test for the opendnp3 OUTSTATION binding
// (the gateway acting as a DNP3 outstation server). It brings up the gateway's
// outstation, pushes a known set of point values into it, then drives it with
// the gateway's OWN DNP3 master over loopback and verifies the master reads back
// exactly what was served — including a fractional analog (proving analogs are
// served as float, not the opendnp3 integer default) and a bad-quality point.
//
// Both roles run in one process, so this also exercises the shared, refcounted
// DNP3Manager (master + outstation alive together).
//
// Build:  go build -tags dnp3_ffi ./cmd/ostnsmoke   (with the opendnp3 cgo flags)
// Run:    ostnsmoke -port 20100 -duration 8s
//
// Exit 0 only if the master connected AND read back every served point with the
// expected value and quality; non-zero otherwise (so it's CI-usable).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"goMqttDnp3/config"
	"goMqttDnp3/dnp3"
	"goMqttDnp3/source"
)

// served is one point the gateway exposes on its outstation, with the value the
// master is expected to read back.
type served struct {
	pt  source.PointType
	idx uint16
	q   source.Quality
	b   bool
	f   float64
	u   uint32
	dbb source.DoubleBitState
}

func (sv served) sample() source.Sample {
	s := source.Sample{PointType: sv.pt, Index: sv.idx, Time: time.Now(), Quality: sv.q}
	switch sv.pt {
	case source.PointBinary, source.PointBinaryOutputStatus:
		s.BoolValue = sv.b
	case source.PointDoubleBitBinary:
		s.DBBValue = sv.dbb
	case source.PointCounter, source.PointFrozenCounter:
		s.UintValue = sv.u
	case source.PointAnalog, source.PointAnalogOutputStatus:
		s.FloatValue = sv.f
	}
	return s
}

const good = source.QualityOnline

// expected is the final value-set the master must read back. The fractional
// analogs (123.5, -7.25) are exactly representable and would NOT survive a 32-bit
// integer wire variation — so reading them back exactly proves the float fix.
// Binary idx 1 is served bad (quality 0) to prove bad quality propagates.
var expected = []served{
	{pt: source.PointBinary, idx: 0, q: good, b: true},
	{pt: source.PointBinary, idx: 1, q: 0, b: false},
	{pt: source.PointAnalog, idx: 0, q: good, f: 123.5},
	{pt: source.PointAnalog, idx: 1, q: good, f: -7.25},
	{pt: source.PointCounter, idx: 0, q: good, u: 4242},
	{pt: source.PointDoubleBitBinary, idx: 0, q: good, dbb: source.DBBOn},
}

// initial is pushed before the master connects, with different values, so the
// post-connect re-push + integrity poll proves the live Update path changes them.
var initial = []served{
	{pt: source.PointBinary, idx: 0, q: good, b: false},
	{pt: source.PointBinary, idx: 1, q: good, b: true},
	{pt: source.PointAnalog, idx: 0, q: good, f: 1},
	{pt: source.PointAnalog, idx: 1, q: good, f: 2},
	{pt: source.PointCounter, idx: 0, q: good, u: 1},
	{pt: source.PointDoubleBitBinary, idx: 0, q: good, dbb: source.DBBOff},
}

type key struct {
	pt  source.PointType
	idx uint16
}

// recorder is the master-side Handler; it keeps the latest sample per point.
type recorder struct {
	mu        sync.Mutex
	latest    map[key]source.Sample
	connected bool
}

func newRecorder() *recorder { return &recorder{latest: make(map[key]source.Sample)} }

func (r *recorder) OnSample(m source.Sample) {
	r.mu.Lock()
	r.latest[key{m.PointType, m.Index}] = m
	r.mu.Unlock()
}

func (r *recorder) OnStatusChange(s source.Status) {
	r.mu.Lock()
	if s.Connected {
		r.connected = true
	}
	r.mu.Unlock()
	log.Printf("[master] status connected=%v err=%q rx=%d", s.Connected, s.LastError, s.MeasurementsRx)
}

func (r *recorder) OnLog(level, msg string) { log.Printf("[master:%s] %s", level, msg) }

// osHandler is the outstation-side Handler (status/log only; never OnSample).
type osHandler struct{}

func (osHandler) OnSample(source.Sample) {}
func (osHandler) OnStatusChange(s source.Status) {
	log.Printf("[ostn] status connected=%v err=%q served=%d", s.Connected, s.LastError, s.MeasurementsRx)
}
func (osHandler) OnLog(level, msg string) { log.Printf("[ostn:%s] %s", level, msg) }

func main() {
	port := flag.Int("port", 20100, "loopback TCP port for the outstation server")
	duration := flag.Duration("duration", 8*time.Second, "total test duration")
	flag.Parse()

	// 1. Bring up the gateway's outstation server.
	sizes := dnp3.OutstationDBSizes{Binary: 2, Analog: 2, Counter: 1, DoubleBit: 1}
	srv := config.DNP3OutstationServer{
		Enabled:         true,
		ID:              "gw-ostn",
		Label:           "Gateway Outstation",
		BindHost:        "127.0.0.1",
		Port:            *port,
		LocalAddress:    1024,
		MasterAddress:   1,
		EventBufferSize: 100,
	}
	ostn := dnp3.NewOutstation(srv, sizes, osHandler{})
	if err := ostn.Start(context.Background()); err != nil {
		log.Fatalf("outstation Start: %v", err)
	}
	defer ostn.Stop()
	log.Printf("outstation serving on 127.0.0.1:%d (addr 1024, master 1)", *port)

	// 2. Seed initial values, then start the master pointed at the outstation.
	for _, sv := range initial {
		ostn.Update(sv.sample())
	}

	rec := newRecorder()
	m := dnp3.New(rec)
	if err := m.AddOutstation(config.DNP3Outstation{
		ID:                "gw",
		Label:             "Gateway (loopback)",
		Host:              "127.0.0.1",
		Port:              *port,
		MasterAddress:     1,
		OutstationAddress: 1024,
		ResponseTimeoutMs: 3000,
		KeepAliveMs:       60000,
		Class1ScanMs:      1000,
		StartupIntegrity:  true,
		Enabled:           true,
	}); err != nil {
		log.Fatalf("AddOutstation: %v", err)
	}
	if err := m.Start(context.Background()); err != nil {
		log.Fatalf("master Start: %v", err)
	}

	// 3. Halfway through, push the final values and re-read with an integrity poll.
	half := *duration / 2
	time.Sleep(half)
	log.Printf("pushing final values + on-demand integrity poll")
	for _, sv := range expected {
		ostn.Update(sv.sample())
	}
	if err := m.IntegrityPoll("gw"); err != nil {
		log.Printf("IntegrityPoll error: %v", err)
	}
	time.Sleep(*duration - half)

	m.Stop()

	// 4. Verify.
	rec.mu.Lock()
	defer rec.mu.Unlock()

	fmt.Println("\n──────── outstation smoke summary ────────")
	fmt.Printf("master connected: %v\n", rec.connected)
	fails := 0
	for _, sv := range expected {
		got, ok := rec.latest[key{sv.pt, sv.idx}]
		if !ok {
			fmt.Printf("  MISS  %-20s idx=%d (never read back)\n", sv.pt, sv.idx)
			fails++
			continue
		}
		valOK := valueMatches(sv, got)
		qOK := got.Quality.Good() == (sv.q == good)
		status := "ok"
		if !valOK || !qOK {
			status = "FAIL"
			fails++
		}
		fmt.Printf("  %-4s  %-20s idx=%d want(%s) got(%s) q=0x%02x good=%v\n",
			status, sv.pt, sv.idx, wantStr(sv), gotStr(got), uint8(got.Quality), got.Quality.Good())
	}
	fmt.Println("──────────────────────────────────────────")

	if !rec.connected {
		fmt.Println("FAIL: master never connected to the gateway outstation")
		os.Exit(1)
	}
	if fails > 0 {
		fmt.Printf("FAIL: %d point(s) mismatched\n", fails)
		os.Exit(1)
	}
	fmt.Println("PASS")
}

func valueMatches(sv served, got source.Sample) bool {
	switch sv.pt {
	case source.PointBinary, source.PointBinaryOutputStatus:
		return got.BoolValue == sv.b
	case source.PointDoubleBitBinary:
		return got.DBBValue == sv.dbb
	case source.PointCounter, source.PointFrozenCounter:
		return got.UintValue == sv.u
	case source.PointAnalog, source.PointAnalogOutputStatus:
		return math.Abs(got.FloatValue-sv.f) < 1e-9
	}
	return false
}

func wantStr(sv served) string {
	switch sv.pt {
	case source.PointBinary, source.PointBinaryOutputStatus:
		return fmt.Sprintf("%v", sv.b)
	case source.PointDoubleBitBinary:
		return fmt.Sprintf("%d", sv.dbb)
	case source.PointCounter, source.PointFrozenCounter:
		return fmt.Sprintf("%d", sv.u)
	case source.PointAnalog, source.PointAnalogOutputStatus:
		return fmt.Sprintf("%g", sv.f)
	}
	return "?"
}

func gotStr(m source.Sample) string {
	switch m.PointType {
	case source.PointBinary, source.PointBinaryOutputStatus:
		return fmt.Sprintf("%v", m.BoolValue)
	case source.PointDoubleBitBinary:
		return fmt.Sprintf("%d", m.DBBValue)
	case source.PointCounter, source.PointFrozenCounter:
		return fmt.Sprintf("%d", m.UintValue)
	case source.PointAnalog, source.PointAnalogOutputStatus:
		return fmt.Sprintf("%g", m.FloatValue)
	}
	return "?"
}
