# Modbus protocol — how the gateway reads registers

This is the reference for the gateway's Modbus source: the frame anatomy, the
data-type/byte-order decoding, a worked end-to-end read, and the answer to
"can I read a float32's registers as int16?". Every claim here is pinned by a
test in `modbus/protocol_test.go` (decode/byte-order matrix + end-to-end against
an in-process slave) and `modbus/rtuovertcp_test.go`.

The same application PDU is used by all three transports the gateway speaks, so
decoding is identical across them — only the framing differs:

| Transport            | Package      | Framing                                  |
|----------------------|--------------|------------------------------------------|
| Modbus/TCP           | `modbus`     | MBAP header, **no** CRC                  |
| Modbus RTU over TCP  | `modbus`     | slave id + PDU + **CRC** over a TCP socket |
| Modbus RTU (serial)  | `modbusrtu`  | slave id + PDU + **CRC** over RS-485/232 |

Set `logFrames: true` on a device to dump every raw ADU (hex) to the gateway log
(the UI **Registro**); the frames below were captured that way.

---

## 1. Frame anatomy

### Modbus/TCP (read holding register, FC 0x03)

```
send  00 01 00 00 00 06 01 03 00 00 00 02
      └─tx─┘ └pro┘ └len┘ U  │  └addr┘ └qty┘
                            FC
recv  00 01 00 00 00 07 01 03 04 41 20 00 00
      └─tx─┘ └pro┘ └len┘ U  │  bc └─reg0─┘ └─reg1─┘
                            FC
```

- **tx** transaction id (echoed), **pro** protocol id (always 0), **len** byte
  count of what follows, **U** unit id.
- Request: function `03`, start address `0x0000`, quantity `0x0002` (2 registers).
- Response: byte count `04`, then registers `0x4120 0x0000` → float32 `10.0`.

### Modbus RTU / RTU-over-TCP (same read, real field device)

```
send  15 03 1b 5e 00 02 a0 29
      U  │  └addr┘ └qty┘ └CRC┘
         FC
recv  15 03 04 42 87 d0 2f 16 7f
      U  │  bc └─reg0─┘ └─reg1─┘ └CRC┘
         FC
```

- No MBAP header; instead the slave id leads (`0x15` = 21) and a 16-bit CRC
  (little-endian on the wire) trails. Address `0x1B5E` = 7006, qty 2.
- Registers `0x4287 0xD02F` → float32 `67.906609`.

### Exception response

When the slave rejects a request it sets the **high bit of the function code**
and returns a one-byte exception code:

```
recv  01 83 02
      U  │  └ exception 0x02 = ILLEGAL DATA ADDRESS
         FC = 0x03 | 0x80
```

Common codes: `01` illegal function, `02` illegal data address, `03` illegal
data value, `04` slave device failure. A 2-register read whose range runs past
the device's last mapped register returns `0x02` for the **whole** request — the
classic off-by-one symptom.

---

## 2. How a register is read, end to end

```
mapping (function, address, quantity, dataType, byteOrder, scale, offset)
   │  ResolvePoint:  quantity 0 → derived from dataType (see table)
   ▼
Read(client, point)        → issues FC03/04/01/02 for [address, address+quantity)
   ▼
raw response bytes         → big-endian per register, high register first
   ▼
reorder(raw, byteOrder)    → rearrange to canonical big-endian
   ▼
decodeNumeric(...)         → int16/uint16/int32/uint32/float32/float64 → float64
   ▼
scale × value + offset     → engineering value   (mapping/engine.go)
   ▼
source.Sample              → publisher → Sparkplug
```

`address` is the **0-based protocol address**. Bench tools and PLC docs usually
show **1-based Modicon points**, so Modicon point `7007` is protocol `address:
7006`. Quantity is in *registers*, not bytes.

---

## 3. Data types and register count

Set `quantity: 0` and the gateway derives the count from `dataType`:

