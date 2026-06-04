package mapping

import (
	"testing"

	"goMqttDnp3/config"
	"goMqttDnp3/source"
)

// TestBirthMetricMatchesApply asserts that the datatype declared in NBIRTH/DBIRTH
// (BirthMetric) is identical to the datatype Apply emits in NDATA/DDATA for the
// same mapping — Sparkplug B requires birth and data to agree. See the shared
// contract in docs/sparkplug-contract.md.
func TestBirthMetricMatchesApply(t *testing.T) {
	cases := []struct {
		name string
		sig  config.SignalMapping
		samp source.Sample
	}{
		{"dnp3 binary", config.SignalMapping{MetricName: "b", PointType: "binary"},
			source.Sample{PointType: source.PointBinary, BoolValue: true}},
		{"dnp3 analog", config.SignalMapping{MetricName: "a", PointType: "analog"},
			source.Sample{PointType: source.PointAnalog, FloatValue: 1.5}},
		{"dnp3 counter raw", config.SignalMapping{MetricName: "c", PointType: "counter"},
			source.Sample{PointType: source.PointCounter, UintValue: 7}},
		{"dnp3 counter scaled", config.SignalMapping{MetricName: "cs", PointType: "counter", Scale: 0.1},
			source.Sample{PointType: source.PointCounter, UintValue: 7}},
		{"dnp3 double-bit", config.SignalMapping{MetricName: "d", PointType: "double_bit_binary"},
			source.Sample{PointType: source.PointDoubleBitBinary, DBBValue: source.DBBOn}},
		{"dnp3 octet", config.SignalMapping{MetricName: "o", PointType: "octet_string"},
			source.Sample{PointType: source.PointOctetString, BytesValue: []byte("x")}},
		{"modbus coil", config.SignalMapping{MetricName: "mc", Protocol: "modbus", Function: "coil", PointType: "coil"},
			source.Sample{PointType: "coil", BoolValue: true}},
		{"modbus holding", config.SignalMapping{MetricName: "mh", Protocol: "modbus", Function: "holding_register", PointType: "holding_register"},
			source.Sample{PointType: "holding_register", FloatValue: 12.3}},
	}
	for _, c := range cases {
		res, err := Apply(c.sig, c.samp)
		if err != nil {
			t.Errorf("%s: Apply: %v", c.name, err)
			continue
		}
		birth := BirthMetric(c.sig, 0)
		if birth.DataType != res.Metric.DataType {
			t.Errorf("%s: birth datatype %d != data datatype %d", c.name, birth.DataType, res.Metric.DataType)
		}
		if birth.Name != c.sig.MetricName {
			t.Errorf("%s: birth name %q != %q", c.name, birth.Name, c.sig.MetricName)
		}
	}
}
