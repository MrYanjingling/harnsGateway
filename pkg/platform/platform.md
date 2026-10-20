# harns-platform 模块开发规范（Vibe Coding）

> 本文件是 `platform` 模块的唯一事实来源。开发/生成代码前先读本文；实现必须满足「约束」，并按「功能清单」序号逐条交付。

---

## 1. 背景与目标

- 1.1 `harns-platform` 多服务物联网平台，已有 `gateway`，现新增 `platform`。
- 1.2 platform 定位：设备接入之外的基础能力（账号、租户、组织、资产、数据存储等，见第 10 节）。
- 1.3 增量开发：不破坏 gateway，不改动通用层语义。

## 2. 技术栈与依赖

- 2.1 Go 1.24（`go.mod` 的 `go` 指令）。
- 2.2 Web：`gin`；配置：`sigs.k8s.io/yaml` + cobra/pflag（`/etc/platform/config.yaml`，`-c` 可覆盖）。
- 2.3 日志：`harns-platform/pkg/logging`（slog 门面，非 klog）。
- 2.4 存储：MySQL（`go-sql-driver/mysql`）、Redis（`redis/go-redis/v9`）、MinIO（`minio/minio-go/v7`）。
- 2.5 主键用 `oklog/ulid/v2`；密码哈希用 `golang.org/x/crypto/bcrypt`。
- 2.6 新增第三方依赖需说明理由，且兼容 Go 1.24。

## 3. 目录与分层

- 3.1 通用层（禁止被模块反向依赖）：`apis`、`generic`、`host`、`i18n`、`logging`、`tsdb`、`utils`、`version`。
- 3.2 模块层：业务代码放 `pkg/platform/`，入口在 `cmd/platform/`；禁止写进 `pkg/gateway` 或通用层。
- 3.3 模块内聚：身份与组织（用户/角色/组织/关联）合并为单包 `pkg/platform/identity/`（model/store/service/server/v1），同包直接调用，避免循环依赖。
- 3.4 模块间调用：工业应用为独立服务，通过 HTTP API 交互；platform 内部跨模块时才由消费方定义接口（contract）并在组合根注入，禁止全局单例。
- 3.5 路径约定：业务接口 `/api/v1/{app}/...`（当前 app=`iot`）；认证/当前用户等共享接口 `/api/v1/auth/...`；健康检查 `/api/v1/health`。审计按路径解析 `app`。

## 4. 通用编码约束

- 4.1 Go 官方风格；提交前 `gofmt`。
- 4.2 错误显式处理，禁止 `_ = err`、禁止 panic 处理可预期错误。
- 4.3 不在业务层直接 `c.JSON`，统一走第 5 节响应封装。
- 4.4 面向接口 + 依赖注入（构造函数传 `*config.Config`），便于测试。
- 4.5 面向用户的错误必须走第 6 节 i18n，禁止硬编码文案。

## 5. API 响应规范

- 5.1 统一返回 `pkg/apis/response` 的 `Result`：
  - 成功 `{"code":200,"message":"success","data":<payload>}`
  - 失败 `{"code":<业务码>,"message":"<本地化消息>","errors":[{"code":..,"message":..}]}`
- 5.2 成功 `response.WriteOK(c, data)`；失败 `response.WriteError(c, status, err...)`。
- 5.3 业务错误（参数校验、重复、不存在等）**HTTP 一律 400**，具体原因由 body 的 `code` 区分；HTTP 语义码只用于非业务场景：401 未认证、403 无权限、500 内部错误。
- 5.4 前端约定：`code===200` 用 `data`，否则用 `errors`。
- 5.5 分页：参数用 `apis.PageQuery`（`Normalize()` 后 `Offset()`/`Limit()`），返回 `apis.PageResp[T]`（`data` 含 `total/page/pageSize/list`）。
- 5.6 字段命名 lowerCamelCase；接口文档放 `apis/platform/`。

## 6. 错误码与国际化

- 6.1 错误码在 `pkg/apis/response/constants.go`（`ErrCode`，10001 起，只能末尾追加）。
- 6.2 在 `pkg/apis/response/errors.go` 的 `messageIDs` 登记语义 key（如 `error.user_not_found`）。
- 6.3 文案放 `pkg/i18n/locales/<lang>.yaml`（至少 en-US、zh-CN，占位符 `%s`/`%d`）。
- 6.4 用 `response.ErrXxx(...)` 构造，不在业务层拼字符串。
- 6.5 语言来自 `?lang=` 或 `Accept-Language`，用 `i18n.Lang(c)` 读取。

