<template>
  <div class="space-y-6 animate-in">

    <!-- KPI strip -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
      <StatCard
        label="Gateway"
        :value="store.isRunning ? 'Running' : 'Stopped'"
        :color="store.isRunning ? 'ok' : 'dim'"
        :sub="store.status?.uptime ? 'up ' + store.status.uptime : undefined"
      />
      <StatCard
        label="MQTT"
        :value="store.mqttConnected ? 'Online' : 'Offline'"
        :color="store.mqttConnected ? 'ok' : 'error'"
        :sub="store.status?.bdSeq !== undefined ? 'bdSeq ' + store.status.bdSeq : undefined"
      />
      <StatCard
        label="Published"
        :value="fmtCount(store.status?.publishCount)"
        color="ok"
      />
      <StatCard
        label="Errors"
        :value="fmtCount(store.status?.errorCount)"
        :color="(store.status?.errorCount ?? 0) > 0 ? 'error' : 'dim'"
      />
    </div>

    <!-- Devices grid -->
    <section>
      <div class="flex items-center gap-3 mb-3">
        <h3 class="font-sans font-bold text-xs uppercase tracking-widest text-text-secondary">Modbus Devices</h3>
        <div class="rule-brand flex-1" />
        <span class="font-mono text-[10px] text-text-dim">{{ devList.length }} device{{ devList.length !== 1 ? 's' : '' }}</span>
      </div>

      <div v-if="!devList.length"
           class="forge-panel text-center py-10">
        <p class="font-mono text-xs text-text-dim">No devices configured — add one in the Devices panel</p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
        <div
          v-for="dev in devList" :key="dev.id"
          class="forge-panel p-4 flex items-start gap-3 transition-all duration-200 hover:shadow-md hover:border-[color:var(--tk-border-bright)]"
        >
          <div :class="['led mt-1 shrink-0', dev.connected ? 'led--green' : 'led--red']" />
          <div class="min-w-0 flex-1">
            <div class="font-sans font-bold text-sm truncate">{{ dev.label || dev.id }}</div>
            <div class="font-mono text-[11px] text-text-secondary mt-0.5">{{ dev.addr }}</div>
            <div class="flex gap-4 mt-2.5 font-mono text-[11px]">
              <span class="text-text-dim">
                reads <span class="text-foreground font-medium">{{ dev.reads }}</span>
              </span>
              <span class="text-text-dim">
                errors
                <span :class="dev.errors > 0 ? 'text-red-base font-semibold' : 'text-foreground font-medium'">{{ dev.errors }}</span>
              </span>
            </div>
          </div>
          <span :class="['signal-badge self-start shrink-0', dev.connected ? 'signal-badge--on' : 'signal-badge--off']">
            {{ dev.connected ? 'OK' : 'FAULT' }}
          </span>
        </div>
      </div>
    </section>

    <!-- Live readings table -->
    <section>
      <div class="flex items-center gap-3 mb-3">
        <h3 class="font-sans font-bold text-xs uppercase tracking-widest text-text-secondary">Live Readings</h3>
        <div class="rule-brand flex-1" />
        <span class="font-mono text-[10px] text-text-dim">{{ readings.length }} signal{{ readings.length !== 1 ? 's' : '' }}</span>
      </div>

      <div v-if="!readings.length"
           class="forge-panel text-center py-10">
        <p class="font-mono text-xs text-text-dim">No readings yet — start the gateway to begin polling</p>
      </div>

      <div v-else class="forge-panel overflow-x-auto">
        <table>
          <thead>
            <tr>
              <th>Metric</th>
              <th class="text-right">Value</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="r in readings" :key="r.metric">
              <td class="font-mono text-xs text-text-secondary">{{ r.metric }}</td>
              <td class="font-mono text-right font-semibold text-bosque">{{ fmtVal(r.value) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useGatewayStore } from '@/stores/gateway'
import StatCard from './StatCard.vue'

const store = useGatewayStore()

const devList = computed(() => Object.values(store.status?.devices ?? {}))

const readings = computed(() =>
  Object.entries(store.status?.lastReadings ?? {})
    .map(([metric, value]) => ({ metric, value }))
    .sort((a, b) => a.metric.localeCompare(b.metric))
)

function fmtVal(v: number) {
  if (Number.isInteger(v)) return v.toString()
  return v.toFixed(4)
}

function fmtCount(n?: number) {
  if (n === undefined) return '—'
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'k'
  return n.toString()
}
</script>
