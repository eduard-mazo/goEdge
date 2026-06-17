<!-- LiveValues — operator live-data grid.
     Joins configured mappings × per-source status × last published readings into
     a dense, searchable/filterable SCADA-style table. Values flash on update. -->
<template>
  <div class="space-y-4">

    <!-- ── Summary strip ─────────────────────────────────────────── -->
    <div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
      <div class="stat-item">
        <div class="stat-item__label">Puntos</div>
        <div class="stat-item__value">{{ rows.length }}</div>
        <div class="stat-item__unit">{{ filtered.length }} visibles</div>
      </div>
      <div class="stat-item">
        <div class="stat-item__label">En vivo</div>
        <div class="stat-item__value text-bosque">{{ liveCount }}</div>
        <div class="stat-item__unit">con valor</div>
      </div>
      <div class="stat-item">
        <div class="stat-item__label">Fuentes</div>
        <div class="stat-item__value">{{ sourcesUp }}<span class="text-text-dim">/{{ sources.length }}</span></div>
        <div class="stat-item__unit">{{ dnp3Count }} dnp3 · {{ modbusCount }} modbus</div>
      </div>
      <div class="stat-item">
        <div class="stat-item__label">En falla</div>
        <div class="stat-item__value" :class="faultCount ? 'text-[color:var(--signal-fault)]' : ''">{{ faultCount }}</div>
        <div class="stat-item__unit">sin conexión</div>
      </div>
    </div>

    <!-- ── Control bar ───────────────────────────────────────────── -->
    <div class="forge-panel p-3 flex flex-wrap items-center gap-2">
      <!-- search -->
      <div class="relative flex-1 min-w-[180px]">
        <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-text-dim pointer-events-none" />
        <input
          v-model="q"
          type="text"
          placeholder="Buscar métrica, fuente…"
          class="forge-input pl-8 font-mono text-[12px]"
        />
      </div>

      <!-- protocol segmented -->
      <div class="seg">
        <button v-for="p in protocolOpts" :key="p.v"
                class="seg-btn" :class="protocol === p.v ? 'seg-btn--active' : ''"
                @click="protocol = p.v">{{ p.l }}</button>
      </div>

      <!-- source filter -->
      <select v-model="sourceFilter" class="forge-input w-auto min-w-[130px] font-mono text-[12px]">
        <option value="">Todas</option>
        <option v-for="s in sources" :key="s.id" :value="s.id">{{ s.label || s.id }}</option>
      </select>

      <!-- status filter -->
      <div class="seg">
        <button v-for="s in statusOpts" :key="s.v"
                class="seg-btn" :class="statusFilter === s.v ? 'seg-btn--active' : ''"
                @click="statusFilter = s.v">{{ s.l }}</button>
      </div>

      <!-- live pulse -->
      <div class="ml-auto flex items-center gap-1.5 font-mono text-[10px] uppercase tracking-wider"
           :class="store.isRunning ? 'text-bosque' : 'text-text-dim'">
        <span class="led" :class="store.isRunning ? 'led--green' : 'led--dim'" />
        {{ store.isRunning ? 'en vivo' : 'inactivo' }}
      </div>
    </div>

    <!-- ── Data grid ─────────────────────────────────────────────── -->
    <div class="forge-panel overflow-hidden scanline-overlay">
      <div class="overflow-auto max-h-[calc(100vh-340px)]">
        <table class="w-full">
          <thead class="sticky top-0 z-10">
            <tr>
              <th class="w-7"></th>
              <th class="cursor-pointer select-none" @click="sortBy('metric')">
                Métrica <SortGlyph :col="'metric'" :sort="sort" />
              </th>
              <th class="cursor-pointer select-none" @click="sortBy('source')">
                Fuente <SortGlyph :col="'source'" :sort="sort" />
              </th>
              <th>Proto</th>
              <th>Punto</th>
              <th class="text-right cursor-pointer select-none" @click="sortBy('value')">
                Valor <SortGlyph :col="'value'" :sort="sort" />
              </th>
              <th class="text-right">Actualizado</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in filtered" :key="r.metric"
                :class="flashing.has(r.metric) ? 'row-flash' : ''">
              <!-- connection LED -->
              <td class="text-center">
                <span class="led inline-block" :class="ledClass(r)" :title="r.connected ? 'source online' : 'source offline'" />
              </td>
              <!-- metric -->
              <td>
                <div class="font-mono text-[12.5px] text-foreground truncate max-w-[260px]">{{ r.metric }}</div>
                <div v-if="r.unit" class="font-mono text-[10px] text-text-dim">{{ r.unit }}</div>
              </td>
              <!-- source -->
              <td class="font-mono text-[12px] text-text-secondary truncate max-w-[150px]" :title="r.source">
                {{ r.sourceLabel }}
              </td>
              <!-- protocol -->
              <td>
                <span class="proto-tag" :class="`proto-tag--${protoCls(r.protocol)}`">
                  {{ protoLabel(r.protocol) }}
                </span>
              </td>
              <!-- point identity -->
              <td class="font-mono text-[11px] text-text-secondary whitespace-nowrap">{{ r.point }}</td>
              <!-- value -->
              <td class="text-right">
                <template v-if="r.value === undefined">
                  <span class="font-mono text-[12px] text-text-dim">—</span>
                </template>
                <template v-else-if="isBinary(r)">
                  <span class="signal-badge" :class="r.value ? 'signal-badge--on' : 'signal-badge--off'">
                    {{ r.value ? 'ON' : 'OFF' }}
                  </span>
                </template>
                <template v-else>
                  <span class="font-mono text-[13px] font-semibold tabular-nums"
                        :class="flashing.has(r.metric) ? 'text-bosque' : 'text-foreground'">
                    {{ fmtNum(r.value) }}<span v-if="r.unit" class="text-text-dim font-normal text-[10px] ml-0.5">{{ r.unit }}</span>
                  </span>
                </template>
              </td>
              <!-- updated -->
              <td class="text-right font-mono text-[10.5px] text-text-dim whitespace-nowrap" :title="r.lastReadAt || ''">{{ ago(r.lastReadAt) }}</td>
            </tr>

            <tr v-if="filtered.length === 0">
              <td colspan="7" class="text-center py-12">
                <div class="text-text-dim font-mono text-[12px]">
                  <template v-if="rows.length === 0">Sin señales. Agrégalas para ver datos en vivo.</template>
                  <template v-else>Sin resultados para el filtro.</template>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <!-- footer rail -->
      <div class="flex items-center justify-between px-3 py-2 border-t border-border bg-muted">
        <span class="font-mono text-[10px] text-text-dim uppercase tracking-wider">
          {{ filtered.length }} / {{ rows.length }} puntos
        </span>
        <span class="font-mono text-[10px] text-text-dim">
          tick {{ tickAgo }}<template v-if="store.status?.publishCount != null"> · pub {{ store.status.publishCount }}</template>
        </span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onUnmounted, h } from 'vue'
