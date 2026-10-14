package device

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	jsonpatch "github.com/evanphx/json-patch"
	"github.com/gin-gonic/gin"
	"harns-platform/pkg/apis"
	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/gateway/registry"
	"harns-platform/pkg/gateway/runtime"
	"harns-platform/pkg/logging"
	"io"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func InstallHandler(group *gin.RouterGroup, mgr *Manager) {
	group.POST("/devices", createDevice(mgr))
	group.DELETE("/devices/:id", deleteDevice(mgr))
	group.PATCH("/devices/:id", patchDeviceById(mgr))
	group.PUT("/devices/:id", updateDeviceById(mgr))
	group.GET("/devices", listDevices(mgr))
	group.GET("/devices/:id", getDeviceById(mgr))
	group.PUT("/devices/:id/:status", switchDeviceStatusById(mgr))
	group.PUT("/devices/:id/action", controlDeviceById(mgr))

}

func createDevice(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logging.V(2).InfoS("Failed to get request body", "err", err)
			response.WriteError(c, http.StatusBadRequest, response.ErrMalformedJSON)
		}

		var target struct {
			DeviceType string `json:"deviceType"`
		}
		err = json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&target)
		if err != nil {
			logging.V(2).InfoS("Failed to parse device type", "err", err)
			response.WriteError(c, http.StatusBadRequest, response.ErrRequestBody)
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		object := registry.DeviceTypeMap[target.DeviceType]()
		if err := c.ShouldBindJSON(object); err != nil {
			logging.V(2).InfoS("Failed to parse Device", "err", err)
			response.WriteError(c, http.StatusBadRequest, response.ErrMalformedJSON)
			return
		}
		d, err := mgr.CreateDevice(object)

		if err != nil {
			response.WriteError(c, http.StatusBadRequest, err)
			return
		}

		// TODO use different scheme
		c.Header(apis.ETag, fmt.Sprintf("%s", d.GetVersion()))
		c.Header(apis.Location, fmt.Sprintf("https://%s%s/%s", c.Request.Host, c.Request.RequestURI, d.GetID()))
		response.WriteOK(c, d)
	}
}

func deleteDevice(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		eTag := c.GetHeader(apis.IfMatch)
		if len(eTag) == 0 {
			c.Status(http.StatusPreconditionRequired)
			return
		}
		device, err := mgr.DeleteDevice(id, eTag)
		if err != nil {
			if os.IsNotExist(err) {
				c.Status(http.StatusNotFound)
			} else if errors.Is(err, apis.ErrMismatch) {
				c.Status(http.StatusPreconditionFailed)
			} else {
				response.WriteError(c, http.StatusBadRequest, err)
			}
			return
		}
		response.WriteOK(c, device)
	}
}

func patchDeviceById(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Request.Body.Close()

		contentType := c.GetHeader("Content-Type")
		// Remove "; charset=" if included in header.
		if idx := strings.Index(contentType, ";"); idx > 0 {
			contentType = contentType[:idx]
		}

		if !patchTypes.Has(contentType) {
			c.Status(http.StatusUnsupportedMediaType)
			return
		}

		eTag := c.GetHeader(apis.IfMatch)
		if len(eTag) == 0 {
			c.Status(http.StatusPreconditionRequired)
			return
		}

		pathBytes, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logging.V(3).InfoS("Failed to read", "err", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		id := c.Param("id")
		old, err := mgr.GetDeviceById(id, true)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		versionedJS, err := json.Marshal(old)
		if err != nil {
			logging.V(3).InfoS("Failed to marshal", "err", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		patchedJS, err := applyJSPatch(types.PatchType(contentType), pathBytes, versionedJS)
		if err != nil {
			response.WriteJSONP(c, http.StatusBadRequest, err)
			return
		}

		newObj := registry.DeviceTypeMap[old.GetDeviceType()]()
		if err := json.NewDecoder(bytes.NewBuffer(patchedJS)).Decode(newObj); err != nil {
			logging.V(3).InfoS("Failed to decode", "err", err)
			response.WriteError(c, http.StatusBadRequest, response.ErrMalformedJSON)
			return
		}

		updated, err := mgr.UpdateDeviceById(id, eTag, newObj)
		if err != nil {
			switch {
			case os.IsNotExist(err):
				c.Status(http.StatusNotFound)
			case errors.Is(err, apis.ErrMismatch):
				c.Status(http.StatusPreconditionFailed)
			default:
				if response.IsResponseError(err) {
					response.WriteError(c, http.StatusBadRequest, err)
				} else {
					c.Status(http.StatusInternalServerError)
				}
			}
			return
		}

		c.Header(apis.ETag, updated.GetVersion())
		response.WriteOK(c, updated)
	}
}

func updateDeviceById(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Request.Body.Close()

		eTag := c.GetHeader(apis.IfMatch)
		if len(eTag) == 0 {
			c.Status(http.StatusPreconditionRequired)
			return
		}

		id := c.Param("id")
		old, err := mgr.GetDeviceById(id, true)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		newObj := registry.DeviceTypeMap[old.GetDeviceType()]()
		if err := json.NewDecoder(c.Request.Body).Decode(newObj); err != nil {
			logging.V(3).InfoS("Failed to decode", "err", err)
			response.WriteError(c, http.StatusBadRequest, response.ErrMalformedJSON)
			return
		}

		updated, err := mgr.UpdateDeviceById(id, eTag, newObj)
		if err != nil {
			switch {
			case os.IsNotExist(err):
				c.Status(http.StatusNotFound)
			case errors.Is(err, apis.ErrMismatch):
				c.Status(http.StatusPreconditionFailed)
			default:
				if response.IsResponseError(err) {
					response.WriteError(c, http.StatusBadRequest, err)
				} else {
					c.Status(http.StatusInternalServerError)
				}
			}
			return
		}

		if updated != nil {
			c.Header(apis.ETag, updated.GetVersion())
		}
		response.WriteOK(c, updated)
	}
}

func listDevices(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Request.Body.Close()

		query := c.Request.URL.Query()
		exploded := false
		filter := runtime.DeviceFilter{}
		if len(query) > 0 {
			v := query.Get(apis.Filter)
			if len(v) > 0 {
				if err := json.Unmarshal([]byte(v), &filter); err != nil {
					response.WriteError(c, http.StatusBadRequest, response.ErrMalformedJSON)
					return
				}
			}
			exploded, _ = strconv.ParseBool(query.Get("exploded"))
		}
		rds, _ := mgr.ListDevices(&filter, exploded)

		response.WriteOK(c, &runtime.ResponseModel{Devices: rds})
	}
}

