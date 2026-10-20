import type {
  DataPointTimeSeries,
  KpiInfo,
  MetricsInfo,
  TagInfo,
} from '../src/api/types'

/** mock 内存态设备:比 DeviceDetail 宽松,便于增删改。 */
export interface StoredDevice {
  name: string
  id: string
  eTag: string
  modTime: string
  deviceCode: string
  deviceType: 'modbus' | 'opcUa' | 's7'
  deviceModel: string
  collectStatus: 'collecting' | 'collectingError' | 'unconnected' | 'emptyVariable' | 'stopped' | 'error'
  topic?: string
  collectorCycle?: number
  variableInterval?: number
  address?: Record<string, unknown>
  slave?: number
  memoryLayout?: string
  positionAddress?: number
  variables?: Array<{
    name: string
    dataType: string
    address: unknown
    bits?: number
    functionCode?: number
    rate?: number
    amount?: number
    namespace?: number
    defaultValue?: unknown
    value?: unknown
    accessMode: 'r' | 'rw'
  }>
}

const now = () => new Date().toISOString()

export const gatewayMeta = {
  name: 'harnsgateway-edge-01',
  id: 'f9f7d42d40b2415d90ed79a242292337',
  eTag: '1844674407370955161',
  secret: 'hg_5e1c9a77b2f0441db8c3e6a90d2f7c41',
  modTime: '2026-08-12T03:21:58.000Z',
}

export const seedDevices: StoredDevice[] = [
  {
    name: '1# 空压机',
    id: 'a1b2c3d4e5f60718293a4b5c6d7e8f90',
    eTag: '1001',
    modTime: '2026-10-19T08:12:31.000Z',
    deviceCode: 'air-compressor-01',
    deviceType: 'modbus',
    deviceModel: 'modbusTcp',
    collectStatus: 'collecting',
    collectorCycle: 1,
    variableInterval: 10,
    slave: 1,
    memoryLayout: 'ABCD',
    positionAddress: 0,
    address: { location: '192.168.10.21', option: { port: 502 } },
    variables: [
      { name: '运行状态', dataType: 'bool', address: 0, bits: 0, functionCode: 3, accessMode: 'r', value: true },
      { name: '排气压力', dataType: 'float32', address: 2, functionCode: 3, rate: 0.1, accessMode: 'r', value: 0.72 },
      { name: '排气温度', dataType: 'int16', address: 4, functionCode: 3, accessMode: 'r', value: 78 },
      { name: '频率给定', dataType: 'uint16', address: 6, functionCode: 3, accessMode: 'rw', value: 42 },
      { name: '启停控制', dataType: 'bool', address: 8, bits: 0, functionCode: 3, accessMode: 'rw', value: true },
    ],
  },
  {
    name: '2# 冷却泵',
    id: 'b2c3d4e5f60718293a4b5c6d7e8f9012',
    eTag: '1002',
    modTime: '2026-10-19T07:48:05.000Z',
    deviceCode: 'cooling-pump-02',
    deviceType: 'modbus',
    deviceModel: 'modbusRtu',
    collectStatus: 'collectingError',
    collectorCycle: 2,
    slave: 2,
    memoryLayout: 'CDAB',
    positionAddress: 0,
    address: {
      location: '/dev/ttyS0',
      option: { baudRate: 9600, dataBits: 8, parity: 'noParity', stopBits: '1' },
    },
    variables: [
      { name: '出口压力', dataType: 'float32', address: 0, functionCode: 3, accessMode: 'r', value: 0.41 },
      { name: '电机电流', dataType: 'int16', address: 2, functionCode: 3, rate: 0.01, accessMode: 'r' },
      { name: '运行频率', dataType: 'uint16', address: 4, functionCode: 3, accessMode: 'rw', value: 35 },
    ],
  },
  {
    name: '3# 注塑机',
    id: 'c3d4e5f60718293a4b5c6d7e8f901234',
    eTag: '1003',
    modTime: '2026-10-19T09:03:44.000Z',
    deviceCode: 'injection-molding-03',
    deviceType: 'opcUa',
    deviceModel: 'opcUa',
    collectStatus: 'collecting',
    collectorCycle: 1,
    variableInterval: 5,
    address: { location: 'opc.tcp://192.168.10.35:4840', option: { port: 4840, username: 'operator', password: '••••••' } },
    variables: [
      { name: '料筒温度', dataType: 'number', address: 'ns=2;s=Barrel.Temperature', namespace: 2, accessMode: 'r', value: 216.4 },
      { name: '注射压力', dataType: 'number', address: 'ns=2;s=Injection.Pressure', namespace: 2, accessMode: 'r', value: 88.2 },
      { name: '设定温度', dataType: 'number', address: 'ns=2;s=Barrel.TargetTemperature', namespace: 2, accessMode: 'rw', value: 220 },
      { name: '模号', dataType: 'string', address: 'ns=2;s=Mold.Code', namespace: 2, accessMode: 'r', value: 'M-2207A' },
    ],
  },
  {
    name: '4# 蒸汽锅炉',
    id: 'd4e5f60718293a4b5c6d7e8f90123456',
    eTag: '1004',
    modTime: '2026-10-18T22:15:10.000Z',
    deviceCode: 'steam-boiler-04',
    deviceType: 's7',
    deviceModel: 's71500',
    collectStatus: 'stopped',
    collectorCycle: 3,
    address: { location: '192.168.10.40', option: { port: 102, rack: 0, slot: 1 } },
    variables: [
      { name: '炉膛温度', dataType: 'float32', address: 'DB1.DBD0', rate: 1, accessMode: 'r', value: 642.5 },
      { name: '蒸汽压力', dataType: 'float32', address: 'DB1.DBD4', rate: 0.01, accessMode: 'r', value: 1.26 },
      { name: '给水阀开度', dataType: 'int16', address: 'DB1.DBW8', rate: 0.1, accessMode: 'rw', value: 55 },
    ],
  },
  {
    name: '5# 传送带',
    id: 'e5f60718293a4b5c6d7e8f9012345678',
    eTag: '1005',
    modTime: '2026-10-19T06:30:00.000Z',
    deviceCode: 'conveyor-05',
    deviceType: 'modbus',
    deviceModel: 'modbusRtuOverTcp',
    collectStatus: 'unconnected',
    collectorCycle: 2,
    slave: 5,
    memoryLayout: 'ABCD',
    address: { location: '192.168.10.51', option: { port: 8899 } },
    variables: [
      { name: '线速度', dataType: 'float32', address: 0, functionCode: 3, accessMode: 'r' },
      { name: '启停', dataType: 'bool', address: 2, bits: 0, functionCode: 3, accessMode: 'rw', value: false },
    ],
  },
  {
    name: '6# 包装线',
    id: 'f60718293a4b5c6d7e8f901234567890',
    eTag: '1006',
    modTime: '2026-10-17T14:02:19.000Z',
    deviceCode: 'packaging-line-06',
    deviceType: 'opcUa',
    deviceModel: 'opcUa',
    collectStatus: 'error',
    collectorCycle: 1,
    address: { location: 'opc.tcp://192.168.10.60:4840', option: { port: 4840 } },
    variables: [
      { name: '包装速度', dataType: 'number', address: 'ns=3;s=Line.Speed', namespace: 3, accessMode: 'r' },
      { name: '产量计数', dataType: 'number', address: 'ns=3;s=Line.Counter', namespace: 3, accessMode: 'r' },
    ],
  },
]

