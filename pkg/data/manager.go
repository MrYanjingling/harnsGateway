package data

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
	"k8s.io/klog/v2"
)

type Option func(*Manager)

type Manager struct {
	db           *sql.DB
	es           *elasticsearch.Client
	scadaURL     string
	escadaURL    string
	scadaClient  *http.Client
	escadaClient *http.Client
}

// WithScadaURL 设置第三方SCADA接口地址并初始化忽略证书校验的HTTPS客户端。
func WithScadaURL(url string) Option {
	return func(m *Manager) {
		m.scadaURL = url
		m.scadaClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}
}

func WithEScadaURL(url string) Option {
	return func(m *Manager) {
		m.escadaURL = url
		m.escadaClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	}
}

func NewDataManager(db *sql.DB, es *elasticsearch.Client, opts ...Option) *Manager {
	m := &Manager{
		db: db,
		es: es,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Manager) Init() {
	klog.InfoS("DataManager initialized")
}

func (m *Manager) DB() *sql.DB {
	return m.db
}

func (m *Manager) ES() *elasticsearch.Client {
	return m.es
}

// QueryMetrics 循环遍历每个metricsCode，先通过metricsCode从ems_statistics_metrics获取id，
// 再从ES metrics_data_YYYYMM 索引中按时间范围对s_data做sum聚合，每个code返回一个汇总值。
func (m *Manager) QueryMetrics(params *MetricsQueryParams) (map[string]*MetricsResult, error) {
	klog.V(4).InfoS("QueryMetrics: start", "metricsCodes", params.MetricsCodes, "startTime", params.StartTime, "endTime", params.EndTime, "granularity", params.Granularity)

	indices := buildIndexNames("metrics_data", params.StartTime, params.EndTime, params.StartTime.Location())

	results := make(map[string]*MetricsResult)

	for _, code := range params.MetricsCodes {
		// Step 1: 通过metricsCode查询ems_statistics_metrics获取id
		meta, err := m.getMetricsMetadata(code)
		if err != nil {
			klog.ErrorS(err, "Failed to get metrics metadata", "code", code)
			results[code] = &MetricsResult{
				MetricsCode: code,
				StartTime:   params.StartTime.Format(TimeLayout),
				EndTime:     params.EndTime.Format(TimeLayout),
			}
			continue
		}

		// Step 2: 从ES查询指标数据，对s_data做sum聚合
		queryBody := buildESQuery(meta.ID, params.StartTime, params.EndTime)
		klog.V(4).InfoS("QueryMetrics: ES query", "metricsCode", code, "metricsId", meta.ID, "indices", indices, "query", queryBody)
		esRaw, err := m.searchES(indices, queryBody)
		if err != nil {
			klog.ErrorS(err, "Failed to query ES", "metricsId", meta.ID, "indices", indices)
			results[code] = &MetricsResult{
				MetricsCode: code,
				StartTime:   params.StartTime.Format(TimeLayout),
				EndTime:     params.EndTime.Format(TimeLayout),
			}
			continue
		}
		klog.V(4).InfoS("QueryMetrics: ES response", "metricsCode", code, "raw", string(esRaw))

		value := parseESSumAggregation(esRaw)
		results[code] = &MetricsResult{
			MetricsCode: code,
			StartTime:   params.StartTime.Format(TimeLayout),
			EndTime:     params.EndTime.Format(TimeLayout),
			Value:       value,
		}
	}

	return results, nil
}

// getMetricsMetadata 通过metricsCode查询ems_statistics_metrics获取元数据。
func (m *Manager) getMetricsMetadata(code string) (*MetricsMetadata, error) {
	meta := &MetricsMetadata{}
	err := m.db.QueryRow(
		`SELECT id, metrics_code, metrics_name, measure_type, calculate_type, time_granularity
		 FROM ems_statistics_metrics
		 WHERE metrics_code = ?`, code,
	).Scan(&meta.ID, &meta.MetricsCode, &meta.MetricsName,
		&meta.MeasureType, &meta.CalculateType, &meta.TimeGranularity)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

// buildIndexNames 根据startTime和endTime生成ES索引名称列表，支持跨月。
func buildIndexNames(prefix string, startTime, endTime time.Time, loc *time.Location) []string {
	var indices []string
	current := time.Date(startTime.Year(), startTime.Month(), 1, 0, 0, 0, 0, loc)
	end := time.Date(endTime.Year(), endTime.Month(), 1, 0, 0, 0, 0, loc)
	for !current.After(end) {
		indices = append(indices, fmt.Sprintf("%s_%s", prefix, current.Format("200601")))
		current = current.AddDate(0, 1, 0)
	}
	return indices
}

// esQueryTemplate ES查询模板：range(s_time) + term(metrics_id) + sum(s_data)聚合。
const esQueryTemplate = `{
    "size": 0,
    "query": {
        "bool": {
            "must": [
                {
                    "range": {
                        "s_time": {
                            "gte": "%s",
                            "lt": "%s"
                        }
                    }
                },
                {
                    "term": {
                        "metrics_id": {
                            "value": "%s"
                        }
                    }
                }
            ]
        }
    },
    "aggs": {
        "total_value": {
            "sum": {
                "field": "s_data"
            }
        }
    }
}`

func buildESQuery(metricsID string, startTime, endTime time.Time) string {
	return fmt.Sprintf(esQueryTemplate,
		startTime.Format(TimeLayout),
		endTime.Format(TimeLayout),
		metricsID,
	)
}

// searchES 执行ES搜索请求，返回原始响应body。
func (m *Manager) searchES(indices []string, queryBody string) ([]byte, error) {
	index := strings.Join(indices, ",")
	req := esapi.SearchRequest{
		Index: []string{index},
		Body:  strings.NewReader(queryBody),
	}

	res, err := req.Do(context.Background(), m.es)
	if err != nil {
		return nil, fmt.Errorf("ES search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("ES search error: %s, body: %s", res.Status(), string(body))
	}

	return io.ReadAll(res.Body)
}

// esSumAggResponse ES sum聚合响应结构体。
type esSumAggResponse struct {
	Aggregations struct {
		TotalValue struct {
			Value float64 `json:"value"`
		} `json:"total_value"`
	} `json:"aggregations"`
}

// parseESSumAggregation 解析ES单sum聚合响应，返回聚合值。
func parseESSumAggregation(raw []byte) float64 {
	var agg esSumAggResponse
	if err := json.Unmarshal(raw, &agg); err != nil {
		klog.ErrorS(err, "Failed to unmarshal ES sum aggregation response")
		return 0
	}
	return agg.Aggregations.TotalValue.Value
}

// QueryKpi 循环遍历每个id，先从ems_statistics_metrics获取元数据，
// 再从ES metrics_data_YYYYMM索引中按时间范围对s_data做sum聚合，每个id返回一个汇总值。
func (m *Manager) QueryKpi(params *KpiQueryParams) (map[string]*KpiResult, error) {
	klog.V(4).InfoS("QueryKpi: start", "ids", params.Ids, "startTime", params.StartTime, "endTime", params.EndTime, "granularity", params.Granularity)

	indices := buildIndexNames("kpi_data", params.StartTime, params.EndTime, params.StartTime.Location())

	results := make(map[string]*KpiResult)

	for _, id := range params.Ids {
		// Step 2: 构建KPI ES查询
		queryBody := buildKpiESQuery(id, params.StartTime, params.EndTime)
		klog.V(4).InfoS("QueryKpi: ES query", "kpiId", id, "indices", indices, "query", queryBody)

		// Step 3: 执行ES查询
		esRaw, err := m.searchES(indices, queryBody)
		if err != nil {
			klog.ErrorS(err, "Failed to query ES", "kpiId", id, "indices", indices)
			results[id] = &KpiResult{
				Id:        id,
				StartTime: params.StartTime.Format(TimeLayout),
				EndTime:   params.EndTime.Format(TimeLayout),
			}
			continue
		}
		klog.V(4).InfoS("QueryKpi: ES response", "kpiId", id, "raw", string(esRaw))

		value := parseESDataAggregation(esRaw)
		results[id] = &KpiResult{
			Id:        id,
			StartTime: params.StartTime.Format(TimeLayout),
			EndTime:   params.EndTime.Format(TimeLayout),
			Value:     value,
		}
	}

	return results, nil
}

// kpiESQueryTemplate KPI查询模板：range(s_time) + term(metrics_id) + sum(s_data)聚合。
const kpiESQueryTemplate = `{
    "size": 0,
    "query": {
        "bool": {
            "must": [
                {
                    "range": {
                        "s_time": {
                            "gte": "%s",
                            "lt": "%s"
                        }
                    }
                },
                {
                    "term": {
                        "kpi_id": "%s"
                    }
                }
            ]
        }
    },
    "aggs": {
        "data": {
            "sum": {
                "field": "s_data"
            }
        }
    }
}`

func buildKpiESQuery(metricsID string, startTime, endTime time.Time) string {
	return fmt.Sprintf(kpiESQueryTemplate,
		startTime.Format(TimeLayout),
		endTime.Format(TimeLayout),
		metricsID,
	)
}

// esDataAggResponse ES "data" 聚合响应结构体。
type esDataAggResponse struct {
	Aggregations struct {
		Data struct {
			Value float64 `json:"value"`
		} `json:"data"`
	} `json:"aggregations"`
}

// parseESDataAggregation 解析ES data聚合响应，返回聚合值。
func parseESDataAggregation(raw []byte) float64 {
	var agg esDataAggResponse
	if err := json.Unmarshal(raw, &agg); err != nil {
		klog.ErrorS(err, "Failed to unmarshal ES data aggregation response")
		return 0
	}
	return agg.Aggregations.Data.Value
}

func (m *Manager) Shutdown() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// ListKpis 查询所有KPI列表。
func (m *Manager) ListKpis() ([]*KpiInfo, error) {
	rows, err := m.db.Query(
		`SELECT id, tenant_id, revision, created_by, created_time,
		        updated_by, updated_time, name, formula, formula_data_name,
		        formula_ori, time_granularity, target_data, single_quota,
		        default_tag, can_delete, unit, target_data_type, status
		 FROM ems_kpi
		 ORDER BY created_time DESC`)
	if err != nil {
		return nil, fmt.Errorf("query ems_kpi failed: %w", err)
	}
	defer rows.Close()

	var result []*KpiInfo
	for rows.Next() {
		var (
			k                                                        KpiInfo
			tenantId, createdBy, updatedBy, formula, formulaDataName sql.NullString
			targetData, unit                                         sql.NullString
			revision                                                 sql.NullInt64
			createdTime, updatedTime                                 sql.NullTime
			singleQuota, defaultTag, canDelete                       []byte
		)

		err := rows.Scan(
			&k.Id, &tenantId, &revision, &createdBy, &createdTime,
			&updatedBy, &updatedTime, &k.Name, &formula, &formulaDataName,
			&k.FormulaOri, &k.TimeGranularity, &targetData, &singleQuota,
			&defaultTag, &canDelete, &unit, &k.TargetDataType, &k.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("scan ems_kpi row failed: %w", err)
		}

		if tenantId.Valid {
			k.TenantId = &tenantId.String
		}
		if revision.Valid {
			v := int(revision.Int64)
			k.Revision = &v
		}
		if createdBy.Valid {
			k.CreatedBy = &createdBy.String
		}
		if createdTime.Valid {
			k.CreatedTime = &createdTime.Time
		}
		if updatedBy.Valid {
			k.UpdatedBy = &updatedBy.String
		}
		if updatedTime.Valid {
			k.UpdatedTime = &updatedTime.Time
		}
		if formula.Valid {
			k.Formula = &formula.String
		}
		if formulaDataName.Valid {
			k.FormulaDataName = &formulaDataName.String
		}
		if targetData.Valid {
			k.TargetData = &targetData.String
		}
		if unit.Valid {
			k.Unit = &unit.String
		}
		k.SingleQuota = len(singleQuota) > 0 && singleQuota[0] == 1
		k.DefaultTag = len(defaultTag) > 0 && defaultTag[0] == 1
		k.CanDelete = len(canDelete) > 0 && canDelete[0] == 1

		result = append(result, &k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ems_kpi rows failed: %w", err)
	}
	return result, nil
}

// ListTags 查询所有测点列表。
func (m *Manager) ListTags() ([]*TagInfo, error) {
	rows, err := m.db.Query(
		`SELECT id, tenant_id, revision, created_by, created_time,
		        updated_by, updated_time, tag_id, tag_name, tag_desc,
		        last_syn_time, tag_type, tag_attribute, enable,
		        scada_id, operator, operate_value
		 FROM ems_tag
		 ORDER BY created_time DESC`)
	if err != nil {
		return nil, fmt.Errorf("query ems_tag failed: %w", err)
	}
	defer rows.Close()

	var result []*TagInfo
	for rows.Next() {
		var (
			t                                                       TagInfo
			tenantId, createdBy, updatedBy, tagId, tagName, tagDesc sql.NullString
			tagType, tagAttribute, operator, operateValue           sql.NullString
			revision                                                sql.NullInt64
			createdTime, updatedTime, lastSynTime                   sql.NullTime
			enable                                                  sql.NullInt64
		)

		err := rows.Scan(
			&t.Id, &tenantId, &revision, &createdBy, &createdTime,
			&updatedBy, &updatedTime, &tagId, &tagName, &tagDesc,
			&lastSynTime, &tagType, &tagAttribute, &enable,
			&t.ScadaId, &operator, &operateValue,
		)
		if err != nil {
			return nil, fmt.Errorf("scan ems_tag row failed: %w", err)
		}

		if tenantId.Valid {
			t.TenantId = &tenantId.String
		}
		if revision.Valid {
			v := int(revision.Int64)
			t.Revision = &v
		}
		if createdBy.Valid {
			t.CreatedBy = &createdBy.String
		}
		if createdTime.Valid {
			t.CreatedTime = &createdTime.Time
		}
		if updatedBy.Valid {
			t.UpdatedBy = &updatedBy.String
		}
		if updatedTime.Valid {
			t.UpdatedTime = &updatedTime.Time
		}
		if tagId.Valid {
			t.TagId = &tagId.String
		}
		if tagName.Valid {
			t.TagName = &tagName.String
		}
		if tagDesc.Valid {
			t.TagDesc = &tagDesc.String
		}
		if lastSynTime.Valid {
			t.LastSynTime = &lastSynTime.Time
		}
		if tagType.Valid {
			t.TagType = &tagType.String
		}
		if tagAttribute.Valid {
			t.TagAttribute = &tagAttribute.String
		}
		if enable.Valid {
			v := int(enable.Int64)
			t.Enable = &v
		}
		if operator.Valid {
			t.Operator = &operator.String
		}
		if operateValue.Valid {
			t.OperateValue = &operateValue.String
		}

		result = append(result, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ems_tag rows failed: %w", err)
	}
	return result, nil
}

// ListMetrics 查询所有统计指标列表。
func (m *Manager) ListMetrics() ([]*MetricsInfo, error) {
	rows, err := m.db.Query(
		`SELECT id, tenant_id, revision, created_by, created_time,
		        updated_by, updated_time, data_point, metrics_code, metrics_name,
		        measure_type, enable, metrics_type, calculate_type, note
		 FROM ems_statistics_metrics
		 ORDER BY created_time DESC`)
	if err != nil {
		return nil, fmt.Errorf("query ems_statistics_metrics failed: %w", err)
	}
	defer rows.Close()

	var result []*MetricsInfo
	for rows.Next() {
		var (
			m                                                          MetricsInfo
			tenantId, createdBy, updatedBy, dataPoint, metricsCode     sql.NullString
			metricsName, measureType, metricsType, calculateType, note sql.NullString
			revision                                                   sql.NullInt64
			createdTime, updatedTime                                   sql.NullTime
			enable                                                     sql.NullBool
		)

		err := rows.Scan(
			&m.Id, &tenantId, &revision, &createdBy, &createdTime,
			&updatedBy, &updatedTime, &dataPoint, &metricsCode, &metricsName,
			&measureType, &enable, &metricsType, &calculateType, &note,
		)
		if err != nil {
			return nil, fmt.Errorf("scan ems_statistics_metrics row failed: %w", err)
		}

		if tenantId.Valid {
			m.TenantId = &tenantId.String
		}
		if revision.Valid {
			v := int(revision.Int64)
			m.Revision = &v
		}
		if createdBy.Valid {
			m.CreatedBy = &createdBy.String
		}
		if createdTime.Valid {
			m.CreatedTime = &createdTime.Time
		}
		if updatedBy.Valid {
			m.UpdatedBy = &updatedBy.String
		}
		if updatedTime.Valid {
			m.UpdatedTime = &updatedTime.Time
		}
		if dataPoint.Valid {
			m.DataPoint = &dataPoint.String
		}
		if metricsCode.Valid {
			m.MetricsCode = &metricsCode.String
		}
		if metricsName.Valid {
			m.MetricsName = &metricsName.String
		}
		if measureType.Valid {
			m.MeasureType = &measureType.String
		}
		if enable.Valid {
			m.Enable = &enable.Bool
		}
		if metricsType.Valid {
			m.MetricsType = &metricsType.String
		}
		if calculateType.Valid {
			m.CalculateType = &calculateType.String
		}
		if note.Valid {
			m.Note = &note.String
		}

		result = append(result, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ems_statistics_metrics rows failed: %w", err)
	}
	return result, nil
}

// QueryTagTimeSeries 调用第三方SCADA接口查询测点时序数据。
func (m *Manager) QueryTagTimeSeries(req *TagTimeSeriesRequest) (*TagTimeSeriesResponse, error) {
	reqBody := map[string]interface{}{
		"select": req.Select,
		"latest": req.Latest,
		"mode":   "raw",
	}
	if req.Start != "" {
		reqBody["start"] = req.Start
	}
	if req.End != "" {
		reqBody["end"] = req.End
	}
	if req.Limit != nil {
		reqBody["limit"] = *req.Limit
	}
	if req.Mode != "" {
		reqBody["mode"] = req.Mode
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body failed: %w", err)
	}

	url := m.scadaURL + "/di/scada/v1/portal/timeSeries"
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create HTTP request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")

	resp, err := m.scadaClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call scada API failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read scada response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scada API returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var tsResp TagTimeSeriesResponse
	if err := json.Unmarshal(respBytes, &tsResp); err != nil {
		return nil, fmt.Errorf("unmarshal scada response failed: %w", err)
	}

	return &tsResp, nil
}

// QueryTagTimeSeries 调用第三方SCADA接口查询测点时序数据。
func (m *Manager) QueryEscadaTagTimeSeries(req *TagTimeSeriesRequest) (*TagTimeSeriesResponse, error) {
	reqBody := map[string]interface{}{
		"select": req.Select,
		"latest": req.Latest,
		"mode":   "raw",
	}
	if req.Start != "" {
		reqBody["start"] = req.Start
	}
	if req.End != "" {
		reqBody["end"] = req.End
	}
	if req.Limit != nil {
		reqBody["limit"] = *req.Limit
	}
	if req.Mode != "" {
		reqBody["mode"] = req.Mode
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body failed: %w", err)
	}

	url := m.escadaURL + "/di/scada/v1/portal/timeSeries"
	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create HTTP request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")

	resp, err := m.escadaClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("call scada API failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read scada response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("scada API returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var tsResp TagTimeSeriesResponse
	if err := json.Unmarshal(respBytes, &tsResp); err != nil {
		return nil, fmt.Errorf("unmarshal scada response failed: %w", err)
	}

	return &tsResp, nil
}
