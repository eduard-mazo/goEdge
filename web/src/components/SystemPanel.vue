<!-- SystemPanel — host telemetry dashboard for the ICR-3232 gateway.
     Reads System/* metrics out of the live readings stream (no extra API),
     renders them as customizable gauge/section widgets, and hosts a tabbed
     metric configurator: each category is a tab with a master toggle plus a
     list of every individual metric to enable/disable. Saving hot-applies to
     the running gateway (no restart). -->
<template>
  <div class="space-y-6 animate-in">

    <!-- ── Config bar ────────────────────────────────────────────── -->
    <section class="forge-panel p-4">
      <div class="flex items-center gap-3 mb-3">
        <Cpu class="h-4 w-4 text-bosque" />
        <h3 class="font-sans font-bold text-xs uppercase tracking-widest text-text-secondary">Monitoreo del host</h3>
        <div class="rule-brand flex-1" />
        <button class="switch" :class="cfg.enabled ? 'switch--on' : ''" role="switch"
                :aria-checked="cfg.enabled" @click="cfg.enabled = !cfg.enabled">
          <span class="switch__track"><span class="switch__knob" /></span>
          <span class="switch__txt" :class="cfg.enabled ? 'text-bosque' : 'text-text-dim'">
            {{ cfg.enabled ? 'ON' : 'OFF' }}
          </span>
        </button>
      </div>

      <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
        <label class="block">
          <span class="form-lbl">Intervalo (ms)</span>
          <input v-model.number="cfg.intervalMs" type="number" min="500" step="500" class="forge-input font-mono text-[12px]" />
        </label>
        <label class="block">
          <span class="form-lbl">Prefijo de métrica</span>
          <input v-model="cfg.metricPrefix" type="text" placeholder="System/" class="forge-input font-mono text-[12px]" />
        </label>
        <label class="block">
          <span class="form-lbl">Montajes (sep. coma)</span>
          <input v-model="mountsStr" type="text" placeholder="/" class="forge-input font-mono text-[12px]" />
        </label>
        <label class="block">
          <span class="form-lbl">Interfaces (vacío = todas)</span>
          <input v-model="ifacesStr" type="text" placeholder="eth0, wwan0" class="forge-input font-mono text-[12px]" />
        </label>
      </div>

      <!-- selection summary + open configurator -->
      <div class="mt-4 flex flex-wrap items-center gap-3">
        <button class="channel-trigger group" @click="openModal">
          <SlidersHorizontal class="h-4 w-4 shrink-0 transition-transform group-hover:rotate-90 duration-300" />
          <span class="font-sans font-bold text-[12px]">Configurar métricas</span>
          <span class="channel-trigger__count">{{ enabledMetricCount }}<span class="opacity-50">/{{ allLeaves.length }}</span></span>
        </button>
        <div class="flex flex-wrap gap-1.5 min-w-0">
          <span v-for="g in groups" :key="g.key" v-show="metrics[g.key]" class="tag-mini">
            {{ g.short }}<span v-if="disabledInGroup(g) > 0" class="opacity-50"> −{{ disabledInGroup(g) }}</span>
          </span>
          <span v-if="enabledMetricCount === 0" class="font-mono text-[10px] text-[color:var(--signal-warn)]">
            ninguna seleccionada — se publican todas
          </span>
        </div>
      </div>

      <div class="flex items-center gap-3 mt-4">
        <button class="btn-primary text-xs py-1.5 px-4" :disabled="saving" @click="save()">
          {{ saving ? 'Guardando…' : 'Guardar' }}
        </button>
        <span v-if="savedMsg" class="font-mono text-[11px]" :class="savedErr ? 'text-[color:var(--signal-fault)]' : 'text-bosque'">{{ savedMsg }}</span>
        <span class="font-mono text-[10px] text-text-dim ml-auto flex items-center gap-1.5">
          <span class="led" :class="store.isRunning ? 'led--green' : 'led--dim'" />
          {{ store.isRunning ? 'se aplica en vivo — sin reinicio' : 'inicia con el gateway' }}
        </span>
      </div>
    </section>

    <!-- ── Empty state ───────────────────────────────────────────── -->
    <div v-if="!hasData" class="forge-panel text-center py-12">
      <p class="font-mono text-xs text-text-dim">
        <template v-if="!cfg.enabled">Activa el monitoreo arriba y guarda.</template>
        <template v-else-if="!store.isRunning">Guardado — inicia el gateway para muestrear.</template>
        <template v-else>Esperando la primera muestra…</template>
      </p>
    </div>

    <template v-else>
      <!-- ── Dashboard customize bar ───────────────────────────────── -->
      <div class="flex items-center gap-2 flex-wrap">
        <span class="font-mono text-[10px] uppercase tracking-[0.16em] text-text-dim mr-1">Ver</span>
        <button v-for="w in widgetDefs" :key="w.key"
                class="vis-pill" :class="show[w.key] ? 'vis-pill--on' : ''"
                @click="show[w.key] = !show[w.key]">
          <component :is="show[w.key] ? Eye : EyeOff" class="h-3 w-3" />
          {{ w.label }}
        </button>
        <button class="vis-pill ml-auto" @click="resetWidgets" title="Restablecer vista">
          <RotateCcw class="h-3 w-3" /> reset
        </button>
      </div>

      <!-- ── Top gauges ──────────────────────────────────────────── -->
      <div v-if="show.gauges" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
        <div v-for="g in topGauges" :key="g.label" class="forge-panel p-4 flex items-center gap-4">
          <div class="ring" :style="g.pct === undefined ? undefined : { background: ringStyle(g.pct) }">
            <div class="ring__hole">
              <span class="font-mono font-bold text-[15px] tabular-nums">{{ g.display }}</span>
              <span class="font-mono text-[9px] text-text-dim">{{ g.unit }}</span>
            </div>
          </div>
          <div class="min-w-0">
            <div class="font-sans font-bold text-xs uppercase tracking-widest text-text-secondary">{{ g.label }}</div>
            <div v-if="g.sub" class="font-mono text-[10px] text-text-dim mt-1 truncate">{{ g.sub }}</div>
          </div>
        </div>
      </div>

      <!-- ── Load averages ───────────────────────────────────────── -->
      <section v-if="show.load && get('CPU/Load1') !== undefined">
        <h4 class="sec-title">Carga promedio</h4>
        <div class="grid grid-cols-3 gap-3">
          <div class="stat-item"><div class="stat-item__label">1 min</div><div class="stat-item__value">{{ fmt(get('CPU/Load1')) }}</div></div>
          <div class="stat-item"><div class="stat-item__label">5 min</div><div class="stat-item__value">{{ fmt(get('CPU/Load5')) }}</div></div>
          <div class="stat-item"><div class="stat-item__label">15 min</div><div class="stat-item__value">{{ fmt(get('CPU/Load15')) }}</div></div>
        </div>
      </section>

      <!-- ── Disks ───────────────────────────────────────────────── -->
      <section v-if="show.disks && disks.length">
        <h4 class="sec-title">Sistemas de archivos</h4>
        <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
          <div v-for="d in disks" :key="d.name" class="forge-panel p-4">
            <div class="flex items-baseline justify-between mb-2">
              <span class="font-mono text-[12px] text-foreground">{{ d.name }}</span>
              <span class="font-mono text-[12px] font-semibold" :class="barColor(d.pct)">{{ fmt(d.pct) }}%</span>
            </div>
            <div class="bar"><div class="bar__fill" :class="barFill(d.pct)" :style="{ width: clampPct(d.pct) + '%' }" /></div>
            <div class="font-mono text-[10px] text-text-dim mt-1.5">{{ fmtMB(d.free) }} libres</div>
          </div>
        </div>
      </section>

      <!-- ── Network ─────────────────────────────────────────────── -->
      <section v-if="show.network && nets.length">
        <h4 class="sec-title">Red</h4>
        <div class="forge-panel overflow-x-auto">
          <table class="w-full">
            <thead><tr>
              <th>Interfaz</th>
              <th class="text-right">Tasa Rx</th>
              <th class="text-right">Tasa Tx</th>
              <th class="text-right">Rx total</th>
              <th class="text-right">Tx total</th>
            </tr></thead>
            <tbody>
              <tr v-for="n in nets" :key="n.name">
                <td class="font-mono text-[12px] text-text-secondary">{{ n.name }}</td>
                <td class="font-mono text-right text-bosque">{{ fmtKbps(n.rx) }}</td>
                <td class="font-mono text-right text-bosque">{{ fmtKbps(n.tx) }}</td>
                <td class="font-mono text-right text-text-dim">{{ fmtMB(n.rxTotal) }}</td>
                <td class="font-mono text-right text-text-dim">{{ fmtMB(n.txTotal) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- ── Raw readings ────────────────────────────────────────── -->
      <section v-if="show.raw">
        <h4 class="sec-title">Todas las métricas publicadas</h4>
        <div class="forge-panel overflow-x-auto">
          <table class="w-full">
            <thead><tr><th>Métrica</th><th class="text-right">Valor</th></tr></thead>
            <tbody>
              <tr v-for="r in rawRows" :key="r.name">
                <td class="font-mono text-[11.5px] text-text-secondary">{{ prefix }}{{ r.name }}</td>
                <td class="font-mono text-right text-foreground tabular-nums">{{ fmt(r.value) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </template>

    <!-- ── Tabbed metric configurator ────────────────────────────── -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="modalOpen" class="modal-scrim">
          <div class="modal-rack" role="dialog" aria-modal="true" aria-label="Configure metrics">
            <!-- header -->
            <header class="modal-rack__head">
              <div class="flex items-center gap-2.5 min-w-0">
                <span class="led" :class="draftEnabledCount ? 'led--green' : 'led--amber'" />
                <div class="min-w-0">
                  <h3 class="font-sans font-extrabold text-[15px] leading-none truncate">Configurador de métricas</h3>
                  <p class="font-mono text-[10px] text-text-dim mt-1">activa categorías y ajusta métricas individuales</p>
                </div>
              </div>
              <button class="rack-x ml-auto" @click="closeModal" aria-label="Close"><X class="h-4 w-4" /></button>
            </header>

            <!-- tabs + panel -->
            <div class="modal-rack__main">
              <!-- vertical category tabs -->
              <nav class="rack-tabs">
                <button v-for="g in groups" :key="g.key"
                        class="rack-tab" :class="activeTab === g.key ? 'rack-tab--active' : ''"
                        @click="activeTab = g.key">
                  <component :is="g.icon" class="h-4 w-4 shrink-0" />
                  <span class="rack-tab__label">{{ g.label }}</span>
                  <span class="rack-tab__dot" :class="tabDotClass(g)" />
                </button>
              </nav>

              <!-- active category panel -->
              <div class="rack-panel" :key="activeTab">
                <div class="rack-panel__head">
                  <div class="min-w-0">
                    <div class="font-sans font-extrabold text-[14px] flex items-center gap-2">
                      <component :is="activeGroup.icon" class="h-4 w-4 text-bosque" />
                      {{ activeGroup.label }}
                    </div>
                    <p class="font-mono text-[10px] text-text-dim mt-1">{{ activeGroup.blurb }}</p>
                  </div>
                  <button class="switch ml-auto shrink-0" :class="draft.metrics[activeTab] ? 'switch--on' : ''"
                          role="switch" :aria-checked="draft.metrics[activeTab]"
                          @click="draft.metrics[activeTab] = !draft.metrics[activeTab]">
                    <span class="switch__track"><span class="switch__knob" /></span>
                    <span class="switch__txt" :class="draft.metrics[activeTab] ? 'text-bosque' : 'text-text-dim'">
                      {{ draft.metrics[activeTab] ? 'ON' : 'OFF' }}
                    </span>
                  </button>
                </div>

                <!-- per-metric list -->
                <div class="rack-list" :class="draft.metrics[activeTab] ? '' : 'rack-list--muted'">
                  <div class="flex items-center gap-2 mb-2">
                    <span class="font-mono text-[9px] uppercase tracking-[0.18em] text-text-dim">Métricas individuales</span>
                    <div class="hairline flex-1" />
                    <button class="rack-link" @click="setAllLeaves(activeGroup, true)">activar todas</button>
                    <span class="text-text-dim">·</span>
                    <button class="rack-link" @click="setAllLeaves(activeGroup, false)">desactivar todas</button>
                  </div>

                  <button
                    v-for="(leaf, li) in activeGroup.leaves" :key="leaf.suffix"
                    class="metric-row" :class="leafOn(leaf) ? 'metric-row--on' : ''"
                    :style="{ animationDelay: (li * 30) + 'ms' }"
                    role="switch" :aria-checked="leafOn(leaf)"
                    :disabled="!draft.metrics[activeTab]"
                    @click="toggleLeaf(leaf)"
                  >
                    <span class="switch__track shrink-0"><span class="switch__knob" /></span>
                    <span class="min-w-0 flex-1 text-left">
                      <code class="block font-mono text-[12px] leading-tight truncate">{{ prefix }}{{ leaf.suffix }}</code>
                      <span class="block font-mono text-[10px] text-text-dim leading-tight mt-0.5 truncate">{{ leaf.desc }}</span>
                    </span>
                    <span class="metric-row__sig" :class="leafOn(leaf) ? 'metric-row__sig--live' : ''">
                      {{ leafOn(leaf) ? 'ON' : 'OFF' }}
                    </span>
                  </button>

                  <p v-if="!draft.metrics[activeTab]" class="font-mono text-[10px] text-text-dim mt-2">
                    Categoría apagada — actívala para publicar estas métricas.
                  </p>
                </div>
              </div>
            </div>

            <!-- footer -->
            <footer class="modal-rack__foot">
              <span class="font-mono text-[10px] text-text-dim">
                <span class="text-foreground font-semibold">{{ draftEnabledCount }}</span> / {{ allLeaves.length }} métricas activas
              </span>
              <div class="ml-auto flex items-center gap-2">
                <button class="btn-ghost text-xs py-1.5 px-4" @click="closeModal">Cancelar</button>
                <button class="btn-primary text-xs py-1.5 px-5" :disabled="saving" @click="applyModal">
                  {{ saving ? 'Aplicando…' : 'Aplicar' }}
                </button>
              </div>
            </footer>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted, onUnmounted } from 'vue'
import { Cpu, X, SlidersHorizontal, MemoryStick, HardDrive, Network, Thermometer, Activity, Eye, EyeOff, RotateCcw } from 'lucide-vue-next'
import { useGatewayStore } from '@/stores/gateway'
import { api, type SystemConfig, type SystemMetrics } from '@/api/client'

const store = useGatewayStore()

// ── metric catalog ────────────────────────────────────────────────
// Each group maps to a SystemMetrics flag (category master) and lists the
// individual metric suffixes it emits. Disk/network use <mount>/<iface>
// placeholders that the backend matches against every concrete instance.
type MetricKey = keyof SystemMetrics
interface Leaf { suffix: string; desc: string }
interface Group { key: MetricKey; label: string; short: string; icon: unknown; blurb: string; leaves: Leaf[] }

const groups: Group[] = [
  { key: 'cpu', label: 'CPU', short: 'CPU', icon: Cpu, blurb: 'Uso del procesador.', leaves: [
    { suffix: 'CPU/Usage_pct', desc: 'uso total %' },
  ] },
  { key: 'load', label: 'Carga', short: 'Carga', icon: Activity, blurb: 'Carga promedio del kernel.', leaves: [
    { suffix: 'CPU/Load1',  desc: 'carga 1 minuto' },
    { suffix: 'CPU/Load5',  desc: 'carga 5 minutos' },
    { suffix: 'CPU/Load15', desc: 'carga 15 minutos' },
  ] },
  { key: 'memory', label: 'Memoria', short: 'Mem', icon: MemoryStick, blurb: 'Uso y capacidad de RAM.', leaves: [
    { suffix: 'Memory/Used_pct',      desc: 'porcentaje usado' },
    { suffix: 'Memory/Used_MB',       desc: 'megabytes usados' },
    { suffix: 'Memory/Available_MB',  desc: 'megabytes disponibles' },
    { suffix: 'Memory/Total_MB',      desc: 'megabytes totales' },
  ] },
  { key: 'swap', label: 'Swap', short: 'Swap', icon: MemoryStick, blurb: 'Uso del espacio de swap.', leaves: [
    { suffix: 'Memory/Swap_Used_pct', desc: 'porcentaje de swap usado' },
  ] },
  { key: 'disk', label: 'Disco', short: 'Disco', icon: HardDrive, blurb: 'Uso por montaje (cada montaje configurado).', leaves: [
    { suffix: 'Disk/<mount>/Used_pct', desc: 'porcentaje usado por montaje' },
    { suffix: 'Disk/<mount>/Free_MB',  desc: 'megabytes libres por montaje' },
  ] },
  { key: 'network', label: 'Red totales', short: 'Red', icon: Network, blurb: 'Contadores de tráfico por interfaz.', leaves: [
    { suffix: 'Network/<iface>/Rx_MB', desc: 'megabytes recibidos por interfaz' },
    { suffix: 'Network/<iface>/Tx_MB', desc: 'megabytes enviados por interfaz' },
  ] },
  { key: 'networkRates', label: 'Red tasas', short: 'Tasa', icon: Network, blurb: 'Caudal instantáneo por interfaz.', leaves: [
    { suffix: 'Network/<iface>/RxRate_kbps', desc: 'tasa de recepción (kbps)' },
    { suffix: 'Network/<iface>/TxRate_kbps', desc: 'tasa de envío (kbps)' },
  ] },
  { key: 'temperature', label: 'Temperatura', short: 'Temp', icon: Thermometer, blurb: 'Sensor térmico CPU / placa.', leaves: [
    { suffix: 'Temperature/CPU_C', desc: 'temperatura CPU °C' },
  ] },
  { key: 'uptime', label: 'Uptime', short: 'Uptime', icon: Activity, blurb: 'Tiempo desde el arranque.', leaves: [
    { suffix: 'Uptime_h', desc: 'horas desde el arranque' },
  ] },
  { key: 'processes', label: 'Procesos', short: 'Proc', icon: Activity, blurb: 'Número de procesos.', leaves: [
    { suffix: 'Process/Count', desc: 'número de procesos' },
  ] },
]
const allLeaves = groups.flatMap((g) => g.leaves)

function allMetrics(v: boolean): Required<SystemMetrics> {
  return { cpu: v, load: v, memory: v, swap: v, disk: v, network: v, networkRates: v, temperature: v, uptime: v, processes: v }
}

// ── config state (committed) ──────────────────────────────────────
const cfg = reactive<SystemConfig>({ enabled: false, intervalMs: 5000, metricPrefix: 'System/', mounts: ['/'], interfaces: [] })
const metrics = reactive<Required<SystemMetrics>>(allMetrics(true))
const disabledMetrics = reactive(new Set<string>())   // individual opt-outs (canonical suffixes)
const mountsStr = ref('/')
const ifacesStr = ref('')
const saving = ref(false)
const savedMsg = ref('')
const savedErr = ref(false)

// a metric (group on AND not individually disabled)
function isLeafEnabled(g: Group, leaf: Leaf): boolean { return metrics[g.key] && !disabledMetrics.has(leaf.suffix) }
const enabledMetricCount = computed(() => groups.reduce((n, g) => n + g.leaves.filter((l) => isLeafEnabled(g, l)).length, 0))
function disabledInGroup(g: Group): number { return metrics[g.key] ? g.leaves.filter((l) => disabledMetrics.has(l.suffix)).length : 0 }

onMounted(async () => {
  try {
    const c = await api.getSystem()
    Object.assign(cfg, c)
    mountsStr.value = (c.mounts ?? ['/']).join(', ')
    ifacesStr.value = (c.interfaces ?? []).join(', ')
    if (c.metrics && Object.values(c.metrics).some(Boolean)) {
      Object.assign(metrics, allMetrics(false), c.metrics)
    }
    disabledMetrics.clear()
    for (const s of c.disabledMetrics ?? []) disabledMetrics.add(s)
  } catch { /* keep defaults */ }
})

