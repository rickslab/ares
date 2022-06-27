package es

import (
	"context"
	"encoding/json"
	"strconv"
)

type Result struct {
	Index   string `json:"_index"`
	Type    string `json:"_type"`
	Id      string `json:"_id"`
	Version int64  `json:"_version"`
	Result  string `json:"result"` // created, deleted
	Shards  struct {
		Total      int64 `json:"total"`
		Successful int64 `json:"successful"`
		Failed     int64 `json:"failed"`
	} `json:"_shards"`
	SeqNo       int64 `json:"_seq_no"`
	PrimaryTerm int64 `json:"_primary_term"`
}

type QueryResult struct {
	Took     int64 `json:"took"`
	TimedOut bool  `json:"timed_out"`
	Shards   struct {
		Total      int64 `json:"total"`
		Successful int64 `json:"successful"`
		Failed     int64 `json:"failed"`
		Skipped    int64 `json:"skipped"`
	} `json:"_shards"`
	Hits QueryHits `json:"hits"`
}

type QueryHits struct {
	Total struct {
		Value    int64  `json:"value"`
		Relation string `json:"relation"`
	} `json:"total"`
	MaxScore float64    `json:"max_score"`
	Rows     []QueryRow `json:"hits"`
}

type QueryRow struct {
	Index     string              `json:"_index"`
	Type      string              `json:"_type"`
	Id        string              `json:"_id"`
	Score     float64             `json:"_score"`
	Source    Doc                 `json:"_source"`
	Highlight map[string][]string `json:"highlight"`
}

func (cli *ElasticSearchClient) Create(index string, id int64, body interface{}) (*Result, error) {
	doc, err := GetDocReader(body)
	if err != nil {
		return nil, err
	}
	resp, err := cli.Client.Create(index, strconv.FormatInt(id, 10), doc)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &Result{}
	err = json.NewDecoder(resp.Body).Decode(result)
	return result, err
}

func (cli *ElasticSearchClient) Delete(index string, id int64) (*Result, error) {
	resp, err := cli.Client.Delete(index, strconv.FormatInt(id, 10))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &Result{}
	err = json.NewDecoder(resp.Body).Decode(result)
	return result, err
}

func (cli *ElasticSearchClient) Update(index string, id int64, body interface{}) (*Result, error) {
	doc, err := GetDocReader(body)
	if err != nil {
		return nil, err
	}
	resp, err := cli.Client.Update(index, strconv.FormatInt(id, 10), doc)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &Result{}
	err = json.NewDecoder(resp.Body).Decode(result)
	return result, err
}

func (cli *ElasticSearchClient) Search(ctx context.Context, index string, query interface{}, from, size int) (*QueryResult, error) {
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

	result := &QueryResult{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
