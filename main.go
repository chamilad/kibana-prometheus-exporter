package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
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

	stateGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:      "overall_state",
			Help:      "Kibana Overall Status",
			Namespace: namespace,
		})
)

type Exporter struct {
	lock  sync.RWMutex
	dumb  prometheus.Counter
	state prometheus.Gauge
}

type KibanaStatus struct {
	Status struct {
		Overall struct {
			State string `json:"state"`
		} `json:"overall"`
	} `json:"status"`
	Metrics struct {
		ConcurrentConnections int `json:"concurrent_connections"`
		Process               struct {
			UptimeInMillis int `json:"uptime_in_millis"`
			Memory         struct {
				Heap struct {
					TotalInBytes int `json:"total_in_bytes"`
					UsedInBytes  int `json:"used_in_bytes"`
				} `json:"heap"`
			} `json:"memory"`
		} `json:"process"`
		Os struct {
			Load struct {
				Load1M  float64 `json:"1m"`
				Load5M  float64 `json:"5m"`
				Load15m float64 `json:"15m"`
			} `json:"load"`
		} `json:"os"`
		ResponseTimes struct {
			AvgInMillis float64 `json:"avg_in_millis"`
			MaxInMillis float64 `json:"max_in_millis"`
		} `json:"response_times"`
		Requests struct {
			Disconnects int `json:"disconnects"`
			Total       int `json:"total"`
		} `json:"requests"`
	} `json:"metrics"`
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.dumb.Desc()
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.lock.Lock()
	defer e.lock.Unlock()

	creds := fmt.Sprintf("%s:%s", *kibanaUsername, *kibanaPassword)
	encCreds := base64.StdEncoding.EncodeToString([]byte(creds))

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/status?extended", *kibanaUri), nil)
	req.Header.Add("Authorization", fmt.Sprintf("Basic %s", encCreds))
	req.Header.Add("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error while reading Kibana status: %s", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("invalid response from Kibana status: %s", resp.Status)
		return
	}

	respContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("error while reading response from Kibana status: %s", err)
	}

	status := &KibanaStatus{}
	err = json.Unmarshal(respContent, &status)
	if err != nil {
		log.Printf("error while unmarshalling Kibana status: %s\nProblematic content:\n%s", err, respContent)
	}

	log.Printf("State: %s", status.Status.Overall.State)
	stateVal := 0.0
	if status.Status.Overall.State == "green" {
		stateVal = 1
	}

	e.dumb.Inc()
	e.state.Set(stateVal)

	ch <- e.dumb
	ch <- e.state
}

// /metrics endpoint
func main() {
	flag.Parse()

	//prometheus.MustRegister(dumbCounter)
	exporter := &Exporter{
		dumb:  dumbCounter,
		state: stateGauge,
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
