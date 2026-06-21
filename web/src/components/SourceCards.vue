<!-- SourceCards — connection-status cards for every polled source across all
     three protocols (DNP3 outstations, Modbus/TCP and Modbus RTU devices).
     GatewayStatus.outstations carries no protocol field, so the protocol is
     derived by cross-referencing each status id against the configured device
     lists (modbusDevices → MB/TCP, serialDevices → RTU, else DNP3). -->
<template>
  <section>
    <div class="flex items-center gap-3 mb-3">
      <h3 class="font-sans font-bold text-xs uppercase tracking-widest text-text-secondary">Fuentes</h3>
      <div class="rule-brand flex-1" />
      <!-- per-protocol up/total chips -->
      <div class="flex items-center gap-1.5">
        <span v-for="p in protoCounts.filter(c => c.total)" :key="p.v"
              class="proto-tag" :class="`proto-tag--${p.cls}`" :title="`${p.label}: ${p.up} en línea / ${p.total}`">
          {{ p.label }} {{ p.up }}/{{ p.total }}
        </span>
      </div>
    </div>

    <!-- protocol filter (only when more than one protocol is present) -->
    <div v-if="protosPresent.length > 1" class="seg mb-3">
      <button v-for="p in filterOpts" :key="p.v"
              class="seg-btn" :class="protocol === p.v ? 'seg-btn--active' : ''"
              @click="protocol = p.v">{{ p.l }}</button>
    </div>

    <div v-if="!filtered.length" class="forge-panel text-center py-10">
      <p class="font-mono text-xs text-text-dim">
        {{ sources.length ? 'Sin fuentes para el filtro.' : 'Sin fuentes — agrégalas en Estaciones, Modbus/TCP o Serial' }}
      </p>
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
      <div
        v-for="o in filtered" :key="o.id"
        class="forge-panel p-4 flex items-start gap-3 transition-all duration-200 hover:shadow-md hover:border-[color:var(--tk-border-bright)]"
      >
        <div :class="['led mt-1 shrink-0', o.connected ? 'led--green' : 'led--red']" />
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2 min-w-0">
            <span class="font-sans font-bold text-sm truncate">{{ o.label || o.id }}</span>
            <span class="proto-tag shrink-0" :class="`proto-tag--${o.cls}`">{{ o.protoLabel }}</span>
          </div>
          <div class="font-mono text-[11px] text-text-secondary mt-0.5 truncate" :title="o.addr">{{ o.addr }}</div>
          <div v-if="o.lastError && !o.connected" class="font-mono text-[10px] text-red-base mt-1 truncate" :title="o.lastError">
            {{ o.lastError }}
          </div>
          <div class="flex gap-4 mt-2.5 font-mono text-[11px]">
            <span class="text-text-dim">
              rx <span class="text-foreground font-medium">{{ o.measurementsRx ?? 0 }}</span>
            </span>
            <span class="text-text-dim" v-if="isRealTime(o.lastReadAt)">
              últ <span class="text-foreground font-medium">{{ relTime(o.lastReadAt) }}</span>
            </span>
            <span class="text-text-dim" v-else>sin lecturas</span>
          </div>
        </div>
        <span :class="['signal-badge self-start shrink-0', o.connected ? 'signal-badge--on' : 'signal-badge--off']">
          {{ o.connected ? 'OK' : 'FALLA' }}
        </span>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useGatewayStore } from '@/stores/gateway'
import { relTime, isRealTime } from '@/lib/time'

type Proto = 'dnp3' | 'modbus' | 'modbusrtu'

const store = useGatewayStore()
const protocol = ref<'' | Proto>('')

// Derive a source's protocol from the configured device lists; default DNP3
// (covers outstations and any source not yet in the config refs).
function protoOf(id: string): Proto {
  if (store.modbusDevices.some((d) => d.id === id)) return 'modbus'
  if (store.serialDevices.some((d) => d.id === id)) return 'modbusrtu'
  return 'dnp3'
}
function meta(p: Proto) {
  return p === 'modbus'
    ? { cls: 'mb', label: 'MB/TCP', protoLabel: 'MB' }
    : p === 'modbusrtu'
      ? { cls: 'rtu', label: 'RTU', protoLabel: 'RTU' }
      : { cls: 'dnp', label: 'DNP3', protoLabel: 'DNP3' }
}

const sources = computed(() =>
  Object.values(store.status?.outstations ?? {})
    .map((o) => {
      const p = protoOf(o.id)
      return { ...o, proto: p, ...meta(p) }
    })
    .sort((a, b) => (a.label || a.id).localeCompare(b.label || b.id)),
)

const protoCounts = computed(() =>
  (['dnp3', 'modbus', 'modbusrtu'] as Proto[]).map((p) => {
    const list = sources.value.filter((s) => s.proto === p)
    return { v: p, cls: meta(p).cls, label: meta(p).protoLabel, total: list.length, up: list.filter((s) => s.connected).length }
  }),
)
const protosPresent = computed(() => protoCounts.value.filter((c) => c.total).map((c) => c.v))
const filterOpts = computed(() => [
  { v: '' as const, l: 'Todas' },
  ...protosPresent.value.map((v) => ({ v, l: meta(v).protoLabel })),
])
const filtered = computed(() => (protocol.value ? sources.value.filter((s) => s.proto === protocol.value) : sources.value))
</script>

<style scoped>
/* segmented control (matches LiveValues / MappingTable) */
.seg { display: inline-flex; border: 1.5px solid var(--border); border-radius: var(--radius); overflow: hidden; background: var(--card); }
.seg-btn { font-family: var(--font-sans); font-weight: 700; font-size: 11px; letter-spacing: 0.03em; padding: 6px 14px; color: var(--muted-foreground); background: transparent; border: none; cursor: pointer; transition: background 0.12s, color 0.12s; }
.seg-btn + .seg-btn { border-left: 1.5px solid var(--border); }
.seg-btn:hover { color: var(--foreground); }
.seg-btn--active { background: color-mix(in srgb, var(--epm-bosque) 12%, transparent); color: var(--epm-bosque); }

/* protocol tag */
.proto-tag { font-family: var(--font-mono); font-size: 9.5px; font-weight: 600; letter-spacing: 0.06em; padding: 2px 6px; border-radius: 3px; border: 1px solid; white-space: nowrap; }
.proto-tag--dnp { color: var(--epm-bosque); border-color: color-mix(in srgb, var(--epm-bosque) 40%, transparent); background: color-mix(in srgb, var(--epm-bosque) 8%, transparent); }
.proto-tag--mb  { color: var(--tk-amber-bright); border-color: color-mix(in srgb, var(--tk-amber-base) 45%, transparent); background: color-mix(in srgb, var(--tk-amber-base) 10%, transparent); }
.proto-tag--rtu { color: var(--tk-amber-bright); border-color: color-mix(in srgb, var(--tk-amber-base) 55%, transparent); background: color-mix(in srgb, var(--tk-amber-base) 16%, transparent); letter-spacing: 0.1em; }
</style>
