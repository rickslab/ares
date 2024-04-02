package grpcex

import (
	"context"
	"regexp"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rickslab/ares/errcode"
	"github.com/rickslab/ares/metrics"
	"github.com/rickslab/ares/ws"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	regFullMethod = regexp.MustCompile("^/([^/]+)/(.+)$")
)

func ContextUCI() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		switch c := ctx.(type) {
		case *gin.Context:
			kv := c.MustGet("MetaKv").([]string)
			ctx = metadata.AppendToOutgoingContext(ctx, kv...)
		case *ws.Context:
			kv := c.Value("MetaKv").([]string)
			ctx = metadata.AppendToOutgoingContext(ctx, kv...)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func TimeoutUCI(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		newCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := invoker(newCtx, method, req, reply, cc, opts...)
		if newCtx.Err() == context.DeadlineExceeded {
			return status.Error(errcode.ErrRpcTimeout, "rpc timeout")
		}
		return err
	}
}

func ContextUSI() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		subs := regFullMethod.FindStringSubmatch(info.FullMethod)
		if len(subs) != 3 {
			return handler(ctx, req)
		}
		c := NewContext(ctx, subs[1], subs[2])

		return handler(c.NewCtx(ctx), req)
	}
}

func LogUSI() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		c := GetContext(ctx)
		if c.Service == "grpc.health.v1.Health" {
			return handler(ctx, req)
		}

		ts := time.Now()
		resp, err := handler(ctx, req)
		dur := time.Since(ts)

		fields := map[string]any{
			"req":     req,
			"latency": dur.Seconds() * 1000, // ms
		}

		code, failed := errcode.From(err)
		fields["code"] = code

		if code == 0 {
			c.Logger.WithFields(fields).Info("OK")
		} else if failed {
			c.Logger.WithFields(fields).Error(err)
		} else {
			c.Logger.WithFields(fields).Warn(err)
		}
		return resp, err
	}
}

func RecoveryUSI() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if ret := recover(); ret != nil {
				GetLogger(ctx).WithFields(logrus.Fields{
					"stack": string(debug.Stack()),
				}).Fatal(ret)

				if retErr, ok := ret.(error); ok {
					err = retErr
				} else {
					err = status.Errorf(errcode.ErrRpcPanic, "recover panic: %v", ret)
				}
			}
		}()
		return handler(ctx, req)
	}
}

func MetricsUSI() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		c := GetContext(ctx)
		if c.Service == "grpc.health.v1.Health" {
			return handler(ctx, req)
		}

		ts := time.Now()
		latency := metrics.NewHistogram("latency", "method", info.FullMethod)
		resp, err := handler(ctx, req)
		latency.Update(time.Since(ts).Nanoseconds())

		code, failed := errcode.From(err)
		status := "success"
		if failed {
			status = "failed"
		}

		call := metrics.NewCounter("call", "method", info.FullMethod, "status", status, "code", strconv.Itoa(code))
		call.Inc(1)
		return resp, err
	}
}

func ErrorMapUSI() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		return resp, errcode.ErrorMap(err)
	}
}

func RequestBindUSI(caller string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		c := GetContext(ctx)
		if c.Caller == caller {
			c.Bind(req)
		}
		return handler(ctx, req)
	}
}
