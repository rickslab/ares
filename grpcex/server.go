package grpcex

import (
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"time"

	"github.com/rickslab/ares/config"
	"github.com/rickslab/ares/consul"
	"github.com/rickslab/ares/metrics"
	"github.com/rickslab/ares/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

const (
	defaultWindowSize = 1024 * 1024 * 1024
)

type GrpcServer struct {
	*grpc.Server
	listener net.Listener
	register *consul.Register
}

func NewServer(opt ...grpc.ServerOption) *GrpcServer {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			ContextUSI(),
			LogUSI(),
			RecoveryUSI(),
			MetricsUSI(),
			ErrorMapUSI(),
		),
		grpc.InitialWindowSize(defaultWindowSize),
		grpc.InitialConnWindowSize(defaultWindowSize),
	}

	return &GrpcServer{
		Server: grpc.NewServer(
			append(opts, opt...)...,
		),
	}
}

func (s *GrpcServer) Serve() {
	var serviceName string
	for name := range s.GetServiceInfo() {
		serviceName = name
		break
	}

	address := config.YamlEnv().GetString(fmt.Sprintf("service.%s", serviceName))

	ip4, port, err := util.AddressToIp4Port(address)
	util.AssertError(err)

	listener, err := net.Listen("tcp", address)
	util.AssertError(err)
	s.listener = listener

	s.Register(serviceName, ip4, port)
	go metrics.ReportInfluxDBV2(serviceName)
	go func() {
		pprofAddr := fmt.Sprintf("%s:%d", ip4, 10000+port)
		log.Printf("pprof debug serving on: %s\n", pprofAddr)
		http.ListenAndServe(pprofAddr, nil)
	}()

	log.Printf("gRPC start serving on: %s\n", address)
	s.Server.Serve(listener)
}

func (s *GrpcServer) Close() {
	if s.register != nil {
		s.register.Deregister()
	}

	if s.listener != nil {
		s.listener.Close()
	}
}

func (s *GrpcServer) Register(name string, address string, port int) {
	grpc_health_v1.RegisterHealthServer(s.Server, new(healthService))

	host := config.YamlEnv().GetString("service.consul")
	s.register = consul.NewRegisterGRPC(host, name, address, port)

	go func() {
		util.AssertError(util.Retry(s.register.Register, 10, time.Second, time.Second))
	}()
}
