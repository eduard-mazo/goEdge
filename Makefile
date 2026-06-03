BINARY   := goMqttDnp3
CONFIG   ?= config.json
PORT     ?= 8080
LOG      ?= info

# TLS for the web UI/API. Set both to serve HTTPS from the run targets, e.g.
#   make run-ffi TLS_CERT=cert.pem TLS_KEY=key.pem
# Generate a throwaway dev cert with `make dev-cert`.
TLS_CERT ?=
TLS_KEY  ?=
TLS_FLAGS = $(if $(strip $(TLS_CERT)),-tls-cert $(TLS_CERT) -tls-key $(TLS_KEY),)

# DNP3 native lib (opendnp3, Apache 2.0) — vendored under third_party/opendnp3/{triple}/
# Built once per host with `make opendnp3-vendor[-arm]`. Static archive, so it
# links into the Go binary; nothing extra to deploy on the target.
DNP3_HOST_TRIPLE   ?= x86_64-unknown-linux-gnu
DNP3_ARM_TRIPLE    ?= armv7-unknown-linux-gnueabihf
DNP3_WIN_TRIPLE    ?= x86_64-w64-mingw32
DNP3_HOST_DIR      := third_party/opendnp3/$(DNP3_HOST_TRIPLE)
DNP3_ARM_DIR       := third_party/opendnp3/$(DNP3_ARM_TRIPLE)
DNP3_WIN_DIR       := third_party/opendnp3/$(DNP3_WIN_TRIPLE)

# The C++ shim (dnp3/opendnp3_c.cpp) is compiled by cgo; it needs the opendnp3
# headers (CXXFLAGS) and the static libs + their TLS/stdc++/pthread deps (LDFLAGS).
DNP3_HOST_CXXFLAGS := -std=c++17 -I$(CURDIR)/$(DNP3_HOST_DIR)/include
DNP3_HOST_LDFLAGS  := -L$(CURDIR)/$(DNP3_HOST_DIR)/lib -lopendnp3 -lssl -lcrypto -lstdc++ -lpthread -lm -ldl
DNP3_ARM_CXXFLAGS  := -std=c++17 -I$(CURDIR)/$(DNP3_ARM_DIR)/include
# The armv7 opendnp3 is vendored with DNP3_TLS=OFF (no armhf OpenSSL on the build
# host), so it links no ssl/crypto. Force-static the C++ runtime via the literal
# archive (-l:libstdc++.a) so the binary needs no libstdc++ on the device — a
# plain -lstdc++ would pull the shared lib and defeat -static-libstdc++.
# To enable secure DNP3 later: vendor with an armhf OpenSSL and add -lssl -lcrypto.
DNP3_ARM_LDFLAGS   := -L$(CURDIR)/$(DNP3_ARM_DIR)/lib -lopendnp3 -l:libstdc++.a -lpthread -lm -ldl -static-libgcc

# Container
IMAGE_ICR   ?= localhost/gomqttdnp3:icr323x
TARBALL_ICR ?= goMqttDnp3-icr323x.tar

# ── ICR-3232 device + deploy (see ICR3232_Dev_Reference.md) ──────────
# The target is BusyBox-init (no systemd, no Docker, no package manager).
# User apps + configs + logs live under /root, which persists via OverlayFS;
# /opt is firmware-reserved and /tmp + /var are volatile. Override DEVICE
# with the router IP, e.g. `make deploy-icr DEVICE=192.168.1.1`.
DEVICE       ?= 192.168.1.1
DEVICE_USER  ?= root
ICR_BIN_DIR  ?= /root/bin
ICR_ETC_DIR  ?= /root/etc
ICR_LOG_DIR  ?= /root/log

# Field-device simulators (see scripts/sim/README.md)
SIM_DNP3_PORT    ?= 20100
SIM_MODBUS_PORT  ?= 1502
SIM_DNP3_BIN     := /tmp/outstation_sim
SIM_DNP3_WIN_BIN := outstation_sim.exe

.PHONY: build build-ffi build-ffi-noembed build-ffi-icr \
        run run-ffi linux windows icr323x icr323x-ffi \
        ui-install ui-build ui-dev \
        web web-dev dev \
        image-icr323x image-save-icr323x \
        check-dnp3-host check-dnp3-arm check-arm-toolchain \
        check-dnp3-windows check-mingw-toolchain \
        opendnp3-vendor opendnp3-vendor-arm opendnp3-vendor-windows \
        sim-dnp3 sim-dnp3-build sim-dnp3-windows sim-modbus dev-cert \
        deploy-icr deploy-icr-ffi service-icr verify-arm \
        test clean

