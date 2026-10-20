package httpx

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"harns-platform/pkg/apis/response"
	"harns-platform/pkg/i18n"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "" || name == "-" {
				return fld.Name
			}
			return name
		})
	}
}

// BindJSON binds the JSON body and writes field-level validation errors on
// failure.
func BindJSON(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		writeBindError(c, err)
		return false
	}
	return true
}

// BindQuery binds the query string and writes field-level validation errors on
// failure.
func BindQuery(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindQuery(obj); err != nil {
		writeBindError(c, err)
		return false
	}
	return true
}

func writeBindError(c *gin.Context, err error) {
	var ves validator.ValidationErrors
	if errors.As(err, &ves) {
		lang := i18n.Lang(c)
		items := make([]response.ErrorItem, 0, len(ves))
		for _, fe := range ves {
			items = append(items, response.ErrorItem{
				Code:    int(response.ErrCodeValidationFailed),
				Message: validationMessage(lang, fe),
				Field:   fe.Field(),
			})
		}
		c.JSON(http.StatusBadRequest, response.Result{
			Code:    int(response.ErrCodeValidationFailed),
			Message: i18n.Translate(lang, "error.validation_failed"),
			Errors:  items,
		})
		c.Abort()
		return
	}
	response.WriteError(c, http.StatusBadRequest, response.ErrMalformedJSON)
	c.Abort()
}

func validationMessage(lang string, fe validator.FieldError) string {
	key := "validate." + fe.Tag()
	switch fe.Tag() {
	case "min", "max", "len", "gte", "lte", "oneof":
		return i18n.Translate(lang, key, fe.Field(), fe.Param())
	default:
		return i18n.Translate(lang, key, fe.Field())
	}
}
