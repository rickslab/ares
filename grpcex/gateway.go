package grpcex

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rickslab/ares/config"
	"github.com/rickslab/ares/consul"
	"github.com/rickslab/ares/metrics"
	"github.com/rickslab/ares/util"
)

type GrpcGatewayServer struct {
	*runtime.ServeMux
	name       string
	httpServer *http.Server
	register   *consul.Register
}

func NewGatewayServer(name string, opts ...runtime.ServeMuxOption) *GrpcGatewayServer {
	s := &GrpcGatewayServer{}
	s.ServeMux = runtime.NewServeMux(opts...)
	s.name = name

	s.ServeMux.HandlePath("GET", "/health", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		w.WriteHeader(200)
	})
	return s
}

func (s *GrpcGatewayServer) Serve() {
	address := config.YamlEnv().GetString(fmt.Sprintf("service.%s", s.name))

	ip4, port, err := util.AddressToIp4Port(address)
	util.AssertError(err)

	listener, err := net.Listen("tcp", address)
	util.AssertError(err)

	s.Register(s.name, ip4, port)
	go metrics.ReportInfluxDBV2(s.name)

	log.Printf("gRPC-gateway start serving on: %s\n", address)
	s.httpServer = &http.Server{
		Handler: s,
	}
	s.httpServer.Serve(listener)
}

func (s *GrpcGatewayServer) Close() {
	if s.register != nil {
		s.register.Deregister()
	}

	if s.httpServer != nil {
		s.httpServer.Close()
	}
}

func (s *GrpcGatewayServer) Register(name string, address string, port int) {
	host := config.YamlEnv().GetString("service.consul")
	s.register = consul.NewRegisterHTTP(host, name, address, port)

	go func() {
		util.AssertError(util.Retry(s.register.Register, 10, time.Second, time.Second))
	}()
}
