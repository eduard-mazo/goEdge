// Package mapping converts DNP3 measurements to Sparkplug B metrics.
package mapping

import (
	"fmt"
	"math"

	"goMqttDnp3/config"
	"goMqttDnp3/source"
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
func Apply(sig config.SignalMapping, m source.Sample) (Result, error) {
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

	// Modbus points are formatted by function: coils/discretes → boolean,
	// registers → engineering double (scale/offset applied).
	if sig.IsModbus() {
		switch sig.Function {
		case "coil", "discrete_input":
			metric = sparkplug.MetricBool(sig.MetricName, tsMs, m.BoolValue)
			if m.BoolValue {
				value = 1
			} else {
				value = 0
			}
		default:
			eng := m.FloatValue*scale + sig.Offset
			metric = sparkplug.MetricDouble(sig.MetricName, tsMs, eng)
			value = eng
		}
		metric.Properties = metricProperties(sig, m.Quality)
		return Result{Value: value, Metric: metric, IsNull: !m.Quality.Good()}, nil
	}

	switch m.PointType {
	case source.PointBinary, source.PointBinaryOutputStatus:
		metric = sparkplug.MetricBool(sig.MetricName, tsMs, m.BoolValue)
		if m.BoolValue {
			value = 1
		} else {
			value = 0
		}

	case source.PointDoubleBitBinary:
		v := uint32(m.DBBValue)
		metric = sparkplug.MetricUInt32(sig.MetricName, tsMs, v)
		value = float64(v)

	case source.PointCounter, source.PointFrozenCounter:
		eng := float64(m.UintValue)*scale + sig.Offset
		if scale == 1.0 && sig.Offset == 0 {
			metric = sparkplug.MetricUInt32(sig.MetricName, tsMs, m.UintValue)
		} else {
			metric = sparkplug.MetricDouble(sig.MetricName, tsMs, eng)
		}
		value = eng

	case source.PointAnalog, source.PointAnalogOutputStatus:
		eng := m.FloatValue*scale + sig.Offset
		metric = sparkplug.MetricDouble(sig.MetricName, tsMs, eng)
		value = eng

	case source.PointOctetString:
		metric = sparkplug.MetricString(sig.MetricName, tsMs, string(m.BytesValue))

	default:
		return Result{}, fmt.Errorf("mapping %q: unsupported point type %q", sig.MetricName, m.PointType)
	}

	metric.Properties = metricProperties(sig, m.Quality)
	return Result{Value: value, Metric: metric, IsNull: !m.Quality.Good()}, nil
}

// BirthMetric returns a zero-valued metric declaring sig's name and the
// datatype it will carry in NDATA/DDATA. NBIRTH/DBIRTH must declare each metric
// with the SAME datatype it later sends (Sparkplug B §6.4.4) — otherwise a
// consumer that takes the datatype from the birth (and decodes data by alias)
// mis-reads the value. The type selection here mirrors Apply exactly.
func BirthMetric(sig config.SignalMapping, ts uint64) *sparkplug.Metric {
	name := sig.MetricName
	var m *sparkplug.Metric
	if sig.IsModbus() {
		switch sig.Function {
		case "coil", "discrete_input":
			m = sparkplug.MetricBool(name, ts, false)
		default: // input/holding register → engineering double
			m = sparkplug.MetricDouble(name, ts, 0)
		}
	} else {
		switch source.PointType(sig.PointType) {
		case source.PointBinary, source.PointBinaryOutputStatus:
			m = sparkplug.MetricBool(name, ts, false)
		case source.PointDoubleBitBinary:
			m = sparkplug.MetricUInt32(name, ts, 0)
		case source.PointCounter, source.PointFrozenCounter:
			// Apply keeps the raw uint32 only when there is no scale/offset.
			scale := sig.Scale
			if scale == 0 {
				scale = 1.0
			}
			if scale == 1.0 && sig.Offset == 0 {
				m = sparkplug.MetricUInt32(name, ts, 0)
			} else {
				m = sparkplug.MetricDouble(name, ts, 0)
			}
		case source.PointOctetString:
			m = sparkplug.MetricString(name, ts, "")
		default: // analog, analog_output_status, and any unknown → double
			m = sparkplug.MetricDouble(name, ts, 0)
		}
	}
	// Declare the universal UNS decomposition (+ engUnit) at birth so the
	// consumer seeds its alias→{code,instance} map without parsing the name.
	ps := sparkplug.UNSProperties(unsCode(sig), unsInstance(sig))
	if sig.EngineeringUnit != "" {
		ps.AddString("engUnit", sig.EngineeringUnit)
	}
	m.Properties = ps
	return m
}

// unsCode returns the canonical UNS Attribute for a mapping: the operator-set
// SignalCode, or the metric name when unset. Universal across protocols.
func unsCode(sig config.SignalMapping) string {
	if sig.SignalCode != "" {
		return sig.SignalCode
	}
	return sig.MetricName
}

// unsInstance returns the UNS entity instance/channel ("default" when unset).
func unsInstance(sig config.SignalMapping) string {
	if sig.Instance != "" {
		return sig.Instance
	}
	return "default"
}

// metricProperties builds the full data-metric PropertySet: quality/engUnit plus
// the universal UNS decomposition (uns/code, uns/instance).
func metricProperties(sig config.SignalMapping, q source.Quality) *sparkplug.PropertySet {
	ps := qualityProperties(q, sig.EngineeringUnit)
	ps.AddString("uns/code", unsCode(sig))
	ps.AddString("uns/instance", unsInstance(sig))
	return ps
}

// qualityProperties packs DNP3 flags into a Sparkplug PropertySet so the
// receiving SCADA can interpret point quality without out-of-band knowledge.
func qualityProperties(q source.Quality, engUnit string) *sparkplug.PropertySet {
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
	addBool("dnp3.online", q&source.QualityOnline != 0)
	addBool("dnp3.restart", q&source.QualityRestart != 0)
	addBool("dnp3.comm_lost", q&source.QualityCommLost != 0)
	if engUnit != "" {
		addStr("engUnit", engUnit)
	}
	return ps
}

// sparkplugQuality maps DNP3 flags to the SCADA quality convention used by
// Sparkplug receivers (192 = GOOD, 64 = STALE, 0 = BAD).
func sparkplugQuality(q source.Quality) uint32 {
	if q.Good() {
		return 192
	}
	if q&source.QualityOnline != 0 {
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
