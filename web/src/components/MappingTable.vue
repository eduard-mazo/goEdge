<template>
  <div class="space-y-4">

    <!-- Toolbar -->
    <div class="flex items-center gap-3 flex-wrap">
      <p class="font-sans text-sm text-text-secondary flex-1">
        Señales: punto → métrica Sparkplug B (DNP3, Modbus/TCP y RTU)
      </p>
      <a :href="api.exportMappingsURL()" target="_blank" class="btn-ghost text-xs py-1">
        ↓ Exportar
      </a>
      <label class="btn-ghost text-xs py-1 cursor-pointer">
        ↑ Importar
        <input type="file" accept=".json" class="hidden" @change="importFile" />
      </label>
      <button class="btn-primary text-xs py-1.5" @click="openAdd">+ Agregar</button>
    </div>

    <!-- Filter -->
    <input v-model="filter" class="forge-input max-w-sm"
           placeholder="Filtrar por métrica, fuente, punto…" />

    <!-- Empty -->
    <div v-if="!sorted.length" class="forge-panel text-center py-12">
      <p class="font-mono text-xs text-text-dim">
        {{ store.mappings.length ? 'Sin resultados para el filtro.' : 'Aún no hay señales.' }}
      </p>
    </div>

    <!-- Table -->
    <div v-else class="forge-panel overflow-hidden">
      <div ref="viewportEl" class="overflow-auto max-h-[calc(100vh-300px)]">
      <!-- table-fixed + colgroup: virtualization keeps only ~30 rows in the DOM,
           so an auto layout would re-fit the columns to the visible window and
           the header would slip out of alignment while scrolling. Fixed widths
           keep the header locked to the data. -->
      <table class="w-full min-w-[1040px] table-fixed">
        <colgroup>
          <col />                          <!-- Métrica (flexes) -->
          <col class="w-[72px]" />         <!-- Proto -->
          <col class="w-[130px]" />        <!-- Fuente -->
          <col class="w-[210px]" />        <!-- Punto -->
          <col class="w-[96px]" />         <!-- Escala -->
          <col class="w-[80px]" />         <!-- UI -->
          <col class="w-[118px]" />        <!-- Banda muerta -->
          <col class="w-[92px]" />         <!-- Estado -->
          <col class="w-[150px]" />        <!-- Acciones -->
        </colgroup>
        <thead class="sticky top-0 z-10">
          <tr>
            <SortTh col="metric"   label="Métrica"      :sort="sort" @sort="toggle" />
            <SortTh col="proto"    label="Proto"        :sort="sort" @sort="toggle" />
            <SortTh col="source"   label="Fuente"       :sort="sort" @sort="toggle" />
            <SortTh col="point"    label="Punto"        :sort="sort" @sort="toggle" />
            <SortTh col="scale"    label="Escala"       align="right" :sort="sort" @sort="toggle" />
            <SortTh col="unit"     label="UI"           :sort="sort" @sort="toggle" />
            <SortTh col="deadband" label="Banda muerta" align="right" :sort="sort" @sort="toggle" />
            <SortTh col="estado"   label="Estado"       align="center" :sort="sort" @sort="toggle" />
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="padTop" class="vspacer" :style="{ height: padTop + 'px' }"><td colspan="9"></td></tr>
          <tr v-for="{ row: m } in visible" :key="m.id" class="vrow">
            <td class="font-mono text-xs font-medium text-foreground">{{ m.metricName }}</td>
            <td>
              <span class="proto-tag" :class="`proto-tag--${protoTag(m).cls}`">
                {{ protoTag(m).label }}
              </span>
            </td>
            <td class="font-mono text-[11px]">{{ srcId(m) }}</td>
            <td class="font-mono text-[11px] text-text-secondary whitespace-nowrap">{{ pointLabel(m) }}</td>
            <td class="font-mono text-[11px] text-text-secondary text-right tabular-nums">
              {{ m.scale ?? 1 }}{{ m.offset ? ' +' + m.offset : '' }}
            </td>
            <td class="font-mono text-[11px] text-text-dim">{{ m.engineeringUnit || '—' }}</td>
            <td class="font-mono text-[11px] text-text-secondary text-right tabular-nums">{{ m.deadband ?? 0 }}</td>
            <td class="text-center">
              <span :class="['signal-badge', m.enabled ? 'signal-badge--on' : 'signal-badge--off']">
                {{ m.enabled ? 'ON' : 'OFF' }}
              </span>
            </td>
            <td>
              <div class="flex gap-1 justify-end">
                <button class="btn-ghost text-[10px] py-0.5 px-2" @click="openEdit(m)">Editar</button>
                <button
                  class="btn-ghost text-[10px] py-0.5 px-2 !border-red-base/40 !text-red-base hover:!border-red-bright hover:!text-red-bright"
                  @click="del(m.id)"
                >Borrar</button>
              </div>
            </td>
          </tr>
          <tr v-if="padBottom" class="vspacer" :style="{ height: padBottom + 'px' }"><td colspan="9"></td></tr>
        </tbody>
      </table>
      </div>
      <div class="flex items-center justify-between px-3 py-2 border-t border-border bg-muted">
        <span class="font-mono text-[10px] text-text-dim uppercase tracking-wider">
          {{ sorted.length }} de {{ store.mappings.length }} señales
        </span>
        <span class="font-mono text-[10px] text-text-dim">ventana virtual</span>
      </div>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="modal" class="modal-backdrop overflow-y-auto py-6">
        <div class="forge-panel w-full max-w-2xl mx-auto">
          <div class="forge-header flex items-center justify-between gap-3">
            <span>{{ editId ? 'Editar señal' : 'Nueva señal' }}</span>
            <button type="button" class="modal-x" @click="modal = false" aria-label="Cerrar">✕</button>
          </div>
          <form @submit.prevent="saveMapping" class="p-5 space-y-4">

            <!-- Protocol switch -->
            <div class="seg">
              <button type="button" v-for="p in protocols" :key="p.value"
                      class="seg-btn" :class="form.protocol === p.value ? 'seg-btn--active' : ''"
                      @click="form.protocol = p.value">
                {{ p.label }}
              </button>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <Field label="Nombre de métrica" hint="Identificador único Sparkplug B">
                <input v-model="form.metricName" class="forge-input" placeholder="feeder/voltage_kv" required />
              </Field>
              <Field label="Device ID Sparkplug" hint="Vacío = métrica de nodo">
                <input v-model="form.deviceId" class="forge-input" placeholder="(nivel de nodo)" />
              </Field>
            </div>

            <!-- Source select (protocol-specific list) -->
            <Field :label="sourceLabel">
              <select v-model="form.sourceId" class="forge-input" required>
                <option value="" disabled>Elegir fuente…</option>
                <template v-if="form.protocol === 'modbus'">
                  <option v-for="d in store.modbusDevices" :key="d.id" :value="d.id">
                    {{ d.label || d.id }} ({{ d.host }}:{{ d.port }})
                  </option>
                </template>
                <template v-else-if="form.protocol === 'modbusrtu'">
                  <option v-for="d in store.serialDevices" :key="d.id" :value="d.id">
                    {{ d.label || d.id }} ({{ d.port }} #{{ d.unitId }})
                  </option>
                </template>
                <template v-else>
                  <option v-for="o in store.outstations" :key="o.id" :value="o.id">
                    {{ o.label || o.id }} ({{ o.host }}:{{ o.port }})
                  </option>
                </template>
              </select>
            </Field>

            <!-- DNP3 point identity -->
            <template v-if="form.protocol === 'dnp3'">
              <div class="grid grid-cols-3 gap-4">
                <Field label="Tipo de punto" hint="grupo/variación DNP3" class="col-span-1">
                  <select v-model="form.pointType" class="forge-input">
                    <option v-for="pt in pointTypes" :key="pt.value" :value="pt.value">{{ pt.label }}</option>
                  </select>
                </Field>
                <Field label="Índice" hint="índice base 0">
                  <input v-model.number="form.index" class="forge-input" type="number" min="0" max="65535" />
                </Field>
                <Field label="Clase de evento" hint="informativo">
                  <select v-model.number="form.eventClass" class="forge-input">
                    <option :value="0">Solo estático</option>
                    <option :value="1">Clase 1</option>
                    <option :value="2">Clase 2</option>
                    <option :value="3">Clase 3</option>
                  </select>
                </Field>
              </div>
            </template>

            <!-- Modbus point identity -->
            <template v-else>
              <div class="grid grid-cols-2 gap-4">
                <Field label="Función">
                  <select v-model="form.function" class="forge-input">
                    <option v-for="f in modbusFunctions" :key="f.value" :value="f.value">{{ f.label }}</option>
                  </select>
                </Field>
                <Field label="Dirección" hint="dirección inicial de registro/coil">
                  <input v-model.number="form.address" class="forge-input" type="number" min="0" max="65535" />
                </Field>
              </div>
              <div class="grid grid-cols-3 gap-4" v-if="isRegister">
                <Field label="Tipo de dato">
                  <select v-model="form.dataType" class="forge-input">
                    <option v-for="dt in modbusDataTypes" :key="dt" :value="dt">{{ dt }}</option>
                  </select>
                </Field>
                <Field label="Orden de bytes" hint="orden para 32/64-bit">
                  <select v-model="form.byteOrder" class="forge-input">
                    <option v-for="bo in byteOrders" :key="bo" :value="bo">{{ bo }}</option>
                  </select>
                </Field>
                <Field label="Cantidad" hint="0 = según tipo de dato">
                  <input v-model.number="form.quantity" class="forge-input" type="number" min="0" max="125" placeholder="0" />
                </Field>
              </div>
            </template>

            <div class="grid grid-cols-3 gap-4">
              <Field label="Escala" hint="valor × escala + offset">
                <input v-model.number="form.scale" class="forge-input" type="number" step="any" placeholder="1" />
              </Field>
              <Field label="Offset">
                <input v-model.number="form.offset" class="forge-input" type="number" step="any" placeholder="0" />
              </Field>
              <Field label="Unidad ing.">
                <input v-model="form.engineeringUnit" class="forge-input" placeholder="kV" />
              </Field>
            </div>

            <Field label="Banda muerta" hint="cambio mínimo (en unidades) para publicar; 0 = siempre">
              <input v-model.number="form.deadband" class="forge-input max-w-[180px]" type="number" step="any" placeholder="0" />
            </Field>

            <!-- UNS / catálogo (contract v3) -->
            <div class="border border-border rounded-sm p-4 space-y-4" style="background:var(--tk-surface)">
              <div class="text-muted-foreground text-[10px] uppercase tracking-widest font-sans font-semibold">
                Catálogo UNS (NBIRTH/DBIRTH)
              </div>
              <div class="grid grid-cols-2 gap-4">
                <Field label="Nombre" hint="uns/name — nombre de la señal en el catálogo">
                  <input v-model="form.nombre" class="forge-input" placeholder="Valvula abierta" />
                </Field>
                <Field label="Descripción" hint="uns/description">
                  <input v-model="form.descripcion" class="forge-input" placeholder="Valvula gas confirmación apertura" />
                </Field>
              </div>
              <div class="grid grid-cols-2 gap-4">
                <Field label="Código de señal" hint="uns/code — vacío = hoja de la métrica (máx. 20)">
                  <input v-model="form.signalCode" class="forge-input" maxlength="20" :placeholder="unsCodeDefault" />
                </Field>
                <Field label="Instancia" hint="uns/instance — vacío = carpeta de la métrica (máx. 30)">
                  <input v-model="form.instance" class="forge-input" maxlength="30" :placeholder="unsInstanceDefault" />
                </Field>
              </div>
              <p class="font-mono text-[11px] text-text-dim">
                uns/code = <span class="text-citrico">{{ form.signalCode || unsCodeDefault }}</span>
                · uns/instance = <span class="text-citrico">{{ form.instance || unsInstanceDefault }}</span>
              </p>
            </div>

            <label v-if="form.protocol === 'dnp3'" class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary">
              <input type="checkbox" v-model="form.publishOnPoll" />
              Publicar también en lecturas estáticas (no solo eventos)
            </label>
            <p v-else class="font-mono text-[11px] text-text-dim">
              Las lecturas {{ form.protocol === 'modbusrtu' ? 'Modbus RTU' : 'Modbus' }} son sondeadas — cada cambio se publica (según banda muerta).
            </p>
            <label class="flex items-center gap-2 cursor-pointer font-sans text-sm text-text-secondary">
              <input type="checkbox" v-model="form.enabled" /> Habilitada
            </label>

            <p v-if="error" class="font-mono text-xs text-red-bright">⚠ {{ error }}</p>

            <div class="flex justify-end gap-3 pt-2 border-t border-border">
              <button type="button" class="btn-ghost text-xs" @click="modal = false">Cancelar</button>
              <button type="submit" class="btn-primary text-xs">Guardar</button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useGatewayStore } from '@/stores/gateway'
