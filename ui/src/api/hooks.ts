import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { isApiError } from '@/lib/http'
import { listKpis, listMetrics, listTags, queryKpi, queryMetrics, queryTimeSeries } from './data'
import {
  controlDevice,
  createDevice,
  deleteDevice,
  getDeviceById,
  listDevices,
  switchDeviceStatus,
  updateDevice,
} from './devices'
import { getGatewayCpu, getGatewayDisk, getGatewayMem, getGatewayMeta } from './gateway'
import { queryKeys } from './query-keys'
import type {
  DeviceAction,
  DevicePayload,
  DeviceStatusAction,
  KpiQueryRequest,
  MetricsQueryRequest,
  TagTimeSeriesRequest,
  TimeSeriesSource,
} from './types'

const RESOURCE_POLL_MS = 10_000

/* ---------------- 网关 ---------------- */

export function useGatewayMeta() {
  return useQuery({
    queryKey: queryKeys.gateway.meta,
    queryFn: getGatewayMeta,
  })
}

export function useGatewayCpu() {
  return useQuery({
    queryKey: queryKeys.gateway.cpu,
    queryFn: getGatewayCpu,
    refetchInterval: RESOURCE_POLL_MS,
  })
}

export function useGatewayMem() {
  return useQuery({
    queryKey: queryKeys.gateway.mem,
    queryFn: getGatewayMem,
    refetchInterval: RESOURCE_POLL_MS,
  })
}

export function useGatewayDisk() {
  return useQuery({
    queryKey: queryKeys.gateway.disk,
    queryFn: getGatewayDisk,
    refetchInterval: RESOURCE_POLL_MS,
  })
}

/* ---------------- 设备 ---------------- */

export function useDevices(exploded = false) {
  return useQuery({
    queryKey: queryKeys.devices.list(exploded),
    queryFn: () => listDevices(exploded),
    refetchInterval: 10_000,
  })
}

export function useDevice(id: string | undefined) {
  return useQuery({
    queryKey: queryKeys.devices.detail(id ?? ''),
    queryFn: () => getDeviceById(id as string, true),
    enabled: Boolean(id),
    refetchInterval: 5_000,
  })
}

function useInvalidateDevicesOnConflict() {
  const queryClient = useQueryClient()
  return (error: unknown) => {
    // 412:别人先改了,强制刷新到最新版本。
    if (isApiError(error) && (error.status === 412 || error.status === 404)) {
      void queryClient.invalidateQueries({ queryKey: queryKeys.devices.all })
    }
  }
}

export function useCreateDevice() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (payload: DevicePayload) => createDevice(payload),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.devices.all }),
  })
}

export function useUpdateDevice() {
  const queryClient = useQueryClient()
  const onConflict = useInvalidateDevicesOnConflict()
  return useMutation({
    mutationFn: (vars: { id: string; payload: DevicePayload; etag: string }) =>
      updateDevice(vars.id, vars.payload, vars.etag),
    onSuccess: (_data, vars) => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.devices.all })
      void queryClient.invalidateQueries({ queryKey: queryKeys.devices.detail(vars.id) })
    },
    onError: onConflict,
  })
}

export function useDeleteDevice() {
  const queryClient = useQueryClient()
  const onConflict = useInvalidateDevicesOnConflict()
  return useMutation({
    mutationFn: (vars: { id: string; etag: string }) => deleteDevice(vars.id, vars.etag),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.devices.all }),
    onError: onConflict,
  })
}

export function useSwitchDeviceStatus() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (vars: { id: string; action: DeviceStatusAction }) =>
      switchDeviceStatus(vars.id, vars.action),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.devices.all }),
  })
}

export function useControlDevice() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (vars: { id: string; actions: DeviceAction[] }) =>
      controlDevice(vars.id, vars.actions),
    onSuccess: (_data, vars) =>
      queryClient.invalidateQueries({ queryKey: queryKeys.devices.detail(vars.id) }),
  })
}

/* ---------------- 数据服务 ---------------- */

export function useMetrics() {
  return useQuery({ queryKey: queryKeys.data.metrics, queryFn: listMetrics })
}

export function useKpis() {
  return useQuery({ queryKey: queryKeys.data.kpis, queryFn: listKpis })
}

export function useTags() {
  return useQuery({ queryKey: queryKeys.data.tags, queryFn: listTags })
}

export function useMetricsValues(req: MetricsQueryRequest | null) {
  return useQuery({
    queryKey: queryKeys.data.metricsValues(req ?? { metricsCodes: [], startTime: '', endTime: '', granularity: 'h' }),
    queryFn: () => queryMetrics(req as MetricsQueryRequest),
    enabled: Boolean(req && req.metricsCodes.length > 0),
    staleTime: 60_000,
  })
}

export function useKpiValues(req: KpiQueryRequest | null) {
  return useQuery({
    queryKey: queryKeys.data.kpiValues(req ?? { ids: [], startTime: '', endTime: '', granularity: 'h' }),
    queryFn: () => queryKpi(req as KpiQueryRequest),
    enabled: Boolean(req && req.ids.length > 0),
    staleTime: 60_000,
  })
}

export function useTimeSeries(
  source: TimeSeriesSource,
  req: TagTimeSeriesRequest | null,
) {
  return useQuery({
    queryKey: queryKeys.data.timeSeries(source, req ?? { select: '' }),
    queryFn: () => queryTimeSeries(source, req as TagTimeSeriesRequest),
    enabled: Boolean(req?.select),
    staleTime: 60_000,
  })
}
