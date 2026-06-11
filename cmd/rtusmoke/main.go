// Command rtusmoke is a self-contained runtime test for the Modbus RTU
// (RS-485) source adapter. It allocates a Linux PTY pair, runs a minimal
// in-process Modbus RTU slave on the master side, points a modbusrtu.Poller at
// the slave device node, and verifies that decoded samples (float32, uint16,
// coil) flow through with the right values over real serial framing.
//
// No hardware, no cgo, no broker. Linux only (uses /dev/ptmx).
//
// Run:  go run ./cmd/rtusmoke   (exit 0 = PASS)
package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"golang.org/x/sys/unix"

	"goMqttDnp3/config"
	"goMqttDnp3/modbusrtu"
	"goMqttDnp3/source"
)

const slaveUnitID = 1

// --- minimal Modbus RTU slave (master side of the PTY) -----------------------

type slave struct {
	mu    sync.Mutex
	regs  [256]uint16
	coils [256]bool
}

func (s *slave) tick(level *float32, count *uint16, coil *bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	bits := math.Float32bits(*level)
	s.regs[0] = uint16(bits >> 16) // HR0 = high word (ABCD big-endian on the wire)
	s.regs[1] = uint16(bits)       // HR1 = low word
	s.regs[2] = *count             // HR2 = uint16 counter
	s.coils[0] = *coil
	*level += 0.5
	*count++
	*coil = !*coil
}

// serve reads fixed 8-byte RTU read requests from the master fd and replies.
// Only the read functions the poller issues (0x01 coils, 0x03 holding, 0x04
// input) are implemented — enough to exercise the decode path end to end.
func (s *slave) serve(f *os.File) {
	req := make([]byte, 8)
	for {
		if err := readFull(f, req); err != nil {
			return // master closed
		}
		if crc16(req[:6]) != binary.LittleEndian.Uint16(req[6:8]) {
			continue // framing/CRC error — ignore, real master would time out
		}
		addr, fn := req[0], req[1]
		start := binary.BigEndian.Uint16(req[2:4])
		qty := binary.BigEndian.Uint16(req[4:6])
		if addr != slaveUnitID {
			continue
		}
		var resp []byte
		switch fn {
		case 0x01: // read coils
			nbytes := int((qty + 7) / 8)
			body := make([]byte, nbytes)
			s.mu.Lock()
			for i := 0; i < int(qty); i++ {
				if s.coils[int(start)+i] {
					body[i/8] |= 1 << (i % 8)
				}
			}
			s.mu.Unlock()
			resp = append([]byte{addr, fn, byte(nbytes)}, body...)
		case 0x03, 0x04: // read holding / input registers
			body := make([]byte, int(qty)*2)
			s.mu.Lock()
			for i := 0; i < int(qty); i++ {
				binary.BigEndian.PutUint16(body[i*2:], s.regs[int(start)+i])
			}
			s.mu.Unlock()
			resp = append([]byte{addr, fn, byte(len(body))}, body...)
		default:
			continue
		}
		resp = binary.LittleEndian.AppendUint16(resp, crc16(resp))
		if _, err := f.Write(resp); err != nil {
			return
		}
	}
}

func readFull(f *os.File, buf []byte) error {
	got := 0
	for got < len(buf) {
		n, err := f.Read(buf[got:])
		if err != nil {
			return err
		}
		got += n
	}
	return nil
}

// crc16 computes the Modbus RTU CRC-16 (poly 0xA001, init 0xFFFF).
func crc16(data []byte) uint16 {
	crc := uint16(0xFFFF)
	for _, b := range data {
		crc ^= uint16(b)
		for range 8 {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}
	return crc
}

// --- sample recorder (source.Handler) ----------------------------------------

type recorder struct {
	mu   sync.Mutex
	last map[string]source.Sample // keyed by "pointType@index"
}

func newRecorder() *recorder { return &recorder{last: map[string]source.Sample{}} }

func (r *recorder) OnSample(s source.Sample) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.last[fmt.Sprintf("%s@%d", s.PointType, s.Index)] = s
}
func (r *recorder) OnStatusChange(s source.Status) {
	log.Printf("status: %s connected=%v err=%q", s.ID, s.Connected, s.LastError)
}
func (r *recorder) OnLog(level, msg string) { log.Printf("[%s] %s", level, msg) }

func (r *recorder) get(key string) (source.Sample, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.last[key]
	return s, ok
}

