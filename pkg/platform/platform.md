# harns-platform 模块开发规范（Vibe Coding）

> 本文件是 `platform`（物联网基础平台）模块的唯一事实来源。开发/生成代码前必须先读本文件，
> 所有实现必须满足「约束」章节，并按「功能清单」的序号逐条交付。
> 与本文档冲突的直觉性实现一律视为错误。

---

## 1. 背景与目标

- 1.1 项目：`harns-platform`，一个多服务的物联网平台。已有 `gateway` 模块，现新增 `platform` 模块。
- 1.2 platform 定位：物联网基础平台服务，承担设备接入之外的基础能力（账号、租户、组织、资产管理、数据存储等——见第 10 节）。
- 1.3 目标：在既有框架与约定下增量开发，不破坏 gateway，不改动通用层语义。

## 2. 技术栈与依赖

- 2.1 语言：Go 1.24（`go.mod` 的 `go` 指令）。
- 2.2 Web：`github.com/gin-gonic/gin`。
- 2.3 配置：`sigs.k8s.io/yaml` + cobra/pflag，配置文件 `/etc/platform/config.yaml`（可用 `-c` 覆盖）。
- 2.4 日志：`harns-platform/pkg/logging`（基于 `log/slog` 的门面，非 klog）。
- 2.5 存储：MySQL（`database/sql` + `go-sql-driver/mysql`）、Redis（`redis/go-redis/v9`）、MinIO（`minio/minio-go/v7`）。
- 2.6 新增第三方依赖需说明理由，且必须兼容 Go 1.24。

## 3. 目录与分层约束

- 3.1 通用层（禁止被模块反向依赖）：`pkg/apis`、`pkg/generic`、`pkg/host`、`pkg/i18n`、`pkg/logging`、`pkg/tsdb`、`pkg/utils`、`pkg/version`。
- 3.2 模块层：platform 的所有业务代码放 `pkg/platform/` 下；服务入口放 `cmd/platform/`。
- 3.3 禁止把 platform 业务代码写入 `pkg/gateway` 或通用层；禁止跨模块直接引用 gateway 的内部包。
- 3.4 目录建议（按需新增子包）：
  ```
  pkg/platform/
    web/        # HTTP server 与路由装配（已有）
    <domain>/   # 业务领域包，按聚合拆（如 account/、asset/）
  cmd/platform/
    app/ config/ options/  # 启动装配（已有）
  ```

## 4. 通用编码约束

- 4.1 遵循 Go 官方风格；提交前 `gofmt`。
- 4.2 错误必须显式处理，禁止 `_ = err` 吞掉业务错误；禁止 panic 处理可预期错误。
- 4.3 不在业务层直接 `c.JSON`；统一走第 5 节的响应封装。
- 4.4 面向接口编程，依赖注入（构造函数传入 `*config.Config` 或具体 client），便于测试。
- 4.5 所有面向用户的错误信息必须走第 6 节的 i18n，不得硬编码中英文。
- 4.6 新增导出符号需有中文/英文注释说明用途。

## 5. API 响应规范

- 5.1 所有接口统一返回 `pkg/apis/response.Result`：
  - 成功：`{"code":200,"message":"success","data":<payload>}`
  - 失败：`{"code":<业务码>,"message":"<本地化消息>","errors":[{"code":..,"message":..}, ...]}`
- 5.2 成功用 `response.WriteOK(c, data)`；失败用 `response.WriteError(c, status, err...)`。
- 5.3 HTTP 状态码保留语义（400/401/403/404/409/412/500），业务码放 body 的 `code`。
- 5.4 前端约定：`code === 200` 用 `data`，否则用 `errors`（数组，单错长度 1）。
- 5.5 组件/JSON 字段命名用 lowerCamelCase；示例接口文档放 `apis/platform/`。

## 6. 错误码与国际化

- 6.1 错误码集中在 `pkg/apis/response/constants.go`（`ErrCode`，从 10001 起），**只能末尾追加**。
- 6.2 每个错误码在 `pkg/apis/response/errors.go` 的 `messageIDs` 里登记语义 key（如 `error.account_not_found`）。
- 6.3 文案放 `pkg/i18n/locales/<lang>.yaml`，至少提供 `en-US` 与 `zh-CN`，占位符用 `%s`/`%d`。
- 6.4 构造错误用 `response` 包提供的构造函数（如 `response.ErrXxx(args...)`），不要在业务层拼字符串。
- 6.5 语言来源：`?lang=`（优先）或 `Accept-Language`，由中间件写入，`i18n.Lang(c)` 读取。

