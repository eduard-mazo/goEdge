BINARY   := goMqttModbus
CONFIG   ?= config.json
PORT     ?= 8080
LOG      ?= info

# Container
IMAGE_ICR   ?= localhost/gomqttmodbus:icr323x
TARBALL_ICR ?= goMqttModbus-icr323x.tar

.PHONY: build run linux windows icr323x \
        ui-install ui-build ui-dev \
        web web-dev dev \
        image-icr323x image-save-icr323x \
        test clean

# ── Go builds ────────────────────────────────────────────────────────

build:
	go build -o $(BINARY) .

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags embed -trimpath -ldflags="-s -w" -o $(BINARY) .

windows: ui-build
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -tags embed -trimpath -ldflags="-s -w" -o $(BINARY) .

# Cross-compile for ICR-323x (linux/arm/v7, ICR_PLATFORM=v3) with embedded UI.
icr323x: ui-build
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
		go build -tags embed -trimpath -ldflags="-s -w" -o $(BINARY) .

# ── Run ──────────────────────────────────────────────────────────────

# Build UI + embed into binary, then run (default workflow).
run: ui-build
	go build -tags embed -o $(BINARY) .
	./$(BINARY) -port $(PORT) -config $(CONFIG) -log $(LOG)

# Alias kept for clarity.
web: run

# Go API only (no embed); run alongside `make ui-dev` for hot-reload dev.
web-dev: build
	./$(BINARY) -port $(PORT) -config $(CONFIG) -log debug

# One-command dev workflow:
#   Terminal 1: make dev      → Go API server on :8080 (no embed)
#   Terminal 2: make ui-dev   → Vite dev server on :5173 (proxies to :8080)
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

image-icr323x: ui-build
	CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
		go build -tags embed -trimpath -ldflags="-s -w" -o $(BINARY) .
	docker build --platform linux/arm/v7 -t $(IMAGE_ICR) .

image-save-icr323x: image-icr323x
	docker save $(IMAGE_ICR) -o $(TARBALL_ICR)
	@echo "Saved $(IMAGE_ICR) → $(TARBALL_ICR)"

# ── Clean ────────────────────────────────────────────────────────────

clean:
	rm -f $(BINARY) $(BINARY).exe $(TARBALL_ICR)
	rm -rf web/dist
