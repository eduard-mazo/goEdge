<template>
  <div class="flex flex-col">

    <!-- Panel header -->
    <div class="flex items-center justify-between px-5 py-3 border-b border-border">
      <div class="flex items-center gap-3">
        <div class="rule-brand w-6 self-stretch shrink-0" style="width:2px;height:auto;display:block" />
        <h3 class="font-sans font-bold text-sm tracking-widest uppercase text-foreground">
          Configuración rápida
        </h3>
        <span class="font-mono text-[10px] text-text-dim">MQTT · Sparkplug B</span>
      </div>
      <button class="btn-ghost text-xs py-1 px-3" @click="$emit('close')">✕ Cerrar</button>
    </div>

    <div class="p-5 grid grid-cols-1 lg:grid-cols-2 gap-6">

      <!-- MQTT config -->
      <section class="space-y-3">
        <p class="forge-label">Broker MQTT</p>
        <div class="space-y-2">
          <div>
            <label class="forge-label">Broker URI</label>
            <input v-model="mqtt.broker" class="forge-input" placeholder="tcp://localhost:1883" />
          </div>
          <div class="grid grid-cols-2 gap-2">
            <div>
              <label class="forge-label">Client ID</label>
              <input v-model="mqtt.clientId" class="forge-input" />
            </div>
            <div>
              <label class="forge-label">QoS</label>
              <select v-model.number="mqtt.qos" class="forge-input">
                <option :value="0">0 – Máx. una vez</option>
                <option :value="1">1 – Al menos una vez</option>
                <option :value="2">2 – Exactamente una vez</option>
              </select>
            </div>
          </div>
          <div class="grid grid-cols-2 gap-2">
            <div>
              <label class="forge-label">Usuario</label>
              <input v-model="mqtt.username" class="forge-input" autocomplete="off" />
            </div>
            <div>
              <label class="forge-label">Contraseña</label>
              <input v-model="mqtt.password" class="forge-input" type="password" autocomplete="off" />
            </div>
          </div>
        </div>
        <button class="btn-primary text-xs py-1.5" :disabled="saving.mqtt" @click="saveMQTT">
          {{ saving.mqtt ? 'Guardando…' : 'Guardar MQTT' }}
        </button>
        <span v-if="saved.mqtt" class="font-mono text-xs text-green-bright ml-3">✓ guardado</span>
      </section>

      <!-- Sparkplug config -->
      <section class="space-y-3">
        <p class="forge-label">Nodo Sparkplug B</p>
        <div class="space-y-2">
          <div>
            <label class="forge-label">Group ID</label>
            <input v-model="sp.groupId" class="forge-input" placeholder="plant-floor" />
          </div>
          <div>
            <label class="forge-label">Node ID</label>
            <input v-model="sp.nodeId" class="forge-input" placeholder="dnp3-gw" />
          </div>
          <!-- Topic preview -->
          <div class="font-mono text-[10px] text-text-secondary border border-border rounded-md p-3 space-y-1 leading-relaxed"
               style="background:var(--tk-surface)">
            <div>NBIRTH → <span class="text-citrico">spBv1.0/{{ sp.groupId || '…' }}/NBIRTH/{{ sp.nodeId || '…' }}</span></div>
            <div>NDATA  → <span class="text-foreground">spBv1.0/{{ sp.groupId || '…' }}/NDATA/{{ sp.nodeId || '…' }}</span></div>
          </div>
        </div>
        <button class="btn-primary text-xs py-1.5" :disabled="saving.sp" @click="saveSP">
          {{ saving.sp ? 'Guardando…' : 'Guardar Sparkplug' }}
        </button>
        <span v-if="saved.sp" class="font-mono text-xs text-green-bright ml-3">✓ guardado</span>
      </section>

    </div>

    <div v-if="error" class="px-5 pb-4 font-mono text-xs text-red-bright">⚠ {{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type MQTTConfig, type SparkplugConfig } from '@/api/client'

defineEmits<{ close: [] }>()

const error = ref('')
const saving = ref({ mqtt: false, sp: false })
const saved  = ref({ mqtt: false, sp: false })

const mqtt = ref<MQTTConfig>({
  broker: 'tcp://localhost:1883', clientId: 'goMqttDnp3',
  username: '', password: '', qos: 1, keepalive: 60,
  tls: { enabled: false },
})
const sp = ref<SparkplugConfig>({ groupId: 'plant-floor', nodeId: 'dnp3-gw' })

onMounted(async () => {
  try { mqtt.value = await api.getMQTT() } catch {}
  try { sp.value   = await api.getSparkplug() } catch {}
})

async function saveMQTT() {
  error.value = ''; saving.value.mqtt = true; saved.value.mqtt = false
  try {
    await api.setMQTT(mqtt.value)
    saved.value.mqtt = true
    setTimeout(() => { saved.value.mqtt = false }, 3000)
  } catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
  finally { saving.value.mqtt = false }
}

async function saveSP() {
  error.value = ''; saving.value.sp = true; saved.value.sp = false
  try {
    await api.setSparkplug(sp.value)
    saved.value.sp = true
    setTimeout(() => { saved.value.sp = false }, 3000)
  } catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
  finally { saving.value.sp = false }
}
</script>
