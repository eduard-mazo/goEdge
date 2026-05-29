// Package mapping converts raw Modbus bytes to Sparkplug B metric values.
package mapping

import (
	"encoding/binary"
	"fmt"
	"math"

	"goMqttModbus/config"
	"goMqttModbus/sparkplug"
)

// ReadFnName maps config function strings to modbus ReadFunc names.
var ReadFnName = map[string]struct{}{
	"coil":             {},
	"discrete_input":   {},
	"input_register":   {},
	"holding_register": {},
}

// Result holds the decoded engineering value and the Sparkplug metric.
type Result struct {
	Value  float64
	IsNull bool
	Metric *sparkplug.Metric
}

// Apply decodes raw Modbus bytes according to SignalMapping and returns a Sparkplug metric.
func Apply(sig config.SignalMapping, rawBytes []byte, ts uint64) (Result, error) {
	reordered, err := reorder(rawBytes, sig.ByteOrder, sig.DataType)
	if err != nil {
		return Result{}, fmt.Errorf("mapping %q byte-order: %w", sig.MetricName, err)
	}

	value, metric, err := decode(sig, reordered, ts)
	if err != nil {
		return Result{}, fmt.Errorf("mapping %q decode: %w", sig.MetricName, err)
	}
	return Result{Value: value, Metric: metric}, nil
}

// BoolFromCoil extracts a single bool from a coil byte slice.
// Modbus coils pack 8 coils per byte, LSB first.
func BoolFromCoil(data []byte, bitIndex uint16) bool {
	byteIdx := bitIndex / 8
	bitIdx := bitIndex % 8
	if int(byteIdx) >= len(data) {
		return false
	}
	return (data[byteIdx]>>bitIdx)&1 == 1
}

// reorder applies byte/word swapping according to byteOrder and dataType.
// Modbus wire format is always big-endian (ABCD). We may need to re-order
// to match the PLC's internal representation.
//
// ByteOrder values:
//
//	ABCD - big-endian (Modbus default, no swap)
//	DCBA - little-endian (byte + word swapped)
//	BADC - byte-swapped within each word, words in BE order
//	CDAB - word-swapped, bytes within each word in BE order
func reorder(data []byte, byteOrder, dataType string) ([]byte, error) {
	if byteOrder == "" || byteOrder == "ABCD" {
		return data, nil
	}
	n := len(data)
	out := make([]byte, n)
	switch byteOrder {
	case "DCBA":
		for i := 0; i < n; i++ {
			out[i] = data[n-1-i]
		}
	case "BADC":
		// swap bytes within each 2-byte word
		for i := 0; i+1 < n; i += 2 {
			out[i] = data[i+1]
			out[i+1] = data[i]
		}
		if n%2 != 0 {
			out[n-1] = data[n-1]
		}
	case "CDAB":
		// swap 2-byte word order
		if n == 4 {
			out[0] = data[2]
			out[1] = data[3]
			out[2] = data[0]
			out[3] = data[1]
		} else if n == 8 {
			out[0] = data[6]
			out[1] = data[7]
			out[2] = data[4]
			out[3] = data[5]
			out[4] = data[2]
			out[5] = data[3]
			out[6] = data[0]
			out[7] = data[1]
		} else {
			return data, nil
		}
	default:
		return nil, fmt.Errorf("unknown byteOrder %q", byteOrder)
	}
	return out, nil
}

