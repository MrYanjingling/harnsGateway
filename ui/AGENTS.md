# ui/AGENTS.md — 前端约定

React 19 + TypeScript + Vite 7 + Tailwind CSS v4 + TanStack Query + React Router 7。
根目录 `../AGENTS.md` 记录了后端契约;本文件只讲前端。

## 目录

```
src/
├── main.tsx / router.tsx / index.css
├── api/          # types.ts(后端契约)、endpoints、query-keys、hooks
├── lib/          # http、format、constants、cn
├── providers/    # ThemeProvider
├── components/
│   ├── layout/   # AppShell + Sidebar
│   ├── ui/       # 无业务语义的原语:Button/Card/Field/Dialog/Table/Badge…
│   ├── charts/   # GaugeRing 等
│   └── shared/   # PageHeader/StatCard/StatusBadge/TimeRangeFields/ErrorState
└── features/     # 按页面域组织:dashboard/devices/metrics/kpi/tags/timeseries/settings
mock/
├── index.ts      # Vite 中间件,内存态 API
└── fixtures.ts   # 种子数据
```

## 数据请求

- 所有 HTTP 走 `src/lib/http.ts` 的 `request(service, path, options)`。
  - `service` 为 `'gateway'` 或 `'data'`,基址默认都是 `/api/v1`,由 Vite proxy 分流。
  - 错误统一抛 `ApiError`,展示时用 `errorMessage(error)`。
  - 设备写操作必须传 `ifMatch`(来自 `DeviceSummary.eTag` / `DeviceDetail.eTag`)。
- Query key 集中定义在 `src/api/query-keys.ts`,不要在组件里手写 key 数组。
- 每个接口一个 hook,放在 `src/api/hooks.ts`:
  - 设备列表/网关资源 10s 轮询,设备详情 5s 轮询。
  - 写设备成功后 `invalidateQueries({ queryKey: queryKeys.devices.all })`。
  - 412/404 冲突时先失效缓存,让用户看到最新版本,不要静默覆盖。
- 数据服务响应统一 `{ data, total }`,解包逻辑在 `src/api/data.ts`。

## 组件

- `components/ui` 只放通用原语,不含业务概念;有业务语义的放 `components/shared` 或对应 `features`。
- 复用现有视觉语义:卡片 `rounded-2xl border border-border bg-surface`,主色 `text-primary`,弱文本 `text-muted/faint`,数字用 `tabular`。
- 颜色只引用 `index.css` 里的语义令牌(`bg-surface`、`border-border`、`text-danger` 等),不写死 zinc/emerald 调色板(状态徽标除外)。
- 弹窗用原生 `<dialog>`(`components/ui/Dialog.tsx`),自带焦点陷阱。
- 表格在窄屏必须可横向滚动(`components/ui/Table.tsx` 已内置 `overflow-x-auto`)。
- 空态用 `EmptyState`,错误用 `ErrorState`,加载用 `Skeleton`/`TableSkeleton`/`PageLoading`。不要只渲染 `null`。

## 表单

- 设备表单:react-hook-form + zod(`features/devices/schema.ts`)。
- 表单是**扁平模型**,协议差异在 `superRefine` 里按 `deviceType` 校验,提交时由 `toPayload` 映射成 v1 形状。
- 默认值必须避开 Go `binding:"required"` 的零值:采集周期 1 秒、slave 1、functionCode 3、namespace 2。
- 编辑时 `deviceType`/`deviceModel` 不可切换;从 `?exploded=true` 载入后用 `toFormValues` 转回表单。
- 新增变量字段时,同时更新 `schema.ts` 的三处映射:`defaultFormValues`、`toFormValues`、`toPayload`。

## Mock

- `VITE_USE_MOCK=true` 时启用 `mock/index.ts`,同时不注册 Vite proxy。
- mock 是**有状态**的:创建/更新/删除/启停/反控都会改变内存数据并递增 eTag;刷新页面重置。
- 新接口必须同时实现 mock 路由;未实现的 `/api/v1/*` 会显式返回 404,避免误判为真实后端。
- mock 响应形状必须与 Go 一致,尤其 `eTag`、`cpus/mem/disk`、`Total/Used/UsedPercent`、`{data,total}`。

## 样式与主题

- Tailwind v4 CSS-first:`index.css` 中 `@theme inline` 把语义变量映射为工具类。
- 深色模式通过 `<html class="dark">`;`ThemeProvider` 持久化到 `localStorage['harns-theme']`,首屏脚本在 `index.html` 内联,避免闪烁。
- 图表配色使用固定色板(`features/timeseries` 中的 `SERIES_COLORS`),保证明暗主题下都可辨识。

## 提交前检查

```bash
pnpm typecheck
pnpm build
VITE_USE_MOCK=true pnpm dev
```

手动走查:概览 → 设备列表(启停/删除)→ 新建设备(三种协议)→ 详情(反控)→ 指标/KPI/测点/时序 → 设置;并切换一次深色主题。