# ── Stub builds (no DNP3 lib needed; emits no measurements) ──────────

# Default dev build: pure Go, default stub master.
build:
	go build -o $(BINARY) .

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags embed -trimpath -ldflags="-s -w" -o $(BINARY) .

windows: ui-build
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -tags embed -trimpath -ldflags="-s -w" -o $(BINARY) .

# Cross-compile (stub) for ICR-323x (linux/arm/v7) with embedded UI.
icr323x: ui-build
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
		go build -tags embed,icr -trimpath -ldflags="-s -w" -o $(BINARY) .

# ── DNP3 FFI builds (require vendored opendnp3; run make opendnp3-vendor) ──

# Host build with real DNP3 master + embedded UI (so browsing / serves the SPA).
# For dev with hot-reload, use `make build-ffi-noembed` alongside `make ui-dev`.
build-ffi: check-dnp3-host ui-build
	CGO_ENABLED=1 \
	CGO_CXXFLAGS="$(DNP3_HOST_CXXFLAGS)" \
	CGO_LDFLAGS="$(DNP3_HOST_LDFLAGS)" \
	go build -tags embed,dnp3_ffi -trimpath -o $(BINARY) .

# Same as build-ffi but without the embed tag; serves no static files at /.
# Use with `make ui-dev` (Vite on :5173 proxies API calls to :8080).
build-ffi-noembed: check-dnp3-host
	CGO_ENABLED=1 \
	CGO_CXXFLAGS="$(DNP3_HOST_CXXFLAGS)" \
	CGO_LDFLAGS="$(DNP3_HOST_LDFLAGS)" \
	go build -tags dnp3_ffi -trimpath -o $(BINARY) .

run-ffi: build-ffi
	./$(BINARY) -port $(PORT) -config $(CONFIG) -log $(LOG) $(TLS_FLAGS)

# Cross-compile with real DNP3 master for ICR-323x. Modbus + sysmon are pure-Go
# (always in); the dnp3_ffi tag links the real opendnp3 master. We FULLY static-
# link (-extldflags -static) because cgo otherwise pulls the build host's glibc
# (Debian links GLIBC_2.38), which is far newer than the ICR firmware userland —
# a dynamic binary would die with "GLIBC_2.xx not found". Static = self-contained,
# runs on the device regardless of its glibc. netgo gives Go a pure-Go DNS resolver
# so MQTT hostname lookups don't need glibc NSS. (A DNP3 outstation set by hostname
# would still want glibc NSS via asio's getaddrinfo — configure outstations by IP.)
icr323x-ffi: check-dnp3-arm check-arm-toolchain ui-build
	CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 \
	CC=arm-linux-gnueabihf-gcc \
	CXX=arm-linux-gnueabihf-g++ \
	CGO_CXXFLAGS="$(DNP3_ARM_CXXFLAGS)" \
	CGO_LDFLAGS="$(DNP3_ARM_LDFLAGS)" \
	go build -tags embed,dnp3_ffi,netgo,icr -trimpath -ldflags="-s -w -extldflags '-static'" -o $(BINARY) .
	@echo ""
	@echo "Deploy notes for $(BINARY) on ICR-3232 (opendnp3 is static-linked; no .so needed):"
	@echo "  make deploy-icr-ffi DEVICE=<ip>  # build, then print scp/ssh steps for $(ICR_BIN_DIR) / $(ICR_ETC_DIR)"
	@echo "  make service-icr DEVICE=<ip>     # generate a BusyBox init.d service + print install steps"

# Alias for consistency.
build-ffi-icr: icr323x-ffi

# ── ICR-3232 deploy (manual — these targets only build + print steps) ──
#
# The device has no package manager and we keep credentials out of the build:
# deploy-icr / deploy-icr-ffi build a verified ARM binary, then PRINT the
# scp/ssh commands to run by hand. Nothing is pushed automatically. Override
# DEVICE=<ip> so the printed commands are copy-paste ready. deploy-icr uses the
# pure-Go stub build (host telemetry, no field protocols); deploy-icr-ffi ships
# the real DNP3 master. Target tree is the OverlayFS-persisted /root (§7).

