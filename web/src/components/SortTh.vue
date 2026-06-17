<!-- SortTh — a sortable table header cell.
     Renders a real <th> (so the global table styling applies) and guarantees the
     header's alignment matches its column's data alignment via the `align` prop —
     right for numeric columns so digits line up under the label. The sort glyph
     always occupies space (idle ⇅ → active ▲/▼) so toggling never shifts the
     layout. Wire it with :sort + col + @sort="toggle". -->
<template>
  <th
    :class="[alignClass, sortable ? 'sort-th' : '', active ? 'sort-th--active' : '']"
    :aria-sort="active ? (sort.dir === 1 ? 'ascending' : 'descending') : undefined"
    @click="sortable && emit('sort', col)"
  >
    <span class="sort-th__inner">
      <slot>{{ label }}</slot>
      <span v-if="sortable" class="sort-th__glyph">{{ glyph }}</span>
    </span>
  </th>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { SortState } from '@/composables/useSort'

const props = withDefaults(
  defineProps<{
    col: string
    sort: SortState
    label?: string
    align?: 'left' | 'right' | 'center'
    sortable?: boolean
  }>(),
  { align: 'left', sortable: true },
)

const emit = defineEmits<{ sort: [col: string] }>()

const active = computed(() => props.sortable && props.sort.col === props.col)
const glyph = computed(() => (active.value ? (props.sort.dir === 1 ? '▲' : '▼') : '⇅'))
const alignClass = computed(
  () => ({ left: 'sort-th--l', right: 'sort-th--r', center: 'sort-th--c' })[props.align],
)
</script>

<style scoped>
.sort-th--l { text-align: left; }
.sort-th--r { text-align: right; }
.sort-th--c { text-align: center; }

.sort-th__inner {
  display: inline-flex;
  align-items: center;
  gap: 5px;
}
.sort-th--r .sort-th__inner { flex-direction: row-reverse; }

.sort-th {
  cursor: pointer;
  user-select: none;
  transition: color 0.12s, background 0.12s;
}
.sort-th:hover {
  color: var(--foreground);
  background: color-mix(in srgb, var(--epm-bosque) 7%, var(--muted));
}
.sort-th--active { color: var(--epm-bosque); }

.sort-th__glyph {
  display: inline-block;
  min-width: 0.7em;
  font-size: 9px;
  line-height: 1;
  color: var(--tk-border-bright);
  transition: color 0.12s, opacity 0.12s;
  opacity: 0.55;
}
.sort-th:hover .sort-th__glyph { opacity: 1; }
.sort-th--active .sort-th__glyph {
  color: var(--epm-bosque);
  opacity: 1;
}
</style>
