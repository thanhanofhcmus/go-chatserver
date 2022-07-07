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
)

func init() {
	prometheus.MustRegister(promClientCounter)
	prometheus.MustRegister(collectors.NewBuildInfoCollector())
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
