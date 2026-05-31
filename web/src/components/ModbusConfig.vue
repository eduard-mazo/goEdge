<template>
  <div class="space-y-5">

    <div class="flex items-center gap-4">
      <p class="font-sans text-sm text-text-secondary flex-1">
        Modbus/TCP devices. Each one is polled on its own cadence; mapped registers are decoded
        per data-type and byte-order and published to Sparkplug.
      </p>
      <button class="btn-primary text-xs py-1.5" @click="openAdd">+ Add Device</button>
    </div>

    <!-- Empty state -->
    <div v-if="!store.modbusDevices.length" class="forge-panel text-center py-12">
      <p class="font-mono text-xs text-text-dim">No Modbus devices configured.</p>
      <button class="btn-ghost mt-4 text-xs" @click="openAdd">Add first device</button>
    </div>

    <!-- Devices table -->
    <div v-else class="forge-panel overflow-x-auto">
      <table class="min-w-[760px]">
        <thead>
          <tr>
            <th>ID</th><th>Label</th><th>Host : Port</th>
            <th>Unit</th><th>Scan</th><th>Timeout</th><th>Status</th><th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="d in store.modbusDevices" :key="d.id">
            <td class="font-mono text-xs text-text-secondary">{{ d.id }}</td>
            <td class="font-sans font-semibold text-sm">{{ d.label }}</td>
            <td class="font-mono text-xs">{{ d.host }}:{{ d.port ?? 502 }}</td>
            <td class="font-mono text-xs text-text-secondary">{{ d.unitId }}</td>
            <td class="font-mono text-[11px] text-text-secondary">{{ fmtMs(d.scanRateMs) }}</td>
            <td class="font-mono text-[11px] text-text-secondary">{{ fmtMs(d.timeoutMs) }}</td>
            <td>
              <span :class="['signal-badge', d.enabled ? 'signal-badge--on' : 'signal-badge--off']">
                {{ d.enabled ? 'enabled' : 'disabled' }}
              </span>
            </td>
            <td>
              <div class="flex gap-2 justify-end">
                <button class="btn-ghost text-xs py-0.5 px-3" @click="openEdit(d)">Edit</button>
                <button
                  class="btn-ghost text-xs py-0.5 px-3 !border-red-base/40 !text-red-base hover:!border-red-bright hover:!text-red-bright"
                  @click="del(d.id)"
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
        <div class="forge-panel w-full max-w-xl mx-auto">
          <div class="forge-header">{{ editId ? 'Edit Modbus Device' : 'Add Modbus Device' }}</div>
          <form @submit.prevent="saveDevice" class="p-5 space-y-4">

            <div class="grid grid-cols-2 gap-4">
              <Field label="Device ID (unique slug)">
                <input v-model="form.id" class="forge-input" :disabled="!!editId" required placeholder="plc-01" />
              </Field>
              <Field label="Label">
                <input v-model="form.label" class="forge-input" placeholder="Main PLC" />
              </Field>
            </div>

            <div class="grid grid-cols-3 gap-4">
              <Field label="Host" class="col-span-2">
                <input v-model="form.host" class="forge-input" placeholder="192.168.1.50" required />
              </Field>
              <Field label="Port" hint="Modbus/TCP default 502">
                <input v-model.number="form.port" class="forge-input" type="number" placeholder="502" />
              </Field>
            </div>

            <Field label="Unit ID" hint="Modbus slave/unit address (typical: 1)">
              <input v-model.number="form.unitId" class="forge-input" type="number" min="0" max="255" />
            </Field>

            <fieldset class="space-y-3">
              <legend class="font-mono text-[10px] uppercase tracking-[0.18em] text-text-dim">Polling</legend>
              <div class="grid grid-cols-2 gap-4">
                <Field label="Scan rate (ms)" hint="default 1000">
                  <input v-model.number="form.scanRateMs" class="forge-input" type="number" placeholder="1000" />
                </Field>
                <Field label="Timeout (ms)" hint="default 3000">
                  <input v-model.number="form.timeoutMs" class="forge-input" type="number" placeholder="3000" />
                </Field>
                <Field label="Retries" hint="default 2">
                  <input v-model.number="form.retries" class="forge-input" type="number" min="0" placeholder="2" />
                </Field>
                <Field label="Retry delay (ms)" hint="default 500">
                  <input v-model.number="form.retryDelayMs" class="forge-input" type="number" placeholder="500" />
                </Field>
              </div>
            </fieldset>

            <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary pt-2 border-t border-border">
              <input type="checkbox" v-model="form.enabled" /> Device enabled
            </label>

            <p v-if="error" class="font-mono text-xs text-red-bright">⚠ {{ error }}</p>
            <div class="flex justify-end gap-3 pt-2 border-t border-border">
              <button type="button" class="btn-ghost text-xs" @click="modal = false">Cancel</button>
              <button type="submit" class="btn-primary text-xs">Save Device</button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useGatewayStore } from '@/stores/gateway'
import { api, type ModbusDevice } from '@/api/client'
import Field from './Field.vue'

const store  = useGatewayStore()
const modal  = ref(false)
const editId = ref('')
const error  = ref('')

const empty = (): ModbusDevice => ({
  id: '', label: '',
  host: '', port: 502, unitId: 1,
  scanRateMs: 1000, timeoutMs: 3000, retries: 2, retryDelayMs: 500,
  enabled: true,
})
const form = ref<ModbusDevice>(empty())

function openAdd() { editId.value = ''; form.value = empty(); error.value = ''; modal.value = true }
function openEdit(d: ModbusDevice) {
  editId.value = d.id
  form.value = { ...empty(), ...d }
  error.value = ''
  modal.value = true
}

async function saveDevice() {
  error.value = ''
  try {
    if (editId.value) await api.updateModbusDevice(editId.value, form.value)
    else await api.addModbusDevice(form.value)
    await store.loadModbusDevices()
    modal.value = false
  } catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
}

async function del(id: string) {
  if (!confirm(`Delete device "${id}" and all its mappings?`)) return
  await api.deleteModbusDevice(id)
  await store.loadModbusDevices()
  await store.loadMappings()
}

function fmtMs(ms?: number) {
  if (!ms) return '—'
  if (ms >= 1000) return (ms / 1000).toFixed(ms % 1000 ? 1 : 0) + 's'
  return ms + 'ms'
}
</script>
