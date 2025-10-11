package options

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/go-sql-driver/mysql"
	flag "github.com/spf13/pflag"
	"harnsgateway/cmd/clickhouse/config"
	"harnsgateway/pkg/ck"
	baseoptions "harnsgateway/pkg/generic/options"
	"k8s.io/klog/v2"
	"time"
)

const (
	// _defaultCimDBUrl      = "10.121.47.11:9030"
	// _defaultCimDBUsername = "ems"
	// _defaultCimDBPassword = "Tm20@2024ems"
	// _defaultEmsDBUrl      = "10.122.70.36:8406"
	// _defaultEmsDBUsername = "root"
	// _defaultEmsDBPassword = "i+aSroC6ak"
	_defaultWait = time.Second * 15
	_defaultPort = "9919"
)

var defaultCimDsn = "%s:%s@tcp(%s)/bdw_prod?charset=utf8mb4&parseTime=True&loc=Local"
var defaultEmsDsn = "%s:%s@tcp(%s)/ems_base?charset=utf8mb4&parseTime=True&loc=Local"

type Options struct {
	CimDBUrl      string        `json:"cim-db-url"`
	CimDBUsername string        `json:"cim-db-username"`
	CimDBPassword string        `json:"cim-db-password"`
	EmsDBUrl      string        `json:"ems-db-url"`
	EmsDBUsername string        `json:"ems-db-username"`
	EmsDBPassword string        `json:"ems-db-password"`
	Wait          time.Duration `json:"graceful-timeout"`
	Port          string        `json:"port"`
	baseoptions.BaseOptions
}

func NewDefaultOptions() *Options {
	return &Options{
		// CimDBUrl:      _defaultCimDBUrl,
		// CimDBUsername: _defaultCimDBUsername,
		// CimDBPassword: _defaultCimDBPassword,
		// EmsDBUrl:      _defaultEmsDBUrl,
		// EmsDBUsername: _defaultEmsDBUsername,
		// EmsDBPassword: _defaultEmsDBPassword,
		Wait:        _defaultWait,
		Port:        _defaultPort,
		BaseOptions: baseoptions.NewDefaultBaseOptions(),
	}
}

func (o *Options) AddFlags(fs *flag.FlagSet) {
	fs.StringVar(&o.CimDBUrl, "cim-db-url", o.CimDBUrl, "The SqlServer URL")
	fs.StringVar(&o.CimDBUsername, "cim-db-username", o.CimDBUsername, "The SqlServer username")
	fs.StringVar(&o.CimDBPassword, "cim-db-password", o.CimDBPassword, "The SqlServer password")
	fs.StringVar(&o.EmsDBUrl, "ems-db-url", o.EmsDBUrl, "The rabbit host")
	fs.StringVar(&o.EmsDBUsername, "ems-db-username", o.EmsDBUsername, "The rabbit port")
	fs.StringVar(&o.EmsDBPassword, "ems-db-password", o.EmsDBPassword, "The rabbit username")
	fs.DurationVar(&o.Wait, "graceful-timeout", o.Wait, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
}

func (o *Options) Config(stopCh <-chan struct{}) (*config.Config, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{"10.56.223.15:8689"},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "Di@clickHouse#456789",
		},
		ClientInfo: clickhouse.ClientInfo{
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "an-example-go-client", Version: "0.1"},
			},
		},

		Debugf: func(format string, v ...interface{}) {
			fmt.Printf(format, v)
		},
	})

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(context.Background()); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			klog.V(2).InfoS("Failed to connect clickhouse", "errCode", exception.Code, "errMsg", exception.Message)
		}
		return nil, err
	}

	manager := ck.NewManager(conn, stopCh)

	manager.Init()

	c := &config.Config{
		CkMgr: manager,
	}

	return c, nil
}
