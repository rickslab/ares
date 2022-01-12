package metrics

import (
	"fmt"
	"os"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/rcrowley/go-metrics"
	"github.com/rickslab/ares/config"
	"github.com/rickslab/ares/util"
)

const (
	reportInterval = 30 * time.Second
)

var (
	client influxdb2.Client
)

func ReportInfluxDBV2(service string) {
	if client != nil {
		return
	}

	v := config.YamlEnv().Sub("influxdb")
	if v == nil {
		return
	}

	client = influxdb2.NewClient(fmt.Sprintf("http://%s", v.GetString("address")), v.GetString("token"))

	host, err := os.Hostname()
	util.AssertError(err)

	tags := map[string]string{
		"host":    host,
		"service": service,
	}

	go watchNumGoroutine()

	t := time.Tick(reportInterval)
	for {
		select {
		case <-t:
			report(metrics.DefaultRegistry, v.GetString("org"), v.GetString("bucket"), tags)
		}
	}
}

func report(r metrics.Registry, org string, bucket string, tags map[string]string) {
	w := client.WriteAPI(org, bucket)

	r.Each(func(name string, i interface{}) {
		measurement, otherTags := getMeasurementAndTags(name)
		p := influxdb2.NewPointWithMeasurement(measurement)
		for k, v := range tags {
			p.AddTag(k, v)
		}
		for k, v := range otherTags {
			p.AddTag(k, v)
		}

		switch metric := i.(type) {
		case metrics.Counter:
			ms := metric.Snapshot()
			p.AddField("count", ms.Count())
		case metrics.Gauge:
			ms := metric.Snapshot()
			p.AddField("gauge", ms.Value())
		case metrics.GaugeFloat64:
			ms := metric.Snapshot()
			p.AddField("gauge", ms.Value())
		case metrics.Histogram:
			ms := metric.Snapshot()
			ps := ms.Percentiles([]float64{0.95, 0.99})
			p.AddField("max", ms.Max())
			p.AddField("mean", ms.Mean())
			p.AddField("min", ms.Min())
			p.AddField("p95", ps[0])
			p.AddField("p99", ps[1])
		case metrics.Meter:
			ms := metric.Snapshot()
			p.AddField("m1", ms.Rate1())
			p.AddField("m5", ms.Rate5())
			p.AddField("m15", ms.Rate15())
			p.AddField("mean", ms.RateMean())
		case metrics.Timer:
			ms := metric.Snapshot()
			ps := ms.Percentiles([]float64{0.95, 0.99})
			p.AddField("max", ms.Max())
			p.AddField("mean", ms.Mean())
			p.AddField("min", ms.Min())
			p.AddField("p95", ps[0])
			p.AddField("p99", ps[1])
			p.AddField("m1", ms.Rate1())
			p.AddField("m5", ms.Rate5())
			p.AddField("m15", ms.Rate15())
		}
		w.WritePoint(p.SetTime(time.Now()))
	})

	w.Flush()
}
