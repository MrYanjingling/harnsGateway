import type { ReactNode } from 'react'
import { Check, Database, Moon, RefreshCw, Server, Sun } from 'lucide-react'
import { useGatewayMeta, useMetrics } from '@/api/hooks'
import { useTheme } from '@/providers/ThemeProvider'
import { cn } from '@/lib/cn'
import { maskSecret } from '@/lib/format'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { CopyButton } from '@/components/ui/CopyButton'
import { PageHeader } from '@/components/shared/PageHeader'

const env = import.meta.env

const gatewayBase = env.VITE_GATEWAY_BASE || '/api/v1'
const dataBase = env.VITE_DATA_BASE || '/api/v1'
const gatewayTarget = env.VITE_GATEWAY_TARGET || 'http://127.0.0.1:32200'
const dataTarget = env.VITE_DATA_TARGET || 'http://127.0.0.1:32201'
const useMock = env.VITE_USE_MOCK === 'true'

export function SettingsPage() {
  const { theme, setTheme } = useTheme()
  const gateway = useGatewayMeta()
  const dataService = useMetrics()

  return (
    <div>
      <PageHeader
        title="设置"
        description="外观、运行环境与后端连接信息。所有配置均通过构建期环境变量注入。"
      />

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <div>
              <CardTitle>外观</CardTitle>
              <CardDescription>主题保存在浏览器 localStorage,首屏无闪烁</CardDescription>
            </div>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 gap-3">
              <ThemeOption
                label="浅色"
                active={theme === 'light'}
                onClick={() => setTheme('light')}
                icon={<Sun className="size-4" />}
                swatches={['bg-white', 'bg-zinc-100', 'bg-emerald-500']}
              />
              <ThemeOption
                label="深色"
                active={theme === 'dark'}
                onClick={() => setTheme('dark')}
                icon={<Moon className="size-4" />}
                swatches={['bg-zinc-900', 'bg-zinc-800', 'bg-emerald-400']}
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div>
              <CardTitle>运行环境</CardTitle>
              <CardDescription>由 ui/.env.local 中的 VITE_* 变量控制</CardDescription>
            </div>
            {useMock ? <Badge tone="warning">Mock 模式</Badge> : <Badge tone="success">真实后端</Badge>}
          </CardHeader>
          <CardContent className="space-y-3">
            <EnvRow label="模式" value={useMock ? '本地 Mock 中间件' : 'Vite Proxy 转发'} />
            <EnvRow label="网关服务基址" value={gatewayBase} />
            <EnvRow label="数据服务基址" value={dataBase} />
            <EnvRow label="网关代理目标" value={useMock ? '不启用' : gatewayTarget} />
            <EnvRow label="数据代理目标" value={useMock ? '不启用' : dataTarget} />
            <p className="text-[11px] leading-relaxed text-faint">
              生产构建不再包含 Vite proxy,需要由 nginx 等反向代理把两套 API 转发到对应端口。
            </p>
          </CardContent>
        </Card>
      </div>

      <Card className="mt-4">
        <CardHeader>
          <div>
            <CardTitle>服务连接</CardTitle>
            <CardDescription>检查两个 Go 服务是否可访问</CardDescription>
          </div>
          <Button
            size="sm"
            variant="outline"
            icon={<RefreshCw className="size-3.5" />}
            onClick={() => {
              void gateway.refetch()
              void dataService.refetch()
            }}
          >
            重新检查
          </Button>
        </CardHeader>
        <CardContent className="space-y-3">
          <ServiceRow
            icon={<Server className="size-4" />}
            name="网关服务"
            endpoint={`${gatewayBase}/devices · 默认端口 32200`}
            loading={gateway.isLoading}
            online={gateway.isSuccess}
          >
            {gateway.data ? (
              <div className="mt-2 grid gap-x-6 gap-y-1 text-[11px] text-muted sm:grid-cols-2">
                <span>
                  名称 <span className="text-fg">{gateway.data.name}</span>
                </span>
                <span className="truncate">
                  Secret <span className="font-mono text-fg">{maskSecret(gateway.data.secret)}</span>
                </span>
                <span className="truncate">
                  ID <span className="font-mono text-fg">{gateway.data.id}</span>
                </span>
                <span>
                  eTag <span className="font-mono text-fg">{gateway.data.eTag}</span>
                </span>
              </div>
            ) : null}
          </ServiceRow>

          <ServiceRow
            icon={<Database className="size-4" />}
            name="数据服务"
            endpoint={`${dataBase}/metrics · 默认端口 32201`}
            loading={dataService.isLoading}
            online={dataService.isSuccess}
          >
            {dataService.data ? (
              <p className="mt-2 text-[11px] text-muted">
                已加载 <span className="font-medium text-fg">{dataService.data.length}</span> 个统计指标
              </p>
            ) : null}
          </ServiceRow>
        </CardContent>
      </Card>

      <Card className="mt-4">
        <CardHeader>
          <div>
            <CardTitle>生产部署</CardTitle>
            <CardDescription>Go 服务不托管静态资源,也不设置 CORS</CardDescription>
          </div>
        </CardHeader>
        <CardContent className="space-y-3 text-xs leading-relaxed text-muted">
          <p>
            将 <code className="rounded bg-surface-2 px-1 py-0.5 font-mono text-[11px]">pnpm build</code>{' '}
            生成的 <code className="rounded bg-surface-2 px-1 py-0.5 font-mono text-[11px]">dist/</code>{' '}
            交给 nginx 托管,并把两套 API 反向代理到对应端口:
          </p>
          <pre className="overflow-x-auto rounded-xl border border-border bg-surface-2 p-4 font-mono text-[11px] leading-relaxed text-fg">
{`location / {
  try_files $uri /index.html;
}

location /api/v1/devices    { proxy_pass http://127.0.0.1:32200; }
location /api/v1/gateway    { proxy_pass http://127.0.0.1:32200; }
location /api/v1/metrics    { proxy_pass http://127.0.0.1:32201; }
location /api/v1/metricsData { proxy_pass http://127.0.0.1:32201; }
location /api/v1/kpi        { proxy_pass http://127.0.0.1:32201; }
location /api/v1/kpis       { proxy_pass http://127.0.0.1:32201; }
location /api/v1/tags       { proxy_pass http://127.0.0.1:32201; }
location /api/v1/fmcs       { proxy_pass http://127.0.0.1:32201; }
location /api/v1/escada     { proxy_pass http://127.0.0.1:32201; }`}
          </pre>
          <p className="text-faint">
            若前端与 API 不同源,可在 nginx 中统一添加 CORS 响应头,或通过{' '}
            <code className="font-mono">VITE_GATEWAY_BASE</code> /{' '}
            <code className="font-mono">VITE_DATA_BASE</code> 指向可访问的完整地址。
          </p>
        </CardContent>
      </Card>

      <Card className="mt-4">
        <CardHeader>
          <CardTitle>关于</CardTitle>
        </CardHeader>
        <CardContent className="space-y-1.5 text-xs text-muted">
          <p>
            harnsGateway 管理后台 <span className="font-mono text-fg">v0.1.0</span>
          </p>
          <p>React 19 · TypeScript · Vite · Tailwind CSS v4 · TanStack Query · Recharts</p>
          <p className="text-faint">
            界面仅调用 /api/v1 下已存在的接口。后端契约变更时,请先同步{' '}
            <code className="font-mono">src/api/types.ts</code>。
          </p>
        </CardContent>
      </Card>
    </div>
  )
}

