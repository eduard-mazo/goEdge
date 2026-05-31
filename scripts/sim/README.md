# Field-device simulators

Two simulators let you exercise the gateway end-to-end without real field
hardware — one per protocol. Each serves changing values so you can watch live
data flow through to MQTT/Sparkplug and the **Live Values** UI.

| Sim | Source | Language | Default port | `make` target |
|-----|--------|----------|--------------|---------------|
| DNP3 outstation | `outstation_sim.cpp`        | C++ (links vendored opendnp3) | `20100` | `make sim-dnp3` |
| Modbus/TCP slave | `modbusslave/main.go`      | pure Go (no deps)             | `1502`  | `make sim-modbus` |

Both run in the foreground; press `Ctrl-C` to stop. Override ports with
`SIM_DNP3_PORT` / `SIM_MODBUS_PORT`.

> The gateway does **not** start polling on boot — open the web UI and click
> **Start** (or `POST /api/gateway/start`). The publisher connects MQTT first,
> so a broker must be reachable.

---

## DNP3 outstation — `outstation_sim.cpp`

An [opendnp3](https://github.com/dnp3/opendnp3) TCP **outstation** (server). It
links the vendored `libopendnp3.a`, so it needs the host opendnp3 first:

```bash
make opendnp3-vendor      # once per host (fetches + builds opendnp3)
make sim-dnp3             # build + run on :20100
# or a different port:
make sim-dnp3 SIM_DNP3_PORT=20200
```

- **Link addressing:** outstation address `1024`, master address `1`.
- **Database:** 5 points of each type. On startup it seeds static values; every
  **1.5 s** it mutates `binary[0]`, `analog[0]`/`analog[1]`, `counter[0]`,
  `double-bit[0]`, and `octet[0]` — generating both static (integrity) data and
  change events (Class 1).
- **Unsolicited:** allowed (the master enables it per its config).

**Matching gateway outstation** (see `configs/example.json` → `rtu-01`):

```json
{ "id": "rtu-01", "host": "127.0.0.1", "port": 20100,
  "masterAddress": 1, "outstationAddress": 1024,
  "class1ScanMs": 2000, "startupIntegrity": true, "enabled": true }
```

Requires a gateway built with DNP3 support (`make build-ffi` / `run-ffi`).

---

## Modbus/TCP slave — `modbusslave/main.go`

A dependency-free Modbus/TCP **slave** (server). Pure Go, so no opendnp3/cgo is
needed — a Modbus-only gateway runs with the plain build.

```bash
make sim-modbus                       # run on :1502
make sim-modbus SIM_MODBUS_PORT=1503  # different port
# or directly:
go run ./scripts/sim/modbusslave 1502
```

Four register/bit banks, refreshed every **1 s**:

| Function | Addr | Type | Meaning | Range |
|----------|------|------|---------|-------|
| holding register (FC03) | 0–1 | float32 (ABCD) | tank level | 10.0–20.0 m |
| holding register (FC03) | 2   | uint16 | pump speed | ~1450 rpm |
| holding register (FC03) | 3   | int16  | temp ×10 | ~230 (23.0 °C) |
| holding register (FC03) | 4–5 | float32 (ABCD) | flow | 3.0–6.0 m³/h |
| input register (FC04)   | 0   | uint16 | pressure | ~101 kPa |
| coil (FC01)             | 0   | bool   | pump running | toggles |
| coil (FC01)             | 1   | bool   | valve open | toggles |
| discrete input (FC02)   | 0   | bool   | fault | occasional |

**Matching gateway device + mapping** (see `configs/example.json` → `plc-01`):

```json
{ "id": "plc-01", "host": "127.0.0.1", "port": 1502, "unitId": 1,
  "scanRateMs": 1000, "enabled": true }
```
```json
{ "protocol": "modbus", "sourceId": "plc-01", "function": "holding_register",
  "address": 0, "dataType": "float32", "byteOrder": "ABCD",
  "metricName": "plc/tank_level", "engineeringUnit": "m", "enabled": true }
```

---

## Run both at once

```bash
# terminal 1 — DNP3 outstation
make sim-dnp3
# terminal 2 — Modbus slave
make sim-modbus
# terminal 3 — MQTT broker (example)
docker run --rm -p 1883:1883 eclipse-mosquitto
# terminal 4 — gateway with embedded UI + both protocols
make run-ffi CONFIG=configs/example.json PORT=9099
```

Open <http://localhost:9099>, click **Start**, and watch the **Live Values**
tab: DNP3 points (`rtu/*`, tagged `DNP3`) and Modbus points (`plc/*`, tagged
`MB`) stream side by side.

`configs/example.json` is pre-wired to both sims. For a Modbus-only test (no
opendnp3 build), use a config with just `modbusDevices` and run
`make run CONFIG=<that>.json`.
