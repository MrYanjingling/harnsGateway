import { createBrowserRouter } from 'react-router-dom'
import { AppShell } from '@/components/layout/AppShell'
import { DashboardPage } from '@/features/dashboard/DashboardPage'
import { DevicesPage } from '@/features/devices/DevicesPage'
import { DeviceCreatePage } from '@/features/devices/DeviceCreatePage'
import { DeviceDetailPage } from '@/features/devices/DeviceDetailPage'
import { DeviceEditPage } from '@/features/devices/DeviceEditPage'
import { MetricsPage } from '@/features/metrics/MetricsPage'
import { KpiPage } from '@/features/kpi/KpiPage'
import { TagsPage } from '@/features/tags/TagsPage'
import { TimeSeriesPage } from '@/features/timeseries/TimeSeriesPage'
import { SettingsPage } from '@/features/settings/SettingsPage'
import { NotFoundPage } from '@/features/NotFoundPage'

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppShell />,
    children: [
      { index: true, element: <DashboardPage /> },
      { path: 'devices', element: <DevicesPage /> },
      { path: 'devices/new', element: <DeviceCreatePage /> },
      { path: 'devices/:id', element: <DeviceDetailPage /> },
      { path: 'devices/:id/edit', element: <DeviceEditPage /> },
      { path: 'data/metrics', element: <MetricsPage /> },
      { path: 'data/kpis', element: <KpiPage /> },
      { path: 'data/tags', element: <TagsPage /> },
      { path: 'data/timeseries', element: <TimeSeriesPage /> },
      { path: 'settings', element: <SettingsPage /> },
      { path: '*', element: <NotFoundPage /> },
    ],
  },
])
