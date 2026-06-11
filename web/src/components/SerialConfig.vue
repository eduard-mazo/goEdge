<template>
  <div class="space-y-5">

    <div class="flex items-center gap-4">
      <p class="font-sans text-sm text-text-secondary flex-1">
        Esclavos Modbus RTU sobre <strong class="text-foreground font-semibold">RS-485</strong>. Los esclavos
        que comparten puerto forman un bus multipunto half-duplex; el gateway serializa cada transacción.
      </p>
      <button class="btn-primary text-xs py-1.5" @click="openAdd">+ Agregar</button>
    </div>

    <!-- Empty state -->
    <div v-if="!store.serialDevices.length" class="forge-panel text-center py-12">
      <div class="bus-empty-glyph mx-auto mb-4">
        <svg width="120" height="34" viewBox="0 0 120 34" fill="none">
          <line x1="6" y1="17" x2="114" y2="17" class="trunk-dim" />
          <g v-for="i in 4" :key="i">
            <line :x1="18 + (i - 1) * 28" y1="17" :x2="18 + (i - 1) * 28" y2="27" class="drop-dim" />
            <circle :cx="18 + (i - 1) * 28" cy="29" r="3" class="node-dim" />
          </g>
        </svg>
      </div>
      <p class="font-mono text-xs text-text-dim">Sin esclavos RS-485.</p>
      <button class="btn-ghost mt-4 text-xs" @click="openAdd">Agregar el primero</button>
    </div>

    <!-- One card per serial bus (= per port) -->
    <div v-else class="space-y-4">
      <div v-for="bus in buses" :key="bus.port" class="forge-panel bus-card overflow-hidden">

        <!-- Bus header: port identity + line settings + live multidrop glyph -->
        <div class="bus-header">
          <div class="flex items-center gap-3 min-w-0">
            <span class="port-badge">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="2" y="7" width="20" height="10" rx="2" /><path d="M7 7V5M12 7V5M17 7V5M9 17v2M15 17v2" />
              </svg>
              {{ bus.port }}
            </span>
            <span class="line-chip" :title="`${baudOf(bus.line)} baud · ${framingOf(bus.line)}`">
              {{ (baudOf(bus.line) / 1000).toFixed(baudOf(bus.line) % 1000 ? 1 : 0) }}k
              <span class="text-text-dim">·</span> {{ framingOf(bus.line) }}
            </span>
            <span class="de-chip" :class="busUsesSoftDE(bus) ? 'de-chip--sw' : 'de-chip--hw'"
                  :title="busUsesSoftDE(bus) ? 'DE/RE driven by software (TIOCSRS485)' : 'DE/RE driven in hardware / auto'">
              DE {{ busUsesSoftDE(bus) ? 'SW' : 'HW' }}
            </span>
          </div>

          <!-- Multidrop trunk glyph — one drop per slave, lit when enabled -->
          <div class="flex items-center gap-3">
            <svg class="bus-glyph" :width="glyphW(bus.devices.length)" height="30" :viewBox="`0 0 ${glyphW(bus.devices.length)} 30`" fill="none">
              <line x1="4" y1="11" :x2="glyphW(bus.devices.length) - 4" y2="11" class="trunk" />
              <g v-for="(d, i) in bus.devices" :key="d.id">
                <line :x1="12 + i * 22" y1="11" :x2="12 + i * 22" y2="20" :class="d.enabled ? 'drop' : 'drop-dim'" />
                <circle :cx="12 + i * 22" cy="23" r="3.5"
                        :class="nodeClass(d)" />
              </g>
            </svg>
            <span class="font-mono text-[10px] uppercase tracking-[0.14em] text-text-dim whitespace-nowrap">
              {{ bus.devices.length }} {{ bus.devices.length === 1 ? 'esclavo' : 'esclavos' }}
            </span>
          </div>
        </div>

        <!-- Slaves on this bus -->
        <div class="overflow-x-auto">
          <table class="min-w-[680px]">
            <thead>
              <tr>
                <th class="w-px"></th><th>Unidad</th><th>ID</th><th>Nombre</th>
                <th>Sondeo</th><th>Timeout</th><th>Estado</th><th></th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="d in bus.devices" :key="d.id">
                <td><span class="addr-tag">{{ d.unitId }}</span></td>
                <td class="font-mono text-[11px] text-text-secondary">#{{ d.unitId }}</td>
                <td class="font-mono text-xs text-text-secondary">{{ d.id }}</td>
                <td class="font-sans font-semibold text-sm">{{ d.label || '—' }}</td>
                <td class="font-mono text-[11px] text-text-secondary">{{ fmtMs(d.scanRateMs) }}</td>
                <td class="font-mono text-[11px] text-text-secondary">{{ fmtMs(d.timeoutMs) }}</td>
                <td>
                  <!-- Live link state when running, else the enabled flag -->
                  <span v-if="store.isRunning && liveOf(d)" class="flex items-center gap-1.5 font-mono text-[10px]"
                        :class="liveOf(d)!.connected ? 'text-[color:var(--signal-ok)]' : 'text-[color:var(--tk-red-bright)]'">
                    <span class="led" :class="liveOf(d)!.connected ? 'led--green' : 'led--red'" />
                    {{ liveOf(d)!.connected ? 'conectado' : 'sin resp.' }}
                    <span v-if="(liveOf(d)!.measurementsRx ?? 0) > 0" class="text-text-dim">
                      · {{ liveOf(d)!.measurementsRx }}
                    </span>
                  </span>
                  <span v-else :class="['signal-badge', d.enabled ? 'signal-badge--on' : 'signal-badge--off']">
                    {{ d.enabled ? 'activo' : 'inactivo' }}
                  </span>
                </td>
                <td>
                  <div class="flex gap-2 justify-end">
                    <button class="btn-ghost text-xs py-0.5 px-3" @click="openEdit(d)">Editar</button>
                    <button
                      class="btn-ghost text-xs py-0.5 px-3 !border-red-base/40 !text-red-base hover:!border-red-bright hover:!text-red-bright"
                      @click="del(d.id)"
                    >Borrar</button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="modal" class="modal-backdrop overflow-y-auto py-6">
        <div class="forge-panel w-full max-w-2xl mx-auto">
          <div class="forge-header flex items-center justify-between gap-3">
            <span>{{ editId ? 'Editar esclavo RS-485' : 'Nuevo esclavo RS-485' }}</span>
            <button type="button" class="modal-x" @click="modal = false" aria-label="Cerrar">✕</button>
          </div>
          <form @submit.prevent="saveDevice" class="p-5 space-y-4">

            <div class="grid grid-cols-2 gap-4">
              <Field label="ID (identificador único)">
                <input v-model="form.id" class="forge-input" :disabled="!!editId" required placeholder="flow-01" />
              </Field>
              <Field label="Nombre">
                <input v-model="form.label" class="forge-input" placeholder="Caudalímetro" />
              </Field>
            </div>

            <Field label="Puerto serial" hint="En el ICR-3232 el puerto RS-485 es /dev/ttyS1; los adaptadores USB son /dev/ttyUSB*">
              <input v-model="form.port" class="forge-input" list="serial-ports" required placeholder="/dev/ttyS1" />
              <datalist id="serial-ports">
                <option value="/dev/ttyS1">RS-485 (ICR-3232)</option>
                <option value="/dev/ttyS5" />
                <option value="/dev/ttyUSB6" /><option value="/dev/ttyUSB7" />
                <option value="/dev/ttyUSB8" /><option value="/dev/ttyUSB9" />
                <option value="COM3" />
              </datalist>
            </Field>

            <fieldset class="space-y-3">
              <legend class="font-mono text-[10px] uppercase tracking-[0.18em] text-text-dim">Parámetros de línea</legend>
              <p v-if="lineConflict" class="font-mono text-[11px] text-[color:var(--signal-warn)] leading-snug">
                ⚠ Otro esclavo ya usa <strong>{{ form.port }}</strong> a {{ lineConflict }}. El bus tiene un solo juego
                de parámetros — manda el primer esclavo; estos se ignorarán.
              </p>
              <div class="grid grid-cols-2 gap-4">
                <Field label="Baudios">
                  <select v-model.number="form.baudRate" class="forge-input">
                    <option v-for="b in bauds" :key="b" :value="b">{{ b }}</option>
                  </select>
                </Field>
                <Field label="Unit ID" hint="Dirección RTU 1–247">
                  <input v-model.number="form.unitId" class="forge-input" type="number" min="1" max="247" required />
                </Field>
                <Field label="Bits datos">
                  <select v-model.number="form.dataBits" class="forge-input">
                    <option v-for="n in [5, 6, 7, 8]" :key="n" :value="n">{{ n }}</option>
                  </select>
                </Field>
                <div class="grid grid-cols-2 gap-3">
                  <Field label="Paridad">
                    <select v-model="form.parity" class="forge-input">
                      <option value="N">Ninguna</option><option value="E">Par</option><option value="O">Impar</option>
                    </select>
                  </Field>
                  <Field label="Bits parada">
                    <select v-model.number="form.stopBits" class="forge-input">
                      <option :value="1">1</option><option :value="2">2</option>
                    </select>
                  </Field>
                </div>
              </div>
              <p class="font-mono text-[11px] text-text-dim">
                Trama: <strong class="text-amber-bright">{{ framingOf(form) }}</strong> @ {{ form.baudRate }} baudios.
              </p>
            </fieldset>

            <fieldset class="space-y-3">
              <legend class="font-mono text-[10px] uppercase tracking-[0.18em] text-text-dim">Sondeo</legend>
              <div class="grid grid-cols-2 gap-4">
                <Field label="Cadencia (ms)" hint="por defecto 1000">
                  <input v-model.number="form.scanRateMs" class="forge-input" type="number" placeholder="1000" />
                </Field>
                <Field label="Timeout (ms)" hint="por defecto 1000">
                  <input v-model.number="form.timeoutMs" class="forge-input" type="number" placeholder="1000" />
                </Field>
                <Field label="Reintentos" hint="por defecto 2">
                  <input v-model.number="form.retries" class="forge-input" type="number" min="0" placeholder="2" />
                </Field>
                <Field label="Espera reintento (ms)" hint="por defecto 200">
                  <input v-model.number="form.retryDelayMs" class="forge-input" type="number" placeholder="200" />
                </Field>
              </div>
            </fieldset>

            <!-- RS-485 direction control -->
            <fieldset class="rs485-box space-y-3">
              <legend class="font-mono text-[10px] uppercase tracking-[0.18em] text-amber-bright px-1">Dirección RS-485 (DE/RE)</legend>
              <label class="flex items-start gap-2.5 cursor-pointer">
                <input type="checkbox" v-model="form.rs485.enabled" class="mt-0.5" />
                <span class="font-sans text-sm text-text-secondary leading-snug">
                  Control DE/RE por software vía <span class="font-mono text-xs">TIOCSRS485</span>
                  <span class="block text-[11px] text-text-dim">
                    Déjalo apagado para <span class="font-mono">ttyS1</span> del ICR-3232 (el kernel maneja DE por hardware).
                    Actívalo solo para adaptadores USB que lo requieran.
                  </span>
                </span>
              </label>

              <Transition name="rs485-reveal">
                <div v-if="form.rs485.enabled" class="space-y-3 pl-7 pt-1">
                  <div class="flex flex-wrap gap-x-5 gap-y-2">
                    <label class="flex items-center gap-2 cursor-pointer font-sans text-[13px] text-text-secondary">
                      <input type="checkbox" v-model="form.rs485.rtsHighDuringSend" /> RTS alto al transmitir
                    </label>
                    <label class="flex items-center gap-2 cursor-pointer font-sans text-[13px] text-text-secondary">
                      <input type="checkbox" v-model="form.rs485.rtsHighAfterSend" /> RTS alto tras transmitir
                    </label>
                    <label class="flex items-center gap-2 cursor-pointer font-sans text-[13px] text-text-secondary">
                      <input type="checkbox" v-model="form.rs485.rxDuringTx" /> Rx durante Tx
                    </label>
                  </div>
                  <div class="grid grid-cols-2 gap-4">
                    <Field label="Espera antes de TX (µs)">
                      <input v-model.number="form.rs485.delayRtsBeforeSendUs" class="forge-input" type="number" min="0" placeholder="0" />
                    </Field>
                    <Field label="Espera tras TX (µs)">
                      <input v-model.number="form.rs485.delayRtsAfterSendUs" class="forge-input" type="number" min="0" placeholder="0" />
                    </Field>
                  </div>
                </div>
              </Transition>
            </fieldset>

            <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary pt-2 border-t border-border">
              <input type="checkbox" v-model="form.enabled" /> Esclavo habilitado
            </label>

            <p v-if="error" class="font-mono text-xs text-red-bright">⚠ {{ error }}</p>
            <div class="flex justify-end gap-3 pt-2 border-t border-border">
              <button type="button" class="btn-ghost text-xs" @click="modal = false">Cancelar</button>
              <button type="submit" class="btn-primary text-xs">Guardar</button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useGatewayStore } from '@/stores/gateway'
