<!-- Dashboard ("Resumen") — the unified operational view. Merges the former
     Resumen + En vivo tabs into one panel: system KPIs, multiprotocol source
     cards (DNP3 / Modbus-TCP / RTU) and the full live SCADA grid. -->
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

  </div>
</template>

<script setup lang="ts">
import { useGatewayStore } from '@/stores/gateway'
import StatCard from './StatCard.vue'
import SourceCards from './SourceCards.vue'
import LiveValues from './LiveValues.vue'

const store = useGatewayStore()

function fmtCount(n?: number) {
  if (n === undefined) return '—'
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(1) + 'M'
  if (n >= 1_000) return (n / 1_000).toFixed(1) + 'k'
  return n.toString()
}
</script>
