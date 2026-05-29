package mapping

import (
	"math"
	"testing"

	"goMqttModbus/config"
)

func TestApplyUInt16Scale(t *testing.T) {
	sig := config.SignalMapping{
		MetricName: "temp",
		DataType:   "uint16",
		ByteOrder:  "ABCD",
		Scale:      0.1,
		Offset:     -40.0,
	}
	// raw = 0x01F4 = 500 → 500 * 0.1 - 40 = 10.0
	raw := []byte{0x01, 0xF4}
	r, err := Apply(sig, raw, 0)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(r.Value-10.0) > 0.001 {
		t.Fatalf("expected 10.0, got %f", r.Value)
	}
}

func TestApplyFloat32(t *testing.T) {
	sig := config.SignalMapping{
		MetricName: "pressure",
		DataType:   "float32",
		ByteOrder:  "ABCD",
		Scale:      1.0,
	}
	// IEEE 754 for 3.14 = 0x4048F5C3
	raw := []byte{0x40, 0x48, 0xF5, 0xC3}
	r, err := Apply(sig, raw, 0)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(r.Value-3.14) > 0.0001 {
		t.Fatalf("expected ~3.14, got %f", r.Value)
	}
}

func TestApplyBool(t *testing.T) {
	sig := config.SignalMapping{
		MetricName: "coil0",
		DataType:   "bool",
	}
	raw := []byte{0x01} // coil 0 = ON
	r, err := Apply(sig, raw, 0)
	if err != nil {
		t.Fatal(err)
	}
	if r.Value != 1.0 {
		t.Fatalf("expected 1.0, got %f", r.Value)
	}
}

func TestByteOrderDCBA(t *testing.T) {
	sig := config.SignalMapping{
		MetricName: "val",
		DataType:   "uint16",
		ByteOrder:  "DCBA",
		Scale:      1.0,
	}
	// wire: 0xF401 → DCBA reorder → 0x01F4 = 500
	raw := []byte{0xF4, 0x01}
	r, err := Apply(sig, raw, 0)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(r.Value-500) > 0.001 {
		t.Fatalf("expected 500, got %f", r.Value)
	}
}

func TestCheckDeadband(t *testing.T) {
	if !CheckDeadband(10.5, math.NaN(), 0.5) {
		t.Fatal("first reading should always publish")
	}
	if CheckDeadband(10.3, 10.0, 0.5) {
		t.Fatal("change 0.3 < deadband 0.5 should not publish")
	}
	if !CheckDeadband(10.6, 10.0, 0.5) {
		t.Fatal("change 0.6 >= deadband 0.5 should publish")
	}
}

func TestBoolFromCoil(t *testing.T) {
	data := []byte{0b10110101} // bits: 7=1,6=0,5=1,4=1,3=0,2=1,1=0,0=1
	cases := []struct {
		bit  uint16
		want bool
	}{
		{0, true}, {1, false}, {2, true}, {3, false},
		{4, true}, {5, true}, {6, false}, {7, true},
	}
	for _, c := range cases {
		got := BoolFromCoil(data, c.bit)
		if got != c.want {
			t.Errorf("bit %d: want %v got %v", c.bit, c.want, got)
		}
	}
}
