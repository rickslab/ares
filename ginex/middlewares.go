package ginex

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rickslab/ares/logger"
	"github.com/rickslab/ares/metrics"
	"github.com/sirupsen/logrus"
)

func getFullMethod(c *gin.Context) string {
	fullPath := c.FullPath()
	if fullPath == "" {
		fullPath = c.Request.URL.Path
	}
	return fmt.Sprintf("%s%s", c.Request.Method, fullPath)
}

func LogMW(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqId := c.GetHeader("X-Request-Id")
		clientIp := c.GetHeader("X-Real-IP")
		userId := c.GetHeader("X-User-Id")
		scope := c.GetHeader("X-Auth-Scope")

		c.Set("MetaKv", []string{
			"request_id", reqId,
			"client_ip", clientIp,
			"user_id", userId,
			"scope", scope,
			"caller", service,
		})

		log := logger.NewEntry(c, map[string]interface{}{
			"request_id": reqId,
			"client_ip":  clientIp,
			"user_id":    userId,
			"scope":      scope,
			"service":    service,
			"method":     getFullMethod(c),
		})
		c.Set("Logger", log)

		ts := time.Now()
		c.Next()
		dur := time.Since(ts)
		code := c.Writer.Status()

		fields := map[string]interface{}{
			"latency": dur.Seconds() * 1000, // ms
			"code":    code,
		}

		var errMsg string
		if len(c.Errors) > 0 {
			errMsg = c.Errors.ByType(gin.ErrorTypePrivate).String()
		}

		if code >= 200 && code < 400 {
			log.WithFields(fields).Info("OK")
		} else if code >= 400 && code < 500 {
			log.WithFields(fields).Warn(errMsg)
		} else {
			log.WithFields(fields).Error(errMsg)
		}
	}
}

func GetLogger(c *gin.Context) *logrus.Entry {
	return c.MustGet("Logger").(*logrus.Entry)
}

func RecoveryMW() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if ret := recover(); ret != nil {
				GetLogger(c).WithFields(logrus.Fields{
					"stack": string(debug.Stack()),
				}).Fatal(ret)

				c.AbortWithStatus(http.StatusServiceUnavailable)
			}
		}()
		c.Next()
	}
}

func MetricsMW() gin.HandlerFunc {
	return func(c *gin.Context) {
		fullMethod := c.FullPath()
		if fullMethod == "" {
			c.Next()
			return
		}

		ts := time.Now()
		latency := metrics.NewHistogram("latency", "method", fullMethod)
		c.Next()
		latency.Update(time.Since(ts).Nanoseconds())

		status := "success"
		if c.Writer.Status() >= 500 {
			status = "failed"
		}

		call := metrics.NewCounter("call", "method", fullMethod, "status", status, "code", strconv.Itoa(c.Writer.Status()))
		call.Inc(1)
	}
}
