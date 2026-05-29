<template>
  <div class="space-y-4">

    <!-- Toolbar -->
    <div class="flex items-center gap-3 flex-wrap">
      <p class="font-sans text-sm text-text-secondary flex-1">
        DNP3 point → Sparkplug B metric mappings
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
           placeholder="Filter by metric name, outstation ID…" />

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
            <th>Sparkplug Dev</th>
            <th>Outstation</th>
            <th>Point Type</th>
            <th>Index</th>
            <th>Class</th>
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
            <td class="font-mono text-[11px] text-text-dim">{{ m.deviceId || '(node)' }}</td>
            <td class="font-mono text-[11px]">{{ m.outstationId }}</td>
            <td>
              <span class="font-mono text-[10px] px-1.5 py-0.5 border border-border rounded-sm text-text-secondary">
                {{ shortPoint(m.pointType) }}
              </span>
            </td>
            <td class="font-mono text-[11px] text-citrico">{{ m.index }}</td>
            <td class="font-mono text-[11px] text-text-secondary">
              {{ m.eventClass === 0 ? 'static' : 'C' + m.eventClass }}
            </td>
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

            <div class="grid grid-cols-2 gap-4">
              <Field label="Metric Name" hint="Unique Sparkplug B metric identifier">
                <input v-model="form.metricName" class="forge-input" placeholder="feeder/voltage_kv" required />
              </Field>
              <Field label="Sparkplug Device ID" hint="Leave blank to publish as node metric">
                <input v-model="form.deviceId" class="forge-input" placeholder="(node level)" />
              </Field>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <Field label="Outstation">
                <select v-model="form.outstationId" class="forge-input" required>
                  <option value="" disabled>Select outstation…</option>
                  <option v-for="o in store.outstations" :key="o.id" :value="o.id">
                    {{ o.label || o.id }} ({{ o.host }}:{{ o.port }})
                  </option>
                </select>
              </Field>
              <Field label="Point Type" hint="Maps to DNP3 group/variation">
                <select v-model="form.pointType" class="forge-input" required>
                  <option v-for="pt in pointTypes" :key="pt.value" :value="pt.value">
                    {{ pt.label }}
                  </option>
                </select>
              </Field>
            </div>

            <div class="grid grid-cols-3 gap-4">
              <Field label="Index" hint="Point index within its type (0-based)">
                <input v-model.number="form.index" class="forge-input" type="number" min="0" max="65535" required />
              </Field>
              <Field label="Event Class" hint="Informational; outstation assigns at config time">
                <select v-model.number="form.eventClass" class="forge-input">
                  <option :value="0">Static only</option>
                  <option :value="1">Class 1</option>
                  <option :value="2">Class 2</option>
                  <option :value="3">Class 3</option>
                </select>
              </Field>
              <Field label="Deadband" hint="Min change (engineering units) to publish">
                <input v-model.number="form.deadband" class="forge-input" type="number" step="any" placeholder="0" />
              </Field>
            </div>

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

            <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary">
              <input type="checkbox" v-model="form.publishOnPoll" />
              Publish on static reads too (not only on event)
            </label>
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
import { api, type SignalMapping, type PointType } from '@/api/client'
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

const shortPoint = (pt: string) => ({
  binary: 'g1',
  double_bit_binary: 'g3',
  binary_output_status: 'g10',
  counter: 'g20',
  frozen_counter: 'g21',
  analog: 'g30',
  analog_output_status: 'g40',
  octet_string: 'g110',
}[pt] ?? pt)

const filtered = computed(() =>
  filter.value
    ? store.mappings.filter(m =>
        m.metricName.includes(filter.value) ||
        (m.deviceId ?? '').includes(filter.value) ||
        m.outstationId.includes(filter.value)
      )
    : store.mappings
)

const emptyForm = (): SignalMapping => ({
  id: '', metricName: '', deviceId: '', outstationId: '',
  pointType: 'analog', index: 0, eventClass: 1,
  scale: 1, offset: 0, engineeringUnit: '',
  deadband: 0, publishOnPoll: false,
  enabled: true,
})
const form = ref<SignalMapping>(emptyForm())

function openAdd() { editId.value = ''; form.value = emptyForm(); error.value = ''; modal.value = true }
function openEdit(m: SignalMapping) { editId.value = m.id; form.value = { ...emptyForm(), ...m }; error.value = ''; modal.value = true }

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