// --- PTY allocation -----------------------------------------------------------

// openPTY returns the master *os.File and the slave device path (/dev/pts/N).
func openPTY() (*os.File, string, error) {
	master, err := unix.Open("/dev/ptmx", unix.O_RDWR|unix.O_NOCTTY, 0)
	if err != nil {
		return nil, "", fmt.Errorf("open /dev/ptmx: %w", err)
	}
	if err := unix.IoctlSetPointerInt(master, unix.TIOCSPTLCK, 0); err != nil { // unlockpt
		unix.Close(master)
		return nil, "", fmt.Errorf("unlockpt: %w", err)
	}
	ptn, err := unix.IoctlGetInt(master, unix.TIOCGPTN)
	if err != nil {
		unix.Close(master)
		return nil, "", fmt.Errorf("get pty number: %w", err)
	}
	return os.NewFile(uintptr(master), "ptmx"), fmt.Sprintf("/dev/pts/%d", ptn), nil
}

// --- driver -------------------------------------------------------------------

func main() {
	master, slavePath, err := openPTY()
	if err != nil {
		log.Fatalf("PTY: %v", err)
	}
	defer master.Close()
	log.Printf("RTU slave on PTY master; poller will open %s", slavePath)

	sl := &slave{}
	level := float32(10.0)
	count := uint16(100)
	coil := true
	sl.tick(&level, &count, &coil)
	go sl.serve(master)

	// Keep the slave registers advancing so we can see fresh reads.
	stopTick := make(chan struct{})
	go func() {
		t := time.NewTicker(100 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-stopTick:
				return
			case <-t.C:
				sl.tick(&level, &count, &coil)
			}
		}
	}()

	dev := config.SerialDevice{
		ID:         "rtu-1",
		Label:      "smoke slave",
		Port:       slavePath,
		BaudRate:   19200,
		UnitID:     slaveUnitID,
		ScanRateMs: 100,
		TimeoutMs:  500,
		Retries:    1,
		Enabled:    true,
	}
	mappings := []config.SignalMapping{
		{ID: "m1", MetricName: "Level", Protocol: "modbusrtu", SourceID: "rtu-1",
			Function: "holding_register", Address: 0, DataType: "float32", ByteOrder: "ABCD", Enabled: true},
		{ID: "m2", MetricName: "Count", Protocol: "modbusrtu", SourceID: "rtu-1",
			Function: "holding_register", Address: 2, DataType: "uint16", Enabled: true},
		{ID: "m3", MetricName: "Run", Protocol: "modbusrtu", SourceID: "rtu-1",
			Function: "coil", Address: 0, Enabled: true},
	}

	rec := newRecorder()
	p := modbusrtu.New(rec)
	if err := p.AddDevice(dev, mappings); err != nil {
		log.Fatalf("AddDevice: %v", err)
	}
	ctx := context.Background()
	if err := p.Start(ctx); err != nil {
		log.Fatalf("Start: %v", err)
	}

	// Give the poller a few scan cycles to read all three points, then snapshot
	// status while still running (Stop cancels in-flight reads → momentarily
	// disconnected, which is expected).
	time.Sleep(1 * time.Second)
	st := p.Status()
	p.Stop()
	close(stopTick)

	ok := true
	if s, found := rec.get("holding_register@0"); !found || s.FloatValue < 10.0 {
		log.Printf("FAIL float32 holding@0: found=%v val=%v", found, s.FloatValue)
		ok = false
	} else {
		log.Printf("PASS float32 holding@0 = %v", s.FloatValue)
	}
	if s, found := rec.get("holding_register@2"); !found || s.FloatValue < 100 {
		log.Printf("FAIL uint16 holding@2: found=%v val=%v", found, s.FloatValue)
		ok = false
	} else {
		log.Printf("PASS uint16 holding@2 = %v", s.FloatValue)
	}
	if s, found := rec.get("coil@0"); !found {
		log.Printf("FAIL coil@0 not received")
		ok = false
	} else {
		log.Printf("PASS coil@0 = %v", s.BoolValue)
	}

	if len(st) != 1 || !st[0].Connected || st[0].MeasurementsRx == 0 {
		log.Printf("FAIL status: %+v", st)
		ok = false
	} else {
		log.Printf("PASS status: connected, %d measurements", st[0].MeasurementsRx)
	}

	if !ok {
		os.Exit(1)
	}
	log.Printf("ALL PASS")
}
