package main

import (
	"flag"
	"net/http"
	"strings"

	"github.com/chamilad/kibana-prometheus-exporter/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	addr           = flag.String("web.listen-address", ":9684", "The address to listen on for HTTP requests.")
	metricsPath    = flag.String("web.telemetry-path", "/metrics", "The address to listen on for HTTP requests.")
	kibanaURI      = flag.String("kibana.uri", "", "The Kibana API to fetch metrics from")
	kibanaUsername = flag.String("kibana.username", "", "The username to use for Kibana API")
	kibanaPassword = flag.String("kibana.password", "", "The password to use for Kibana API")
	kibanaSkipTLS  = flag.Bool("kibana.skip-tls", false, "Skip TLS verification for TLS secured Kibana URLs")
	debug          = flag.Bool("debug", false, "Output verbose details during metrics collection, use for development only")
	wait           = flag.Bool(
		"wait",
		false,
		"Wait for Kibana to be responsive before starting, setting this to false would cause the exporter to error out instead of waiting",
	)
	namespace = "kibana"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

	flag.Parse()
	*kibanaURI = strings.TrimSpace(*kibanaURI)
	*kibanaUsername = strings.TrimSpace(*kibanaUsername)
	*kibanaPassword = strings.TrimSpace(*kibanaPassword)

	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	if *kibanaURI == "" {
		log.Fatal().
			Msg("required flag -kibana.uri not provided, aborting")
	}

	*kibanaURI = strings.TrimSuffix(*kibanaURI, "/")
	log.Printf("using Kibana URL: %s", *kibanaURI)

	collector, err := exporter.NewCollector(*kibanaURI, *kibanaUsername, *kibanaPassword, *kibanaSkipTLS)
	if err != nil {
		log.Fatal().
			Msgf("error while initializing collector: %s", err)
	}

	exporter, err := exporter.NewExporter(namespace, collector)
	if err != nil {
		log.Fatal().
			Msgf("error while initializing exporter: %s", err)
	}

	if *wait {
		// blocking wait for Kibana to be responsive
		collector.WaitForConnection()
	} else {
		if !collector.TestConnection() {
			log.Fatal().
				Msg("not waiting for Kibana to be responsive")
		}
	}

	prometheus.MustRegister(exporter)

	// readable output
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err = w.Write([]byte(`<html>
             <head><title>Kibana Exporter</title></head>
             <body>
             <h1>Kibana Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
		log.Warn().
			Msgf("error while writing response to /metrics call: %s", err)
	})

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
			return
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})

	http.Handle(*metricsPath, promhttp.Handler())

	log.Info().
		Msgf("starting metrics server at %s", *addr)
	log.Fatal().
		Msgf("%s", http.ListenAndServe(*addr, nil))
}
