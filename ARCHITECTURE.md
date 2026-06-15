# goMqttDnp3 — Architecture & Code Status

> Industrial protocol gateway: polls **Modbus/TCP** and **Modbus RTU (RS-485)**
> devices and acts as a **DNP3** master, normalizes their measurements, and
> republishes them as **MQTT Sparkplug B** metrics. Ships with an embedded Vue web UI for
> configuration and live monitoring. Primary deployment target is the Advantech
> **ICR-3232** industrial router (linux/arm/v7); also builds for linux/amd64 and
> windows/amd64.

Module: `goMqttDnp3` · Go 1.24 · ~6.3k LOC Go + Vue 3 SPA.

---

## 1. High-level data flow

```
  ┌─────────┐ ┌──────────┐ ┌──────────┐ ┌─────────┐
  │  DNP3   │ │ Modbus   │ │ Modbus   │ │ sysmon  │  field / host sources
  │  master │ │ /TCP     │ │ RTU/485  │ │ collect │  (each a source.Source)
  └────┬────┘ └────┬─────┘ └────┬─────┘ └────┬────┘
       │ OnSample  │ OnSample   │ OnSample   │ Collect() (ticker)
       ▼           ▼            ▼            ▼
        ┌──────────────────────────────────────────┐
        │            publisher.Publisher            │
        │  ingest chan (non-blocking, 4096 buffer)  │
        │              │                            │
        │   mapping.Apply (scale/offset + quality)  │
        │              │                            │
        │   deadband filter (per-metric lastVal)    │
        │              │                            │
        │   Sparkplug encode → MQTT publish         │
        │              │         └─ offline buffer ─┤  (store-and-forward)
        └──────────────┼────────────────────────────┘
                       ▼
              MQTT broker (NBIRTH / NDATA / DDATA / NDEATH)
                       ▲
        ┌──────────────┴────────────────────────────┐
        │  api.Server (REST + WebSocket) + Vue SPA   │  config + live status
        │  config.Store (atomic JSON persistence)    │
        └────────────────────────────────────────────┘
```

The design is **event-driven**: field drivers push `source.Sample` values into a
bounded `ingest` channel; a single drain goroutine maps, filters, and publishes
them. A slow or unreachable broker never back-pressures the protocol stack —
samples are buffered offline or shed (counted in `Status.DroppedCount`).

---

## 2. Package map

| Package | Responsibility | Status |
|---|---|---|
| `main` | Flags, logger, TLS selection, embedded-UI serving, HTTP server lifecycle + graceful shutdown. | Stable |
| `api` | REST handlers, WebSocket hub, request/response types. Owns the gateway lifecycle (`start`/`stop`). | Stable |
| `config` | `AppConfig` schema, defaults, and `Store` (thread-safe, atomic JSON persistence). | Stable |
| `source` | Protocol-agnostic boundary: `Source`, `Handler`, `Sample`, `Quality`, `Status`, `PointType`. | Stable |
| `publisher` | Orchestrates sources → mapping → deadband → Sparkplug/MQTT; offline buffer; live status push. | Stable |
| `dnp3` | DNP3 master `Source`. Stub (default) or `dnp3_ffi` cgo binding to opendnp3. | Stub stable; FFI requires vendored lib |
| `modbus` | Modbus/TCP poller `Source`: per-device connection, typed register decoding, byte-order handling. Exports the transport-agnostic decode helpers (`Point`, `ResolvePoint`, `Read`, `DecodeSample`) reused by `modbusrtu`. | Stable |
| `modbusrtu` | Modbus RTU (serial/RS-485) `Source`. Groups slaves by serial port into a half-duplex **bus** (one goroutine per port, serialized transactions); per-slave cadence/timeout/retry; optional `TIOCSRS485` direction control. Reuses `modbus`'s decode. | Stable |
| `mapping` | `Apply` — converts a `Sample` to a Sparkplug `Metric` (scale/offset, quality properties); deadband. Modbus/TCP and RTU share one formatting path. | Stable |
| `sparkplug` | Hand-rolled Sparkplug B v1.0 protobuf encoder, NBIRTH/NDATA/NDEATH, alias registry, topic builder. | Stable |
| `sysmon` | Host telemetry collector (gopsutil) + ICR-323x vendor sensors (`icr` build tag). | Stable |
| `cmd/dnp3smoke`, `cmd/modbussmoke`, `cmd/rtusmoke` | Manual smoke-test entry points. `rtusmoke` allocates a PTY pair and runs an in-process RTU slave — no hardware needed (Linux). | Dev tooling |
| `scripts/sim` | Field-device simulators: DNP3 outstation (C++/opendnp3) and Modbus slave (Go). | Dev tooling |

