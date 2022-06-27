package es

import (
	"context"
	"encoding/json"
	"strconv"
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

type SearchResult struct {
	TimedOut bool       `json:"timed_out"`
	Hits     SearchHits `json:"hits"`
}

type SearchHits struct {
	Total struct {
		Value    int64  `json:"value"`
		Relation string `json:"relation"`
	} `json:"total"`
	MaxScore float64   `json:"max_score"`
	Hits     []DocType `json:"hits"`
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

func (cli *ElasticSearchClient) Create(index string, id int64, body interface{}) error {
	doc, err := GetDocReader(body)
	if err != nil {
		return err
	}
	resp, err := cli.Client.Create(index, strconv.FormatInt(id, 10), doc)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (cli *ElasticSearchClient) Delete(index string, id int64) error {
	resp, err := cli.Client.Delete(index, strconv.FormatInt(id, 10))
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (cli *ElasticSearchClient) Update(index string, id int64, body interface{}) error {
	doc, err := GetDocReader(body)
	if err != nil {
		return err
	}
	resp, err := cli.Client.Update(index, strconv.FormatInt(id, 10), doc)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (cli *ElasticSearchClient) Search(ctx context.Context, index string, query interface{}, from, size int) (*SearchResult, error) {
	doc, err := GetDocReader(query)
	if err != nil {
		return nil, err
	}

	es := cli.Client
	resp, err := es.Search(
		es.Search.WithContext(ctx),
		es.Search.WithIndex(index),
		es.Search.WithBody(doc),
		es.Search.WithFrom(from),
		es.Search.WithSize(size),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &SearchResult{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
