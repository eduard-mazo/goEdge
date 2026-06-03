//go:build icr

package sysmon

import "testing"

func TestParseStatusPanel(t *testing.T) {
	out := `Part Number        : ICR-3232
Product Type       : ICR-3232
Product Name       : ICR-323x
Firmware Version   : 6.6.1 (2026-04-24)
Serial Number      : ACZ1100002518501
Hardware UUID      : 47519ee2-ae5a-11ec-8856-000a148e9bb6
RTC Battery        : Ok
Supply Voltage     : 12.1 V
Temperature        : 41 C
Time               : 2026-06-02 21:50:31`

	f := parseStatusPanel(out)
	want := map[string]string{
		"Part Number":      "ICR-3232",
		"Product Name":     "ICR-323x",
		"Firmware Version": "6.6.1 (2026-04-24)",
		"Hardware UUID":    "47519ee2-ae5a-11ec-8856-000a148e9bb6",
		"RTC Battery":      "Ok",
		"Temperature":      "41 C",
		// Value has colons (HH:MM:SS) — only the first ":" splits.
		"Time": "2026-06-02 21:50:31",
	}
	for k, v := range want {
		if f[k] != v {
			t.Errorf("%s = %q; want %q", k, f[k], v)
		}
	}
}

func TestLeadingFloat(t *testing.T) {
	cases := map[string]struct {
		v  float64
		ok bool
	}{
		"42 C":   {42, true},
		"12.1 V": {12.1, true},
		"Ok":     {0, false},
		"":       {0, false},
	}
	for in, want := range cases {
		v, ok := leadingFloat(in)
		if ok != want.ok || (ok && v != want.v) {
			t.Errorf("leadingFloat(%q) = %v,%v; want %v,%v", in, v, ok, want.v, want.ok)
		}
	}
}

func TestBoolFloat(t *testing.T) {
	if boolFloat(true) != 1 || boolFloat(false) != 0 {
		t.Error("boolFloat mapping wrong")
	}
}