func getDeviceById(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Request.Body.Close()

		id := c.Param("id")
		query := c.Request.URL.Query()
		exploded := false
		if len(query) > 0 {
			exploded, _ = strconv.ParseBool(query.Get("exploded"))
		}
		rd, err := mgr.GetDeviceById(id, exploded)
		if err != nil {
			if os.IsNotExist(err) {
				c.Status(http.StatusNotFound)
			} else {
				c.Status(http.StatusInternalServerError)
			}
			return
		}

		c.Header(apis.ETag, fmt.Sprintf("%s", rd.GetVersion()))
		response.WriteOK(c, rd)
	}
}

func switchDeviceStatusById(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Request.Body.Close()

		id := c.Param("id")
		status := c.Param("status")
		if err := mgr.SwitchDeviceStatus(id, status); err != nil {
			if os.IsNotExist(err) {
				c.Status(http.StatusNotFound)
			} else {
				response.WriteError(c, http.StatusBadRequest, err)
			}
		}
		c.Status(http.StatusAccepted)
	}
}

func controlDeviceById(mgr *Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer c.Request.Body.Close()

		id := c.Param("id")
		var actions []map[string]interface{}
		if err := json.NewDecoder(c.Request.Body).Decode(&actions); err != nil {
			logging.V(3).InfoS("Failed to parse action", "err", err)
			response.WriteError(c, http.StatusBadRequest, response.ErrMalformedJSON)
			return
		}

		err := mgr.DeliverAction(id, actions)

		if err != nil {
			response.WriteError(c, http.StatusBadRequest, err)
			return
		}

		c.Status(http.StatusAccepted)
	}
}

func applyJSPatch(patchType types.PatchType, patchBytes, versionedJS []byte) (patchedJS []byte, err error) {
	switch patchType {
	case types.JSONPatchType:
		patchObj, err := jsonpatch.DecodePatch(patchBytes)
		if err != nil {
			return nil, response.ErrMalformedJSON
		}
		if len(patchObj) > maxJSONPatchOperations {
			logging.V(3).InfoS("Too many json patch operations", "count", len(patchObj))
			return nil, response.ErrTooManyJsonPatchOperations(maxJSONPatchOperations)
		}
		patchedJS, err := patchObj.Apply(versionedJS)
		if err != nil {
			logging.V(3).InfoS("Failed to apply json patch", "err", err)
			return nil, response.ErrMalformedJSON
		}
		return patchedJS, nil
	case types.MergePatchType:
		patchedJS, err = jsonpatch.MergePatch(versionedJS, patchBytes)
		if err != nil {
			logging.V(3).InfoS("Failed to apply json merge patch", "err", err)
			return nil, response.ErrMalformedJSON
		}
		return patchedJS, err
	default:
		// only here as a safety net - gin filters content-type
		return nil, fmt.Errorf("unknown Content-Type header for patch: %v", patchType)
	}
}
