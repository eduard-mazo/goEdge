<template>
  <div class="stat-item group">
    <div class="stat-item__label">{{ label }}</div>
    <div class="flex items-center gap-2.5 mt-1">
      <div :class="['led shrink-0', ledClass]" />
      <span :class="['stat-item__value', valueColorClass]">{{ value }}</span>
    </div>
    <div v-if="sub" class="stat-item__unit mt-2">{{ sub }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
const props = defineProps<{ label: string; value: string; color?: string; sub?: string }>()

const ledClass = computed(() => ({
  ok:    'led--green',
  warn:  'led--amber',
  error: 'led--red',
  dim:   'led--dim',
}[props.color ?? 'dim'] ?? 'led--dim'))

const valueColorClass = computed(() => ({
  ok:    'text-[color:var(--signal-ok)]',
  warn:  'text-[color:var(--signal-warn)]',
  error: 'text-[color:var(--signal-fault)]',
  dim:   'text-foreground',
}[props.color ?? 'dim'] ?? 'text-foreground'))
</script>