// ── modal (tabbed; draft applied on confirm) ──────────────────────
const modalOpen = ref(false)
const activeTab = ref<MetricKey>('cpu')
const draft = reactive<{ metrics: Required<SystemMetrics>; disabled: Set<string> }>({
  metrics: allMetrics(true),
  disabled: new Set<string>(),
})

const activeGroup = computed<Group>(() => groups.find((g) => g.key === activeTab.value) ?? groups[0])
const draftEnabledCount = computed(() =>
  groups.reduce((n, g) => n + (draft.metrics[g.key] ? g.leaves.filter((l) => !draft.disabled.has(l.suffix)).length : 0), 0),
)

function leafOn(leaf: Leaf): boolean { return draft.metrics[activeTab.value] && !draft.disabled.has(leaf.suffix) }
function toggleLeaf(leaf: Leaf) {
  if (!draft.metrics[activeTab.value]) return
  if (draft.disabled.has(leaf.suffix)) draft.disabled.delete(leaf.suffix)
  else draft.disabled.add(leaf.suffix)
}
function setAllLeaves(g: Group, on: boolean) {
  if (!draft.metrics[g.key] && on) draft.metrics[g.key] = true
  for (const l of g.leaves) { if (on) draft.disabled.delete(l.suffix); else draft.disabled.add(l.suffix) }
}
function tabDotClass(g: Group): string {
  if (!draft.metrics[g.key]) return 'dot--off'
  const anyDisabled = g.leaves.some((l) => draft.disabled.has(l.suffix))
  return anyDisabled ? 'dot--partial' : 'dot--on'
}

