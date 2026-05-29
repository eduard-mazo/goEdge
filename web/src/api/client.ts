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

  getOutstations:    ()  => req<DNP3Outstation[]>('GET', '/outstations'),
  addOutstation:     (o: DNP3Outstation) => req<DNP3Outstation>('POST', '/outstations', o),
  updateOutstation:  (id: string, o: DNP3Outstation) => req<DNP3Outstation>('PUT', `/outstations/${id}`, o),
  deleteOutstation:  (id: string) => req<null>('DELETE', `/outstations/${id}`),

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

export interface DNP3Outstation {
  id: string
  label: string
  host: string
  port: number                   // DNP3/IP default 20000
  masterAddress: number          // local link-layer address (typical 1)
  outstationAddress: number      // remote link-layer address (typical 1024+)
  responseTimeoutMs?: number     // app-layer response timeout; default 5000
  keepAliveMs?: number           // keep-alive interval; default 60000

  integrityScanMs?: number       // 0 = disabled; default 3600000
  class1ScanMs?: number          // 0 = disabled; default 1000
  class2ScanMs?: number          // 0 = disabled; default 5000
  class3ScanMs?: number          // 0 = disabled; default 30000

  unsolicitedEnabled?: boolean
  unsolicitedClass1?: boolean
  unsolicitedClass2?: boolean
  unsolicitedClass3?: boolean

  disableUnsolOnStartup?: boolean
  startupIntegrity?: boolean

  enabled: boolean
}

export type PointType =
  | 'binary'
  | 'double_bit_binary'
  | 'binary_output_status'
  | 'counter'
  | 'frozen_counter'
  | 'analog'
  | 'analog_output_status'
  | 'octet_string'

export interface SignalMapping {
  id: string
  metricName: string
  deviceId?: string          // Sparkplug device ID; empty = node metric
  outstationId: string       // ref to DNP3Outstation.id

  pointType: PointType
  index: number              // point index within its type
  eventClass: number         // 0=static, 1|2|3 = event class assignment (informational)

  scale?: number             // value × scale + offset
  offset?: number
  engineeringUnit?: string

  deadband?: number          // 0 = always publish on event
  publishOnPoll?: boolean    // also publish static reads

  enabled: boolean
}

export interface AppConfig {
  mqtt: MQTTConfig
  sparkplug: SparkplugConfig
  outstations: DNP3Outstation[]
  mappings: SignalMapping[]
}

export interface GatewayStatus {
  running: boolean
  mqttConnected: boolean
  bdSeq: number
  outstations: Record<string, OutstationStatus>
  publishCount: number
  errorCount: number
  uptime: string
  lastReadings: Record<string, number>
}

export interface OutstationStatus {
  id: string
  label: string
  addr: string
  connected: boolean
  lastError?: string
  measurementsRx?: number
  integrityPolls?: number
  classPolls?: number
  unsolicitedRsps?: number
  lastReadAt?: string
}