function ThemeOption({
  label,
  active,
  onClick,
  icon,
  swatches,
}: {
  label: string
  active: boolean
  onClick: () => void
  icon: ReactNode
  swatches: string[]
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      aria-pressed={active}
      className={cn(
        'rounded-xl border p-3 text-left transition-colors',
        active
          ? 'border-primary bg-primary-soft/50 ring-1 ring-primary/30'
          : 'border-border hover:border-border-strong hover:bg-surface-2',
      )}
    >
      <span className="flex items-center justify-between">
        <span className="flex items-center gap-2 text-xs font-medium">
          {icon}
          {label}
        </span>
        {active ? <Check className="size-3.5 text-primary" /> : null}
      </span>
      <span className="mt-3 flex gap-1.5">
        {swatches.map((swatch) => (
          <span
            key={swatch}
            className={cn('h-7 flex-1 rounded-md border border-black/5 dark:border-white/10', swatch)}
          />
        ))}
      </span>
    </button>
  )
}

function EnvRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 border-b border-border pb-2 last:border-0 last:pb-0">
      <span className="text-xs text-muted">{label}</span>
      <span className="flex min-w-0 items-center gap-1">
        <code className="truncate font-mono text-[11px] text-fg">{value}</code>
        <CopyButton value={value} />
      </span>
    </div>
  )
}

function ServiceRow({
  icon,
  name,
  endpoint,
  loading,
  online,
  children,
}: {
  icon: ReactNode
  name: string
  endpoint: string
  loading: boolean
  online: boolean
  children?: ReactNode
}) {
  return (
    <div className="rounded-xl border border-border px-4 py-3">
      <div className="flex flex-wrap items-center gap-3">
        <span className="flex size-8 items-center justify-center rounded-lg bg-surface-2 text-muted">
          {icon}
        </span>
        <div className="min-w-0 flex-1">
          <p className="text-sm font-medium">{name}</p>
          <p className="truncate font-mono text-[11px] text-faint">{endpoint}</p>
        </div>
        {loading ? (
          <Badge tone="neutral" dot>
            检查中
          </Badge>
        ) : online ? (
          <Badge tone="success" dot>
            在线
          </Badge>
        ) : (
          <Badge tone="danger" dot>
            不可达
          </Badge>
        )}
      </div>
      {children}
    </div>
  )
}
