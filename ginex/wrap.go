package ginex

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

func getResultIfExists(obj any) any {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return obj
	}

	arrVal := val.FieldByName("Result")
	if arrVal.IsValid() {
		return arrVal.Interface()
	}

	return obj
}

func Wrap(f func(*gin.Context) (any, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		obj, err := f(c)
		if err != nil {
			handleError(c, err)
			return
		}

		if obj != nil {
			switch c.Request.Method {
			case "POST":
				c.Status(http.StatusCreated)
				render.WriteJSON(c.Writer, obj)
			case "DELETE":
				c.Status(http.StatusNoContent)
			default:
				//c.SecureJSON(http.StatusOK, obj)
				c.Status(http.StatusOK)
				render.WriteJSON(c.Writer, getResultIfExists(obj))
			}
		}
	}
}