## 7. 配置规范

- 7.1 配置项定义在 `cmd/platform/options/Options`，带 `json` tag；默认值在 `NewDefaultOptions`。
- 7.2 同步更新 `etc/platform/config.yaml` 模板（含注释）。
- 7.3 敏感信息（密码/密钥）默认留空，不入库、不打日志。
- 7.4 配置优先级：CLI flag > 配置文件 > 内置默认值（框架已实现，勿破坏）。
- 7.5 启动时校验配置（`options.Validate`），非法配置要报错并退出。

## 8. 日志规范

- 8.1 使用 `harns-platform/pkg/logging`：`logging.V(n).InfoS(msg, kv...)`、`logging.ErrorS(err, msg, kv...)`。
- 8.2 禁止 `fmt.Println`/标准库 `log`；禁止把日志写到 stdout 业务输出。
- 8.3 级别使用：V(1) 里程碑、V(2) 日常与失败（默认可见）、V(3+) 细节；错误用 `ErrorS` 并带 `err`。
- 8.4 文件输出与切割由 `logging` 配置控制（`max-size`/`max-backups`/`max-age`/`compress`），勿自行实现。

## 9. 存储规范

- 9.1 MySQL：使用 `cmd/platform/config.Config.DB`（`*sql.DB`）；查询用参数化，禁止拼接 SQL。
- 9.2 Redis：使用 `config.Config.Redis`（`redis.UniversalClient`）；key 需带业务前缀，设置 TTL。
- 9.3 MinIO：使用 `config.Config.MinIO`（`*minio.Client`）与 `MinIOBucket`；对象命名需可追溯。
- 9.4 表结构变更需提供 SQL 迁移脚本，放 `deploy/` 或 `migrations/`（按需约定）。

## 10. 功能清单（按序号编写）

> 每个功能用如下格式，序号不可跳号；未实现前保持 `状态: TODO`。

### F1. <功能名称>
- 状态：TODO
- 目标：<一句话说明价值>
- 输入：<请求/数据来源>
- 处理：<核心逻辑步骤，可列子项>
- 输出：<响应/副作用>
- 约束：<特殊限制，引用上面章节，如「见 5.1、6.2」>
- 验收标准：
  - [ ] <可验证的断言 1>
  - [ ] <可验证的断言 2>

### F2. <功能名称>
- 状态：TODO
- 目标：
- 输入：
- 处理：
- 输出：
- 约束：
- 验收标准：
  - [ ]

<!-- 继续 F3、F4 ... -->

## 11. 任务提示词模板（每次 vibe coding 用）

> 把下面模板连同本文件一起发给 AI，明确「只做某一个序号」。

```
请先阅读 pkg/platform/platform.md（模块规范）与涉及的通用包源码，然后实现 F<序号>。

范围：
- 只改动：<列出允许改的文件/目录>
- 不要改动：gateway 模块、通用层语义、其他功能

要求：
- 遵守 platform.md 第 3~9 节的约束（分层、响应、错误码/i18n、配置、日志、存储）
- 复用现有框架：response.WriteOK/WriteError、i18n、logging、config.Config
- 新增错误码时：按 6.1/6.2/6.3 同时补 constants.go、messageIDs、locales
- 新增配置项时：按 7.1/7.2 同步 options 与 etc/platform/config.yaml

验收：
- 满足 F<序号> 的「验收标准」全部勾选项
- go build ./... 通过；gofmt 干净
- 如有接口，给出请求/响应示例

请先给出实现计划与改动文件清单，再动手写代码。
```

## 12. 禁止事项

- 12.1 禁止引入 klog/component-base 等 k8s 日志链路；日志统一走 `pkg/logging`。
- 12.2 禁止在业务层返回裸对象（必须包 `Result`）。
- 12.3 禁止硬编码文案、魔法数、明文密钥。
- 12.4 禁止破坏 `code===200?data:errors` 的前端约定。
- 12.5 禁止让通用层 `import` 模块层。

## 13. 交付自检清单

- [ ] 功能对应序号已从 TODO 改为 DONE
- [ ] `go build ./...` 通过
- [ ] `gofmt -l` 无输出
- [ ] 新增错误码/文案/配置三项齐全
- [ ] 有接口时附请求/响应示例
- [ ] 未触碰禁止事项
