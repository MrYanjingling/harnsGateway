/* ============================================================
 * 后端契约类型 — 严格对应 Go 结构体的 JSON 输出。
 * 来源:pkg/runtime、pkg/v1、pkg/gateway、pkg/data、pkg/protocol/*。
 * 修改前请先阅读对应 server.go 与 type.go。
 * ============================================================ */

export type DeviceType = 'modbus' | 'opcUa' | 's7'

export type DeviceStatus =
  | 'collecting'
  | 'collectingError'
  | 'unconnected'
  | 'emptyVariable'
  | 'stopped'
  | 'error'

export type AccessMode = 'r' | 'rw'

export type DataType =
  | 'bool'
  | 'int16'
  | 'float32'
  | 'float64'
  | 'int32'
  | 'int64'
  | 'uint16'
  | 'number'
  | 'string'

export type MemoryLayout = 'ABCD' | 'BADC' | 'CDAB' | 'DCBA'

export type DeviceStatusAction = 'start' | 'restart' | 'stop'

/* ---------- 设备(折叠列表形态) ---------- */

/** GET /devices 默认返回的 8 个字段。eTag 用于 If-Match。 */
export interface DeviceSummary {
  name: string
  id: string
  eTag: string
  modTime: string
  deviceCode: string
  deviceType: DeviceType
  deviceModel: string
  collectStatus: DeviceStatus
  topic?: string
}

export interface DeviceListResponse {
  devices?: DeviceSummary[]
}

/* ---------- 协议地址 ---------- */

export interface ModbusAddressOption {
  port?: number
  baudRate?: number
  dataBits?: number
  parity?: string
  stopBits?: string
}

export interface ModbusAddress {
  location: string
  option?: ModbusAddressOption
}

export interface OpcUaAddressOption {
  port?: number
  username?: string
  password?: string
}

export interface OpcUaAddress {
  location: string
  option?: OpcUaAddressOption
}

export interface S7AddressOption {
  port?: number
  rack?: number
  slot?: number
}

export interface S7Address {
  location: string
  option?: S7AddressOption
}

/* ---------- 变量(运行态,exploded=true) ---------- */

export interface ModbusVariable {
  name: string
  dataType: DataType
  address: number
  bits: number
  functionCode: number
  rate: number
  amount: number
  defaultValue?: unknown
  value?: unknown
  accessMode: AccessMode
}

export interface OpcUaVariable {
  name: string
  dataType: DataType
  address: unknown
  namespace: number
  defaultValue?: unknown
  value?: unknown
  accessMode: AccessMode
}

export interface S7Variable {
  name: string
  dataType: DataType
  address: string
  rate: number
  defaultValue?: unknown
  value?: unknown
  accessMode: AccessMode
}

/* ---------- 设备(exploded=true) ---------- */

export interface ModbusDevice extends DeviceSummary {
  deviceType: 'modbus'
  collectorCycle: number
  variableInterval: number
  address: ModbusAddress
  slave: number
  memoryLayout: MemoryLayout
  positionAddress: number
  variables: ModbusVariable[]
}

export interface OpcUaDevice extends DeviceSummary {
  deviceType: 'opcUa'
  collectorCycle: number
  variableInterval: number
  address: OpcUaAddress
  variables: OpcUaVariable[]
}

export interface S7Device extends DeviceSummary {
  deviceType: 's7'
  collectorCycle: number
  variableInterval: number
  address: S7Address
  variables: S7Variable[]
}

export type DeviceDetail = ModbusDevice | OpcUaDevice | S7Device

/* ---------- 创建 / 更新载荷(v1 创建形状) ---------- */

export interface DeviceBasePayload {
  name: string
  deviceCode: string
  deviceType: DeviceType
  deviceModel: string
  collectorCycle: number
  variableInterval?: number
  topic?: string
}

export interface ModbusDevicePayload extends DeviceBasePayload {
  deviceType: 'modbus'
  address: ModbusAddress
  slave: number
  memoryLayout: MemoryLayout
  positionAddress?: number
  variables: Array<{
    name: string
    dataType: DataType
    address: number
    bits?: number
    functionCode: number
    rate?: number
    amount?: number
    defaultValue?: unknown
    accessMode: AccessMode
  }>
}

