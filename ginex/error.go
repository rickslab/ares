package ginex

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rickslab/ares/errcode"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Error(code int, msg string) error {
	return status.Error(codes.Code(code), msg)
}

func handleError(c *gin.Context, err error) {
	statusCode := http.StatusInternalServerError
	err = errcode.ErrorMap(err)

	st, ok := status.FromError(err)
	if !ok {
		c.JSON(statusCode, gin.H{
			"code":    errcode.ErrGinFailed,
			"message": err.Error(),
		})
		return
	}

	code := int(st.Code())
	if code > 1000 {
		statusCode = code / 1000
	}

	c.JSON(statusCode, st.Proto())
}
