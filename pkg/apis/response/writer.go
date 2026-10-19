package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"harns-platform/pkg/i18n"
)

const (
	// SuccessCode is the business code returned on success.
	SuccessCode = 200
	// SuccessMessage is the message returned on success.
	SuccessMessage = "success"
)

// Result is the unified response envelope shared by every endpoint.
//
//	success: {"code":200,"message":"success","data":{...}}
//	error:   {"code":10012,"message":"Device [x] not found.","errors":[{"code":..,"message":..}, ...]}
//
// The frontend only needs one rule: code == 200 ? use data : use errors.
type Result struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Errors  []ErrorItem `json:"errors,omitempty"`
}

// ErrorItem is a single localized error entry carried in Result.Errors.
type ErrorItem struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OK builds a successful result.
func OK(data interface{}) Result {
	return Result{Code: SuccessCode, Message: SuccessMessage, Data: data}
}

// WriteOK writes a successful result with HTTP 200.
func WriteOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, OK(data))
}

// ErrorResult flattens errs (expanding MultiError) and builds the localized
// error result. Code/Message summarize the first error, and Errors always
// carries the full list (length 1 for a single error).
func ErrorResult(lang string, errs ...error) Result {
	flat := flatten(errs)
	if len(flat) == 0 {
		return Result{Code: SuccessCode, Message: SuccessMessage}
	}

	items := make([]ErrorItem, 0, len(flat))
	for _, err := range flat {
		items = append(items, localizeOne(lang, err))
	}
	return Result{Code: items[0].Code, Message: items[0].Message, Errors: items}
}

// WriteError writes errs as a localized JSON error response. The HTTP status is
// preserved so callers keep REST semantics (400/404/409/412/...).
func WriteError(c *gin.Context, status int, errs ...error) {
	c.JSON(status, ErrorResult(i18n.Lang(c), errs...))
}

// WriteJSONP writes errs as a localized JSONP error response.
func WriteJSONP(c *gin.Context, status int, errs ...error) {
	c.JSONP(status, ErrorResult(i18n.Lang(c), errs...))
}

func flatten(errs []error) []error {
	flat := make([]error, 0, len(errs))
	for _, err := range errs {
		if err == nil {
			continue
		}
		if me, ok := err.(*MultiError); ok {
			flat = append(flat, me.Errors()...)
			continue
		}
		flat = append(flat, err)
	}
	return flat
}

func localizeOne(lang string, err error) ErrorItem {
	if re, ok := err.(*responseError); ok {
		loc := re.Localized(lang)
		return ErrorItem{Code: int(loc.Code), Message: loc.Message, Data: loc.data}
	}
	// Unknown error types keep the raw message and use code 0.
	return ErrorItem{Code: 0, Message: err.Error()}
}

// ErrorCode returns the business code of a response error, or 0 for other errors.
func ErrorCode(err error) int {
	if re, ok := err.(*responseError); ok && re != nil {
		return int(re.Code)
	}
	return 0
}
