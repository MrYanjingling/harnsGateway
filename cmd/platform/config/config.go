package config

import (
	"database/sql"

	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"

	"harns-platform/pkg/platform/audit"
	"harns-platform/pkg/platform/auth"
	"harns-platform/pkg/platform/httpx"
)

type Config struct {
	CertFile string
	KeyFile  string

	Redis       redis.UniversalClient
	DB          *sql.DB
	MinIO       *minio.Client
	MinIOBucket string
	Security    auth.Config
	CORS        httpx.CORSConfig
	Audit       audit.Config
}
