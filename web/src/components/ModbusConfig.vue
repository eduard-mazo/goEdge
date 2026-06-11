<template>
  <div class="space-y-5">

    <div class="flex items-center gap-4">
      <p class="font-sans text-sm text-text-secondary flex-1">
        Dispositivos Modbus/TCP. Cada uno se sondea con su propia cadencia.
      </p>
      <button class="btn-primary text-xs py-1.5" @click="openAdd">+ Agregar</button>
    </div>

    <!-- Empty state -->
    <div v-if="!store.modbusDevices.length" class="forge-panel text-center py-12">
      <p class="font-mono text-xs text-text-dim">Sin dispositivos Modbus.</p>
      <button class="btn-ghost mt-4 text-xs" @click="openAdd">Agregar el primero</button>
    </div>

    <!-- Devices table -->
    <div v-else class="forge-panel overflow-x-auto">
      <table class="min-w-[760px]">
        <thead>
          <tr>
            <th>ID</th><th>Nombre</th><th>Host : Puerto</th>
            <th>Unidad</th><th>Sondeo</th><th>Timeout</th><th>Estado</th><th></th>
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

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="modal" class="modal-backdrop overflow-y-auto py-6">
        <div class="forge-panel w-full max-w-2xl mx-auto">
          <div class="forge-header flex items-center justify-between gap-3">
            <span>{{ editId ? 'Editar dispositivo' : 'Nuevo dispositivo Modbus' }}</span>
            <button type="button" class="modal-x" @click="modal = false" aria-label="Cerrar">✕</button>
          </div>
          <form @submit.prevent="saveDevice" class="p-5 space-y-4">

            <div class="grid grid-cols-2 gap-4">
              <Field label="ID (identificador único)">
                <input v-model="form.id" class="forge-input" :disabled="!!editId" required placeholder="plc-01" />
              </Field>
              <Field label="Nombre">
                <input v-model="form.label" class="forge-input" placeholder="PLC principal" />
              </Field>
            </div>

            <div class="grid grid-cols-3 gap-4">
              <Field label="Host" class="col-span-2">
                <input v-model="form.host" class="forge-input" placeholder="192.168.1.50" required />
              </Field>
              <Field label="Puerto" hint="Modbus/TCP por defecto 502">
                <input v-model.number="form.port" class="forge-input" type="number" placeholder="502" />
              </Field>
            </div>

            <Field label="Unit ID" hint="Dirección del esclavo (típico: 1)">
              <input v-model.number="form.unitId" class="forge-input" type="number" min="0" max="255" />
            </Field>

            <fieldset class="space-y-3">
              <legend class="font-mono text-[10px] uppercase tracking-[0.18em] text-text-dim">Sondeo</legend>
              <div class="grid grid-cols-2 gap-4">
                <Field label="Cadencia (ms)" hint="por defecto 1000">
                  <input v-model.number="form.scanRateMs" class="forge-input" type="number" placeholder="1000" />
                </Field>
                <Field label="Timeout (ms)" hint="por defecto 3000">
                  <input v-model.number="form.timeoutMs" class="forge-input" type="number" placeholder="3000" />
                </Field>
                <Field label="Reintentos" hint="por defecto 2">
                  <input v-model.number="form.retries" class="forge-input" type="number" min="0" placeholder="2" />
                </Field>
                <Field label="Espera reintento (ms)" hint="por defecto 500">
                  <input v-model.number="form.retryDelayMs" class="forge-input" type="number" placeholder="500" />
                </Field>
              </div>
            </fieldset>

            <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary pt-2 border-t border-border">
              <input type="checkbox" v-model="form.enabled" /> Dispositivo habilitado
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
  if (!confirm(`¿Eliminar el dispositivo "${id}" y todas sus señales?`)) return
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