import { api, type SignalMapping, type Protocol, type PointType, type ModbusFunction, type ModbusDataType } from '@/api/client'
import Field from './Field.vue'
import SortTh from './SortTh.vue'
import { useSort } from '@/composables/useSort'
import { useVirtualRows } from '@/composables/useVirtualRows'

const store  = useGatewayStore()
const modal  = ref(false)
const editId = ref('')
const error  = ref('')
const filter = ref('')

const protocols: { value: Protocol; label: string }[] = [
  { value: 'dnp3',      label: 'DNP3' },
  { value: 'modbus',    label: 'Modbus/TCP' },
  { value: 'modbusrtu', label: 'Modbus RTU' },
]

const sourceLabel = computed(() => ({
  modbus:    'Dispositivo Modbus',
  modbusrtu: 'Dispositivo serial (RS-485)',
  dnp3:      'Estación',
}[form.value.protocol ?? 'dnp3']))

const pointTypes: { value: PointType; label: string }[] = [
  { value: 'binary',               label: 'Entrada binaria (g1/g2)' },
  { value: 'double_bit_binary',    label: 'Binaria doble bit (g3/g4)' },
  { value: 'binary_output_status', label: 'Estado salida binaria (g10/g11)' },
  { value: 'counter',              label: 'Contador (g20/g22)' },
  { value: 'frozen_counter',       label: 'Contador congelado (g21/g23)' },
  { value: 'analog',               label: 'Entrada analógica (g30/g32)' },
  { value: 'analog_output_status', label: 'Estado salida analógica (g40/g42)' },
  { value: 'octet_string',         label: 'Cadena de octetos (g110/g111)' },
]

