<template>
  <div class="h-screen w-screen overflow-hidden flex bg-background text-foreground" :class="{ dark }">

    <!-- Mobile backdrop (only when the drawer is open) -->
    <Transition name="fade">
      <div v-if="mobileOpen" class="fixed inset-0 bg-black/50 z-40 md:hidden" @click="mobileOpen = false" />
    </Transition>

    <!-- ── SIDEBAR ────────────────────────────────────────────────── -->
    <!-- Mobile: off-canvas drawer (fixed, slides in). Desktop: in-flow, collapsible. -->
    <aside
      class="bg-sidebar text-sidebar-foreground border-r border-sidebar-border flex flex-col h-screen overflow-hidden z-50
             fixed inset-y-0 left-0 w-[270px] transition-transform duration-300 ease-out
             md:static md:transition-[width]"
      :class="[
        mobileOpen ? 'translate-x-0' : '-translate-x-full md:translate-x-0',
        collapsed ? 'md:w-[68px]' : 'md:w-[232px]',
      ]"
    >
      <!-- Brand -->
      <div class="flex items-center h-16 border-b border-sidebar-border shrink-0"
           :class="collapsed ? 'justify-center px-0' : 'gap-3 px-4'">
        <div class="grid place-items-center w-9 h-9 rounded-md bg-[color:var(--epm-citrico)] text-[color:var(--epm-bosque)] font-black text-base shrink-0 leading-none">
          M
        </div>
        <div v-if="!collapsed" class="leading-none min-w-0">
          <div class="font-sans text-base font-extrabold tracking-tight text-white truncate">DNP3 GW</div>
          <div class="text-[10px] uppercase tracking-[0.18em] text-[color:var(--epm-citrico)] mt-0.5 font-semibold">Sparkplug B</div>
        </div>
        <!-- Mobile: close drawer -->
        <button class="md:hidden ml-auto text-white/70 hover:text-white p-1" @click="mobileOpen = false" title="Cerrar">
          <X class="h-5 w-5" />
        </button>
      </div>

      <!-- Control section -->
      <div class="border-b border-sidebar-border shrink-0 space-y-2" :class="collapsed ? 'p-2' : 'px-3 py-3'">
        <!-- MQTT status -->
        <div v-if="!collapsed" class="flex items-center gap-2 px-2 py-1.5 rounded-md"
             :class="store.mqttConnected ? 'bg-sidebar-accent' : ''">
          <div :class="['led', store.mqttConnected ? 'led--green' : 'led--dim']" />
          <span class="font-mono text-[11px] truncate" :class="store.mqttConnected ? 'text-white' : 'text-white/50'">
            {{ store.mqttConnected ? 'MQTT en línea' : 'MQTT sin conexión' }}
          </span>
        </div>
        <div v-else class="flex justify-center py-1" :title="store.mqttConnected ? 'MQTT en línea' : 'MQTT sin conexión'">
          <div :class="['led', store.mqttConnected ? 'led--green' : 'led--dim']" />
        </div>

        <!-- Run / stop -->
        <button v-if="!store.isRunning" @click="doStart"
                class="w-full btn-primary text-xs" :class="collapsed ? 'px-0 h-9' : 'py-1.5'"
                :title="collapsed ? 'Iniciar gateway' : ''">
          <span v-if="collapsed" class="text-sm leading-none">▶</span>
          <template v-else>▶ Iniciar</template>
        </button>
        <button v-else @click="doStop"
                class="w-full btn-danger text-xs" :class="collapsed ? 'px-0 h-9' : 'py-1.5'"
                :title="collapsed ? 'Detener gateway' : ''">
          <span v-if="collapsed" class="text-sm leading-none">■</span>
          <template v-else>■ Detener</template>
        </button>

        <!-- Configure toggle -->
        <button @click="showConfig = !showConfig"
                class="w-full flex items-center rounded-md text-xs font-semibold transition-colors"
                :class="[collapsed ? 'justify-center h-9 px-0' : 'gap-2 px-2 py-1.5',
                         showConfig ? 'bg-[color:var(--epm-citrico)] text-[color:var(--epm-bosque)]' : 'text-white/70 hover:bg-sidebar-accent']"
                title="Configuración rápida">
          <component :is="IconCog" class="h-4 w-4 shrink-0" />
          <span v-if="!collapsed">{{ showConfig ? 'Ocultar' : 'Configurar' }}</span>
        </button>
      </div>

      <!-- Nav -->
      <nav class="flex-1 min-h-0 overflow-y-auto py-3 px-2 space-y-0.5 sidebar-nav-scroll">
        <div v-if="!collapsed" class="text-[10px] uppercase tracking-[0.2em] text-[color:var(--epm-citrico)] font-semibold mb-2 px-2">Paneles</div>
        <button
          v-for="tab in tabs"
          :key="tab.id"
          class="group w-full flex items-center rounded-md text-sm transition-colors hover:bg-sidebar-accent relative"
          :class="[collapsed ? 'justify-center h-11 px-0' : 'gap-3 px-3 py-2.5',
                   activeTab === tab.id ? 'bg-sidebar-accent text-sidebar-accent-foreground' : 'text-sidebar-foreground/70']"
          :title="tab.label"
          @click="activeTab = tab.id; mobileOpen = false">
          <span v-if="activeTab === tab.id" class="absolute left-0 top-1.5 bottom-1.5 w-0.5 bg-[color:var(--sidebar-primary)] rounded-r" />
          <component :is="tab.icon" class="h-[18px] w-[18px] shrink-0" />
          <span v-if="!collapsed" class="truncate font-semibold">{{ tab.label }}</span>
        </button>
      </nav>

      <!-- Bottom controls -->
      <div class="border-t border-sidebar-border p-2 space-y-0.5 shrink-0">
        <!-- Running indicator (expanded only) -->
        <div v-if="store.isRunning && !collapsed"
             class="flex items-center gap-2 px-3 py-1.5 mb-1 rounded-md bg-[color:color-mix(in_srgb,var(--signal-warn)_15%,transparent)] border border-[color:var(--signal-warn)]">
          <div class="led led--amber" />
          <span class="text-[11px] font-semibold text-[color:var(--signal-warn)] font-mono truncate">
            pub {{ store.status?.publishCount ?? 0 }}
          </span>
        </div>

        <button @click="toggleTheme"
                class="w-full flex items-center rounded-md text-sm hover:bg-sidebar-accent transition-colors text-sidebar-foreground/70"
                :class="collapsed ? 'justify-center h-10 px-0' : 'gap-3 px-3 py-2'"
                :title="dark ? 'Modo claro' : 'Modo oscuro'">
          <svg v-if="dark" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="5"/><path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/></svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12.79A9 9 0 1111.21 3a7 7 0 009.79 9.79z"/></svg>
          <span v-if="!collapsed">{{ dark ? 'Modo claro' : 'Modo oscuro' }}</span>
        </button>
        <button @click="toggleSidebar"
                class="hidden md:flex w-full items-center rounded-md text-sm hover:bg-sidebar-accent transition-colors text-sidebar-foreground/70"
                :class="collapsed ? 'justify-center h-10 px-0' : 'gap-3 px-3 py-2'"
                :title="collapsed ? 'Expandir menú' : 'Contraer menú'">
          <svg v-if="collapsed" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M9 3v18M15 9l3 3-3 3"/></svg>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="18" height="18" rx="2"/><path d="M9 3v18M9 9l-3 3 3 3"/></svg>
          <span v-if="!collapsed">{{ collapsed ? 'Expandir' : 'Contraer' }}</span>
        </button>
      </div>
    </aside>

    <!-- ── MAIN COLUMN ────────────────────────────────────────────── -->
    <div class="flex-1 flex flex-col min-w-0 h-screen overflow-hidden">

      <!-- Top rail -->
      <header class="shrink-0 flex items-center gap-3 h-14 px-4 sm:px-6 border-b border-border bg-card z-30"
              style="box-shadow: var(--shadow-sm)">
        <button class="md:hidden btn-ghost px-2 py-1.5" @click="mobileOpen = !mobileOpen" title="Menú">
          <Menu class="h-4 w-4" />
        </button>
        <div class="flex items-center gap-2 min-w-0 flex-1">
          <h2 class="font-sans font-extrabold text-base leading-none truncate">{{ activeTabLabel }}</h2>
          <span
            v-if="store.isRunning"
            class="hidden sm:inline-flex items-center gap-1.5 font-mono text-[10px] font-semibold px-2 py-0.5 rounded-full"
            style="background:color-mix(in srgb,var(--signal-ok) 12%,transparent); color:var(--signal-ok); border:1px solid color-mix(in srgb,var(--signal-ok) 30%,transparent)"
          >
            <span class="w-1.5 h-1.5 rounded-full bg-[color:var(--signal-ok)] animate-pulse inline-block" />
            En vivo
          </span>
        </div>

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
        <SerialConfig    v-if="activeTab === 'serial'" />
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
import SerialConfig   from '@/components/SerialConfig.vue'
import MappingTable  from '@/components/MappingTable.vue'
import LogConsole    from '@/components/LogConsole.vue'
import QuickConfig   from '@/components/QuickConfig.vue'
import SystemPanel   from '@/components/SystemPanel.vue'
import {
  LayoutDashboard, Activity, Wifi, Zap, Server, Cpu, Cable,
  Waypoints, Gauge, ScrollText, Settings, Menu, X,
} from 'lucide-vue-next'