### Source abstraction

Adding a protocol means implementing `source.Source` (lifecycle: `Start`/`Stop`/
`Status`) and pushing `source.Sample` to a `source.Handler`. The
mapping/Sparkplug/MQTT layer is untouched. `Quality` uses DNP3 (IEEE 1815)
semantics as the common model; Modbus maps a good read to `QualityOnline`.

---

## 3. Build configurations (build tags)

| Tag | Effect |
|---|---|
| *(none)* | Stub DNP3 master (no C deps). UI not embedded — `/` serves nothing. |
| `embed` | Embeds `web/dist` (built SPA) into the binary via `embed_prod.go`. |
| `dnp3_ffi` | Links the real opendnp3 master via cgo (`dnp3/opendnp3_c.cpp`). Requires `make opendnp3-vendor`. |
| `icr` | Enables the ICR-323x vendor sensor collector (`sysmon/vendor_icr.go`); reads `status -v sys`. |
| `netgo` | Pure-Go DNS resolver (used in the fully-static ICR cross-build). |

`embed_dev.go` (no `embed` tag) returns a nil FS so dev builds rely on the Vite
dev server. See the `Makefile` (`make help`) for the canonical build matrix.

---

## 4. Runtime model & concurrency

- **HTTP server** (`main`): one `http.Server` with `ReadHeaderTimeout`,
  `ReadTimeout`, `IdleTimeout` set. SIGINT/SIGTERM trigger graceful shutdown:
  stop the publisher (NDEATH + clean MQTT disconnect) then `http.Server.Shutdown`.
- **Publisher** owns: one `ingestLoop` goroutine (drains `ingest`), one `sysLoop`
  goroutine (host-telemetry ticker), and the MQTT client's own callback threads.
  Counters are `atomic`; `lastVal`/`osStatus`/`sysCollector` are mutex-guarded.
- **DNP3 master (FFI)**: opendnp3 delivers measurements on internal "strand"
  threads → `OnSample`. **Modbus/TCP poller**: one goroutine per device on a
  scan-rate ticker. **Modbus RTU poller**: one goroutine per *serial port* (the
  RS-485 bus is half-duplex, so transactions across the slaves on a port are
  serialized; each slave keeps its own cadence). All feed the same `ingest`
  channel via a non-blocking send.
- **WebSocket hub**: tracks connected clients; `Broadcast` serializes writes
  under a mutex (gorilla requires serialized writes per connection).
- **Config store**: `sync.RWMutex`; `Get()` returns a **deep copy** of the
  config so readers never race with in-place mutators.

---

## 5. Configuration model

`AppConfig` (persisted as pretty-printed JSON, written atomically via temp file +
rename) has these sections:

- **`mqtt`** — broker URL, client ID, credentials, QoS, keepalive, optional TLS
  (`caFile`/`certFile`/`keyFile`/`insecure`).
- **`sparkplug`** — `groupId`, `nodeId`, `birthOnConfigChange`.
- **`outstations`** — DNP3 outstations (host/port, link addresses, poll cadences,
  unsolicited config, startup-integrity class selection, static-poll period).
- **`modbusDevices`** — Modbus/TCP devices (host/port, unit ID, scan rate,
  timeout, retries).
- **`serialDevices`** — Modbus RTU slaves on an RS-485/RS-232 port (`port` e.g.
  `/dev/ttyS1`, `baudRate`/`dataBits`/`parity`/`stopBits`, `unitId`, cadence/
  timeout/retries, optional `rs485` `TIOCSRS485` direction control). Devices
  sharing a `port` form one multidrop bus; line settings come from the first.