| dataType            | registers | bytes | notes                          |
|---------------------|-----------|-------|--------------------------------|
| `int16` / `uint16`  | 1         | 2     | `""` defaults to int16         |
| `bool` (coil/discr.)| 1 bit     | —     | FC 01/02                       |
| `int32` / `uint32`  | 2         | 4     |                                |
| `float32`           | 2         | 4     | IEEE-754 single                |
| `float64`           | 4         | 8     | IEEE-754 double                |

Getting `quantity` wrong is the #1 config error: a `float32` with `quantity: 1`
reads only 2 bytes, the decoder needs 4, and the sample is **silently dropped**
(no value, no error).

---

## 4. Byte order

Modbus is big-endian per register, but multi-register values disagree on word
order across vendors. `byteOrder` rearranges the raw bytes `[A B C D]` to
canonical big-endian before decoding:

| byteOrder | result      | meaning                              |
|-----------|-------------|--------------------------------------|
| `ABCD`    | `A B C D`   | big-endian, high word first (default, Modicon float) |
| `DCBA`    | `D C B A`   | full byte reverse                    |
| `BADC`    | `B A D C`   | swap bytes within each register      |
| `CDAB`    | `C D A B`   | swap the two registers (word swap)   |

Example — the field float `0x4287 0xD02F`:

| byteOrder | decoded float32 |
|-----------|-----------------|
| `ABCD`    | **67.906609** ✅ |
| `CDAB`    | -11761490944    |
| `BADC`    | -1.46e-34       |
| `DCBA`    | 3.79e-10        |

Only one order yields a sane number; that's how you discover a device's order
when its docs don't say.

---

## 5. Can I read a float32 (2 registers) as int16?

**Yes, the wire doesn't care — but you get nonsense, not the value.** Each
register holds *half of the IEEE-754 bit pattern*, not a number.

Tank level `10.0 m` is stored ABCD as bits `0x41200000` → reg0 `0x4120`, reg1
`0x0000`:

| read as                  | result      |
|--------------------------|-------------|
| float32 (both registers) | **10.0**  ← the real value |
| reg0 as int16/uint16     | 16672  (`0x4120`, high half of the float bits) |
| reg1 as int16            | 0      (low half) |

So `16672` is *not* "tank level as an integer" — it's the top 16 bits of the
float encoding. Reading one register of a float32 as int16 is only meaningful if
the data was never a float to begin with (e.g. a scaled integer like `temp×10`
stored as a genuine int16 in a single register — see `holding[3]` below).

Rule of thumb:
- Value spans 2 registers and is IEEE-754 → read it as `float32`. Period.
- Value is a scaled integer in 1 register → read `int16`/`uint16` and apply
  `scale`/`offset` (e.g. `scale: 0.1` for temp×10).

(Verified in `TestFloat32RegistersAsInt16`.)

---

## 6. Test coverage

`modbus/protocol_test.go` mirrors the soak harness's `scripts/sim/modbusslave`
bank layout (`holding 0-1 float32, 2 uint16, 3 int16, 4-5 float32; input 0
uint16; coil 0/1; discrete 0`) and verifies:

| test                          | what it pins                                   |
|-------------------------------|------------------------------------------------|
| `TestReorder`                 | ABCD/DCBA/BADC/CDAB byte rearrangement         |
| `TestDecodeNumeric`           | every dataType + signed/unsigned edges + short-buffer failure |
| `TestDecodeSample`            | FC → PointType/value + quality                 |
| `TestResolvePointQuantity`    | register count auto-derived per dataType       |
| `TestFloat32RegistersAsInt16` | the float32-as-int16 question (§5)             |
| `TestModbusE2EAllFunctions`   | live read of FC 01/02/03/04, real frames       |
| `TestModbusE2EIllegalAddress` | out-of-range read → exception, no sample       |
| `TestRTUOverTCPDecodesFloat`  | RTU-over-TCP transport decodes the field float |

Run:

```bash
go test ./modbus/ -v          # protocol suite (prints captured frames)
go test -race ./modbus/       # concurrency-sensitive poller paths
```

Full edge → gateway → TimescaleDB soak (exercises the Modbus path under load):

```bash
# in the goGateway repo
scripts/soak.sh run 10        # builds this edge + modbusslave, soaks 10 min
```
