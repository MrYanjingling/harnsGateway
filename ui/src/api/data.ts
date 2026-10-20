import { request } from '@/lib/http'
import type {
  DataPointTimeSeries,
  DataResponse,
  KpiInfo,
  KpiQueryRequest,
  KpiResult,
  MetricsInfo,
  MetricsQueryRequest,
  MetricsResult,
  TagInfo,
  TagTimeSeriesRequest,
  TimeSeriesSource,
} from './types'

async function list<T>(path: string): Promise<T[]> {
  const { data } = await request<DataResponse<T[]>>('data', path)
  return data.data ?? []
}

export const listMetrics = () => list<MetricsInfo>('/metrics')
export const listKpis = () => list<KpiInfo>('/kpis')
export const listTags = () => list<TagInfo>('/tags')

export async function queryMetrics(req: MetricsQueryRequest): Promise<MetricsResult[]> {
  const { data } = await request<DataResponse<MetricsResult[]>>('data', '/metricsData', {
    method: 'POST',
    body: req,
  })
  return data.data ?? []
}

export async function queryKpi(req: KpiQueryRequest): Promise<KpiResult[]> {
  const { data } = await request<DataResponse<KpiResult[]>>('data', '/kpi', {
    method: 'POST',
    body: req,
  })
  return data.data ?? []
}

export async function queryTimeSeries(
  source: TimeSeriesSource,
  req: TagTimeSeriesRequest,
): Promise<DataPointTimeSeries[]> {
  const { data } = await request<DataResponse<DataPointTimeSeries[]>>(
    'data',
    `/${source}/tagTimeSeries`,
    { method: 'POST', body: req },
  )
  return data.data ?? []
}