const modbusFunctions: { value: ModbusFunction; label: string }[] = [
  { value: 'coil',             label: 'Coil (FC01)' },
  { value: 'discrete_input',   label: 'Discrete Input (FC02)' },
  { value: 'holding_register', label: 'Holding Register (FC03)' },
  { value: 'input_register',   label: 'Input Register (FC04)' },
]
const modbusDataTypes: ModbusDataType[] = ['int16', 'uint16', 'int32', 'uint32', 'float32', 'float64']
const byteOrders = ['ABCD', 'CDAB', 'BADC', 'DCBA']

const isRegister = computed(
  () => form.value.function === 'holding_register' || form.value.function === 'input_register',
)

// Mirror of the Go defaults (mapping.unsCode / unsInstance): uns/code is the
// LEAF of the metric name, uns/instance the folder path ("default" when flat).
const unsCodeDefault = computed(() => {
  const name = form.value.metricName
  const i = name.lastIndexOf('/')
  return (i >= 0 ? name.slice(i + 1) : name) || '…'
})
const unsInstanceDefault = computed(() => {
  const name = form.value.metricName
  const i = name.lastIndexOf('/')
  return i >= 0 ? name.slice(0, i) : 'default'
})

function isModbusFraming(m: SignalMapping) { return m.protocol === 'modbus' || m.protocol === 'modbusrtu' }
function srcId(m: SignalMapping) { return m.sourceId || m.outstationId || '—' }

