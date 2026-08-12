package data

import (
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

func InstallHandler(group *gin.RouterGroup, mgr *Manager) {
	group.POST("/metricsData", queryMetrics(mgr))
	group.GET("/metrics", listMetrics(mgr))
	group.POST("/kpi", queryKpi(mgr))
	group.GET("/kpis", listKpis(mgr))
	group.GET("/tags", listTags(mgr))
	group.POST("/fmcs/tagTimeSeries", queryFmcsTagTimeSeries(mgr))
	group.POST("/escada/tagTimeSeries", queryEscadaTagTimeSeries(mgr))
}

func queryMetrics(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req MetricsQueryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			klog.V(4).InfoS("queryMetrics: invalid request", "error", err)
			c.JSON(http.StatusBadRequest, ResponseModel{
				Data: "invalid request: " + err.Error(),
			})
			return
		}
		klog.V(4).InfoS("queryMetrics", "metricsCodes", req.MetricsCodes, "startTime", req.StartTime, "endTime", req.EndTime, "granularity", req.Granularity)

		startTime, err := time.ParseInLocation(TimeLayout, req.StartTime, time.Local)
		if err != nil {
			klog.V(4).InfoS("queryMetrics: invalid startTime", "startTime", req.StartTime, "error", err)
			c.JSON(http.StatusBadRequest, ResponseModel{
				Data: "invalid startTime, expected format: " + TimeLayout,
			})
			return
		}

		endTime, err := time.ParseInLocation(TimeLayout, req.EndTime, time.Local)
		if err != nil {
			klog.V(4).InfoS("queryMetrics: invalid endTime", "endTime", req.EndTime, "error", err)
			c.JSON(http.StatusBadRequest, ResponseModel{
				Data: "invalid endTime, expected format: " + TimeLayout,
			})
			return
		}

		params := &MetricsQueryParams{
			MetricsCodes: req.MetricsCodes,
			StartTime:    startTime,
			EndTime:      endTime,
			Granularity:  req.Granularity,
		}

		results, err := mgr.QueryMetrics(params)
		if err != nil {
			klog.ErrorS(err, "Failed to query metrics")
			c.JSON(http.StatusInternalServerError, ResponseModel{
				Data: "failed to query metrics",
			})
			return
		}

		// 保持入参顺序输出
		resultList := make([]*MetricsResult, 0, len(req.MetricsCodes))
		for _, code := range req.MetricsCodes {
			if r, ok := results[code]; ok {
				resultList = append(resultList, r)
			}
		}
		sort.SliceStable(resultList, func(i, j int) bool {
			return indexOf(req.MetricsCodes, resultList[i].MetricsCode) <
				indexOf(req.MetricsCodes, resultList[j].MetricsCode)
		})

		klog.V(4).InfoS("queryMetrics: success", "resultCount", len(resultList))
		c.JSON(http.StatusOK, ResponseModel{
			Data:  resultList,
			Total: int64(len(req.MetricsCodes)),
		})
	}
}

func queryKpi(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req KpiQueryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			klog.V(4).InfoS("queryKpi: invalid request", "error", err)
			c.JSON(http.StatusBadRequest, ResponseModel{
				Data: "invalid request: " + err.Error(),
			})
			return
		}
		klog.V(4).InfoS("queryKpi", "ids", req.Ids, "startTime", req.StartTime, "endTime", req.EndTime, "granularity", req.Granularity)

		startTime, err := time.ParseInLocation(TimeLayout, req.StartTime, time.Local)
		if err != nil {
			klog.V(4).InfoS("queryKpi: invalid startTime", "startTime", req.StartTime, "error", err)
			c.JSON(http.StatusBadRequest, ResponseModel{
				Data: "invalid startTime, expected format: " + TimeLayout,
			})
			return
		}

		endTime, err := time.ParseInLocation(TimeLayout, req.EndTime, time.Local)
		if err != nil {
			klog.V(4).InfoS("queryKpi: invalid endTime", "endTime", req.EndTime, "error", err)
			c.JSON(http.StatusBadRequest, ResponseModel{
				Data: "invalid endTime, expected format: " + TimeLayout,
			})
			return
		}

		params := &KpiQueryParams{
			Ids:         req.Ids,
			StartTime:   startTime,
			EndTime:     endTime,
			Granularity: req.Granularity,
		}

		results, err := mgr.QueryKpi(params)
		if err != nil {
			klog.ErrorS(err, "Failed to query kpi")
			c.JSON(http.StatusInternalServerError, ResponseModel{
				Data: "failed to query kpi",
			})
			return
		}

		// 保持入参顺序输出
		resultList := make([]*KpiResult, 0, len(req.Ids))
		for _, id := range req.Ids {
			if r, ok := results[id]; ok {
				resultList = append(resultList, r)
			}
		}
		sort.SliceStable(resultList, func(i, j int) bool {
			return indexOf(req.Ids, resultList[i].Id) <
				indexOf(req.Ids, resultList[j].Id)
		})

		klog.V(4).InfoS("queryKpi: success", "resultCount", len(resultList))
		c.JSON(http.StatusOK, ResponseModel{
			Data:  resultList,
			Total: int64(len(req.Ids)),
		})
	}
}

