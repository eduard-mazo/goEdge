import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api, type GatewayStatus, type DNP3Outstation, type ModbusDevice, type SignalMapping } from '@/api/client'

export const useGatewayStore = defineStore('gateway', () => {
  const status = ref<GatewayStatus | null>(null)
  const outstations = ref<DNP3Outstation[]>([])
  const modbusDevices = ref<ModbusDevice[]>([])
  const mappings = ref<SignalMapping[]>([])
  const logs = ref<{ level: string; message: string; time: string }[]>([])
  const ws = ref<WebSocket | null>(null)

  const isRunning = computed(() => status.value?.running ?? false)
  const mqttConnected = computed(() => status.value?.mqttConnected ?? false)

  async function loadStatus() {
    try {
      status.value = await api.status()
    } catch {}
  }

  async function loadOutstations() {
    outstations.value = await api.getOutstations()
  }

  async function loadModbusDevices() {
    modbusDevices.value = await api.getModbusDevices()
  }

  async function loadMappings() {
    mappings.value = await api.getMappings()
  }

  async function startGateway() {
    const st = await api.startGateway()
    status.value = st
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
    status, outstations, modbusDevices, mappings, logs, isRunning, mqttConnected,
    loadStatus, loadOutstations, loadModbusDevices, loadMappings,
    startGateway, stopGateway,
    connectWS, addLog,
  }
})