import { api, type SerialDevice, type OutstationStatus } from '@/api/client'
import Field from './Field.vue'

const store  = useGatewayStore()
const modal  = ref(false)
const editId = ref('')
const error  = ref('')

const bauds = [1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200]

const empty = (): SerialDevice => ({
  id: '', label: '',
  port: '/dev/ttyS1', baudRate: 9600, dataBits: 8, parity: 'N', stopBits: 1, unitId: 1,
  scanRateMs: 1000, timeoutMs: 1000, retries: 2, retryDelayMs: 200,
  rs485: { enabled: false, rtsHighDuringSend: true, rtsHighAfterSend: false, rxDuringTx: false, delayRtsBeforeSendUs: 0, delayRtsAfterSendUs: 0 },
  enabled: true,
})
const form = ref<SerialDevice>(empty())

// Group slaves into buses by port — mirrors the backend's per-port bus model.
const buses = computed(() => {
  const map = new Map<string, SerialDevice[]>()
  for (const d of store.serialDevices) {
    const key = d.port || '(unset)'
    if (!map.has(key)) map.set(key, [])
    map.get(key)!.push(d)
  }
  return [...map.entries()]
    .map(([port, devices]) => ({
      port,
      devices: [...devices].sort((a, b) => a.unitId - b.unitId),
      line: devices[0],
    }))
    .sort((a, b) => a.port.localeCompare(b.port))
})

