import { request } from '@/lib/http'
import type {
  DeviceAction,
  DeviceDetail,
  DeviceListResponse,
  DevicePayload,
  DeviceStatusAction,
  DeviceSummary,
} from './types'

export async function listDevices(exploded = false): Promise<DeviceSummary[]> {
  const suffix = exploded ? '?exploded=true' : ''
  const { data } = await request<DeviceListResponse>('gateway', `/devices${suffix}`)
  return data.devices ?? []
}

export async function getDeviceById(id: string, exploded = true): Promise<DeviceDetail> {
  const suffix = exploded ? '?exploded=true' : ''
  const { data } = await request<DeviceDetail>('gateway', `/devices/${id}${suffix}`)
  return data
}

export async function createDevice(payload: DevicePayload): Promise<DeviceDetail> {
  const { data } = await request<DeviceDetail>('gateway', '/devices', {
    method: 'POST',
    body: payload,
  })
  return data
}

export async function updateDevice(
  id: string,
  payload: DevicePayload,
  etag: string,
): Promise<DeviceDetail> {
  const { data } = await request<DeviceDetail>('gateway', `/devices/${id}`, {
    method: 'PUT',
    body: payload,
    ifMatch: etag,
  })
  return data
}

export async function deleteDevice(id: string, etag: string): Promise<void> {
  await request<unknown>('gateway', `/devices/${id}`, { method: 'DELETE', ifMatch: etag })
}

export async function switchDeviceStatus(
  id: string,
  action: DeviceStatusAction,
): Promise<void> {
  await request<unknown>('gateway', `/devices/${id}/${action}`, { method: 'PUT' })
}

export async function controlDevice(id: string, actions: DeviceAction[]): Promise<void> {
  await request<unknown>('gateway', `/devices/${id}/action`, {
    method: 'PUT',
    body: actions,
  })
}
