package grpcex

import (
	"fmt"
	"sync"
	"time"

	"github.com/rickslab/ares/config"
	_ "github.com/rickslab/ares/grpcex/balancer/chash"
	_ "github.com/rickslab/ares/grpcex/resolver/consul"
	"github.com/rickslab/ares/util"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultRpcTimeout = 5 * time.Second
)

var (
	clients = map[string]*grpc.ClientConn{}
	mu      = sync.RWMutex{}
)

func Dial(address string, balancer string, mws ...grpc.UnaryClientInterceptor) (*grpc.ClientConn, error) {
	serverConfig := fmt.Sprintf(`{"loadBalancingConfig": [{"%s": {}}]}`, balancer)
	return grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(serverConfig),
		grpc.WithChainUnaryInterceptor(mws...),
		grpc.WithInitialWindowSize(defaultWindowSize),
		grpc.WithInitialConnWindowSize(defaultWindowSize),
	)
}

func DialByName(target string, balancer string, mws ...grpc.UnaryClientInterceptor) (*grpc.ClientConn, error) {
	consulAddress := config.YamlEnv().GetString("service.consul")
	return Dial(fmt.Sprintf("consul://%s/%s", consulAddress, target), balancer, mws...)
}

func initGrpcConn(name string, balancer string) *grpc.ClientConn {
	mu.Lock()
	defer mu.Unlock()

	conn, ok := clients[name]
	if ok {
		return conn
	}

	conn, err := DialByName(name, balancer,
		ContextUCI(),
		TimeoutUCI(defaultRpcTimeout),
	)
	util.AssertError(err)

	clients[name] = conn
	return conn
}

func getGrpcConn(name string) *grpc.ClientConn {
	mu.RLock()
	defer mu.RUnlock()

	return clients[name]
}

func Client(name string, balancers ...string) *grpc.ClientConn {
	conn := getGrpcConn(name)
	if conn != nil {
		return conn
	}

	balancer := "round_robin"
	if len(balancers) > 0 {
		balancer = balancers[0]
	}
	return initGrpcConn(name, balancer)
}
