package device

import (
	"context"
	"errors"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/go-redis/redis/v8"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"harnsgateway/pkg/apis"
	"harnsgateway/pkg/apis/response"
	"harnsgateway/pkg/gateway"
	"harnsgateway/pkg/generic"
	runtime2 "harnsgateway/pkg/protocol/modbus/runtime"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/runtime/constant"
	"harnsgateway/pkg/ts"
	v1 "harnsgateway/pkg/v1"
	"k8s.io/klog/v2"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Option func(*Manager)

type Manager struct {
	gatewayMeta      *gateway.GatewayMeta
	redisClient      *redis.Client
	mu               *sync.Mutex
	deviceManager    map[string]DeviceManager
	devices          *sync.Map
	heartBeatDevices *sync.Map
	tsManager        *ts.TsManager
	store            *generic.Store
	brokers          map[string]runtime.Broker
	brokerReturnCh   map[string]chan *runtime.ParseVariableResult
	stopCh           <-chan struct{}
	deviceStatusCh   chan string
	closers          []runtime.LabeledCloser
	placeholder      string
}

func NewManager(store *generic.Store, tsManager *ts.TsManager, redisClient *redis.Client, gatewayMeta *gateway.GatewayMeta, placeholder string, stop <-chan struct{}, opts ...Option) *Manager {
	m := &Manager{
		gatewayMeta:      gatewayMeta,
		redisClient:      redisClient,
		mu:               &sync.Mutex{},
		devices:          &sync.Map{},
		heartBeatDevices: &sync.Map{},
		deviceManager:    DeviceManagers,
		tsManager:        tsManager,
		brokers:          make(map[string]runtime.Broker, 0),
		brokerReturnCh:   make(map[string]chan *runtime.ParseVariableResult, 0),
		store:            store,
		stopCh:           stop,
		deviceStatusCh:   make(chan string, 0),
		placeholder:      placeholder,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Manager) Init() {
	devices, _ := m.store.LoadResource()
	for _, object := range devices {
		object.IndexDevice()
		obj, _ := runtime.AccessorDevice(object)
		m.devices.Store(obj.GetID(), obj)

		if err := m.readyCollect(obj); err != nil {
			if errors.Is(err, constant.ErrConnectDevice) {
				// 开启探测协程 15S一次
				m.heartBeatDevices.Store(obj.GetID(), obj)
			} else {
				klog.V(2).InfoS("Failed to start process collect device data", "deviceId", obj.GetID())
			}
		}
	}

	go m.heartBeatDetection()
	go m.listeningDeviceStatusCh()
}

func (m *Manager) CreateDevice(object v1.DeviceType) (runtime.Device, error) {
	device, err := m.deviceManager[object.GetDeviceType()].CreateDevice(object)
	if err != nil {
		klog.V(2).InfoS("Failed to create device", "error", err)
		return nil, err
	}

	created, err := m.store.Create(device)
	if err != nil {
		klog.V(2).InfoS("Failed to store device", "error", err)
		return nil, err
	}
	rd := created.(runtime.Device)
	m.devices.Store(rd.GetID(), rd)
	_, _ = runtime.AccessorDevice(created)

	if err = m.readyCollect(rd); err != nil {
		if errors.Is(err, constant.ErrConnectDevice) {
			// 开启探测协程 15S一次
			m.heartBeatDevices.Store(rd.GetID(), rd)
		} else {
			klog.V(2).InfoS("Failed to start process collect device data", "deviceId", rd.GetID())
			return nil, err
		}
	}

	return rd, nil
}

func (m *Manager) DeleteDevice(id string, version string) (runtime.Device, error) {
	device, err := m.GetDeviceById(id, false)
	if err != nil {
		return nil, err
	}

	if device.GetVersion() != version {
		return nil, apis.ErrMismatch
	}

	d, err := m.deviceManager[device.GetDeviceType()].DeleteDevice(device)
	if err != nil {
		klog.V(2).InfoS("Failed to delete device", "error", err)
		return nil, err
	}

	if _, err := m.store.Delete(d); err != nil {
		klog.V(2).InfoS("Failed to delete device", "deviceId", device.GetID())
	}

	klog.V(2).InfoS("Deleted device", "deviceId", device.GetID())

	go func() {
		if err := m.cancelCollect(device); err != nil {
			klog.V(2).InfoS("Failed to cancel collect process", "deviceId", device.GetID())
		}
	}()

	m.devices.Delete(device.GetID())
	return device, nil
}

// 模拟modbus服务
func (m *Manager) UpdateDeviceVariableByName(name string, file *multipart.FileHeader) error {
	devices, _ := m.ListDevices(&runtime.DeviceFilter{
		Name: name,
	}, true)

	device := devices[0]
	md := device.(*runtime2.ModBusDevice)

	open, _ := file.Open()
	excel, _ := excelize.OpenReader(open)

	sheetMap := excel.GetSheetMap()
	rows := excel.GetRows(sheetMap[1])

	if len(rows) > 1 {
		variables := make([]*runtime2.Variable, 0)
		variablesMap := make(map[string]*runtime2.Variable, 0)
		m.cancelCollect(md)
		for _, row := range rows[1:] {
			atoi, _ := strconv.Atoi(row[2])
			bit, _ := strconv.Atoi(row[3])
			functionCode, _ := strconv.Atoi(row[4])
			v := &runtime2.Variable{
				DataType:     constant.StringToDataType[row[1]],
				Name:         row[0],
				Address:      uint(atoi),
				Bits:         uint8(bit),
				FunctionCode: uint8(functionCode),
				Rate:         0,
				Amount:       0,
				AccessMode:   constant.AccessModeReadWrite,
			}
			variables = append(variables, v)
			variablesMap[row[0]] = v
		}

		md.Variables = variables
		md.VariablesMap = variablesMap
		m.store.Update(md)

		m.readyCollect(md)
	}

	return nil
}

func (m *Manager) UpdateDeviceById(id string, version string, newObj v1.DeviceType) (runtime.Device, error) {
	d, err := m.GetDeviceById(id, true)
	if err != nil {
		return nil, err
	}

	if version != d.GetVersion() {
		return nil, apis.ErrMismatch
	}

	copied := d.DeepCopyObject()
	cd := copied.(runtime.Device)

	if err = m.deviceManager[d.GetDeviceType()].UpdateValidation(newObj, cd); err != nil {
		return nil, err
	}

	device, err := m.deviceManager[d.GetDeviceType()].UpdateDevice(id, newObj, cd)
	if err != nil {
		klog.V(2).InfoS("Failed to update device", "error", err)
		return nil, err
	}

	updated, err := m.store.Update(device)
	if err != nil {
		klog.V(2).InfoS("Failed to update device", "error", err)
		return nil, err
	}
	rd := updated.(runtime.Device)
	m.devices.Store(rd.GetID(), updated)

	return updated, nil
}

func (m *Manager) ListDevices(filter *runtime.DeviceFilter, exploded bool) ([]runtime.Device, error) {
	rds := make([]runtime.Device, 0)
	predicates := runtime.ParseTypeFilter(filter)

	// descend
	byModTime := func(d1, d2 runtime.Device) bool { return d1.GetModTime().Before(d2.GetModTime()) }
	sorter := runtime.ByDevice(byModTime)

	m.devices.Range(func(key, value interface{}) bool {
		isMatch := true
		v := value.(runtime.Device)
		for _, p := range predicates {
			if !p(v) {
				isMatch = false
				break
			}
		}
		if isMatch {
			rds = sorter.Insert(rds, v)
		}
		return true
	})

	if !exploded {
		for i := range rds {
			rds[i] = m.foldDevice(rds[i])
		}
	}

	return rds, nil
}

func (m *Manager) GetDeviceById(id string, exploded bool) (runtime.Device, error) {
	d, isExist := m.devices.Load(id)
	if !isExist {
		return nil, os.ErrNotExist
	}
	device, _ := d.(runtime.Device)
	if !exploded {
		return m.foldDevice(device), nil
	}
	return device, nil
}

func (m *Manager) SwitchDeviceStatus(id string, status string) error {
	if _, err := m.GetDeviceById(id, true); err != nil {
		klog.V(2).InfoS("Failed to find device", "deviceId", id)
		return err
	}
	if _, ok := runtime.StringToDeviceStatusCh[status]; !ok {
		klog.V(2).InfoS("Unsupported device status", "status", status)
		return response.ErrDeviceOperatorUnSupported(status)
	}
	dsc := id + "-" + status
	m.deviceStatusCh <- dsc
	return nil
}

func (m *Manager) DeliverAction(id string, actions []map[string]interface{}) error {
	device, err := m.GetDeviceById(id, true)
	if err != nil {
		klog.V(2).InfoS("Failed to find device", "deviceId", id)
		return response.NewMultiError(response.ErrDeviceNotFound(id))
	}

	errs := &response.MultiError{}
	legalActions := make(map[string]interface{}, 0)
	for _, item := range actions {
		for k, v := range item {
			if _, exist := legalActions[k]; exist {
				errs.Add(response.ErrResourceExists(k))
				continue
			}
			if v, ok := device.GetVariable(k); !ok {
				errs.Add(response.ErrResourceNotFound(k))
				continue
			} else if v.GetVariableAccessMode() != constant.AccessModeReadWrite {
				errs.Add(response.ErrResourceNotFound(k))
				continue
			}
			legalActions[k] = v
		}
	}

	if errs.Len() > 0 {
		return errs
	}

	if len(legalActions) == 0 {
		return response.NewMultiError(response.ErrLegalActionNotFound)
	}

	if device.GetCollectStatus() == runtime.CollectStatusToString[runtime.Unconnected] {
		klog.V(2).InfoS("Failed to connect device", "deviceId", id)
		return response.NewMultiError(response.ErrDeviceNotConnect(id))
	}

	return m.brokers[id].DeliverAction(context.Background(), legalActions)
}

func (m Manager) cancelCollect(obj runtime.Device) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// switch status
	obj.SetCollectStatus(runtime.CollectStatusToString[runtime.Stopped])
	// delete heartBeat devices if exist
	if _, exist := m.heartBeatDevices.Load(obj.GetID()); exist {
		m.heartBeatDevices.Delete(obj.GetID())
	}
	if v, ok := m.brokers[obj.GetID()]; ok {
		v.Destroy(context.Background())
		delete(m.brokers, obj.GetID())
		delete(m.brokerReturnCh, obj.GetID())
	}
	return nil
}

func (m *Manager) readyCollect(obj runtime.Device) error {
	broker, results, err := generic.DeviceTypeBrokerMap[obj.GetDeviceType()](obj)
	if err != nil {
		switch {
		case errors.Is(err, constant.ErrConnectDevice):
			obj.SetCollectStatus(runtime.CollectStatusToString[runtime.Unconnected])
			return err
		case errors.Is(err, constant.ErrDeviceEmptyVariable):
			obj.SetCollectStatus(runtime.CollectStatusToString[runtime.EmptyVariable])
			return nil
		default:
			return err
		}
	}
	obj.SetCollectStatus(runtime.CollectStatusToString[runtime.Collecting])
	klog.V(2).InfoS("Succeed to collect data", "deviceId", obj.GetID())
	m.mu.Lock()
	defer m.mu.Unlock()
	m.brokers[obj.GetID()] = broker
	m.brokerReturnCh[obj.GetID()] = results

	// topic := obj.GetTopic()
	// if len(topic) == 0 {
	// 	topic = fmt.Sprintf("data/%s/v1/%s", m.gatewayMeta.ID, obj.GetID())
	// 	obj.SetTopic(topic)
	// }

	broker.Collect(context.Background())
	go func(deviceId string, ch chan *runtime.ParseVariableResult) {
		for {
			select {
			case _, ok := <-m.stopCh:
				if !ok {
					return
				}
			case pvr, ok := <-results:
				if ok {
					if v, ok := m.devices.Load(deviceId); ok {
						if len(pvr.Err) == 0 {
							if v.(runtime.Device).GetCollectStatus() != runtime.CollectStatusToString[runtime.Collecting] {
								v.(runtime.Device).SetCollectStatus(runtime.CollectStatusToString[runtime.Collecting])
							}
							pds := make([]runtime.PointData, 0, len(pvr.VariableSlice))
							for _, value := range pvr.VariableSlice {
								pd := runtime.PointData{
									DataPointId: value.GetVariableName(),
									Value:       value.GetValue(),
								}
								pds = append(pds, pd)
							}

							m.processData(pds)
							// publishData := runtime.PublishData{Payload: runtime.Payload{Data: []runtime.TimeSeriesData{{
							// 	Timestamp: time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
							// 	Values:    pds,
							// }}}}
							//
							// marshal, _ := json.Marshal(publishData)
							// token := m.mqttClient.Publish(topic, 1, false, marshal)
							// if token.WaitTimeout(mqttTimeout) && token.Error() == nil {
							// 	klog.V(5).InfoS("Succeed to publish MQTT", "topic", topic, "data", publishData)
							// } else {
							// 	klog.V(1).InfoS("Failed to publish MQTT", "topic", topic, "err", token.Error())
							// }
						} else {
							v.(runtime.Device).SetCollectStatus(runtime.CollectStatusToString[runtime.CollectingError])
						}
					} else {
						klog.V(2).InfoS("Failed to load device", "deviceId", deviceId)
					}
				} else {
					klog.V(2).InfoS("Stopped to collect data", "deviceId", deviceId)
					return
				}
			}
		}
	}(obj.GetID(), results)
	return nil
}

func (m *Manager) ShutdownDaemon(ctx context.Context) error {
	return nil
}

func (m *Manager) Shutdown(context context.Context) error {
	for _, c := range m.brokers {
		c.Destroy(context)
	}

	_ = m.redisClient.Close()
	var errs []string
	for i := len(m.closers); i > 0; i-- {
		lc := m.closers[i-1]
		if err := lc.Closer(context); err != nil {
			klog.V(2).InfoS("Failed to stopped Dependencies service", "service", lc.Label)
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("Failed to shutdown server: [%s]\n", strings.Join(errs, ","))
	}
	return nil
}

func (m *Manager) foldDevice(device runtime.Device) runtime.Device {
	return &runtime.DeviceMeta{
		ObjectMeta: runtime.ObjectMeta{
			Name:    device.GetName(),
			ID:      device.GetID(),
			Version: device.GetVersion(),
			ModTime: device.GetModTime(),
		},
		DeviceModel:   device.GetDeviceModel(),
		DeviceCode:    device.GetDeviceCode(),
		DeviceType:    device.GetDeviceType(),
		CollectStatus: device.GetCollectStatus(),
	}
}

func (m *Manager) heartBeatDetection() {
	tick := time.Tick(heartBeatTimeInterval)
	for {
		select {
		case _, ok := <-m.stopCh:
			if !ok {
				return
			}
		case <-tick:
			resumeDevices := make([]string, 0, 0)
			m.heartBeatDevices.Range(func(key, value any) bool {
				d := value.(runtime.Device)
				if err := m.readyCollect(d); err == nil {
					resumeDevices = append(resumeDevices, key.(string))
					return true
				}
				return false
			})
			if len(resumeDevices) > 0 {
				for _, deviceId := range resumeDevices {
					m.heartBeatDevices.Delete(deviceId)
				}
			}
		}
	}
}

func (m *Manager) listeningDeviceStatusCh() {
	for {
		select {
		case _, ok := <-m.stopCh:
			if !ok {
				return
			}
		case statusCh, ok := <-m.deviceStatusCh:
			if !ok {
				return
			}
			split := strings.Split(statusCh, "-")
			deviceId := split[0]
			status := split[1]
			d, exist := m.devices.Load(deviceId)
			if !exist {
				klog.V(2).InfoS("Failed to find device", "deviceId", deviceId)
			}
			m.switchDeviceStatus(d.(runtime.Device), status)
		}
	}
}

func (m *Manager) switchDeviceStatus(device runtime.Device, status string) {
	cs := device.GetCollectStatus()
	switch runtime.StringToCollectStatus[cs] {
	case runtime.Collecting:
		switch runtime.StringToDeviceStatusCh[status] {
		case runtime.Start:
			return
		case runtime.Restart:
			_ = m.cancelCollect(device)
			if err := m.readyCollect(device); err != nil {
				if errors.Is(err, constant.ErrConnectDevice) {
					m.heartBeatDevices.Store(device.GetID(), device)
				} else {
					klog.V(2).InfoS("Failed to start process collect device data", "deviceId", device.GetID())
				}
			}
			return
		case runtime.Stop:
			_ = m.cancelCollect(device)
			return
		}
	case runtime.CollectingError, runtime.Error:
		switch runtime.StringToDeviceStatusCh[status] {
		case runtime.Restart, runtime.Start:
			_ = m.cancelCollect(device)
			if err := m.readyCollect(device); err != nil {
				if errors.Is(err, constant.ErrConnectDevice) {
					m.heartBeatDevices.Store(device.GetID(), device)
				} else {
					klog.V(2).InfoS("Failed to start process collect device data", "deviceId", device.GetID())
				}
			}
			return
		case runtime.Stop:
			_ = m.cancelCollect(device)
			return
		}
	case runtime.EmptyVariable, runtime.Unconnected:
		switch runtime.StringToDeviceStatusCh[status] {
		case runtime.Restart, runtime.Start:
			_ = m.cancelCollect(device)
			if err := m.readyCollect(device); err != nil {
				if errors.Is(err, constant.ErrConnectDevice) {
					m.heartBeatDevices.Store(device.GetID(), device)
				} else {
					klog.V(2).InfoS("Failed to start process collect device data", "deviceId", device.GetID())
				}
			}
			return
		case runtime.Stop:
			_ = m.cancelCollect(device)
			return
		}
	case runtime.Stopped:
		switch runtime.StringToDeviceStatusCh[status] {
		case runtime.Restart, runtime.Start:
			if err := m.readyCollect(device); err != nil {
				if errors.Is(err, constant.ErrConnectDevice) {
					m.heartBeatDevices.Store(device.GetID(), device)
				} else {
					klog.V(2).InfoS("Failed to start process collect device data", "deviceId", device.GetID())
				}
			}
			return
		case runtime.Stop:
			return
		}
	}
}

func (m *Manager) processData(pds []runtime.PointData) {
	start := time.Now()

	thingTimeSeries := make(map[string]map[string]interface{}, 0)

	for _, pd := range pds {
		deviceProperty := strings.Split(pd.DataPointId, m.placeholder)
		if v, exist := thingTimeSeries[deviceProperty[0]]; exist {
			v[deviceProperty[1]] = pd.Value
		} else {
			pv := map[string]interface{}{deviceProperty[1]: pd.Value}
			thingTimeSeries[deviceProperty[0]] = pv
		}
	}

	for thingCode, kv := range thingTimeSeries {

		set := m.redisClient.HSet(context.Background(), thingCode, kv)
		if set.Err() != nil {
			klog.V(2).InfoS("Failed save data to redis", "thingCode", thingCode, "err", set.Err())
		}
	}
	end := time.Now()
	klog.V(3).InfoS("Insert into redis", "time", end.Sub(start).Seconds())
}

func (m *Manager) Daemon() {
	// now := time.Now()
	loc, _ := time.LoadLocation("Asia/Shanghai")

	t := time.Now().In(loc)
	t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, loc)

	// points := make([]*write.Point, 0)
	// thingTimeSeries := make(map[string]map[string]interface{}, 0)

	m.devices.Range(func(key, value any) bool {
		v := value.(runtime.Device)
		variables := v.GetVariables()

		points := make([]*write.Point, 0)
		thingTimeSeries := make(map[string]map[string]interface{}, 0)

		for _, variable := range variables {
			k := variable.GetVariableName()
			vv := variable.GetValue()
			deviceProperty := strings.Split(k, m.placeholder)
			if v, exist := thingTimeSeries[deviceProperty[0]]; exist {
				v[deviceProperty[1]] = vv
			} else {
				pv := map[string]interface{}{deviceProperty[1]: vv}
				thingTimeSeries[deviceProperty[0]] = pv
			}
		}

		go m.insertIntoInfluxdb(thingTimeSeries, points, t)

		return true
	})

	// m.insertIntoInfluxdb(thingTimeSeries, points, t)

}

func (m *Manager) insertIntoInfluxdb(thingTimeSeries map[string]map[string]interface{}, points []*write.Point, t time.Time) {
	start := time.Now()
	for thingCode, kv := range thingTimeSeries {
		point := write.NewPoint("device_data_electric_meter", map[string]string{"ti": thingCode}, kv, t)
		points = append(points, point)
	}

	m.tsManager.SaveOrUpdateTimeSeries(points)
	end := time.Now()
	klog.V(3).InfoS("Insert into redis", "time", end.Sub(start).Seconds())
}
