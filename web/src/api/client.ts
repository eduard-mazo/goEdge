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

  getModbusDevices:    ()  => req<ModbusDevice[]>('GET', '/modbusDevices'),
  addModbusDevice:     (d: ModbusDevice) => req<ModbusDevice>('POST', '/modbusDevices', d),
  updateModbusDevice:  (id: string, d: ModbusDevice) => req<ModbusDevice>('PUT', `/modbusDevices/${id}`, d),
  deleteModbusDevice:  (id: string) => req<null>('DELETE', `/modbusDevices/${id}`),

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

export type Protocol = 'dnp3' | 'modbus'

export type ModbusFunction =
  | 'coil'
  | 'discrete_input'
  | 'input_register'
  | 'holding_register'

export type ModbusDataType =
  | 'bool'
  | 'int16'
  | 'uint16'
  | 'int32'
  | 'uint32'
  | 'float32'
  | 'float64'

export interface ModbusDevice {
  id: string
  label: string
  host: string
  port: number               // default 502
  unitId: number             // Modbus slave/unit id
  scanRateMs?: number        // poll cadence; default 1000
  timeoutMs?: number         // default 3000
  retries?: number           // default 2
  retryDelayMs?: number      // default 500
  enabled: boolean
}

export interface SignalMapping {
  id: string
  metricName: string
  deviceId?: string          // Sparkplug device ID; empty = node metric

  protocol?: Protocol        // 'dnp3' (default) | 'modbus'
  sourceId?: string          // ref to source (outstation/device); falls back to outstationId
  outstationId: string       // legacy DNP3 alias of sourceId

  // DNP3 point identity
  pointType: PointType
  index: number
  eventClass: number

  // Modbus point identity
  function?: ModbusFunction
  address?: number
  quantity?: number
  dataType?: ModbusDataType
  byteOrder?: string         // ABCD|DCBA|BADC|CDAB

  scale?: number
  offset?: number
  engineeringUnit?: string

  deadband?: number
  publishOnPoll?: boolean

  enabled: boolean
}

export interface AppConfig {
  mqtt: MQTTConfig
  sparkplug: SparkplugConfig
  outstations: DNP3Outstation[]
  modbusDevices: ModbusDevice[]
  mappings: SignalMapping[]
}

export interface GatewayStatus {
  running: boolean
  mqttConnected: boolean
  bdSeq: number
  outstations: Record<string, OutstationStatus>
  publishCount: number
  errorCount: number
  droppedCount?: number
  uptime: string
  lastReadings: Record<string, number>
}

// OutstationStatus is the per-source status snapshot (DNP3 outstation or Modbus
// device — the gateway reports both under GatewayStatus.outstations).
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
