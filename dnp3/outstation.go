package dnp3

import (
	"context"

	"goMqttDnp3/config"
	"goMqttDnp3/source"
)

// Outstation is the DNP3 outstation server: the gateway serves aggregated field
// data northbound to a SCADA master over DNP3.
//
// Unlike Master, an Outstation is a sink, not a source.Source — the publisher
// pushes values into it via Update as samples flow in from the field, rather
// than the outstation pushing samples to a Handler. The Handler it is given is
// used only for status changes (SCADA connect/disconnect) and log lines.
//
// Lifecycle: NewOutstation → Start (binds the TCP server, begins serving) →
// Update… (any goroutine) → Stop (tears down the server channel).
//
// Implementations:
//   - stubOutstation in outstation_stub.go (default; no C deps)
//   - ffiOutstation in outstation_ffi.go  (build tag dnp3_ffi; wraps opendnp3)
type Outstation interface {
	// Start brings up the server channel and enables the outstation.
	Start(ctx context.Context) error

	// Stop tears down the server channel and releases the shared manager ref.
	Stop()

	// Update pushes one measurement into the served database. The sample's
	// PointType/Index address a point in the outstation database (not the field
	// source), and the value carried per PointType is served as-is — callers pass
	// the engineering-scaled value and the quality they want SCADA to see. Safe to
	// call from any goroutine; a no-op before Start / after Stop.
	Update(s source.Sample)

	// Status returns the current server status snapshot (one server, so a single
	// Status rather than the Master's slice). Connected reflects whether a SCADA
	// master is currently connected.
	Status() source.Status
}

// OutstationDBSizes is the per-type point count used to size the outstation
// database. The database is built with contiguous indices [0, count) for each
// type; an Update to an out-of-range index is dropped. The publisher computes
// these from the ServeDNP3 mappings (max index per type + 1).
//
// FrozenCounter is sized for completeness but has no direct setter (opendnp3
// derives frozen counters by freezing a counter); a field frozen-counter is
// served as a plain counter point instead.
type OutstationDBSizes struct {
	Binary             uint16
	DoubleBit          uint16
	Analog             uint16
	Counter            uint16
	FrozenCounter      uint16
	BinaryOutputStatus uint16
	AnalogOutputStatus uint16
	OctetString        uint16
}

// NewOutstation constructs an Outstation. The concrete type is selected by build
// tag (see newOutstation in outstation_{stub,ffi}.go).
func NewOutstation(cfg config.DNP3OutstationServer, sizes OutstationDBSizes, h source.Handler) Outstation {
	return newOutstation(cfg, sizes, h)
}
