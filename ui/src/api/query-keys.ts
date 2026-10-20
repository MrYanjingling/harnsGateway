import type { KpiQueryRequest, MetricsQueryRequest, TagTimeSeriesRequest, TimeSeriesSource } from './types'

export const queryKeys = {
  gateway: {
    all: ['gateway'] as const,
    meta: ['gateway', 'meta'] as const,
    cpu: ['gateway', 'cpu'] as const,
    mem: ['gateway', 'mem'] as const,
    disk: ['gateway', 'disk'] as const,
  },
  devices: {
    all: ['devices'] as const,
    list: (exploded: boolean) => ['devices', 'list', exploded] as const,
    detail: (id: string) => ['devices', 'detail', id] as const,
  },
  data: {
    metrics: ['data', 'metrics'] as const,
    kpis: ['data', 'kpis'] as const,
    tags: ['data', 'tags'] as const,
    metricsValues: (req: MetricsQueryRequest) => ['data', 'metricsValues', req] as const,
    kpiValues: (req: KpiQueryRequest) => ['data', 'kpiValues', req] as const,
    timeSeries: (source: TimeSeriesSource, req: TagTimeSeriesRequest) =>
      ['data', 'timeSeries', source, req] as const,
  },
}
