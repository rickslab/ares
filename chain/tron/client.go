package tron

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

type TronClient struct {
	apiUrl   string
	feeLimit uint64
}

var (
	httpClient = &http.Client{
		Timeout: 8 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Proxy: http.ProxyFromEnvironment,
		},
	}
)

func Client(apiUrl string, feeLimit uint64) *TronClient {
	return &TronClient{
		apiUrl:   apiUrl,
		feeLimit: feeLimit,
	}
}

func (cli *TronClient) httpPost(path string, in interface{}, out interface{}) error {
	url := cli.apiUrl + path

	req, err := json.Marshal(in)
	if err != nil {
		return err
	}

	r, err := httpClient.Post(url, "application/json", bytes.NewReader(req))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(out)
	if err != nil {
		logrus.Errorf("Tron api failed: url='POST%s' req='%s'", url, req)
		return err
	}
	return nil
}

func (cli *TronClient) httpGet(path string, in *url.Values, out interface{}) error {
	url := cli.apiUrl + path
	req := in.Encode()

	r, err := httpClient.Get(fmt.Sprintf("%s?%s", url, req))
	if err != nil {
		return err
	}
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(out)
	if err != nil {
		logrus.Errorf("Tron api failed: url='GET%s' req='%s'", url, req)
		return err
	}
	return nil
}
