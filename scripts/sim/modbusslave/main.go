// Command modbusslave is a standalone Modbus/TCP slave simulator for exercising
// the gateway's Modbus source — the counterpart to scripts/sim/outstation_sim
// (DNP3). Pure Go, no dependencies.
//
// It serves four banks (holding/input registers, coils, discrete inputs) and
// mutates them on a timer so the gateway sees changing values. The data layout
// matches config.modbussim.json.
//
//	holding  0-1 float32 tank level (m, ABCD)   2 uint16 pump speed (rpm)
//	         3   int16   temp ×10 (°C)         4-5 float32 flow (m³/h, ABCD)
//	input    0   uint16  pressure (kPa)
//	coils    0   pump running    1 valve open
//	discrete 0   fault
//
// Usage:  go run ./scripts/sim/modbusslave [port]   (default 1502)
package main

import (
	"encoding/binary"
	"io"
	"log"
	"math"
	"net"
	"os"
	"sync"
	"time"
)

type banks struct {
	mu       sync.Mutex
	holding  [256]uint16
	input    [256]uint16
	coils    [256]bool
	discrete [256]bool
}

func (b *banks) putF32(bank *[256]uint16, addr int, v float32) {
	bits := math.Float32bits(v)
	bank[addr] = uint16(bits >> 16) // high word first → ABCD on the wire
	bank[addr+1] = uint16(bits)
}

func (b *banks) tick(n int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.putF32(&b.holding, 0, 10.0+float32(n%50)*0.2)  // tank level 10..20 m sawtooth
	b.holding[2] = uint16(1450 + n%20)               // pump speed rpm
	b.holding[3] = uint16(int16(235 + (n%11 - 5)))   // temp ×10 → 23.0..24.0 °C
	b.putF32(&b.holding, 4, 3.0+float32(n%30)*0.1)   // flow m³/h
	b.input[0] = uint16(101 + n%8)                   // pressure kPa
	b.coils[0] = n%2 == 0                            // pump running
	b.coils[1] = (n/3)%2 == 0                        // valve open
	b.discrete[0] = n%17 == 0                        // fault (occasional)
}

func main() {
	port := "1502"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	ln, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatalf("modbusslave: bind :%s: %v", port, err)
	}
	b := &banks{}
	b.tick(0)
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		n := 1
		for range t.C {
			b.tick(n)
			n++
		}
	}()
	log.Printf("modbusslave: listening on 0.0.0.0:%s (unit id ignored; banks update every 1s)", port)

	for {
		c, err := ln.Accept()
		if err != nil {
			log.Fatalf("modbusslave: accept: %v", err)
		}
		go b.handle(c)
	}
}

func (b *banks) handle(c net.Conn) {
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
		resp := b.respond(pdu)

		out := make([]byte, 7+len(resp))
		copy(out[0:4], hdr[0:4]) // transaction id + protocol id (0)
		binary.BigEndian.PutUint16(out[4:6], uint16(1+len(resp)))
		out[6] = hdr[6] // unit id
		copy(out[7:], resp)
		if _, err := c.Write(out); err != nil {
			return
		}
	}
}

// respond builds the response PDU for a request PDU (read function codes only).
func (b *banks) respond(pdu []byte) []byte {
	if len(pdu) < 5 {
		return []byte{pdu[0] | 0x80, 0x03} // illegal data value
	}
	fc := pdu[0]
	addr := binary.BigEndian.Uint16(pdu[1:3])
	qty := binary.BigEndian.Uint16(pdu[3:5])

	b.mu.Lock()
	defer b.mu.Unlock()

	switch fc {
	case 0x03, 0x04: // read holding / input registers
		src := &b.holding
		if fc == 0x04 {
			src = &b.input
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
	case 0x01, 0x02: // read coils / discrete inputs
		src := &b.coils
		if fc == 0x02 {
			src = &b.discrete
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
