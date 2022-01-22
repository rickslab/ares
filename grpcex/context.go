package grpcex

import (
	"context"
	"reflect"
	"strconv"

	"github.com/google/uuid"
	"github.com/rickslab/ares/logger"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
)

type GrpcCtxKey string

const (
	grpcCtxKey GrpcCtxKey = "grpc_ctx"
)

type Context struct {
	Service   string
	Method    string
	Caller    string
	RequestId string
	ConnectId string
	ClientIp  string
	UserId    int64
	Scope     string
	Logger    *logrus.Entry
}

func NewContext(ctx context.Context, service string, method string) *Context {
	c := &Context{
		Service: service,
		Method:  method,
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		c.Caller = getMeta(md, "caller")
		c.RequestId = getMeta(md, "request_id")
		c.ConnectId = getMeta(md, "connect_id")
		c.ClientIp = getMeta(md, "client_ip")
		userId := getMeta(md, "user_id")
		if userId != "" {
			uid, _ := strconv.ParseInt(userId, 10, 64)
			c.UserId = uid
		}
		c.Scope = getMeta(md, "scope")
	} else {
		c.RequestId = uuid.New().String()
	}

	c.Logger = logger.NewEntry(ctx, logrus.Fields{
		"service":    c.Service,
		"method":     c.Method,
		"caller":     c.Caller,
		"request_id": c.RequestId,
		"connect_id": c.ConnectId,
		"client_ip":  c.ClientIp,
		"user_id":    c.UserId,
		"scope":      c.Scope,
	})
	return c
}

func (c *Context) NewCtx(ctx context.Context) context.Context {
	kv := []string{"caller", c.Service, "request_id", c.RequestId}
	if c.ConnectId != "" {
		kv = append(kv, "connect_id", c.ConnectId)
	}
	if c.ClientIp != "" {
		kv = append(kv, "client_ip", c.ClientIp)
	}
	if c.UserId > 0 {
		kv = append(kv, "user_id", strconv.FormatInt(c.UserId, 10))
	}
	if c.Scope != "" {
		kv = append(kv, "scope", c.Scope)
	}

	ctx = metadata.AppendToOutgoingContext(ctx, kv...)
	return context.WithValue(ctx, grpcCtxKey, c)
}

func (c *Context) Bind(req interface{}) {
	r := reflect.ValueOf(req)
	if r.Kind() == reflect.Ptr {
		r = r.Elem()
	}

	t := r.Type()
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		tag := ft.Tag.Get("ctx")
		if tag == "" {
			continue
		}

		fv := r.Field(i)
		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}
		if !fv.IsValid() {
			continue
		}

		switch tag {
		case "user_id":
			fv.SetInt(c.UserId)
		}
	}
}

func getMeta(md metadata.MD, key string) string {
	vals := md.Get(key)
	if len(vals) == 0 || vals[0] == "" {
		return ""
	}
	return vals[0]
}

func GetContext(ctx context.Context) *Context {
	return ctx.Value(grpcCtxKey).(*Context)
}

func GetLogger(ctx context.Context) *logrus.Entry {
	return GetContext(ctx).Logger
}
