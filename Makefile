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
DNP3_HOST_DIR      := third_party/opendnp3/$(DNP3_HOST_TRIPLE)
DNP3_ARM_DIR       := third_party/opendnp3/$(DNP3_ARM_TRIPLE)

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

# Field-device simulators (see scripts/sim/README.md)
SIM_DNP3_PORT   ?= 20100
SIM_MODBUS_PORT ?= 1502
SIM_DNP3_BIN    := /tmp/outstation_sim

.PHONY: build build-ffi build-ffi-noembed build-ffi-icr \
        run run-ffi linux windows icr323x icr323x-ffi \
        ui-install ui-build ui-dev \
        web web-dev dev \
        image-icr323x image-save-icr323x \
        check-dnp3-host check-dnp3-arm check-arm-toolchain \
        opendnp3-vendor opendnp3-vendor-arm \
        sim-dnp3 sim-dnp3-build sim-modbus dev-cert \
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
		go build -tags embed -trimpath -ldflags="-s -w" -o $(BINARY) .

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

# Cross-compile with real DNP3 master for ICR-323x. Static-links libopendnp3.a
# (and libstdc++); the resulting binary is self-contained for DNP3.
icr323x-ffi: check-dnp3-arm check-arm-toolchain ui-build
	CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 \
	CC=arm-linux-gnueabihf-gcc \
	CXX=arm-linux-gnueabihf-g++ \
	CGO_CXXFLAGS="$(DNP3_ARM_CXXFLAGS)" \
	CGO_LDFLAGS="$(DNP3_ARM_LDFLAGS)" \
	go build -tags embed,dnp3_ffi -trimpath -ldflags="-s -w" -o $(BINARY) .
	@echo ""
	@echo "Deploy notes for $(BINARY) on ICR-3232:"
	@echo "  scp $(BINARY)   root@ICR:/usr/local/bin/   # opendnp3 is static-linked; no .so needed"

# Alias for consistency.
build-ffi-icr: icr323x-ffi

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
	rm -f $(BINARY) $(BINARY).exe $(TARBALL_ICR)
	rm -rf web/dist
