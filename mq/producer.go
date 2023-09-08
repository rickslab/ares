package mq

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/nsqio/go-nsq"
	"github.com/rickslab/ares/config"
	"github.com/rickslab/ares/util"
)

var (
	producerCli  *nsq.Producer
	producerOnce = sync.Once{}
)

func getProducer() *nsq.Producer {
	producerOnce.Do(func() {
		address := config.YamlEnv().GetString("service.nsqd")

		var err error
		producerCli, err = nsq.NewProducer(address, nsq.NewConfig())
		util.AssertError(err)
	})
	return producerCli
}

func Publish(topic string, body []byte) error {
	return getProducer().Publish(topic, body)
}

func DeferredPublish(topic string, delay time.Duration, body []byte) error {
	return getProducer().DeferredPublish(topic, delay, body)
}

func PublishJSON(topic string, value any) error {
	data, err := json.Marshal(value)
	util.AssertError(err)

	return Publish(topic, data)
}

func DeferredPublishJSON(topic string, delay time.Duration, value any) error {
	data, err := json.Marshal(value)
	util.AssertError(err)

	return DeferredPublish(topic, delay, data)
}

func PublishProto(topic string, value proto.Message) error {
	data, err := proto.Marshal(value)
	util.AssertError(err)

	return Publish(topic, data)
}

func DeferredPublishProto(topic string, delay time.Duration, value proto.Message) error {
	data, err := proto.Marshal(value)
	util.AssertError(err)

	return DeferredPublish(topic, delay, data)
}