# Confirm the build really is a static ARM ELF before shipping (reference §6).
verify-arm:
	@file $(BINARY) | grep -q "ARM" || { echo "ERROR: $(BINARY) is not an ARM binary — build with 'make icr323x' first"; exit 1; }
	@file $(BINARY) | grep -q "statically linked" || echo "WARN: $(BINARY) is not statically linked — check CGO/toolchain"
	@file $(BINARY); ls -lh $(BINARY)

define print_deploy_steps
	@echo ""
	@echo "Built $(BINARY) for ICR-3232. Copy it to the device by hand (/root persists via OverlayFS):"
	@echo ""
	@echo "  # 1. create dirs, then copy binary + config"
	@echo "  ssh $(DEVICE_USER)@$(DEVICE) 'mkdir -p $(ICR_BIN_DIR) $(ICR_ETC_DIR) $(ICR_LOG_DIR)'"
	@echo "  scp $(BINARY) $(DEVICE_USER)@$(DEVICE):$(ICR_BIN_DIR)/$(BINARY)"
	@echo "  scp $(CONFIG) $(DEVICE_USER)@$(DEVICE):$(ICR_ETC_DIR)/$(BINARY).json"
	@echo ""
	@echo "  # 2. make executable + test run"
	@echo "  ssh $(DEVICE_USER)@$(DEVICE) 'chmod +x $(ICR_BIN_DIR)/$(BINARY)'"
	@echo "  ssh $(DEVICE_USER)@$(DEVICE) '$(ICR_BIN_DIR)/$(BINARY) -config $(ICR_ETC_DIR)/$(BINARY).json -port $(PORT) -log $(LOG)'"
	@echo ""
	@echo "  # 3. (optional) generate + install a boot service:  make service-icr"
endef

deploy-icr: icr323x verify-arm
	$(print_deploy_steps)

deploy-icr-ffi: icr323x-ffi verify-arm
	$(print_deploy_steps)

# Generate a BusyBox init.d service (start/stop/restart/status) so the gateway
# survives reboots. service-icr only writes the script locally and prints the
# manual install steps — nothing is copied to the device.
define ICR_INITD
#!/bin/sh
# $(BINARY) — Modbus/DNP3 → MQTT Sparkplug B gateway (BusyBox init.d)
# Generated by `make service-icr`; see ICR3232_Dev_Reference.md §5.
#
# Plain POSIX shell — the ICR firmware's BusyBox has no start-stop-daemon
# applet, so we manage the process with nohup + a PID file directly.
DAEMON=$(ICR_BIN_DIR)/$(BINARY)
RUNDIR=/root/run
PIDFILE=$$RUNDIR/$(BINARY).pid
LOGFILE=$(ICR_LOG_DIR)/$(BINARY).log
ARGS="-config $(ICR_ETC_DIR)/$(BINARY).json -port $(PORT) -log $(LOG)"

running() {
    [ -f "$$PIDFILE" ] && kill -0 "$$(cat "$$PIDFILE")" 2>/dev/null
}
start() {
    if running; then
        echo "$(BINARY) already running (PID $$(cat "$$PIDFILE"))"
        return 0
    fi
    mkdir -p "$$RUNDIR" $(ICR_LOG_DIR)
    echo "Starting $(BINARY)..."
    nohup "$$DAEMON" $$ARGS >> "$$LOGFILE" 2>&1 &
    echo $$! > "$$PIDFILE"
    echo "OK (PID $$(cat "$$PIDFILE"))"
}
stop() {
    echo "Stopping $(BINARY)..."
    if [ -f "$$PIDFILE" ]; then
        PID=$$(cat "$$PIDFILE")
        kill "$$PID" 2>/dev/null
        i=0
        while kill -0 "$$PID" 2>/dev/null && [ $$i -lt 10 ]; do sleep 1; i=$$((i+1)); done
        kill -9 "$$PID" 2>/dev/null
        rm -f "$$PIDFILE"
    fi
    echo "OK"
}
case "$$1" in
    start)   start ;;
    stop)    stop ;;
    restart) stop; sleep 2; start ;;
    status)
        if running; then
            echo "$(BINARY) running (PID $$(cat "$$PIDFILE"))"
        else
            echo "$(BINARY) stopped"
        fi ;;
    *) echo "Usage: $$0 {start|stop|restart|status}" ;;