// If editing/adding on a port that already has a bus with different line
// settings, warn — the first slave's line settings win at runtime.
const lineConflict = computed(() => {
  const existing = store.serialDevices.find(
    (d) => d.port === form.value.port && d.id !== form.value.id,
  )
  if (!existing) return ''
  if (baudOf(existing) === baudOf(form.value) && framingOf(existing) === framingOf(form.value)) return ''
  return `${baudOf(existing)} ${framingOf(existing)}`
})

const baudOf = (d: SerialDevice) => d.baudRate || 9600
const framingOf = (d: SerialDevice) => `${d.dataBits || 8}${d.parity || 'N'}${d.stopBits || 1}`
const busUsesSoftDE = (bus: { devices: SerialDevice[] }) => bus.devices.some((d) => d.rs485?.enabled)

const liveOf = (d: SerialDevice): OutstationStatus | undefined =>
  store.status?.outstations?.[d.id]

function nodeClass(d: SerialDevice) {
  if (store.isRunning) {
    const l = liveOf(d)
    if (l) return l.connected ? 'node node--live' : 'node node--down'
  }
  return d.enabled ? 'node' : 'node-dim'
}

const glyphW = (n: number) => Math.max(40, 24 + n * 22)

function openAdd() { editId.value = ''; form.value = empty(); error.value = ''; modal.value = true }
function openEdit(d: SerialDevice) {
  editId.value = d.id
  form.value = { ...empty(), ...d, rs485: { ...empty().rs485, ...d.rs485 } }
  error.value = ''
  modal.value = true
}

