import type { DeviceStatus, DeviceStatusAction, DeviceType, DataType } from '@/api/types'

export interface StatusMeta {
  label: string
  /** tailwind 类:文本 + 背景 + 边框 */
  className: string
  dotClassName: string
  description: string
}

export const DEVICE_STATUS: Record<DeviceStatus, StatusMeta> = {
  collecting: {
    label: '采集中',
    className: 'bg-emerald-500/10 text-emerald-700 border-emerald-600/20 dark:text-emerald-300',
    dotClassName: 'bg-emerald-500',
    description: '设备连接正常,正在采集数据',
  },
  collectingError: {
    label: '采集异常',
    className: 'bg-amber-500/10 text-amber-700 border-amber-600/20 dark:text-amber-300',
    dotClassName: 'bg-amber-500',
    description: '设备可连接,但部分数据采集失败',
  },
  unconnected: {
    label: '未连接',
    className: 'bg-zinc-500/10 text-zinc-600 border-zinc-500/20 dark:text-zinc-300',
    dotClassName: 'bg-zinc-400',
    description: '网关无法连接到设备',
  },
  emptyVariable: {
    label: '无变量',
    className: 'bg-violet-500/10 text-violet-700 border-violet-600/20 dark:text-violet-300',
    dotClassName: 'bg-violet-500',
    description: '设备没有配置任何采集变量',
  },
  stopped: {
    label: '已停止',
    className: 'bg-slate-500/10 text-slate-600 border-slate-500/20 dark:text-slate-300',
    dotClassName: 'bg-slate-400',
    description: '采集已停止',
  },
  error: {
    label: '错误',
    className: 'bg-rose-500/10 text-rose-700 border-rose-600/20 dark:text-rose-300',
    dotClassName: 'bg-rose-500',
    description: '设备运行发生错误',
  },
}

export const UNKNOWN_STATUS: StatusMeta = {
  label: '未知',
  className: 'bg-zinc-500/10 text-zinc-600 border-zinc-500/20 dark:text-zinc-300',
  dotClassName: 'bg-zinc-400',
  description: '未知状态',
}

export interface ProtocolMeta {
  type: DeviceType
  label: string
  short: string
  description: string
  defaultModel: string
  models: Array<{ value: string; label: string; hint: string }>
  /** 用于徽标配色 */
  accent: string
}

export const PROTOCOLS: Record<DeviceType, ProtocolMeta> = {
  modbus: {
    type: 'modbus',
    label: 'Modbus',
    short: 'MB',
    description: 'Modbus TCP / RTU / RTU over TCP',
    defaultModel: 'modbusTcp',
    models: [
      { value: 'modbusTcp', label: 'Modbus TCP', hint: '以太网,默认端口 502' },
      { value: 'modbusRtu', label: 'Modbus RTU', hint: '串口 (RS-485/232)' },
      { value: 'modbusRtuOverTcp', label: 'Modbus RTU over TCP', hint: '串口协议经 TCP 透传' },
    ],
    accent: 'emerald',
  },
  opcUa: {
    type: 'opcUa',
    label: 'OPC UA',
    short: 'UA',
    description: 'OPC UA 二进制协议',
    defaultModel: 'opcUa',
    models: [{ value: 'opcUa', label: 'OPC UA', hint: 'Endpoint: opc.tcp://host:port' }],
    accent: 'sky',
  },
  s7: {
    type: 's7',
    label: 'S7',
    short: 'S7',
    description: '西门子 S7-1500',
    defaultModel: 's71500',
    models: [{ value: 's71500', label: 'S7-1500', hint: 'ISO-on-TCP,默认端口 102' }],
    accent: 'amber',
  },
}

export const DEVICE_STATUS_ACTIONS: Array<{
  action: DeviceStatusAction
  label: string
}> = [
  { action: 'start', label: '启动' },
  { action: 'restart', label: '重启' },
  { action: 'stop', label: '停止' },
]

export const DATA_TYPES: Array<{ value: DataType; label: string }> = [
  { value: 'bool', label: 'bool' },
  { value: 'int16', label: 'int16' },
  { value: 'uint16', label: 'uint16' },
  { value: 'int32', label: 'int32' },
  { value: 'int64', label: 'int64' },
  { value: 'float32', label: 'float32' },
  { value: 'float64', label: 'float64' },
  { value: 'number', label: 'number' },
  { value: 'string', label: 'string' },
]

export const MEMORY_LAYOUTS = ['ABCD', 'BADC', 'CDAB', 'DCBA'] as const

export const BAUD_RATES = [1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200]
export const DATA_BITS = [5, 6, 7, 8]
export const PARITIES = [
  { value: 'noParity', label: '无校验 (None)' },
  { value: 'oddParity', label: '奇校验 (Odd)' },
  { value: 'evenParity', label: '偶校验 (Even)' },
]
export const STOP_BITS = [
  { value: '1', label: '1' },
  { value: '1.5', label: '1.5' },
  { value: '2', label: '2' },
]
