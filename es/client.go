package es

import (
	"sync"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/rickslab/ares/config"
	"github.com/rickslab/ares/util"
)

var (
	client *ElasticSearchClient
	once   = sync.Once{}
)

type ElasticSearchClient struct {
	*elasticsearch.Client
}

func Client() *ElasticSearchClient {
	once.Do(func() {
		conf := config.YamlEnv().Sub("elasticsearch")
		cfg := elasticsearch.Config{
			Addresses: []string{
				conf.GetString("host"),
			},
			Username: conf.GetString("user"),
			Password: conf.GetString("password"),
		}

		cli, err := elasticsearch.NewClient(cfg)
		util.AssertError(err)

		client = &ElasticSearchClient{
			Client: cli,
		}
	})
	return client
}
