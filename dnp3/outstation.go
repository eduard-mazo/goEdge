package dnp3

import (
	"context"

	godnp3 "goDnp3"

	"goMqttDnp3/config"
	"goMqttDnp3/source"
)

// Outstation is the DNP3 outstation server: a sink the publisher pushes mapped
// values into via Update so a SCADA master can poll them. It wraps
// goDnp3.Outstation. The Handler is used only for status/log callbacks.
type Outstation interface {
	Start(ctx context.Context) error
	Stop()
	Update(s source.Sample)
	Status() source.Status

	// SetCommandSink registers a sink for controls a SCADA master issues (Phase 8
	// control passthrough). Must be called before Start; with no sink, controls
	// are rejected (NOT_SUPPORTED).
	SetCommandSink(sink CommandSink)
}

// ControlStatus is the result a CommandSink returns for a control.
type ControlStatus int

const (
	CtrlAccepted     ControlStatus = 0
	CtrlNotSupported ControlStatus = 4
)

// CommandSink receives controls a SCADA master issues to the gateway's
// outstation (a CROB becomes on/off; an analog output carries a value). The
// gateway maps these to field writes. Return CtrlAccepted to accept or
// CtrlNotSupported to reject. Called from a DNP3 thread — must not block.
type CommandSink interface {
	OnControlBinary(index uint16, on bool, isSelect bool) ControlStatus
	OnControlAnalog(index uint16, value float64, isSelect bool) ControlStatus
}

// OutstationDBSizes is the per-type point count used to size the served
// database. The publisher computes it from the ServeDNP3 mappings (max served
// index per type + 1).
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

// NewOutstation constructs an Outstation backed by goDnp3.
func NewOutstation(cfg config.DNP3OutstationServer, sizes OutstationDBSizes, h source.Handler) Outstation {
	return &outstationAdapter{
		lib: godnp3.NewOutstation(toServerConfig(cfg), toDBSizes(sizes), handlerAdapter{h: h}),
	}
}

type outstationAdapter struct{ lib godnp3.Outstation }

func (o *outstationAdapter) Start(ctx context.Context) error { return o.lib.Start(ctx) }
func (o *outstationAdapter) Stop()                           { o.lib.Stop() }
func (o *outstationAdapter) Update(s source.Sample)          { o.lib.Update(toMeasurement(s)) }
func (o *outstationAdapter) Status() source.Status           { return toStatus(o.lib.Status()) }
func (o *outstationAdapter) SetCommandSink(sink CommandSink) {
	o.lib.SetCommandSink(commandSinkAdapter{sink: sink})
}