import { Search } from 'lucide-vue-next'
import { useGatewayStore } from '@/stores/gateway'
import type { SignalMapping } from '@/api/client'
import { ago as agoFmt } from '@/lib/time'

const store = useGatewayStore()

// ── filters / sort state ──────────────────────────────────────────
const q = ref('')
const protocol = ref<'' | 'dnp3' | 'modbus' | 'modbusrtu'>('')
const sourceFilter = ref('')
const statusFilter = ref<'' | 'live' | 'fault'>('')
const sort = ref<{ col: 'metric' | 'source' | 'value'; dir: 1 | -1 }>({ col: 'metric', dir: 1 })

const protocolOpts = [
  { v: '' as const, l: 'Todos' },
  { v: 'dnp3' as const, l: 'DNP3' },
  { v: 'modbus' as const, l: 'MB' },
  { v: 'modbusrtu' as const, l: 'RTU' },
]

type Proto = 'dnp3' | 'modbus' | 'modbusrtu'
function protoLabel(p: Proto) { return p === 'modbus' ? 'MB' : p === 'modbusrtu' ? 'RTU' : 'DNP3' }
function protoCls(p: Proto) { return p === 'modbus' ? 'mb' : p === 'modbusrtu' ? 'rtu' : 'dnp' }
const statusOpts = [
  { v: '' as const, l: 'Todo' },
  { v: 'live' as const, l: 'En vivo' },
  { v: 'fault' as const, l: 'Falla' },
]