## 7. 配置规范

- 7.1 配置项在 `cmd/platform/options/Options`（`json` tag），默认值在 `NewDefaultOptions`。
- 7.2 同步更新 `etc/platform/config.yaml` 模板（含注释）。
- 7.3 敏感信息默认留空，不入库、不打日志。
- 7.4 优先级：CLI flag > 配置文件 > 内置默认（框架已实现，勿破坏）；启动时 `options.Validate` 校验。
- 7.5 认证配置 `security`（见 F4）：`token-secret` 必须 ≥32 字节，变更会使所有旧 token 失效；`rsa-enabled=true` 时启动依赖 MinIO。

## 8. 日志规范

- 8.1 用 `pkg/logging`：`logging.V(n).InfoS(...)`、`logging.ErrorS(err, ...)`；禁止 `fmt.Println`/标准库 `log`。
- 8.2 级别：V(1) 里程碑、V(2) 日常（默认可见）、V(3+) 细节；错误带 `err`。
- 8.3 文件输出与切割由 logging 配置控制（`max-size`/`max-backups`/`max-age`/`compress`），勿自行实现。
- 8.4 方法入口/出口：Service 层每个导出方法**第一行**写 `defer logging.Trace("identity.Xxx", kv...)()`；`Trace` 以 V(3) 输出 `enter`/`exit` 及 `cost`。
- 8.5 入口 kv 只带关键入参（如 `id`、`name`、`code`），**禁止打印 password/authKey 等敏感信息**。
- 8.6 handler 与 store 不重复记录入口/出口（避免日志风暴）；错误在出错处用 `logging.ErrorS`/`V(2).InfoS` 记录。
- 8.7 排查时把 `logging.verbosity` 调到 3 即可看到全部方法入口/出口。

## 9. 存储规范

- 9.1 数据访问用 `config.Config`：MySQL `DB`、Redis `Redis`、MinIO `MinIO`；SQL 参数化，Redis key 带业务前缀并设 TTL。
- 9.2 表结构变更用 SQL 迁移脚本，放 `migrations/`。
- 9.3 实体统一内嵌 `apis.Meta`（`id`/`version`/`createdUser`/`updatedUser`/`createdTime`/`updatedTime`，勿重复定义）；`id` 用 ULID（列宽 32）；`version` 每次更新 +1，`ETag`=version 仅作参考，不做 If-Match 校验。用户关联用 `user_roles`/`user_orgs`（无外键），删除时同事务清理。
- 9.4 审计字段：创建时 `CreatedUser`/`UpdatedUser`、更新时 `UpdatedUser` 一律取 `principal.UserID(ctx)`（鉴权中间件已注入）；禁止硬编码或留空。

## 10. 功能清单（按序号编写）

> 每个功能用如下格式，序号不可跳号；未实现前保持 `状态: TODO`。

### F1. 用户管理（CRUD + 分页）
- 状态：TODO
- 目标：用户创建、按 id 查询、更新、删除、分页查询 5 个 REST 接口。
- 输入：字段见下表。
- 处理：1) 迁移脚本建 `users`、`user_roles`、`user_orgs` 表；2) 实体 `User`（内嵌 `apis.Meta`）；3) 单包 `pkg/platform/identity/`（model/store/service/server/v1）；4) 路由注册到 `pkg/platform/web`，文档到 `apis/platform/user.yaml`。
- 输出：5 个接口 + 迁移脚本 + API 文档。
- 字段：

  | 字段 | Go 类型 | 约束 |
  |---|---|---|
  | userName | string | 必填，唯一 |
  | password | string | 可选，默认 `用户名+@123`；bcrypt 落库，响应不回传 |
  | userNo | string | 必填，唯一 |
  | nickName | string | 必填 |
  | birthDate | string `2006-01-02` | 可空 |
  | gender | string | 枚举 male/female |
  | mail | string | 可空 |
  | phone | string | 可空 |
  | note | string | 可空 |
  | authKey | string | 可空，响应不回传 |
  | authMode | string | 枚举 local/ldap/radius |
  | lastLoginTime | *time.Time | 可空 |
  | loginFailureTimes | int | 默认 0 |
  | expireTimes | int | 默认 0 |
  | status | string | VARCHAR，枚举 enabled/disabled |
  | roleIds | []string | 可多选，校验角色存在 |
  | orgIds | []string | 可多选，校验组织存在 |

  > 基础字段来自 `apis.Meta`（见 §9.3），不要 `deleted` 列。
