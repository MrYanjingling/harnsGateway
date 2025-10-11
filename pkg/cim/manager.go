package cim

import (
	"context"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
	"strconv"
	"time"
)

var CimMap = map[string]string{
	"A1": "车载",
	"T1": "IT",
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
	// ts := yesterday.Format("20060102")

	m.dataDay(yesterday)

	// data := make([]*CimData, 0)
	// where := "DATE_TIMEKEY=" + ts
	//
	// // DWS_OUT_FACTORY_EMS_V
	// // DWS_ERP_CUX_WUHUMOD_SHIP_V_F
	// // ADS_OUT_FACTORY_EMS_M
	// // DWS_OUT_FACTORY_EMS_V
	// err := m.CimDb.Table("ADS_OUT_FACTORY_EMS_M").Select("FACTORY_NAME,DATE_TIMEKEY,OUT_TOTAL,PLAN_INPUT_QTY,REAL_INPUT_QTY").Where(where).Find(&data).Error
	// if err != nil {
	// 	klog.V(2).InfoS("Failed to get CIM data", "err", err)
	// }
	// klog.V(5).InfoS("Debug cim data", "length", len(data))
	// year := yesterday.Year()
	// month := int(yesterday.Month())
	// day := yesterday.Day()
	//
	// for _, datum := range data {
	// 	klog.V(5).InfoS("Debug cim data", "Factory", datum.Factory, "count", datum.Count, "time", datum.Time)
	// 	if n, ok := CimMap[datum.Factory]; ok {
	// 		klog.V(5).InfoS("Debug", "year", year, "month", month, "tree_name", n)
	// 		m.InsertIntoEms(n+"产量", datum.Count, year, month, day)
	//
	// 		m.InsertIntoEms(n+"计划投片量", datum.PlanInput, year, month, day)
	//
	// 		m.InsertIntoEms(n+"实际投片量", datum.RealInput, year, month, day)
	//
	// 	} else {
	// 		klog.V(2).InfoS("Failed to add cim production unit", "factory", datum.Factory)
	// 	}
	// }

}

func (m *Manager) Shutdown(ctx context.Context) error {

	return nil
}

func (m *Manager) InsertIntoEms(treeName string, data int, year int, month int, day int) {

	var ed EmsData
	if err := m.EmsDB.Table("ems_manual_production").Select("id,p_year,p_month,p_data,tree_name").Where("tree_name = ? and p_year = ? and p_month = ?", treeName, year, month).First(&ed).Error; err != nil {
		klog.V(5).InfoS("Failed to get ems data", "year", year, "month", month, "tree_name", treeName)
		return
	}

	if len(ed.Id) != 0 {
		pdm := make(map[string]interface{})
		err := json.Unmarshal([]byte(ed.PData), &pdm)
		if err != nil {
			klog.V(2).InfoS("Failed to unmarshal production data", "err", err)
			return
		} else {
			klog.V(3).InfoS("Success to unmarshal production data", "data", pdm)
		}
		// day := yesterday.Day()
		ds := fmt.Sprintf("d%d", day)
		pdm[ds] = fmt.Sprintf("%d", data)
		klog.V(3).InfoS("Production data", "day", ds, "count", data)

		if _, flag := pdm["total"].(float64); flag {
			nt := 0
			for i := 1; i <= day; i++ {
				if ct := pdm[fmt.Sprintf("d%d", i)]; ct != nil {
					klog.V(3).InfoS("ct data", "ct", ct)
					if atoi, err := strconv.Atoi(ct.(string)); err == nil {
						klog.V(2).InfoS("Success to get production data", "day", i, "data", atoi)
						nt += atoi
					} else {
						klog.V(2).InfoS("Failed to get production data", "day", i)
					}
				} else {
					klog.V(3).InfoS("ct data is nil", "ct", ct)
				}
			}
			// total = total + float64(data)
			pdm["total"] = nt
		} else {
			klog.V(2).InfoS("Production data 0")
			pdm["total"] = 0
		}

		if marshal, err := json.Marshal(pdm); err != nil {
			klog.V(2).InfoS("Failed to marshall production data", "err", err)
			return
		} else {
			upd := string(marshal)
			klog.V(5).InfoS("Success marshal Production data", "data", upd)
			if err := m.EmsDB.Table("ems_manual_production").Where("id = ?", ed.Id).Update("p_data", upd).Error; err != nil {
				klog.V(2).InfoS("Failed to insert into ems data", "err", err)
			}
		}
		// 计算count

	} else {
		klog.V(5).InfoS("Failed to get ems data", "year", year, "month", month, "tree_name", treeName)
	}

	return
}

func (m *Manager) Data(timeType string, t string) {
	if timeType == "d" {
		klog.V(3).InfoS("day")
		parse, err := time.Parse("20060102", t)
		if err != nil {
			return
		}
		klog.V(3).InfoS("sync data", "time", parse.Format("20060102"))
		m.dataDay(parse)
		return
	} else if timeType == "m" {
		klog.V(3).InfoS("month")
		parse, err := time.Parse("20060102", t+"01")
		if err != nil {
			return
		}
		for parse.Before(time.Now().AddDate(0, 0, -1).Add(20 * time.Hour)) {
			klog.V(3).InfoS("sync data", "time", parse.Format("20060102"))
			m.dataDay(parse)
			parse = parse.AddDate(0, 0, 1)
		}
		return
	}
}

func (m *Manager) dataDay(yesterday time.Time) {
	ts := yesterday.Format("20060102")

	data := make([]*CimData, 0)
	where := "DATE_TIMEKEY=" + ts

	err := m.CimDb.Table("ADS_OUT_FACTORY_EMS_M").Select("FACTORY_NAME,DATE_TIMEKEY,OUT_TOTAL,PLAN_INPUT_QTY,REAL_INPUT_QTY").Where(where).Find(&data).Error
	if err != nil {
		klog.V(2).InfoS("Failed to get CIM data", "err", err)
	}
	klog.V(5).InfoS("Debug cim data", "length", len(data))
	year := yesterday.Year()
	month := int(yesterday.Month())
	day := yesterday.Day()

	for _, datum := range data {
		klog.V(5).InfoS("Debug cim data", "Factory", datum.Factory, "count", datum.Count, "time", datum.Time)
		if n, ok := CimMap[datum.Factory]; ok {
			klog.V(5).InfoS("Debug", "year", year, "month", month, "tree_name", n)
			m.InsertIntoEms(n+"产量", datum.Count, year, month, day)

			m.InsertIntoEms(n+"计划投片量", datum.PlanInput, year, month, day)

			m.InsertIntoEms(n+"实际投片量", datum.RealInput, year, month, day)

		} else {
			klog.V(2).InfoS("Failed to add cim production unit", "factory", datum.Factory)
		}
	}
}
