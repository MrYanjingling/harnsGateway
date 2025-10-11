package ts

import (
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"k8s.io/klog/v2"
)

type TsManager struct {
	service http.Service
	client  influxdb2.Client
}

func NewTsManager(influxdbUrl, influxdbToken string) *TsManager {
	service := http.NewService(influxdbUrl, influxdbToken, http.DefaultOptions())
	client := influxdb2.NewClientWithOptions(influxdbUrl, "c6jGYUCinwzeTWdeUh32", influxdb2.DefaultOptions().SetLogLevel(3))

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
	defer writeApi.Close()
	errCh := writeApi.Errors()
	go func() {
		for err := range errCh {
			klog.V(2).InfoS("Failed to write data into influxdb", "err", err.Error())
		}
	}()
	for _, p := range points {
		writeApi.WritePoint(p)
	}
	return nil
}
