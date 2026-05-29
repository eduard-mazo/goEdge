<template>
  <div class="space-y-5">

    <div class="flex items-center gap-4">
      <p class="font-sans text-sm text-text-secondary flex-1">
        Modbus TCP target devices. Each device needs a unique slug ID.
      </p>
      <button class="btn-primary text-xs py-1.5" @click="openAdd">+ Add Device</button>
    </div>

    <!-- Empty state -->
    <div v-if="!store.devices.length"
         class="forge-panel text-center py-12">
      <p class="font-mono text-xs text-text-dim">No Modbus devices configured.</p>
      <button class="btn-ghost mt-4 text-xs" @click="openAdd">Add first device</button>
    </div>

    <!-- Devices table -->
    <div v-else class="forge-panel overflow-x-auto">
      <table class="min-w-[700px]">
        <thead>
          <tr>
            <th>ID</th><th>Label</th><th>Host : Port</th>
            <th>Timeout</th><th>Retries</th><th>Status</th><th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="dev in store.devices" :key="dev.id">
            <td class="font-mono text-xs text-text-secondary">{{ dev.id }}</td>
            <td class="font-sans font-semibold text-sm">{{ dev.label }}</td>
            <td class="font-mono text-xs">{{ dev.host }}:{{ dev.port ?? 502 }}</td>
            <td class="font-mono text-xs text-text-secondary">{{ dev.timeoutMs ?? 3000 }}&nbsp;ms</td>
            <td class="font-mono text-xs text-text-secondary">{{ dev.retries ?? 2 }}</td>
            <td>
              <span :class="['signal-badge', dev.enabled ? 'signal-badge--on' : 'signal-badge--off']">
                {{ dev.enabled ? 'enabled' : 'disabled' }}
              </span>
            </td>
            <td>
              <div class="flex gap-2 justify-end">
                <button class="btn-ghost text-xs py-0.5 px-3" @click="openEdit(dev)">Edit</button>
                <button
                  class="btn-ghost text-xs py-0.5 px-3 !border-red-base/40 !text-red-base hover:!border-red-bright hover:!text-red-bright"
                  @click="del(dev.id)"
                >Del</button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="modal" class="modal-backdrop" @click.self="modal = false">
        <div class="forge-panel w-full max-w-md">
          <div class="forge-header">{{ editId ? 'Edit Device' : 'Add Modbus Device' }}</div>
          <form @submit.prevent="saveDevice" class="p-5 space-y-3">
            <Field label="Device ID (unique slug)">
              <input v-model="form.id" class="forge-input" :disabled="!!editId" required placeholder="plc-01" />
            </Field>
            <Field label="Label">
              <input v-model="form.label" class="forge-input" placeholder="Main PLC" />
            </Field>
            <div class="grid grid-cols-3 gap-3">
              <Field label="Host" class="col-span-2">
                <input v-model="form.host" class="forge-input" placeholder="192.168.1.100" required />
              </Field>
              <Field label="Port">
                <input v-model.number="form.port" class="forge-input" type="number" placeholder="502" />
              </Field>
            </div>
            <div class="grid grid-cols-3 gap-3">
              <Field label="Timeout ms">
                <input v-model.number="form.timeoutMs" class="forge-input" type="number" placeholder="3000" />
              </Field>
              <Field label="Retries">
                <input v-model.number="form.retries" class="forge-input" type="number" placeholder="2" />
              </Field>
              <Field label="Retry delay ms">
                <input v-model.number="form.retryDelayMs" class="forge-input" type="number" placeholder="500" />
              </Field>
            </div>
            <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary">
              <input type="checkbox" v-model="form.enabled" /> Enabled
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
  id: '', label: '', host: '', port: 502,
  timeoutMs: 3000, retries: 2, retryDelayMs: 500, enabled: true,
})
const form = ref<ModbusDevice>(empty())

function openAdd() { editId.value = ''; form.value = empty(); error.value = ''; modal.value = true }
function openEdit(d: ModbusDevice) { editId.value = d.id; form.value = { ...d }; error.value = ''; modal.value = true }

async function saveDevice() {
  error.value = ''
  try {
    if (editId.value) await api.updateDevice(editId.value, form.value)
    else await api.addDevice(form.value)
    await store.loadDevices()
    modal.value = false
  } catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
}

async function del(id: string) {
  if (!confirm(`Delete device "${id}" and all its mappings?`)) return
  await api.deleteDevice(id)
  await store.loadDevices()
  await store.loadMappings()
}
</script>
