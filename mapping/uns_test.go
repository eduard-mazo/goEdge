package mapping

import (
	"testing"

	"goMqttDnp3/config"
	"goMqttDnp3/sparkplug"
)

// TestUNSDecomposition pins the contract-v3 §5.1 defaults: uns/code is the
// LEAF of the metric name and uns/instance the folder path; explicit
// signalCode/instance overrides always win.
func TestUNSDecomposition(t *testing.T) {
	cases := []struct {
		name     string
		sig      config.SignalMapping
		code     string
		instance string
	}{
		{"flat", config.SignalMapping{MetricName: "tank_press"}, "tank_press", "default"},
		{"one folder", config.SignalMapping{MetricName: "PLC/tank_level"}, "tank_level", "PLC"},
		{"valve folder", config.SignalMapping{MetricName: "VALV/VALV_ON"}, "VALV_ON", "VALV"},
		{"deep folder", config.SignalMapping{MetricName: "Feeder1/PhaseA/Voltage"}, "Voltage", "Feeder1/PhaseA"},
		{"explicit code", config.SignalMapping{MetricName: "PLC/tank_level", SignalCode: "TL"}, "TL", "PLC"},
		{"explicit instance", config.SignalMapping{MetricName: "tank_level", Instance: "PLC"}, "tank_level", "PLC"},
		{"both explicit", config.SignalMapping{MetricName: "x/y", SignalCode: "C", Instance: "I"}, "C", "I"},
	}
	for _, c := range cases {
		if got := unsCode(c.sig); got != c.code {
			t.Errorf("%s: unsCode = %q; want %q", c.name, got, c.code)
		}
		if got := unsInstance(c.sig); got != c.instance {
			t.Errorf("%s: unsInstance = %q; want %q", c.name, got, c.instance)
		}
	}
}

// TestBirthMetricMetadata asserts that BirthMetric declares uns/code,
// uns/instance and — when configured — the birth-only uns/name and
// uns/description catalog metadata (contract v3 §5).
func TestBirthMetricMetadata(t *testing.T) {
	sig := config.SignalMapping{
		MetricName:  "VALV/VALV_ON",
		PointType:   "binary",
		Nombre:      "Valvula abierta",
		Descripcion: "Valvula gas confirmación apertura",
	}
	m := BirthMetric(sig, 1)
	props := propsToMap(t, m.Properties)
	want := map[string]string{
		"uns/code":        "VALV_ON",
		"uns/instance":    "VALV",
		"uns/name":        "Valvula abierta",
		"uns/description": "Valvula gas confirmación apertura",
	}
	for k, v := range want {
		if props[k] != v {
			t.Errorf("birth property %q = %q; want %q", k, props[k], v)
		}
	}

	// Without nombre/descripcion the metadata keys must be absent.
	m = BirthMetric(config.SignalMapping{MetricName: "tank_press", PointType: "analog"}, 1)
	props = propsToMap(t, m.Properties)
	for _, k := range []string{"uns/name", "uns/description"} {
		if _, ok := props[k]; ok {
			t.Errorf("birth property %q present; want absent", k)
		}
	}
	if props["uns/code"] != "tank_press" || props["uns/instance"] != "default" {
		t.Errorf("flat metric uns = (%q,%q); want (tank_press,default)",
			props["uns/code"], props["uns/instance"])
	}
}

func propsToMap(t *testing.T, ps *sparkplug.PropertySet) map[string]string {
	t.Helper()
	if ps == nil {
		t.Fatal("nil PropertySet")
	}
	out := make(map[string]string, len(ps.Keys))
	for i, k := range ps.Keys {
		if i < len(ps.Values) && ps.Values[i] != nil && ps.Values[i].StringValue != nil {
			out[k] = *ps.Values[i].StringValue
		}
	}
	return out
}
