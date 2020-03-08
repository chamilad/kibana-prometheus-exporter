package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

// args
var (
	addr           = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	kibanaUri      = flag.String("kibana.uri", "http://kibana:5601", "The Kibana API to fetch metrics from")
	kibanaUsername = flag.String("kibana.username", "elastic", "The username to use for Kibana API")
	kibanaPassword = flag.String("kibana.password", "", "The password to use for Kibana API")
	namespace      = "kibana"
)

var (
	dumbCounter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name:      "dumb_kibana_count",
			Help:      "Dumb kibana counter",
			Namespace: namespace,
		})
)

// /metrics endpoint
func main() {
	flag.Parse()

	prometheus.MustRegister(dumbCounter)
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}

// fetch metrics