esac
endef

$(BINARY).init: Makefile
	$(file >$@,$(ICR_INITD))
	@echo "wrote $@"

service-icr: $(BINARY).init
	@echo ""
	@echo "Wrote $(BINARY).init (BusyBox init.d service). Install it by hand:"
	@echo ""
	@echo "  scp $(BINARY).init $(DEVICE_USER)@$(DEVICE):/etc/init.d/$(BINARY)"
	@echo "  ssh $(DEVICE_USER)@$(DEVICE) 'chmod +x /etc/init.d/$(BINARY)'"
	@echo "  ssh $(DEVICE_USER)@$(DEVICE) '/etc/init.d/$(BINARY) start'"
	@echo ""
	@echo "  # start on boot: add this line to /etc/rc.local, before the final 'exit 0':"
	@echo "      /etc/init.d/$(BINARY) start"

# ── Run ──────────────────────────────────────────────────────────────

run: ui-build
	go build -tags embed -o $(BINARY) .
	./$(BINARY) -port $(PORT) -config $(CONFIG) -log $(LOG) $(TLS_FLAGS)

web: run

web-dev: build
	./$(BINARY) -port $(PORT) -config $(CONFIG) -log debug

dev: web-dev

# ── Tests ────────────────────────────────────────────────────────────

test:
	go test ./...

# ── Vue frontend ─────────────────────────────────────────────────────

ui-install:
	cd web && pnpm install

ui-build:
	cd web && pnpm run build

ui-dev:
	cd web && pnpm run dev

# ── Container (ICR-323x, linux/arm/v7) ──────────────────────────────

image-icr323x: ui-build icr323x-ffi
	docker build --platform linux/arm/v7 -t $(IMAGE_ICR) .

image-save-icr323x: image-icr323x
	docker save $(IMAGE_ICR) -o $(TARBALL_ICR)
	@echo "Saved $(IMAGE_ICR) → $(TARBALL_ICR)"

# ── Preflight checks ─────────────────────────────────────────────────

check-dnp3-host:
	@if [ ! -f $(DNP3_HOST_DIR)/include/opendnp3/DNP3Manager.h ] || [ ! -f $(DNP3_HOST_DIR)/lib/libopendnp3.a ]; then \
		echo "ERROR: missing $(DNP3_HOST_DIR)/{include/opendnp3/DNP3Manager.h,lib/libopendnp3.a}"; \
		echo "  Run: make opendnp3-vendor"; \
		exit 1; \
	fi

check-dnp3-arm:
	@if [ ! -f $(DNP3_ARM_DIR)/include/opendnp3/DNP3Manager.h ] || [ ! -f $(DNP3_ARM_DIR)/lib/libopendnp3.a ]; then \
		echo "ERROR: missing $(DNP3_ARM_DIR)/{include/opendnp3/DNP3Manager.h,lib/libopendnp3.a}"; \
		echo "  Run: make opendnp3-vendor-arm"; \
		exit 1; \
	fi

check-arm-toolchain:
	@if ! command -v arm-linux-gnueabihf-gcc >/dev/null 2>&1; then \
		echo "ERROR: arm-linux-gnueabihf-gcc not found."; \
		echo "  Install on Debian/Ubuntu: sudo apt install gcc-arm-linux-gnueabihf"; \
		exit 1; \
	fi

check-dnp3-windows:
	@if [ ! -f $(DNP3_WIN_DIR)/include/opendnp3/DNP3Manager.h ] || [ ! -f $(DNP3_WIN_DIR)/lib/libopendnp3.a ]; then \
		echo "ERROR: missing $(DNP3_WIN_DIR)/{include/opendnp3/DNP3Manager.h,lib/libopendnp3.a}"; \
		echo "  Run: make opendnp3-vendor-windows"; \
		exit 1; \
	fi

check-mingw-toolchain:
	@if ! command -v x86_64-w64-mingw32-g++ >/dev/null 2>&1; then \
		echo "ERROR: x86_64-w64-mingw32-g++ not found."; \
		echo "  Install on Debian/Ubuntu: sudo apt install g++-mingw-w64-x86-64"; \
		exit 1; \
	fi

