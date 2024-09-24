package options

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	flag "github.com/spf13/pflag"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"harnsgateway/cmd/cim/config"
	"harnsgateway/pkg/cim"
	baseoptions "harnsgateway/pkg/generic/options"
	"k8s.io/klog/v2"
	"time"
)

const (
	_defaultCimDBUrl      = "10.121.47.11:9030"
	_defaultCimDBUsername = "ems"
	_defaultCimDBPassword = "Tm20@2024ems"
	_defaultEmsDBUrl      = "10.122.70.36:8406"
	_defaultEmsDBUsername = "root"
	_defaultEmsDBPassword = "i+aSroC6ak"
	_defaultWait          = time.Second * 15
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
	baseoptions.BaseOptions
}

func NewDefaultOptions() *Options {
	return &Options{
		CimDBUrl:      _defaultCimDBUrl,
		CimDBUsername: _defaultCimDBUsername,
		CimDBPassword: _defaultCimDBPassword,
		EmsDBUrl:      _defaultEmsDBUrl,
		EmsDBUsername: _defaultEmsDBUsername,
		EmsDBPassword: _defaultEmsDBPassword,
		Wait:          _defaultWait,
		BaseOptions:   baseoptions.NewDefaultBaseOptions(),
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
	gc := &gorm.Config{
		CreateBatchSize: 5000,
	}

	cimUrl := fmt.Sprintf(defaultCimDsn, o.CimDBUsername, o.CimDBPassword, o.CimDBUrl)

	cimDb, err := gorm.Open(mysql.Open(cimUrl), gc)
	if err != nil {
		klog.V(1).InfoS("Failed to connect CIM database", "err", err)
		return nil, err
	}

	emsUrl := fmt.Sprintf(defaultEmsDsn, o.EmsDBUsername, o.EmsDBPassword, o.EmsDBUrl)

	emsDb, err := gorm.Open(mysql.Open(emsUrl), gc)
	if err != nil {
		klog.V(1).InfoS("Failed to connect EMS database", "err", err)
		return nil, err
	}

	manager := cim.NewManager(cimDb, emsDb, stopCh)

	manager.Init()

	c := &config.Config{
		CimMgr: manager,
	}

	return c, nil
}
