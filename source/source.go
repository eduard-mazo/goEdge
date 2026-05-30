// Package source defines the protocol-agnostic boundary between a field
// protocol driver (DNP3 master, Modbus poller, …) and the publisher.
//
// A driver implements Source (lifecycle + status) and pushes Sample values to
// a Handler. The publisher implements Handler and aggregates one or more
// Sources, so adding a protocol means writing a new Source — the
// mapping/Sparkplug/MQTT layer is unchanged.
package source

import (
	"context"
	"time"
)

// PointType is the normalized class of a measurement. DNP3 fills these from its
// object groups; other protocols map their native types onto the same set
// (e.g. a Modbus holding register → analog, a coil → binary).
type PointType string

const (
	PointBinary             PointType = "binary"
	PointDoubleBitBinary    PointType = "double_bit_binary"
	PointBinaryOutputStatus PointType = "binary_output_status"
	PointCounter            PointType = "counter"
	PointFrozenCounter      PointType = "frozen_counter"
	PointAnalog             PointType = "analog"
	PointAnalogOutputStatus PointType = "analog_output_status"
	PointOctetString        PointType = "octet_string"
)

// Quality is a flag bitfield using DNP3 semantics (IEEE 1815 §A.4) as the
// common model. Protocols without rich quality (e.g. Modbus) set QualityOnline
// on a good read and clear it on failure.
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

// DoubleBitState encodes the DNP3 g3/g4 double-bit binary state.
type DoubleBitState uint8

const (
	DBBIntermediate DoubleBitState = 0
	DBBOff          DoubleBitState = 1
	DBBOn           DoubleBitState = 2
	DBBIndeterm     DoubleBitState = 3
)

// Sample is a single point update delivered by a Source to the Handler.
type Sample struct {
	SourceID  string // gateway-side source ID (outstation/device ID)
	PointType PointType
	Index     uint16
	Time      time.Time // value timestamp from the device (or time.Now if absent)
	Quality   Quality

	// IsEvent is true when the value arrived as a spontaneous event rather than
	// a poll/static response. Lets the publisher honor PublishOnPoll.
	IsEvent bool

	// Exactly one of the following carries the value, per PointType.
	BoolValue  bool
	DBBValue   DoubleBitState // double-bit binary
	UintValue  uint32         // counter / frozen counter
	FloatValue float64        // analog / analog output status
	BytesValue []byte         // octet string
}

// Status is a per-source snapshot. JSON tags match the web UI's expectations.
type Status struct {
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

// Handler receives samples and status changes from a Source. Implementations
// must be safe for concurrent calls from multiple Source goroutines/threads.
type Handler interface {
	OnSample(s Sample)
	OnStatusChange(s Status)
	OnLog(level, msg string)
}

// Source is the lifecycle a protocol driver exposes to the publisher.
type Source interface {
	Start(ctx context.Context) error
	Stop()
	Status() []Status
}
