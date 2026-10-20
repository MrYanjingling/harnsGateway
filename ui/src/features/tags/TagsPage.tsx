import { useMemo, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { LineChart, Search } from 'lucide-react'
import { useTags } from '@/api/hooks'
import { formatRelative } from '@/lib/format'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { EmptyState } from '@/components/ui/EmptyState'
import { Field, Input, Select } from '@/components/ui/Field'
import { TableSkeleton } from '@/components/ui/Skeleton'
import { TBody, TD, TH, THead, TR, Table } from '@/components/ui/Table'
import { ErrorState } from '@/components/shared/ErrorState'
import { PageHeader } from '@/components/shared/PageHeader'

type EnableFilter = 'all' | 'enabled' | 'disabled'

export function TagsPage() {
  const navigate = useNavigate()
  const tags = useTags()
  const [search, setSearch] = useState('')
  const [scadaFilter, setScadaFilter] = useState('all')
  const [typeFilter, setTypeFilter] = useState('all')
  const [enableFilter, setEnableFilter] = useState<EnableFilter>('all')

  const scadaIds = useMemo(
    () => Array.from(new Set((tags.data ?? []).map((tag) => tag.scadaId).filter(Boolean))).sort(),
    [tags.data],
  )
  const types = useMemo(
    () => Array.from(new Set((tags.data ?? []).map((tag) => tag.tagType).filter(Boolean))) as string[],
    [tags.data],
  )

  const filtered = useMemo(() => {
    const keyword = search.trim().toLowerCase()
    return (tags.data ?? []).filter((tag) => {
      if (scadaFilter !== 'all' && tag.scadaId !== scadaFilter) return false
      if (typeFilter !== 'all' && tag.tagType !== typeFilter) return false
      if (enableFilter === 'enabled' && tag.enable !== 1) return false
      if (enableFilter === 'disabled' && tag.enable === 1) return false
      if (!keyword) return true
      return (
        (tag.tagId ?? '').toLowerCase().includes(keyword) ||
        (tag.tagName ?? '').toLowerCase().includes(keyword) ||
        (tag.tagDesc ?? '').toLowerCase().includes(keyword)
      )
    })
  }, [tags.data, search, scadaFilter, typeFilter, enableFilter])

  const reset = () => {
    setSearch('')
    setScadaFilter('all')
    setTypeFilter('all')
    setEnableFilter('all')
  }

  const enabledCount = (tags.data ?? []).filter((tag) => tag.enable === 1).length

  return (
    <div>
      <PageHeader
        title="测点管理"
        description="查看数据服务同步的测点(SCADA Tag),可直接跳转查询历史时序。"
      />

      <Card className="mb-4 flex flex-wrap items-end gap-3 p-4">
        <Field label="搜索" className="min-w-56 flex-1">
          <div className="relative">
            <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-faint" />
            <Input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="测点 ID / 名称 / 描述"
              className="pl-9"
            />
          </div>
        </Field>
        <Field label="SCADA" className="w-40">
          <Select value={scadaFilter} onChange={(event) => setScadaFilter(event.target.value)}>
            <option value="all">全部</option>
            {scadaIds.map((id) => (
              <option key={id} value={id}>
                {id}
              </option>
            ))}
          </Select>
        </Field>
        <Field label="类型" className="w-36">
          <Select value={typeFilter} onChange={(event) => setTypeFilter(event.target.value)}>
            <option value="all">全部类型</option>
            {types.map((type) => (
              <option key={type} value={type}>
                {type}
              </option>
            ))}
          </Select>
        </Field>
        <Field label="状态" className="w-36">
          <Select
            value={enableFilter}
            onChange={(event) => setEnableFilter(event.target.value as EnableFilter)}
          >
            <option value="all">全部状态</option>
            <option value="enabled">已启用</option>
            <option value="disabled">已停用</option>
          </Select>
        </Field>
        <Button variant="ghost" onClick={reset}>
          重置
        </Button>
      </Card>

      <Card className="overflow-hidden">
        <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border px-5 py-3">
          <p className="text-xs text-muted">
            共 <span className="font-medium text-fg tabular">{filtered.length}</span> 个测点
            {filtered.length !== (tags.data?.length ?? 0)
              ? ` (总计 ${tags.data?.length ?? 0})`
              : ''}
          </p>
          <p className="text-[11px] text-faint">
            启用 <span className="text-success tabular">{enabledCount}</span> · 停用{' '}
            <span className="tabular">{(tags.data?.length ?? 0) - enabledCount}</span>
          </p>
        </div>

        {tags.isLoading ? (
          <TableSkeleton rows={6} cols={5} />
        ) : tags.isError ? (
          <div className="p-5">
            <ErrorState error={tags.error} onRetry={tags.refetch} compact />
          </div>
        ) : filtered.length === 0 ? (
          <EmptyState
            title={tags.data?.length ? '没有匹配的测点' : '没有测点'}
            description={
              tags.data?.length ? '尝试调整搜索或筛选条件。' : '数据服务尚未同步任何测点。'
            }
            action={
              tags.data?.length ? (
                <Button variant="outline" onClick={reset}>
                  清除筛选
                </Button>
              ) : undefined
            }
          />
        ) : (
          <Table>
            <THead>
              <tr>
                <TH>测点</TH>
                <TH>描述</TH>
                <TH>类型 / 属性</TH>
                <TH>SCADA</TH>
                <TH>状态</TH>
                <TH>最后同步</TH>
                <TH className="text-right">操作</TH>
              </tr>
            </THead>
            <TBody>
              {filtered.map((tag) => (
                <TR key={tag.id}>
                  <TD>
                    <p className="font-medium">{tag.tagName || tag.tagId || '未命名测点'}</p>
                    <p className="font-mono text-[11px] text-faint">{tag.tagId || tag.id}</p>
                  </TD>
                  <TD className="max-w-56 truncate text-xs text-muted" title={tag.tagDesc ?? ''}>
                    {tag.tagDesc || '—'}
                  </TD>
                  <TD className="text-xs text-muted">
                    {tag.tagType || '—'}
                    {tag.tagAttribute ? (
                      <span className="ml-1 font-mono text-[10px] text-faint">
                        / {tag.tagAttribute}
                      </span>
                    ) : null}
                  </TD>
                  <TD className="font-mono text-xs text-muted">{tag.scadaId}</TD>
                  <TD>
                    {tag.enable === 1 ? (
                      <Badge tone="success" dot>
                        启用
                      </Badge>
                    ) : (
                      <Badge tone="neutral" dot>
                        停用
                      </Badge>
                    )}
                  </TD>
                  <TD className="text-xs text-muted">
                    {tag.lastSynTime ? formatRelative(tag.lastSynTime) : '—'}
                  </TD>
                  <TD>
                    <div className="flex justify-end">
                      <Button
                        size="sm"
                        variant="outline"
                        icon={<LineChart className="size-3.5" />}
                        onClick={() =>
                          navigate(
                            `/data/timeseries?select=${encodeURIComponent(tag.tagId ?? '')}`,
                          )
                        }
                      >
                        时序
                      </Button>
                    </div>
                  </TD>
                </TR>
              ))}
            </TBody>
          </Table>
        )}
      </Card>
    </div>
  )
}