- 接口：
  - POST `/api/v1/iot/users` → 200 + UserVo
  - GET `/api/v1/iot/users/:id` → 200 + UserVo
  - PUT `/api/v1/iot/users/:id` → 200 + UserVo
  - DELETE `/api/v1/iot/users/:id` → 200（物理删除，同事务清理关联）
  - PUT `/api/v1/iot/users/:id/password` → 200 + UserVo（重置密码为 `用户名+@123`）
  - GET `/api/v1/iot/users` → 200 + `PageResp[UserVo]`
- 约束：遵循 §3~§9；DTO 放 `pkg/platform/identity/v1`（`XxxReq`/`XxxQuery`）；对外一律 `XxxVo`，禁止返回实体（防泄露 password/authKey）；`createdUser`/`updatedUser` 取自登录上下文（`principal.UserID(ctx)`）。
- 验收标准：
  - [ ] 5 个接口按 §5.1 返回
  - [ ] 列表用 `PageQuery`/`PageResp`，`data` 含 total/page/pageSize/list
  - [ ] userName/userNo 重复 → 400 + 业务码 `error.user_name_exists`/`error.user_no_exists`；id 不存在 → 400 + `error.user_not_found`
  - [ ] 响应体不含 password/authKey；password 落库为 bcrypt
  - [ ] 创建时未传 password → 使用 `用户名+@123`；重置密码接口将密码置为 `用户名+@123`
  - [ ] 创建时 `roleIds`/`orgIds` 校验存在（不存在 → 400 `error.role_not_found`/`error.org_not_found`）
  - [ ] 用户 + `user_roles` + `user_orgs` 同一事务写入，失败整体回滚
  - [ ] `UserVo` 返回 `roleIds`/`orgIds`，以及关联的 `roles`/`orgs` 对象数组（列表亦是）
  - [ ] 删除用户同事务清理关联；更新后 `version` +1、`ETag` 反映新值
  - [ ] 错误文案支持 zh-CN/en-US；产出迁移脚本与 API 文档

### F2. 角色管理（CRUD + 分页）
- 状态：TODO
- 目标：角色创建、按 id 查询、更新、删除、分页查询 5 个 REST 接口。
- 输入：字段见下表。
- 处理：
  1. 迁移脚本建 `roles` 表；
  2. 实体 `Role`（内嵌 `apis.Meta`）；
  3. 单包 `pkg/platform/identity/`（model/store/service/server/v1）；
  4. 路由注册到 `pkg/platform/web`，文档到 `apis/platform/role.yaml`。
- 输出：5 个接口 + 迁移脚本 + API 文档。
- 字段：

  | 字段 | Go 类型 | 约束 |
  |---|---|---|
  | name | string | 必填，唯一 |
  | code | string | 必填，唯一 |
  | description | string | 可空 |
  | status | string | VARCHAR，枚举 enabled/disabled |

  > 基础字段来自 `apis.Meta`（见 §9.3），不要 `deleted` 列。
- 接口：
  - POST `/api/v1/iot/roles` → 200 + RoleVo
  - GET `/api/v1/iot/roles/:id` → 200 + RoleVo
  - PUT `/api/v1/iot/roles/:id` → 200 + RoleVo
  - DELETE `/api/v1/iot/roles/:id` → 200（物理删除，同事务清理 `user_roles`）
  - GET `/api/v1/iot/roles` → 200 + `PageResp[RoleVo]`
- 约束：遵循 §3~§9；DTO 放 `pkg/platform/identity/v1`（`XxxReq`/`XxxQuery`）；对外一律 `RoleVo`（无敏感字段，含全部业务字段）；`createdUser`/`updatedUser` 取自登录上下文（`principal.UserID(ctx)`）。
- 验收标准：
  - [ ] 5 个接口按 §5.1 返回
  - [ ] 列表用 `PageQuery`/`PageResp`，`data` 含 total/page/pageSize/list
  - [ ] name/code 重复 → 400 + `error.role_name_exists`/`error.role_code_exists`；id 不存在 → 400 + `error.role_not_found`
  - [ ] 删除角色同事务清理 `user_roles`；更新后 `version` +1
  - [ ] 错误文案支持 zh-CN/en-US；产出迁移脚本与 API 文档

### F3. 组织(部门)管理（CRUD + 树状查询）
- 状态：TODO
- 目标：组织创建、按 id 查询、更新、删除、树状查询 5 个接口（**无分页**）。
- 输入：字段见下表。
- 处理：
  1. 迁移脚本建 `orgs` 表；
  2. 实体 `Org`（内嵌 `apis.Meta`）；
  3. 单包 `pkg/platform/identity/`（model/store/service/server/v1）；
  4. 路由注册到 `pkg/platform/web`，文档到 `apis/platform/org.yaml`。
