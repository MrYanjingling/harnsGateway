**其他语言版本: [English](README.md), [中文](README_zh.md).**

## harnsGateway是什么

HarnsGateway 可以安装在嵌入式设备上作为硬网关采集工业数据, 也可以安装在边缘系统上面作为软网关采集数据.

## harnsGateway主要功能是什么

* **从南端采集数据**  
  支持的协议列表:

1. ModbusTcp ModbusRtu ModbusRtuOverTcp
2. S71500
3. OpcUA

* **获取北端输入反控设备**
  支持的协议列表:

1. ModbusTcp ModbusRtu ModbusRtuOverTcp
2. S71500

* **边缘计算**

## 如何构建

1. git clone https://github.com/harnsFactory/harnsgateway.git
2. cd harnsgateway
3. make
4. cd harnsgateway/_output/bin/

## 如何启动

1. QuickStart</br> ./gateway --mqtt-broker-urls=127.0.0.1:1883 &
2. Systemd

## 如何使用

例如 **连接Modbus设备**

1. 打开Modbus模拟软件(ModSim32) 并且更新如下参数: deviceId = 1,functionCode =
   3,并且设置第一个寄存器的值等于188.</br>[第一步](https://postimg.cc/sBFyrN2M) </br>然后通过502端口启动服务.
2. 在harnsGateway上创建一个Modbus设备([接口文档](apis/gateway/create-modbus-device.yaml)).设备的id属性会在MQTT topic中使用.</br> [第二步](https://postimg.cc/svYFZdpy)
3. 获取harnsGateway元信息( [接口文档](apis/gateway/gateway.yaml) ).网关的id属性会在MQTT topic中使用.</br> [第三步](https://postimg.cc/GHYxf9zP)
4. 订阅MQTT topic.</br> [第四步](https://postimg.cc/ppTGRwqq) </br>Topic为'data/{gatewayId}/v1/{deviceId}'.
5. 删除设备.

## 如何启动测试用例



## 平台服务（harns-platform）

仓库同时包含物联网平台服务（`cmd/platform`）。

### 构建与启动

```bash
make                                   # 构建所有服务到 _output/bin
# 或
go build -o platform ./cmd/platform

cp etc/platform/config.yaml /etc/platform/config.yaml
./platform -c /etc/platform/config.yaml
```

### 依赖

- 需要 MySQL、Redis、MinIO。
- 启动时会 Ping MySQL/Redis；`migrate: true` 时用 golang-migrate 执行内嵌迁移（`migrations/NNNNNN_name.up.sql`）。

### 接口

- 认证（共享）：`POST /api/v1/auth/login`、`POST /api/v1/auth/logout`、
  `GET /api/v1/auth/index`、`GET /api/v1/auth/publicKey`、
  `POST /api/v1/auth/verificationCode`
- 业务（app=IOT）：`/api/v1/iot/users`、`/roles`、`/orgs`、`/auditLogs`
- 探针：`GET /api/v1/health`、`GET /api/v1/ready`（检查 DB、Redis、MinIO）
- 文档：`GET /swagger`、`GET /api/v1/openapi`、`GET /api/v1/openapi/{file}`
- 运维：`GET /metrics`（Prometheus）、`GET /debug/pprof/`

### 认证

- 迁移会初始化超级管理员角色（`code=admin`）与账号 `admin / admin@qwer`（首次登录后请修改）。
- JWT(HS256) + Redis 会话；密码支持 RSA-OAEP(SHA-256) 加密。
- 连续 5 次登录失败锁定 60 分钟。

### 审计

- 仅记录写操作；入参/出参脱敏并按 16KB 截断。
- 日志异步批量写入（`audit.batch-size` 默认 100，`audit.flush-interval-ms` 默认 1000），
  保留 `audit.retention-days`（默认 180 天），清理按 `audit.cleanup-batch-size` 分批删除。
- `app` 表示来源服务：`IOT`、`EMS`、`EAM`。EMS/EAM 使用同一审计模块，各自绑定自己的 app。

同仓库的其他 Go 服务可直接使用 `pkg/audit`：

```go
cfg := audit.DefaultConfig()
m := audit.NewModule(db, cfg, audit.AppEMS)
defer m.Close(ctx)

_ = m.Record(ctx, audit.Event{OperationType: audit.OpCreate, Path: "/api/v1/ems/..."})
list, total, err := m.List(ctx, audit.Query{PageSize: 20})
```

表结构和 180 天清理只由 platform 负责。服务需要安装 `generic.Trace()`，出站调用使用
`generic.SetTraceHeader(ctx, req)` 透传 `traceId`，这样不同服务的记录可以关联。

### 开发

```bash
make fmt     # gofmt
make lint    # gofmt + go vet（排除 tsdb）
make test    # go test
make package # 构建并打包二进制与 etc/ 配置
```

### Docker

```bash
docker compose up --build
# 平台：http://127.0.0.1:32300
# 文档：http://127.0.0.1:32300/swagger
# MinIO：http://127.0.0.1:9001
```

compose 会启动 MySQL、Redis、MinIO 与平台服务。正式环境请先修改默认凭据与
`security.token-secret`。