function openModal() {
  Object.assign(draft.metrics, metrics)
  draft.disabled = new Set(disabledMetrics)
  modalOpen.value = true
  document.body.style.overflow = 'hidden'
}
function closeModal() {
  modalOpen.value = false
  document.body.style.overflow = ''
}
async function applyModal() {
  Object.assign(metrics, draft.metrics)
  disabledMetrics.clear()
  for (const s of draft.disabled) disabledMetrics.add(s)
  await save('✓ métricas aplicadas en vivo')
  if (!savedErr.value) closeModal()
}

function onKey(e: KeyboardEvent) { if (e.key === 'Escape' && modalOpen.value) closeModal() }
onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => { window.removeEventListener('keydown', onKey); document.body.style.overflow = '' })

// ── save (hot-applies to the running gateway) ─────────────────────
async function save(okMsg = '✓ guardado — aplicado en vivo') {
  saving.value = true
  savedMsg.value = ''
  savedErr.value = false
  try {
    cfg.mounts = splitList(mountsStr.value)
    cfg.interfaces = splitList(ifacesStr.value)
    cfg.metrics = { ...metrics }
    cfg.disabledMetrics = [...disabledMetrics]
    const c = await api.setSystem({ ...cfg })
    Object.assign(cfg, c)
    savedMsg.value = store.isRunning ? okMsg : '✓ guardado — inicia con el gateway'
    setTimeout(() => (savedMsg.value = ''), 4000)
  } catch (e: unknown) {
    savedErr.value = true
    savedMsg.value = 'error: ' + (e instanceof Error ? e.message : String(e))
  } finally {
    saving.value = false
  }
}