function sortBy(col: 'metric' | 'source' | 'value') {
  if (sort.value.col === col) sort.value.dir = (sort.value.dir * -1) as 1 | -1
  else sort.value = { col, dir: 1 }
}

// ── derived rows ──────────────────────────────────────────────────
interface Row {
  metric: string
  source: string
  sourceLabel: string
  protocol: Proto
  ptype: string // normalized point type for value formatting
  point: string // human point identity
  unit: string
  value: number | undefined
  connected: boolean
  lastReadAt?: string
}

function srcId(m: SignalMapping): string {
  return m.sourceId || m.outstationId || ''
}

const rows = computed<Row[]>(() => {
  const st = store.status
  const readings = st?.lastReadings ?? {}
  const srcStatus = st?.outstations ?? {}
  return store.mappings
    .filter((m) => m.enabled !== false)
    .map((m) => {
      const sid = srcId(m)
      const proto = (m.protocol as Proto) || 'dnp3'
      const ss = srcStatus[sid]
      const isMb = proto === 'modbus' || proto === 'modbusrtu'
      const ptype = isMb
        ? (m.function === 'coil' || m.function === 'discrete_input' ? 'binary' : 'analog')
        : m.pointType
      const point = isMb
        ? `${(m.function ?? '').replace(/_/g, ' ')} @${m.address ?? 0}`
        : `${m.pointType} #${m.index}`
      return {
        metric: m.metricName,
        source: sid,
        sourceLabel: ss?.label || sid,
        protocol: proto,
        ptype,
        point,
        unit: m.engineeringUnit ?? '',
        value: readings[m.metricName],
        connected: ss?.connected ?? false,
        lastReadAt: ss?.lastReadAt,
      }
    })
})

// ── unique sources for the dropdown ───────────────────────────────
const sources = computed(() => {
  const seen = new Map<string, { id: string; label: string }>()
  for (const r of rows.value) {
    if (!seen.has(r.source)) seen.set(r.source, { id: r.source, label: r.sourceLabel })
  }
  return [...seen.values()].sort((a, b) => (a.label || a.id).localeCompare(b.label || b.id))
})

// ── filtering + sorting ───────────────────────────────────────────
const filtered = computed<Row[]>(() => {
  const term = q.value.trim().toLowerCase()
  let out = rows.value.filter((r) => {
    if (protocol.value && r.protocol !== protocol.value) return false
    if (sourceFilter.value && r.source !== sourceFilter.value) return false
    if (statusFilter.value === 'live' && !(r.connected && r.value !== undefined)) return false
    if (statusFilter.value === 'fault' && r.connected) return false
    if (term) {
      const hay = `${r.metric} ${r.sourceLabel} ${r.source} ${r.unit} ${r.point}`.toLowerCase()
      if (!hay.includes(term)) return false
    }
    return true
  })
  const { col, dir } = sort.value
  out = [...out].sort((a, b) => {
    let c = 0
    if (col === 'value') c = (a.value ?? -Infinity) - (b.value ?? -Infinity)
    else if (col === 'source') c = (a.sourceLabel).localeCompare(b.sourceLabel)
    else c = a.metric.localeCompare(b.metric)
    return c * dir
  })
  return out
})

// ── summary counters ──────────────────────────────────────────────
const liveCount = computed(() => rows.value.filter((r) => r.connected && r.value !== undefined).length)
const sourcesUp = computed(() => sources.value.filter((s) => store.status?.outstations?.[s.id]?.connected).length)
const faultCount = computed(() => sources.value.filter((s) => !(store.status?.outstations?.[s.id]?.connected)).length)
const dnp3Count = computed(() => new Set(rows.value.filter((r) => r.protocol === 'dnp3').map((r) => r.source)).size)
const modbusCount = computed(() => new Set(rows.value.filter((r) => r.protocol !== 'dnp3').map((r) => r.source)).size)

// ── value-change flash ────────────────────────────────────────────
const flashing = reactive(new Set<string>())
const timers = new Map<string, number>()
let prev: Record<string, number> = {}