- 输出：5 个接口 + 迁移脚本 + API 文档。
- 字段：

  | 字段 | Go 类型 | 约束 |
  |---|---|---|
  | code | string | 必填，唯一 |
  | name | string | 必填 |
  | parentId | string | 可空，空/0 为根 |
  | sort | int | 排序，默认 0 |
  | description | string | 可空 |
  | remark | string | 可空 |
  | status | string | VARCHAR，枚举 enabled/disabled |

  > 基础字段来自 `apis.Meta`（见 §9.3），不要 `deleted` 列。
- 接口：
  - POST `/api/v1/iot/orgs` → 200 + OrgVo
  - GET `/api/v1/iot/orgs/:id` → 200 + OrgVo
  - PUT `/api/v1/iot/orgs/:id` → 200 + OrgVo
  - DELETE `/api/v1/iot/orgs/:id` → 200（物理删除；**存在子节点则拒绝**；同事务清理 `user_orgs`）
  - GET `/api/v1/iot/orgs/tree` → 200 + `[]OrgVo`（树，`children` 嵌套）
- 约束：遵循 §3~§9；DTO 放 `pkg/platform/identity/v1`；对外一律 `OrgVo`（树状携带 `children`）；`createdUser`/`updatedUser` 取自登录上下文（`principal.UserID(ctx)`）；更新 `parentId` 需防成环（不能设为自己或其后代）。
- 验收标准：
  - [ ] 5 个接口按 §5.1 返回
  - [ ] 树状查询返回根节点数组，`children` 嵌套
  - [ ] code 重复 → 400 + `error.org_code_exists`；id 不存在 → 400 + `error.org_not_found`
  - [ ] 父节点不存在 → 400 + `error.org_parent_not_found`；parentId 设为自己/后代 → 400 + `error.org_parent_cycle`
  - [ ] 删除存在子节点的组织 → 400 + `error.org_has_children`
  - [ ] 删除组织同事务清理 `user_orgs`；更新后 `version` +1
  - [ ] 错误文案支持 zh-CN/en-US；产出迁移脚本与 API 文档

### F4. 认证
- 状态：TODO
- 目标：登录、登出、当前用户、公钥；JWT(HS256) + Redis 会话。
- 接口（`/api/v1`）：
  - POST `/login` → `{token}`；body `{username,password,code?,key?}`
  - POST `/verificationCode` → `{uuid,img}`（数学验证码，base64 图片）
  - POST `/logout` → 删除会话（幂等）
  - GET `/index?refresh=` → 当前用户（含 roles/orgs）；`refresh=true` 查库并写回 Redis
  - GET `/publicKey` → Base64(X509 DER) 公钥
- 规则：
  - 创建用户密码 bcrypt；登录用 `bcrypt.CompareHashAndPassword`。
  - 密码错 5 次 → 锁定 1 小时（`status=disabled`、`locked_until=+60m`）；锁定返回 `errors[0].data.unlockAt`（Unix 毫秒）。
  - 会话 TTL 默认 60 分钟（用户 `expireTimes` 覆盖；<0 永不过期）；剩余 <10 分钟时在 filter 中滑动续期。
  - JWT 只含 `login_user`(会话UUID)/`sub`/`iat`，不设 `exp`；`token-secret` ≥32 字节。
  - RSA：启动时在 MinIO 检查 `login-rsa-private.pem`/`login-rsa-public.pem`，缺失才生成；MinIO 不可达 → 启动失败；解密用 PKCS1v15，失败按明文。
  - 验证码：生成后存 Redis `platform:captcha:<uuid>`（TTL `captcha-ttl-seconds`，默认 300s）；**登录中的校验默认注释未启用**（`auth/service.go` 的 login）。
  - Trace：由 `httpx.Trace()` 独立处理（`X-Trace-Id` 生成/回写 + 请求结果日志），与认证解耦；`auth.Auth()` 只管认证。
  - 上下文：认证成功后把**完整用户**（`identity/v1.UserVo`，含 roles/orgs）注入 `c.Request.Context()`；业务层用 `principal.UserID(ctx)` / `principal.FromContext(ctx)` 取用户信息（见 §9.4）。
