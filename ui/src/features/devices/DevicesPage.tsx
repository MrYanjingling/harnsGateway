import { useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import {
  ChevronRight,
  Play,
  Plus,
  RotateCw,
  Search,
  Square,
  Trash2,
} from 'lucide-react'
import { toast } from 'sonner'
import { useDeleteDevice, useDevices, useSwitchDeviceStatus } from '@/api/hooks'
import type { DeviceStatus, DeviceStatusAction, DeviceSummary, DeviceType } from '@/api/types'
import { DEVICE_STATUS, PROTOCOLS } from '@/lib/constants'
import { errorMessage } from '@/lib/http'
import { formatDateTimeShort, formatRelative } from '@/lib/format'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card } from '@/components/ui/Card'
import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { EmptyState } from '@/components/ui/EmptyState'
import { Field, Input, Select } from '@/components/ui/Field'
import { TableSkeleton } from '@/components/ui/Skeleton'
import { TBody, TD, TH, THead, TR, Table } from '@/components/ui/Table'
import { DeviceTypeBadge } from '@/components/shared/DeviceTypeBadge'
import { ErrorState } from '@/components/shared/ErrorState'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatusBadge } from '@/components/shared/StatusBadge'

const STATUS_OPTIONS: Array<{ value: DeviceStatus; label: string }> = Object.entries(
  DEVICE_STATUS,
).map(([value, meta]) => ({ value: value as DeviceStatus, label: meta.label }))

