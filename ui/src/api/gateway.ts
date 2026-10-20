import { request } from '@/lib/http'
import type {
  GatewayCpuResponse,
  GatewayDiskResponse,
  GatewayMemResponse,
  GatewayMeta,
} from './types'

export async function getGatewayMeta(): Promise<GatewayMeta> {
  const { data, etag } = await request<GatewayMeta>('gateway', '/gatewayMeta')
  // 后端同时写入 ETag 响应头与 body.eTag,优先使用 body。
  return { ...data, eTag: data.eTag || etag || '' }
}

export async function getGatewayCpu(): Promise<GatewayCpuResponse> {
  const { data } = await request<GatewayCpuResponse>('gateway', '/gatewayCpu')
  return data
}

export async function getGatewayMem(): Promise<GatewayMemResponse> {
  const { data } = await request<GatewayMemResponse>('gateway', '/gatewayMem')
  return data
}

export async function getGatewayDisk(): Promise<GatewayDiskResponse> {
  const { data } = await request<GatewayDiskResponse>('gateway', '/gatewayDisk')
  return data
}
