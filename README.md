**Read this in other languages: [English](README.md), [中文](README_zh.md).**
## What's harnsGateway

HarnsGateway is used to connect industrial equipment.  
It can be installed on embedded devices to as edge industrial gateway.   
It can also be run as a service on the edge IoT platform to as soft gateway.

## What's the functions of harnsGateway

* **Collect equipment data from south end**  
  Supported protocol list:

1. ModbusTcp ModbusRtu ModbusRtuOverTcp
2. S71500
3. OpcUA

* **Control equipment from north end input**
  Supported protocol list:

1. ModbusTcp ModbusRtu ModbusRtuOverTcp
2. S71500

* **Edge computing**

## How to Build

1. git clone https://github.com/harnsFactory/harnsgateway.git
2. cd harnsgateway
3. make
4. cd harnsgateway/_output/bin/

## How to Start

1. QuickStart</br> ./gateway --mqtt-broker-urls=127.0.0.1:1883 &
2. Systemd

## How to Use

example **Connect Modbus device**

1. Started modbus simulator(ModSim32) And update some parameters: deviceId = 1,functionCode = 3,And set first address
   value = 188.</br>[stepOne](https://postimg.cc/sBFyrN2M) </br>Then start service on port 502.
2. Create modbus( [api doc](apis/gateway/create-modbus-device.yaml) )device in harnsGateway.The device id property userd in MQTT topic.</br> [stepTow.png](https://postimg.cc/svYFZdpy)
3. Get gateway meta information( [api doc](apis/gateway/gateway.yaml) ).The gateway id property used in MQTT topic.</br> [stepThree.png](https://postimg.cc/GHYxf9zP)
4. Subscript MQTT topic.</br> [stepFour.png](https://postimg.cc/ppTGRwqq) </br>Topic is 'data/{gatewayId}/v1/{deviceId}'.
5. Delete the Device.

## How to Run Test



## Platform Service (harns-platform)

The monorepo also hosts the IoT platform service (`cmd/platform`).

### Build and run

```bash
make                                   # build all services into _output/bin
# or
go build -o platform ./cmd/platform

cp etc/platform/config.yaml /etc/platform/config.yaml
./platform -c /etc/platform/config.yaml
```

### Dependencies

- MySQL, Redis and MinIO are required.
- On startup the service pings MySQL/Redis and applies embedded SQL migrations
  (golang-migrate, files `migrations/NNNNNN_name.up.sql`) when `migrate: true`.

### Endpoints

- Auth (shared): `POST /api/v1/auth/login`, `POST /api/v1/auth/logout`,
  `GET /api/v1/auth/index`, `GET /api/v1/auth/publicKey`,
  `POST /api/v1/auth/verificationCode`
- IoT (`app=IOT`): `/api/v1/iot/users`, `/roles`, `/orgs`, `/auditLogs`
- Probes: `GET /api/v1/health`, `GET /api/v1/ready` (DB, Redis and MinIO)
- Docs: `GET /swagger`, `GET /api/v1/openapi`, `GET /api/v1/openapi/{file}`
- Ops: `GET /metrics` (Prometheus), `GET /debug/pprof/`

### Auth

- A super administrator role (`code=admin`) and account `admin / admin@qwer` are seeded by
  migration; change the password after first login.
- JWT (HS256) + Redis session; password may be RSA-OAEP(SHA-256) encrypted.
- Account locks for 60 minutes after 5 consecutive failures.

### Audit

- Write operations only; request/response bodies are masked and truncated (16KB).
- Logs are written asynchronously in batches (`audit.batch-size`, default 100;
  `audit.flush-interval-ms`, default 1000) and retained for
  `audit.retention-days` (default 180 days), with cleanup batched via
  `audit.cleanup-batch-size`.
- `app` identifies the source service: `IOT`, `EMS` or `EAM`. EMS/EAM use the
  same shared audit module with their own `app`.

Other Go services in this repo can use `pkg/audit` directly:

```go
cfg := audit.DefaultConfig()
m := audit.NewModule(db, cfg, audit.AppEMS)
defer m.Close(ctx)

_ = m.Record(ctx, audit.Event{OperationType: audit.OpCreate, Path: "/api/v1/ems/..."})
list, total, err := m.List(ctx, audit.Query{PageSize: 20})
```

Only the platform owns migrations and retention cleanup. Install `generic.Trace()`
and call `generic.SetTraceHeader(ctx, req)` for outbound calls so records from
different services share the same `traceId`.

### Development

```bash
make fmt     # gofmt
make lint    # gofmt + go vet (tsdb excluded)
make test    # go test
make package # build + zip binaries and etc/ config
```

### Docker

```bash
docker compose up --build
# platform: http://127.0.0.1:32300
# swagger:  http://127.0.0.1:32300/swagger
# minio:    http://127.0.0.1:9001
```

The compose stack starts MySQL, Redis, MinIO and the platform. Override the
default credentials and `security.token-secret` before using it anywhere real.