function splitList(s: string): string[] {
  return s.split(',').map((x) => x.trim()).filter(Boolean)
}

// ── dashboard widget visibility (persisted) ───────────────────────
type WidgetKey = 'gauges' | 'load' | 'disks' | 'network' | 'raw'
const widgetDefs: { key: WidgetKey; label: string }[] = [
  { key: 'gauges',  label: 'Medidores' },
  { key: 'load',    label: 'Carga' },
  { key: 'disks',   label: 'Discos' },
  { key: 'network', label: 'Red' },
  { key: 'raw',     label: 'Lista' },
]
const WIDGET_LS = 'sysmon:widgets'
const defaultWidgets = (): Record<WidgetKey, boolean> => ({ gauges: true, load: true, disks: true, network: true, raw: false })
const show = reactive<Record<WidgetKey, boolean>>(loadWidgets())
function loadWidgets(): Record<WidgetKey, boolean> {
  try { return { ...defaultWidgets(), ...JSON.parse(localStorage.getItem(WIDGET_LS) || '{}') } }
  catch { return defaultWidgets() }
}
function resetWidgets() { Object.assign(show, defaultWidgets()) }
// persist widget visibility on any change
watch(show, (v) => localStorage.setItem(WIDGET_LS, JSON.stringify(v)), { deep: true })

