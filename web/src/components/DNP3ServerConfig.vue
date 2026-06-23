<template>
  <div class="max-w-xl space-y-5">

    <div class="forge-panel">
      <div class="forge-header">Servidor DNP3 (outstation)</div>
      <div class="p-5 space-y-4">

        <p class="font-sans text-sm text-text-secondary">
          Sirve los datos del campo a un master SCADA por DNP3.
        </p>

        <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary">
          <input type="checkbox" v-model="form.enabled" />
          Habilitar servidor
        </label>

        <Field label="Nombre" hint="opcional">
          <input v-model="form.label" class="forge-input" placeholder="Outstation SCADA" />
        </Field>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <Field label="Escucha" hint="0.0.0.0 = todas">
            <input v-model="form.bindHost" class="forge-input" placeholder="0.0.0.0" />
          </Field>
          <Field label="Puerto" hint="def. 20000">
            <input v-model.number="form.port" class="forge-input" type="number" min="1" max="65535" placeholder="20000" />
          </Field>
        </div>

        <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <Field label="Dir. outstation" hint="típico 1024">
            <input v-model.number="form.localAddress" class="forge-input" type="number" min="0" max="65519" placeholder="1024" />
          </Field>
          <Field label="Dir. master" hint="típico 1">
            <input v-model.number="form.masterAddress" class="forge-input" type="number" min="0" max="65519" placeholder="1" />
          </Field>
        </div>

        <Field label="Buffer de eventos" hint="def. 100">
          <input v-model.number="form.eventBufferSize" class="forge-input max-w-[160px]" type="number" min="0" placeholder="100" />
        </Field>

        <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary">
          <input type="checkbox" v-model="form.allowUnsolicited" />
          Respuestas no solicitadas
        </label>

        <!-- Endpoint preview -->
        <div class="border border-border rounded-sm p-4 font-mono text-[11px] space-y-1.5"
             style="background:var(--tk-surface)">
          <div class="text-muted-foreground mb-2 text-[10px] uppercase tracking-widest font-sans font-semibold">Endpoint</div>
          <div class="flex gap-2 items-baseline">
            <span class="signal-badge shrink-0" :class="form.enabled ? 'signal-badge--on' : 'signal-badge--off'">
              {{ form.enabled ? 'ON' : 'OFF' }}
            </span>
            <span class="text-text-secondary break-all">
              <span class="text-citrico">{{ form.bindHost || '0.0.0.0' }}</span>:<span class="text-foreground">{{ form.port || 20000 }}</span>
              · os <span class="text-citrico">{{ form.localAddress ?? 1024 }}</span>
              ← master <span class="text-citrico">{{ form.masterAddress ?? 1 }}</span>
            </span>
          </div>
          <div class="text-text-dim">{{ servedCount }} señales servidas</div>
        </div>

      </div>
    </div>

    <div class="flex items-center gap-4">
      <button class="btn-primary" :disabled="saving" @click="save">
        {{ saving ? 'Guardando…' : 'Guardar' }}
      </button>
      <span v-if="saved" class="font-mono text-xs text-green-bright">✓ Guardado</span>
      <span v-if="error" class="font-mono text-xs text-red-bright">⚠ {{ error }}</span>
    </div>

    <p class="font-mono text-[10px] text-text-dim">Aplica al reiniciar el gateway.</p>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { api, type DNP3OutstationServer } from '@/api/client'
import { useGatewayStore } from '@/stores/gateway'
import Field from './Field.vue'

const store  = useGatewayStore()
const saving = ref(false)
const saved  = ref(false)
const error  = ref('')

const form = ref<DNP3OutstationServer>({
  enabled: false, label: '', bindHost: '0.0.0.0', port: 20000,
  localAddress: 1024, masterAddress: 1, allowUnsolicited: false, eventBufferSize: 100,
})

const servedCount = computed(() => store.mappings.filter(m => m.enabled && m.serveDnp3).length)

onMounted(async () => {
  try { form.value = await api.getDNP3Server() } catch {}
  if (!store.mappings.length) { try { await store.loadMappings() } catch {} }
})

async function save() {
  saving.value = true; saved.value = false; error.value = ''
  try {
    await api.setDNP3Server(form.value)
    saved.value = true
    setTimeout(() => { saved.value = false }, 4000)
  } catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
  finally { saving.value = false }
}
</script>
