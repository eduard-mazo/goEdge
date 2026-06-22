import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api, type GatewayStatus, type DNP3Outstation, type ModbusDevice, type SerialDevice, type SignalMapping } from '@/api/client'

export const useGatewayStore = defineStore('gateway', () => {
  const status = ref<GatewayStatus | null>(null)
  const outstations = ref<DNP3Outstation[]>([])
  const modbusDevices = ref<ModbusDevice[]>([])
  const serialDevices = ref<SerialDevice[]>([])
  const mappings = ref<SignalMapping[]>([])
  const logs = ref<{ level: string; message: string; time: string }[]>([])
  const ws = ref<WebSocket | null>(null)

  // Clock anchor: the gateway's clock at the last status, paired with the browser
  // clock when it arrived. serverNow() reconstructs the gateway's current time so
  // freshness ages don't depend on the device and browser clocks agreeing (an
  // embedded box can be hours off). Falls back to the browser clock until a
  // status with serverTime is seen.
  const serverTimeMs = ref(0)
  const serverRecvAtMs = ref(0)

  const isRunning = computed(() => status.value?.running ?? false)
  const mqttConnected = computed(() => status.value?.mqttConnected ?? false)

  function noteServerTime(s: GatewayStatus | null) {
    const t = s?.serverTime ? Date.parse(s.serverTime) : NaN
    if (Number.isFinite(t)) {
      serverTimeMs.value = t
      serverRecvAtMs.value = Date.now()
    }
  }

  /** The gateway clock now, extrapolated from the last status via the browser's
   *  elapsed time. Pass a reactive `clientNow` so callers re-render on a ticker. */
  function serverNow(clientNow: number = Date.now()): number {
    if (!serverTimeMs.value) return clientNow
    return serverTimeMs.value + (clientNow - serverRecvAtMs.value)
  }

  async function loadStatus() {
    try {
      status.value = await api.status()
      noteServerTime(status.value)
    } catch {}
  }

  // The `?? []` guards are defensive: the API returns [] for empty lists, but a
  // null (older backend, or a transport hiccup) must never leave these refs null
  // — the list panels read `.length`/`.filter` directly in their templates.
  async function loadOutstations() {
    outstations.value = (await api.getOutstations()) ?? []
  }

  async function loadModbusDevices() {
    modbusDevices.value = (await api.getModbusDevices()) ?? []
  }

  async function loadSerialDevices() {
    serialDevices.value = (await api.getSerialDevices()) ?? []
  }

  async function loadMappings() {
    mappings.value = (await api.getMappings()) ?? []
  }

  async function startGateway() {
    const st = await api.startGateway()
    status.value = st
    noteServerTime(st)
  }

  async function stopGateway() {
    await api.stopGateway()
    await loadStatus()
  }

  function connectWS() {
    if (ws.value) return
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const sock = new WebSocket(`${proto}://${location.host}/ws`)

    sock.onmessage = (ev) => {
      try {
        const evt = JSON.parse(ev.data)
        if (evt.type === 'log') {
          logs.value.unshift(evt.payload)
          if (logs.value.length > 200) logs.value.pop()
        } else if (evt.type === 'status') {
          status.value = evt.payload
          noteServerTime(evt.payload)
        } else if (evt.type === 'readings') {
          if (status.value) {
            status.value.lastReadings = { ...status.value.lastReadings, ...evt.payload }
          }
        }
      } catch {}
    }

    sock.onclose = () => {
      ws.value = null
      setTimeout(connectWS, 3000) // reconnect
    }

    ws.value = sock
  }

  function addLog(level: string, message: string) {
    logs.value.unshift({ level, message, time: new Date().toLocaleTimeString() })
    if (logs.value.length > 200) logs.value.pop()
  }

  return {
    status, outstations, modbusDevices, serialDevices, mappings, logs, isRunning, mqttConnected,
    serverNow,
    loadStatus, loadOutstations, loadModbusDevices, loadSerialDevices, loadMappings,
    startGateway, stopGateway,
    connectWS, addLog,
  }
})
