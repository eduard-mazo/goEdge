<template>
  <div class="space-y-6 animate-in">

    <!-- KPI strip -->
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

    <!-- Outstations grid -->
    <section>
      <div class="flex items-center gap-3 mb-3">
        <h3 class="font-sans font-bold text-xs uppercase tracking-widest text-text-secondary">Estaciones DNP3</h3>
        <div class="rule-brand flex-1" />
        <span class="font-mono text-[10px] text-text-dim">{{ osList.length }} est.</span>
      </div>

      <div v-if="!osList.length"
           class="forge-panel text-center py-10">
        <p class="font-mono text-xs text-text-dim">Sin estaciones — agrégalas en el panel Estaciones</p>
      </div>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-3">
        <div
          v-for="o in osList" :key="o.id"
          class="forge-panel p-4 flex items-start gap-3 transition-all duration-200 hover:shadow-md hover:border-[color:var(--tk-border-bright)]"
        >
          <div :class="['led mt-1 shrink-0', o.connected ? 'led--green' : 'led--red']" />
          <div class="min-w-0 flex-1">
            <div class="font-sans font-bold text-sm truncate">{{ o.label || o.id }}</div>
            <div class="font-mono text-[11px] text-text-secondary mt-0.5">{{ o.addr }}</div>
            <div v-if="o.lastError && !o.connected" class="font-mono text-[10px] text-red-base mt-1 truncate" :title="o.lastError">
              {{ o.lastError }}
            </div>
            <div class="flex gap-4 mt-2.5 font-mono text-[11px]">
              <span class="text-text-dim">
                rx <span class="text-foreground font-medium">{{ o.measurementsRx ?? 0 }}</span>
              </span>
              <span class="text-text-dim" v-if="o.lastReadAt">
                últ <span class="text-foreground font-medium">{{ relTime(o.lastReadAt) }}</span>
              </span>
            </div>
          </div>
          <span :class="['signal-badge self-start shrink-0', o.connected ? 'signal-badge--on' : 'signal-badge--off']">
            {{ o.connected ? 'OK' : 'FALLA' }}
          </span>
        </div>
      </div>
    </section>

    <!-- Live readings table -->
    <section>
      <div class="flex items-center gap-3 mb-3">
        <h3 class="font-sans font-bold text-xs uppercase tracking-widest text-text-secondary">Lecturas en vivo</h3>
        <div class="rule-brand flex-1" />
        <span class="font-mono text-[10px] text-text-dim">{{ readings.length }} señales</span>
      </div>

      <div v-if="!readings.length"
           class="forge-panel text-center py-10">
        <p class="font-mono text-xs text-text-dim">Sin lecturas — inicia el gateway para recibir datos</p>
      </div>

      <div v-else class="forge-panel overflow-x-auto">
        <table>
          <thead>
            <tr>
              <th>Métrica</th>
              <th class="text-right">Valor</th>
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

const osList = computed(() => Object.values(store.status?.outstations ?? {}))

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

function relTime(iso: string) {
  const t = new Date(iso).getTime()
  if (!isFinite(t)) return '—'
  const delta = Math.max(0, Date.now() - t)
  if (delta < 60_000) return 'hace ' + Math.round(delta / 1000) + 's'
  if (delta < 3_600_000) return 'hace ' + Math.round(delta / 60_000) + 'm'
  return 'hace ' + Math.round(delta / 3_600_000) + 'h'
}
</script>
