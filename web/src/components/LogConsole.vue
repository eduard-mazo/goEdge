<template>
  <div class="space-y-3 h-full flex flex-col">

    <div class="flex items-center gap-3 shrink-0">
      <p class="font-sans text-sm text-text-secondary flex-1">Eventos del gateway en vivo</p>
      <span class="font-mono text-[10px] text-text-dim">{{ store.logs.length }} entradas</span>
      <button class="btn-ghost text-xs py-1 px-3" @click="store.logs.length = 0">Limpiar</button>
    </div>

    <!-- Log terminal -->
    <div
      class="forge-panel flex-1 min-h-0 overflow-y-auto font-mono text-[12px] scanline-overlay"
      style="min-height:60vh; background:#0E1C12; border-color: rgba(159,207,103,0.15);"
      ref="logEl"
    >
      <div class="p-4 space-y-[3px]">
        <!-- Cursor blink when empty -->
        <div v-if="!store.logs.length" class="flex items-center gap-2" style="color:#5A7860">
          <span style="color:#7DC850">›</span>
          <span>Esperando eventos…</span>
          <span class="inline-block w-[6px] h-[13px] animate-pulse" style="background:#5A7860;opacity:0.7" />
        </div>

        <div
          v-for="(entry, i) in store.logs"
          :key="i"
          class="flex gap-3 leading-relaxed relative z-10"
        >
          <span class="shrink-0 tabular-nums select-none" style="color:#3D6045">{{ entry.time }}</span>
          <span
            class="shrink-0 w-[38px] font-semibold text-[10px] uppercase tracking-wider"
            :style="levelStyle(entry.level)"
          >{{ entry.level }}</span>
          <span :style="msgStyle(entry.level)">{{ entry.message }}</span>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { useGatewayStore } from '@/stores/gateway'

const store = useGatewayStore()
const logEl = ref<HTMLElement>()

watch(() => store.logs.length, async () => {
  await nextTick()
  if (logEl.value) logEl.value.scrollTop = logEl.value.scrollHeight
})

function levelStyle(level: string) {
  return {
    error: 'color:#F07060',
    warn:  'color:#F0B848',
    info:  'color:#7DC850',
    debug: 'color:#5A7860',
  }[level] ?? 'color:#5A7860'
}

function msgStyle(level: string) {
  return {
    error: 'color:#F8A898',
    warn:  'color:#E0D090',
    info:  'color:#C8E0B0',
    debug: 'color:#5A7860',
  }[level] ?? 'color:#8EA88A'
}
</script>