func listKpis(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		klog.V(4).InfoS("listKpis")

		result, err := mgr.ListKpis()
		if err != nil {
			klog.ErrorS(err, "Failed to list kpis")
			c.JSON(http.StatusInternalServerError, ResponseModel{
				Data: "failed to list kpis",
			})
			return
		}

		klog.V(4).InfoS("listKpis: success", "count", len(result))
		c.JSON(http.StatusOK, ResponseModel{
			Data:  result,
			Total: int64(len(result)),
		})
	}
}

func listMetrics(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		klog.V(4).InfoS("listMetrics")

		result, err := mgr.ListMetrics()
		if err != nil {
			klog.ErrorS(err, "Failed to list metrics")
			c.JSON(http.StatusInternalServerError, ResponseModel{
				Data: "failed to list metrics",
			})
			return
		}

		klog.V(4).InfoS("listMetrics: success", "count", len(result))
		c.JSON(http.StatusOK, ResponseModel{
			Data:  result,
			Total: int64(len(result)),
		})
	}
}

func listTags(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		klog.V(4).InfoS("listTags")

		result, err := mgr.ListTags()
		if err != nil {
			klog.ErrorS(err, "Failed to list tags")
			c.JSON(http.StatusInternalServerError, ResponseModel{
				Data: "failed to list tags",
			})
			return
		}

		klog.V(4).InfoS("listTags: success", "count", len(result))
		c.JSON(http.StatusOK, ResponseModel{
			Data:  result,
			Total: int64(len(result)),
		})
	}
}

func queryEscadaTagTimeSeries(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TagTimeSeriesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			klog.V(4).InfoS("queryTagTimeSeries: invalid request", "error", err)
			c.JSON(http.StatusBadRequest, ResponseModel{
				Data: "invalid request: " + err.Error(),
			})
			return
		}
		klog.V(4).InfoS("queryTagTimeSeries", "select", req.Select, "start", req.Start, "end", req.End, "latest", req.Latest, "limit", req.Limit, "mode", req.Mode)

		result, err := mgr.QueryEscadaTagTimeSeries(&req)
		if err != nil {
			klog.ErrorS(err, "Failed to query tag time series")
			c.JSON(http.StatusInternalServerError, ResponseModel{
				Data: "failed to query tag time series",
			})
			return
		}

		if !result.Success {
			klog.V(4).InfoS("queryTagTimeSeries: scada returned non-200", "code", result.Code, "message", result.Message)
			c.JSON(http.StatusInternalServerError, ResponseModel{
				Data: result.Message,
			})
			return
		}

		klog.V(4).InfoS("queryTagTimeSeries: success", "dataCount", len(result.Data))
		c.JSON(http.StatusOK, ResponseModel{
			Data: result.Data,
		})
	}
}

func queryFmcsTagTimeSeries(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TagTimeSeriesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			klog.V(4).InfoS("queryTagTimeSeries: invalid request", "error", err)
			c.JSON(http.StatusBadRequest, ResponseModel{
				Data: "invalid request: " + err.Error(),
			})
			return
		}
		klog.V(4).InfoS("queryTagTimeSeries", "select", req.Select, "start", req.Start, "end", req.End, "latest", req.Latest, "limit", req.Limit, "mode", req.Mode)

		result, err := mgr.QueryTagTimeSeries(&req)
		if err != nil {
			klog.ErrorS(err, "Failed to query tag time series")
			c.JSON(http.StatusInternalServerError, ResponseModel{
				Data: "failed to query tag time series",
			})
			return
		}

		if !result.Success {
			klog.V(4).InfoS("queryTagTimeSeries: scada returned non-200", "code", result.Code, "message", result.Message)
			c.JSON(http.StatusInternalServerError, ResponseModel{
				Data: result.Message,
			})
			return
		}

		klog.V(4).InfoS("queryTagTimeSeries: success", "dataCount", len(result.Data))
		c.JSON(http.StatusOK, ResponseModel{
			Data: result.Data,
		})
	}
}

func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}
