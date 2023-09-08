package turn

import (
	"fmt"
	"log"
	"net"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/pion/turn/v2"
	"github.com/rickslab/ares/cache"
	"github.com/rickslab/ares/config"
	"github.com/rickslab/ares/crypto"
	"github.com/rickslab/ares/env"
	"github.com/rickslab/ares/ginex"
	"github.com/rickslab/ares/util"
)

type TurnServer struct {
	turn.Server
	Realm         string
	PublicAddress string
	TTL           int64
}

func NewTurnServer(confKey string, publicIp string) *TurnServer {
	conf := config.YamlEnv().Sub(confKey)
	port := conf.GetInt("port")
	realm := conf.GetString("realm")
	ttl := conf.GetInt64("ttl")

	address := fmt.Sprintf(":%d", port)
	conn, err := net.ListenPacket("udp4", address)
	util.AssertError(err)

	if env.IsTest() {
		conn = &StunLogger{conn}
	}

	log.Printf("TURN server start serving on: %s udp\n", address)

	t, err := turn.NewServer(turn.ServerConfig{
		Realm: realm,
		AuthHandler: func(username string, realm string, srcAddr net.Addr) ([]byte, bool) {
			key, err := redis.Bytes(cache.Redis("turn").Do("GET", username))
			if err != nil {
				return nil, false
			}
			return key, true
		},
		// PacketConnConfigs is a list of UDP Listeners and the configuration around them
		PacketConnConfigs: []turn.PacketConnConfig{
			{
				PacketConn: conn,
				RelayAddressGenerator: &turn.RelayAddressGeneratorStatic{
					RelayAddress: net.ParseIP(publicIp), // Claim that we are listening on IP passed by user (This should be your Public IP)
					Address:      "0.0.0.0",             // But actually be listening on every interface
				},
			},
		},
	})
	util.AssertError(err)

	return &TurnServer{
		Server:        *t,
		Realm:         realm,
		PublicAddress: fmt.Sprintf("%s:%d", publicIp, port),
		TTL:           ttl,
	}
}

func (s *TurnServer) CreateAccessPoint(username string) (map[string]any, error) {
	password := crypto.Sha1Hash(util.RandomString(64))
	key := turn.GenerateAuthKey(username, s.Realm, password)
	_, err := cache.Redis("turn").Do("SET", username, key, "EX", s.TTL)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"url":        fmt.Sprintf("turn:%s?transport=udp", s.PublicAddress),
		"username":   username,
		"credential": password,
	}, nil
}

func (s *TurnServer) RouteApi(g *gin.RouterGroup) {
	g.POST("turn", ginex.Wrap(func(c *gin.Context) (any, error) {
		authInfo := ginex.GetAuthInfo(c)
		return s.CreateAccessPoint(fmt.Sprintf("u%d", authInfo.UserId))
	}))
}