- **`mappings`** — `SignalMapping` per point. Carries a `protocol` discriminator
  (`dnp3`/`modbus`/`modbusrtu`), source reference (`sourceId`, legacy
  `outstationId`), point
  identity, engineering transform (scale/offset/unit), and publish controls
  (deadband, publish-on-poll).
- **`system`** — host telemetry: enable flag, interval, metric prefix, mounts,
  interfaces, per-group `metrics` toggles, and `disabledMetrics` opt-outs.

Defaults live in `config.DefaultAppConfig()`; validation lives in `api/handlers.go`
(`validateMQTT`, `validateOutstation`, `validateModbusDevice`,
`validateSerialDevice`, `validateMapping`, `validateSystem`).

---

## 6. HTTP / WebSocket API

All REST responses use the envelope `{ "success": bool, "message"?: string,
"data"?: any }`. CORS is open (`Access-Control-Allow-Origin: *`).

| Method(s) | Path | Purpose |
|---|---|---|
| GET | `/api/health` | Liveness + UTC time. |
| GET | `/api/status` | Publisher runtime `Status`. |
| GET | `/api/config` | Full `AppConfig`. |
| GET/PUT | `/api/config/mqtt` | MQTT section. |
| GET/PUT | `/api/config/sparkplug` | Sparkplug section. |
| GET/PUT | `/api/config/system` | Host-telemetry section (hot-applied when running). |
| GET | `/api/system/telemetry` | Flat host-telemetry snapshot (reuses publisher's sample when running; samples on demand otherwise — works with the broker down). |
| GET/POST | `/api/outstations` | List / upsert DNP3 outstations. |
| PUT/DELETE | `/api/outstations/{id}` | Update / delete (cascades to mappings). |
| GET/POST | `/api/modbusDevices` | List / upsert Modbus/TCP devices. |
| PUT/DELETE | `/api/modbusDevices/{id}` | Update / delete (cascades to mappings). |
| GET/POST | `/api/serialDevices` | List / upsert Modbus RTU (RS-485) slaves. |
| PUT/DELETE | `/api/serialDevices/{id}` | Update / delete (cascades to mappings). |
| GET/POST | `/api/mappings` | List / add mappings. |
| GET | `/api/mappings/export` | Download mappings JSON. |
| POST | `/api/mappings/import` | Replace all mappings (validated). |
| PUT/DELETE | `/api/mappings/{id}` | Update / delete a mapping. |
| POST | `/api/gateway/start` | Build a fresh publisher from current config and start it. |
| POST | `/api/gateway/stop` | Stop the running publisher. |
| GET (upgrade) | `/ws` | WebSocket: pushes `log` and `status` events. |
| GET | `/` | Embedded SPA (when built with `embed`). |

`gateway/start` never blocks on the broker: MQTT connects in the background with
retry, and the field sources start polling immediately so an edge gateway
collects data even when the broker is unreachable (cloud broker, link down, not
yet provisioned). Samples accumulate in the store-and-forward buffer; the
`OnConnect` handler publishes NBIRTH/DBIRTH, (re)subscribes NCMD, and drains the
buffer once the broker is reached — and it re-runs on every auto-reconnect, so a
fresh birth follows each reconnection. `Status.mqttConnected` reflects the gateway's
own connect/disconnect tracking, not paho's `IsConnected()` (which reports true
while merely retrying).

---

## 7. Sparkplug B notes

`sparkplug` is a self-contained Sparkplug B v1.0 implementation (no external
protobuf dependency): hand-written wire encoder (`proto.go`), node/device birth
& death (`node.go`), metric-name→alias registry for NDATA compression
(`registry.go`), topic builder (`topic.go`), and datatype constants
(`datatype.go`). `bdSeq` continuity across reconnects is preserved within a
process lifetime. Point quality is attached as metric properties
(`quality`/`dnp3.*`/`engUnit`) so receivers interpret it without out-of-band
knowledge.

---

## 8. Deployment — ICR-3232

The target runs BusyBox init (no systemd, no Docker, no package manager); `/root`
persists via OverlayFS. Build a fully-static arm/v7 binary, copy it to `/root`,
and install a BusyBox-friendly init script:

