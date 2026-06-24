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
