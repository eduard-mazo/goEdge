package dnp3

import (
	"context"

	"goMqttDnp3/config"
)

// Master abstracts the DNP3 master runtime.
//
// Lifecycle: New → AddOutstation* → Start (connects TCP channels, sends
// startup integrity polls per association) → ... → Stop (closes channels,
// cancels associations).
//
// Implementations:
//   - stubMaster in master_stub.go (default; no C deps)
//   - ffiMaster in master_ffi.go  (build tag dnp3_ffi; wraps libdnp3_ffi)
type Master interface {
	// AddOutstation registers an outstation. Must be called before Start.
	// Re-adding the same ID replaces the prior config.
	AddOutstation(o config.DNP3Outstation) error

	// RemoveOutstation removes an outstation and tears down its channel.
	// Safe to call while running.
	RemoveOutstation(id string) error

	// Start brings up all enabled outstation channels and begins polling.
	// The provided context cancels long-running setup. The Handler will
	// receive measurements asynchronously from internal goroutines.
	Start(ctx context.Context) error

	// Stop tears down all channels and waits for in-flight callbacks.
	Stop()

	// Status returns the current per-outstation status snapshot.
	Status() []OutstationStatus

	// IntegrityPoll triggers an on-demand integrity poll for one outstation.
	// Returns immediately; the response arrives via Handler.OnMeasurement.
	IntegrityPoll(outstationID string) error
}

// New constructs a Master. The concrete type is selected by build tag.
// See newMaster in master_{stub,ffi}.go.
func New(h Handler) Master {
	return newMaster(h)
}
