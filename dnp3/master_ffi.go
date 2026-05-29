//go:build dnp3_ffi

package dnp3

// This file is the CGO binding to Step Function I/O libdnp3_ffi.
// It is intentionally left as a stub until the headers are vendored at
//   third_party/dnp3/{triple}/include/dnp3.h
// and reviewed. Once present, this file will:
//
//   1. #cgo CFLAGS  -I${SRCDIR}/../third_party/dnp3/<triple>/include
//      #cgo LDFLAGS -L${SRCDIR}/../third_party/dnp3/<triple>/lib -ldnp3_ffi -lpthread -ldl -lm
//      #include "dnp3.h"
//
//   2. Wrap the runtime (dnp3_runtime_create / destroy).
//
//   3. Wrap the master channel (dnp3_master_channel_create_tcp).
//
//   4. For each outstation, dnp3_master_channel_add_association with
//      a ReadHandler vtable whose function pointers are exported Go funcs
//      using cgo.Handle to route into the registered Handler.
//
//   5. Schedule integrity + class polls via dnp3_master_channel_add_poll
//      with the per-outstation cadence from config.
//
//   6. Enable unsolicited per the config flags using
//      dnp3_master_channel_enable_unsolicited.
//
//   7. Translate g1/g2/g3/g4/g20/g21/g30/g32/g110 events to Measurement
//      and call h.OnMeasurement on a worker goroutine (so the cgo
//      callback returns promptly).
//
// Until then, building with -tags dnp3_ffi will fail at link time — by
// design, so it is impossible to ship a binary that claims DNP3 support
// without the real lib.

import (
	"context"
	"errors"

	"goMqttDnp3/config"
)

type ffiMaster struct {
	h Handler
}

func newMaster(h Handler) Master {
	return &ffiMaster{h: h}
}

func (m *ffiMaster) AddOutstation(o config.DNP3Outstation) error {
	return errors.New("dnp3_ffi: binding not implemented — vendor libdnp3_ffi headers first")
}

func (m *ffiMaster) RemoveOutstation(id string) error {
	return errors.New("dnp3_ffi: binding not implemented")
}

func (m *ffiMaster) Start(_ context.Context) error {
	return errors.New("dnp3_ffi: binding not implemented")
}

func (m *ffiMaster) Stop() {}

func (m *ffiMaster) Status() []OutstationStatus { return nil }

func (m *ffiMaster) IntegrityPoll(id string) error {
	return errors.New("dnp3_ffi: binding not implemented")
}
