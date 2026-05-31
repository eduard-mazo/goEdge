<template>
  <div class="space-y-4">

    <!-- Toolbar -->
    <div class="flex items-center gap-3 flex-wrap">
      <p class="font-sans text-sm text-text-secondary flex-1">
        Point → Sparkplug B metric mappings (DNP3 &amp; Modbus)
      </p>
      <a :href="api.exportMappingsURL()" target="_blank" class="btn-ghost text-xs py-1">
        ↓ Export JSON
      </a>
      <label class="btn-ghost text-xs py-1 cursor-pointer">
        ↑ Import JSON
        <input type="file" accept=".json" class="hidden" @change="importFile" />
      </label>
      <button class="btn-primary text-xs py-1.5" @click="openAdd">+ Add Mapping</button>
    </div>

    <!-- Filter -->
    <input v-model="filter" class="forge-input max-w-sm"
           placeholder="Filter by metric, source, point…" />

    <!-- Empty -->
    <div v-if="!filtered.length" class="forge-panel text-center py-12">
      <p class="font-mono text-xs text-text-dim">
        {{ store.mappings.length ? 'No results for filter.' : 'No signal mappings yet.' }}
      </p>
    </div>

    <!-- Table -->
    <div v-else class="forge-panel overflow-x-auto">
      <table class="min-w-[960px]">
        <thead>
          <tr>
            <th>Metric Name</th>
            <th>Proto</th>
            <th>Source</th>
            <th>Point</th>
            <th>Scale</th>
            <th>EU</th>
            <th>Deadband</th>
            <th>Status</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in filtered" :key="m.id">
            <td class="font-mono text-xs font-medium text-foreground">{{ m.metricName }}</td>
            <td>
              <span class="proto-tag" :class="isModbus(m) ? 'proto-tag--mb' : 'proto-tag--dnp'">
                {{ isModbus(m) ? 'MB' : 'DNP3' }}
              </span>
            </td>
            <td class="font-mono text-[11px]">{{ srcId(m) }}</td>
            <td class="font-mono text-[11px] text-text-secondary whitespace-nowrap">{{ pointLabel(m) }}</td>
            <td class="font-mono text-[11px] text-text-secondary">
              {{ m.scale ?? 1 }}{{ m.offset ? ' +' + m.offset : '' }}
            </td>
            <td class="font-mono text-[11px] text-text-dim">{{ m.engineeringUnit || '—' }}</td>
            <td class="font-mono text-[11px] text-text-secondary">{{ m.deadband ?? 0 }}</td>
            <td>
              <span :class="['signal-badge', m.enabled ? 'signal-badge--on' : 'signal-badge--off']">
                {{ m.enabled ? 'ON' : 'OFF' }}
              </span>
            </td>
            <td>
              <div class="flex gap-1 justify-end">
                <button class="btn-ghost text-[10px] py-0.5 px-2" @click="openEdit(m)">Edit</button>
                <button
                  class="btn-ghost text-[10px] py-0.5 px-2 !border-red-base/40 !text-red-base hover:!border-red-bright hover:!text-red-bright"
                  @click="del(m.id)"
                >Del</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="modal" class="modal-backdrop overflow-y-auto py-6" @click.self="modal = false">
        <div class="forge-panel w-full max-w-2xl mx-auto">
          <div class="forge-header">{{ editId ? 'Edit Signal Mapping' : 'Add Signal Mapping' }}</div>
          <form @submit.prevent="saveMapping" class="p-5 space-y-4">

            <!-- Protocol switch -->
            <div class="seg">
              <button type="button" v-for="p in ['dnp3','modbus']" :key="p"
                      class="seg-btn" :class="form.protocol === p ? 'seg-btn--active' : ''"
                      @click="form.protocol = p as 'dnp3' | 'modbus'">
                {{ p === 'dnp3' ? 'DNP3' : 'Modbus' }}
              </button>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <Field label="Metric Name" hint="Unique Sparkplug B metric identifier">
                <input v-model="form.metricName" class="forge-input" placeholder="feeder/voltage_kv" required />
              </Field>
              <Field label="Sparkplug Device ID" hint="Leave blank to publish as node metric">
                <input v-model="form.deviceId" class="forge-input" placeholder="(node level)" />
              </Field>
            </div>

            <!-- Source select (protocol-specific list) -->
            <Field :label="form.protocol === 'modbus' ? 'Modbus Device' : 'Outstation'">
              <select v-model="form.sourceId" class="forge-input" required>
                <option value="" disabled>Select source…</option>
                <template v-if="form.protocol === 'modbus'">
                  <option v-for="d in store.modbusDevices" :key="d.id" :value="d.id">
                    {{ d.label || d.id }} ({{ d.host }}:{{ d.port }})
                  </option>
                </template>
                <template v-else>
                  <option v-for="o in store.outstations" :key="o.id" :value="o.id">
                    {{ o.label || o.id }} ({{ o.host }}:{{ o.port }})
                  </option>
                </template>
              </select>
            </Field>

            <!-- DNP3 point identity -->
            <template v-if="form.protocol !== 'modbus'">
              <div class="grid grid-cols-3 gap-4">
                <Field label="Point Type" hint="DNP3 group/variation" class="col-span-1">
                  <select v-model="form.pointType" class="forge-input">
                    <option v-for="pt in pointTypes" :key="pt.value" :value="pt.value">{{ pt.label }}</option>
                  </select>
                </Field>
                <Field label="Index" hint="0-based point index">
                  <input v-model.number="form.index" class="forge-input" type="number" min="0" max="65535" />
                </Field>
                <Field label="Event Class" hint="Informational">
                  <select v-model.number="form.eventClass" class="forge-input">
                    <option :value="0">Static only</option>
                    <option :value="1">Class 1</option>
                    <option :value="2">Class 2</option>
                    <option :value="3">Class 3</option>
                  </select>
                </Field>
              </div>
            </template>

            <!-- Modbus point identity -->
            <template v-else>
              <div class="grid grid-cols-2 gap-4">
                <Field label="Function">
                  <select v-model="form.function" class="forge-input">
                    <option v-for="f in modbusFunctions" :key="f.value" :value="f.value">{{ f.label }}</option>
                  </select>
                </Field>
                <Field label="Address" hint="Register / coil start address">
                  <input v-model.number="form.address" class="forge-input" type="number" min="0" max="65535" />
                </Field>
              </div>
              <div class="grid grid-cols-3 gap-4" v-if="isRegister">
                <Field label="Data Type">
                  <select v-model="form.dataType" class="forge-input">
                    <option v-for="dt in modbusDataTypes" :key="dt" :value="dt">{{ dt }}</option>
                  </select>
                </Field>
                <Field label="Byte Order" hint="word/byte order for 32/64-bit">
                  <select v-model="form.byteOrder" class="forge-input">
                    <option v-for="bo in byteOrders" :key="bo" :value="bo">{{ bo }}</option>
                  </select>
                </Field>
                <Field label="Quantity" hint="0 = derive from data type">
                  <input v-model.number="form.quantity" class="forge-input" type="number" min="0" max="125" placeholder="0" />
                </Field>
              </div>
            </template>

            <div class="grid grid-cols-3 gap-4">
              <Field label="Scale" hint="value × scale + offset">
                <input v-model.number="form.scale" class="forge-input" type="number" step="any" placeholder="1" />
              </Field>
              <Field label="Offset">
                <input v-model.number="form.offset" class="forge-input" type="number" step="any" placeholder="0" />
              </Field>
              <Field label="Eng. Unit">
                <input v-model="form.engineeringUnit" class="forge-input" placeholder="kV" />
              </Field>
            </div>

            <Field label="Deadband" hint="Min change (engineering units) to publish; 0 = always">
              <input v-model.number="form.deadband" class="forge-input max-w-[180px]" type="number" step="any" placeholder="0" />
            </Field>

            <label v-if="form.protocol !== 'modbus'" class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary">
              <input type="checkbox" v-model="form.publishOnPoll" />
              Publish on static reads too (not only on event)
            </label>
            <p v-else class="font-mono text-[11px] text-text-dim">
              Modbus reads are polled — every change is published (subject to deadband).
            </p>
            <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary">
              <input type="checkbox" v-model="form.enabled" /> Enabled
            </label>

            <p v-if="error" class="font-mono text-xs text-red-bright">⚠ {{ error }}</p>

            <div class="flex justify-end gap-3 pt-3 border-t border-border">
              <button type="button" class="btn-ghost text-xs" @click="modal = false">Cancel</button>
              <button type="submit" class="btn-primary text-xs">Save Mapping</button>
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
import { api, type SignalMapping, type PointType, type ModbusFunction, type ModbusDataType } from '@/api/client'
import Field from './Field.vue'

