package mq

import (
	"context"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/nsqio/go-nsq"
	"github.com/rickslab/ares/config"
	"github.com/rickslab/ares/errcode"
	"github.com/rickslab/ares/grpcex"
	"github.com/rickslab/ares/metrics"
	"github.com/rickslab/ares/util"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/status"
)

type Consumer struct {
	address string
	channel string
	conf    *nsq.Config
	clients []*nsq.Consumer
}

func NewConsumer(channel string) *Consumer {
	address := config.YamlEnv().GetString("service.nsqlookupd")

	return &Consumer{
		address: address,
		channel: channel,
		conf:    nsq.NewConfig(),
	}
}

func (c *Consumer) WithMaxInFlight(maxInFlight int) *Consumer {
	c.conf.MaxInFlight = maxInFlight
	return c
}

func (c *Consumer) StartConsumer(topic string, f func(ctx context.Context, m *nsq.Message) error, concurrency int, timeout time.Duration) {
	consumer, err := nsq.NewConsumer(topic, c.channel, c.conf)
	util.AssertError(err)

	consumer.AddConcurrentHandlers(nsq.HandlerFunc(func(m *nsq.Message) (err error) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		ctx = grpcex.NewContext(ctx, c.channel, topic).NewCtx(ctx)
		t0 := time.Unix(0, m.Timestamp)
		ts := time.Now()
		metrics.NewHistogram("mq_latency", "topic", topic, "channel", c.channel).Update(ts.Sub(t0).Nanoseconds())
		latency := metrics.NewHistogram("latency", "topic", topic, "channel", c.channel)

		defer func() {
			log := grpcex.GetLogger(ctx)
			if ret := recover(); ret != nil {
				stack := string(debug.Stack())
				log.WithFields(logrus.Fields{
					"stack": stack,
				}).Error("Recover panic", ret)

				if retErr, ok := ret.(error); ok {
					err = retErr
				} else {
					err = status.Errorf(errcode.ErrRpcPanic, "recover panic: %v", ret)
				}
			}

			dur := time.Since(ts)
			latency.Update(dur.Nanoseconds())

			code, failed := errcode.From(err)
			status := "success"
			if failed {
				status = "failed"
			}
			call := metrics.NewCounter("call", "topic", topic, "channel", c.channel, "status", status, "code", strconv.Itoa(code))
			call.Inc(1)

			fields := logrus.Fields{
				"latency": dur.Seconds() * 1000, // ms
				"code":    code,
			}
			if err != nil {
				log.WithFields(fields).Error("Consume failed", err)
			} else {
				log.WithFields(fields).Info("Consume success")
			}
		}()

		err = f(ctx, m)
		return
	}), concurrency)

	err = consumer.ConnectToNSQLookupd(c.address)
	util.AssertError(err)

	c.clients = append(c.clients, consumer)
}

func (c *Consumer) OnWithTimeout(topic string, f func(ctx context.Context, m *nsq.Message) error, timeout time.Duration) {
	c.StartConsumer(topic, f, 1, timeout)
}

func (c *Consumer) OnWithConcurrency(topic string, f func(ctx context.Context, m *nsq.Message) error, concurrency int) {
	c.StartConsumer(topic, f, concurrency, 8*time.Second)
}

func (c *Consumer) On(topic string, f func(ctx context.Context, m *nsq.Message) error) {
	c.StartConsumer(topic, f, 1, 8*time.Second)
}

func (c *Consumer) Stop() {
	for _, cli := range c.clients {
		cli.Stop()
	}
}
