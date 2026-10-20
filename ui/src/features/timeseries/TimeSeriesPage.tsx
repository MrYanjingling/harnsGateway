import { useEffect, useMemo, useRef, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import {
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { AreaChart, Search } from 'lucide-react'
import { useTags, useTimeSeries } from '@/api/hooks'
import type { DataPointTimeSeries, TagTimeSeriesRequest, TimeSeriesMode, TimeSeriesSource } from '@/api/types'
import { formatDateTimeShort, formatNumber } from '@/lib/format'
import { cn } from '@/lib/cn'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { EmptyState } from '@/components/ui/EmptyState'
import { Field, Input, Select } from '@/components/ui/Field'
import { PageLoading } from '@/components/ui/Spinner'
import { TBody, TD, TH, THead, TR, Table } from '@/components/ui/Table'
import { ErrorState } from '@/components/shared/ErrorState'
import { PageHeader } from '@/components/shared/PageHeader'
import { TimeRangeFields, useTimeRange } from '@/components/shared/TimeRangeFields'

const SERIES_COLORS = [
  '#10b981',
  '#0ea5e9',
  '#f59e0b',
  '#8b5cf6',
  '#ef4444',
  '#14b8a6',
  '#6366f1',
  '#f43f5e',
  '#84cc16',
  '#06b6d4',
]

const MODES: Array<{ value: TimeSeriesMode; label: string; hint: string }> = [
  { value: 'raw', label: '原始值', hint: '逐点返回原始采集值' },
  { value: 'mean', label: '平均值', hint: '按时间桶计算平均值' },
  { value: 'usage', label: '差值', hint: '用于累计量,返回桶内差值' },
]

/**
 * 把多个测点的时序按时间戳合并成图表行。
 * 使用 s0/s1 作为列名,避免测点 ID 中的 "." 被 Recharts 当作嵌套路径。
 */
function mergeSeries(series: DataPointTimeSeries[]) {
  const rows = new Map<string, Record<string, string | number>>()
  series.forEach((item, index) => {
    const key = `s${index}`
    for (const point of item.data ?? []) {
      const value = Number(point.value)
      if (!Number.isFinite(value)) continue
      const row = rows.get(point.time) ?? { time: point.time }
      row[key] = value
      rows.set(point.time, row)
    }
  })
  return Array.from(rows.values()).sort((a, b) =>
    String(a.time).localeCompare(String(b.time)),
  )
}

function summarize(series: DataPointTimeSeries) {
  const values = (series.data ?? [])
    .map((point) => Number(point.value))
    .filter((value) => Number.isFinite(value))
  if (values.length === 0) return null
  return {
    count: values.length,
    min: Math.min(...values),
    max: Math.max(...values),
    latest: values[values.length - 1],
  }
}

export function TimeSeriesPage() {
  const [searchParams] = useSearchParams()
  const tags = useTags()
  const range = useTimeRange(24)
  const [source, setSource] = useState<TimeSeriesSource>('fmcs')
  const [selectText, setSelectText] = useState(searchParams.get('select') ?? '')
  const [mode, setMode] = useState<TimeSeriesMode>('raw')
  const [limit, setLimit] = useState(48)
  const [tagSearch, setTagSearch] = useState('')
  const [request, setRequest] = useState<TagTimeSeriesRequest | null>(null)
  const series = useTimeSeries(source, request)
  const initialized = useRef(Boolean(searchParams.get('select')))

  useEffect(() => {
    if (initialized.current || !tags.data?.length) return
    initialized.current = true
    const first = tags.data.find((tag) => tag.enable === 1 && tag.tagId) ?? tags.data[0]
    if (first?.tagId) setSelectText(first.tagId)
  }, [tags.data])

  const selectedIds = useMemo(
    () =>
      selectText
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean),
    [selectText],
  )

  const filteredTags = useMemo(() => {
    const keyword = tagSearch.trim().toLowerCase()
    return (tags.data ?? []).filter((tag) => {
      if (!keyword) return true
      return (
        (tag.tagId ?? '').toLowerCase().includes(keyword) ||
        (tag.tagName ?? '').toLowerCase().includes(keyword)
      )
    })
  }, [tags.data, tagSearch])

  const toggleTag = (tagId: string) => {
    const next = selectedIds.includes(tagId)
      ? selectedIds.filter((item) => item !== tagId)
      : [...selectedIds, tagId]
    setSelectText(next.join(','))
  }

  const runQuery = () => {
    const api = range.toApi()
    const select = selectText.trim()
    if (!api || !select) return
    setRequest({
      select,
      start: `${api.startTime}.000`,
      end: `${api.endTime}.000`,
      limit,
      mode,
    })
  }

  const chartData = useMemo(() => mergeSeries(series.data ?? []), [series.data])
  const chartSeries = series.data ?? []

  const canQuery = selectedIds.length > 0 && range.valid && limit >= 2

  return (
    <div>
      <PageHeader
        title="时序查询"
        description="按测点与时间范围查询历史数据,支持原始值、平均值与差值模式。"
        actions={
          <Button
            variant="primary"
            icon={<Search className="size-4" />}
            disabled={!canQuery}
            onClick={runQuery}
          >
            查询时序
          </Button>
        }
      />

      <div className="grid gap-4 lg:grid-cols-[1.35fr_1fr]">
        <Card>
          <CardHeader>
            <div>
              <CardTitle>查询条件</CardTitle>
              <p className="mt-0.5 text-xs text-muted">
                多个测点用英文逗号分隔,最多建议 10 个
              </p>
            </div>
            <Badge tone="neutral">{selectedIds.length} 个测点</Badge>
          </CardHeader>
          <CardContent className="space-y-4">
            <Field
              label="测点 (select)"
              hint="支持直接粘贴测点 ID,或从下方列表勾选"
            >
              <Input
                value={selectText}
                onChange={(event) => setSelectText(event.target.value)}
                placeholder="例如: SCADA.compressor.pressure,SCADA.chiller.flow"
                className="font-mono text-xs"
              />
            </Field>

            <div className="flex flex-wrap items-end gap-3">
              <Field label="数据源" className="w-36">
                <Select
                  value={source}
                  onChange={(event) => setSource(event.target.value as TimeSeriesSource)}
                >
                  <option value="fmcs">FMCS</option>
                  <option value="escada">Escada</option>
                </Select>
              </Field>
              <Field label="模式" className="w-36">
                <Select
                  value={mode}
                  onChange={(event) => setMode(event.target.value as TimeSeriesMode)}
                >
                  {MODES.map((item) => (
                    <option key={item.value} value={item.value}>
                      {item.label}
                    </option>
                  ))}
                </Select>
              </Field>
              <Field label="条数上限" className="w-28" error={limit < 2 ? '至少 2 条' : undefined}>
                <Input
                  type="number"
                  min={2}
                  max={2000}
                  value={limit}
                  onChange={(event) => setLimit(Number(event.target.value))}
                />
              </Field>
            </div>

            <TimeRangeFields range={range} />

            <p className="text-[11px] leading-relaxed text-faint">
              {MODES.find((item) => item.value === mode)?.hint}。请求会发送到{' '}
              <code className="font-mono">POST /{source}/tagTimeSeries</code>,时间格式为{' '}
              <code className="font-mono">yyyy-MM-dd HH:mm:ss.SSS</code>。
            </p>

            <Button
              className="w-full"
              variant="primary"
              icon={<Search className="size-4" />}
              disabled={!canQuery}
              onClick={runQuery}
            >
              查询 {selectedIds.length > 0 ? `${selectedIds.length} 个测点` : '时序'}
            </Button>
          </CardContent>
        </Card>

        <Card className="flex flex-col">
          <CardHeader>
            <div>
              <CardTitle>可用测点</CardTitle>
              <p className="mt-0.5 text-xs text-muted">点击勾选,自动写入 select</p>
            </div>
            <Input
              value={tagSearch}
              onChange={(event) => setTagSearch(event.target.value)}
              placeholder="筛选测点"
              className="h-8 w-36 text-xs"
            />
          </CardHeader>
          <div className="max-h-96 overflow-y-auto px-2 pb-2">
            {tags.isLoading ? (
              <PageLoading label="加载测点…" />
            ) : tags.isError ? (
              <div className="p-3">
                <ErrorState error={tags.error} onRetry={tags.refetch} compact />
              </div>
            ) : filteredTags.length === 0 ? (
              <p className="px-3 py-8 text-center text-xs text-muted">没有匹配的测点</p>
            ) : (
              <ul className="space-y-0.5">
                {filteredTags.map((tag) => {
                  const tagId = tag.tagId ?? ''
                  const checked = selectedIds.includes(tagId)
                  return (
                    <li key={tag.id}>
                      <label
                        className={cn(
                          'flex cursor-pointer items-start gap-2.5 rounded-lg px-3 py-2 transition-colors',
                          checked ? 'bg-primary-soft/60' : 'hover:bg-surface-2',
                        )}
                      >
                        <input
                          type="checkbox"
                          checked={checked}
                          onChange={() => toggleTag(tagId)}
                          className="mt-0.5 size-4 shrink-0 cursor-pointer accent-[var(--primary)]"
                        />
                        <span className="min-w-0">
                          <span className="block truncate text-xs font-medium">
                            {tag.tagName || tagId}
                          </span>
                          <span className="block truncate font-mono text-[10px] text-faint">
                            {tagId}
                          </span>
                        </span>
                      </label>
                    </li>
                  )
                })}
              </ul>
            )}
          </div>
        </Card>
      </div>

      <Card className="mt-4">
        <CardHeader>
          <div>
            <CardTitle>趋势图</CardTitle>
            <p className="mt-0.5 text-xs text-muted">
              {chartSeries.length > 0
                ? `${chartSeries.length} 条曲线 · ${chartData.length} 个时间点`
                : '查询后展示多测点折线'}
            </p>
          </div>
          {request && series.isFetching ? (
            <span className="flex items-center gap-1.5 text-[11px] text-faint">
              <span className="size-1.5 animate-pulse rounded-full bg-primary" />
              加载中
            </span>
          ) : null}
        </CardHeader>
        <CardContent>
          {!request ? (
            <EmptyState
              icon={<AreaChart className="size-5" />}
              title="尚未查询"
              description="选择测点、时间范围与模式后点击查询。"
            />
          ) : series.isLoading ? (
            <PageLoading label="正在查询时序数据…" />
          ) : series.isError ? (
            <ErrorState error={series.error} onRetry={series.refetch} compact />
          ) : chartSeries.length === 0 || chartData.length === 0 ? (
            <EmptyState
              title="没有时序数据"
              description="该测点在所选时间范围内没有数据,请调整时间范围或测点。"
            />
          ) : (
            <div className="h-[340px] w-full">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={chartData} margin={{ top: 8, right: 16, bottom: 4, left: 0 }}>
                  <CartesianGrid stroke="var(--color-border)" strokeDasharray="3 3" vertical={false} />
                  <XAxis
                    dataKey="time"
                    tickFormatter={(value: string) => value.slice(11, 16)}
                    tick={{ fill: 'var(--color-faint)', fontSize: 11 }}
                    tickLine={false}
                    axisLine={{ stroke: 'var(--color-border)' }}
                    minTickGap={28}
                  />
                  <YAxis
                    tick={{ fill: 'var(--color-faint)', fontSize: 11 }}
                    tickLine={false}
                    axisLine={false}
                    width={56}
                    tickFormatter={(value: number) => formatNumber(value, 2)}
                  />
                  <Tooltip
                    contentStyle={{
                      background: 'var(--color-surface)',
                      border: '1px solid var(--color-border)',
                      borderRadius: 12,
                      fontSize: 12,
                      color: 'var(--color-fg)',
                      boxShadow: '0 8px 24px rgba(0,0,0,0.12)',
                    }}
                    labelStyle={{ color: 'var(--color-muted)', marginBottom: 4 }}
                    labelFormatter={(value) => formatDateTimeShort(String(value))}
                    formatter={(value) => formatNumber(Number(value), 3)}
                  />
                  <Legend
                    iconType="plainline"
                    wrapperStyle={{ fontSize: 12, paddingTop: 8 }}
                  />
                  {chartSeries.map((item, index) => (
                    <Line
                      key={item.tagId}
                      type="monotone"
                      dataKey={`s${index}`}
                      name={item.tagName || item.tagId}
                      stroke={SERIES_COLORS[index % SERIES_COLORS.length]}
                      strokeWidth={2}
                      dot={false}
                      activeDot={{ r: 3 }}
                      connectNulls
                      isAnimationActive={false}
                    />
                  ))}
                </LineChart>
              </ResponsiveContainer>
            </div>
          )}
        </CardContent>
      </Card>

      {request && chartSeries.length > 0 ? (
        <Card className="mt-4 overflow-hidden">
          <CardHeader>
            <CardTitle>测点概览</CardTitle>
            <Badge tone="neutral">{source.toUpperCase()}</Badge>
          </CardHeader>
          <Table>
            <THead>
              <tr>
                <TH>测点</TH>
                <TH>ID</TH>
                <TH className="text-right">数据点</TH>
                <TH className="text-right">最小值</TH>
                <TH className="text-right">最大值</TH>
                <TH className="text-right">最新值</TH>
              </tr>
            </THead>
            <TBody>
              {chartSeries.map((item, index) => {
                const summary = summarize(item)
                return (
                  <TR key={item.tagId}>
                    <TD>
                      <span className="flex items-center gap-2">
                        <span
                          className="size-2 shrink-0 rounded-full"
                          style={{ background: SERIES_COLORS[index % SERIES_COLORS.length] }}
                        />
                        <span className="font-medium">{item.tagName || item.tagId}</span>
                      </span>
                      {item.tagDesc ? (
                        <p className="mt-0.5 text-[11px] text-faint">{item.tagDesc}</p>
                      ) : null}
                    </TD>
                    <TD className="font-mono text-xs text-muted">{item.tagId}</TD>
                    <TD className="text-right text-xs tabular">{summary?.count ?? 0}</TD>
                    <TD className="text-right font-mono text-xs tabular">
                      {summary ? formatNumber(summary.min, 3) : '—'}
                    </TD>
                    <TD className="text-right font-mono text-xs tabular">
                      {summary ? formatNumber(summary.max, 3) : '—'}
                    </TD>
                    <TD className="text-right font-mono text-xs tabular">
                      {summary ? formatNumber(summary.latest, 3) : '—'}
                    </TD>
                  </TR>
                )
              })}
            </TBody>
          </Table>
        </Card>
      ) : null}
    </div>
  )
}
