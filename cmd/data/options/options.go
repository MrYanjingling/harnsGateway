package options

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/go-sql-driver/mysql"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"harnsgateway/cmd/data/config"
	"harnsgateway/pkg/data"

	baseoptions "harnsgateway/pkg/generic/options"
)

type Options struct {
	Port           string        `json:"port"`
	Wait           time.Duration `json:"graceful-timeout"`
	MySQLHost      string        `json:"mysql-host"`
	MySQLPort      string        `json:"mysql-port"`
	MySQLDB        string        `json:"mysql-db"`
	MySQLUser      string        `json:"mysql-user"`
	MySQLPassword  string        `json:"mysql-password"`
	MySQLMaxActive int           `json:"mysql-max-active"`
	ESURLs         []string      `json:"es-urls"`
	ESUsername     string        `json:"es-username"`
	ESPassword     string        `json:"es-password"`
	ScadaURL       string        `json:"scada-url"`
	EScadaURL      string        `json:"escada-url"`
	CertFile       string        `json:"cert-file"`
	KeyFile        string        `json:"key-file"`
	baseoptions.BaseOptions
}

const (
	defaultPort           = "32201"
	defaultWait           = 15 * time.Second
	defaultMySQLHost      = "192.168.10.71"
	defaultFMCSScadaUrl   = "https://192.168.10.71:8591"
	defaultEScadaUrl      = "https://192.168.10.71:8589"
	defaultMySQLPort      = "8597"
	defaultMySQLDB        = "ems_base"
	defaultMySQLUser      = "root"
	defaultMySQLPassword  = "Di@mysql#123456"
	defaultMySQLMaxActive = 50
	defaultESUsername     = "ems"
	defaultESPassword     = "aEFCaY6OR2MPQLhdx1GX"
)

var (
	defaultESURLs = []string{"http://192.168.10.71:8598"}
)

func NewDefaultOptions() *Options {
	return &Options{
		Port:           defaultPort,
		Wait:           defaultWait,
		MySQLHost:      defaultMySQLHost,
		MySQLPort:      defaultMySQLPort,
		MySQLDB:        defaultMySQLDB,
		MySQLUser:      defaultMySQLUser,
		MySQLPassword:  defaultMySQLPassword,
		MySQLMaxActive: defaultMySQLMaxActive,
		ScadaURL:       defaultFMCSScadaUrl,
		EScadaURL:      defaultEScadaUrl,
		ESURLs:         defaultESURLs,
		ESUsername:     defaultESUsername,
		ESPassword:     defaultESPassword,
		BaseOptions:    baseoptions.NewDefaultBaseOptions(),
		CertFile:       "",
		KeyFile:        "",
	}
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&o.Port, "port", "P", o.Port, "Port exposed by the data service")
	fs.DurationVar(&o.Wait, "graceful-timeout", o.Wait, "The duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	fs.StringVarP(&o.MySQLHost, "mysql-host", "", o.MySQLHost, "MySQL host")
	fs.StringVarP(&o.MySQLPort, "mysql-port", "", o.MySQLPort, "MySQL port")
	fs.StringVarP(&o.MySQLDB, "mysql-db", "", o.MySQLDB, "MySQL database name")
	fs.StringVarP(&o.MySQLUser, "mysql-user", "", o.MySQLUser, "MySQL username")
	fs.StringVarP(&o.MySQLPassword, "mysql-password", "", o.MySQLPassword, "MySQL password")
	fs.IntVarP(&o.MySQLMaxActive, "mysql-max-active", "", o.MySQLMaxActive, "MySQL max active connections")
	fs.StringSliceVarP(&o.ESURLs, "es-urls", "", o.ESURLs, "Elasticsearch URLs, e.g. http://127.0.0.1:9200")
	fs.StringVarP(&o.ESUsername, "es-username", "", o.ESUsername, "Elasticsearch username")
	fs.StringVarP(&o.ESPassword, "es-password", "", o.ESPassword, "Elasticsearch password")
	fs.StringVarP(&o.ScadaURL, "scada-url", "", o.ScadaURL, "Scada third-party API base URL, e.g. https://scada.example.com")
	fs.StringVarP(&o.EScadaURL, "escada-url", "", o.EScadaURL, "EScada third-party API base URL, e.g. https://scada.example.com")
	fs.StringVarP(&o.CertFile, "cert-file", "", o.CertFile, "TLS certificate file")
	fs.StringVarP(&o.KeyFile, "key-file", "", o.KeyFile, "TLS key file")
}

func (o *Options) buildMySQLDSN() string {
	cfg := mysql.NewConfig()
	cfg.User = o.MySQLUser
	cfg.Passwd = o.MySQLPassword
	cfg.Net = "tcp"
	cfg.Addr = fmt.Sprintf("%s:%s", o.MySQLHost, o.MySQLPort)
	cfg.DBName = o.MySQLDB
	cfg.ParseTime = true
	cfg.Loc = time.Local
	cfg.Params = map[string]string{
		"charset": "utf8mb4",
	}
	cfg.TLSConfig = "false"
	return cfg.FormatDSN()
}

func (o *Options) Config(stopCh <-chan struct{}) (*config.Config, error) {
	c := &config.Config{}

	// Initialize MySQL
	if o.MySQLHost != "" {
		dsn := o.buildMySQLDSN()
		db, err := sql.Open("mysql", dsn)
		if err != nil {
			klog.ErrorS(err, "Failed to open MySQL connection")
			return nil, err
		}
		db.SetMaxOpenConns(o.MySQLMaxActive)
		db.SetMaxIdleConns(o.MySQLMaxActive / 5)
		db.SetConnMaxLifetime(5 * time.Minute)
		if err := db.Ping(); err != nil {
			klog.ErrorS(err, "Failed to ping MySQL")
			return nil, err
		}
		c.DB = db
		klog.InfoS("Connected to MySQL", "host", o.MySQLHost, "port", o.MySQLPort, "db", o.MySQLDB)
	}

	// Initialize Elasticsearch
	esCfg := elasticsearch.Config{
		Addresses: o.ESURLs,
		Username:  o.ESUsername,
		Password:  o.ESPassword,
	}
	es, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		klog.ErrorS(err, "Failed to create Elasticsearch client")
		return nil, err
	}
	res, err := es.Ping()
	if err != nil {
		klog.ErrorS(err, "Failed to ping Elasticsearch")
		return nil, err
	}
	defer res.Body.Close()
	if res.IsError() {
		klog.ErrorS(nil, "Elasticsearch ping returned error", "status", res.Status())
		return nil, err
	}
	c.ES = es
	klog.InfoS("Connected to Elasticsearch", "urls", o.ESURLs)

	// Initialize DataManager
	c.ScadaURL = o.ScadaURL
	c.EscadaURL = o.EScadaURL
	dataMgr := data.NewDataManager(c.DB, c.ES, data.WithScadaURL(c.ScadaURL), data.WithEScadaURL(c.EscadaURL))
	dataMgr.Init()
	c.DataManager = dataMgr

	c.CertFile = o.CertFile
	c.KeyFile = o.KeyFile

	return c, nil
}