// ── live readings, filtered to the System prefix ──────────────────
const prefix = computed(() => cfg.metricPrefix || 'System/')

const sys = computed<Record<string, number>>(() => {
  const out: Record<string, number> = {}
  const lr = store.status?.lastReadings ?? {}
  const p = prefix.value
  for (const k in lr) {
    if (k.startsWith(p)) out[k.slice(p.length)] = lr[k]
  }
  return out
})

const hasData = computed(() => Object.keys(sys.value).length > 0)
function get(suffix: string): number | undefined { return sys.value[suffix] }

const rawRows = computed(() =>
  Object.entries(sys.value).map(([name, value]) => ({ name, value })).sort((a, b) => a.name.localeCompare(b.name)),
)

const memSub = computed(() => {
  const used = get('Memory/Used_MB'); const total = get('Memory/Total_MB')
  if (used === undefined || total === undefined) return undefined
  return `${fmtMB(used)} / ${fmtMB(total)}`
})

const tempPct = computed(() => {
  const t = get('Temperature/CPU_C')
  return t === undefined ? undefined : Math.min(100, (t / 90) * 100)
})

const disks = computed(() => {
  const names = new Set<string>()
  for (const k in sys.value) {
    const m = /^Disk\/(.+)\/Used_pct$/.exec(k)
    if (m) names.add(m[1])
  }
  return [...names].sort().map((name) => ({
    name,
    pct: get(`Disk/${name}/Used_pct`) ?? 0,
    free: get(`Disk/${name}/Free_MB`) ?? 0,
  }))
})