async function saveDevice() {
  error.value = ''
  try {
    if (editId.value) await api.updateSerialDevice(editId.value, form.value)
    else await api.addSerialDevice(form.value)
    await store.loadSerialDevices()
    modal.value = false
  } catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
}

async function del(id: string) {
  if (!confirm(`¿Eliminar el esclavo "${id}" y todas sus señales?`)) return
  await api.deleteSerialDevice(id)
  await store.loadSerialDevices()
  await store.loadMappings()
}

function fmtMs(ms?: number) {
  if (!ms) return '—'
  if (ms >= 1000) return (ms / 1000).toFixed(ms % 1000 ? 1 : 0) + 's'
  return ms + 'ms'
}
</script>

<style scoped>
.bus-card { padding: 0; }

.bus-header {
  display: flex; align-items: center; justify-content: space-between; gap: 1rem;
  padding: 0.65rem 0.9rem;
  border-bottom: 1.5px solid var(--border);
  background:
    linear-gradient(to right,
      color-mix(in srgb, var(--tk-amber-base) 9%, transparent),
      color-mix(in srgb, var(--tk-amber-base) 2%, transparent) 60%,
      transparent);
}

.port-badge {
  display: inline-flex; align-items: center; gap: 0.4rem;
  font-family: var(--font-mono); font-size: 12px; font-weight: 700;
  color: var(--tk-amber-bright);
  padding: 3px 9px; border-radius: 5px;
  border: 1px solid color-mix(in srgb, var(--tk-amber-base) 38%, transparent);
  background: color-mix(in srgb, var(--tk-amber-base) 10%, transparent);
}

