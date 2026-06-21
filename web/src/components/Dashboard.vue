<!-- Dashboard ("Resumen") — the unified operational view. Merges the former
     Resumen + En vivo tabs into one panel: system KPIs, multiprotocol source
     cards (DNP3 / Modbus-TCP / RTU), the full live SCADA grid and the
     host/system metrics that bypass the mapping path. -->
<template>
  <div class="space-y-6 animate-in">

    <!-- System KPI strip -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
      <StatCard
        label="Gateway"
        :value="store.isRunning ? 'Activo' : 'Detenido'"
        :color="store.isRunning ? 'ok' : 'dim'"
        :sub="store.status?.uptime ? 'activo ' + store.status.uptime : undefined"
      />
      <StatCard
        label="MQTT"
        :value="store.mqttConnected ? 'En línea' : 'Sin conexión'"
        :color="store.mqttConnected ? 'ok' : 'error'"
        :sub="store.status?.bdSeq !== undefined ? 'bdSeq ' + store.status.bdSeq : undefined"
      />
      <StatCard
        label="Publicados"
        :value="fmtCount(store.status?.publishCount)"
        color="ok"
      />
      <StatCard
        label="Errores"
        :value="fmtCount(store.status?.errorCount)"
        :color="(store.status?.errorCount ?? 0) > 0 ? 'error' : 'dim'"
      />
    </div>

    <!-- Multiprotocol source cards -->
    <SourceCards />

    <!-- Live readings — full SCADA grid (search / protocol / source / status) -->
    <section>
      <div class="flex items-center gap-3 mb-3">
        <h3 class="font-sans font-bold text-xs uppercase tracking-widest text-text-secondary">Lecturas en vivo</h3>
        <div class="rule-brand flex-1" />
      </div>
      <LiveValues />
    </section>

    <!-- System / host metrics — host telemetry (sysmon) is mirrored into
         lastReadings but bypasses the mapping path, so it never appears in the
         grid above. Surface every reading without a configured mapping here. -->
    <section v-if="systemReadings.length">
      <div class="flex items-center gap-3 mb-3">
        <h3 class="font-sans font-bold text-xs uppercase tracking-widest text-text-secondary">Métricas de sistema</h3>
        <div class="rule-brand flex-1" />
        <span class="font-mono text-[10px] text-text-dim">{{ systemReadings.length }} métricas</span>
      </div>

      <div class="forge-panel overflow-hidden">
        <div class="overflow-auto max-h-[360px]">
          <table>
            <thead class="sticky top-0 z-10">
              <tr>
                <th>Métrica</th>
                <th class="text-right">Valor</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in systemReadings" :key="r.metric">
                <td class="font-mono text-xs text-text-secondary">{{ r.metric }}</td>
                <td class="font-mono text-right font-semibold text-bosque tabular-nums">{{ fmtVal(r.value) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>

  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useGatewayStore } from '@/stores/gateway'
import StatCard from './StatCard.vue'
import SourceCards from './SourceCards.vue'
import LiveValues from './LiveValues.vue'

const store = useGatewayStore()

// Readings with no configured mapping are the host/system metrics synthesized by
// sysmon (e.g. "System/CPU/Usage_pct"). The LiveValues grid joins mappings ×
// readings, so they only show here.
const systemReadings = computed(() => {
  const mapped = new Set(store.mappings.map((m) => m.metricName))
  return Object.entries(store.status?.lastReadings ?? {})
    .filter(([metric]) => !mapped.has(metric))
    .map(([metric, value]) => ({ metric, value }))
    .sort((a, b) => a.metric.localeCompare(b.metric))
})

function fmtVal(v: number) {
  if (Number.isInteger(v)) return v.toString()
  const a = Math.abs(v)
  return v.toFixed(a >= 1000 ? 1 : a >= 1 ? 3 : 4)
}

function fmtCount(n?: number) {
  if (n === undefined) return '—'
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'k'
  return n.toString()
}
</script>

<style scoped>
.tabular-nums { font-variant-numeric: tabular-nums; }
</style>
