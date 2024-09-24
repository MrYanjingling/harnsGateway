package ts

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type TsManager struct {
	service http.Service
	client  influxdb2.Client
}

func NewTsManager(influxdbUrl, influxdbToken string) *TsManager {
	service := http.NewService(influxdbUrl, influxdbToken, http.DefaultOptions())
	client := influxdb2.NewClient(influxdbUrl, "c6jGYUCinwzeTWdeUh32")

	s := &TsManager{
		service: service,
		client:  client,
	}
	return s
}

func (s *TsManager) SaveOrUpdateTimeSeries(points []*write.Point) error {
	// end := time.Now().AddDate(0, 0, 5)
	// start := end.AddDate(0, -5, 1)
	//
	// deleteAPI := s.client.DeleteAPI()
	// err := deleteAPI.DeleteWithName(context.Background(), "main", "data-raw", start, end, "_measurement=device_data_electric_meter")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	if len(points) == 0 {
		return nil
	}
	writeApi := api.NewWriteAPI("main", "data-raw", s.service, write.DefaultOptions().SetBatchSize(uint(len(points))).SetUseGZip(true))
	for _, p := range points {
		writeApi.WritePoint(p)
	}
	writeApi.Flush()
	return nil
}
