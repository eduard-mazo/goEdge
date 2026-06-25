// Package dnp3 adapts the standalone goDnp3 DNP3 binding to the gateway's
// source.Source / source.Handler contract.
//
// The opendnp3 binding itself — the Master/Outstation implementations, the cgo
// shim, the dnp3_ffi build tag, and the vendored native library — lives in the
// goDnp3 module. This package is a thin, pure-Go adapter: it converts between
// goDnp3's neutral types (Measurement/Status/OutstationConfig/ServerConfig) and
// the gateway's source.Sample/source.Status and config.* types, so the
// publisher, API, and UI are unchanged.
//
// Build tags: goDnp3 selects its real opendnp3 binding under -tags dnp3_ffi and
// a pure-Go stub otherwise; building this module with that tag propagates it.
package dnp3

import (
	"context"

	godnp3 "goDnp3"

	"goMqttDnp3/config"
	"goMqttDnp3/source"
)

// Master abstracts the DNP3 master runtime: a source.Source plus DNP3-specific
// outstation management. It wraps goDnp3.Master.
type Master interface {
	// AddOutstation registers an outstation. Must be called before Start.
	AddOutstation(o config.DNP3Outstation) error

	// RemoveOutstation removes an outstation and tears down its channel.
	RemoveOutstation(id string) error

	// Start brings up all outstation channels and begins polling.
	Start(ctx context.Context) error

	// Stop tears down all channels and waits for in-flight callbacks.
	Stop()

	// Status returns the current per-outstation status snapshot.
	Status() []source.Status

	// IntegrityPoll triggers an on-demand integrity poll for one outstation.
	IntegrityPoll(outstationID string) error

	// OperateBinary / OperateAnalog issue a DirectOperate control to a field
	// outstation (Phase 8 control passthrough, DNP3 target). index is the point
	// index on that outstation.
	OperateBinary(outstationID string, index uint16, on bool) error
	OperateAnalog(outstationID string, index uint16, value float64) error
}

// Master is a source.Source plus DNP3-specific outstation management.
var _ source.Source = (Master)(nil)

// New constructs a Master backed by goDnp3.
func New(h source.Handler) Master {
	return &masterAdapter{lib: godnp3.NewMaster(handlerAdapter{h: h})}
}

type masterAdapter struct{ lib godnp3.Master }

func (m *masterAdapter) AddOutstation(o config.DNP3Outstation) error {
	return m.lib.AddOutstation(toOutstationConfig(o))
}
func (m *masterAdapter) RemoveOutstation(id string) error { return m.lib.RemoveOutstation(id) }
func (m *masterAdapter) Start(ctx context.Context) error  { return m.lib.Start(ctx) }
func (m *masterAdapter) Stop()                            { m.lib.Stop() }
func (m *masterAdapter) Status() []source.Status          { return toStatuses(m.lib.Status()) }
func (m *masterAdapter) IntegrityPoll(id string) error    { return m.lib.IntegrityPoll(id) }
func (m *masterAdapter) OperateBinary(id string, index uint16, on bool) error {
	return m.lib.OperateBinary(id, index, on)
}
func (m *masterAdapter) OperateAnalog(id string, index uint16, value float64) error {
	return m.lib.OperateAnalog(id, index, value)
}
