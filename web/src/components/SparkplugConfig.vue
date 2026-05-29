<template>
  <div class="max-w-xl space-y-5">

    <div class="forge-panel">
      <div class="forge-header">Sparkplug B Namespace</div>
      <div class="p-5 space-y-4">

        <Field label="Group ID" hint="Logical site/plant group (e.g. plant-floor)">
          <input v-model="form.groupId" class="forge-input" placeholder="plant-floor" required />
        </Field>

        <Field label="Node ID" hint="Unique EoN node identifier on this broker">
          <input v-model="form.nodeId" class="forge-input" placeholder="dnp3-gw-01" required />
        </Field>

        <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary">
          <input type="checkbox" v-model="form.birthOnConfigChange" />
          Re-publish NBIRTH when configuration changes (while running)
        </label>

        <!-- Topic preview box -->
        <div class="border border-border rounded-sm p-4 space-y-1.5 font-mono text-[11px]"
             style="background:var(--tk-surface)">
          <div class="text-muted-foreground mb-2 text-[10px] uppercase tracking-widest font-sans font-semibold">Topic Preview</div>
          <div class="flex gap-2">
            <span class="signal-badge signal-badge--warn shrink-0">NBIRTH</span>
            <span class="text-text-secondary truncate">
              spBv1.0/<span class="text-citrico">{{ form.groupId || '…' }}</span>/NBIRTH/<span class="text-foreground">{{ form.nodeId || '…' }}</span>
            </span>
          </div>
          <div class="flex gap-2">
            <span class="signal-badge shrink-0">NDATA</span>
            <span class="text-text-secondary truncate">
              spBv1.0/<span class="text-citrico">{{ form.groupId || '…' }}</span>/NDATA/<span class="text-foreground">{{ form.nodeId || '…' }}</span>
            </span>
          </div>
          <div class="flex gap-2">
            <span class="signal-badge signal-badge--off shrink-0">NDEATH</span>
            <span class="text-text-secondary truncate">
              spBv1.0/<span class="text-citrico">{{ form.groupId || '…' }}</span>/NDEATH/<span class="text-foreground">{{ form.nodeId || '…' }}</span>
            </span>
          </div>
          <div class="flex gap-2">
            <span class="signal-badge signal-badge--on shrink-0">DDATA</span>
            <span class="text-text-secondary truncate">
              spBv1.0/…/DDATA/{{ form.nodeId || '…' }}/<span class="text-citrico">{deviceId}</span>
            </span>
          </div>
        </div>

      </div>
    </div>

    <div class="flex items-center gap-4">
      <button class="btn-primary" :disabled="saving" @click="save">
        {{ saving ? 'Saving…' : 'Save Changes' }}
      </button>
      <span v-if="saved" class="font-mono text-xs text-green-bright">✓ Saved</span>
      <span v-if="error" class="font-mono text-xs text-red-bright">⚠ {{ error }}</span>
    </div>

    <p class="font-mono text-[10px] text-text-dim">
      Changes take effect on next gateway start. Restart triggers a new NBIRTH sequence.
    </p>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api, type SparkplugConfig } from '@/api/client'
import Field from './Field.vue'

const saving = ref(false)
const saved  = ref(false)
const error  = ref('')

const form = ref<SparkplugConfig>({ groupId: 'plant-floor', nodeId: 'dnp3-gw' })

onMounted(async () => { try { form.value = await api.getSparkplug() } catch {} })

async function save() {
  saving.value = true; saved.value = false; error.value = ''
  try {
    await api.setSparkplug(form.value)
    saved.value = true
    setTimeout(() => { saved.value = false }, 4000)
  } catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
  finally { saving.value = false }
}
</script>
