import { NavLink } from 'react-router-dom'
import {
  Activity,
  ChartColumn,
  ChartLine,
  Cpu,
  LayoutDashboard,
  Settings,
  Target,
  Waypoints,
} from 'lucide-react'
import { cn } from '@/lib/cn'
import { useGatewayMeta } from '@/api/hooks'

interface NavItem {
  to: string
  label: string
  icon: typeof Cpu
  end?: boolean
}

const GROUPS: Array<{ label?: string; items: NavItem[] }> = [
  {
    items: [
      { to: '/', label: '概览', icon: LayoutDashboard, end: true },
      { to: '/devices', label: '设备管理', icon: Cpu },
    ],
  },
  {
    label: '数据服务',
    items: [
      { to: '/data/metrics', label: '统计指标', icon: ChartColumn },
      { to: '/data/kpis', label: 'KPI', icon: Target },
      { to: '/data/tags', label: '测点', icon: Waypoints },
      { to: '/data/timeseries', label: '时序查询', icon: ChartLine },
    ],
  },
  {
    label: '系统',
    items: [{ to: '/settings', label: '设置', icon: Settings }],
  },
]

export function Sidebar({ onNavigate }: { onNavigate?: () => void }) {
  const meta = useGatewayMeta()

  return (
    <div className="flex h-full flex-col">
      <div className="flex h-14 items-center gap-2.5 px-5">
        <div className="flex size-8 items-center justify-center rounded-xl bg-gradient-to-br from-emerald-400 to-emerald-700 text-white shadow-sm">
          <svg viewBox="0 0 24 24" className="size-4.5" fill="none" aria-hidden="true">
            <path
              d="M7 15V8m0 3.5h5M12 15V8m0 3.5h5V15"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
          </svg>
        </div>
        <div className="min-w-0">
          <p className="text-sm font-semibold leading-tight tracking-tight">harnsGateway</p>
          <p className="text-[10px] leading-tight text-faint">边缘网关管理后台</p>
        </div>
      </div>

      <nav className="flex-1 space-y-5 overflow-y-auto px-3 py-3">
        {GROUPS.map((group, index) => (
          <div key={group.label ?? index}>
            {group.label ? (
              <p className="mb-1.5 px-2 text-[10px] font-medium tracking-wider text-faint uppercase">
                {group.label}
              </p>
            ) : null}
            <div className="space-y-0.5">
              {group.items.map((item) => (
                <NavLink
                  key={item.to}
                  to={item.to}
                  end={item.end}
                  onClick={onNavigate}
                  className={({ isActive }) =>
                    cn(
                      'group flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-sm transition-colors',
                      isActive
                        ? 'bg-primary-soft font-medium text-primary'
                        : 'text-muted hover:bg-surface-2 hover:text-fg',
                    )
                  }
                >
                  <item.icon className="size-4 shrink-0" />
                  <span className="truncate">{item.label}</span>
                </NavLink>
              ))}
            </div>
          </div>
        ))}
      </nav>

      <div className="border-t border-border p-3">
        <div className="rounded-xl border border-border bg-surface-2 px-3 py-2.5">
          <div className="flex items-center gap-2 text-xs font-medium">
            <Activity className="size-3.5 text-primary" />
            网关状态
          </div>
          <div className="mt-1.5 flex items-center gap-1.5 text-[11px] text-muted">
            <span
              className={cn(
                'size-1.5 rounded-full',
                meta.isError
                  ? 'bg-rose-500'
                  : meta.isSuccess
                    ? 'bg-emerald-500'
                    : 'animate-pulse bg-amber-500',
              )}
            />
            <span className="truncate">
              {meta.isError ? '服务不可用' : meta.isSuccess ? '服务运行中' : '正在连接…'}
            </span>
          </div>
          {meta.data ? (
            <p className="mt-1 truncate font-mono text-[10px] text-faint" title={meta.data.id}>
              ID {meta.data.id}
            </p>
          ) : null}
        </div>
      </div>
    </div>
  )
}
