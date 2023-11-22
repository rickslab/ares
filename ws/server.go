package ws

import (
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rickslab/ares/config"
	"github.com/rickslab/ares/consul"
	"github.com/rickslab/ares/metrics"
	"github.com/rickslab/ares/util"
	"github.com/sirupsen/logrus"
)

const (
	readLimit           = 64 * 1024
	healthCheckInterval = 30 * time.Second
	watchInterval       = 10 * time.Second
)

type Server struct {
	name       string
	httpServer *http.Server
	serveMux   *http.ServeMux
	register   *consul.Register
	upgrader   *websocket.Upgrader
	connCount  int64
}

func NewServer(name string, upgrader *websocket.Upgrader) *Server {
	s := &Server{
		name:     name,
		serveMux: http.NewServeMux(),
		upgrader: upgrader,
	}

	s.serveMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	return s
}

func (s *Server) Serve() {
	address := config.YamlEnv().GetString(fmt.Sprintf("service.%s", s.name))

	ip4, port, err := util.AddressToIp4Port(address)
	util.AssertError(err)

	listener, err := net.Listen("tcp", address)
	util.AssertError(err)

	s.Register(s.name, ip4, port)
	go metrics.ReportInfluxDBV2(s.name)
	go s.watchConnCount()
	go func() {
		pprofAddr := fmt.Sprintf("%s:%d", ip4, 10000+port)
		log.Printf("pprof debug serving on: %s\n", pprofAddr)
		http.ListenAndServe(pprofAddr, nil)
	}()

	log.Printf("WebSocket start serving on: %s\n", address)
	s.httpServer = &http.Server{
		Handler: s.serveMux,
	}
	s.httpServer.Serve(listener)
}

func (s *Server) Close() {
	if s.register != nil {
		s.register.Deregister()
	}

	if s.httpServer != nil {
		s.httpServer.Close()
	}
}

func (s *Server) Register(name string, address string, port int) {
	host := config.YamlEnv().GetString("service.consul")
	s.register = consul.NewRegisterHTTP(host, name, address, port)

	go func() {
		util.AssertError(util.Retry(s.register.Register, 10, time.Second, time.Second))
	}()
}

type Handler interface {
	Open(c *Context) error
	Close(c *Context)
	Receive(c *Context, mt int, data []byte) error
}

func (s *Server) Handle(pattern string, h Handler) {
	log.Println("->", pattern)

	s.serveMux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&s.connCount, 1)
		defer atomic.AddInt64(&s.connCount, -1)

		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		c := NewContext(r, conn, s.name, pattern)
		defer c.Cancel()

		conn.SetReadLimit(readLimit)
		conn.SetReadDeadline(time.Now().Add(healthCheckInterval))
		conn.SetPingHandler(func(data string) error {
			c.WriteMessage(websocket.PongMessage, []byte(data))
			return conn.SetReadDeadline(time.Now().Add(healthCheckInterval))
		})

		logger := c.Logger()
		defer func() {
			if ret := recover(); ret != nil {
				logger.WithFields(logrus.Fields{
					"stack": string(debug.Stack()),
				}).Error("Recover panic", ret)
			}
		}()

		err = h.Open(c)
		if err != nil {
			logger.Error("Open error", err)
			return
		}
		logger.Trace("Open")

		defer func() {
			h.Close(c)
			logger.Trace("Closed")
		}()

		for {
			msgType, data, err := c.ReadMessage()
			if err != nil {
				return
			}

			err = h.Receive(c, msgType, data)
			if err != nil {
				logger.Error("Receive error", err)
				return
			}
			logger.Tracef("Receive %d bytes", len(data))
		}
	})
}

func (s *Server) watchConnCount() {
	g := metrics.NewGauge("numsocket")
	t := time.Tick(watchInterval)
	for {
		select {
		case <-t:
			g.Update(atomic.LoadInt64(&s.connCount))
		}
	}
}
