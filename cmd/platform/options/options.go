package options

import (
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
)

type Options struct {
	Port     string               `json:"port"`
	Wait     baseoptions.Duration `json:"graceful-timeout"`
	CertFile string               `json:"cert-file"`
	KeyFile  string               `json:"key-file"`
	Redis    RedisOptions         `json:"redis"`
	MySQL    MySQLOptions         `json:"mysql"`
	MinIO    MinIOOptions         `json:"minio"`
	baseoptions.BaseOptions
}

type RedisOptions struct {
	Addrs    []string `json:"addrs"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	DB       int      `json:"db"`
}

type MySQLOptions struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
	Charset  string `json:"charset"`
}

type MinIOOptions struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"access-key"`
	SecretKey string `json:"secret-key"`
	UseSSL    bool   `json:"use-ssl"`
	Bucket    string `json:"bucket"`
}

const (
	_defaultPort          = "32300"
	_defaultWait          = baseoptions.Duration(15 * time.Second)
	_defaultConfigFile    = "/etc/platform/config.yaml"
	_defaultMySQLHost     = "127.0.0.1"
	_defaultMySQLPort     = 3306
	_defaultMySQLCharset  = "utf8mb4"
	_defaultMinIOEndpoint = "127.0.0.1:9000"
	_defaultMinIOBucket   = "harns"
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
			Addrs: _defaultRedisAddrs,
		},
		MySQL: MySQLOptions{
			Host:    _defaultMySQLHost,
			Port:    _defaultMySQLPort,
			Charset: _defaultMySQLCharset,
		},
		MinIO: MinIOOptions{
			Endpoint: _defaultMinIOEndpoint,
			Bucket:   _defaultMinIOBucket,
		},
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
	}

	c.Redis = redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:    o.Redis.Addrs,
		Username: o.Redis.Username,
		Password: o.Redis.Password,
		DB:       o.Redis.DB,
	})

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		o.MySQL.Username, o.MySQL.Password, o.MySQL.Host, o.MySQL.Port, o.MySQL.Database, o.MySQL.Charset)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	c.DB = db

	minioClient, err := minio.New(o.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(o.MinIO.AccessKey, o.MinIO.SecretKey, ""),
		Secure: o.MinIO.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	c.MinIO = minioClient

	return c, nil
}
