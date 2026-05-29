<template>
  <div class="max-w-2xl space-y-5">

    <div class="forge-panel">
      <div class="forge-header">MQTT Broker Connection</div>
      <div class="p-5 space-y-4">

        <div class="grid grid-cols-2 gap-4">
          <Field label="Broker URI" hint="tcp://host:port  ·  ssl://host:port">
            <input v-model="form.broker" class="forge-input" placeholder="tcp://localhost:1883" required />
          </Field>
          <Field label="Client ID">
            <input v-model="form.clientId" class="forge-input" placeholder="goMqttDnp3" required />
          </Field>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <Field label="Username">
            <input v-model="form.username" class="forge-input" autocomplete="off" />
          </Field>
          <Field label="Password">
            <input v-model="form.password" class="forge-input" type="password" autocomplete="off" />
          </Field>
        </div>

        <div class="grid grid-cols-2 gap-4">
          <Field label="QoS">
            <select v-model.number="form.qos" class="forge-input">
              <option :value="0">0 – At most once</option>
              <option :value="1">1 – At least once</option>
              <option :value="2">2 – Exactly once</option>
            </select>
          </Field>
          <Field label="Keepalive (s)">
            <input v-model.number="form.keepalive" class="forge-input" type="number" min="0" />
          </Field>
        </div>

      </div>
    </div>

    <!-- TLS -->
    <div class="forge-panel">
      <div class="forge-header flex items-center gap-3">
        <label class="flex items-center gap-2 cursor-pointer">
          <input type="checkbox" v-model="form.tls.enabled" />
          TLS / SSL
        </label>
        <span class="font-mono text-[10px] text-text-dim font-400 normal-case tracking-normal">
          {{ form.tls.enabled ? 'enabled' : 'disabled — plain TCP' }}
        </span>
      </div>
      <div v-if="form.tls.enabled" class="p-5 space-y-4">
        <Field label="CA Certificate file">
          <input v-model="form.tls.caFile" class="forge-input" placeholder="/etc/certs/ca.pem" />
        </Field>
        <div class="grid grid-cols-2 gap-4">
          <Field label="Client cert file">
            <input v-model="form.tls.certFile" class="forge-input" placeholder="/etc/certs/client.crt" />
          </Field>
          <Field label="Client key file">
            <input v-model="form.tls.keyFile" class="forge-input" placeholder="/etc/certs/client.key" />
          </Field>
        </div>
        <label class="flex items-center gap-2 cursor-pointer font-mono text-xs text-signal-warn">
          <input type="checkbox" v-model="form.tls.insecure" />
          Skip server cert verification (insecure — dev only)
        </label>
      </div>
    </div>

    <div class="flex items-center gap-4">
      <button class="btn-primary" :disabled="saving" @click="save">
        {{ saving ? 'Saving…' : 'Save Changes' }}
      </button>
      <span v-if="saved" class="font-mono text-xs text-green-bright">✓ Configuration saved</span>
      <span v-if="error" class="font-mono text-xs text-red-bright">⚠ {{ error }}</span>
    </div>

    <p class="font-mono text-[10px] text-text-dim">
      Changes take effect on next gateway start.
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type MQTTConfig } from '@/api/client'
import Field from './Field.vue'

const saving = ref(false)
const saved  = ref(false)
const error  = ref('')

const form = ref<MQTTConfig>({
  broker: 'tcp://localhost:1883', clientId: 'goMqttDnp3',
  username: '', password: '', qos: 1, keepalive: 60,
  tls: { enabled: false },
})

onMounted(async () => { try { form.value = await api.getMQTT() } catch {} })

async function save() {
  saving.value = true; saved.value = false; error.value = ''
  try {
    await api.setMQTT(form.value)
    saved.value = true
    setTimeout(() => { saved.value = false }, 4000)
  } catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
  finally { saving.value = false }
}
</script>