.line-chip {
  font-family: var(--font-mono); font-size: 11px; font-weight: 600;
  color: var(--tk-text-secondary); letter-spacing: 0.02em;
}

.de-chip {
  font-family: var(--font-mono); font-size: 9px; font-weight: 700; letter-spacing: 0.1em;
  padding: 2px 6px; border-radius: 3px; border: 1px solid; text-transform: uppercase;
}
.de-chip--hw { color: var(--signal-ok); border-color: color-mix(in srgb, var(--signal-ok) 35%, transparent); background: color-mix(in srgb, var(--signal-ok) 8%, transparent); }
.de-chip--sw { color: var(--tk-amber-bright); border-color: color-mix(in srgb, var(--tk-amber-base) 40%, transparent); background: color-mix(in srgb, var(--tk-amber-base) 8%, transparent); }

/* multidrop trunk glyph */
.bus-glyph .trunk { stroke: color-mix(in srgb, var(--tk-amber-base) 55%, var(--border)); stroke-width: 2; stroke-linecap: round; }
.bus-glyph .trunk-dim, .bus-empty-glyph .trunk-dim { stroke: var(--border); stroke-width: 2; stroke-linecap: round; }
.bus-glyph .drop { stroke: color-mix(in srgb, var(--tk-amber-base) 50%, var(--border)); stroke-width: 1.5; }
.bus-glyph .drop-dim, .bus-empty-glyph .drop-dim { stroke: var(--border); stroke-width: 1.5; }
.bus-glyph .node { fill: var(--tk-amber-bright); }
.bus-glyph .node-dim, .bus-empty-glyph .node-dim { fill: none; stroke: var(--tk-border-bright); stroke-width: 1.5; }
.bus-glyph .node--live { fill: var(--signal-ok); animation: nodePulse 1.6s ease-in-out infinite; }
.bus-glyph .node--down { fill: none; stroke: var(--tk-red-bright); stroke-width: 1.75; }

@keyframes nodePulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.45; }
}

.addr-tag {
  display: inline-grid; place-items: center; min-width: 22px; height: 22px;
  font-family: var(--font-mono); font-size: 10px; font-weight: 700;
  color: var(--tk-amber-bright);
  border: 1px solid color-mix(in srgb, var(--tk-amber-base) 35%, transparent);
  border-radius: 4px; padding: 0 5px;
  background: color-mix(in srgb, var(--tk-amber-base) 7%, transparent);
}

.rs485-box {
  border: 1px solid color-mix(in srgb, var(--tk-amber-base) 28%, transparent);
  border-radius: var(--radius);
  padding: 0.85rem 0.9rem;
  background: color-mix(in srgb, var(--tk-amber-base) 4%, transparent);
}

.rs485-reveal-enter-active, .rs485-reveal-leave-active { transition: opacity 0.2s ease, transform 0.2s ease; }
.rs485-reveal-enter-from, .rs485-reveal-leave-to { opacity: 0; transform: translateY(-4px); }
</style>