export const metrics: MetricsInfo[] = [
  {
    id: 'm-001',
    metricsCode: 'energy_total',
    metricsName: '总能耗',
    measureType: 'kWh',
    enable: true,
    metricsType: 'A',
    calculateType: 'sum',
    dataPoint: 'plant.energy',
    note: '全厂总用电量',
  },
  {
    id: 'm-002',
    metricsCode: 'power_factor',
    metricsName: '功率因数',
    measureType: '',
    enable: true,
    metricsType: 'A',
    calculateType: 'avg',
    dataPoint: 'plant.power-factor',
    note: '并网点平均功率因数',
  },
  {
    id: 'm-003',
    metricsCode: 'water_total',
    metricsName: '总用水量',
    measureType: 'm³',
    enable: true,
    metricsType: 'A',
    calculateType: 'sum',
    dataPoint: 'plant.water',
    note: '',
  },
  {
    id: 'm-004',
    metricsCode: 'steam_total',
    metricsName: '总蒸汽量',
    measureType: 't',
    enable: true,
    metricsType: 'A',
    calculateType: 'sum',
    dataPoint: 'plant.steam',
    note: '',
  },
  {
    id: 'm-005',
    metricsCode: 'run_hours',
    metricsName: '设备运行时长',
    measureType: 'h',
    enable: false,
    metricsType: 'M',
    calculateType: 'sum',
    dataPoint: 'plant.runtime',
    note: '手动抄表',
  },
]

export const kpis: KpiInfo[] = [
  {
    id: 'kpi_001',
    name: '日总能耗',
    formula: 'energy_total',
    formulaOri: 'A',
    timeGranularity: 'd',
    singleQuota: true,
    defaultTag: false,
    canDelete: true,
    unit: 'kWh',
    targetDataType: 'number',
    status: 'active',
    targetData: '52000',
  },
  {
    id: 'kpi_002',
    name: '单位产品能耗',
    formula: 'energy_total/output_count',
    formulaOri: 'A/B',
    timeGranularity: 'd',
    singleQuota: false,
    defaultTag: false,
    canDelete: true,
    unit: 'kWh/件',
    targetDataType: 'number',
    status: 'active',
    targetData: '1.8',
  },
  {
    id: 'kpi_003',
    name: '设备综合效率 (OEE)',
    formula: 'availability*performance*quality',
    formulaOri: 'A*B*C',
    timeGranularity: 'd',
    singleQuota: false,
    defaultTag: true,
    canDelete: false,
    unit: '%',
    targetDataType: 'number',
    status: 'active',
    targetData: '85',
  },
  {
    id: 'kpi_004',
    name: '峰谷电费占比',
    formula: 'peak_cost/total_cost',
    formulaOri: 'A/B',
    timeGranularity: 'm',
    singleQuota: false,
    defaultTag: false,
    canDelete: true,
    unit: '%',
    targetDataType: 'number',
    status: 'draft',
  },
]