- 约束：遵循 §3~§9；受保护接口（users/roles/orgs）挂鉴权中间件；`health`/login/logout/publicKey 放行。
- 错误码：`10030 login_expired(401)`、`10031 unauthorized(401)`、`10032 username_password_error(400)`、`10033 account_disabled(400)`、`10034 account_locked(400)`、`10035 login_failed(500)`。
- 验收标准：
  - [ ] 登录成功返回 token；密码错 400 + 业务码
  - [ ] 连续 5 次失败锁定 1h，返回 `unlockAt`（毫秒）；1h 后自动解锁
  - [ ] 无/过期 token 访问受保护接口 → 401 `login_expired`
  - [ ] 登出后原 token 失效
  - [ ] filter 滑动续期（剩余 <10min 重置 60min）
  - [ ] `index?refresh=true` 查库并写回 Redis；`false` 读 Redis
  - [ ] 验证码接口返回 `{uuid,img}`，答案存 Redis（TTL 300s）；登录校验默认注释未启用
  - [ ] 受保护请求响应头返回 `X-Trace-Id`，日志包含 traceId、校验结果与耗时
  - [ ] RSA 加密密码与明文均可登录；MinIO 不可达启动失败

### F5. 审计日志
- 状态：TODO
- 目标：记录写操作的入参/出参，支持查询。
- 中间件（拆分）：
  - `httpx.Trace()`：traceId（`X-Trace-Id`）生成/透传。
  - `auth.Auth()`：仅认证；从 `httpx.TraceID` 取 trace。
  - `audit.Middleware()`：捕获入参/出参并异步入库。
  - 顺序：`Trace → Auth → Audit → handler`。
- 审计范围：仅 `POST/PUT/PATCH/DELETE`（含 login/logout）；`GET` 不记录。排除 `health`/`auth/publicKey`/`auth/verificationCode`。
- 落库（异步）：有界队列；满则丢弃并告警；失败不影响响应。
- 字段：`app`（由路径解析，auth 记 `platform`）、`traceId`、`userId/userName`（`principal`）、`operationType/operationDesc`（路由 `httpx.WithOperation` 显式声明）、`method/path/query/ip/userAgent`、`requestBody/responseBody`、`status/resultCode/errorMsg`、`durationMs`。
- 入参/出参：仅 JSON；按 16KB 截断；敏感键（password/authKey/token/accessKey/secretKey/ak/sk/secret）脱敏为 `***`；非 JSON 记 `<non-json body omitted>`。
- 查询：`GET /api/v1/iot/audit-logs`（分页 + app/userId/path/method/resultCode/start/end）。
- 验收标准：
  - [ ] 写操作落 `audit_logs`，GET 不落
  - [ ] app/traceId/userId/operationType/入参/出参/耗时/业务码正确
  - [ ] 敏感字段脱敏、16KB 截断
  - [ ] 异步不阻塞；队列满丢弃并告警
  - [ ] 登录/登出审计 app=`platform`，业务审计 app=`iot`
  - [ ] 审计查询分页与筛选可用

<!-- 继续 F6 ... -->

## 11. 任务提示词模板（每次 vibe coding 用）

> 连同本文件发给 AI，明确「只做某一个序号」。

```
请先阅读 pkg/platform/platform.md 与涉及的通用包源码，然后实现 F<序号>。

范围：
- 只改动：<允许改的文件/目录>
- 不要改动：gateway 模块、通用层语义、其他功能

要求：
- 遵守 platform.md §3~§9（分层、响应、错误码/i18n、配置、日志、存储）
- 复用现有框架：response.WriteOK/WriteError、i18n、logging、config.Config
- 新增错误码按 §6 同时补 constants.go、messageIDs、locales
- 新增配置项按 §7 同步 options 与 etc/platform/config.yaml

验收：
- 满足 F<序号> 的验收标准
- go build ./... 通过；gofmt 干净
- 有接口时附请求/响应示例

请先给出实现计划与改动文件清单，再动手写代码。
```

## 12. 禁止事项

- 12.1 引入 klog/component-base 等 k8s 日志链路（日志统一 `pkg/logging`）。
- 12.2 业务层返回裸对象（必须包 `Result`）。
- 12.3 硬编码文案、魔法数、明文密钥。
- 12.4 破坏 `code===200?data:errors` 约定。
- 12.5 通用层 import 模块层。

## 13. 交付自检清单

- [ ] 功能序号改为 DONE
- [ ] `go build ./...` 通过、`gofmt -l` 无输出
- [ ] 新增错误码/文案/配置齐全
- [ ] Service 层方法入口/出口已加 `logging.Trace`（见 §8.4）
- [ ] 有接口时附请求/响应示例
- [ ] 未触碰禁止事项
