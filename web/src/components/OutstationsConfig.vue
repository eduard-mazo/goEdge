<template>
  <div class="space-y-5">

    <div class="flex items-center gap-4">
      <p class="font-sans text-sm text-text-secondary flex-1">
        DNP3 outstations. Each one is a TCP target with its own link-layer addresses and poll cadence.
      </p>
      <button class="btn-primary text-xs py-1.5" @click="openAdd">+ Add Outstation</button>
    </div>

    <!-- Empty state -->
    <div v-if="!store.outstations.length"
         class="forge-panel text-center py-12">
      <p class="font-mono text-xs text-text-dim">No DNP3 outstations configured.</p>
      <button class="btn-ghost mt-4 text-xs" @click="openAdd">Add first outstation</button>
    </div>

    <!-- Outstations table -->
    <div v-else class="forge-panel overflow-x-auto">
      <table class="min-w-[900px]">
        <thead>
          <tr>
            <th>ID</th><th>Label</th><th>Host : Port</th>
            <th>Master ↔ Outstation</th>
            <th>Polls (ms)</th>
            <th>Unsol</th>
            <th>Status</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="o in store.outstations" :key="o.id">
            <td class="font-mono text-xs text-text-secondary">{{ o.id }}</td>
            <td class="font-sans font-semibold text-sm">{{ o.label }}</td>
            <td class="font-mono text-xs">{{ o.host }}:{{ o.port ?? 20000 }}</td>
            <td class="font-mono text-xs text-text-secondary">{{ o.masterAddress }} → {{ o.outstationAddress }}</td>
            <td class="font-mono text-[11px] text-text-secondary">
              <span :title="`integrity ${o.integrityScanMs ?? 0}, c1 ${o.class1ScanMs ?? 0}, c2 ${o.class2ScanMs ?? 0}, c3 ${o.class3ScanMs ?? 0}`">
                int {{ fmtMs(o.integrityScanMs) }} / c1 {{ fmtMs(o.class1ScanMs) }} / c2 {{ fmtMs(o.class2ScanMs) }} / c3 {{ fmtMs(o.class3ScanMs) }}
              </span>
            </td>
            <td class="font-mono text-[11px]">
              <span v-if="o.unsolicitedEnabled" class="text-citrico">
                {{ [o.unsolicitedClass1 && '1', o.unsolicitedClass2 && '2', o.unsolicitedClass3 && '3'].filter(Boolean).join(',') || '–' }}
              </span>
              <span v-else class="text-text-dim">off</span>
            </td>
            <td>
              <span :class="['signal-badge', o.enabled ? 'signal-badge--on' : 'signal-badge--off']">
                {{ o.enabled ? 'enabled' : 'disabled' }}
              </span>
            </td>
            <td>
              <div class="flex gap-2 justify-end">
                <button class="btn-ghost text-xs py-0.5 px-3" @click="openEdit(o)">Edit</button>
                <button
                  class="btn-ghost text-xs py-0.5 px-3 !border-red-base/40 !text-red-base hover:!border-red-bright hover:!text-red-bright"
                  @click="del(o.id)"
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
          <div class="forge-header">{{ editId ? 'Edit Outstation' : 'Add DNP3 Outstation' }}</div>
          <form @submit.prevent="saveOutstation" class="p-5 space-y-4">

            <div class="grid grid-cols-2 gap-4">
              <Field label="Outstation ID (unique slug)">
                <input v-model="form.id" class="forge-input" :disabled="!!editId" required placeholder="rtu-01" />
              </Field>
              <Field label="Label">
                <input v-model="form.label" class="forge-input" placeholder="Substation RTU 01" />
              </Field>
            </div>

            <div class="grid grid-cols-3 gap-4">
              <Field label="Host" class="col-span-2">
                <input v-model="form.host" class="forge-input" placeholder="192.168.1.100" required />
              </Field>
              <Field label="Port" hint="DNP3/IP default 20000">
                <input v-model.number="form.port" class="forge-input" type="number" placeholder="20000" />
              </Field>
            </div>

            <fieldset class="space-y-3">
              <legend class="font-mono text-[10px] uppercase tracking-[0.18em] text-text-dim">Link layer</legend>
              <div class="grid grid-cols-2 gap-4">
                <Field label="Master Address" hint="Local link-layer address (typical: 1)">
                  <input v-model.number="form.masterAddress" class="forge-input" type="number" min="1" max="65519" required />
                </Field>
                <Field label="Outstation Address" hint="Remote link-layer address (typical: 1024+)">
                  <input v-model.number="form.outstationAddress" class="forge-input" type="number" min="1" max="65519" required />
                </Field>
              </div>
            </fieldset>

            <fieldset class="space-y-3">
              <legend class="font-mono text-[10px] uppercase tracking-[0.18em] text-text-dim">Application layer</legend>
              <div class="grid grid-cols-2 gap-4">
                <Field label="Response Timeout (ms)" hint="default 5000">
                  <input v-model.number="form.responseTimeoutMs" class="forge-input" type="number" placeholder="5000" />
                </Field>
                <Field label="Keep-Alive (ms)" hint="default 60000; 0 = disabled">
                  <input v-model.number="form.keepAliveMs" class="forge-input" type="number" placeholder="60000" />
                </Field>
              </div>
            </fieldset>

            <fieldset class="space-y-3">
              <legend class="font-mono text-[10px] uppercase tracking-[0.18em] text-text-dim">
                Periodic polls (0 = disabled)
              </legend>
              <div class="grid grid-cols-4 gap-3">
                <Field label="Integrity (ms)" hint="all classes">
                  <input v-model.number="form.integrityScanMs" class="forge-input" type="number" placeholder="3600000" />
                </Field>
                <Field label="Class 1 (ms)">
                  <input v-model.number="form.class1ScanMs" class="forge-input" type="number" placeholder="1000" />
                </Field>
                <Field label="Class 2 (ms)">
                  <input v-model.number="form.class2ScanMs" class="forge-input" type="number" placeholder="5000" />
                </Field>
                <Field label="Class 3 (ms)">
                  <input v-model.number="form.class3ScanMs" class="forge-input" type="number" placeholder="30000" />
                </Field>
              </div>
            </fieldset>

            <fieldset class="space-y-3">
              <legend class="font-mono text-[10px] uppercase tracking-[0.18em] text-text-dim">Unsolicited responses</legend>
              <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary">
                <input type="checkbox" v-model="form.unsolicitedEnabled" />
                Enable unsolicited responses at startup
              </label>
              <div class="grid grid-cols-3 gap-3 pl-6" v-if="form.unsolicitedEnabled">
                <label class="flex items-center gap-2 cursor-pointer font-sans text-sm">
                  <input type="checkbox" v-model="form.unsolicitedClass1" /> Class 1
                </label>
                <label class="flex items-center gap-2 cursor-pointer font-sans text-sm">
                  <input type="checkbox" v-model="form.unsolicitedClass2" /> Class 2
                </label>
                <label class="flex items-center gap-2 cursor-pointer font-sans text-sm">
                  <input type="checkbox" v-model="form.unsolicitedClass3" /> Class 3
                </label>
              </div>
              <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary">
                <input type="checkbox" v-model="form.disableUnsolOnStartup" />
                Send DISABLE_UNSOLICITED before initial integrity poll
              </label>
              <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary">
                <input type="checkbox" v-model="form.startupIntegrity" />
                Perform integrity poll on connect (recommended)
              </label>
            </fieldset>

            <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary pt-2 border-t border-border">
              <input type="checkbox" v-model="form.enabled" /> Outstation enabled
            </label>

            <p v-if="error" class="font-mono text-xs text-red-bright">⚠ {{ error }}</p>
            <div class="flex justify-end gap-3 pt-2 border-t border-border">
              <button type="button" class="btn-ghost text-xs" @click="modal = false">Cancel</button>
              <button type="submit" class="btn-primary text-xs">Save Outstation</button>
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
import { api, type DNP3Outstation } from '@/api/client'
import Field from './Field.vue'

