module harns-platform

go 1.24.0

require (
	github.com/eclipse/paho.mqtt.golang v1.4.3
	github.com/evanphx/json-patch v5.6.0+incompatible
	github.com/gin-contrib/gzip v1.0.1
	github.com/gin-gonic/gin v1.9.1
	github.com/go-playground/validator/v10 v10.20.0
	github.com/go-sql-driver/mysql v1.10.1
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/golang-migrate/migrate/v4 v4.19.0
	github.com/google/uuid v1.6.0
	github.com/gopcua/opcua v0.5.1
	github.com/minio/minio-go/v7 v7.0.95
	github.com/mitchellh/mapstructure v1.5.0
	github.com/mojocn/base64Captcha v1.3.8
	github.com/oklog/ulid/v2 v2.1.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.23.2
	github.com/redis/go-redis/v9 v9.20.0
	github.com/shirou/gopsutil/v3 v3.23.10
	github.com/spf13/cobra v1.10.2
	github.com/spf13/pflag v1.0.10
	github.com/stretchr/testify v1.12.1
	go.bug.st/serial v1.6.1
	go.opentelemetry.io/otel v1.37.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.37.0
	go.opentelemetry.io/otel/sdk v1.37.0
	go.uber.org/atomic v1.12.0
	golang.org/x/crypto v0.41.0
	golang.org/x/mod v0.26.0
	golang.org/x/sys v0.35.0
	golang.org/x/text v0.28.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	k8s.io/apimachinery v0.28.1
	sigs.k8s.io/yaml v1.3.0
)

require (
	filippo.io/edwards25519 v1.2.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/sonic v1.11.6 // indirect
	github.com/bytedance/sonic/loader v0.1.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.2 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/creack/goselect v0.1.2 // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.19.2 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/minio/crc64nvme v1.1.1 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/tinylib/msgp v1.6.4 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.37.0 // indirect
	go.opentelemetry.io/otel/metric v1.37.0 // indirect
	go.opentelemetry.io/otel/trace v1.37.0 // indirect
	go.opentelemetry.io/proto/otlp v1.7.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/arch v0.7.0 // indirect
	golang.org/x/exp v0.0.0-20230522175609-2e198f4a06a1 // indirect
	golang.org/x/image v0.23.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250603155806-513f23925822 // indirect
	google.golang.org/grpc v1.73.0 // indirect
	google.golang.org/protobuf v1.36.12 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/klog/v2 v2.100.1 // indirect
	k8s.io/utils v0.0.0-20230406110748-d93618cff8a2 // indirect
)
