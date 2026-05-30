// Package dnp3 is a DNP3-master Source for the gateway.
//
// It implements the protocol-agnostic source.Source / source.Handler contract
// (see package source); the gateway-facing measurement, quality, and status
// types live there so other protocols (e.g. Modbus) share them.
//
// Two implementations are provided via build tags:
//
//   - default ("dnp3_stub"): in-process stub that compiles without any C library.
//     Lets the UI/API/build pipeline work end-to-end without the native stack.
//     Emits no measurements; reports connected=false for all outstations.
//
//   - "dnp3_ffi": CGO binding to opendnp3 (Apache 2.0,
//     https://github.com/dnp3/opendnp3) via the C++ shim in opendnp3_c.{h,cpp}.
//     Vendor the static lib + headers under third_party/opendnp3/{triple}/ with
//     `make opendnp3-vendor` before building with this tag.
//
// Build the real binding with: go build -tags dnp3_ffi
package dnp3
