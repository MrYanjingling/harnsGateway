# harnsGateway 管理后台

harnsGateway 的 Web 管理界面,覆盖网关服务与数据服务的全部现有 API:

- **概览**:网关身份、CPU / 内存 / 磁盘、设备状态分布、最近设备
- **设备管理**:列表筛选、创建、编辑、启动 / 停止 / 重启、删除、变量反控
- **数据服务**:统计指标、KPI、测点、时序查询(折线图)
- **设置**:主题、运行环境、服务连通性、生产部署说明

技术栈:React 19 · TypeScript · Vite 7 · Tailwind CSS v4 · TanStack Query · React Router 7 · React Hook Form + Zod · Recharts。

## 快速开始

要求 Node.js ≥ 20(推荐 22 LTS)与 pnpm。

```bash
cd ui
pnpm install

# 方式一:纯前端演示(mock,无需 Go 后端)
cp .env.example .env.local     # .env.example 默认 VITE_USE_MOCK=true
pnpm dev                       # http://localhost:5173

# 方式二:连接真实后端
VITE_USE_MOCK=false pnpm dev
```

不创建 `.env.local` 时,也可以直接通过命令行覆盖:

```bash
VITE_USE_MOCK=true pnpm dev
```

## 环境变量

| 变量 | 默认值 | 说明 |
|---|---|---|
| `VITE_USE_MOCK` | `false` | `true` 启用内置 mock,且不注册任何 proxy |
| `VITE_GATEWAY_TARGET` | `http://127.0.0.1:32200` | 开发期网关服务代理目标 |
| `VITE_DATA_TARGET` | `http://127.0.0.1:32201` | 开发期数据服务代理目标 |
| `VITE_GATEWAY_BASE` | `/api/v1` | 生产环境网关服务基址(可选) |
| `VITE_DATA_BASE` | `/api/v1` | 生产环境数据服务基址(可选) |

## 开发期代理

两个 Go 服务共用 `/api/v1` 前缀且**没有 CORS**,因此开发期由 Vite 按路径分流(见 `vite.config.ts`):

| 路径前缀 | 目标 |
|---|---|
| `/api/v1/devices`、`/api/v1/gatewayMeta`、`/gatewayCpu`、`/gatewayMem`、`/gatewayDisk` | 网关 32200 |
| `/api/v1/metrics`、`/metricsData`、`/kpi`、`/kpis`、`/tags`、`/fmcs`、`/escada` | 数据 32201 |

网关相关键必须写在前面,避免被通用规则抢先匹配。

```bash
# 根目录启动后端
go run ./cmd/gateway    # :32200
go run ./cmd/data       # :32201
```

## 构建与部署

```bash
pnpm build     # 先 tsc --noEmit,再 vite build -> dist/
pnpm preview   # 本地预览构建产物
```

Go 服务不托管静态资源。生产环境用 nginx 托管 `dist/`,并把 API 反代到对应端口:

```nginx
location / {
  try_files $uri /index.html;
}

location /api/v1/devices     { proxy_pass http://127.0.0.1:32200; }
location /api/v1/gateway     { proxy_pass http://127.0.0.1:32200; }
location /api/v1/metrics     { proxy_pass http://127.0.0.1:32201; }
location /api/v1/metricsData { proxy_pass http://127.0.0.1:32201; }
location /api/v1/kpi         { proxy_pass http://127.0.0.1:32201; }
location /api/v1/kpis        { proxy_pass http://127.0.0.1:32201; }
location /api/v1/tags        { proxy_pass http://127.0.0.1:32201; }
location /api/v1/fmcs        { proxy_pass http://127.0.0.1:32201; }
location /api/v1/escada      { proxy_pass http://127.0.0.1:32201; }
```

若前后端不同源且无法使用反向代理,需要自行在网关层添加 CORS,或把 `VITE_GATEWAY_BASE` / `VITE_DATA_BASE` 指向可直接访问的完整地址。

## 目录说明

```
src/
├── api/        # 后端契约类型、端点封装、query keys、TanStack Query hooks
├── components/ # layout / ui 原语 / shared 业务组件
├── features/   # 按页面组织
├── lib/        # http、format、constants、cn
├── providers/  # 主题
mock/           # 开发期内存态 API(Vite 中间件)
```

更细的约定见 [`AGENTS.md`](./AGENTS.md),仓库级后端契约见 [`../AGENTS.md`](../AGENTS.md)。
