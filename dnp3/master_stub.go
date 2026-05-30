//go:build !dnp3_ffi

package dnp3

import (
	"context"
	"fmt"
	"sync"

	"goMqttDnp3/config"
	"goMqttDnp3/source"
)

// stubMaster is the default no-op implementation. It tracks registered
// outstations and reports them as disconnected without performing any I/O.
// Lets the rest of the gateway (UI, API, MQTT, Sparkplug) be built, run, and
// exercised end-to-end before the CGO binding is in place.
type stubMaster struct {
	h  source.Handler
	mu sync.Mutex
	os map[string]config.DNP3Outstation
}

func newMaster(h source.Handler) Master {
	return &stubMaster{
		h:  h,
		os: make(map[string]config.DNP3Outstation),
	}
}

func (m *stubMaster) AddOutstation(o config.DNP3Outstation) error {
	m.mu.Lock()
	m.os[o.ID] = o
	m.mu.Unlock()
	return nil
}

func (m *stubMaster) RemoveOutstation(id string) error {
	m.mu.Lock()
	delete(m.os, id)
	m.mu.Unlock()
	return nil
}

func (m *stubMaster) Start(_ context.Context) error {
	if m.h != nil {
		m.h.OnLog("warn", "DNP3 master running in STUB mode — no measurements will be received. Build with -tags dnp3_ffi after running `make opendnp3-vendor`.")
	}
	return nil
}

func (m *stubMaster) Stop() {}

func (m *stubMaster) Status() []source.Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]source.Status, 0, len(m.os))
	for _, o := range m.os {
		out = append(out, source.Status{
			ID:        o.ID,
			Label:     o.Label,
			Addr:      o.Addr(),
			Connected: false,
			LastError: "stub build (no DNP3 library)",
		})
	}
	return out
}

func (m *stubMaster) IntegrityPoll(id string) error {
	return fmt.Errorf("stub master: integrity poll not implemented (build with -tags dnp3_ffi)")
}