export function DevicesPage() {
  const navigate = useNavigate()
  const devices = useDevices()
  const switchStatus = useSwitchDeviceStatus()
  const deleteDevice = useDeleteDevice()

  const [search, setSearch] = useState('')
  const [typeFilter, setTypeFilter] = useState<'all' | DeviceType>('all')
  const [statusFilter, setStatusFilter] = useState<'all' | DeviceStatus>('all')
  const [pendingDelete, setPendingDelete] = useState<DeviceSummary | null>(null)

  const filtered = useMemo(() => {
    const keyword = search.trim().toLowerCase()
    return (devices.data ?? []).filter((device) => {
      if (typeFilter !== 'all' && device.deviceType !== typeFilter) return false
      if (statusFilter !== 'all' && device.collectStatus !== statusFilter) return false
      if (!keyword) return true
      return (
        device.name.toLowerCase().includes(keyword) ||
        device.deviceCode.toLowerCase().includes(keyword) ||
        device.id.toLowerCase().includes(keyword)
      )
    })
  }, [devices.data, search, typeFilter, statusFilter])

  const runAction = (device: DeviceSummary, action: DeviceStatusAction) => {
    const labels: Record<DeviceStatusAction, string> = {
      start: '启动',
      restart: '重启',
      stop: '停止',
    }
    switchStatus.mutate(
      { id: device.id, action },
      {
        onSuccess: () => toast.success(`${device.name} 已${labels[action]}`),
        onError: (error) => toast.error(errorMessage(error)),
      },
    )
  }

  const confirmDelete = () => {
    if (!pendingDelete) return
    deleteDevice.mutate(
      { id: pendingDelete.id, etag: pendingDelete.eTag },
      {
        onSuccess: () => {
          toast.success(`已删除 ${pendingDelete.name}`)
          setPendingDelete(null)
        },
        onError: (error) => toast.error(errorMessage(error)),
      },
    )
  }

  const total = devices.data?.length ?? 0
  const pending = switchStatus.isPending || deleteDevice.isPending

  return (
    <div>
      <PageHeader
        title="设备管理"
        description="管理接入网关的工业设备,支持启动、停止、重启、反控与删除。"
        actions={
          <Button
            variant="primary"
            onClick={() => navigate('/devices/new')}
            icon={<Plus className="size-4" />}
          >
            新建设备
          </Button>
        }
      />

      <Card className="mb-4 flex flex-wrap items-end gap-3 p-4">
        <Field label="搜索" className="min-w-56 flex-1">
          <div className="relative">
            <Search className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-faint" />
            <Input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="设备名称 / 设备编码 / ID"
              className="pl-9"
            />
          </div>
        </Field>
        <Field label="协议" className="w-40">
          <Select
            value={typeFilter}
            onChange={(event) => setTypeFilter(event.target.value as 'all' | DeviceType)}
          >
            <option value="all">全部协议</option>
            {Object.values(PROTOCOLS).map((protocol) => (
              <option key={protocol.type} value={protocol.type}>
                {protocol.label}
              </option>
            ))}
          </Select>
        </Field>
        <Field label="状态" className="w-40">
          <Select
            value={statusFilter}
            onChange={(event) => setStatusFilter(event.target.value as 'all' | DeviceStatus)}
          >
            <option value="all">全部状态</option>
            {STATUS_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </Select>
        </Field>
        <Button
          variant="ghost"
          onClick={() => {
            setSearch('')
            setTypeFilter('all')
            setStatusFilter('all')
          }}
        >
          重置
        </Button>
      </Card>

      <Card className="overflow-hidden">
        <div className="flex items-center justify-between border-b border-border px-5 py-3">
          <p className="text-xs text-muted">
            共 <span className="font-medium text-fg tabular">{filtered.length}</span> 台设备
            {filtered.length !== total ? ` (总计 ${total})` : ''}
          </p>
          {devices.isFetching && !devices.isLoading ? (
            <span className="flex items-center gap-1.5 text-[11px] text-faint">
              <span className="size-1.5 animate-pulse rounded-full bg-primary" />
              正在刷新
            </span>
          ) : null}
        </div>

        {devices.isLoading ? (
          <TableSkeleton rows={6} cols={5} />
        ) : devices.isError ? (
          <div className="p-5">
            <ErrorState error={devices.error} onRetry={devices.refetch} compact />
          </div>
        ) : filtered.length === 0 ? (
          total === 0 ? (
            <EmptyState
              title="还没有设备"
              description="创建 Modbus、OPC UA 或 S7 设备,开始采集工业数据。"
              action={
                <Button variant="primary" onClick={() => navigate('/devices/new')} icon={<Plus className="size-4" />}>
                  新建设备
                </Button>
              }
            />
          ) : (
            <EmptyState
              title="没有匹配的设备"
              description="尝试调整搜索关键词或筛选条件。"
              action={
                <Button
                  variant="outline"
                  onClick={() => {
                    setSearch('')
                    setTypeFilter('all')
                    setStatusFilter('all')
                  }}
                >
                  清除筛选
                </Button>
              }
            />
          )
        ) : (
          <Table>
            <THead>
              <tr>
                <TH>设备</TH>
                <TH>协议 / 型号</TH>
                <TH>状态</TH>
                <TH>设备编码</TH>
                <TH>最后更新</TH>
                <TH className="text-right">操作</TH>
              </tr>
            </THead>
            <TBody>
              {filtered.map((device) => {
                const running =
                  device.collectStatus === 'collecting' ||
                  device.collectStatus === 'collectingError'
                return (
                  <TR key={device.id}>
                    <TD>
                      <div className="flex items-center gap-3">
                        <div className="flex size-8 shrink-0 items-center justify-center rounded-lg border border-border bg-surface-2 text-[10px] font-semibold text-muted">
                          {PROTOCOLS[device.deviceType]?.short ?? '?'}
                        </div>
                        <div className="min-w-0">
                          <Link
                            to={`/devices/${device.id}`}
                            className="block truncate font-medium transition-colors hover:text-primary"
                          >
                            {device.name}
                          </Link>
                          <p className="truncate font-mono text-[10px] text-faint">{device.id}</p>
                        </div>
                      </div>
                    </TD>
                    <TD>
                      <DeviceTypeBadge type={device.deviceType} />
                      <p className="mt-1 text-[11px] text-muted">{device.deviceModel}</p>
                    </TD>
                    <TD>
                      <StatusBadge status={device.collectStatus} />
                    </TD>
                    <TD className="font-mono text-xs text-muted">{device.deviceCode}</TD>
                    <TD>
                      <p className="text-xs text-fg">{formatRelative(device.modTime)}</p>
                      <p className="text-[10px] text-faint">{formatDateTimeShort(device.modTime)}</p>
                    </TD>
                    <TD>
                      <div className="flex items-center justify-end gap-1">
                        {running ? (
                          <>
                            <Button
                              size="icon"
                              variant="ghost"
                              title="重启"
                              disabled={pending}
                              onClick={() => runAction(device, 'restart')}
                            >
                              <RotateCw className="size-3.5" />
                            </Button>
                            <Button
                              size="icon"
                              variant="ghost"
                              title="停止"
                              disabled={pending}
                              onClick={() => runAction(device, 'stop')}
                            >
                              <Square className="size-3.5" />
                            </Button>
                          </>
                        ) : (
                          <Button
                            size="icon"
                            variant="ghost"
                            title="启动"
                            disabled={pending}
                            onClick={() => runAction(device, 'start')}
                          >
                            <Play className="size-3.5" />
                          </Button>
                        )}
                        <Button
                          size="icon"
                          variant="ghost"
                          title="删除"
                          className="text-danger hover:bg-danger/10 hover:text-danger"
                          disabled={pending}
                          onClick={() => setPendingDelete(device)}
                        >
                          <Trash2 className="size-3.5" />
                        </Button>
                        <Link
                          to={`/devices/${device.id}`}
                          title="详情"
                          className="inline-flex size-8 items-center justify-center rounded-lg text-muted transition-colors hover:bg-surface-2 hover:text-fg"
                        >
                          <ChevronRight className="size-4" />
                        </Link>
                      </div>
                    </TD>
                  </TR>
                )
              })}
            </TBody>
          </Table>
        )}
      </Card>

      <div className="mt-3 flex flex-wrap items-center gap-2 text-[11px] text-faint">
        <Badge tone="neutral">提示</Badge>
        <span>启动 / 停止 / 重启为异步操作,状态将在数秒内刷新。</span>
      </div>

      <ConfirmDialog
        open={Boolean(pendingDelete)}
        title="删除设备"
        description={
          pendingDelete
            ? `确定删除「${pendingDelete.name}」吗?删除后该设备将停止采集,此操作不可撤销。`
            : undefined
        }
        confirmText="删除"
        danger
        loading={deleteDevice.isPending}
        onConfirm={confirmDelete}
        onClose={() => setPendingDelete(null)}
      />
    </div>
  )
}
