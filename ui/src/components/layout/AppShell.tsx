import { useEffect, useState } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import { Menu, Moon, Sun } from 'lucide-react'
import { cn } from '@/lib/cn'
import { useTheme } from '@/providers/ThemeProvider'
import { useGatewayMeta } from '@/api/hooks'
import { Button } from '@/components/ui/Button'
import { Sidebar } from './Sidebar'

const TITLES: Record<string, string> = {
  '/': '概览',
  '/devices': '设备管理',
  '/devices/new': '新建设备',
  '/data/metrics': '统计指标',
  '/data/kpis': 'KPI',
  '/data/tags': '测点',
  '/data/timeseries': '时序查询',
  '/settings': '设置',
}

function usePageTitle() {
  const { pathname } = useLocation()
  if (TITLES[pathname]) return TITLES[pathname]
  if (pathname.startsWith('/devices/') && pathname.endsWith('/edit')) return '编辑设备'
  if (pathname.startsWith('/devices/')) return '设备详情'
  return 'harnsGateway'
}

export function AppShell() {
  const [mobileOpen, setMobileOpen] = useState(false)
  const { theme, toggle } = useTheme()
  const { pathname } = useLocation()
  const title = usePageTitle()
  const meta = useGatewayMeta()
  const isMock = import.meta.env.VITE_USE_MOCK === 'true'

  useEffect(() => {
    setMobileOpen(false)
  }, [pathname])

  useEffect(() => {
    document.title = `${title} · harnsGateway`
  }, [title])

  const online = meta.isSuccess && !meta.isError

  return (
    <div className="flex min-h-dvh bg-bg">
      {/* 桌面侧边栏 */}
      <aside className="fixed inset-y-0 left-0 z-30 hidden w-60 border-r border-border bg-surface lg:block">
        <Sidebar />
      </aside>

      {/* 移动端抽屉 */}
      <div
        className={cn(
          'fixed inset-0 z-40 lg:hidden',
          mobileOpen ? 'pointer-events-auto' : 'pointer-events-none',
        )}
      >
        <div
          className={cn(
            'absolute inset-0 bg-black/40 backdrop-blur-[1px] transition-opacity',
            mobileOpen ? 'opacity-100' : 'opacity-0',
          )}
          onClick={() => setMobileOpen(false)}
        />
        <aside
          className={cn(
            'absolute inset-y-0 left-0 w-64 border-r border-border bg-surface shadow-2xl transition-transform duration-200',
            mobileOpen ? 'translate-x-0' : '-translate-x-full',
          )}
        >
          <Sidebar onNavigate={() => setMobileOpen(false)} />
        </aside>
      </div>

      <div className="flex min-w-0 flex-1 flex-col lg:pl-60">
        <header className="sticky top-0 z-20 flex h-14 items-center gap-3 border-b border-border bg-bg/80 px-4 backdrop-blur-md sm:px-6">
          <Button
            variant="ghost"
            size="icon"
            className="lg:hidden"
            onClick={() => setMobileOpen(true)}
            aria-label="打开菜单"
          >
            <Menu className="size-4.5" />
          </Button>

          <div className="flex min-w-0 items-center gap-2 text-sm">
            <span className="hidden text-muted sm:inline">harnsGateway</span>
            <span className="hidden text-faint sm:inline">/</span>
            <span className="truncate font-medium">{title}</span>
          </div>

          <div className="ml-auto flex items-center gap-2">
            {isMock ? (
              <span className="hidden rounded-full border border-amber-600/25 bg-amber-500/10 px-2 py-0.5 text-[11px] font-medium text-amber-700 sm:inline dark:text-amber-300">
                Mock 数据
              </span>
            ) : null}
            <span
              className="hidden items-center gap-1.5 rounded-full border border-border bg-surface px-2.5 py-1 text-[11px] text-muted sm:inline-flex"
              title={online ? '网关服务已连接' : '网关服务不可用'}
            >
              <span
                className={cn(
                  'size-1.5 rounded-full',
                  meta.isError ? 'bg-rose-500' : online ? 'bg-emerald-500' : 'bg-amber-500',
                )}
              />
              <span className="max-w-36 truncate">
                {meta.isError ? '网关离线' : (meta.data?.name ?? '连接中…')}
              </span>
            </span>
            <Button variant="ghost" size="icon" onClick={toggle} aria-label="切换主题">
              {theme === 'dark' ? <Sun className="size-4" /> : <Moon className="size-4" />}
            </Button>
          </div>
        </header>

        <main className="mx-auto w-full max-w-[1400px] flex-1 px-4 py-6 sm:px-6 lg:px-8">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