// Configure-toggle icon (referenced in template as IconCog)
const IconCog = Settings

const store = useGatewayStore()

const tabs = [
  { id: 'dashboard',  label: 'Resumen',      icon: LayoutDashboard },
  { id: 'live',       label: 'En vivo',      icon: Activity        },
  { id: 'broker',     label: 'Broker',       icon: Wifi            },
  { id: 'sparkplug',  label: 'Sparkplug',    icon: Zap             },
  { id: 'outstations',label: 'Estaciones',   icon: Server          },
  { id: 'modbus',     label: 'Modbus/TCP',   icon: Cpu             },
  { id: 'serial',     label: 'Serial · RTU', icon: Cable           },
  { id: 'mappings',   label: 'Señales',      icon: Waypoints       },
  { id: 'system',     label: 'Sistema',      icon: Gauge           },
  { id: 'logs',       label: 'Registro',     icon: ScrollText      },
] as const

type TabId = typeof tabs[number]['id']
const activeTab  = ref<TabId>('dashboard')
const collapsed  = ref(false)   // desktop-only: narrow rail
const mobileOpen = ref(false)   // mobile-only: drawer open
const showConfig = ref(false)
const dark       = ref(false)

const activeTabLabel = computed(() => tabs.find(t => t.id === activeTab.value)?.label ?? '')

