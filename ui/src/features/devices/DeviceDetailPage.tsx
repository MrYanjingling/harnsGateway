import { useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import {
  Pencil,
  Play,
  RotateCw,
  SlidersHorizontal,
  Square,
  Trash2,
} from 'lucide-react'
import { toast } from 'sonner'
import { useDeleteDevice, useDevice, useSwitchDeviceStatus } from '@/api/hooks'
import type { DeviceStatusAction } from '@/api/types'
import { DEVICE_STATUS, PROTOCOLS } from '@/lib/constants'
import { errorMessage } from '@/lib/http'
import { formatDateTime } from '@/lib/format'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { ConfirmDialog } from '@/components/ui/ConfirmDialog'
import { CopyButton } from '@/components/ui/CopyButton'
import { EmptyState } from '@/components/ui/EmptyState'
import { PageLoading } from '@/components/ui/Spinner'
import { TBody, TD, TH, THead, TR, Table } from '@/components/ui/Table'
import { DeviceTypeBadge } from '@/components/shared/DeviceTypeBadge'
import { ErrorState } from '@/components/shared/ErrorState'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { ControlDialog } from './ControlDialog'

function displayValue(value: unknown): string {
  if (value === undefined || value === null || value === '') return '—'
  if (typeof value === 'boolean') return value ? 'true' : 'false'
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

export function DeviceDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const device = useDevice(id)
  const switchStatus = useSwitchDeviceStatus()
  const deleteDevice = useDeleteDevice()
  const [controlOpen, setControlOpen] = useState(false)
  const [controlTarget, setControlTarget] = useState<string | undefined>(undefined)
  const [confirmDelete, setConfirmDelete] = useState(false)

  const detail = device.data

  const rows = useMemo(() => {
    if (!detail) return []
    type Row = {
      name: string
      dataType: string
      address: string
      detail: string
      accessMode: string
      value: unknown
    }
    if (detail.deviceType === 'modbus') {
      return detail.variables.map<Row>((variable) => ({
        name: variable.name,
        dataType: variable.dataType,
        address: String(variable.address),
        detail: `FC${variable.functionCode}${variable.dataType === 'bool' ? ` · bit ${variable.bits}` : ''} · ×${variable.rate} · 数量 ${variable.amount}`,
        accessMode: variable.accessMode,
        value: variable.value,
      }))
    }
    if (detail.deviceType === 'opcUa') {
      return detail.variables.map<Row>((variable) => ({
        name: variable.name,
        dataType: variable.dataType,
        address: displayValue(variable.address),
        detail: `ns=${variable.namespace}`,
        accessMode: variable.accessMode,
        value: variable.value,
      }))
    }
    return detail.variables.map<Row>((variable) => ({
      name: variable.name,
      dataType: variable.dataType,
      address: variable.address,
      detail: `×${variable.rate}`,
      accessMode: variable.accessMode,
      value: variable.value,
    }))
  }, [detail])

  if (device.isLoading) {
    return <PageLoading label="正在加载设备详情…" />
  }

  if (device.isError || !detail) {
    return (
      <div>
        <PageHeader backTo="/devices" backLabel="返回设备列表" title="设备详情" />
        <ErrorState error={device.error} onRetry={device.refetch} />
      </div>
    )
  }

  const protocol = PROTOCOLS[detail.deviceType]
  const running =
    detail.collectStatus === 'collecting' || detail.collectStatus === 'collectingError'
  const writableCount = rows.filter((row) => row.accessMode === 'rw').length

  const runAction = (action: DeviceStatusAction) => {
    const labels: Record<DeviceStatusAction, string> = {
      start: '启动',
      restart: '重启',
      stop: '停止',
    }
    switchStatus.mutate(
      { id: detail.id, action },
      {
        onSuccess: () => toast.success(`${detail.name} 已${labels[action]}`),
        onError: (error) => toast.error(errorMessage(error)),
      },
    )
  }

  const openControl = (name?: string) => {
    setControlTarget(name)
    setControlOpen(true)
  }

  const confirmRemove = () => {
    deleteDevice.mutate(
      { id: detail.id, etag: detail.eTag },
      {
        onSuccess: () => {
          toast.success(`已删除 ${detail.name}`)
          navigate('/devices')
        },
        onError: (error) => toast.error(errorMessage(error)),
      },
    )
  }

  return (
    <div>
      <PageHeader
        backTo="/devices"
        backLabel="返回设备列表"
        title={
          <span className="flex flex-wrap items-center gap-3">
            {detail.name}
            <StatusBadge status={detail.collectStatus} />
          </span>
        }
        description={
          <span className="flex flex-wrap items-center gap-2">
            <DeviceTypeBadge type={detail.deviceType} />
            <span className="font-mono text-xs">{detail.deviceModel}</span>
            <span className="text-faint">·</span>
            <span className="font-mono text-xs">{detail.deviceCode}</span>
            <span className="text-faint">·</span>
            <span>{DEVICE_STATUS[detail.collectStatus]?.description}</span>
          </span>
        }
        actions={
          <>
            {writableCount > 0 ? (
              <Button
                variant="secondary"
                icon={<SlidersHorizontal className="size-4" />}
                onClick={() => openControl()}
              >
                反控
              </Button>
            ) : null}
            <Button
              variant="outline"
              icon={<Pencil className="size-4" />}
              onClick={() => navigate(`/devices/${detail.id}/edit`)}
            >
              编辑
            </Button>
            {running ? (
              <>
                <Button
                  variant="outline"
                  icon={<RotateCw className="size-4" />}
                  disabled={switchStatus.isPending}
                  onClick={() => runAction('restart')}
                >
                  重启
                </Button>
                <Button
                  variant="outline"
                  icon={<Square className="size-4" />}
                  disabled={switchStatus.isPending}
                  onClick={() => runAction('stop')}
                >
                  停止
                </Button>
              </>
            ) : (
              <Button
                variant="primary"
                icon={<Play className="size-4" />}
                disabled={switchStatus.isPending}
                onClick={() => runAction('start')}
              >
                启动
              </Button>
            )}
            <Button
              variant="outline"
              className="text-danger hover:bg-danger/10 hover:text-danger"
              icon={<Trash2 className="size-4" />}
              disabled={deleteDevice.isPending}
              onClick={() => setConfirmDelete(true)}
            >
              删除
            </Button>
          </>
        }
      />

      <div className="grid gap-4 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <div>
              <CardTitle>设备信息</CardTitle>
              <CardDescription>标识、采集参数与版本信息</CardDescription>
            </div>
            {device.isFetching ? (
              <span className="flex items-center gap-1.5 text-[11px] text-faint">
                <span className="size-1.5 animate-pulse rounded-full bg-primary" />
                同步中
              </span>
            ) : null}
          </CardHeader>
          <CardContent>
            <dl className="grid gap-x-6 gap-y-4 sm:grid-cols-2">
              <InfoItem label="设备编码" value={detail.deviceCode} mono />
              <InfoItem
                label="协议 / 型号"
                value={`${protocol.label} / ${detail.deviceModel}`}
              />
              <InfoItem label="采集周期" value={`${detail.collectorCycle} 秒`} />
              <InfoItem
                label="变量间隔"
                value={detail.variableInterval > 0 ? `${detail.variableInterval} 毫秒` : '不等待'}
              />
              <InfoItem label="Topic" value={detail.topic || '默认规则'} mono />
              <InfoItem label="变量数量" value={`${rows.length} 个`} />
              <InfoItem
                label="设备 ID"
                value={detail.id}
                mono
                copyValue={detail.id}
              />
              <InfoItem label="eTag" value={detail.eTag} mono copyValue={detail.eTag} />
              <InfoItem label="最后修改" value={formatDateTime(detail.modTime)} />
            </dl>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div>
              <CardTitle>连接参数</CardTitle>
              <CardDescription>{protocol.description}</CardDescription>
            </div>
          </CardHeader>
          <CardContent>
            <dl className="space-y-4">
              {detail.deviceType === 'modbus' ? (
                <>
                  <InfoItem
                    label={detail.deviceModel === 'modbusRtu' ? '串口设备' : '设备地址'}
                    value={detail.address.location}
                    mono
                  />
                  {detail.deviceModel === 'modbusRtu' ? (
                    <InfoItem
                      label="串口参数"
                      value={`${detail.address.option?.baudRate ?? '—'} baud · ${detail.address.option?.dataBits ?? '—'} 数据位 · ${detail.address.option?.parity ?? '—'} · ${detail.address.option?.stopBits ?? '—'} 停止位`}
                    />
                  ) : (
                    <InfoItem label="端口" value={String(detail.address.option?.port ?? '—')} mono />
                  )}
                  <InfoItem label="从站地址" value={String(detail.slave)} />
                  <InfoItem label="内存布局" value={detail.memoryLayout} mono />
                  <InfoItem label="起始偏移" value={String(detail.positionAddress)} />
                </>
              ) : null}

              {detail.deviceType === 'opcUa' ? (
                <>
                  <InfoItem label="Endpoint" value={detail.address.location} mono />
                  <InfoItem label="端口" value={String(detail.address.option?.port ?? '—')} mono />
                  <InfoItem
                    label="用户名"
                    value={detail.address.option?.username || '匿名'}
                    mono
                  />
                  <InfoItem
                    label="密码"
                    value={detail.address.option?.password ? '••••••••' : '未设置'}
                    mono
                  />
                </>
              ) : null}

              {detail.deviceType === 's7' ? (
                <>
                  <InfoItem label="PLC 地址" value={detail.address.location} mono />
                  <InfoItem label="端口" value={String(detail.address.option?.port ?? '—')} mono />
                  <InfoItem
                    label="Rack / Slot"
                    value={`${detail.address.option?.rack ?? 0} / ${detail.address.option?.slot ?? 1}`}
                    mono
                  />
                </>
              ) : null}
            </dl>
          </CardContent>
        </Card>
      </div>

      <Card className="mt-4 overflow-hidden">
        <CardHeader>
          <div>
            <CardTitle>采集变量</CardTitle>
            <CardDescription>
              当前值来自设备最近一次采集;标记为 rw 的变量支持反控写入
            </CardDescription>
          </div>
          <Badge tone="neutral">{rows.length} 个变量</Badge>
        </CardHeader>

        {rows.length === 0 ? (
          <EmptyState
            title="没有采集变量"
            description="编辑设备并添加至少一个变量后,设备才能开始采集。"
            action={
              <Button
                variant="outline"
                onClick={() => navigate(`/devices/${detail.id}/edit`)}
                icon={<Pencil className="size-4" />}
              >
                添加变量
              </Button>
            }
          />
        ) : (
          <Table>
            <THead>
              <tr>
                <TH>变量名</TH>
                <TH>类型</TH>
                <TH>地址</TH>
                <TH>采集参数</TH>
                <TH>当前值</TH>
                <TH>读写</TH>
                <TH className="text-right">操作</TH>
              </tr>
            </THead>
            <TBody>
              {rows.map((row) => (
                <TR key={row.name}>
                  <TD className="font-medium">{row.name}</TD>
                  <TD>
                    <code className="rounded bg-surface-2 px-1.5 py-0.5 font-mono text-[11px] text-muted">
                      {row.dataType}
                    </code>
                  </TD>
                  <TD className="font-mono text-xs text-muted">{row.address}</TD>
                  <TD className="text-xs text-muted">{row.detail}</TD>
                  <TD>
                    <span className="font-mono text-xs text-fg">{displayValue(row.value)}</span>
                  </TD>
                  <TD>
                    {row.accessMode === 'rw' ? (
                      <Badge tone="warning">rw</Badge>
                    ) : (
                      <Badge tone="neutral">r</Badge>
                    )}
                  </TD>
                  <TD>
                    <div className="flex justify-end">
                      {row.accessMode === 'rw' ? (
                        <Button
                          size="sm"
                          variant="outline"
                          onClick={() => openControl(row.name)}
                          icon={<SlidersHorizontal className="size-3.5" />}
                        >
                          下发
                        </Button>
                      ) : (
                        <span className="text-[11px] text-faint">只读</span>
                      )}
                    </div>
                  </TD>
                </TR>
              ))}
            </TBody>
          </Table>
        )}
      </Card>

      <p className="mt-3 text-[11px] text-faint">
        设备详情每 5 秒自动刷新。修改设备时若提示版本冲突,请返回详情页重新加载后再编辑。
      </p>

      <ControlDialog
        open={controlOpen}
        device={detail}
        only={controlTarget}
        onClose={() => setControlOpen(false)}
      />

      <ConfirmDialog
        open={confirmDelete}
        title="删除设备"
        description={`确定删除「${detail.name}」吗?删除后该设备将停止采集,此操作不可撤销。`}
        confirmText="删除"
        danger
        loading={deleteDevice.isPending}
        onConfirm={confirmRemove}
        onClose={() => setConfirmDelete(false)}
      />
    </div>
  )
}

function InfoItem({
  label,
  value,
  mono = false,
  copyValue,
}: {
  label: string
  value: string
  mono?: boolean
  copyValue?: string
}) {
  return (
    <div className="min-w-0">
      <dt className="text-[11px] text-faint">{label}</dt>
      <dd className="mt-1 flex items-center gap-1">
        <span
          className={
            mono
              ? 'min-w-0 truncate font-mono text-xs text-fg'
              : 'min-w-0 truncate text-sm text-fg'
          }
          title={value}
        >
          {value}
        </span>
        {copyValue ? <CopyButton value={copyValue} /> : null}
      </dd>
    </div>
  )
}
