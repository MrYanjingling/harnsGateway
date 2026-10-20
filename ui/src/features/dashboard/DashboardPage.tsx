import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  ArrowRight,
  Cpu,
  Eye,
  EyeOff,
  HardDrive,
  MemoryStick,
  RefreshCw,
  Router,
  Server,
  ShieldAlert,
  Activity,
} from 'lucide-react'
import { useGatewayCpu, useGatewayDisk, useGatewayMem, useGatewayMeta, useDevices } from '@/api/hooks'
import type { DeviceStatus } from '@/api/types'
import { DEVICE_STATUS } from '@/lib/constants'
import { formatRelative, maskSecret, percentValue } from '@/lib/format'
import { cn } from '@/lib/cn'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { CopyButton } from '@/components/ui/CopyButton'
import { EmptyState } from '@/components/ui/EmptyState'
import { ProgressBar } from '@/components/ui/ProgressBar'
import { Skeleton, TableSkeleton } from '@/components/ui/Skeleton'
import { TD, TH, THead, TBody, TR, Table } from '@/components/ui/Table'
import { GaugeRing, toneForPercent } from '@/components/charts/GaugeRing'
import { PageHeader } from '@/components/shared/PageHeader'
import { StatCard } from '@/components/shared/StatCard'
import { StatusBadge } from '@/components/shared/StatusBadge'
import { DeviceTypeBadge } from '@/components/shared/DeviceTypeBadge'
import { ErrorState } from '@/components/shared/ErrorState'
import { toast } from 'sonner'

const STATUS_ORDER: DeviceStatus[] = [
  'collecting',
  'collectingError',
  'unconnected',
  'emptyVariable',
  'stopped',
  'error',
]

const STATUS_BAR: Record<DeviceStatus, string> = {
  collecting: 'bg-emerald-500',
  collectingError: 'bg-amber-500',
  unconnected: 'bg-zinc-400',
  emptyVariable: 'bg-violet-500',
  stopped: 'bg-slate-400',
  error: 'bg-rose-500',
}

