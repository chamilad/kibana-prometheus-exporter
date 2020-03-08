package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"sync"
)

// args
var (
	addr           = flag.String("web.listen-address", ":8080", "The address to listen on for HTTP requests.")
	metricsPath    = flag.String("web.telemetry-path", "/metrics", "The address to listen on for HTTP requests.")
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

type Exporter struct {
	lock sync.RWMutex
	dumb prometheus.Counter
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.dumb.Desc()
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.dumb.Inc()

	ch <- e.dumb
}

// /metrics endpoint
func main() {
	flag.Parse()

	//prometheus.MustRegister(dumbCounter)
	exporter := &Exporter{
		dumb: dumbCounter,
	}

	prometheus.MustRegister(exporter)
	//prometheus.MustRegister(prometheus.NewBuildInfoCollector())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Kibana Exporter</title></head>
             <body>
             <h1>Kibana Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	http.Handle(*metricsPath, promhttp.Handler())
	log.Printf("starting metrics server at %s,", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

// fetch metrics
