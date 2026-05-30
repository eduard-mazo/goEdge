#!/usr/bin/env bash
#
# dnp3-smoke.sh — end-to-end runtime smoke test for the opendnp3 master binding.
#
# Builds the C++ outstation simulator (scripts/sim/outstation_sim.cpp) and the
# Go smoke harness (cmd/dnp3smoke), both against the vendored opendnp3, then
# runs the harness against the sim on a loopback TCP port and reports pass/fail.
#
# Usage:  bash scripts/dnp3-smoke.sh [port] [duration]
#   port      loopback TCP port for the sim          (default 20000)
#   duration  how long the harness runs, Go format   (default 12s)
#
# Requires: the host opendnp3 vendored under third_party/opendnp3/<triple>/
#           (run `make opendnp3-vendor` first), plus g++ and a CGO toolchain.
set -euo pipefail

PORT="${1:-20000}"
DURATION="${2:-12s}"
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

SIM_BIN="$(mktemp -t outstation_sim.XXXXXX)"
HARNESS_BIN="$(mktemp -t dnp3smoke.XXXXXX)"
SIM_LOG="$(mktemp -t outstation_sim_log.XXXXXX)"
SIM_PID=""

cleanup() {
    [[ -n "$SIM_PID" ]] && kill "$SIM_PID" 2>/dev/null || true
    [[ -n "$SIM_PID" ]] && wait "$SIM_PID" 2>/dev/null || true
    rm -f "$SIM_BIN" "$HARNESS_BIN" "$SIM_LOG"
}
trap cleanup EXIT

echo "==> building outstation simulator"
g++ -std=c++17 -I "$INC" \
    scripts/sim/outstation_sim.cpp "$LIB" \
    -lssl -lcrypto -lpthread -o "$SIM_BIN"

echo "==> building Go smoke harness (-tags dnp3_ffi)"
CGO_ENABLED=1 \
CGO_CXXFLAGS="-std=c++17 -I$INC" \
CGO_LDFLAGS="-L$ODIR/lib -lopendnp3 -lssl -lcrypto -lstdc++ -lpthread -lm -ldl" \
go build -tags dnp3_ffi -o "$HARNESS_BIN" ./cmd/dnp3smoke

echo "==> starting simulator on :$PORT"
"$SIM_BIN" "$PORT" >"$SIM_LOG" 2>&1 &
SIM_PID=$!
sleep 1
if ! kill -0 "$SIM_PID" 2>/dev/null; then
    echo "ERROR: simulator failed to start:" >&2
    cat "$SIM_LOG" >&2
    exit 2
fi

echo "==> running harness against 127.0.0.1:$PORT for $DURATION"
set +e
"$HARNESS_BIN" -addr "127.0.0.1:$PORT" -duration "$DURATION"
RC=$?
set -e

echo
echo "==> simulator log (tail):"
tail -n 15 "$SIM_LOG" | sed 's/^/    /'

exit $RC
