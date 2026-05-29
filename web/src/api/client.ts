const BASE = '/api'

export interface APIResponse<T = unknown> {
  success: boolean
  message?: string
  data?: T
}

async function req<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch(BASE + path, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  })
  const json: APIResponse<T> = await res.json()
  if (!json.success) throw new Error(json.message ?? 'Request failed')
  return json.data as T
}

export const api = {
  status:      ()      => req<GatewayStatus>('GET',  '/status'),
  config:      ()      => req<AppConfig>('GET',  '/config'),

  getMQTT:     ()      => req<MQTTConfig>('GET',  '/config/mqtt'),
  setMQTT:     (c: MQTTConfig) => req<MQTTConfig>('PUT', '/config/mqtt', c),

  getSparkplug: ()     => req<SparkplugConfig>('GET',  '/config/sparkplug'),
  setSparkplug: (c: SparkplugConfig) => req<SparkplugConfig>('PUT', '/config/sparkplug', c),

  getDevices:  ()      => req<ModbusDevice[]>('GET',  '/devices'),
  addDevice:   (d: ModbusDevice) => req<ModbusDevice>('POST', '/devices', d),
  updateDevice: (id: string, d: ModbusDevice) => req<ModbusDevice>('PUT', `/devices/${id}`, d),
  deleteDevice: (id: string) => req<null>('DELETE', `/devices/${id}`),

  getMappings: ()      => req<SignalMapping[]>('GET', '/mappings'),
  addMapping:  (m: SignalMapping) => req<SignalMapping>('POST', '/mappings', m),
  updateMapping: (id: string, m: SignalMapping) => req<SignalMapping>('PUT', `/mappings/${id}`, m),
  deleteMapping: (id: string) => req<null>('DELETE', `/mappings/${id}`),
  importMappings: (ms: SignalMapping[]) => req<{imported: number}>('POST', '/mappings/import', ms),
  exportMappingsURL: () => BASE + '/mappings/export',

  startGateway: () => req<GatewayStatus>('POST', '/gateway/start'),
  stopGateway:  () => req<null>('POST', '/gateway/stop'),
}

// --- Domain types ---

export interface MQTTConfig {
  broker: string
  clientId: string
  username?: string
  password?: string
  qos: number
  keepalive: number
  tls: TLSConfig
}

export interface TLSConfig {
  enabled: boolean
  caFile?: string
  certFile?: string
  keyFile?: string
  insecure?: boolean
}

export interface SparkplugConfig {
  groupId: string
  nodeId: string
  birthOnConfigChange?: boolean
}

export interface ModbusDevice {
  id: string
  label: string
  host: string
  port: number
  timeoutMs?: number
  retries?: number
  retryDelayMs?: number
  enabled: boolean
}

export interface SignalMapping {
  id: string
  metricName: string
  deviceId?: string
  modbusDeviceId: string
  unitId: number
  function: 'coil' | 'discrete_input' | 'input_register' | 'holding_register'
  address: number
  quantity: number
  dataType: string
  byteOrder?: string
  scale?: number
  offset?: number
  engineeringUnit?: string
  scanRateMs: number
  deadband?: number
  qualityPolicy?: string
  enabled: boolean
}

export interface AppConfig {
  mqtt: MQTTConfig
  sparkplug: SparkplugConfig
  devices: ModbusDevice[]
  mappings: SignalMapping[]
}

export interface GatewayStatus {
  running: boolean
  mqttConnected: boolean
  bdSeq: number
  devices: Record<string, DeviceStatus>
  publishCount: number
  errorCount: number
  uptime: string
  lastReadings: Record<string, number>
}

export interface DeviceStatus {
  id: string
  label: string
  addr: string
  connected: boolean
  errors: number
  reads: number
}
