package ginex

import (
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rickslab/ares/config"
	"github.com/rickslab/ares/consul"
	"github.com/rickslab/ares/errcode"
	"github.com/rickslab/ares/metrics"
	"github.com/rickslab/ares/util"
	"google.golang.org/grpc/status"
)

type GinServer struct {
	*gin.Engine
	name     string
	listener net.Listener
	register *consul.Register
}

func NewServer(name string) *GinServer {
	s := &GinServer{}
	s.Engine = gin.New()
	s.name = name

	s.GET("health", func(c *gin.Context) {
		c.Status(200)
	})

	s.Use(
		LogMW(name),
		RecoveryMW(),
		MetricsMW(),
	)
	return s
}

func (s *GinServer) Serve() {
	address := config.YamlEnv().GetString(fmt.Sprintf("service.%s", s.name))

	ip4, port, err := util.AddressToIp4Port(address)
	util.AssertError(err)

	listener, err := net.Listen("tcp", address)
	util.AssertError(err)
	s.listener = listener

	s.Register(s.name, ip4, port)
	go metrics.ReportInfluxDBV2(s.name)
	go func() {
		pprofAddr := fmt.Sprintf("%s:%d", ip4, 10000+port)
		log.Printf("pprof debug serving on: %s\n", pprofAddr)
		http.ListenAndServe(pprofAddr, nil)
	}()

	log.Printf("Gin start serving on: %s\n", address)
	s.RunListener(listener)
}

func (s *GinServer) Close() {
	if s.register != nil {
		s.register.Deregister()
	}

	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *GinServer) Register(name string, address string, port int) {
	host := config.YamlEnv().GetString("service.consul")
	s.register = consul.NewRegisterHTTP(host, name, address, port)

	go func() {
		util.AssertError(util.Retry(s.register.Register, 10, time.Second, time.Second))
	}()
}

func ParamInt(c *gin.Context, key string) (int64, error) {
	str := c.Param(key)

	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, status.Errorf(errcode.ErrGinParam, "param :%s need integer", key)
	}

	return val, nil
}
