//go:build !dnp3_ffi

package dnp3

import (
	"context"

	"goMqttDnp3/config"
	"goMqttDnp3/source"
)

// stubOutstation is the default no-op implementation. It lets the gateway (UI,
// API, MQTT, config) be built and run without the opendnp3 library; no DNP3
// server is actually bound, so a SCADA master cannot connect.
type stubOutstation struct {
	cfg config.DNP3OutstationServer
	h   source.Handler
}

func newOutstation(cfg config.DNP3OutstationServer, _ OutstationDBSizes, h source.Handler) Outstation {
	return &stubOutstation{cfg: cfg, h: h}
}

func (o *stubOutstation) Start(_ context.Context) error {
	if o.h != nil {
		o.h.OnLog("warn", "DNP3 outstation server running in STUB mode — no SCADA master can connect. Build with -tags dnp3_ffi after running `make opendnp3-vendor`.")
	}
	return nil
}

func (o *stubOutstation) Stop() {}

func (o *stubOutstation) Update(source.Sample) {}

func (o *stubOutstation) Status() source.Status {
	return source.Status{
		ID:        o.cfg.ServerID(),
		Label:     o.cfg.Label,
		Addr:      o.cfg.Addr(),
		Connected: false,
		LastError: "stub build (no DNP3 library)",
	}
}
