package data

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/boltdb/bolt"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
	"mime/multipart"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Manager struct {
	Store  *bolt.DB
	Db     *gorm.DB
	Conn   *amqp.Connection
	Mc     *MqConfig
	StopCh <-chan struct{}
	Tags   *sync.Map
}

func NewManager(Db *gorm.DB, Conn *amqp.Connection, store *bolt.DB, Mc *MqConfig, stopCh <-chan struct{}) *Manager {
	return &Manager{
		Store:  store,
		Db:     Db,
		Conn:   Conn,
		Mc:     Mc,
		StopCh: stopCh,
		Tags:   &sync.Map{},
	}
}

func (m *Manager) Init() {
	if err := m.Store.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists([]byte("tag")); err != nil {
			return err
		}
		return nil
	}); err != nil {
		klog.V(2).InfoS("Failed to create database tag", "err", err)
	}

	tag := &Tag{}
	if err := m.Store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("tag"))
		get := bucket.Get([]byte("tagNames"))
		if get == nil {
			return nil
		}
		if err := json.NewDecoder(bytes.NewReader(get)).Decode(tag); err != nil {
			return err
		}
		return nil
	}); err != nil {
		klog.V(2).InfoS("Failed to view tag names", "err", err)
	}

	klog.V(5).InfoS("Success init tagNames", "tagNames", tag.TagNames)
	m.Tags.Store("tag", tag)
}

func (m *Manager) ImportTagNames(file *multipart.FileHeader) error {
	tagNames := make([]string, 0)

	open, _ := file.Open()
	excel, _ := excelize.OpenReader(open)

	sheetMap := excel.GetSheetMap()
	rows := excel.GetRows(sheetMap[1])

	if len(rows) > 1 {
		for _, row := range rows[1:] {
			tn := row[0]
			if !strings.Contains(tn, ".VAL_Actl") {
				tn = fmt.Sprintf("%s%s", tn, ".VAL_Actl")

			}
			tagNames = append(tagNames, tn)
			klog.V(5).InfoS("Success import tagNames", "tagNames", tagNames)
		}
	}
	t := &Tag{TagNames: tagNames}

	m.Tags.Store("tag", t)

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(t); err != nil {
		klog.V(2).InfoS("Failed to encode tag", "err", err)
	}

	err := m.Store.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("tag"))

		if err := bucket.Put([]byte("tagNames"), buf.Bytes()); err != nil {
			klog.V(2).InfoS("Failed to save tagNames", "err", err)
			return err
		}
		return nil
	})

	return err
}

func (m *Manager) ListTagNames() []string {
	load, ok := m.Tags.Load("tag")
	if !ok {
		return []string{}
	}
	t := load.(*Tag)
	if len(t.TagNames) == 0 {
		return []string{}
	}
	return t.TagNames
}

func (m *Manager) Polling() {
	spcs := make([]*Spc, 0)

	load, ok := m.Tags.Load("tag")
	if !ok {
		return
	}
	t := load.(*Tag)
	if len(t.TagNames) == 0 {
		return
	}

	tagNames := t.TagNames

	data := make([]*TimeSeries, 0)
	err := m.Db.Table("dbo.Live").Select("TagName as tagName, DateTime as time,Value as value").Where("tagName IN ?", tagNames).Find(&data).Error
	if err != nil {
		klog.V(2).InfoS("Failed to connect FMCS database", "err", err)
	}

	items := make([]*Item, 0)
	for index, dataTime := range data {
		tn := dataTime.TagName
		if suffix, found := strings.CutSuffix(tn, ".VAL_Actl"); found {
			tn = suffix
		}

		is := strconv.Itoa(index)
		item := NewItem(tn, is, dataTime.Value)
		items = append(items, item)
	}

	ts := time.Now().Format("2006-01-02 15:04:05.000")
	spc := NewSpc(items, ts)
	spcs = append(spcs, spc)

	// push
	channel, err := m.Conn.Channel()

	defer channel.Close()
	if err != nil {
		klog.V(2).InfoS("Failed to get rabbit channel", "err", err)
	}
	marshal, err := json.Marshal(spcs)
	if err != nil {
		klog.V(2).InfoS("Failed to marshal spcs data", "err", err)
	} else {
		klog.V(5).InfoS("Success to marshal spcs data", "data", spcs)
	}

	q, err := channel.QueueDeclare(
		m.Mc.Queue, // 队列名称
		true,       // 是否持久化
		false,      // 是否自动删除
		false,      // 是否为排他性队列
		false,      // 是否等待服务器返回响应
		nil,        // 额外参数
	)
	if err != nil {
		klog.V(2).InfoS("Failed to declare MQ Queue", "err", err)
	}

	err = channel.Publish(m.Mc.Exchange, q.Name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        marshal,
	})
	if err != nil {
		klog.V(2).InfoS("Failed to publish data to rabbit", "err", err)
	}

}

func (m *Manager) Shutdown(ctx context.Context) error {
	if err := m.Store.Close(); err != nil {
		return err
	}
	if err := m.Conn.Close(); err != nil {
		return err
	}
	return nil
}
