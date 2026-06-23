#!/usr/bin/env bash
#
# dnp3-ostn-smoke.sh — runtime smoke test for the opendnp3 OUTSTATION binding
# (the gateway acting as a DNP3 outstation server).
#
# Builds the Go harness (cmd/ostnsmoke) against the vendored opendnp3 and runs
# it. The harness runs BOTH the gateway's DNP3 outstation server and its own DNP3
# master in one process over loopback, pushes a known set of point values, and
# asserts the master reads them back (values + quality). No external simulator is
# needed — both sides are the gateway's own code, which also exercises the shared
# refcounted DNP3Manager.
#
# Usage:  bash scripts/dnp3-ostn-smoke.sh [port] [duration]
#   port      loopback TCP port for the outstation    (default 20100)
#   duration  how long the harness runs, Go format     (default 8s)
#
# Requires: the host opendnp3 vendored under third_party/opendnp3/<triple>/
#           (run `make opendnp3-vendor` first), plus a CGO toolchain.
set -euo pipefail

PORT="${1:-20100}"
DURATION="${2:-8s}"
TRIPLE="${DNP3_HOST_TRIPLE:-x86_64-unknown-linux-gnu}"

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

ODIR="$ROOT/third_party/opendnp3/$TRIPLE"
INC="$ODIR/include"
LIB="$ODIR/lib/libopendnp3.a"

if [[ ! -f "$LIB" || ! -d "$INC" ]]; then
    echo "ERROR: missing vendored opendnp3 at $ODIR (run: make opendnp3-vendor)" >&2
    exit 2
fi

HARNESS_BIN="$(mktemp -t ostnsmoke.XXXXXX)"
cleanup() { rm -f "$HARNESS_BIN"; }
trap cleanup EXIT

echo "==> building Go outstation smoke harness (-tags dnp3_ffi)"
CGO_ENABLED=1 \
CGO_CXXFLAGS="-std=c++17 -I$INC" \
CGO_LDFLAGS="-L$ODIR/lib -lopendnp3 -lssl -lcrypto -lstdc++ -lpthread -lm -ldl" \
go build -tags dnp3_ffi -o "$HARNESS_BIN" ./cmd/ostnsmoke

echo "==> running outstation+master loopback on 127.0.0.1:$PORT for $DURATION"
set +e
"$HARNESS_BIN" -port "$PORT" -duration "$DURATION"
RC=$?
set -e
exit $RC
