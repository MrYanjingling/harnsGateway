package config

import (
	"database/sql"

	"github.com/elastic/go-elasticsearch/v7"

	"harnsgateway/pkg/data"
)

type Config struct {
	DB          *sql.DB
	ES          *elasticsearch.Client
	DataManager *data.Manager
	ScadaURL    string
	EscadaURL   string
	CertFile    string
	KeyFile     string
}
