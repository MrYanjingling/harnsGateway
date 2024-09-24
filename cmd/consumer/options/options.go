package options

import (
	"fmt"
	"github.com/boltdb/bolt"
	flag "github.com/spf13/pflag"
	"github.com/streadway/amqp"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"harnsgateway/cmd/consumer/config"
	"harnsgateway/pkg/data"
	baseoptions "harnsgateway/pkg/generic/options"
	"k8s.io/klog/v2"
	"time"
)

const (
	_defaultPort                = "32200"
	_defaultSqlServerUrl        = "192.168.10.17:1433"
	_defaultSqlServerUsername   = "sa"
	_defaultSqlServerPassword   = "Rockwell123"
	_defaultRabbitmqHost        = "192.168.167.41"
	_defaultRabbitmqPort        = "5672"
	_defaultRabbitmqUsername    = "guest"
	_defaultRabbitmqPassword    = "guest"
	_defaultRabbitmqExchange    = "spc"
	_defaultRabbitmqVirtualHost = "SPC"
	_defaultRabbitmqQueue       = "SPC_RET_MFG_DATA"
	_defaultWait                = time.Second * 15
)

var defaultDsn = "sqlserver://%s:%s@%s?database=Runtime"
var defaultMq = "amqp://%s:%s@%s:%s/%s"

type Options struct {
	Port                string        `json:"port"`
	SqlServerUrl        string        `json:"db-url"`
	SqlServerUsername   string        `json:"db-username"`
	SqlServerPassword   string        `json:"db-password"`
	RabbitmqHost        string        `json:"mq-host"`
	RabbitmqPort        string        `json:"mq-port"`
	RabbitmqUsername    string        `json:"mq-username"`
	RabbitmqPassword    string        `json:"mq-password"`
	RabbitmqExchange    string        `json:"mq-exchange"`
	RabbitmqVirtualHost string        `json:"mq-virtual-host"`
	RabbitmqQueue       string        `json:"mq-queue"`
	Wait                time.Duration `json:"graceful-timeout"`
	baseoptions.BaseOptions
}

func NewDefaultOptions() *Options {
	return &Options{
		SqlServerUrl:        _defaultSqlServerUrl,
		SqlServerUsername:   _defaultSqlServerUsername,
		SqlServerPassword:   _defaultSqlServerPassword,
		RabbitmqHost:        _defaultRabbitmqHost,
		RabbitmqPort:        _defaultRabbitmqPort,
		RabbitmqUsername:    _defaultRabbitmqUsername,
		RabbitmqPassword:    _defaultRabbitmqPassword,
		RabbitmqExchange:    _defaultRabbitmqExchange,
		RabbitmqVirtualHost: _defaultRabbitmqVirtualHost,
		RabbitmqQueue:       _defaultRabbitmqQueue,
		Wait:                _defaultWait,
		Port:                _defaultPort,
		BaseOptions:         baseoptions.NewDefaultBaseOptions(),
	}
}

func (o *Options) AddFlags(fs *flag.FlagSet) {
	fs.StringVarP(&o.Port, "port", "P", o.Port, "Port exposed")
	fs.StringVar(&o.SqlServerUrl, "sqlServer-url", o.SqlServerUrl, "The SqlServer URL")
	fs.StringVar(&o.SqlServerUsername, "username", o.SqlServerUsername, "The SqlServer username")
	fs.StringVar(&o.SqlServerPassword, "password", o.SqlServerPassword, "The SqlServer password")
	fs.StringVar(&o.RabbitmqHost, "mq-host", o.RabbitmqHost, "The rabbit host")
	fs.StringVar(&o.RabbitmqPort, "mq-port", o.RabbitmqPort, "The rabbit port")
	fs.StringVar(&o.RabbitmqUsername, "mq-username", o.RabbitmqUsername, "The rabbit username")
	fs.StringVar(&o.RabbitmqPassword, "mq-password", o.RabbitmqPassword, "The rabbit password")
	fs.StringVar(&o.RabbitmqExchange, "mq-exchange", o.RabbitmqExchange, "The rabbit exchange")
	fs.StringVar(&o.RabbitmqVirtualHost, "mq-virtual-host", o.RabbitmqVirtualHost, "The rabbit virtual host")
	fs.StringVar(&o.RabbitmqQueue, "mq-queue", o.RabbitmqQueue, "The rabbit queue")
	fs.DurationVar(&o.Wait, "graceful-timeout", o.Wait, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
}

func (o *Options) Config(stopCh <-chan struct{}) (*config.Config, error) {
	gc := &gorm.Config{
		CreateBatchSize: 5000,
	}

	sqlUrl := fmt.Sprintf(defaultDsn, o.SqlServerUsername, o.SqlServerPassword, o.SqlServerUrl)

	db, err := gorm.Open(sqlserver.Open(sqlUrl), gc)
	if err != nil {
		klog.V(1).InfoS("Failed to connect FMCS database", "err", err)
		return nil, err
	}

	mqUrl := fmt.Sprintf(defaultMq, o.RabbitmqUsername, o.RabbitmqPassword, o.RabbitmqHost, o.RabbitmqPort, o.RabbitmqVirtualHost)

	dial, err := amqp.Dial(mqUrl)
	if err != nil {
		klog.V(1).InfoS("Failed to connect MQ", "err", err)
		return nil, err
	} else {
		klog.V(3).InfoS("Success to connect MQ", "url", mqUrl)
	}

	store, err := bolt.Open("tag.db", 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		klog.V(1).InfoS("Failed to connect open bolt db", "err", err)
		return nil, err
	}

	manager := data.NewManager(db, dial, store, &data.MqConfig{
		Exchange:    o.RabbitmqExchange,
		VirtualHost: o.RabbitmqVirtualHost,
		Queue:       o.RabbitmqQueue,
	}, stopCh)

	manager.Init()

	c := &config.Config{
		DataMgr: manager,
	}

	return c, nil
}