const nets = computed(() => {
  const names = new Set<string>()
  for (const k in sys.value) {
    const m = /^Network\/(.+)\/(Rx_MB|Tx_MB|RxRate_kbps|TxRate_kbps)$/.exec(k)
    if (m) names.add(m[1])
  }
  return [...names].sort().map((name) => ({
    name,
    rx: get(`Network/${name}/RxRate_kbps`),
    tx: get(`Network/${name}/TxRate_kbps`),
    rxTotal: get(`Network/${name}/Rx_MB`) ?? 0,
    txTotal: get(`Network/${name}/Tx_MB`) ?? 0,
  }))
})

// ── formatting / colors ───────────────────────────────────────────
function fmt(v?: number): string { return v === undefined ? '—' : (Number.isInteger(v) ? v.toString() : v.toFixed(2)) }
function fmtMB(mb?: number): string {
  if (mb === undefined) return '—'
  return mb >= 1024 ? (mb / 1024).toFixed(2) + ' GB' : mb.toFixed(0) + ' MB'
}
function fmtKbps(k?: number): string {
  if (k === undefined) return '—'
  return k >= 1000 ? (k / 1000).toFixed(2) + ' Mbps' : k.toFixed(1) + ' kbps'
}
function clampPct(v: number): number { return Math.max(0, Math.min(100, v)) }
function barColor(v: number): string { return v >= 90 ? 'text-[color:var(--signal-fault)]' : v >= 75 ? 'text-[color:var(--signal-warn)]' : 'text-bosque' }
function barFill(v: number): string { return v >= 90 ? 'bar__fill--err' : v >= 75 ? 'bar__fill--warn' : '' }

function ringStyle(pct: number): string {
  const col = pct >= 90 ? 'var(--signal-fault)' : pct >= 75 ? 'var(--signal-warn)' : 'var(--epm-bosque)'
  return `conic-gradient(${col} ${clampPct(pct) * 3.6}deg, color-mix(in srgb, var(--border) 60%, transparent) 0deg)`
}

interface Gauge { label: string; display: string; unit: string; pct?: number; sub?: string }
const topGauges = computed<Gauge[]>(() => {
  const cpu = get('CPU/Usage_pct')
  const memPct = get('Memory/Used_pct')
  const temp = get('Temperature/CPU_C')
  const up = get('Uptime_h')
  const procs = get('Process/Count')
  return [
    { label: 'CPU', display: fmt1(cpu), unit: '%', pct: cpu },
    { label: 'Memoria', display: fmt1(memPct), unit: '%', pct: memPct, sub: memSub.value },
    { label: 'Temperatura', display: fmt1(temp), unit: '°C', pct: tempPct.value, sub: temp === undefined ? 'n/d' : undefined },
    { label: 'Uptime', display: fmt1(up), unit: 'h', pct: undefined, sub: procs !== undefined ? `${procs} proc` : undefined },
  ]
})
function fmt1(v?: number): string {
  if (v === undefined) return '—'
  return Number.isInteger(v) ? v.toString() : v.toFixed(1)
}
</script>

<style scoped>
.sec-title { font-family: var(--font-sans); font-weight: 700; font-size: 11px; text-transform: uppercase; letter-spacing: 0.12em; color: var(--text-secondary, var(--muted-foreground)); margin-bottom: 0.75rem; }
.form-lbl { display: block; font-family: var(--font-mono); font-size: 10px; text-transform: uppercase; letter-spacing: 0.08em; color: var(--muted-foreground); margin-bottom: 4px; }

