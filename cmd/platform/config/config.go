package config

import (
	"context"
	"database/sql"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"

	"harns-platform/pkg/audit"
	"harns-platform/pkg/generic"
	"harns-platform/pkg/platform/auth"
	"harns-platform/pkg/telemetry"
)

type Config struct {
	CertFile     string
	KeyFile      string
	MaxBodyBytes int64

	Redis       redis.UniversalClient
	DB          *sql.DB
	MinIO       *minio.Client
	MinIOBucket string
	Security    auth.Config
	CORS        generic.CORSConfig
	Audit       audit.Config
	Telemetry   telemetry.Config

	TelemetryShutdown func(context.Context) error
}