const store  = useGatewayStore()
const modal  = ref(false)
const editId = ref('')
const error  = ref('')
const filter = ref('')

const pointTypes: { value: PointType; label: string }[] = [
  { value: 'binary',               label: 'Binary Input (g1/g2)' },
  { value: 'double_bit_binary',    label: 'Double-Bit Binary (g3/g4)' },
  { value: 'binary_output_status', label: 'Binary Output Status (g10/g11)' },
  { value: 'counter',              label: 'Counter (g20/g22)' },
  { value: 'frozen_counter',       label: 'Frozen Counter (g21/g23)' },
  { value: 'analog',               label: 'Analog Input (g30/g32)' },
  { value: 'analog_output_status', label: 'Analog Output Status (g40/g42)' },
  { value: 'octet_string',         label: 'Octet String (g110/g111)' },
]

const modbusFunctions: { value: ModbusFunction; label: string }[] = [
  { value: 'coil',             label: 'Coil (FC01)' },
  { value: 'discrete_input',   label: 'Discrete Input (FC02)' },
  { value: 'holding_register', label: 'Holding Register (FC03)' },
  { value: 'input_register',   label: 'Input Register (FC04)' },
]
const modbusDataTypes: ModbusDataType[] = ['int16', 'uint16', 'int32', 'uint32', 'float32', 'float64']
const byteOrders = ['ABCD', 'CDAB', 'BADC', 'DCBA']

