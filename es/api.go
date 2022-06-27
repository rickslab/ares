package es

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/rickslab/ares/errcode"
	"google.golang.org/grpc/status"
)

type Result struct {
	Index   string `json:"_index"`
	Type    string `json:"_type"`
	Id      string `json:"_id"`
	Version int64  `json:"_version"`
	Result  string `json:"result"` // created, deleted
	Error   struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
	} `json:"error"`
	Shards struct {
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
	Error    struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
	} `json:"error"`
	Shards struct {
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
	Source    Object              `json:"_source"`
	Highlight map[string][]string `json:"highlight"`
}

func getId(id interface{}) string {
	switch val := id.(type) {
	case uint64:
		return strconv.FormatUint(val, 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case string:
		return val
	}
	return fmt.Sprintf("%v", id)
}

func (cli *ElasticSearchClient) Create(ctx context.Context, index string, id interface{}, body interface{}) (*Result, error) {
	doc, err := GetObjectReader(body)
	if err != nil {
		return nil, err
	}

	es := cli.Client
	resp, err := es.Create(index, getId(id), doc, es.Create.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &Result{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("es.Create failed[code=%d]: type=%s reason: %s", resp.StatusCode, result.Error.Type, result.Error.Reason)
	}
	return result, nil
}

func (cli *ElasticSearchClient) Delete(ctx context.Context, index string, id interface{}) (*Result, error) {
	es := cli.Client
	resp, err := es.Delete(index, getId(id), es.Delete.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &Result{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if result.Result == "not_found" {
			return nil, status.Error(errcode.ErrElasticNotFound, "elastic delete not found")
		}
		return nil, fmt.Errorf("es.Delete failed[code=%d]: type=%s reason: %s", resp.StatusCode, result.Error.Type, result.Error.Reason)
	}
	return result, nil
}

func (cli *ElasticSearchClient) Update(ctx context.Context, index string, id interface{}, body interface{}) (*Result, error) {
	doc, err := GetObjectReader(body)
	if err != nil {
		return nil, err
	}

	es := cli.Client
	resp, err := es.Update(index, getId(id), doc, es.Update.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &Result{}
	err = json.NewDecoder(resp.Body).Decode(result)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("es.Update failed[code=%d]: type=%s reason: %s", resp.StatusCode, result.Error.Type, result.Error.Reason)
	}
	return result, nil
}

func (cli *ElasticSearchClient) Search(ctx context.Context, index string, query interface{}, from, size int) (*QueryResult, error) {
	doc, err := GetObjectReader(query)
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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("es.Search failed[code=%d]: type=%s reason: %s", resp.StatusCode, result.Error.Type, result.Error.Reason)
	}
	return result, nil
}