function protoTag(m: SignalMapping): { cls: string; label: string } {
  if (m.protocol === 'modbusrtu') return { cls: 'rtu', label: 'RTU' }
  if (m.protocol === 'modbus')    return { cls: 'mb',  label: 'MB' }
  return { cls: 'dnp', label: 'DNP3' }
}

const shortPoint = (pt: string) => ({
  binary: 'g1', double_bit_binary: 'g3', binary_output_status: 'g10',
  counter: 'g20', frozen_counter: 'g21', analog: 'g30',
  analog_output_status: 'g40', octet_string: 'g110',
}[pt] ?? pt)

function pointLabel(m: SignalMapping): string {
  if (isModbusFraming(m)) {
    const fn = (m.function ?? '').replace(/_/g, ' ')
    const dt = m.dataType ? ` · ${m.dataType}` : ''
    return `${fn} @${m.address ?? 0}${dt}`
  }
  return `${shortPoint(m.pointType)} #${m.index}${m.eventClass ? ' · C' + m.eventClass : ''}`
}

const matched = computed(() => {
  const f = filter.value.toLowerCase()
  if (!f) return store.mappings
  return store.mappings.filter((m) =>
    `${m.metricName} ${m.deviceId ?? ''} ${srcId(m)} ${pointLabel(m)} ${m.nombre ?? ''} ${m.signalCode ?? ''}`.toLowerCase().includes(f),
  )
})

// Sortable view over the filtered rows. Computed columns (proto/source/point)
// resolve through accessors; the rest read the matching mapping field.
const { sort, toggle, sorted } = useSort(matched, {
  initial: 'metric',
  accessors: {
    metric:   (m) => m.metricName,
    proto:    (m) => protoTag(m).label,
    source:   (m) => srcId(m),
    point:    (m) => pointLabel(m),
    scale:    (m) => m.scale ?? 1,
    unit:     (m) => m.engineeringUnit ?? '',
    deadband: (m) => m.deadband ?? 0,
    estado:   (m) => (m.enabled ? 1 : 0),
  },
})

