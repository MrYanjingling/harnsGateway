package data

import "time"

// MetricsQueryRequest is the request body for querying metrics data.
type MetricsQueryRequest struct {
	MetricsCodes []string `json:"metricsCodes" binding:"required"`
	StartTime    string   `json:"startTime" binding:"required"` // "2026-07-23 00:00:00"
	EndTime      string   `json:"endTime" binding:"required"`   // "2026-07-23 00:00:00"
	Granularity  string   `json:"granularity" binding:"required,oneof=h"`
}

// MetricsQueryParams holds the parsed query parameters.
type MetricsQueryParams struct {
	MetricsCodes []string
	StartTime    time.Time
	EndTime      time.Time
	Granularity  string
}

// MetricsMetadata holds the metadata of a metrics code from ems_statistics_metrics.
type MetricsMetadata struct {
	ID              string `json:"id"`
	MetricsCode     string `json:"metricsCode"`
	MetricsName     string `json:"metricsName"`
	MeasureType     string `json:"measureType"`
	CalculateType   string `json:"calculateType"`
	TimeGranularity string `json:"timeGranularity"`
}

// MetricsResult holds the aggregated result for a single metrics code.
type MetricsResult struct {
	MetricsCode string  `json:"metricsCode"`
	StartTime   string  `json:"startTime"`
	EndTime     string  `json:"endTime"`
	Value       float64 `json:"value"`
}

// KpiQueryRequest is the request body for querying KPI data.
type KpiQueryRequest struct {
	Ids         []string `json:"ids" binding:"required"`
	StartTime   string   `json:"startTime" binding:"required"` // "2026-07-23 00:00:00"
	EndTime     string   `json:"endTime" binding:"required"`   // "2026-07-23 00:00:00"
	Granularity string   `json:"granularity" binding:"required,oneof=h"`
}

// KpiQueryParams holds the parsed query parameters for KPI.
type KpiQueryParams struct {
	Ids         []string
	StartTime   time.Time
	EndTime     time.Time
	Granularity string
}

// KpiResult holds the aggregated result for a single KPI id.
type KpiResult struct {
	Id        string  `json:"id"`
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
	Value     float64 `json:"value"`
}

// ResponseModel wraps API response data.
type ResponseModel struct {
	Data  interface{} `json:"data,omitempty"`
	Total int64       `json:"total,omitempty"`
}

// KpiInfo 对应 ems_kpi 表的 KPI 信息。
type KpiInfo struct {
	Id              string     `json:"id"`
	TenantId        *string    `json:"tenantId,omitempty"`
	Revision        *int       `json:"revision,omitempty"`
	CreatedBy       *string    `json:"createdBy,omitempty"`
	CreatedTime     *time.Time `json:"createdTime,omitempty"`
	UpdatedBy       *string    `json:"updatedBy,omitempty"`
	UpdatedTime     *time.Time `json:"updatedTime,omitempty"`
	Name            string     `json:"name"`
	Formula         *string    `json:"formula,omitempty"`
	FormulaDataName *string    `json:"formulaDataName,omitempty"`
	FormulaOri      string     `json:"formulaOri"`
	TimeGranularity string     `json:"timeGranularity"`
	TargetData      *string    `json:"targetData,omitempty"`
	SingleQuota     bool       `json:"singleQuota"`
	DefaultTag      bool       `json:"defaultTag"`
	CanDelete       bool       `json:"canDelete"`
	Unit            *string    `json:"unit,omitempty"`
	TargetDataType  string     `json:"targetDataType"`
	Status          string     `json:"status"`
}

// TagInfo 对应 ems_tag 表的测点配置信息。
type TagInfo struct {
	Id           string     `json:"id"`
	TenantId     *string    `json:"tenantId,omitempty"`
	Revision     *int       `json:"revision,omitempty"`
	CreatedBy    *string    `json:"createdBy,omitempty"`
	CreatedTime  *time.Time `json:"createdTime,omitempty"`
	UpdatedBy    *string    `json:"updatedBy,omitempty"`
	UpdatedTime  *time.Time `json:"updatedTime,omitempty"`
	TagId        *string    `json:"tagId,omitempty"`
	TagName      *string    `json:"tagName,omitempty"`
	TagDesc      *string    `json:"tagDesc,omitempty"`
	LastSynTime  *time.Time `json:"lastSynTime,omitempty"`
	TagType      *string    `json:"tagType,omitempty"`
	TagAttribute *string    `json:"tagAttribute,omitempty"`
	Enable       *int       `json:"enable,omitempty"`
	ScadaId      string     `json:"scadaId"`
	Operator     *string    `json:"operator,omitempty"`
	OperateValue *string    `json:"operateValue,omitempty"`
}

// MetricsInfo 对应 ems_statistics_metrics 表的统计指标信息。
type MetricsInfo struct {
	Id            string     `json:"id"`
	TenantId      *string    `json:"tenantId,omitempty"`
	Revision      *int       `json:"revision,omitempty"`
	CreatedBy     *string    `json:"createdBy,omitempty"`
	CreatedTime   *time.Time `json:"createdTime,omitempty"`
	UpdatedBy     *string    `json:"updatedBy,omitempty"`
	UpdatedTime   *time.Time `json:"updatedTime,omitempty"`
	DataPoint     *string    `json:"dataPoint,omitempty"`
	MetricsCode   *string    `json:"metricsCode,omitempty"`
	MetricsName   *string    `json:"metricsName,omitempty"`
	MeasureType   *string    `json:"measureType,omitempty"`
	Enable        *bool      `json:"enable,omitempty"`
	MetricsType   *string    `json:"metricsType,omitempty"`
	CalculateType *string    `json:"calculateType,omitempty"`
	Note          *string    `json:"note,omitempty"`
}

// TagTimeSeriesRequest is the request body for querying tag time series data.
type TagTimeSeriesRequest struct {
	Select string `json:"select" binding:"required"` // 测点ID，多个用逗号分隔
	Start  string `json:"start"`                     // yyyy-MM-dd HH:mm:ss.SSS
	End    string `json:"end"`                       // yyyy-MM-dd HH:mm:ss.SSS
	Latest bool   `json:"latest"`                    // 为true代表最新值
	Limit  *int   `json:"limit"`                     // 数据条数
	Mode   string `json:"mode"`                      // 值类型 默认是raw 平均值mean 差值usage
}

// TagTimeSeriesResponse 第三方接口 DataPointResponse<List<DataPointTimeSeriesTO>> 的映射。
type TagTimeSeriesResponse struct {
	Success bool                   `json:"success"`
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    []*DataPointTimeSeries `json:"data"`
}

// DataPointTimeSeries 第三方接口 DataPointTimeSeriesTO 的映射。
type DataPointTimeSeries struct {
	TagId   string          `json:"tagId"`
	TagName string          `json:"tagName"`
	TagDesc string          `json:"tagDesc"`
	Data    []*TimeSeriesTO `json:"data"`
}

// TimeSeriesTO 第三方接口 TimeSeriesTO 的映射。
type TimeSeriesTO struct {
	Time  string `json:"time"`
	Value string `json:"value"`
}

const (
	TimeLayout = "2006-01-02 15:04:05"
)
