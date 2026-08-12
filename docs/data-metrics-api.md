# 数据服务 API 接口文档

## 通用说明

- 基础路径: `/api/v1`
- Content-Type: `application/json`
- 统一响应结构:

```json
{
  "data": "<具体数据>",
  "total": 0
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `data` | `any` | 响应数据，可为对象、数组或字符串 |
| `total` | `number` | 列表接口返回总条数，非列表接口省略 |

---

## 1. POST /api/v1/metricsData

### 接口描述

通过 `metricsCodes` 批量查询统计指标在指定时间范围内的汇总值（sum 聚合），支持跨月 ES 索引查询。

### 请求参数

| 参数 | 类型 | 必填 | 校验 | 说明 |
|------|------|------|------|------|
| `metricsCodes` | `string[]` | 是 | 非空数组 | 统计指标编码列表 |
| `startTime` | `string` | 是 | `yyyy-MM-dd HH:mm:ss` | 查询起始时间 |
| `endTime` | `string` | 是 | `yyyy-MM-dd HH:mm:ss` | 查询结束时间 |
| `granularity` | `string` | 是 | `h` | 时间粒度，仅支持 h |

### 请求示例

```json
{
  "metricsCodes": ["energy_total", "power_factor"],
  "startTime": "2026-07-01 00:00:00",
  "endTime": "2026-07-02 00:00:00",
  "granularity": "h"
}
```

### 响应格式

`data` 为数组，保持入参顺序：

| 字段 | 类型 | 说明 |
|------|------|------|
| `metricsCode` | `string` | 统计指标编码 |
| `startTime` | `string` | 查询起始时间 |
| `endTime` | `string` | 查询结束时间 |
| `value` | `number` | s_data 的 sum 聚合值 |

### 成功响应

```json
{
  "data": [
    {
      "metricsCode": "energy_total",
      "startTime": "2026-07-01 00:00:00",
      "endTime": "2026-07-23 23:59:59",
      "value": 15420.50
    },
    {
      "metricsCode": "power_factor",
      "startTime": "2026-07-01 00:00:00",
      "endTime": "2026-07-23 23:59:59",
      "value": 0.92
    }
  ],
  "total": 2
}
```

指标不存在时 `value` 为 `0`：

```json
{
  "data": [
    {
      "metricsCode": "unknown_metric",
      "startTime": "2026-07-01 00:00:00",
      "endTime": "2026-07-23 23:59:59",
      "value": 0
    }
  ],
  "total": 1
}
```

### 错误响应

**参数错误 (400):**

```json
{ "data": "invalid startTime, expected format: 2006-01-02 15:04:05" }
```

```json
{ "data": "invalid request: Key: 'MetricsQueryRequest.metricsCodes' Error:Field validation for 'metricsCodes' failed on the 'required' tag" }
```

**服务异常 (500):**

```json
{ "data": "failed to query metrics" }
```

### 数据源说明

| 数据源 | 表/索引 | 用途 |
|--------|---------|------|
| MySQL | `ems_statistics_metrics` | `WHERE metrics_code = ?` 获取指标 ID |
| Elasticsearch | `metrics_data_YYYYMM` | 根据 `metrics_id` + `s_time` 范围做 sum 聚合 |

ES 索引按月分片，查询跨月时自动拼接多个索引名。

---

## 2. GET /api/v1/metrics

### 接口描述

查询所有统计指标列表，返回 `ems_statistics_metrics` 表全量数据。

### 请求参数

无。

### 响应格式

`data` 为 `MetricsInfo` 数组，`total` 为总条数。

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | `string` | 主键 ID |
| `tenantId` | `string/null` | 租户号 |
| `revision` | `number/null` | 乐观锁版本 |
| `createdBy` | `string/null` | 创建人 |
| `createdTime` | `string/null` | 创建时间 |
| `updatedBy` | `string/null` | 更新人 |
| `updatedTime` | `string/null` | 更新时间 |
| `dataPoint` | `string/null` | 测点名称 |
| `metricsCode` | `string/null` | 统计指标编码 |
| `metricsName` | `string/null` | 统计指标名称 |
| `measureType` | `string/null` | 计量单位 ID |
| `enable` | `boolean/null` | 是否启用 |
| `metricsType` | `string/null` | M=手动表, A=自动表 |
| `calculateType` | `string/null` | 聚合类型: sum/avg/max/min |
| `note` | `string/null` | 备注 |

### 成功响应

```json
{
  "data": [
    {
      "id": "abc123",
      "tenantId": "tenant-01",
      "revision": 3,
      "createdBy": "admin",
      "createdTime": "2026-01-01T10:00:00Z",
      "updatedBy": "admin",
      "updatedTime": "2026-07-01T15:30:00Z",
      "dataPoint": "meter.power",
      "metricsCode": "energy_total",
      "metricsName": "总能耗",
      "measureType": "kWh",
      "enable": true,
      "metricsType": "A",
      "calculateType": "sum",
      "note": "全厂总能耗指标"
    }
  ],
  "total": 1
}
```

### 错误响应

```json
{ "data": "failed to list metrics" }
```

---

## 3. POST /api/v1/kpi

### 接口描述

通过 KPI ID 批量查询 KPI 在指定时间范围内的汇总值（sum 聚合）。

### 请求参数

| 参数 | 类型 | 必填 | 校验 | 说明 |
|------|------|------|------|------|
| `ids` | `string[]` | 是 | 非空数组 | KPI ID 列表 |
| `startTime` | `string` | 是 | `yyyy-MM-dd HH:mm:ss` | 查询起始时间 |
| `endTime` | `string` | 是 | `yyyy-MM-dd HH:mm:ss` | 查询结束时间 |
| `granularity` | `string` | 是 | `h` | 时间粒度，仅支持 h |

### 请求示例

```json
{
  "ids": ["kpi_001", "kpi_002"],
  "startTime": "2026-07-01 00:00:00",
  "endTime": "2026-07-31 23:59:59",
  "granularity": "h"
}
```

### 响应格式

`data` 为数组，保持入参顺序：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | `string` | KPI ID |
| `startTime` | `string` | 查询起始时间 |
| `endTime` | `string` | 查询结束时间 |
| `value` | `number` | s_data 的 sum 聚合值 |

### 成功响应

```json
{
  "data": [
    {
      "id": "kpi_001",
      "startTime": "2026-07-01 00:00:00",
      "endTime": "2026-07-31 23:59:59",
      "value": 3200.80
    }
  ],
  "total": 1
}
```

### 错误响应

**参数错误 (400):**

```json
{ "data": "invalid request: Key: 'KpiQueryRequest.Ids' Error:Field validation for 'Ids' failed on the 'required' tag" }
```

**服务异常 (500):**

```json
{ "data": "failed to query kpi" }
```

### 数据源说明

| 数据源 | 表/索引 | 用途 |
|--------|---------|------|
| MySQL | `ems_statistics_metrics` | 根据 `metrics_code = ?` 获取 KPI 元数据 |
| Elasticsearch | `kpi_data_YYYYMM` | 根据 `kpi_id` + `s_time` 范围查询 `s_data` sum 聚合 |

---

## 4. GET /api/v1/kpis

### 接口描述

查询所有 KPI 列表，返回 `ems_kpi` 表全量数据。

### 请求参数

无。

### 响应格式

`data` 为 `KpiInfo` 数组：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | `string` | KPI ID |
| `name` | `string` | KPI 名称 |
| `formula` | `string/null` | 解析后的公式内容 |
| `formulaDataName` | `string/null` | 公式涉及数据对象名称 |
| `formulaOri` | `string` | 原始公式内容 |
| `timeGranularity` | `string` | 时间粒度: h/d/m |
| `targetData` | `string/null` | 目标值 |
| `singleQuota` | `boolean` | 是否单一 KPI |
| `defaultTag` | `boolean` | 是否出厂默认 |
| `canDelete` | `boolean` | 是否能删除 |
| `unit` | `string/null` | 单位 |
| `targetDataType` | `string` | 目标值特性 |
| `status` | `string` | 状态 |
| `tenantId` | `string/null` | 租户号 |
| `revision` | `number/null` | 乐观锁版本 |
| `createdBy` | `string/null` | 创建人 |
| `createdTime` | `string/null` | 创建时间 |
| `updatedBy` | `string/null` | 更新人 |
| `updatedTime` | `string/null` | 更新时间 |

### 成功响应

```json
{
  "data": [
    {
      "id": "kpi_001",
      "tenantId": "tenant-01",
      "name": "日总能耗",
      "formula": "tag1+tag2",
      "formulaOri": "A+B",
      "timeGranularity": "d",
      "singleQuota": true,
      "defaultTag": false,
      "canDelete": true,
      "unit": "kWh",
      "targetDataType": "number",
      "status": "active"
    }
  ],
  "total": 1
}
```

### 错误响应

```json
{ "data": "failed to list kpis" }
```

---

## 5. GET /api/v1/tags

### 接口描述

查询所有测点列表，返回 `ems_tag` 表全量数据。

### 请求参数

无。

### 响应格式

`data` 为 `TagInfo` 数组：

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | `string` | 主键 ID |
| `tenantId` | `string/null` | 租户号 |
| `revision` | `number/null` | 乐观锁版本 |
| `createdBy` | `string/null` | 创建人 |
| `createdTime` | `string/null` | 创建时间 |
| `updatedBy` | `string/null` | 更新人 |
| `updatedTime` | `string/null` | 更新时间 |
| `tagId` | `string/null` | 测点 ID |
| `tagName` | `string/null` | 测点名称 |
| `tagDesc` | `string/null` | 测点描述 |
| `lastSynTime` | `string/null` | 最近同步时间 |
| `tagType` | `string/null` | 测点类型 |
| `tagAttribute` | `string/null` | 测点属性 |
| `enable` | `number/null` | 启用状态: 1=启用, 0=禁用 |
| `scadaId` | `string` | 所属 SCADA |
| `operator` | `string/null` | 运算符 |
| `operateValue` | `string/null` | 运算值 |

### 成功响应

```json
{
  "data": [
    {
      "id": "tag_001",
      "tagId": "SCADA.temperature",
      "tagName": "温度测点",
      "tagDesc": "1号机组温度",
      "enable": 1,
      "scadaId": "scada-01",
      "tagType": "analog"
    }
  ],
  "total": 1
}
```

### 错误响应

```json
{ "data": "failed to list tags" }
```

---

## 6. POST /api/v1/tagTimeSeries

### 接口描述

查询测点时序数据，代理转发至第三方 SCADA 接口 `POST /di/scada/v1/portal/timeSeries`。最多支持 10 个测点同时查询。

### 前置配置

启动参数 `--scada-url` 配置第三方 SCADA 服务地址，例如：
```
--scada-url=https://scada.example.com
```

### 请求参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `select` | `string` | 是 | 测点 ID，单个如 `tag1`，多个逗号分隔如 `tag1,tag2` |
| `start` | `string` | 否 | 起始时间，格式 `yyyy-MM-dd HH:mm:ss.SSS` |
| `end` | `string` | 否 | 结束时间，格式 `yyyy-MM-dd HH:mm:ss.SSS` |
| `latest` | `boolean` | 否 | 是否查询最新值，默认 `false` |
| `limit` | `number` | 否 | 返回数据条数 |
| `mode` | `string` | 否 | 值类型：`raw`(默认)/`mean`(平均值)/`usage`(差值) |

### 请求示例

**查询时序数据:**

```json
{
  "select": "SCADA.temp1,SCADA.power1",
  "start": "2026-08-01 00:00:00.000",
  "end": "2026-08-09 23:59:59.999",
  "latest": false,
  "limit": 100,
  "mode": "raw"
}
```

**查询最新值:**

```json
{
  "select": "SCADA.temp1",
  "latest": true
}
```

### 响应格式（code=200 成功时）

`data` 为 `DataPointTimeSeries` 数组：

| 字段 | 类型 | 说明 |
|------|------|------|
| `tagId` | `string` | 测点 ID |
| `tagName` | `string` | 测点名称 |
| `tagDesc` | `string` | 测点描述 |
| `data` | `TimeSeriesTO[]` | 时序数据点数组 |

`TimeSeriesTO`:

| 字段 | 类型 | 说明 |
|------|------|------|
| `time` | `string` | 数据时间戳 |
| `value` | `string` | 数据值 |

### 成功响应

```json
{
  "data": [
    {
      "tagId": "SCADA.temp1",
      "tagName": "温度测点1",
      "tagDesc": "1号机组温度",
      "data": [
        { "time": "2026-08-01 00:00:00", "value": "25.5" },
        { "time": "2026-08-01 01:00:00", "value": "26.1" }
      ]
    }
  ]
}
```

### 错误响应


**参数错误 (400):**

```json
{ "data": "invalid request: Key: 'TagTimeSeriesRequest.Select' Error:Field validation for 'Select' failed on the 'required' tag" }
```

**第三方接口返回非 200 (500):**

```json
{ "data": "<第三方接口的 message>" }
```

**服务异常 (500):**

```json
{ "data": "failed to query tag time series" }
```

---

## 附录

### 接口总览

| 方法 | 路径 | 说明 | 数据源 |
|------|------|------|--------|
| POST | `/api/v1/metricsData` | 批量查询指标聚合值 | MySQL + ES |
| GET | `/api/v1/metrics` | 查询所有统计指标 | MySQL |
| POST | `/api/v1/kpi` | 批量查询 KPI 聚合值 | MySQL + ES |
| GET | `/api/v1/kpis` | 查询所有 KPI | MySQL |
| GET | `/api/v1/tags` | 查询所有测点 | MySQL |
| POST | `/api/v1/tagTimeSeries` | 查询测点时序数据 | 第三方 SCADA |



