package config

import (
	"sync"

	"github.com/rickslab/ares/env"
	"github.com/rickslab/ares/util"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

var (
	vipers = map[string]*viper.Viper{}
	mu     = sync.RWMutex{}
)

func InitYaml(name string, paths ...string) *viper.Viper {
	mu.Lock()
	defer mu.Unlock()

	v := vipers[name]
	if v != nil {
		return v
	}

	v = viper.New()
	v.SetConfigName(name)
	v.SetConfigType("yaml")
	for _, path := range paths {
		v.AddConfigPath(path)
	}

	err := v.ReadInConfig()
	util.AssertError(err)

	v.WatchConfig()

	vipers[name] = v
	return v
}

func InitRemoteYaml(provider string, endpoint string, name string) *viper.Viper {
	mu.Lock()
	defer mu.Unlock()

	v := vipers[name]
	if v != nil {
		return v
	}

	v = viper.New()
	v.AddRemoteProvider(provider, endpoint, name)
	v.SetConfigType("yaml")

	err := v.ReadRemoteConfig()
	util.AssertError(err)

	err = v.WatchRemoteConfigOnChannel()
	util.AssertError(err)

	vipers[name] = v
	return v
}

func GetViper(name string) *viper.Viper {
	mu.RLock()
	defer mu.RUnlock()

	return vipers[name]
}

func Yaml(name string) *viper.Viper {
	v := GetViper(name)
	if v != nil {
		return v
	}
	return InitYaml(name, env.GetConfPath(), "../../conf")
}

func YamlEnv() *viper.Viper {
	return Yaml(env.GetEnvFlag())
}

func RemoteYaml(name string) *viper.Viper {
	v := GetViper(name)
	if v != nil {
		return v
	}
	return InitRemoteYaml("consul", "http://127.0.0.1:8500", name)
}
