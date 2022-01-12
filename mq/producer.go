package mq

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/nsqio/go-nsq"
	"github.com/rickslab/ares/config"
	"github.com/rickslab/ares/util"
	"github.com/sirupsen/logrus"
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

func Publish(topic string, body []byte) {
	err := getProducer().Publish(topic, body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": topic,
		}).Error("Publish failed", err)
	}
}

func DeferredPublish(topic string, delay time.Duration, body []byte) {
	err := getProducer().DeferredPublish(topic, delay, body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"topic": topic,
			"delay": delay.Seconds(),
		}).Error("DeferredPublish failed", err)
	}
}

func PublishJSON(topic string, value interface{}) {
	data, err := json.Marshal(value)
	util.AssertError(err)

	Publish(topic, data)
}

func DeferredPublishJSON(topic string, delay time.Duration, value interface{}) {
	data, err := json.Marshal(value)
	util.AssertError(err)

	DeferredPublish(topic, delay, data)
}

func PublishProto(topic string, value proto.Message) {
	data, err := proto.Marshal(value)
	util.AssertError(err)

	Publish(topic, data)
}

func DeferredPublishProto(topic string, delay time.Duration, value proto.Message) {
	data, err := proto.Marshal(value)
	util.AssertError(err)

	DeferredPublish(topic, delay, data)
}
