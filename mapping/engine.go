// Package mapping converts DNP3 measurements to Sparkplug B metrics.
package mapping

import (
	"fmt"
	"math"

	"goMqttDnp3/config"
	"goMqttDnp3/dnp3"
	"goMqttDnp3/sparkplug"
)

// Result holds the engineering-scaled value and the Sparkplug metric.
type Result struct {
	Value  float64
	IsNull bool
	Metric *sparkplug.Metric
}

// Apply turns a DNP3 Measurement into a Sparkplug Metric, applying scale/offset
// and attaching engineering unit + DNP3 quality flags as metric properties.
//
// Sparkplug timestamps are taken from the measurement (the outstation's clock,
// already plausibility-checked by the master), not from the gateway clock.
func Apply(sig config.SignalMapping, m dnp3.Measurement) (Result, error) {
	if string(m.PointType) != sig.PointType {
		return Result{}, fmt.Errorf("mapping %q: point type mismatch (mapping=%s, measurement=%s)",
			sig.MetricName, sig.PointType, m.PointType)
	}
	tsMs := uint64(m.Time.UnixMilli())
	scale := sig.Scale
	if scale == 0 {
		scale = 1.0
	}

	var metric *sparkplug.Metric
	var value float64 = math.NaN()

	switch m.PointType {
	case dnp3.PointBinary, dnp3.PointBinaryOutputStatus:
		metric = sparkplug.MetricBool(sig.MetricName, tsMs, m.BoolValue)
		if m.BoolValue {
			value = 1
		} else {
			value = 0
		}

	case dnp3.PointDoubleBitBinary:
		v := uint32(m.DBBValue)
		metric = sparkplug.MetricUInt32(sig.MetricName, tsMs, v)
		value = float64(v)

	case dnp3.PointCounter, dnp3.PointFrozenCounter:
		eng := float64(m.UintValue)*scale + sig.Offset
		if scale == 1.0 && sig.Offset == 0 {
			metric = sparkplug.MetricUInt32(sig.MetricName, tsMs, m.UintValue)
		} else {
			metric = sparkplug.MetricDouble(sig.MetricName, tsMs, eng)
		}
		value = eng

	case dnp3.PointAnalog, dnp3.PointAnalogOutputStatus:
		eng := m.FloatValue*scale + sig.Offset
		metric = sparkplug.MetricDouble(sig.MetricName, tsMs, eng)
		value = eng

	case dnp3.PointOctetString:
		metric = sparkplug.MetricString(sig.MetricName, tsMs, string(m.BytesValue))

	default:
		return Result{}, fmt.Errorf("mapping %q: unsupported point type %q", sig.MetricName, m.PointType)
	}

	metric.Properties = qualityProperties(m.Quality, sig.EngineeringUnit)
	return Result{Value: value, Metric: metric, IsNull: !m.Quality.Good()}, nil
}

// qualityProperties packs DNP3 flags into a Sparkplug PropertySet so the
// receiving SCADA can interpret point quality without out-of-band knowledge.
func qualityProperties(q dnp3.Quality, engUnit string) *sparkplug.PropertySet {
	ps := &sparkplug.PropertySet{}
	addUint := func(k string, v uint32) {
		val := v
		ps.Keys = append(ps.Keys, k)
		ps.Values = append(ps.Values, &sparkplug.PropertyValue{
			Type: sparkplug.DataTypeUInt32, IntValue: &val,
		})
	}
	addBool := func(k string, v bool) {
		val := v
		ps.Keys = append(ps.Keys, k)
		ps.Values = append(ps.Values, &sparkplug.PropertyValue{
			Type: sparkplug.DataTypeBoolean, BoolValue: &val,
		})
	}
	addStr := func(k, v string) {
		val := v
		ps.Keys = append(ps.Keys, k)
		ps.Values = append(ps.Values, &sparkplug.PropertyValue{
			Type: sparkplug.DataTypeString, StringValue: &val,
		})
	}

	addUint("quality", sparkplugQuality(q))
	addUint("dnp3.flags", uint32(q))
	addBool("dnp3.online", q&dnp3.QualityOnline != 0)
	addBool("dnp3.restart", q&dnp3.QualityRestart != 0)
	addBool("dnp3.comm_lost", q&dnp3.QualityCommLost != 0)
	if engUnit != "" {
		addStr("engUnit", engUnit)
	}
	return ps
}

// sparkplugQuality maps DNP3 flags to the SCADA quality convention used by
// Sparkplug receivers (192 = GOOD, 64 = STALE, 0 = BAD).
func sparkplugQuality(q dnp3.Quality) uint32 {
	if q.Good() {
		return 192
	}
	if q&dnp3.QualityOnline != 0 {
		return 64
	}
	return 0
}

// CheckDeadband returns true if the value should be published based on deadband.
func CheckDeadband(current, last, deadband float64) bool {
	if math.IsNaN(current) {
		return true
	}
	if math.IsNaN(last) {
		return true
	}
	if deadband <= 0 {
		return true
	}
	return math.Abs(current-last) >= deadband
}
