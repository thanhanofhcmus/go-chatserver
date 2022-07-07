package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const PROM_SLEEP_TIME = time.Microsecond * 150

var (
	promClientCounter = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "total_client_gauge",
		Help: "The total number of clients currently connecting",
	})

	promRequestCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_request_counter",
		Help: "The total number of requests the server received",
	})

	promRedisRequestCounter = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "total_redis_request_counter",
		Help: "The total number of requests the server received form redis",
	})
)

func init() {
	prometheus.MustRegister(promClientCounter)
	prometheus.MustRegister(promRequestCounter)
	prometheus.MustRegister(collectors.NewBuildInfoCollector())
}

func IncreaseRequestCounter() {
	go func() {
		promRequestCounter.Inc()
	}()
}

func IncreaseRedisRequestCounter() {
	go func() {
		promRedisRequestCounter.Inc()
	}()
}

func StartSendMetric() {

	go func() {
		for {
			c := gClients.Count()
			promClientCounter.Set(float64(c))
			time.Sleep(PROM_SLEEP_TIME)
		}
	}()

	// Expose the registered metrics via HTTP.
	http.Handle("/metrics", promhttp.Handler())

	log.Println("Prepare prometheus complete")
}