watch(
  () => store.status?.lastReadings,
  (lr) => {
    if (!lr) return
    for (const k in lr) {
      if (prev[k] !== undefined && prev[k] !== lr[k]) {
        flashing.add(k)
        if (timers.has(k)) window.clearTimeout(timers.get(k))
        timers.set(k, window.setTimeout(() => flashing.delete(k), 750))
      }
    }
    prev = { ...lr }
  },
  { deep: true },
)
onUnmounted(() => timers.forEach((t) => window.clearTimeout(t)))

// ── "tick" freshness indicator ────────────────────────────────────
const now = ref(Date.now())
const ticker = window.setInterval(() => (now.value = Date.now()), 1000)
onUnmounted(() => window.clearInterval(ticker))
const lastTick = computed(() => {
  let max = 0
  for (const s of Object.values(store.status?.outstations ?? {})) {
    if (s.lastReadAt) max = Math.max(max, new Date(s.lastReadAt).getTime())
  }
  return max
})
const tickAgo = computed(() => (lastTick.value ? ago(new Date(lastTick.value).toISOString()) : '—'))

// ── helpers ───────────────────────────────────────────────────────
function isBinary(r: Row) {
  return r.ptype === 'binary' || r.ptype === 'binary_output_status'
}
function fmtNum(v: number): string {
  if (Number.isInteger(v)) return v.toString()
  const a = Math.abs(v)
  const dp = a >= 1000 ? 1 : a >= 1 ? 3 : 4
  return v.toFixed(dp)
}
function ledClass(r: Row) {
  if (!r.connected) return 'led--red'
  return r.value !== undefined ? 'led--green' : 'led--dim'
}
// Reactive wrapper over the shared formatter: re-renders each `now` tick and is
// hardened against the Go zero-time / epoch sentinels (→ "—").
const ago = (iso?: string) => agoFmt(iso, now.value)

// SortGlyph — tiny inline indicator. Uses a render function (not a string
// template) so it works in the runtime-only production build.
const SortGlyph = (props: { col: string; sort: { col: string; dir: number } }) =>
  props.sort.col === props.col
    ? h('span', { class: 'inline-block text-bosque ml-0.5' }, props.sort.dir === 1 ? '▲' : '▼')
    : null
</script>

<style scoped>
/* segmented control */
.seg {
  display: inline-flex;
  border: 1.5px solid var(--border);
  border-radius: var(--radius);
  overflow: hidden;
  background: var(--card);
}
.seg-btn {
  font-family: var(--font-sans);
  font-weight: 700;
  font-size: 11px;
  letter-spacing: 0.03em;
  padding: 7px 12px;
  color: var(--muted-foreground);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: background 0.12s, color 0.12s;
}
.seg-btn + .seg-btn { border-left: 1.5px solid var(--border); }
.seg-btn:hover { color: var(--foreground); }
.seg-btn--active {
  background: color-mix(in srgb, var(--epm-bosque) 12%, transparent);
  color: var(--epm-bosque);
}

/* protocol tag */
.proto-tag {
  font-family: var(--font-mono);
  font-size: 9.5px;
  font-weight: 600;
  letter-spacing: 0.06em;
  padding: 2px 6px;
  border-radius: 3px;
  border: 1px solid;
}
.proto-tag--dnp { color: var(--epm-bosque); border-color: color-mix(in srgb, var(--epm-bosque) 40%, transparent); background: color-mix(in srgb, var(--epm-bosque) 8%, transparent); }
.proto-tag--mb  { color: var(--tk-amber-bright); border-color: color-mix(in srgb, var(--tk-amber-base) 45%, transparent); background: color-mix(in srgb, var(--tk-amber-base) 10%, transparent); }
.proto-tag--rtu { color: var(--tk-amber-bright); border-color: color-mix(in srgb, var(--tk-amber-base) 55%, transparent); background: color-mix(in srgb, var(--tk-amber-base) 16%, transparent); letter-spacing: 0.1em; }

/* sticky header sits on the muted band */
thead th { background: var(--muted); }

/* value-update flash — a quick citrico sweep across the row */
.row-flash td {
  animation: row-flash 0.75s ease-out;
}
@keyframes row-flash {
  0%   { background: color-mix(in srgb, var(--epm-citrico) 38%, transparent); }
  100% { background: transparent; }
}
.tabular-nums { font-variant-numeric: tabular-nums; }
</style>