# ── opendnp3 vendoring (Apache 2.0 DNP3 stack) ──────────────────────
#
# Run once per build host. Fetches opendnp3 source, builds static libs, and
# installs them under third_party/opendnp3/<triple>/. The CGO shim
# (dnp3/opendnp3_c.cpp) links against these.
#
# Requires on host: cmake, build-essential, libssl-dev (apt install).
# For ARMv7 cross: also gcc-arm-linux-gnueabihf + g++-arm-linux-gnueabihf,
# and OPENSSL_ROOT_DIR pointing at a static armv7 OpenSSL build (or accept
# the TLS-off fallback).

opendnp3-vendor:
	bash scripts/build-opendnp3.sh host

opendnp3-vendor-arm:
	bash scripts/build-opendnp3.sh armv7-linux

# Cross-build opendnp3 for Windows x64 (MinGW-w64, TLS off). Feeds sim-dnp3-windows.
opendnp3-vendor-windows:
	bash scripts/build-opendnp3.sh windows-mingw

# ── Field-device simulators (docs: scripts/sim/README.md) ────────────
#
# Each runs in the foreground (Ctrl-C to stop). Override ports with
# SIM_DNP3_PORT / SIM_MODBUS_PORT. Point a gateway config at them
# (see configs/example.json) and `make run-ffi` to exercise both.

# Compile the DNP3 outstation sim (C++; needs vendored opendnp3 + g++).
sim-dnp3-build: check-dnp3-host
	@command -v g++ >/dev/null || { echo "ERROR: g++ not found (apt install g++)"; exit 1; }
	g++ -std=c++17 -I$(DNP3_HOST_DIR)/include \
		scripts/sim/outstation_sim.cpp $(DNP3_HOST_DIR)/lib/libopendnp3.a \
		-lssl -lcrypto -lpthread -o $(SIM_DNP3_BIN)
	@echo "built $(SIM_DNP3_BIN)"

# Build + run the DNP3 outstation sim (outstation addr 1024, master addr 1).
sim-dnp3: sim-dnp3-build
	$(SIM_DNP3_BIN) $(SIM_DNP3_PORT)

# Cross-compile the DNP3 outstation sim as a self-contained Windows .exe
# (MinGW-w64 + the TLS-off Windows opendnp3, statically linked so it needs no
# MinGW runtime DLLs on the target). Copy $(SIM_DNP3_WIN_BIN) to Windows and run
# it there: `outstation_sim.exe 20100`. opendnp3's networking is ASIO/Winsock,
# hence -lws2_32; -lwsock32; the C++ runtime + winpthread are pulled statically.
sim-dnp3-windows: check-mingw-toolchain check-dnp3-windows
	x86_64-w64-mingw32-g++ -std=c++17 -D_WIN32_WINNT=0x0601 \
		-I$(DNP3_WIN_DIR)/include \
		scripts/sim/outstation_sim.cpp $(DNP3_WIN_DIR)/lib/libopendnp3.a \
		-static -static-libgcc -static-libstdc++ \
		-lws2_32 -lwsock32 -lpthread \
		-o $(SIM_DNP3_WIN_BIN)
	@echo "built $(SIM_DNP3_WIN_BIN) — copy to Windows and run: $(SIM_DNP3_WIN_BIN) $(SIM_DNP3_PORT)"

# Run the Modbus/TCP slave sim (pure Go; no opendnp3 needed).
sim-modbus:
	go run ./scripts/sim/modbusslave $(SIM_MODBUS_PORT)

# ── TLS helper ───────────────────────────────────────────────────────
#
# Generate a throwaway self-signed cert (cert.pem/key.pem) for local HTTPS.
# It encrypts traffic but browsers still warn (not CA-trusted). For a
# warning-free cert use mkcert or your internal CA — the SAN must list the
# host you browse to. Override the name via TLS_HOST.
TLS_HOST ?= localhost
dev-cert:
	openssl req -x509 -newkey rsa:2048 -nodes -keyout key.pem -out cert.pem -days 365 \
		-subj "/CN=$(TLS_HOST)" \
		-addext "subjectAltName=DNS:$(TLS_HOST),DNS:localhost,IP:127.0.0.1"
	@echo "wrote cert.pem + key.pem  →  make run-ffi TLS_CERT=cert.pem TLS_KEY=key.pem"

# ── Clean ────────────────────────────────────────────────────────────

clean:
	rm -f $(BINARY) $(BINARY).exe $(TARBALL_ICR) $(BINARY).init $(SIM_DNP3_WIN_BIN)
	rm -rf web/dist
