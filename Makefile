BINARY   := goMqttDnp3
CONFIG   ?= config.json
PORT     ?= 8080
LOG      ?= info

# DNP3 native lib (Step Function I/O libdnp3_ffi) — vendored under third_party/dnp3/{triple}/
DNP3_HOST_TRIPLE   ?= x86_64-unknown-linux-gnu
DNP3_ARM_TRIPLE    ?= armv7-unknown-linux-gnueabihf
DNP3_HOST_DIR      := third_party/dnp3/$(DNP3_HOST_TRIPLE)
DNP3_ARM_DIR       := third_party/dnp3/$(DNP3_ARM_TRIPLE)

# Container
IMAGE_ICR   ?= localhost/gomqttdnp3:icr323x
TARBALL_ICR ?= goMqttDnp3-icr323x.tar

.PHONY: build build-ffi build-ffi-noembed build-ffi-icr \
        run run-ffi linux windows icr323x icr323x-ffi \
        ui-install ui-build ui-dev \
        web web-dev dev \
        image-icr323x image-save-icr323x \
        check-dnp3-host check-dnp3-arm check-arm-toolchain \
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

# ── DNP3 FFI builds (require vendored libdnp3_ffi) ───────────────────

# Host build with real DNP3 master + embedded UI (so browsing / serves the SPA).
# For dev with hot-reload, use `make build-ffi-noembed` alongside `make ui-dev`.
build-ffi: check-dnp3-host ui-build
	CGO_ENABLED=1 \
	CGO_CFLAGS="-I$(CURDIR)/$(DNP3_HOST_DIR)/include" \
	CGO_LDFLAGS="-L$(CURDIR)/$(DNP3_HOST_DIR)/lib -ldnp3_ffi -lpthread -ldl -lm -Wl,-rpath,$(CURDIR)/$(DNP3_HOST_DIR)/lib" \
	go build -tags embed,dnp3_ffi -trimpath -o $(BINARY) .

# Same as build-ffi but without the embed tag; serves no static files at /.
# Use with `make ui-dev` (Vite on :5173 proxies API calls to :8080).
build-ffi-noembed: check-dnp3-host
	CGO_ENABLED=1 \
	CGO_CFLAGS="-I$(CURDIR)/$(DNP3_HOST_DIR)/include" \
	CGO_LDFLAGS="-L$(CURDIR)/$(DNP3_HOST_DIR)/lib -ldnp3_ffi -lpthread -ldl -lm -Wl,-rpath,$(CURDIR)/$(DNP3_HOST_DIR)/lib" \
	go build -tags dnp3_ffi -trimpath -o $(BINARY) .

run-ffi: build-ffi
	./$(BINARY) -port $(PORT) -config $(CONFIG) -log $(LOG)

# Cross-compile with real DNP3 master for ICR-323x. Static-links libdnp3_ffi.a.
icr323x-ffi: check-dnp3-arm check-arm-toolchain ui-build
	CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 \
	CC=arm-linux-gnueabihf-gcc \
	CGO_CFLAGS="-I$(CURDIR)/$(DNP3_ARM_DIR)/include" \
	CGO_LDFLAGS="-L$(CURDIR)/$(DNP3_ARM_DIR)/lib -ldnp3_ffi -lpthread -ldl -lm -Wl,-rpath,/usr/local/lib" \
	go build -tags embed,dnp3_ffi -trimpath -ldflags="-s -w" -o $(BINARY) .
	@echo ""
	@echo "Deploy notes for $(BINARY) on ICR-3232:"
	@echo "  scp $(BINARY)                                            root@ICR:/usr/local/bin/"
	@echo "  scp $(DNP3_ARM_DIR)/lib/libdnp3_ffi.so                   root@ICR:/usr/local/lib/"
	@echo "  ssh root@ICR 'ldconfig'   # refresh dynamic linker cache"

# Alias for consistency.
build-ffi-icr: icr323x-ffi

# ── Run ──────────────────────────────────────────────────────────────

run: ui-build
	go build -tags embed -o $(BINARY) .
	./$(BINARY) -port $(PORT) -config $(CONFIG) -log $(LOG)

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
	@if [ ! -f $(DNP3_HOST_DIR)/include/dnp3.h ] || [ ! -f $(DNP3_HOST_DIR)/lib/libdnp3_ffi.so ]; then \
		echo "ERROR: missing $(DNP3_HOST_DIR)/{include/dnp3.h,lib/libdnp3_ffi.so}"; \
		echo "  Fetch from https://github.com/stepfunc/dnp3/releases (see README)."; \
		exit 1; \
	fi

check-dnp3-arm:
	@if [ ! -f $(DNP3_ARM_DIR)/include/dnp3.h ] || [ ! -f $(DNP3_ARM_DIR)/lib/libdnp3_ffi.so ]; then \
		echo "ERROR: missing $(DNP3_ARM_DIR)/{include/dnp3.h,lib/libdnp3_ffi.so}"; \
		echo "  Fetch from https://github.com/stepfunc/dnp3/releases (see README)."; \
		exit 1; \
	fi

check-arm-toolchain:
	@if ! command -v arm-linux-gnueabihf-gcc >/dev/null 2>&1; then \
		echo "ERROR: arm-linux-gnueabihf-gcc not found."; \
		echo "  Install on Debian/Ubuntu: sudo apt install gcc-arm-linux-gnueabihf"; \
		exit 1; \
	fi

# ── Clean ────────────────────────────────────────────────────────────

clean:
	rm -f $(BINARY) $(BINARY).exe $(TARBALL_ICR)
	rm -rf web/dist
