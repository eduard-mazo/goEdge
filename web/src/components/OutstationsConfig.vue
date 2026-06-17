<template>
  <div class="space-y-5">

    <div class="flex items-center gap-4">
      <p class="font-sans text-sm text-text-secondary flex-1">
        Estaciones DNP3 sobre TCP, cada una con su cadencia de sondeo.
      </p>
      <button class="btn-primary text-xs py-1.5" @click="openAdd">+ Agregar</button>
    </div>

    <!-- Empty state -->
    <div v-if="!store.outstations.length"
         class="forge-panel text-center py-12">
      <p class="font-mono text-xs text-text-dim">Sin estaciones DNP3.</p>
      <button class="btn-ghost mt-4 text-xs" @click="openAdd">Agregar la primera</button>
    </div>

    <!-- Outstations table -->
    <div v-else class="forge-panel overflow-x-auto">
      <table class="min-w-[900px]">
        <thead>
          <tr>
            <SortTh col="id"     label="ID"                :sort="sort" @sort="toggle" />
            <SortTh col="label"  label="Nombre"            :sort="sort" @sort="toggle" />
            <SortTh col="addr"   label="Host : Puerto"     :sort="sort" @sort="toggle" />
            <SortTh col="link"   label="Maestro ↔ Estación" :sort="sort" @sort="toggle" />
            <th>Sondeos (ms)</th>
            <th>Unsol</th>
            <SortTh col="estado" label="Estado" align="center" :sort="sort" @sort="toggle" />
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="o in sorted" :key="o.id">
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
              <span v-else class="text-text-dim">no</span>
            </td>
            <td class="text-center">
              <span :class="['signal-badge', o.enabled ? 'signal-badge--on' : 'signal-badge--off']">
                {{ o.enabled ? 'activa' : 'inactiva' }}
              </span>
            </td>
            <td>
              <div class="flex gap-2 justify-end">
                <button class="btn-ghost text-xs py-0.5 px-3" @click="openEdit(o)">Editar</button>
                <button
                  class="btn-ghost text-xs py-0.5 px-3 !border-red-base/40 !text-red-base hover:!border-red-bright hover:!text-red-bright"
                  @click="del(o.id)"
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
            <span>{{ editId ? 'Editar estación' : 'Nueva estación DNP3' }}</span>
            <button type="button" class="modal-x" @click="modal = false" aria-label="Cerrar">✕</button>
          </div>
          <form @submit.prevent="saveOutstation" class="p-5 space-y-4">

            <div class="grid grid-cols-2 gap-4">
              <Field label="ID (identificador único)">
                <input v-model="form.id" class="forge-input" :disabled="!!editId" required placeholder="rtu-01" />
              </Field>
              <Field label="Nombre">
                <input v-model="form.label" class="forge-input" placeholder="RTU subestación 01" />
              </Field>
            </div>

            <div class="grid grid-cols-3 gap-4">
              <Field label="Host" class="col-span-2">
                <input v-model="form.host" class="forge-input" placeholder="192.168.1.100" required />
              </Field>
              <Field label="Puerto" hint="DNP3/IP por defecto 20000">
                <input v-model.number="form.port" class="forge-input" type="number" placeholder="20000" />
              </Field>
            </div>

            <fieldset class="space-y-3">
              <legend class="font-mono text-[10px] uppercase tracking-[0.18em] text-text-dim">Capa de enlace</legend>
              <div class="grid grid-cols-2 gap-4">
                <Field label="Dir. maestro" hint="Dirección local (típico: 1)">
                  <input v-model.number="form.masterAddress" class="forge-input" type="number" min="1" max="65519" required />
                </Field>
                <Field label="Dir. estación" hint="Dirección remota (típico: 1024+)">
                  <input v-model.number="form.outstationAddress" class="forge-input" type="number" min="1" max="65519" required />
                </Field>
              </div>
            </fieldset>

            <fieldset class="space-y-3">
              <legend class="font-mono text-[10px] uppercase tracking-[0.18em] text-text-dim">Capa de aplicación</legend>
              <div class="grid grid-cols-2 gap-4">
                <Field label="Timeout respuesta (ms)" hint="por defecto 5000">
                  <input v-model.number="form.responseTimeoutMs" class="forge-input" type="number" placeholder="5000" />
                </Field>
                <Field label="Keep-Alive (ms)" hint="por defecto 60000; 0 = off">
                  <input v-model.number="form.keepAliveMs" class="forge-input" type="number" placeholder="60000" />
                </Field>
              </div>
            </fieldset>

            <fieldset class="space-y-3">
              <legend class="font-mono text-[10px] uppercase tracking-[0.18em] text-text-dim">
                Sondeos periódicos (0 = off)
              </legend>
              <div class="grid grid-cols-4 gap-3">
                <Field label="Integridad (ms)" hint="todas las clases">
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
              <legend class="font-mono text-[10px] uppercase tracking-[0.18em] text-text-dim">Respuestas no solicitadas</legend>
              <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary">
                <input type="checkbox" v-model="form.unsolicitedEnabled" />
                Habilitar no solicitadas al iniciar
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
                Enviar DISABLE_UNSOLICITED antes del sondeo inicial
              </label>
              <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary">
                <input type="checkbox" v-model="form.startupIntegrity" />
                Sondeo de integridad al conectar (recomendado)
              </label>
            </fieldset>

            <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary pt-2 border-t border-border">
              <input type="checkbox" v-model="form.enabled" /> Estación habilitada
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
import { api, type DNP3Outstation } from '@/api/client'
import Field from './Field.vue'
import SortTh from './SortTh.vue'
import { useSort } from '@/composables/useSort'

const store  = useGatewayStore()

const { sort, toggle, sorted } = useSort(() => store.outstations, {
  initial: 'id',
  accessors: {
    id:     (o) => o.id,
    label:  (o) => o.label ?? '',
    addr:   (o) => `${o.host}:${o.port ?? 20000}`,
    link:   (o) => o.masterAddress ?? 0,
    estado: (o) => (o.enabled ? 1 : 0),
  },
})
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
  if (!confirm(`¿Eliminar la estación "${id}" y todas sus señales?`)) return
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
