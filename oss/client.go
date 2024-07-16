package oss

import (
	"fmt"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/rickslab/ares/config"
)

var (
	clients = map[string]*oss.Client{}
	mu      = sync.RWMutex{}
)

func initOssClient(name string) (*oss.Client, error) {
	mu.Lock()
	defer mu.Unlock()

	cli, ok := clients[name]
	if ok {
		return cli, nil
	}

	provider, err := oss.NewEnvironmentVariableCredentialsProvider()
	if err != nil {
		return nil, err
	}

	conf := config.YamlEnv().Sub(fmt.Sprintf("oss.%s", name))

	cli, err = oss.New(conf.GetString("endpoint"), "", "",
		oss.SetCredentialsProvider(&provider),
		oss.Timeout(10, 60),
	)
	if err != nil {
		return nil, err
	}

	clients[name] = cli
	return cli, nil
}

func getOssClient(name string) *oss.Client {
	mu.RLock()
	defer mu.RUnlock()

	return clients[name]
}

func Bucket(name string) (*oss.Bucket, error) {
	cli := getOssClient(name)
	if cli == nil {
		var err error
		cli, err = initOssClient(name)
		if err != nil {
			return nil, err
		}
	}
	return cli.Bucket(name)
}