func decode(sig config.SignalMapping, data []byte, ts uint64) (float64, *sparkplug.Metric, error) {
	name := sig.MetricName
	scale := sig.Scale
	if scale == 0 {
		scale = 1.0
	}
	applyLinear := func(raw float64) float64 { return raw*scale + sig.Offset }

	switch sig.DataType {
	case "bool":
		val := BoolFromCoil(data, 0)
		return boolToFloat(val), sparkplug.MetricBool(name, ts, val), nil

	case "int16":
		if len(data) < 2 {
			return 0, nil, fmt.Errorf("int16 needs 2 bytes, got %d", len(data))
		}
		raw := int16(binary.BigEndian.Uint16(data[:2]))
		eng := applyLinear(float64(raw))
		return eng, metricDouble(name, ts, eng, sig), nil

	case "uint16":
		if len(data) < 2 {
			return 0, nil, fmt.Errorf("uint16 needs 2 bytes, got %d", len(data))
		}
		raw := binary.BigEndian.Uint16(data[:2])
		eng := applyLinear(float64(raw))
		return eng, metricDouble(name, ts, eng, sig), nil

	case "int32":
		if len(data) < 4 {
			return 0, nil, fmt.Errorf("int32 needs 4 bytes, got %d", len(data))
		}
		raw := int32(binary.BigEndian.Uint32(data[:4]))
		eng := applyLinear(float64(raw))
		return eng, metricDouble(name, ts, eng, sig), nil

	case "uint32":
		if len(data) < 4 {
			return 0, nil, fmt.Errorf("uint32 needs 4 bytes, got %d", len(data))
		}
		raw := binary.BigEndian.Uint32(data[:4])
		eng := applyLinear(float64(raw))
		return eng, metricDouble(name, ts, eng, sig), nil

	case "int64":
		if len(data) < 8 {
			return 0, nil, fmt.Errorf("int64 needs 8 bytes, got %d", len(data))
		}
		raw := int64(binary.BigEndian.Uint64(data[:8]))
		eng := applyLinear(float64(raw))
		return eng, metricDouble(name, ts, eng, sig), nil

	case "uint64":
		if len(data) < 8 {
			return 0, nil, fmt.Errorf("uint64 needs 8 bytes, got %d", len(data))
		}
		raw := binary.BigEndian.Uint64(data[:8])
		eng := applyLinear(float64(raw))
		return eng, metricDouble(name, ts, eng, sig), nil

	case "float32":
		if len(data) < 4 {
			return 0, nil, fmt.Errorf("float32 needs 4 bytes, got %d", len(data))
		}
		raw := math.Float32frombits(binary.BigEndian.Uint32(data[:4]))
		eng := applyLinear(float64(raw))
		return eng, metricDouble(name, ts, eng, sig), nil

	case "float64":
		if len(data) < 8 {
			return 0, nil, fmt.Errorf("float64 needs 8 bytes, got %d", len(data))
		}
		raw := math.Float64frombits(binary.BigEndian.Uint64(data[:8]))
		eng := applyLinear(raw)
		return eng, metricDouble(name, ts, eng, sig), nil

	default:
		return 0, nil, fmt.Errorf("unsupported dataType %q", sig.DataType)
	}
}

func boolToFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// metricDouble returns a Sparkplug Double metric.
// We always publish doubles for numeric types for SCADA-system compatibility.
func metricDouble(name string, ts uint64, value float64, sig config.SignalMapping) *sparkplug.Metric {
	m := sparkplug.MetricDouble(name, ts, value)
	if sig.EngineeringUnit != "" {
		m.Properties = &sparkplug.PropertySet{
			Keys: []string{"engUnit"},
			Values: []*sparkplug.PropertyValue{
				{Type: sparkplug.DataTypeString, StringValue: strPtr(sig.EngineeringUnit)},
			},
		}
	}
	return m
}

func strPtr(s string) *string { return &s }

// CheckDeadband returns true if the value should be published based on deadband policy.
// last is the previous published value (NaN if never published).
func CheckDeadband(current, last, deadband float64) bool {
	if math.IsNaN(last) {
		return true
	}
	if deadband <= 0 {
		return true
	}
	return math.Abs(current-last) >= deadband
}

// RequiredBytes returns the minimum byte count for a given dataType.
func RequiredBytes(dataType string) int {
	switch dataType {
	case "bool":
		return 1
	case "int16", "uint16":
		return 2
	case "int32", "uint32", "float32":
		return 4
	case "int64", "uint64", "float64":
		return 8
	default:
		return 0
	}
}

// RequiredRegisters returns how many 16-bit registers a dataType needs.
func RequiredRegisters(dataType string) uint16 {
	switch dataType {
	case "int32", "uint32", "float32":
		return 2
	case "int64", "uint64", "float64":
		return 4
	default:
		return 1
	}
}