const store  = useGatewayStore()
const modal  = ref(false)
const editId = ref('')
const error  = ref('')

const empty = (): DNP3Outstation => ({
  id: '', label: '',
  host: '', port: 20000,
  masterAddress: 1, outstationAddress: 1024,
  responseTimeoutMs: 5000, keepAliveMs: 60000,
  integrityScanMs: 3600000, class1ScanMs: 1000, class2ScanMs: 5000, class3ScanMs: 30000,
  unsolicitedEnabled: false,
  unsolicitedClass1: true, unsolicitedClass2: false, unsolicitedClass3: false,
  disableUnsolOnStartup: true, startupIntegrity: true,
  enabled: true,
})
const form = ref<DNP3Outstation>(empty())

function openAdd() { editId.value = ''; form.value = empty(); error.value = ''; modal.value = true }
function openEdit(o: DNP3Outstation) {
  editId.value = o.id
  form.value = { ...empty(), ...o }
  error.value = ''
  modal.value = true
}

async function saveOutstation() {
  error.value = ''
  try {
    if (editId.value) await api.updateOutstation(editId.value, form.value)
    else await api.addOutstation(form.value)
    await store.loadOutstations()
    modal.value = false
  } catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
}

async function del(id: string) {
  if (!confirm(`Delete outstation "${id}" and all its mappings?`)) return
  await api.deleteOutstation(id)
  await store.loadOutstations()
  await store.loadMappings()
}

function fmtMs(ms?: number) {
  if (!ms) return '—'
  if (ms >= 60000) return (ms / 60000).toFixed(0) + 'm'
  if (ms >= 1000) return (ms / 1000).toFixed(0) + 's'
  return ms + ''
}
</script>
