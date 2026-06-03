//go:build icr

package sysmon

import (
	"context"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// statusCmd is the Advantech firmware tool that prints the device status panel.
// `status sys` lists the board sensors gopsutil cannot read on this hardware
// (temperature, supply voltage, RTC battery) in a "Label : value" format.
const statusCmd = "/usr/bin/status"

// collectVendor augments the gopsutil sample with ICR-323x board sensors by
// shelling out to `status sys` once per tick. It is compiled only into the
// on-device build (-tags icr). A missing/slow tool is skipped (warned once)
// rather than blocking the collector — the rest of the batch still publishes.
func (c *Collector) collectVendor(add func(string, float64)) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, statusCmd, "sys").Output()
	if err != nil {
		c.warn("icr-status", err)
		return
	}
	fields := parseStatusPanel(string(out))

	// Temperature shares the standard suffix so it slots into the same metric
	// the publisher already expects; gated by the Temperature group toggle.
	if c.sel.Temperature {
		if v, ok := leadingFloat(fields["Temperature"]); ok { // e.g. "42 C"
			add("Temperature/CPU_C", round2(v))
		}
	}
	// Supply voltage and RTC battery are ICR-specific — no gopsutil equivalent —
	// so they publish whenever the panel reports them. Opt out per-metric via
	// DisabledMetrics ("Power/Supply_V", "Power/RTC_Battery_OK").
	if v, ok := leadingFloat(fields["Supply Voltage"]); ok { // e.g. "12.1 V"
		add("Power/Supply_V", round2(v))
	}
	if s, ok := fields["RTC Battery"]; ok { // "Ok" / "Low" / ...
		add("Power/RTC_Battery_OK", boolFloat(strings.EqualFold(strings.TrimSpace(s), "Ok")))
	}
}

// parseStatusPanel splits "Label : value" lines into a map keyed by the trimmed
// label. Labels are right-padded with spaces, so both sides are trimmed.
func parseStatusPanel(out string) map[string]string {
	fields := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		k, v, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		fields[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return fields
}

// leadingFloat parses the numeric prefix of a value like "42 C" or "12.1 V".
func leadingFloat(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, ' '); i > 0 {
		s = s[:i]
	}
	v, err := strconv.ParseFloat(s, 64)
	return v, err == nil
}

func boolFloat(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