export function DashboardPage() {
  const meta = useGatewayMeta()
  const cpu = useGatewayCpu()
  const mem = useGatewayMem()
  const disk = useGatewayDisk()
  const devices = useDeviceList()
  const [secretVisible, setSecretVisible] = useState(false)

  const list = devices.data ?? []
  const counts = useMemo(() => {
    const result = Object.fromEntries(STATUS_ORDER.map((s) => [s, 0])) as Record<DeviceStatus, number>
    for (const device of list) {
      if (device.collectStatus in result) result[device.collectStatus] += 1
      else result.error += 1
    }
    return result
  }, [list])

  const total = list.length
  const healthy = counts.collecting
  const abnormal = counts.collectingError + counts.error
  const inactive = counts.stopped + counts.unconnected + counts.emptyVariable

  const recent = useMemo(
    () =>
      [...list]
        .sort((a, b) => new Date(b.modTime).getTime() - new Date(a.modTime).getTime())
        .slice(0, 5),
    [list],
  )

  const refreshAll = () => {
    void meta.refetch()
    void cpu.refetch()
    void mem.refetch()
    void disk.refetch()
    void devices.refetch()
    toast.success('已刷新')
  }

  const cpuData = cpu.data?.cpus
  const memData = mem.data?.mem
  const diskData = disk.data?.disk ?? {}

  return (
    <div className="space-y-6">
      <PageHeader
        title="概览"
        description="网关运行状态、设备采集情况与资源占用一览。"
        actions={
          <Button variant="outline" onClick={refreshAll} icon={<RefreshCw className="size-4" />}>
            刷新
          </Button>
        }
      />

      {/* 网关身份 */}
      <Card className="relative overflow-hidden">
        <div
          aria-hidden
          className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top_right,var(--primary-soft),transparent_60%)] opacity-70"
        />
        <div className="relative flex flex-col gap-6 p-5 md:flex-row md:items-start md:justify-between">
          <div className="min-w-0">
            <div className="flex flex-wrap items-center gap-2">
              <Badge tone={meta.isError ? 'danger' : 'success'} dot>
                {meta.isError ? '网关不可用' : meta.isSuccess ? '网关运行中' : '连接中'}
              </Badge>
              <Badge tone="neutral">边缘节点</Badge>
            </div>

            {meta.isLoading ? (
              <div className="mt-3 space-y-2">
                <Skeleton className="h-7 w-56" />
                <Skeleton className="h-4 w-72" />
              </div>
            ) : meta.isError ? (
              <div className="mt-3">
                <h2 className="text-lg font-semibold">无法读取网关信息</h2>
                <p className="mt-1 text-xs text-muted">
                  请确认网关服务已启动 (默认端口 32200),或在 `.env` 中开启 `VITE_USE_MOCK`。
                </p>
              </div>
            ) : (
              <>
                <h2 className="mt-3 text-2xl font-semibold tracking-tight">{meta.data?.name}</h2>
                <div className="mt-1.5 flex flex-wrap items-center gap-2 text-xs text-muted">
                  <span className="font-mono">ID {meta.data?.id}</span>
                  <CopyButton value={meta.data?.id ?? ''} />
                </div>
              </>
            )}
          </div>

          <div className="grid shrink-0 grid-cols-2 gap-x-8 gap-y-3 text-xs md:min-w-72">
            <div>
              <p className="text-faint">版本号 (ETag)</p>
              <p className="mt-1 font-mono text-fg">{meta.data?.eTag ?? '—'}</p>
            </div>
            <div>
              <p className="text-faint">最近更新</p>
              <p className="mt-1 text-fg">{formatRelative(meta.data?.modTime)}</p>
            </div>
            <div className="col-span-2">
              <p className="text-faint">网关密钥</p>
              <div className="mt-1 flex items-center gap-1.5">
                <span className="truncate font-mono text-fg">
                  {meta.data
                    ? secretVisible
                      ? meta.data.secret || '未设置'
                      : maskSecret(meta.data.secret)
                    : '—'}
                </span>
                {meta.data?.secret ? (
                  <>
                    <button
                      type="button"
                      onClick={() => setSecretVisible((v) => !v)}
                      className="rounded p-1 text-faint hover:bg-surface-2 hover:text-fg"
                      aria-label={secretVisible ? '隐藏密钥' : '显示密钥'}
                    >
                      {secretVisible ? <EyeOff className="size-3.5" /> : <Eye className="size-3.5" />}
                    </button>
                    <CopyButton value={meta.data.secret} label="密钥已复制" />
                  </>
                ) : null}
              </div>
            </div>
          </div>
        </div>
      </Card>

      {/* 设备统计 */}
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="设备总数"
          value={devices.isLoading ? '—' : total}
          hint="已接入网关的采集设备"
          icon={<Router className="size-4" />}
        />
        <StatCard
          label="采集中"
          value={devices.isLoading ? '—' : healthy}
          hint={total ? `占比 ${Math.round((healthy / total) * 100)}%` : '暂无设备'}
          icon={<Activity className="size-4" />}
          footer={<ProgressBar value={total ? (healthy / total) * 100 : 0} tone="success" />}
        />
        <StatCard
          label="采集异常"
          value={devices.isLoading ? '—' : abnormal}
          hint="需要关注连接或采集错误"
          icon={<ShieldAlert className="size-4" />}
          footer={<ProgressBar value={total ? (abnormal / total) * 100 : 0} tone="warning" />}
        />
        <StatCard
          label="未运行"
          value={devices.isLoading ? '—' : inactive}
          hint="已停止、未连接或无变量"
          icon={<Cpu className="size-4" />}
          footer={<ProgressBar value={total ? (inactive / total) * 100 : 0} tone="primary" />}
        />
      </div>

      <div className="grid gap-4 lg:grid-cols-5">
        {/* 资源 */}
        <Card className="lg:col-span-3">
          <CardHeader>
            <div>
              <CardTitle>资源使用</CardTitle>
              <CardDescription>每 10 秒自动刷新</CardDescription>
            </div>
            <Server className="size-4 text-faint" />
          </CardHeader>
          <CardContent className="space-y-5">
            <div className="grid grid-cols-2 gap-4 rounded-xl border border-border bg-surface-2/50 p-4">
              <div>
                <p className="mb-3 flex items-center gap-1.5 text-xs font-medium text-muted">
                  <Cpu className="size-3.5" /> CPU
                </p>
                {cpu.isError ? (
                  <p className="text-xs text-danger">CPU 数据不可用</p>
                ) : (
                  <div className="flex items-end justify-around">
                    {(['1min', '5min', '15min'] as const).map((key) => (
                      <GaugeRing
                        key={key}
                        value={percentValue(cpuData?.[key])}
                        label={key.replace('min', ' 分钟')}
                        size={78}
                        strokeWidth={7}
                        tone={toneForPercent(percentValue(cpuData?.[key]))}
                      />
                    ))}
                  </div>
                )}
              </div>
              <div className="border-l border-border pl-4">
                <p className="mb-3 flex items-center gap-1.5 text-xs font-medium text-muted">
                  <MemoryStick className="size-3.5" /> 内存
                </p>
                {mem.isError || !memData ? (
                  <p className="text-xs text-danger">内存数据不可用</p>
                ) : (
                  <div className="flex flex-col items-center gap-2">
                    <GaugeRing
                      value={percentValue(memData.UsedPercent)}
                      label={`${memData.Used} / ${memData.Total}`}
                      size={92}
                      tone={toneForPercent(percentValue(memData.UsedPercent))}
                    />
                  </div>
                )}
              </div>
            </div>

            <div>
              <p className="mb-3 flex items-center gap-1.5 text-xs font-medium text-muted">
                <HardDrive className="size-3.5" /> 磁盘
              </p>
              {disk.isError ? (
                <p className="text-xs text-danger">磁盘数据不可用</p>
              ) : Object.keys(diskData).length === 0 ? (
                <p className="text-xs text-faint">暂无挂载点信息</p>
              ) : (
                <div className="space-y-3">
                  {Object.entries(diskData).map(([mount, info]) => {
                    const percent = percentValue(info.UsedPercent)
                    return (
                      <div key={mount}>
                        <div className="mb-1.5 flex items-baseline justify-between gap-3 text-xs">
                          <span className="font-mono text-fg">{mount}</span>
                          <span className="text-muted tabular">
                            {info.Used} / {info.Total}
                            <span className="ml-2 font-medium text-fg">{info.UsedPercent}</span>
                          </span>
                        </div>
                        <ProgressBar
                          value={percent}
                          tone={
                            percent >= 90 ? 'danger' : percent >= 75 ? 'warning' : 'success'
                          }
                        />
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          </CardContent>
        </Card>

        {/* 状态分布 */}
        <Card className="lg:col-span-2">
          <CardHeader>
            <div>
              <CardTitle>设备状态分布</CardTitle>
              <CardDescription>按采集状态统计</CardDescription>
            </div>
            <Link
              to="/devices"
              className="inline-flex items-center gap-1 text-xs text-muted transition-colors hover:text-fg"
            >
              查看全部 <ArrowRight className="size-3" />
            </Link>
          </CardHeader>
          <CardContent>
            {devices.isLoading ? (
              <div className="space-y-3">
                <Skeleton className="h-3 w-full" />
                {Array.from({ length: 4 }).map((_, i) => (
                  <Skeleton key={i} className="h-5 w-full" />
                ))}
              </div>
            ) : devices.isError ? (
              <ErrorState error={devices.error} onRetry={devices.refetch} compact />
            ) : total === 0 ? (
              <EmptyState title="还没有设备" description="创建第一台设备后即可查看状态分布。" className="py-8" />
            ) : (
              <>
                <div className="flex h-2.5 w-full overflow-hidden rounded-full bg-surface-3">
                  {STATUS_ORDER.map((status) =>
                    counts[status] > 0 ? (
                      <div
                        key={status}
                        className={cn('h-full transition-all', STATUS_BAR[status])}
                        style={{ width: `${(counts[status] / total) * 100}%` }}
                        title={`${DEVICE_STATUS[status].label} ${counts[status]}`}
                      />
                    ) : null,
                  )}
                </div>
                <ul className="mt-4 space-y-2.5">
                  {STATUS_ORDER.map((status) => (
                    <li key={status} className="flex items-center gap-2 text-xs">
                      <span className={cn('size-2 rounded-full', STATUS_BAR[status])} />
                      <span className="text-muted">{DEVICE_STATUS[status].label}</span>
                      <span className="ml-auto font-medium tabular">{counts[status]}</span>
                      <span className="w-10 text-right text-faint tabular">
                        {total ? Math.round((counts[status] / total) * 100) : 0}%
                      </span>
                    </li>
                  ))}
                </ul>
              </>
            )}
          </CardContent>
        </Card>
      </div>

      {/* 最近设备 */}
      <Card>
        <CardHeader>
          <div>
            <CardTitle>最近更新设备</CardTitle>
            <CardDescription>按最后修改时间排序</CardDescription>
          </div>
          <Link
            to="/devices"
            className="inline-flex items-center gap-1 text-xs text-muted transition-colors hover:text-fg"
          >
            设备管理 <ArrowRight className="size-3" />
          </Link>
        </CardHeader>
        {devices.isLoading ? (
          <TableSkeleton rows={4} cols={4} />
        ) : devices.isError ? (
          <div className="p-5">
            <ErrorState error={devices.error} onRetry={devices.refetch} compact />
          </div>
        ) : recent.length === 0 ? (
          <EmptyState
            title="暂无设备"
            description="创建 Modbus、OPC UA 或 S7 设备,开始采集数据。"
            action={
              <Link
                to="/devices/new"
                className="inline-flex h-9.5 items-center rounded-lg bg-primary px-4 text-sm font-medium text-primary-fg hover:bg-primary-hover"
              >
                新建设备
              </Link>
            }
          />
        ) : (
          <Table>
            <THead>
              <tr>
                <TH>设备名称</TH>
                <TH>协议</TH>
                <TH>状态</TH>
                <TH>设备编码</TH>
                <TH className="text-right">更新时间</TH>
              </tr>
            </THead>
            <TBody>
              {recent.map((device) => (
                <TR key={device.id}>
                  <TD>
                    <Link
                      to={`/devices/${device.id}`}
                      className="font-medium transition-colors hover:text-primary"
                    >
                      {device.name}
                    </Link>
                  </TD>
                  <TD>
                    <DeviceTypeBadge type={device.deviceType} />
                  </TD>
                  <TD>
                    <StatusBadge status={device.collectStatus} />
                  </TD>
                  <TD className="font-mono text-xs text-muted">{device.deviceCode}</TD>
                  <TD className="text-right text-xs text-muted">{formatRelative(device.modTime)}</TD>
                </TR>
              ))}
            </TBody>
          </Table>
        )}
      </Card>
    </div>
  )
}

/** 独立 hook 便于后续替换为带过滤条件的列表查询。 */
function useDeviceList() {
  return useDevices()
}
