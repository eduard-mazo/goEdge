<template>
  <div class="h-screen w-screen overflow-hidden flex bg-background text-foreground" :class="dark ? 'dark' : ''">

    <!-- ── SIDEBAR ────────────────────────────────────────────────── -->
    <aside
      class="bg-sidebar text-sidebar-foreground border-r border-sidebar-border flex flex-col h-screen overflow-hidden transition-[width,transform] duration-300 ease-out shrink-0"
      :class="collapsed ? 'w-[64px] sidebar-mini' : 'w-[240px]'"
    >
      <!-- Brand -->
      <div class="flex items-center gap-3 px-4 h-16 border-b border-sidebar-border shrink-0">
        <div class="grid place-items-center w-8 h-8 rounded-sm bg-[color:var(--epm-citrico)] text-[color:var(--epm-bosque)] font-black text-sm shrink-0 leading-none">
          M
        </div>
        <div class="sidebar-wide-only leading-none flex-1 min-w-0">
          <div class="font-sans text-base font-extrabold tracking-tight text-white truncate">DNP3 GW</div>
          <div class="text-[10px] uppercase tracking-[0.18em] text-[color:var(--epm-citrico)] mt-0.5 font-semibold">
            Sparkplug B
          </div>
        </div>
      </div>

      <!-- Connection section -->
      <div class="px-3 pt-4 pb-2 sidebar-wide-only border-b border-sidebar-border shrink-0">
        <div class="text-[10px] uppercase tracking-[0.2em] text-[color:var(--epm-citrico)] font-semibold mb-2 px-1">Status</div>

        <!-- MQTT row -->
        <div class="flex items-center gap-2 px-2 py-1.5 rounded-sm"
             :class="store.mqttConnected ? 'bg-sidebar-accent' : ''">
          <div :class="['led', store.mqttConnected ? 'led--green' : 'led--dim']" />
          <span class="font-mono text-[11px] truncate"
                :class="store.mqttConnected ? 'text-white' : 'text-white/50'">
            {{ store.mqttConnected ? 'MQTT online' : 'MQTT offline' }}
          </span>
        </div>

        <!-- Gateway run/stop -->
        <button
          v-if="!store.isRunning"
          class="mt-2 w-full btn-primary text-xs py-1.5"
          @click="doStart"
        >▶ Start Gateway</button>
        <button
          v-else
          class="mt-2 w-full btn-danger text-xs py-1.5"
          @click="doStop"
        >■ Stop Gateway</button>

        <button
          class="mt-2 w-full text-left px-2 py-1.5 rounded-sm text-xs font-semibold transition-colors"
          :class="showConfig
            ? 'bg-[color:var(--epm-citrico)] text-[color:var(--epm-bosque)]'
            : 'text-white/70 hover:bg-sidebar-accent'"
          @click="showConfig = !showConfig">
          {{ showConfig ? 'Hide settings' : 'Configure…' }}
        </button>
      </div>

      <!-- Nav -->
      <nav class="flex-1 min-h-0 overflow-y-auto py-3 px-2 space-y-0.5 sidebar-nav-scroll">
        <div class="sidebar-wide-only text-[10px] uppercase tracking-[0.2em] text-[color:var(--epm-citrico)] font-semibold mb-2 px-2">Panels</div>
        <button
          v-for="tab in tabs"
          :key="tab.id"
          class="group w-full flex items-center gap-3 rounded-sm px-3 py-2.5 text-sm transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground relative"
          :class="activeTab === tab.id
            ? 'bg-sidebar-accent text-sidebar-accent-foreground'
            : 'text-sidebar-foreground/70'"
          :title="tab.label"
          @click="activeTab = tab.id">
          <span v-if="activeTab === tab.id"
                class="absolute left-0 top-1.5 bottom-1.5 w-0.5 bg-[color:var(--sidebar-primary)] rounded-r" />
          <component :is="tab.icon" class="h-4 w-4 shrink-0" />
          <span class="sidebar-label truncate font-semibold">{{ tab.label }}</span>
        </button>
      </nav>

      <!-- Bottom controls -->
      <div class="border-t border-sidebar-border p-2 space-y-0.5 shrink-0">
        <!-- Running indicator -->
        <div v-if="store.isRunning"
             class="sidebar-wide-only flex items-center gap-2 px-3 py-1.5 mb-1 rounded-sm bg-[color:color-mix(in_srgb,var(--signal-warn)_15%,transparent)] border border-[color:var(--signal-warn)]">
          <div class="led led--amber" />
          <span class="text-[11px] font-semibold text-[color:var(--signal-warn)] font-mono truncate">
            pub {{ store.status?.publishCount ?? 0 }}
          </span>
        </div>

        <button
          class="w-full flex items-center gap-3 rounded-sm px-3 py-2 text-sm hover:bg-sidebar-accent transition-colors text-sidebar-foreground/70"
          :title="dark ? 'Light mode' : 'Dark mode'"
          @click="toggleTheme">
          <svg v-if="dark" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="5"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>
          <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12.79A9 9 0 1111.21 3a7 7 0 009.79 9.79z"/></svg>
          <span class="sidebar-label">{{ dark ? 'Light mode' : 'Dark mode' }}</span>
        </button>
        <button
          class="hidden md:flex w-full items-center gap-3 rounded-sm px-3 py-2 text-sm hover:bg-sidebar-accent transition-colors text-sidebar-foreground/70"
          :title="collapsed ? 'Expand sidebar' : 'Collapse sidebar'"
          @click="toggleSidebar">
          <svg v-if="collapsed" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M9 3v18M15 9l3 3-3 3"/></svg>
          <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M9 3v18M9 9l-3 3 3 3"/></svg>
          <span class="sidebar-label">{{ collapsed ? 'Expand' : 'Collapse' }}</span>
        </button>
      </div>
    </aside>

    <!-- ── MAIN COLUMN ────────────────────────────────────────────── -->
    <div class="flex-1 flex flex-col min-w-0 h-screen overflow-hidden">

      <!-- Top rail -->
      <header class="shrink-0 flex items-center gap-3 h-14 px-4 sm:px-6 border-b border-border bg-card z-30"
              style="box-shadow: var(--shadow-sm)">
        <div class="flex items-center gap-2 min-w-0 flex-1">
          <h2 class="font-sans font-extrabold text-base leading-none truncate">{{ activeTabLabel }}</h2>
          <!-- running pill -->
          <span
            v-if="store.isRunning"
            class="hidden sm:inline-flex items-center gap-1.5 font-mono text-[10px] font-semibold px-2 py-0.5 rounded-full"
            style="background:color-mix(in srgb,var(--signal-ok) 12%,transparent); color:var(--signal-ok); border:1px solid color-mix(in srgb,var(--signal-ok) 30%,transparent)"
          >
            <span class="w-1.5 h-1.5 rounded-full bg-[color:var(--signal-ok)] animate-pulse inline-block" />
            Live
          </span>
        </div>

        <!-- Status badges -->
        <div class="hidden md:flex items-center gap-3 text-xs font-mono text-muted-foreground">
          <span v-if="store.mqttConnected" class="signal-badge signal-badge--on">{{ brokerHost }}</span>
          <span v-if="store.status?.bdSeq !== undefined" class="text-text-dim">
            bdSeq <span class="text-foreground font-semibold">{{ store.status.bdSeq }}</span>
          </span>
          <span v-if="(store.status?.errorCount ?? 0) > 0" class="signal-badge signal-badge--off">
            {{ store.status!.errorCount }} err
          </span>
        </div>
      </header>

      <!-- Config slide-in panel -->
      <Transition name="slide-down">
        <div v-if="showConfig" class="shrink-0 border-b border-border bg-card overflow-y-auto max-h-[65vh]"
             style="box-shadow: var(--shadow-md)">
          <QuickConfig @close="showConfig = false" />
        </div>
      </Transition>

      <!-- Tab content -->
      <main class="flex-1 min-h-0 overflow-auto p-4 md:p-6">
        <Dashboard       v-if="activeTab === 'dashboard'" />
        <LiveValues      v-if="activeTab === 'live'" />
        <BrokerConfig    v-if="activeTab === 'broker'" />
        <SparkplugCfg    v-if="activeTab === 'sparkplug'" />
        <OutstationsConfig v-if="activeTab === 'outstations'" />
        <ModbusConfig    v-if="activeTab === 'modbus'" />
        <MappingTable    v-if="activeTab === 'mappings'" />
        <SystemPanel     v-if="activeTab === 'system'" />
        <LogConsole      v-if="activeTab === 'logs'" />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useGatewayStore } from '@/stores/gateway'
