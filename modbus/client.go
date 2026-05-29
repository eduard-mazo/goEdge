// Package modbus provides a pooled, reconnecting Modbus TCP client.
package modbus

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	mb "github.com/goburrow/modbus"
)

// deviceConn wraps a single Modbus TCP handler+client behind a mutex.
// One connection per (host:port); all reads to the same device serialize here.
type deviceConn struct {
	mu      sync.Mutex
	handler *mb.TCPClientHandler
	client  mb.Client
	addr    string
	alive   bool
}

func newDeviceConn(addr string, timeoutMs, retries int) *deviceConn {
	h := mb.NewTCPClientHandler(addr)
	h.Timeout = time.Duration(timeoutMs) * time.Millisecond
	h.IdleTimeout = 30 * time.Second
	h.Logger = nil // suppress library's own logger; we log through slog
	return &deviceConn{
		handler: h,
		client:  mb.NewClient(h),
		addr:    addr,
	}
}

func (dc *deviceConn) connect() error {
	if dc.alive {
		return nil
	}
	if err := dc.handler.Connect(); err != nil {
		return fmt.Errorf("modbus connect %s: %w", dc.addr, err)
	}
	dc.alive = true
	return nil
}

func (dc *deviceConn) close() {
	dc.mu.Lock()
	defer dc.mu.Unlock()
	_ = dc.handler.Close()
	dc.alive = false
}

// Pool manages one *deviceConn per address, creating lazily and reconnecting on failure.
type Pool struct {
	mu      sync.Mutex
	conns   map[string]*deviceConn
	timeout int // ms
	retries int
}

// NewPool creates a connection pool.
func NewPool(timeoutMs, retries int) *Pool {
	if timeoutMs <= 0 {
		timeoutMs = 3000
	}
	if retries < 0 {
		retries = 2
	}
	return &Pool{
		conns:   make(map[string]*deviceConn),
		timeout: timeoutMs,
		retries: retries,
	}
}

func (p *Pool) get(addr string, timeoutMs, retries int) *deviceConn {
	p.mu.Lock()
	defer p.mu.Unlock()
	if dc, ok := p.conns[addr]; ok {
		return dc
	}
	dc := newDeviceConn(addr, timeoutMs, retries)
	p.conns[addr] = dc
	return dc
}

// Read performs a Modbus read with the given function code, retrying on transient errors.
// Returns raw bytes in Modbus big-endian wire order.
func (p *Pool) Read(ctx context.Context, addr string, unitID uint8, fn ReadFunc, regAddr, qty uint16, timeoutMs, retries, retryDelayMs int) ([]byte, error) {
	dc := p.get(addr, timeoutMs, retries)
	dc.mu.Lock()
	defer dc.mu.Unlock()

	var lastErr error
	for attempt := 0; attempt <= retries; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if err := dc.connect(); err != nil {
			lastErr = err
			dc.alive = false
			sleep(ctx, retryDelayMs)
			continue
		}
		dc.handler.SlaveId = unitID
		data, err := fn(dc.client, regAddr, qty)
		if err == nil {
			return data, nil
		}
		lastErr = fmt.Errorf("modbus read unit=%d addr=%d qty=%d: %w", unitID, regAddr, qty, err)
		slog.Warn("modbus read error", "addr", addr, "attempt", attempt, "err", err)
		dc.alive = false
		_ = dc.handler.Close()
		sleep(ctx, retryDelayMs)
	}
	return nil, lastErr
}

// CloseAll closes all pooled connections.
func (p *Pool) CloseAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, dc := range p.conns {
		_ = dc.handler.Close()
		dc.alive = false
	}
	p.conns = make(map[string]*deviceConn)
}

// ReadFunc abstracts the four Modbus read operations.
type ReadFunc func(client mb.Client, address, quantity uint16) ([]byte, error)

// ReadCoils is a ReadFunc for FC01.
func ReadCoils(client mb.Client, address, quantity uint16) ([]byte, error) {
	return client.ReadCoils(address, quantity)
}

// ReadDiscreteInputs is a ReadFunc for FC02.
func ReadDiscreteInputs(client mb.Client, address, quantity uint16) ([]byte, error) {
	return client.ReadDiscreteInputs(address, quantity)
}

// ReadInputRegisters is a ReadFunc for FC04.
func ReadInputRegisters(client mb.Client, address, quantity uint16) ([]byte, error) {
	return client.ReadInputRegisters(address, quantity)
}

// ReadHoldingRegisters is a ReadFunc for FC03.
func ReadHoldingRegisters(client mb.Client, address, quantity uint16) ([]byte, error) {
	return client.ReadHoldingRegisters(address, quantity)
}

func sleep(ctx context.Context, ms int) {
	if ms <= 0 {
		ms = 500
	}
	select {
	case <-time.After(time.Duration(ms) * time.Millisecond):
	case <-ctx.Done():
	}
}