export interface OpcUaDevicePayload extends DeviceBasePayload {
  deviceType: 'opcUa'
  address: OpcUaAddress
  variables: Array<{
    name: string
    dataType: DataType
    address: unknown
    namespace: number
    defaultValue?: unknown
    accessMode: AccessMode
  }>
}

export interface S7DevicePayload extends DeviceBasePayload {
  deviceType: 's7'
  address: S7Address
  variables: Array<{
    name: string
    dataType: DataType
    address: string
    rate?: number
    defaultValue?: unknown
    accessMode: AccessMode
  }>
}

export type DevicePayload = ModbusDevicePayload | OpcUaDevicePayload | S7DevicePayload

/** PUT /devices/:id/action 的请求体:单键对象数组。 */
export type DeviceAction = Record<string, unknown>

/* ---------- 网关资源 ---------- */

export interface GatewayMeta {
  name: string
  id: string
  eTag: string
  secret: string
  modTime: string
}

/** 注意:后端字段名是首字母大写的 Total/Used/UsedPercent。 */
export interface UsageInfo {
  Total: string
  Used: string
  UsedPercent: string
}

export interface GatewayCpuResponse {
  cpus?: {
    '1min': string
    '5min': string
    '15min': string
  }
}

export interface GatewayMemResponse {
  mem?: UsageInfo
}

export interface GatewayDiskResponse {
  disk?: Record<string, UsageInfo>
}

/* ---------- 数据服务统一响应 ---------- */

export interface DataResponse<T> {
  data?: T
  total?: number
}

export interface MetricsInfo {
  id: string
  tenantId?: string | null
  revision?: number | null
  createdBy?: string | null
  createdTime?: string | null
  updatedBy?: string | null
  updatedTime?: string | null
  dataPoint?: string | null
  metricsCode?: string | null
  metricsName?: string | null
  measureType?: string | null
  enable?: boolean | null
  metricsType?: string | null
  calculateType?: string | null
  note?: string | null
}

export interface KpiInfo {
  id: string
  tenantId?: string | null
  revision?: number | null
  createdBy?: string | null
  createdTime?: string | null
  updatedBy?: string | null
  updatedTime?: string | null
  name: string
  formula?: string | null
  formulaDataName?: string | null
  formulaOri: string
  timeGranularity: string
  targetData?: string | null
  singleQuota: boolean
  defaultTag: boolean
  canDelete: boolean
  unit?: string | null
  targetDataType: string
  status: string
}

export interface TagInfo {
  id: string
  tenantId?: string | null
  revision?: number | null
  createdBy?: string | null
  createdTime?: string | null
  updatedBy?: string | null
  updatedTime?: string | null
  tagId?: string | null
  tagName?: string | null
  tagDesc?: string | null
  lastSynTime?: string | null
  tagType?: string | null
  tagAttribute?: string | null
  enable?: number | null
  scadaId: string
  operator?: string | null
  operateValue?: string | null
}

export interface MetricsQueryRequest {
  metricsCodes: string[]
  startTime: string
  endTime: string
  granularity: 'h'
}

export interface MetricsResult {
  metricsCode: string
  startTime: string
  endTime: string
  value: number
}

export interface KpiQueryRequest {
  ids: string[]
  startTime: string
  endTime: string
  granularity: 'h'
}

export interface KpiResult {
  id: string
  startTime: string
  endTime: string
  value: number
}

export type TimeSeriesMode = 'raw' | 'mean' | 'usage'

export interface TagTimeSeriesRequest {
  select: string
  start?: string
  end?: string
  latest?: boolean
  limit?: number
  mode?: TimeSeriesMode
}

export interface TimeSeriesPoint {
  time: string
  value: string
}

export interface DataPointTimeSeries {
  tagId: string
  tagName: string
  tagDesc: string
  data: TimeSeriesPoint[]
}

export type TimeSeriesSource = 'fmcs' | 'escada'