export const tags: TagInfo[] = [
  { id: 't-001', tagId: 'SCADA.compressor.pressure', tagName: '空压机排气压力', tagDesc: '1# 空压机', tagType: 'analog', tagAttribute: 'float', enable: 1, scadaId: 'scada-01', lastSynTime: '2026-10-19T09:20:00.000Z' },
  { id: 't-002', tagId: 'SCADA.compressor.temperature', tagName: '空压机排气温度', tagDesc: '1# 空压机', tagType: 'analog', tagAttribute: 'float', enable: 1, scadaId: 'scada-01', lastSynTime: '2026-10-19T09:20:00.000Z' },
  { id: 't-003', tagId: 'SCADA.chiller.flow', tagName: '冷却水流量', tagDesc: '循环水系统', tagType: 'analog', tagAttribute: 'float', enable: 1, scadaId: 'scada-01', lastSynTime: '2026-10-19T09:20:00.000Z' },
  { id: 't-004', tagId: 'SCADA.boiler.steam', tagName: '蒸汽流量', tagDesc: '4# 锅炉', tagType: 'analog', tagAttribute: 'float', enable: 1, scadaId: 'scada-02', lastSynTime: '2026-10-19T09:19:30.000Z' },
  { id: 't-005', tagId: 'SCADA.boiler.level', tagName: '汽包液位', tagDesc: '4# 锅炉', tagType: 'analog', tagAttribute: 'float', enable: 1, scadaId: 'scada-02', lastSynTime: '2026-10-19T09:19:30.000Z' },
  { id: 't-006', tagId: 'SCADA.molding.pressure', tagName: '注塑压力', tagDesc: '3# 注塑机', tagType: 'analog', tagAttribute: 'float', enable: 1, scadaId: 'scada-02', lastSynTime: '2026-10-19T09:18:00.000Z' },
  { id: 't-007', tagId: 'SCADA.molding.status', tagName: '注塑机运行状态', tagDesc: '3# 注塑机', tagType: 'digital', tagAttribute: 'bool', enable: 1, scadaId: 'scada-02', lastSynTime: '2026-10-19T09:18:00.000Z' },
  { id: 't-008', tagId: 'SCADA.legacy.temp', tagName: '旧线温度', tagDesc: '已停用产线', tagType: 'analog', tagAttribute: 'float', enable: 0, scadaId: 'scada-01', lastSynTime: '2026-09-30T11:02:00.000Z' },
]

const TAG_BASE: Record<string, { name: string; desc: string; base: number; amp: number }> = {
  'SCADA.compressor.pressure': { name: '空压机排气压力', desc: '1# 空压机', base: 0.72, amp: 0.06 },
  'SCADA.compressor.temperature': { name: '空压机排气温度', desc: '1# 空压机', base: 78, amp: 4 },
  'SCADA.chiller.flow': { name: '冷却水流量', desc: '循环水系统', base: 126, amp: 9 },
  'SCADA.boiler.steam': { name: '蒸汽流量', desc: '4# 锅炉', base: 12.4, amp: 1.1 },
  'SCADA.boiler.level': { name: '汽包液位', desc: '4# 锅炉', base: 52, amp: 3 },
  'SCADA.molding.pressure': { name: '注塑压力', desc: '3# 注塑机', base: 88, amp: 6 },
  'SCADA.molding.status': { name: '注塑机运行状态', desc: '3# 注塑机', base: 1, amp: 0 },
  'SCADA.legacy.temp': { name: '旧线温度', desc: '已停用产线', base: 30, amp: 0.2 },
}

const pad = (n: number) => String(n).padStart(2, '0')
function formatTime(d: Date) {
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

/** 生成最近 24 小时的确定性时序数据。 */
export function buildTimeSeries(select: string, limit = 48): DataPointTimeSeries[] {
  const ids = select
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean)
    .slice(0, 10)
  const nowDate = new Date()
  const count = Math.max(2, Math.min(limit || 48, 48))
  const stepMs = (24 * 60 * 60 * 1000) / (count - 1)

  return ids.map((tagId) => {
    const meta = TAG_BASE[tagId] ?? { name: tagId, desc: '', base: 50, amp: 5 }
    return {
      tagId,
      tagName: meta.name,
      tagDesc: meta.desc,
      data: Array.from({ length: count }, (_, i) => {
        const time = new Date(nowDate.getTime() - (count - 1 - i) * stepMs)
        const wave = Math.sin(i / 3.2) * meta.amp + Math.sin(i / 1.7) * meta.amp * 0.35
        const value = meta.amp === 0 ? meta.base : meta.base + wave
        return {
          time: formatTime(time),
          value: value.toFixed(meta.base >= 10 ? 2 : 3),
        }
      }),
    }
  })
}

export function createDeviceId(): string {
  const hex = '0123456789abcdef'
  return Array.from({ length: 32 }, () => hex[Math.floor(Math.random() * 16)]).join('')
}

export function nowIso() {
  return now()
}