import { api } from '@/api/client'
import Dashboard     from '@/components/Dashboard.vue'
import LiveValues    from '@/components/LiveValues.vue'
import BrokerConfig  from '@/components/BrokerConfig.vue'
import SparkplugCfg  from '@/components/SparkplugConfig.vue'
import OutstationsConfig from '@/components/OutstationsConfig.vue'
import ModbusConfig   from '@/components/ModbusConfig.vue'
import MappingTable  from '@/components/MappingTable.vue'
import LogConsole    from '@/components/LogConsole.vue'
import QuickConfig   from '@/components/QuickConfig.vue'
import SystemPanel   from '@/components/SystemPanel.vue'

// ── Inline SVG icon components ──────────────────────────────────
const IconGauge  = { template: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22C6.48 22 2 17.52 2 12S6.48 2 12 2s10 4.48 10 10"/><path d="m12 12-3-5"/><circle cx="12" cy="12" r="1.5"/></svg>' }
const IconLive   = { template: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 12h-4l-3 9L9 3l-3 9H2"/></svg>' }
const IconWifi   = { template: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M5 12.55a11 11 0 0 1 14.08 0"/><path d="M1.42 9a16 16 0 0 1 21.16 0"/><path d="M8.53 16.11a6 6 0 0 1 6.95 0"/><circle cx="12" cy="20" r="1"/></svg>' }
const IconSpark  = { template: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/></svg>' }
const IconDevice = { template: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="4" y="2" width="16" height="20" rx="1"/><line x1="8" y1="6" x2="16" y2="6"/><line x1="8" y1="10" x2="16" y2="10"/><circle cx="12" cy="17" r="1"/></svg>' }
const IconChip   = { template: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="6" y="6" width="12" height="12" rx="1"/><path d="M9 2v3M15 2v3M9 19v3M15 19v3M2 9h3M2 15h3M19 9h3M19 15h3"/></svg>' }
const IconMap    = { template: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>' }
const IconLog    = { template: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>' }
const IconCpu    = { template: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="6" height="6"/><path d="M9 2v2M15 2v2M9 20v2M15 20v2M2 9h2M2 15h2M20 9h2M20 15h2"/></svg>' }

const store = useGatewayStore()

const tabs = [
  { id: 'dashboard', label: 'Dashboard', icon: IconGauge  },
  { id: 'live',      label: 'Live Values', icon: IconLive },
  { id: 'broker',    label: 'Broker',    icon: IconWifi   },
  { id: 'sparkplug', label: 'Sparkplug', icon: IconSpark  },
  { id: 'outstations', label: 'Outstations', icon: IconDevice },
  { id: 'modbus',    label: 'Modbus',    icon: IconChip   },
  { id: 'mappings',  label: 'Mappings',  icon: IconMap    },
  { id: 'system',    label: 'System',    icon: IconCpu    },
  { id: 'logs',      label: 'Log',       icon: IconLog    },
] as const

type TabId = typeof tabs[number]['id']
const activeTab  = ref<TabId>('dashboard')
const collapsed  = ref(false)
const showConfig = ref(false)
const dark       = ref(false)

const activeTabLabel = computed(() => tabs.find(t => t.id === activeTab.value)?.label ?? '')

const brokerHost = computed(() => {
  if (!store.status?.mqttConnected) return '—'
  // Strip scheme prefix (tcp://, ssl://) from the configured broker URL for display.
  const broker = mqttBroker.value
  return broker.replace(/^[a-z]+:\/\//, '') || '—'
})

const mqttBroker = ref('')
api.getMQTT().then(c => { mqttBroker.value = c.broker }).catch(() => {})

async function doStart() {
  try { await store.startGateway() }
  catch (e: unknown) { store.addLog('error', 'Start: ' + (e instanceof Error ? e.message : String(e))) }
}
async function doStop() { await store.stopGateway() }

function toggleSidebar() {
  collapsed.value = !collapsed.value
  localStorage.setItem('sb:collapsed', collapsed.value ? '1' : '0')
}
function toggleTheme() {
  dark.value = !dark.value
  localStorage.setItem('theme', dark.value ? 'dark' : 'light')
  document.documentElement.classList.toggle('dark', dark.value)
}

onMounted(() => {
  collapsed.value = localStorage.getItem('sb:collapsed') === '1'
  dark.value = localStorage.getItem('theme') === 'dark' ||
    (!localStorage.getItem('theme') && window.matchMedia('(prefers-color-scheme: dark)').matches)
  document.documentElement.classList.toggle('dark', dark.value)

  store.loadStatus()
  store.loadOutstations()
  store.loadModbusDevices()
  store.loadMappings()
  store.connectWS()
  setInterval(() => store.loadStatus(), 4000)
})
</script>

<style scoped>
.slide-down-enter-active { transition: max-height 0.25s ease, opacity 0.2s ease; overflow: hidden; }
.slide-down-leave-active { transition: max-height 0.2s ease, opacity 0.15s ease; overflow: hidden; }
.slide-down-enter-from, .slide-down-leave-to { max-height: 0; opacity: 0; }
.slide-down-enter-to, .slide-down-leave-from { max-height: 65vh; opacity: 1; }
</style>
