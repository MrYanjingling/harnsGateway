export type Service = 'gateway' | 'data'

const BASES: Record<Service, string> = {
  gateway: import.meta.env.VITE_GATEWAY_BASE || '/api/v1',
  data: import.meta.env.VITE_DATA_BASE || '/api/v1',
}

export interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: unknown
  /** 乐观锁:写入设备时必须携带最近一次读取到的 eTag。 */
  ifMatch?: string
  headers?: Record<string, string>
  signal?: AbortSignal
}

export interface ApiResponse<T> {
  data: T
  etag?: string
}

export class ApiError extends Error {
  readonly status: number
  readonly code?: number

  constructor(message: string, status: number, code?: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

interface ErrorPayload {
  code?: number
  message?: string
  errors?: Array<{ code?: number; message?: string }>
  data?: unknown
}

function extractMessage(payload: unknown, status: number): { message: string; code?: number } {
  if (payload && typeof payload === 'object') {
    const p = payload as ErrorPayload
    if (Array.isArray(p.errors) && p.errors.length > 0) {
      const messages = p.errors.map((e) => e?.message).filter(Boolean)
      if (messages.length > 0) return { message: messages.join(';'), code: p.errors[0]?.code }
    }
    if (typeof p.message === 'string' && p.message) return { message: p.message, code: p.code }
    if (typeof p.data === 'string' && p.data) return { message: p.data }
    if (p.data && typeof p.data === 'object' && 'message' in p.data) {
      const m = (p.data as { message?: unknown }).message
      if (typeof m === 'string') return { message: m }
    }
  }
  return { message: statusMessage(status) }
}

function statusMessage(status: number): string {
  switch (status) {
    case 400:
      return '请求参数有误'
    case 404:
      return '资源不存在'
    case 405:
      return '请求方法不被允许'
    case 409:
      return '资源冲突'
    case 412:
      return '数据已被修改,请刷新后重试'
    case 428:
      return '缺少版本校验 (If-Match)'
    case 500:
      return '服务内部错误'
    case 502:
    case 503:
    case 504:
      return '无法连接到后端服务'
    default:
      return `请求失败 (HTTP ${status})`
  }
}

/** 统一请求入口:JSON 编解码、错误归一化、ETag 读取、空响应兼容。 */
export async function request<T>(
  service: Service,
  path: string,
  options: RequestOptions = {},
): Promise<ApiResponse<T>> {
  const { method = 'GET', body, ifMatch, headers = {}, signal } = options
  const url = `${BASES[service]}${path.startsWith('/') ? path : `/${path}`}`

  const finalHeaders: Record<string, string> = { Accept: 'application/json', ...headers }
  if (body !== undefined) finalHeaders['Content-Type'] = 'application/json'
  if (ifMatch) finalHeaders['If-Match'] = ifMatch

  let response: Response
  try {
    response = await fetch(url, {
      method,
      headers: finalHeaders,
      body: body === undefined ? undefined : JSON.stringify(body),
      signal,
    })
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') throw error
    throw new ApiError('无法连接到后端服务,请确认网关服务已启动', 0)
  }

  const text = await response.text()
  let payload: unknown = undefined
  if (text) {
    try {
      payload = JSON.parse(text)
    } catch {
      payload = text
    }
  }

  if (!response.ok) {
    const { message, code } = extractMessage(payload, response.status)
    throw new ApiError(message, response.status, code)
  }

  return {
    data: payload as T,
    etag: response.headers.get('ETag') ?? undefined,
  }
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}

export function errorMessage(error: unknown): string {
  if (isApiError(error)) return error.message
  if (error instanceof Error) return error.message
  return '发生未知错误'
}
