package sysmon

import (
	"strings"
	"testing"

	"goMqttDnp3/config"
)

func TestSanitizeMount(t *testing.T) {
	cases := map[string]string{
		"/":          "root",
		"":           "root",
		"/data":      "data",
		"/mnt/sdcard": "mnt_sdcard",
		"eth0":       "eth0",
		"/var/log/":  "var_log",
	}
	for in, want := range cases {
		if got := sanitizeMount(in); got != want {
			t.Errorf("sanitizeMount(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestRateKbps(t *testing.T) {
	// 1 MiB over 2s = 1048576 bytes * 8 / 1000 / 2 = 4194.304 kbps
	if got := rateKbps(1048576, 0, 2); got != 4194.3 {
		t.Errorf("rateKbps = %v, want 4194.3", got)
	}
	// counter reset (cur < prev) yields 0, never negative.
	if got := rateKbps(5, 100, 1); got != 0 {
		t.Errorf("rateKbps(reset) = %v, want 0", got)
	}
}

func TestCanonicalSuffix(t *testing.T) {
	cases := map[string]string{
		"Disk/root/Free_MB":        "Disk/<mount>/Free_MB",
		"Disk/mnt_sd/Used_pct":     "Disk/<mount>/Used_pct",
		"Network/eth0/RxRate_kbps": "Network/<iface>/RxRate_kbps",
		"CPU/Usage_pct":            "CPU/Usage_pct", // unchanged
		"Uptime_h":                 "Uptime_h",
	}
	for in, want := range cases {
		if got := canonicalSuffix(in); got != want {
			t.Errorf("canonicalSuffix(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDisabledMetricsExcluded(t *testing.T) {
	c := New(config.SystemConfig{
		DisabledMetrics: []string{"CPU/Usage_pct", "Disk/<mount>/Free_MB"},
	})
	for _, m := range c.Collect(1) {
		// names carry the default "System/" prefix
		if m.Name == "System/CPU/Usage_pct" {
			t.Error("disabled metric CPU/Usage_pct was published")
		}
		if strings.HasPrefix(m.Name, "System/Disk/") && strings.HasSuffix(m.Name, "/Free_MB") {
			t.Errorf("placeholder-disabled metric %q was published", m.Name)
		}
	}
}

func TestRound2(t *testing.T) {
	if got := round2(1.23456); got != 1.23 {
		t.Errorf("round2 = %v, want 1.23", got)
	}
}

// TestCollectSmoke exercises a real Collect against the host. It must not panic
// and should yield CPU/memory metrics on any Linux test runner.
func TestCollectSmoke(t *testing.T) {
	c := New(config.SystemConfig{})
	ms := c.Collect(1)
	if len(ms) == 0 {
		t.Fatal("Collect returned no metrics")
	}
	want := map[string]bool{
		"System/CPU/Usage_pct":  false,
		"System/Memory/Used_pct": false,
	}
	for _, m := range ms {
		if _, ok := want[m.Name]; ok {
			want[m.Name] = true
		}
	}
	for name, seen := range want {
		if !seen {
			t.Errorf("expected metric %q not collected", name)
		}
	}
}

// TestHostUNS pins the contract-v3 §5.1 leaf/folder split for host telemetry
// (the cosmetic metricPrefix never enters uns/*).
func TestHostUNS(t *testing.T) {
	cases := []struct {
		path     string
		code     string
		instance string
	}{
		{"Uptime_h", "Uptime_h", "default"},
		{"CPU/Usage_pct", "Usage_pct", "CPU"},
		{"Memory/Free_MB", "Free_MB", "Memory"},
		{"Disk/root/Used_pct", "Used_pct", "Disk/root"},
		{"Network/eth0/Rx_MB", "Rx_MB", "Network/eth0"},
		{"Process/Count", "Count", "Process"},
	}
	for _, c := range cases {
		code, instance := hostUNS(c.path)
		if code != c.code || instance != c.instance {
			t.Errorf("hostUNS(%q) = (%q,%q); want (%q,%q)",
				c.path, code, instance, c.code, c.instance)
		}
	}
}