/* ── industrial rocker switch ─────────────────────────────────── */
.switch { display: inline-flex; align-items: center; gap: 8px; background: none; border: none; cursor: pointer; padding: 0; }
.switch__track {
  position: relative; width: 38px; height: 20px; border-radius: 999px;
  background: color-mix(in srgb, var(--border) 80%, transparent);
  border: 1.5px solid var(--border);
  transition: background 0.22s ease, border-color 0.22s ease;
}
.switch__knob {
  position: absolute; top: 1px; left: 1px; width: 14px; height: 14px; border-radius: 50%;
  background: var(--card); box-shadow: var(--shadow-sm);
  transition: transform 0.22s cubic-bezier(0.3, 1.4, 0.5, 1);
}
.switch__txt { font-family: var(--font-mono); font-size: 10px; font-weight: 600; letter-spacing: 0.05em; }
.switch--on .switch__track,
.metric-row--on .switch__track {
  background: var(--epm-bosque); border-color: var(--epm-bosque);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--epm-bosque) 18%, transparent);
}
.switch--on .switch__knob,
.metric-row--on .switch__knob { transform: translateX(18px); }

/* ── trigger + summary tags ───────────────────────────────────── */
.channel-trigger {
  display: inline-flex; align-items: center; gap: 8px;
  padding: 8px 14px; border-radius: var(--radius);
  border: 1.5px solid color-mix(in srgb, var(--epm-bosque) 35%, transparent);
  background: color-mix(in srgb, var(--epm-bosque) 7%, transparent);
  color: var(--epm-bosque); cursor: pointer;
  transition: background 0.15s, box-shadow 0.15s;
}
.channel-trigger:hover { background: color-mix(in srgb, var(--epm-bosque) 12%, transparent); box-shadow: var(--shadow-sm); }
.channel-trigger__count {
  font-family: var(--font-mono); font-size: 11px; font-weight: 700;
  padding: 1px 7px; border-radius: 999px; background: var(--epm-bosque); color: var(--primary-foreground);
}
.tag-mini {
  font-family: var(--font-mono); font-size: 10px; padding: 2px 7px; border-radius: 999px;
  border: 1px solid var(--border); color: var(--muted-foreground); background: var(--card); white-space: nowrap;
}

/* ── dashboard visibility pills ───────────────────────────────── */
.vis-pill {
  display: inline-flex; align-items: center; gap: 5px;
  font-family: var(--font-mono); font-size: 10.5px;
  padding: 4px 10px; border-radius: 999px;
  border: 1.5px solid var(--border); background: var(--card);
  color: var(--muted-foreground); cursor: pointer;
  transition: color 0.12s, border-color 0.12s, background 0.12s;
}
.vis-pill:hover { color: var(--foreground); }
.vis-pill--on {
  color: var(--epm-bosque); border-color: color-mix(in srgb, var(--epm-bosque) 40%, transparent);
  background: color-mix(in srgb, var(--epm-bosque) 8%, transparent);
}