```sh
make icr323x-ffi               # or icr323x (stub, no DNP3 lib)
make deploy-icr-ffi DEVICE=<ip>   # prints scp/ssh steps (nothing auto-pushed)
make service-icr DEVICE=<ip>      # generates /etc/init.d script + install steps
```

The generated init script uses plain `nohup` + PID file (the firmware's BusyBox
has **no** `start-stop-daemon` applet). Boot hook: add `/etc/init.d/goMqttDnp3
start` to `/etc/rc.local`. The full device reference is `ICR3232_Dev_Reference.md`.

---

## 9. Security posture & known limitations

The gateway is designed as a **trusted-LAN appliance**. The following are
deliberate trade-offs to document for any internet-adjacent deployment:

| Item | Status | Note |
|---|---|---|
| HTTP request timeouts | **Fixed** | `ReadHeaderTimeout`/`ReadTimeout`/`IdleTimeout` set (Slowloris mitigation). |
| Graceful shutdown | **Fixed** | SIGTERM stops the publisher (NDEATH) and drains HTTP. |
| Config read/write race | **Fixed** | `Store.Get()` deep-copies; verified under `-race`. |
| **No API authentication** | Known | All REST/WS endpoints are unauthenticated. Bind to a trusted network or front with a reverse proxy + auth. |
| **WebSocket `CheckOrigin` allows all origins** | Known | No auth means any page could open `/ws`. Acceptable on a trusted LAN; tighten if exposed. |
| **MQTT password returned in plaintext** | Known | `GET /api/config[/mqtt]` echoes the stored password. Persisted file is `0640`. |
| TLS UI optional | By design | HTTPS only when `-tls-cert`/`-tls-key` are both set; one-without-the-other fails loudly. |
| MQTT TLS `insecure` flag | By design | `tls.insecure` skips cert verification — dev only. |
| CORS `*` | By design | Convenience for the bundled SPA; restrict if multi-origin. |

There is no secret redaction, rate limiting, or audit logging — out of scope for
the current trusted-LAN model.

---

## 10. Testing

- `go test ./...` — unit tests currently cover `sysmon` (collector + ICR vendor
  parsing). Run with `-race` for the concurrency-sensitive paths.
- `cmd/dnp3smoke`, `cmd/modbussmoke` and `scripts/sim/*` provide end-to-end smoke
  testing against the bundled simulators (`make sim-dnp3`, `make sim-modbus`).
- The DNP3 FFI path requires `make opendnp3-vendor` and a `dnp3_ffi` build.
- **Full edge → gateway → TimescaleDB soak:** the end-to-end runbook and a
  one-command harness live in the **goGateway** repo — `docs/e2e-test-guide.md`
  and `scripts/soak.sh` (`scripts/soak.sh run 10`). It builds this producer,
  stands up a throwaway stack, autodiscovers + approves the edge, runs a timed
  soak, and tears down. Point its `EDGE_DIR` at this repo (the default assumes
  a sibling checkout).

**Coverage gaps worth closing:** `config.Store`, `mapping.Apply`, `modbus`
decode/byte-order, and the `api` handlers have no automated tests yet.

---

## 11. Code review — changes applied this pass

- **Removed container packaging**: deleted the stale `Dockerfile` (it referenced
  a wrong binary name) and the `image-*` Make targets / image vars.
- **Purged & hardened the Makefile**: dropped redundant alias targets
  (`web`/`web-dev`/`dev`/`build-ffi-icr`), added a self-documenting `help` target
  as the default goal.
- **Deleted dead code**: `sparkplug.StringPropertySet` (unreachable),
  `api.ReadingSnapshot` + `EventReadings` (unused), and a hand-rolled `itoa`
  reimplementation (replaced with `strconv.Itoa`).
- **Fixed a config data race**, **added HTTP server timeouts**, and **added
  graceful shutdown with NDEATH** (see §9).
- Verified: `go build ./...`, `go vet ./...`, `go test -race ./...`, and the
  `embed,icr` arm/v7 cross-build all pass.
