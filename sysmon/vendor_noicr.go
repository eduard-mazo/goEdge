//go:build !icr

package sysmon

// collectVendor is a no-op off the ICR-323x: the host telemetry on dev,
// Windows, and generic Linux builds comes entirely from gopsutil. The real
// implementation lives in vendor_icr.go behind the `icr` build tag.
func (c *Collector) collectVendor(add func(string, float64), addStr func(string, string)) {}