/* ── modal scrim + rack ───────────────────────────────────────── */
.modal-scrim {
  position: fixed; inset: 0; z-index: 100;
  display: flex; align-items: center; justify-content: center; padding: 1rem;
  background: color-mix(in srgb, var(--background) 55%, #000 45%);
  backdrop-filter: blur(6px) saturate(0.9);
}
.modal-rack {
  width: min(760px, 100%); max-height: min(88vh, 720px);
  display: flex; flex-direction: column;
  background: var(--card); color: var(--foreground);
  border: 1.5px solid var(--tk-border-bright);
  border-radius: calc(var(--radius) + 6px);
  box-shadow: var(--shadow-lg), 0 0 0 1px color-mix(in srgb, var(--epm-bosque) 10%, transparent);
  overflow: hidden;
}
.modal-rack__head {
  display: flex; align-items: center; gap: 10px; padding: 14px 16px;
  border-bottom: 1.5px solid var(--border);
  background: linear-gradient(180deg, color-mix(in srgb, var(--epm-bosque) 8%, var(--card)), var(--card));
}
.modal-rack__main { display: flex; flex: 1; min-height: 0; }
.modal-rack__foot {
  display: flex; align-items: center; gap: 10px; padding: 12px 16px;
  border-top: 1.5px solid var(--border); background: var(--card);
}
.rack-x {
  display: grid; place-items: center; width: 28px; height: 28px;
  border-radius: var(--radius); border: 1px solid var(--border);
  background: var(--card); color: var(--muted-foreground); cursor: pointer;
  transition: color 0.12s, border-color 0.12s;
}
.rack-x:hover { color: var(--signal-fault); border-color: color-mix(in srgb, var(--signal-fault) 45%, transparent); }
.rack-link {
  font-family: var(--font-mono); font-size: 10px; text-transform: uppercase; letter-spacing: 0.06em;
  color: var(--epm-bosque); background: none; border: none; cursor: pointer; padding: 0;
}
.rack-link:hover { text-decoration: underline; }
.hairline { height: 1px; background: color-mix(in srgb, var(--border) 70%, transparent); }

/* ── vertical tabs ────────────────────────────────────────────── */
.rack-tabs {
  width: 180px; shrink: 0; flex-shrink: 0; overflow-y: auto;
  border-right: 1.5px solid var(--border); background: var(--muted);
  padding: 8px; display: flex; flex-direction: column; gap: 2px;
}
@media (max-width: 560px) { .rack-tabs { width: 128px; } }
.rack-tab {
  display: flex; align-items: center; gap: 9px;
  padding: 9px 10px; border-radius: var(--radius);
  font-family: var(--font-sans); font-size: 12.5px; font-weight: 600;
  color: var(--muted-foreground); background: none; border: none; cursor: pointer;
  text-align: left; transition: background 0.12s, color 0.12s; position: relative;
}
.rack-tab:hover { background: color-mix(in srgb, var(--epm-bosque) 8%, transparent); color: var(--foreground); }
.rack-tab--active { background: var(--card); color: var(--foreground); box-shadow: var(--shadow-sm); }
.rack-tab--active::before {
  content: ''; position: absolute; left: 0; top: 6px; bottom: 6px; width: 3px;
  border-radius: 0 3px 3px 0; background: var(--epm-bosque);
}
.rack-tab__label { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.rack-tab__dot { width: 8px; height: 8px; border-radius: 50%; shrink: 0; flex-shrink: 0; }
.dot--on      { background: var(--epm-bosque); box-shadow: 0 0 0 3px color-mix(in srgb, var(--epm-bosque) 18%, transparent); }
.dot--partial { background: var(--signal-warn); box-shadow: 0 0 0 3px color-mix(in srgb, var(--signal-warn) 18%, transparent); }
.dot--off     { background: var(--tk-border-bright); }

/* ── active panel + per-metric list ───────────────────────────── */
.rack-panel { flex: 1; min-width: 0; overflow-y: auto; padding: 14px 16px; animation: panel-in 0.22s ease-out; }
.rack-panel__head { display: flex; align-items: flex-start; gap: 12px; padding-bottom: 12px; border-bottom: 1px solid color-mix(in srgb, var(--border) 60%, transparent); }
.rack-list { margin-top: 12px; display: flex; flex-direction: column; gap: 7px; transition: opacity 0.18s; }
.rack-list--muted { opacity: 0.5; }
.metric-row {
  display: flex; align-items: center; gap: 10px; padding: 9px 11px;
  border-radius: var(--radius); border: 1.5px solid var(--border); background: var(--card);
  cursor: pointer; text-align: left; animation: channel-in 0.28s ease-out both;
  transition: border-color 0.15s, background 0.15s, box-shadow 0.15s, transform 0.08s;
}
.metric-row:hover:not(:disabled) { border-color: var(--tk-border-bright); box-shadow: var(--shadow-sm); }
.metric-row:active:not(:disabled) { transform: scale(0.99); }
.metric-row:disabled { cursor: not-allowed; }
.metric-row--on { border-color: color-mix(in srgb, var(--epm-bosque) 45%, transparent); background: color-mix(in srgb, var(--epm-bosque) 6%, transparent); }
.metric-row code { color: var(--foreground); }
.metric-row__sig {
  font-family: var(--font-mono); font-size: 9px; font-weight: 700; letter-spacing: 0.08em;
  padding: 2px 6px; border-radius: 4px; shrink: 0; flex-shrink: 0;
  color: var(--muted-foreground); border: 1px solid var(--border); background: var(--muted);
}
.metric-row__sig--live {
  color: var(--epm-bosque); border-color: color-mix(in srgb, var(--epm-bosque) 40%, transparent);
  background: color-mix(in srgb, var(--epm-bosque) 12%, transparent);
}

/* ── dashboard widgets ────────────────────────────────────────── */
.bar { height: 8px; border-radius: 4px; background: color-mix(in srgb, var(--border) 60%, transparent); overflow: hidden; }
.bar__fill { height: 100%; border-radius: 4px; background: var(--epm-bosque); transition: width 0.4s ease; }
.bar__fill--warn { background: var(--signal-warn); }
.bar__fill--err  { background: var(--signal-fault); }
.ring { width: 64px; height: 64px; border-radius: 50%; display: grid; place-items: center; flex-shrink: 0; background: color-mix(in srgb, var(--border) 60%, transparent); }
.ring__hole { width: 50px; height: 50px; border-radius: 50%; background: var(--card); display: flex; flex-direction: column; align-items: center; justify-content: center; line-height: 1; gap: 1px; }
.tabular-nums { font-variant-numeric: tabular-nums; }

/* ── transitions ──────────────────────────────────────────────── */
@keyframes channel-in { from { opacity: 0; transform: translateY(5px); } to { opacity: 1; transform: translateY(0); } }
@keyframes panel-in { from { opacity: 0; transform: translateX(6px); } to { opacity: 1; transform: translateX(0); } }
.modal-enter-active, .modal-leave-active { transition: opacity 0.2s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-active .modal-rack { transition: transform 0.28s cubic-bezier(0.2, 1, 0.4, 1), opacity 0.2s ease; }
.modal-enter-from .modal-rack { transform: translateY(14px) scale(0.97); opacity: 0; }
</style>
