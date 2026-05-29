<template>
  <div class="space-y-1.5">
    <label class="forge-label">{{ label }}</label>
    <slot />
    <p v-if="hint" class="font-sans text-[11px] text-text-secondary mt-1 leading-snug">{{ hint }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps<{ label: string; hint?: string }>()
</script>