// Windowed rendering so 1000+ mappings stay smooth (only visible rows in DOM).
const viewportEl = ref<HTMLElement | null>(null)
const { visible, padTop, padBottom } = useVirtualRows(sorted, viewportEl, { rowHeight: 40 })

const emptyForm = (): SignalMapping => ({
  id: '', metricName: '', deviceId: '',
  protocol: 'dnp3', sourceId: '', outstationId: '',
  pointType: 'analog', index: 0, eventClass: 1,
  function: 'holding_register', address: 0, quantity: 0, dataType: 'float32', byteOrder: 'ABCD',
  scale: 1, offset: 0, engineeringUnit: '',
  signalCode: '', instance: '', nombre: '', descripcion: '',
  deadband: 0, publishOnPoll: false,
  enabled: true,
})
const form = ref<SignalMapping>(emptyForm())

function openAdd() { editId.value = ''; form.value = emptyForm(); error.value = ''; modal.value = true }
function openEdit(m: SignalMapping) {
  editId.value = m.id
  form.value = { ...emptyForm(), ...m, protocol: m.protocol ?? 'dnp3', sourceId: m.sourceId || m.outstationId || '' }
  error.value = ''
  modal.value = true
}

async function saveMapping() {
  error.value = ''
  // Keep legacy outstationId aligned with sourceId for DNP3 mappings only;
  // Modbus/TCP and RTU reference their device via sourceId.
  const payload: SignalMapping = { ...form.value }
  payload.outstationId = payload.protocol === 'dnp3' ? (payload.sourceId ?? '') : ''
  try {
    if (editId.value) await api.updateMapping(editId.value, payload)
    else await api.addMapping(payload)
    await store.loadMappings()
    modal.value = false
  } catch (e: unknown) { error.value = e instanceof Error ? e.message : String(e) }
}

async function del(id: string) {
  if (!confirm('¿Eliminar esta señal?')) return
  await api.deleteMapping(id)
  await store.loadMappings()
}

async function importFile(ev: Event) {
  const file = (ev.target as HTMLInputElement).files?.[0]
  if (!file) return
  try {
    const mappings: SignalMapping[] = JSON.parse(await file.text())
    await api.importMappings(mappings)
    await store.loadMappings()
  } catch (e: unknown) {
    alert('Error al importar: ' + (e instanceof Error ? e.message : String(e)))
  }
}
</script>

<style scoped>
.seg { display: inline-flex; border: 1.5px solid var(--border); border-radius: var(--radius); overflow: hidden; background: var(--card); }
.seg-btn { font-family: var(--font-sans); font-weight: 700; font-size: 11px; padding: 7px 16px; color: var(--muted-foreground); background: transparent; border: none; cursor: pointer; transition: background 0.12s, color 0.12s; }
.seg-btn + .seg-btn { border-left: 1.5px solid var(--border); }
.seg-btn--active { background: color-mix(in srgb, var(--epm-bosque) 12%, transparent); color: var(--epm-bosque); }
.proto-tag { font-family: var(--font-mono); font-size: 9.5px; font-weight: 600; letter-spacing: 0.06em; padding: 2px 6px; border-radius: 3px; border: 1px solid; }
.proto-tag--dnp { color: var(--epm-bosque); border-color: color-mix(in srgb, var(--epm-bosque) 40%, transparent); background: color-mix(in srgb, var(--epm-bosque) 8%, transparent); }
.proto-tag--mb  { color: var(--tk-amber-bright); border-color: color-mix(in srgb, var(--tk-amber-base) 45%, transparent); background: color-mix(in srgb, var(--tk-amber-base) 10%, transparent); }
.proto-tag--rtu { color: var(--tk-amber-bright); border-color: color-mix(in srgb, var(--tk-amber-base) 55%, transparent); background: color-mix(in srgb, var(--tk-amber-base) 16%, transparent); letter-spacing: 0.1em; }

/* sticky header sits on the muted band when the body scrolls */
thead th { background: var(--muted); }

/* virtualized rows: EXACT 40px height (required by useVirtualRows); the row
   separator is an inset shadow so it never grows the row box. */
.vrow { height: 40px; }
.vrow > td {
  height: 40px;
  padding-top: 0;
  padding-bottom: 0;
  border-bottom: none;
  box-shadow: inset 0 -1px 0 var(--tk-border-dim);
  white-space: nowrap;
}
.vspacer > td { padding: 0; border: none; box-shadow: none; }
.vspacer:hover > td { background: transparent; }
</style>
