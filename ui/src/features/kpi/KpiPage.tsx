import { useEffect, useMemo, useRef, useState } from 'react'
import { Search, Target } from 'lucide-react'
import { useKpis, useKpiValues } from '@/api/hooks'
import type { KpiQueryRequest } from '@/api/types'
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

const GRAIN_LABEL: Record<string, string> = {
  h: '小时',
  d: '天',
  m: '月',
  y: '年',
}

export function KpiPage() {
  const kpis = useKpis()
  const range = useTimeRange(24)
  const [selected, setSelected] = useState<string[]>([])
  const [request, setRequest] = useState<KpiQueryRequest | null>(null)
  const results = useKpiValues(request)
  const initialized = useRef(false)

  useEffect(() => {
    if (initialized.current || !kpis.data) return
    initialized.current = true
    setSelected(kpis.data.filter((item) => item.status === 'active').map((item) => item.id))
  }, [kpis.data])

  const metaById = useMemo(
    () => new Map((kpis.data ?? []).map((item) => [item.id, item] as const)),
    [kpis.data],
  )

  const toggle = (id: string) =>
    setSelected((prev) => (prev.includes(id) ? prev.filter((item) => item !== id) : [...prev, id]))

  const allIds = (kpis.data ?? []).map((item) => item.id)
  const allSelected = allIds.length > 0 && selected.length === allIds.length

  const runQuery = () => {
    const api = range.toApi()
    if (!api || selected.length === 0) return
    setRequest({ ids: selected, granularity: 'h', ...api })
  }

  return (
    <div>
      <PageHeader
        title="KPI 指标"
        description="查看关键绩效指标定义,并按时间范围计算 KPI 值。"
        actions={
          <Button
            variant="primary"
            icon={<Search className="size-4" />}
            disabled={selected.length === 0 || !range.valid}
            onClick={runQuery}
          >
            计算 KPI
          </Button>
        }
      />

      <div className="grid gap-4 lg:grid-cols-[1.4fr_1fr]">
        <Card className="overflow-hidden">
          <CardHeader>
            <div>
              <CardTitle>KPI 列表</CardTitle>
              <p className="mt-0.5 text-xs text-muted">
                已选择 {selected.length} / {allIds.length} 项
              </p>
            </div>
            <Button size="sm" variant="ghost" onClick={() => setSelected(allSelected ? [] : allIds)}>
              {allSelected ? '取消全选' : '全选'}
            </Button>
          </CardHeader>

          {kpis.isLoading ? (
            <TableSkeleton rows={4} cols={5} />
          ) : kpis.isError ? (
            <div className="p-5">
              <ErrorState error={kpis.error} onRetry={kpis.refetch} compact />
            </div>
          ) : (kpis.data ?? []).length === 0 ? (
            <EmptyState title="没有 KPI" description="数据服务尚未配置 KPI 指标。" />
          ) : (
            <Table>
              <THead>
                <tr>
                  <TH className="w-10">
                    <input
                      type="checkbox"
                      aria-label="全选 KPI"
                      checked={allSelected}
                      onChange={() => setSelected(allSelected ? [] : allIds)}
                      className="size-4 cursor-pointer accent-[var(--primary)]"
                    />
                  </TH>
                  <TH>名称</TH>
                  <TH>公式</TH>
                  <TH>粒度</TH>
                  <TH>状态</TH>
                </tr>
              </THead>
              <TBody>
                {(kpis.data ?? []).map((item) => {
                  const checked = selected.includes(item.id)
                  return (
                    <TR
                      key={item.id}
                      className={cn('cursor-pointer', checked && 'bg-primary-soft/40')}
                      onClick={() => toggle(item.id)}
                    >
                      <TD>
                        <input
                          type="checkbox"
                          aria-label={`选择 ${item.name}`}
                          checked={checked}
                          onChange={() => toggle(item.id)}
                          onClick={(event) => event.stopPropagation()}
                          className="size-4 cursor-pointer accent-[var(--primary)]"
                        />
                      </TD>
                      <TD>
                        <p className="font-medium">{item.name}</p>
                        <p className="font-mono text-[11px] text-faint">{item.id}</p>
                      </TD>
                      <TD>
                        <code
                          className="block max-w-48 truncate font-mono text-[11px] text-muted"
                          title={item.formula ?? item.formulaOri}
                        >
                          {item.formula || item.formulaOri}
                        </code>
                      </TD>
                      <TD className="text-xs text-muted">
                        {GRAIN_LABEL[item.timeGranularity] ?? item.timeGranularity}
                      </TD>
                      <TD>
                        <Badge
                          tone={item.status === 'active' ? 'success' : 'warning'}
                          dot
                        >
                          {item.status === 'active' ? '启用' : item.status === 'draft' ? '草稿' : item.status}
                        </Badge>
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
              <CardTitle>计算条件</CardTitle>
              <p className="mt-0.5 text-xs text-muted">KPI 值按小时粒度聚合</p>
            </div>
            <Target className="size-4 text-faint" />
          </CardHeader>
          <CardContent className="space-y-4">
            <TimeRangeFields range={range} />
            <p className="text-[11px] leading-relaxed text-faint">
              通过 <code className="font-mono">POST /kpi</code> 提交 KPI ID 列表与时间范围。
              公式中的指标编码由数据服务解析。
            </p>
            <Button
              className="w-full"
              variant="primary"
              icon={<Search className="size-4" />}
              disabled={selected.length === 0 || !range.valid}
              onClick={runQuery}
            >
              计算 {selected.length > 0 ? `${selected.length} 项 KPI` : ''}
            </Button>
          </CardContent>
        </Card>
      </div>

      <div className="mt-6">
        <h2 className="mb-3 text-sm font-semibold">计算结果</h2>
        {!request ? (
          <Card>
            <EmptyState
              icon={<Target className="size-5" />}
              title="尚未计算"
              description="选择 KPI 与时间范围后点击计算,结果会显示在这里。"
            />
          </Card>
        ) : results.isLoading ? (
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
            {request.ids.map((id) => (
              <Skeleton key={id} className="h-28 rounded-2xl" />
            ))}
          </div>
        ) : results.isError ? (
          <ErrorState error={results.error} onRetry={results.refetch} />
        ) : (results.data ?? []).length === 0 ? (
          <Card>
            <EmptyState title="没有结果" description="该时间范围内没有计算出 KPI 值。" />
          </Card>
        ) : (
          <>
            <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
              {(results.data ?? []).map((result) => {
                const meta = metaById.get(result.id)
                const target = meta?.targetData ? Number(meta.targetData) : Number.NaN
                const hasTarget = Number.isFinite(target)
                return (
                  <StatCard
                    key={result.id}
                    label={meta?.name ?? result.id}
                    value={
                      <span>
                        {formatNumber(result.value, 2)}
                        {meta?.unit ? (
                          <span className="ml-1 text-sm font-normal text-muted">{meta.unit}</span>
                        ) : null}
                      </span>
                    }
                    hint={`${formatDateTimeShort(result.startTime)} ~ ${formatDateTimeShort(result.endTime)}`}
                    icon={<Target className="size-4" />}
                    footer={
                      hasTarget ? (
                        <div className="flex items-center justify-between text-[11px]">
                          <span className="text-faint">
                            目标 {meta?.targetData}
                            {meta?.unit}
                          </span>
                          <Badge tone={result.value >= target ? 'success' : 'warning'}>
                            {result.value >= target ? '达标' : '未达标'}
                          </Badge>
                        </div>
                      ) : null
                    }
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
