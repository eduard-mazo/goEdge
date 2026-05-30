// Package dnp3 is the gateway-facing abstraction over a DNP3 master.
//
// Two implementations are provided via build tags:
//
//   - default ("dnp3_stub"): in-process stub that compiles without any C library.
//     Lets the UI/API/build pipeline work end-to-end while libdnp3_ffi is being
//     vendored. Emits no measurements; reports connected=false for all outstations.
//
//   - "dnp3_ffi": CGO binding to Step Function I/O libdnp3_ffi
//     (https://github.com/stepfunc/dnp3). Vendor the C headers and static lib
//     under third_party/dnp3/{triple}/ before building with this tag.
//
// Build the real binding with: go build -tags dnp3_ffi
package dnp3

import "time"

// PointType enumerates the DNP3 static object types this gateway exposes.
// Each one maps to one or more (Group, Variation) pairs in IEEE 1815.
type PointType string

const (
	PointBinary             PointType = "binary"               // g1 / g2 events
	PointDoubleBitBinary    PointType = "double_bit_binary"    // g3 / g4 events
	PointBinaryOutputStatus PointType = "binary_output_status" // g10 / g11 events
	PointCounter            PointType = "counter"              // g20 / g22 events
	PointFrozenCounter      PointType = "frozen_counter"       // g21 / g23 events
	PointAnalog             PointType = "analog"               // g30 / g32 events
	PointAnalogOutputStatus PointType = "analog_output_status" // g40 / g42 events
	PointOctetString        PointType = "octet_string"         // g110 / g111 events
)

// Quality is the DNP3 flag bitfield reported per measurement (IEEE 1815 §A.4).
// The bits vary slightly per point type; we preserve the raw byte and expose
// helper accessors for the bits that are common across most types.
type Quality uint8

const (
	QualityOnline       Quality = 1 << 0
	QualityRestart      Quality = 1 << 1
	QualityCommLost     Quality = 1 << 2
	QualityRemoteForced Quality = 1 << 3
	QualityLocalForced  Quality = 1 << 4
	QualityOverRange    Quality = 1 << 5 // analog only
	QualityRefError     Quality = 1 << 6 // analog only
	QualityChatter      Quality = 1 << 5 // binary only (alias of OverRange bit)
)

// Good reports whether the flags indicate a usable measurement
// (online and not restart/comm-lost).
func (q Quality) Good() bool {
	return q&QualityOnline != 0 &&
		q&QualityRestart == 0 &&
		q&QualityCommLost == 0
}

// Measurement is a single point update delivered by the master to the gateway.
// The lib normalizes group/variation differences into a typed value.
type Measurement struct {
	OutstationID string    // gateway-side ID (config.DNP3Outstation.ID)
	PointType    PointType
	Index        uint16
	Time         time.Time // measurement timestamp from outstation (or time.Now if absent)
	Quality      Quality

	// Exactly one of the following carries the value, per PointType.
	BoolValue   bool
	DBBValue    DoubleBitState // double-bit binary
	UintValue   uint32         // counter / frozen counter
	FloatValue  float64        // analog / analog output status
	BytesValue  []byte         // octet string
}

// DoubleBitState encodes the DNP3 g3/g4 double-bit binary state.
type DoubleBitState uint8

const (
	DBBIntermediate DoubleBitState = 0
	DBBOff          DoubleBitState = 1
	DBBOn           DoubleBitState = 2
	DBBIndeterm     DoubleBitState = 3
)

// OutstationStatus is reported by the master per association.
// JSON tags must match web/src/api/client.ts OutstationStatus interface.
type OutstationStatus struct {
	ID              string    `json:"id"`
	Label           string    `json:"label"`
	Addr            string    `json:"addr"`
	Connected       bool      `json:"connected"`
	LastError       string    `json:"lastError"`
	MeasurementsRx  int64     `json:"measurementsRx"`
	IntegrityPolls  int64     `json:"integrityPolls"`
	ClassPolls      int64     `json:"classPolls"`
	UnsolicitedRsps int64     `json:"unsolicitedRsps"`
	LastReadAt      time.Time `json:"lastReadAt"`
}

// Handler is implemented by the publisher to receive measurements and
// status changes from the master.
type Handler interface {
	OnMeasurement(m Measurement)
	OnStatusChange(s OutstationStatus)
	OnLog(level, msg string)
}
