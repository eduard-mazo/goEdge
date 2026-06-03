//go:build icr

package sysmon

import "testing"

func TestParseStatusPanel(t *testing.T) {
	out := `Part Number        : ICR-3232
RTC Battery        : Ok
Supply Voltage     : 12.1 V
Temperature        : 42 C
Uptime             : 0 days, 2 hours, 12 minutes`

	f := parseStatusPanel(out)
	if f["Part Number"] != "ICR-3232" {
		t.Errorf("Part Number = %q", f["Part Number"])
	}
	if f["RTC Battery"] != "Ok" {
		t.Errorf("RTC Battery = %q", f["RTC Battery"])
	}
	// "Uptime" value contains colons in the time — Cut on the first ":" only.
	if f["Temperature"] != "42 C" {
		t.Errorf("Temperature = %q", f["Temperature"])
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
