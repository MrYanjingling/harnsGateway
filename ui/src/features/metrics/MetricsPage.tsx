import { useEffect, useMemo, useRef, useState } from 'react'
import { Activity, BarChart3, Search } from 'lucide-react'
import { useMetrics, useMetricsValues } from '@/api/hooks'
import type { MetricsQueryRequest } from '@/api/types'
import { formatDateTimeShort, formatNumber } from '@/lib/format'
import { cn } from '@/lib/cn'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { EmptyState } from '@/components/ui/EmptyState'
import { Skeleton, TableSkeleton } from '@/components/ui/Skeleton'
import { TBody, TD, TH, THead, TR, Table } from '@/components/ui/Table'
import { ErrorState } from '@/components/shared/ErrorState'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatCard } from '@/components/shared/StatCard'
import { TimeRangeFields, useTimeRange } from '@/components/shared/TimeRangeFields'

export function MetricsPage() {
  const metrics = useMetrics()
  const range = useTimeRange(24)
  const [selected, setSelected] = useState<string[]>([])
  const [request, setRequest] = useState<MetricsQueryRequest | null>(null)
  const results = useMetricsValues(request)
  const initialized = useRef(false)

  useEffect(() => {
    if (initialized.current || !metrics.data) return
    initialized.current = true
    setSelected(
      metrics.data
        .filter((item) => item.enable)
        .slice(0, 4)
        .map((item) => item.metricsCode ?? '')
        .filter(Boolean),
    )
  }, [metrics.data])

  const metaByCode = useMemo(
    () =>
      new Map(
        (metrics.data ?? [])
          .filter((item) => item.metricsCode)
          .map((item) => [item.metricsCode as string, item] as const),
      ),
    [metrics.data],
  )

  const toggle = (code: string) => {
    setSelected((prev) =>
      prev.includes(code) ? prev.filter((item) => item !== code) : [...prev, code],
    )
  }

  const runQuery = () => {
    const api = range.toApi()
    if (!api || selected.length === 0) return
    setRequest({ metricsCodes: selected, granularity: 'h', ...api })
  }

  const allCodes = (metrics.data ?? [])
    .map((item) => item.metricsCode ?? '')
    .filter(Boolean)
  const allSelected = allCodes.length > 0 && selected.length === allCodes.length

  return (
    <div>
      <PageHeader
        title="统计指标"
        description="查看数据服务中的指标定义,按时间范围查询聚合值(小时粒度)。"
        actions={
          <Button
            variant="primary"
            icon={<Search className="size-4" />}
            disabled={selected.length === 0 || !range.valid}
            onClick={runQuery}
          >
            查询指标值
          </Button>
        }
      />

      <div className="grid gap-4 lg:grid-cols-[1.4fr_1fr]">
        <Card className="overflow-hidden">
          <CardHeader>
            <div>
              <CardTitle>指标列表</CardTitle>
              <p className="mt-0.5 text-xs text-muted">
                已选择 {selected.length} / {allCodes.length} 个指标
              </p>
            </div>
            <div className="flex items-center gap-1">
              <Button
                size="sm"
                variant="ghost"
                onClick={() => setSelected(allSelected ? [] : allCodes)}
              >
                {allSelected ? '取消全选' : '全选'}
              </Button>
            </div>
          </CardHeader>

          {metrics.isLoading ? (
            <TableSkeleton rows={5} cols={4} />
          ) : metrics.isError ? (
            <div className="p-5">
              <ErrorState error={metrics.error} onRetry={metrics.refetch} compact />
            </div>
          ) : (metrics.data ?? []).length === 0 ? (
            <EmptyState title="没有指标" description="数据服务尚未配置任何统计指标。" />
          ) : (
            <Table>
              <THead>
                <tr>
                  <TH className="w-10">
                    <input
                      type="checkbox"
                      aria-label="全选指标"
                      checked={allSelected}
                      onChange={() => setSelected(allSelected ? [] : allCodes)}
                      className="size-4 cursor-pointer accent-[var(--primary)]"
                    />
                  </TH>
                  <TH>指标</TH>
                  <TH>计算</TH>
                  <TH>单位</TH>
                  <TH>状态</TH>
                </tr>
              </THead>
              <TBody>
                {(metrics.data ?? []).map((item) => {
                  const code = item.metricsCode ?? item.id
                  const checked = selected.includes(item.metricsCode ?? '')
                  return (
                    <TR
                      key={item.id}
                      className={cn('cursor-pointer', checked && 'bg-primary-soft/40')}
                      onClick={() => toggle(item.metricsCode ?? '')}
                    >
                      <TD>
                        <input
                          type="checkbox"
                          aria-label={`选择 ${code}`}
                          checked={checked}
                          onChange={() => toggle(item.metricsCode ?? '')}
                          onClick={(event) => event.stopPropagation()}
                          className="size-4 cursor-pointer accent-[var(--primary)]"
                        />
                      </TD>
                      <TD>
                        <p className="font-medium">{item.metricsName || code}</p>
                        <p className="font-mono text-[11px] text-faint">{code}</p>
                      </TD>
                      <TD className="text-xs text-muted">
                        {item.calculateType || '—'}
                        {item.metricsType ? (
                          <span className="ml-1.5 text-faint">/ {item.metricsType}</span>
                        ) : null}
                      </TD>
                      <TD className="text-xs text-muted">{item.measureType || '—'}</TD>
                      <TD>
                        {item.enable ? (
                          <Badge tone="success" dot>
                            启用
                          </Badge>
                        ) : (
                          <Badge tone="neutral" dot>
                            停用
                          </Badge>
                        )}
                      </TD>
                    </TR>
                  )
                })}
              </TBody>
            </Table>
          )}
        </Card>

        <Card>
          <CardHeader>
            <div>
              <CardTitle>查询条件</CardTitle>
              <p className="mt-0.5 text-xs text-muted">聚合粒度固定为小时</p>
            </div>
            <BarChart3 className="size-4 text-faint" />
          </CardHeader>
          <CardContent className="space-y-4">
            <TimeRangeFields range={range} />
            <p className="text-[11px] leading-relaxed text-faint">
              指标值由数据服务按 <code className="font-mono">POST /metricsData</code> 聚合返回。
              开始时间需早于结束时间,且至少选择一个指标。
            </p>
            <Button
              className="w-full"
              variant="primary"
              icon={<Search className="size-4" />}
              disabled={selected.length === 0 || !range.valid}
              onClick={runQuery}
            >
              查询 {selected.length > 0 ? `${selected.length} 个指标` : ''}
            </Button>
          </CardContent>
        </Card>
      </div>

      <div className="mt-6">
        <h2 className="mb-3 text-sm font-semibold">查询结果</h2>
        {!request ? (
          <Card>
            <EmptyState
              icon={<Activity className="size-5" />}
              title="尚未查询"
              description="选择指标与时间范围后点击查询,结果会显示在这里。"
            />
          </Card>
        ) : results.isLoading ? (
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            {request.metricsCodes.map((code) => (
              <Skeleton key={code} className="h-28 rounded-2xl" />
            ))}
          </div>
        ) : results.isError ? (
          <ErrorState error={results.error} onRetry={results.refetch} />
        ) : (results.data ?? []).length === 0 ? (
          <Card>
            <EmptyState title="没有数据" description="该时间范围内没有查询到指标值。" />
          </Card>
        ) : (
          <>
            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
              {(results.data ?? []).map((result) => {
                const meta = metaByCode.get(result.metricsCode)
                return (
                  <StatCard
                    key={result.metricsCode}
                    label={meta?.metricsName || result.metricsCode}
                    value={
                      <span>
                        {formatNumber(result.value, 3)}
                        {meta?.measureType ? (
                          <span className="ml-1 text-sm font-normal text-muted">
                            {meta.measureType}
                          </span>
                        ) : null}
                      </span>
                    }
                    hint={`${formatDateTimeShort(result.startTime)} ~ ${formatDateTimeShort(result.endTime)}`}
                    icon={<Activity className="size-4" />}
                  />
                )
              })}
            </div>
            <p className="mt-3 text-[11px] text-faint">
              查询区间:{formatDateTimeShort(request.startTime)} ~ {formatDateTimeShort(request.endTime)}
              {' · '}粒度 {request.granularity}
            </p>
          </>
        )}
      </div>
    </div>
  )
}