const brokerHost = computed(() => {
  if (!store.status?.mqttConnected) return '—'
  return mqttBroker.value.replace(/^[a-z]+:\/\//, '') || '—'
})

const mqttBroker = ref('')
api.getMQTT().then(c => { mqttBroker.value = c.broker }).catch(() => {})

async function doStart() {
  try { await store.startGateway() }
  catch (e: unknown) { store.addLog('error', 'Inicio: ' + (e instanceof Error ? e.message : String(e))) }
}
async function doStop() { await store.stopGateway() }

function toggleSidebar() {
  collapsed.value = !collapsed.value
  localStorage.setItem('sb:collapsed', collapsed.value ? '1' : '0')
}
function applyTheme() {
  document.documentElement.classList.toggle('dark', dark.value)
}
function toggleTheme() {
  dark.value = !dark.value
  localStorage.setItem('theme', dark.value ? 'dark' : 'light')
  applyTheme()
}

onMounted(() => {
  collapsed.value = localStorage.getItem('sb:collapsed') === '1'
  const saved = localStorage.getItem('theme')
  dark.value = saved === 'dark' ||
    (!saved && window.matchMedia('(prefers-color-scheme: dark)').matches)
  applyTheme()

  store.loadStatus()
  store.loadOutstations()
  store.loadModbusDevices()
  store.loadSerialDevices()
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

.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
