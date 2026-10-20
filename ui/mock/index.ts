import type { IncomingMessage, ServerResponse } from 'node:http'
import type { Plugin } from 'vite'
import {
  buildTimeSeries,
  createDeviceId,
  gatewayMeta,
  kpis,
  metrics,
  nowIso,
  seedDevices,
  tags,
  type StoredDevice,
} from './fixtures'

/**
 * 仅开发期生效的 mock API。
 * 与真实后端保持同一套 /api/v1 路径与 JSON 形状,
 * 让前端可以在 Go 服务未启动时完整演示。
 */

const PREFIX = '/api/v1'
const LATENCY_MS = 120

let devices: StoredDevice[] = structuredClone(seedDevices)

function send(
  res: ServerResponse,
  status: number,
  body?: unknown,
  headers?: Record<string, string>,
) {
  if (headers) {
    for (const [key, value] of Object.entries(headers)) res.setHeader(key, value)
  }
  if (body === undefined) {
    res.statusCode = status
    res.end()
    return
  }
  res.statusCode = status
  res.setHeader('Content-Type', 'application/json; charset=utf-8')
  res.end(JSON.stringify(body))
}

function sendErrors(res: ServerResponse, status: number, messages: string[], code = 10001) {
  send(res, status, { errors: messages.map((message) => ({ code, message })) })
}

function readJson(req: IncomingMessage): Promise<unknown> {
  return new Promise((resolve, reject) => {
    const chunks: Buffer[] = []
    req.on('data', (chunk: Buffer) => chunks.push(chunk))
    req.on('end', () => {
      const raw = Buffer.concat(chunks).toString('utf8')
      if (!raw) return resolve(undefined)
      try {
        resolve(JSON.parse(raw))
      } catch {
        reject(new Error('malformed json'))
      }
    })
    req.on('error', reject)
  })
}

function toSummary(device: StoredDevice) {
  return {
    name: device.name,
    id: device.id,
    eTag: device.eTag,
    modTime: device.modTime,
    deviceCode: device.deviceCode,
    deviceType: device.deviceType,
    deviceModel: device.deviceModel,
    collectStatus: device.collectStatus,
    ...(device.topic ? { topic: device.topic } : {}),
  }
}

function bump(device: StoredDevice) {
  const current = Number.parseInt(device.eTag, 10)
  device.eTag = String(Number.isFinite(current) ? current + 1 : Date.now())
  device.modTime = nowIso()
}

function checkPrecondition(
  req: IncomingMessage,
  res: ServerResponse,
  device: StoredDevice,
): boolean {
  const ifMatch = req.headers['if-match']
  if (!ifMatch) {
    send(res, 428)
    return false
  }
  if (ifMatch !== device.eTag) {
    send(res, 412)
    return false
  }
  return true
}

function baseValue(code: string): number {
  switch (code) {
    case 'energy_total':
      return 48213.5
    case 'power_factor':
      return 0.93
    case 'water_total':
      return 1862.4
    case 'steam_total':
      return 342.8
    case 'run_hours':
      return 720
    default:
      return 0
  }
}

function kpiValue(id: string): number {
  switch (id) {
    case 'kpi_001':
      return 48213.5
    case 'kpi_002':
      return 1.72
    case 'kpi_003':
      return 86.4
    case 'kpi_004':
      return 37.8
    default:
      return 0
  }
}

function hourRange(body: Record<string, unknown>, keys: ['startTime', 'endTime'] | ['start', 'end']) {
  return {
    start: typeof body[keys[0]] === 'string' ? (body[keys[0]] as string) : '',
    end: typeof body[keys[1]] === 'string' ? (body[keys[1]] as string) : '',
  }
}

