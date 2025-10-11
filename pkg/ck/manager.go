package ck

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"k8s.io/klog/v2"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type Manager struct {
	Conn     driver.Conn
	things   int
	property int
	StopCh   <-chan struct{}
}

func NewManager(conn driver.Conn, stopCh <-chan struct{}) *Manager {
	return &Manager{
		Conn:   conn,
		StopCh: stopCh,
	}
}

func (m *Manager) Init() {
	m.things = 100
	m.property = 70
}

func (m *Manager) Polling() {
	t := time.Now()
	sw := &sync.WaitGroup{}
	sw.Add(2)
	go m.insertIntoTags(t, sw)
	go m.insertIOT(t, sw)
	// go m.insertIOT2(t, sw)

	sw.Wait()
}

func (m *Manager) Shutdown(ctx context.Context) error {
	m.Conn.Close()
	return nil
}

func (m *Manager) insertIntoTags(t time.Time, sw *sync.WaitGroup) {
	start := time.Now().UnixMilli()

	insertSql := "INSERT INTO tags (dataTag, value, dataTime, insertTime)"
	batch, err := m.Conn.PrepareBatch(context.Background(), insertSql)
	if err != nil {
		klog.V(2).InfoS("Failed to get batch")
	}
	for i := 1; i <= m.things; i++ {
		for j := i; j <= m.property; j++ {
			tagName := "thing" + strconv.Itoa(i) + "_a" + strconv.Itoa(j)
			value := rand.Float64()
			// 0.418073193481601
			float := strconv.FormatFloat(value, 'f', 15, 64)
			if err := batch.Append(tagName, float, t, t); err != nil {
				klog.V(2).InfoS("Failed to batch append", "error", err)
			}
		}
	}
	if err := batch.Send(); err != nil {
		klog.V(2).InfoS("Failed to batch send", "error", err)
	}
	end := time.Now().UnixMilli()
	klog.V(2).InfoS("Success insert into tags time", "cost", end-start)
	sw.Done()
}

func (m *Manager) insertIOT(t time.Time, sw *sync.WaitGroup) {
	start := time.Now().UnixMilli()
	insertSql := "INSERT INTO things ("
	for i := 1; i <= m.property; i++ {
		insertSql += "a" + strconv.Itoa(i) + ", "
	}
	insertSql += "thingId, time)"

	batch, err := m.Conn.PrepareBatch(context.Background(), insertSql)
	if err != nil {
		klog.V(2).InfoS("Failed to get batch")
	}
	for i := 1; i <= m.things; i++ {
		values := make([]interface{}, 0, m.property+2)
		for j := 1; j <= m.property; j++ {
			value := rand.Float64()
			values = append(values, value)
		}
		values = append(values, "thing"+strconv.Itoa(i))
		values = append(values, t)

		if err := batch.Append(values...); err != nil {
			klog.V(2).InfoS("Failed to batch append", "error", err)
		}

	}
	if err := batch.Send(); err != nil {
		klog.V(2).InfoS("Failed to batch send", "error", err)
	}
	end := time.Now().UnixMilli()
	klog.V(2).InfoS("Success insert into IOT time", "cost", end-start)
	sw.Done()
}

func (m *Manager) insertIOT2(t time.Time, sw *sync.WaitGroup) {
	start := time.Now().UnixMilli()
	insertSql := "INSERT INTO things2 ("
	for i := 1; i <= m.property; i++ {
		insertSql += "a" + strconv.Itoa(i) + ", "
	}
	insertSql += "thingId, time)"

	batch, err := m.Conn.PrepareBatch(context.Background(), insertSql)
	if err != nil {
		klog.V(2).InfoS("Failed to get batch")
	}
	for i := 1; i <= m.things; i++ {
		values := make([]interface{}, 0, m.property+2)
		for j := 1; j <= m.property; j++ {
			value := rand.Float64()
			values = append(values, value)
		}
		values = append(values, "thing"+strconv.Itoa(i))
		values = append(values, t)

		if err := batch.Append(values...); err != nil {
			klog.V(2).InfoS("Failed to batch append", "error", err)
		}

	}
	if err := batch.Send(); err != nil {
		klog.V(2).InfoS("Failed to batch send", "error", err)
	}
	end := time.Now().UnixMilli()
	klog.V(2).InfoS("Success insert into IOT2 time", "cost", end-start)
	sw.Done()
}
