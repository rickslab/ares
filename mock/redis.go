package mock

import (
	"github.com/alicebob/miniredis/v2"
	"github.com/rickslab/ares/cache"
	"github.com/rickslab/ares/config"
	"github.com/rickslab/ares/util"
)

func InitRedisCli() {
	s, err := miniredis.Run()
	util.AssertError(err)

	cli := cache.NewRedisClient(s.Addr(), "", "")

	redisConf := config.YamlEnv().GetStringMap("redis")
	for name := range redisConf {
		cache.SetRedis(name, cli)
	}
}
