package options

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/pflag"

	_ "github.com/go-sql-driver/mysql"

	"harns-platform/cmd/platform/config"
	baseoptions "harns-platform/pkg/generic/options"
	"harns-platform/pkg/platform/audit"
	"harns-platform/pkg/platform/auth"
	"harns-platform/pkg/platform/migrate"
)

type Options struct {
	Port     string               `json:"port"`
	Wait     baseoptions.Duration `json:"graceful-timeout"`
	CertFile string               `json:"cert-file"`
	KeyFile  string               `json:"key-file"`
	Redis    RedisOptions         `json:"redis"`
	MySQL    MySQLOptions         `json:"mysql"`
	MinIO    MinIOOptions         `json:"minio"`
	Security auth.Config          `json:"security"`
	Audit    audit.Config         `json:"audit"`
	Migrate  bool                 `json:"migrate"`
	baseoptions.BaseOptions
}

type RedisOptions struct {
	Addrs               []string `json:"addrs"`
	Username            string   `json:"username"`
	Password            string   `json:"password"`
	DB                  int      `json:"db"`
	ReadTimeoutSeconds  int      `json:"read-timeout-seconds"`
	WriteTimeoutSeconds int      `json:"write-timeout-seconds"`
}

type MySQLOptions struct {
	Host                   string `json:"host"`
	Port                   int    `json:"port"`
	Username               string `json:"username"`
	Password               string `json:"password"`
	Database               string `json:"database"`
	Charset                string `json:"charset"`
	MaxOpenConns           int    `json:"max-open-conns"`
	MaxIdleConns           int    `json:"max-idle-conns"`
	ConnMaxLifetimeMinutes int    `json:"conn-max-lifetime-minutes"`
}

type MinIOOptions struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access-key"`
	SecretKey string `json:"secret-key"`
	UseSSL    bool   `json:"use-ssl"`
	Bucket    string `json:"bucket"`
}

const (
	_defaultPort                        = "32300"
	_defaultWait                        = baseoptions.Duration(15 * time.Second)
	_defaultConfigFile                  = "/etc/platform/config.yaml"
	_defaultMySQLHost                   = "127.0.0.1"
	_defaultMySQLPort                   = 3306
	_defaultMySQLCharset                = "utf8mb4"
	_defaultMySQLMaxOpenConns           = 50
	_defaultMySQLMaxIdleConns           = 10
	_defaultMySQLConnMaxLifetimeMinutes = 30
	_defaultRedisTimeoutSeconds         = 3
	_defaultMinIOEndpoint               = "127.0.0.1:9000"
	_defaultMinIOBucket                 = "harns"
)

var (
	_defaultRedisAddrs = []string{"127.0.0.1:6379"}
)

func NewDefaultOptions() *Options {
	base := baseoptions.NewDefaultBaseOptions()
	base.ConfigFile = _defaultConfigFile
	return &Options{
		Port:        _defaultPort,
		Wait:        _defaultWait,
		CertFile:    "",
		KeyFile:     "",
		BaseOptions: base,
		Redis: RedisOptions{
			Addrs:               _defaultRedisAddrs,
			ReadTimeoutSeconds:  _defaultRedisTimeoutSeconds,
			WriteTimeoutSeconds: _defaultRedisTimeoutSeconds,
		},
		MySQL: MySQLOptions{
			Host:                   _defaultMySQLHost,
			Port:                   _defaultMySQLPort,
			Charset:                _defaultMySQLCharset,
			MaxOpenConns:           _defaultMySQLMaxOpenConns,
			MaxIdleConns:           _defaultMySQLMaxIdleConns,
			ConnMaxLifetimeMinutes: _defaultMySQLConnMaxLifetimeMinutes,
		},
		MinIO: MinIOOptions{
			Endpoint: _defaultMinIOEndpoint,
			Bucket:   _defaultMinIOBucket,
		},
		Security: auth.DefaultConfig(),
		Audit:    audit.DefaultConfig(),
		Migrate:  true,
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Port, "port", "P", o.Port, "Port exposed")
	baseoptions.DurationVar(fs, &o.Wait, "graceful-timeout", "The duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	fs.StringVarP(&o.CertFile, "cert-file", "", o.CertFile, "The Cert file")
	fs.StringVarP(&o.KeyFile, "key-file", "", o.KeyFile, "The Key file")
}

func (o *Options) Config(stopCh <-chan struct{}) (*config.Config, error) {
	c := &config.Config{
		CertFile:    o.CertFile,
		KeyFile:     o.KeyFile,
		MinIOBucket: o.MinIO.Bucket,
		Security:    o.Security,
		Audit:       o.Audit,
	}

	c.Redis = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:        o.Redis.Addrs,
		Username:     o.Redis.Username,
		Password:     o.Redis.Password,
		DB:           o.Redis.DB,
		ReadTimeout:  time.Duration(o.Redis.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(o.Redis.WriteTimeoutSeconds) * time.Second,
	})
	redisCtx, cancelRedis := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelRedis()
	if err := c.Redis.Ping(redisCtx).Err(); err != nil {
		return nil, fmt.Errorf("connect redis failed: %w", err)
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local&timeout=5s&readTimeout=10s&writeTimeout=10s",
		o.MySQL.Username, o.MySQL.Password, o.MySQL.Host, o.MySQL.Port, o.MySQL.Database, o.MySQL.Charset)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql failed: %w", err)
	}
	db.SetMaxOpenConns(o.MySQL.MaxOpenConns)
	db.SetMaxIdleConns(o.MySQL.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(o.MySQL.ConnMaxLifetimeMinutes) * time.Minute)

	pingCtx, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("connect mysql failed: %w", err)
	}
	c.DB = db

	if o.Migrate {
		migCtx, cancelMig := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancelMig()
		if err := migrate.Run(migCtx, db); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("run migrations failed: %w", err)
		}
	}

	minioClient, err := minio.New(o.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(o.MinIO.AccessKey, o.MinIO.SecretKey, ""),
		Secure: o.MinIO.UseSSL,
	})
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	c.MinIO = minioClient

	return c, nil
}
