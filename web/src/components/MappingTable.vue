<template>
  <div class="space-y-4">

    <!-- Toolbar -->
    <div class="flex items-center gap-3 flex-wrap">
      <p class="font-sans text-sm text-text-secondary flex-1">
        Modbus register → Sparkplug B metric mappings
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
           placeholder="Filter by metric name, device ID…" />

    <!-- Empty -->
    <div v-if="!filtered.length"
         class="forge-panel text-center py-12">
      <p class="font-mono text-xs text-text-dim">
        {{ store.mappings.length ? 'No results for filter.' : 'No signal mappings yet.' }}
      </p>
    </div>

    <!-- Table -->
    <div v-else class="forge-panel overflow-x-auto">
      <table class="min-w-[1000px]">
        <thead>
          <tr>
            <th>Metric Name</th>
            <th>Device ID</th>
            <th>Modbus Dev</th>
            <th>Unit</th>
            <th>Fn</th>
            <th>Addr</th>
            <th>Qty</th>
            <th>Type</th>
            <th>Scale</th>
            <th>EU</th>
            <th>Rate</th>
            <th>Status</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="m in filtered" :key="m.id">
            <td class="font-mono text-xs font-medium text-foreground">{{ m.metricName }}</td>
            <td class="font-mono text-[11px] text-text-dim">{{ m.deviceId || '(node)' }}</td>
            <td class="font-mono text-[11px]">{{ m.modbusDeviceId }}</td>
            <td class="font-mono text-[11px] text-text-secondary">{{ m.unitId }}</td>
            <td>
              <span class="font-mono text-[10px] px-1.5 py-0.5 border border-border rounded-sm text-text-secondary">
                {{ fnCode(m.function) }}
              </span>
            </td>
            <td class="font-mono text-[11px] text-citrico">{{ m.address }}</td>
            <td class="font-mono text-[11px] text-text-secondary">{{ m.quantity }}</td>
            <td class="font-mono text-[11px]">{{ m.dataType }}</td>
            <td class="font-mono text-[11px] text-text-secondary">
              {{ m.scale ?? 1 }}{{ m.offset ? ' +' + m.offset : '' }}
            </td>
            <td class="font-mono text-[11px] text-text-dim">{{ m.engineeringUnit || '—' }}</td>
            <td class="font-mono text-[11px] text-text-secondary">{{ m.scanRateMs }}ms</td>
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

            <div class="grid grid-cols-2 gap-4">
              <Field label="Metric Name" hint="Unique Sparkplug B metric identifier">
                <input v-model="form.metricName" class="forge-input" placeholder="tank/level_m" required />
              </Field>
              <Field label="Sparkplug Device ID" hint="Leave blank to publish as node metric">
                <input v-model="form.deviceId" class="forge-input" placeholder="(node level)" />
              </Field>
            </div>

            <div class="grid grid-cols-3 gap-4">
              <Field label="Modbus Device">
                <select v-model="form.modbusDeviceId" class="forge-input" required>
                  <option value="" disabled>Select device…</option>
                  <option v-for="d in store.devices" :key="d.id" :value="d.id">
                    {{ d.label || d.id }}
                  </option>
                </select>
              </Field>
              <Field label="Unit ID (0–255)">
                <input v-model.number="form.unitId" class="forge-input" type="number" min="0" max="255" required />
              </Field>
              <Field label="Function Code">
                <select v-model="form.function" class="forge-input" required>
                  <option value="coil">FC01 – Coil</option>
                  <option value="discrete_input">FC02 – Discrete Input</option>
                  <option value="holding_register">FC03 – Holding Register</option>
                  <option value="input_register">FC04 – Input Register</option>
                </select>
              </Field>
            </div>

            <div class="grid grid-cols-4 gap-4">
              <Field label="Address">
                <input v-model.number="form.address" class="forge-input" type="number" min="0" max="65535" required />
              </Field>
              <Field label="Quantity">
                <input v-model.number="form.quantity" class="forge-input" type="number" min="1" max="125" required />
              </Field>
              <Field label="Data Type">
                <select v-model="form.dataType" class="forge-input" required>
                  <option v-for="t in dataTypes" :key="t" :value="t">{{ t }}</option>
                </select>
              </Field>
              <Field label="Byte Order">
                <select v-model="form.byteOrder" class="forge-input">
                  <option value="ABCD">ABCD – Big-Endian</option>
                  <option value="DCBA">DCBA – Little-Endian</option>
                  <option value="BADC">BADC – Byte-Swap</option>
                  <option value="CDAB">CDAB – Word-Swap</option>
                </select>
              </Field>
            </div>

            <div class="grid grid-cols-4 gap-4">
              <Field label="Scale" hint="value × scale + offset">
                <input v-model.number="form.scale" class="forge-input" type="number" step="any" placeholder="1" />
              </Field>
              <Field label="Offset">
                <input v-model.number="form.offset" class="forge-input" type="number" step="any" placeholder="0" />
              </Field>
              <Field label="Eng. Unit">
                <input v-model="form.engineeringUnit" class="forge-input" placeholder="°C" />
              </Field>
              <Field label="Deadband">
                <input v-model.number="form.deadband" class="forge-input" type="number" step="any" placeholder="0" />
              </Field>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <Field label="Scan Rate (ms)" hint="Minimum 100ms">
                <input v-model.number="form.scanRateMs" class="forge-input" type="number" min="100" step="100" required />
              </Field>
              <Field label="Quality Policy">
                <select v-model="form.qualityPolicy" class="forge-input">
                  <option value="good">Good — skip on error</option>
                  <option value="bad_on_error">Bad — publish null on error</option>
                  <option value="last_known">Last known — keep previous value</option>
                </select>
              </Field>
            </div>

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
import { api, type SignalMapping } from '@/api/client'
import Field from './Field.vue'

const store  = useGatewayStore()
const modal  = ref(false)
const editId = ref('')
const error  = ref('')
const filter = ref('')

const dataTypes = ['bool','int16','uint16','int32','uint32','float32','int64','uint64','float64']

const fnCode = (fn: string) => ({ coil:'FC01', discrete_input:'FC02', holding_register:'FC03', input_register:'FC04' }[fn] ?? fn)

const filtered = computed(() =>
  filter.value
    ? store.mappings.filter(m =>
        m.metricName.includes(filter.value) ||
        (m.deviceId ?? '').includes(filter.value) ||
        m.modbusDeviceId.includes(filter.value)
      )
    : store.mappings
)

const emptyForm = (): SignalMapping => ({
  id: '', metricName: '', deviceId: '', modbusDeviceId: '',
  unitId: 1, function: 'holding_register', address: 0, quantity: 1,
  dataType: 'uint16', byteOrder: 'ABCD', scale: 1, offset: 0,
  engineeringUnit: '', scanRateMs: 1000, deadband: 0,
  qualityPolicy: 'last_known', enabled: true,
})
const form = ref<SignalMapping>(emptyForm())

function openAdd() { editId.value = ''; form.value = emptyForm(); error.value = ''; modal.value = true }
function openEdit(m: SignalMapping) { editId.value = m.id; form.value = { ...m }; error.value = ''; modal.value = true }

async function saveMapping() {
  error.value = ''
  try {
    if (editId.value) await api.updateMapping(editId.value, form.value)
    else await api.addMapping(form.value)
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
