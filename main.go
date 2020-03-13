package main

import (
	"flag"
	"github.com/chamilad/kibana-prometheus-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	addr           = flag.String("web.listen-address", ":9684", "The address to listen on for HTTP requests.")
	metricsPath    = flag.String("web.telemetry-path", "/metrics", "The address to listen on for HTTP requests.")
	kibanaUri      = flag.String("kibana.uri", "", "The Kibana API to fetch metrics from")
	kibanaUsername = flag.String("kibana.username", "", "The username to use for Kibana API")
	kibanaPassword = flag.String("kibana.password", "", "The password to use for Kibana API")
	namespace      = "kibana"
)

func main() {
	flag.Parse()
	*kibanaUri = strings.TrimSpace(*kibanaUri)
	*kibanaUsername = strings.TrimSpace(*kibanaUsername)
	*kibanaPassword = strings.TrimSpace(*kibanaPassword)

	if *kibanaUri == "" {
		log.Fatal("required flag -kibana.uri not provided, aborting")
		os.Exit(1)
	}

	log.Printf("using Kibana URL: %s", *kibanaUri)

	err, exporter := exporter.NewExporter(*kibanaUri, *kibanaUsername, *kibanaPassword, namespace)
	if err != nil {
		log.Fatal("error while initializing exporter: ", err)
		os.Exit(1)
	}

	prometheus.MustRegister(exporter)

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

	log.Printf("starting metrics server at %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
