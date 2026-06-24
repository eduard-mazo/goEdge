package dnp3

import (
	godnp3 "goDnp3"

	"goMqttDnp3/config"
	"goMqttDnp3/source"
)

// handlerAdapter forwards goDnp3 callbacks to the gateway's source.Handler,
// converting goDnp3.Measurement/Status to source.Sample/source.Status. A value
// receiver keeps it cheap to pass to goDnp3.NewMaster/NewOutstation.
type handlerAdapter struct{ h source.Handler }

func (a handlerAdapter) OnMeasurement(m godnp3.Measurement) { a.h.OnSample(toSample(m)) }
func (a handlerAdapter) OnStatusChange(s godnp3.Status)     { a.h.OnStatusChange(toStatus(s)) }
func (a handlerAdapter) OnLog(level, msg string)            { a.h.OnLog(level, msg) }

// toSample converts a goDnp3 measurement to the gateway's source.Sample. The
// flag/string values are identical across the two packages, so the typed
// conversions are plain casts.
func toSample(m godnp3.Measurement) source.Sample {
	return source.Sample{
		SourceID:   m.OutstationID,
		PointType:  source.PointType(m.PointType),
		Index:      m.Index,
		Time:       m.Time,
		Quality:    source.Quality(m.Quality),
		IsEvent:    m.IsEvent,
		BoolValue:  m.BoolValue,
		DBBValue:   source.DoubleBitState(m.DBBValue),
		UintValue:  m.UintValue,
		FloatValue: m.FloatValue,
		BytesValue: m.BytesValue,
	}
}

// toMeasurement is the reverse of toSample (used when the publisher pushes a
// value into the outstation server).
func toMeasurement(s source.Sample) godnp3.Measurement {
	return godnp3.Measurement{
		OutstationID: s.SourceID,
		PointType:    godnp3.PointType(s.PointType),
		Index:        s.Index,
		Time:         s.Time,
		Quality:      godnp3.Quality(s.Quality),
		IsEvent:      s.IsEvent,
		BoolValue:    s.BoolValue,
		DBBValue:     godnp3.DoubleBitState(s.DBBValue),
		UintValue:    s.UintValue,
		FloatValue:   s.FloatValue,
		BytesValue:   s.BytesValue,
	}
}

func toStatus(s godnp3.Status) source.Status {
	return source.Status{
		ID:              s.ID,
		Label:           s.Label,
		Addr:            s.Addr,
		Connected:       s.Connected,
		LastError:       s.LastError,
		MeasurementsRx:  s.MeasurementsRx,
		IntegrityPolls:  s.IntegrityPolls,
		ClassPolls:      s.ClassPolls,
		UnsolicitedRsps: s.UnsolicitedRsps,
		LastReadAt:      s.LastReadAt,
	}
}

func toStatuses(ss []godnp3.Status) []source.Status {
	out := make([]source.Status, len(ss))
	for i, s := range ss {
		out[i] = toStatus(s)
	}
	return out
}

func toOutstationConfig(o config.DNP3Outstation) godnp3.OutstationConfig {
	return godnp3.OutstationConfig{
		ID:                    o.ID,
		Label:                 o.Label,
		Host:                  o.Host,
		Port:                  o.Port,
		MasterAddress:         o.MasterAddress,
		OutstationAddress:     o.OutstationAddress,
		ResponseTimeoutMs:     o.ResponseTimeoutMs,
		KeepAliveMs:           o.KeepAliveMs,
		IntegrityScanMs:       o.IntegrityScanMs,
		Class1ScanMs:          o.Class1ScanMs,
		Class2ScanMs:          o.Class2ScanMs,
		Class3ScanMs:          o.Class3ScanMs,
		UnsolicitedEnabled:    o.UnsolicitedEnabled,
		UnsolicitedClass1:     o.UnsolicitedClass1,
		UnsolicitedClass2:     o.UnsolicitedClass2,
		UnsolicitedClass3:     o.UnsolicitedClass3,
		DisableUnsolOnStartup: o.DisableUnsolOnStartup,
		StartupIntegrity:      o.StartupIntegrity,
		IntegrityClass0:       o.IntegrityClass0,
		IntegrityClass1:       o.IntegrityClass1,
		IntegrityClass2:       o.IntegrityClass2,
		IntegrityClass3:       o.IntegrityClass3,
		StaticPollMs:          o.StaticPollMs,
	}
}

func toServerConfig(s config.DNP3OutstationServer) godnp3.ServerConfig {
	return godnp3.ServerConfig{
		ID:               s.ID,
		Label:            s.Label,
		BindHost:         s.BindHost,
		Port:             s.Port,
		LocalAddress:     s.LocalAddress,
		MasterAddress:    s.MasterAddress,
		AllowUnsolicited: s.AllowUnsolicited,
		EventBufferSize:  s.EventBufferSize,
	}
}

func toDBSizes(s OutstationDBSizes) godnp3.DBSizes {
	return godnp3.DBSizes{
		Binary:             s.Binary,
		DoubleBit:          s.DoubleBit,
		Analog:             s.Analog,
		Counter:            s.Counter,
		FrozenCounter:      s.FrozenCounter,
		BinaryOutputStatus: s.BinaryOutputStatus,
		AnalogOutputStatus: s.AnalogOutputStatus,
		OctetString:        s.OctetString,
	}
}