async function handle(req: IncomingMessage, res: ServerResponse): Promise<boolean> {
  const method = (req.method ?? 'GET').toUpperCase()
  const url = new URL(req.url ?? '/', 'http://mock.local')
  const path = url.pathname

  if (!path.startsWith(PREFIX)) return false

  const pathname = path.slice(PREFIX.length) || '/'
  const query = url.searchParams

  /* ---------------- 网关 ---------------- */

  if (method === 'GET' && pathname === '/gatewayMeta') {
    send(res, 200, gatewayMeta, { ETag: gatewayMeta.eTag })
    return true
  }

  if (method === 'GET' && pathname === '/gatewayCpu') {
    const t = new Date().getSeconds()
    const base = 18 + Math.sin(t / 9) * 6
    send(res, 200, {
      cpus: {
        '1min': `${base.toFixed(0)}%`,
        '5min': `${(base - 2).toFixed(0)}%`,
        '15min': `${(base - 4).toFixed(0)}%`,
      },
    })
    return true
  }

  if (method === 'GET' && pathname === '/gatewayMem') {
    const usedPercent = 54 + Math.sin(new Date().getSeconds() / 12) * 3
    send(res, 200, {
      mem: {
        Total: '3906MB',
        Used: `${Math.round((3906 * usedPercent) / 100)}MB`,
        UsedPercent: `${usedPercent.toFixed(0)}%`,
      },
    })
    return true
  }

  if (method === 'GET' && pathname === '/gatewayDisk') {
    send(res, 200, {
      disk: {
        '/': { Total: '29072MB', Used: '11248MB', UsedPercent: '39%' },
        '/data': { Total: '119234MB', Used: '74821MB', UsedPercent: '63%' },
      },
    })
    return true
  }

  /* ---------------- 设备 ---------------- */

  if (method === 'GET' && pathname === '/devices') {
    const exploded = query.get('exploded') === 'true'
    const list = exploded ? devices : devices.map(toSummary)
    send(res, 200, { devices: list })
    return true
  }

  if (method === 'POST' && pathname === '/devices') {
    const body = (await readJson(req)) as Record<string, unknown> | undefined
    if (!body || typeof body.name !== 'string' || typeof body.deviceType !== 'string') {
      sendErrors(res, 400, ['设备名称与设备类型不能为空'])
      return true
    }
    const device: StoredDevice = {
      ...(body as unknown as StoredDevice),
      id: createDeviceId(),
      eTag: '1',
      modTime: nowIso(),
      collectStatus: 'stopped',
    }
    devices = [...devices, device]
    send(res, 201, device, { ETag: device.eTag })
    return true
  }

  const actionMatch = pathname.match(/^\/devices\/([^/]+)\/action$/)
  if (method === 'PUT' && actionMatch) {
    const device = devices.find((d) => d.id === actionMatch[1])
    if (!device) {
      send(res, 404)
      return true
    }
    const actions = (await readJson(req)) as Array<Record<string, unknown>> | undefined
    if (!Array.isArray(actions)) {
      sendErrors(res, 400, ['请求体必须是数组'])
      return true
    }
    const errors: string[] = []
    for (const action of actions) {
      for (const [name, value] of Object.entries(action)) {
        const variable = device.variables?.find((v) => v.name === name)
        if (!variable) {
          errors.push(`变量 ${name} 不存在`)
          continue
        }
        if (variable.accessMode !== 'rw') {
          errors.push(`变量 ${name} 不可写`)
          continue
        }
        variable.value = value
      }
    }
    if (errors.length > 0) {
      sendErrors(res, 400, errors, 10011)
      return true
    }
    send(res, 202)
    return true
  }

  const statusMatch = pathname.match(/^\/devices\/([^/]+)\/(start|restart|stop)$/)
  if (method === 'PUT' && statusMatch) {
    const device = devices.find((d) => d.id === statusMatch[1])
    if (!device) {
      send(res, 404)
      return true
    }
    const action = statusMatch[2]
    device.collectStatus = action === 'stop' ? 'stopped' : 'collecting'
    device.modTime = nowIso()
    send(res, 202)
    return true
  }

  const deviceMatch = pathname.match(/^\/devices\/([^/]+)$/)
  if (deviceMatch && ['GET', 'PUT', 'DELETE'].includes(method)) {
    const device = devices.find((d) => d.id === deviceMatch[1])
    if (!device) {
      send(res, 404)
      return true
    }

    if (method === 'GET') {
      const exploded = query.get('exploded') === 'true'
      send(res, 200, exploded ? device : toSummary(device), { ETag: device.eTag })
      return true
    }

    if (!checkPrecondition(req, res, device)) return true

    if (method === 'DELETE') {
      devices = devices.filter((d) => d.id !== device.id)
      send(res, 200, device)
      return true
    }

    const body = (await readJson(req)) as Record<string, unknown> | undefined
    if (!body) {
      sendErrors(res, 400, ['请求体不能为空'])
      return true
    }
    const updated: StoredDevice = {
      ...device,
      ...(body as unknown as StoredDevice),
      id: device.id,
      eTag: device.eTag,
    }
    bump(updated)
    devices = devices.map((d) => (d.id === device.id ? updated : d))
    send(res, 200, updated, { ETag: updated.eTag })
    return true
  }

  /* ---------------- 数据服务 ---------------- */

  if (method === 'GET' && pathname === '/metrics') {
    send(res, 200, { data: metrics, total: metrics.length })
    return true
  }

  if (method === 'GET' && pathname === '/kpis') {
    send(res, 200, { data: kpis, total: kpis.length })
    return true
  }

  if (method === 'GET' && pathname === '/tags') {
    send(res, 200, { data: tags, total: tags.length })
    return true
  }

  if (method === 'POST' && pathname === '/metricsData') {
    const body = (await readJson(req)) as Record<string, unknown> | undefined
    const codes = Array.isArray(body?.metricsCodes) ? (body?.metricsCodes as string[]) : []
    if (codes.length === 0) {
      sendErrors(res, 400, ["Field validation for 'metricsCodes' failed on the 'required' tag"])
      return true
    }
    const { start, end } = hourRange(body as Record<string, unknown>, ['startTime', 'endTime'])
    const jitter = Math.sin(new Date().getHours()) * 0.01 + 1
    send(res, 200, {
      data: codes.map((metricsCode) => ({
        metricsCode,
        startTime: start,
        endTime: end,
        value: Number((baseValue(metricsCode) * jitter).toFixed(2)),
      })),
      total: codes.length,
    })
    return true
  }

  if (method === 'POST' && pathname === '/kpi') {
    const body = (await readJson(req)) as Record<string, unknown> | undefined
    const ids = Array.isArray(body?.ids) ? (body?.ids as string[]) : []
    if (ids.length === 0) {
      sendErrors(res, 400, ["Field validation for 'Ids' failed on the 'required' tag"])
      return true
    }
    const { start, end } = hourRange(body as Record<string, unknown>, ['startTime', 'endTime'])
    send(res, 200, {
      data: ids.map((id) => ({
        id,
        startTime: start,
        endTime: end,
        value: kpiValue(id),
      })),
      total: ids.length,
    })
    return true
  }

  const seriesMatch = pathname.match(/^\/(fmcs|escada)\/tagTimeSeries$/)
  if (method === 'POST' && seriesMatch) {
    const body = (await readJson(req)) as Record<string, unknown> | undefined
    const select = typeof body?.select === 'string' ? body.select : ''
    if (!select) {
      sendErrors(res, 400, ["Field validation for 'Select' failed on the 'required' tag"])
      return true
    }
    const limit = typeof body?.limit === 'number' ? body.limit : 48
    send(res, 200, { data: buildTimeSeries(select, limit) })
    return true
  }

  // 未实现的 /api/v1 路径显式返回 404,避免误以为连到了真实后端。
  send(res, 404, { errors: [{ code: 10004, message: `mock 未实现: ${method} ${pathname}` }] })
  return true
}

export function mockApiPlugin(): Plugin {
  return {
    name: 'harns-mock-api',
    apply: 'serve',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        const path = (req.url ?? '').split('?')[0]
        if (!path.startsWith(PREFIX)) {
          next()
          return
        }

        void (async () => {
          await new Promise((resolve) => setTimeout(resolve, LATENCY_MS))
          try {
            const handled = await handle(req, res)
            if (!handled) next()
          } catch (error) {
            const message = error instanceof Error ? error.message : 'mock error'
            sendErrors(res, 400, [message])
          }
        })()
      })
    },
  }
}
