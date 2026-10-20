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
- 启动时会 Ping MySQL/Redis；`migrate: true` 时执行内嵌迁移（`migrations/*.sql`）。

### 接口

- 认证（共享）：`POST /api/v1/auth/login`、`POST /api/v1/auth/logout`、
  `GET /api/v1/auth/index`、`GET /api/v1/auth/publicKey`、
  `POST /api/v1/auth/verificationCode`
- 业务（app=iot）：`/api/v1/iot/users`、`/roles`、`/orgs`、`/audit-logs`
- 探针：`GET /api/v1/health`、`GET /api/v1/ready`

### 认证

- 迁移会初始化管理员：`admin / admin@123`（首次登录后请修改）。
- JWT(HS256) + Redis 会话；密码支持 RSA-OAEP(SHA-256) 加密。
- 连续 5 次登录失败锁定 60 分钟。

### 审计

- 仅记录写操作；入参/出参脱敏并按 16KB 截断。
- 日志保留 `audit.retention-days`（默认 180 天）。

### 开发

```bash
make fmt     # gofmt
make lint    # gofmt + go vet（排除 tsdb）
make test    # go test
make package # 构建并打包二进制与 etc/ 配置
```
