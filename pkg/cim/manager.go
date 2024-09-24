package cim

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
	"time"
)

var CimMap = map[string]string{
	"A1": "车载产量",
	"T1": "IT产量",
}

type Manager struct {
	EmsDB  *gorm.DB
	CimDb  *gorm.DB
	StopCh <-chan struct{}
}

func NewManager(CimDb *gorm.DB, EmsDb *gorm.DB, stopCh <-chan struct{}) *Manager {
	return &Manager{
		CimDb:  CimDb,
		EmsDB:  EmsDb,
		StopCh: stopCh,
	}
}

func (m *Manager) Init() {

}

func (m *Manager) Polling() {

	yesterday := time.Now().AddDate(0, 0, -1)
	ts := yesterday.Format("20060102")

	data := make([]*CimData, 0)
	where := "DATE_TIMEKEY=" + ts

	err := m.CimDb.Table("DWS_OUT_FACTORY_EMS_V").Select("FACTORY,DATE_TIMEKEY,OUT_TOTAL").Where(where).Find(&data).Error
	if err != nil {
		klog.V(2).InfoS("Failed to get CIM data", "err", err)
	}
	klog.V(5).InfoS("Debug cim data", "length", len(data))
	year := yesterday.Year()
	month := int(yesterday.Month())

	for _, datum := range data {
		klog.V(5).InfoS("Debug cim data", "Factory", datum.Factory, "count", datum.Count, "time", datum.Time)
		if n, ok := CimMap[datum.Factory]; ok {
			klog.V(5).InfoS("Debug", "year", year, "month", month, "tree_name", n)

			var ed EmsData
			if err := m.EmsDB.Table("ems_manual_production").Select("id,p_year,p_month,p_data,tree_name").Where("tree_name = ? and p_year = ? and p_month = ?", n, year, month).First(&ed).Error; err != nil {
				klog.V(5).InfoS("Failed to get ems data", "year", year, "month", month, "tree_name", n)
				continue
			}

			if len(ed.Id) != 0 {
				pdm := make(map[string]string)
				err := json.Unmarshal([]byte(ed.PData), &pdm)
				if err != nil {
					klog.V(2).InfoS("Failed to unmarshal production data", "err", err)
					continue
				}
				day := yesterday.Day()
				ds := fmt.Sprintf("d%d", day)
				pdm[ds] = fmt.Sprintf("%d", datum.Count)
				klog.V(3).InfoS("Production data", "day", ds, "count", datum.Count)
				if marshal, err := json.Marshal(pdm); err != nil {
					klog.V(2).InfoS("Failed to marshall production data", "err", err)
					continue
				} else {
					upd := string(marshal)
					klog.V(5).InfoS("Success marshal Production data", "data", upd)
					if err := m.EmsDB.Table("ems_manual_production").Where("id = ?", ed.Id).Update("p_data", upd).Error; err != nil {
						klog.V(2).InfoS("Failed to insert into ems data", "err", err)
					}
				}

			} else {
				klog.V(5).InfoS("Failed to get ems data", "year", year, "month", month, "tree_name", n)
			}

		} else {
			klog.V(2).InfoS("Failed to add cim production unit", "factory", datum.Factory)
		}
	}

}

func (m *Manager) Shutdown(ctx context.Context) error {

	return nil
}