const isRegister = computed(
  () => form.value.function === 'holding_register' || form.value.function === 'input_register',
)

function isModbus(m: SignalMapping) { return m.protocol === 'modbus' }
function srcId(m: SignalMapping) { return m.sourceId || m.outstationId || '—' }

const shortPoint = (pt: string) => ({
  binary: 'g1', double_bit_binary: 'g3', binary_output_status: 'g10',
  counter: 'g20', frozen_counter: 'g21', analog: 'g30',
  analog_output_status: 'g40', octet_string: 'g110',
}[pt] ?? pt)

function pointLabel(m: SignalMapping): string {
  if (isModbus(m)) {
    const fn = (m.function ?? '').replace(/_/g, ' ')
    const dt = m.dataType ? ` · ${m.dataType}` : ''
    return `${fn} @${m.address ?? 0}${dt}`
  }
  return `${shortPoint(m.pointType)} #${m.index}${m.eventClass ? ' · C' + m.eventClass : ''}`
}

const filtered = computed(() => {
  const f = filter.value.toLowerCase()
  if (!f) return store.mappings
  return store.mappings.filter((m) =>
    `${m.metricName} ${m.deviceId ?? ''} ${srcId(m)} ${pointLabel(m)}`.toLowerCase().includes(f),
  )
})

const emptyForm = (): SignalMapping => ({
  id: '', metricName: '', deviceId: '',
  protocol: 'dnp3', sourceId: '', outstationId: '',
  pointType: 'analog', index: 0, eventClass: 1,
  function: 'holding_register', address: 0, quantity: 0, dataType: 'float32', byteOrder: 'ABCD',
  scale: 1, offset: 0, engineeringUnit: '',
  deadband: 0, publishOnPoll: false,
  enabled: true,
})
const form = ref<SignalMapping>(emptyForm())

function openAdd() { editId.value = ''; form.value = emptyForm(); error.value = ''; modal.value = true }
function openEdit(m: SignalMapping) {
  editId.value = m.id
  form.value = { ...emptyForm(), ...m, protocol: m.protocol ?? 'dnp3', sourceId: m.sourceId || m.outstationId || '' }
  error.value = ''
  modal.value = true
}

async function saveMapping() {
  error.value = ''
  // Keep legacy outstationId aligned with sourceId for DNP3 mappings.
  const payload: SignalMapping = { ...form.value }
  if (payload.protocol !== 'modbus') payload.outstationId = payload.sourceId ?? ''
  try {
    if (editId.value) await api.updateMapping(editId.value, payload)
    else await api.addMapping(payload)
    await store.loadMappings()
    modal.value = false
  } catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
}

async function del(id: string) {
  if (!confirm('Delete this signal mapping?')) return
  await api.deleteMapping(id)
  await store.loadMappings()
}

async function importFile(ev: Event) {
  const file = (ev.target as HTMLInputElement).files?.[0]
  if (!file) return
  try {
    const mappings: SignalMapping[] = JSON.parse(await file.text())
    await api.importMappings(mappings)
    await store.loadMappings()
  } catch (e: unknown) {
    alert('Import failed: ' + (e instanceof Error ? e.message : String(e)))
  }
}
</script>

<style scoped>
.seg { display: inline-flex; border: 1.5px solid var(--border); border-radius: var(--radius); overflow: hidden; background: var(--card); }
.seg-btn { font-family: var(--font-sans); font-weight: 700; font-size: 11px; padding: 7px 16px; color: var(--muted-foreground); background: transparent; border: none; cursor: pointer; transition: background 0.12s, color 0.12s; }
.seg-btn + .seg-btn { border-left: 1.5px solid var(--border); }
.seg-btn--active { background: color-mix(in srgb, var(--epm-bosque) 12%, transparent); color: var(--epm-bosque); }
.proto-tag { font-family: var(--font-mono); font-size: 9.5px; font-weight: 600; letter-spacing: 0.06em; padding: 2px 6px; border-radius: 3px; border: 1px solid; }
.proto-tag--dnp { color: var(--epm-bosque); border-color: color-mix(in srgb, var(--epm-bosque) 40%, transparent); background: color-mix(in srgb, var(--epm-bosque) 8%, transparent); }
.proto-tag--mb  { color: var(--tk-amber-bright); border-color: color-mix(in srgb, var(--tk-amber-base) 45%, transparent); background: color-mix(in srgb, var(--tk-amber-base) 10%, transparent); }
</style>
