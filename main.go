package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

var (
	addr           = flag.String("web.listen-address", ":8080", "The address to listen on for HTTP requests.")
	metricsPath    = flag.String("web.telemetry-path", "/metrics", "The address to listen on for HTTP requests.")
	kibanaUri      = flag.String("kibana.uri", "http://kibana:5601", "The Kibana API to fetch metrics from")
	kibanaUsername = flag.String("kibana.username", "elastic", "The username to use for Kibana API")
	kibanaPassword = flag.String("kibana.password", "", "The password to use for Kibana API")
	namespace      = "kibana"
)

type KibanaCollector struct {
	url        string
	authHeader string
	client     *http.Client
}

type Exporter struct {
	lock      sync.RWMutex
	collector *KibanaCollector

	status                prometheus.Gauge
	concurrentConnections prometheus.Gauge
	uptime                prometheus.Gauge
	heapTotal             prometheus.Gauge
	heapUsed              prometheus.Gauge
	load1m                prometheus.Gauge
	load5m                prometheus.Gauge
	load15m               prometheus.Gauge
	respTimeAvg           prometheus.Gauge
	respTimeMax           prometheus.Gauge
	reqDisconnects        prometheus.Gauge
	reqTotal              prometheus.Gauge
}

type KibanaMetrics struct {
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
				Load1m  float64 `json:"1m"`
				Load5m  float64 `json:"5m"`
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

func (c *KibanaCollector) scrape() (error, *KibanaMetrics) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/status?extended", c.url), nil)
	if err != nil {
		return errors.New(fmt.Sprintf("could not initialize a request to scrape metrics: %s", err)), nil
	}

	if c.authHeader != "" {
		req.Header.Add("Authorization", c.authHeader)
	}

	req.Header.Add("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("error while reading Kibana status: %s", err)), nil
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("invalid response from Kibana status: %s", resp.Status)), nil

	}

	respContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("error while reading response from Kibana status: %s", err)), nil
	}

	metrics := &KibanaMetrics{}
	err = json.Unmarshal(respContent, &metrics)
	if err != nil {
		return errors.New(fmt.Sprintf("error while unmarshalling Kibana status: %s\nProblematic content:\n%s", err, respContent)), nil
	}

	return nil, metrics
}

func NewExporter(kUrl, kUname, kPwd, namespace string) *Exporter {
	collector := &KibanaCollector{}
	collector.url = kUrl
	collector.client = &http.Client{}

	if kUname != "" && kPwd != "" {
		creds := fmt.Sprintf("%s:%s", *kibanaUsername, *kibanaPassword)
		encCreds := base64.StdEncoding.EncodeToString([]byte(creds))
		collector.authHeader = fmt.Sprintf("Basic %s", encCreds)
	}

	exporter := &Exporter{
		collector: collector,

		status: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "status",
				Help:      "Kibana overall status",
				Namespace: namespace,
			}),
		concurrentConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "concurrent_connections",
				Namespace: namespace,
				Help:      "Kibana Concurrent Connections",
			}),
		uptime: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "millis_uptime",
				Namespace: namespace,
				Help:      "Kibana uptime in milliseconds",
			}),
		heapTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "heap_max_in_bytes",
				Namespace: namespace,
				Help:      "Kibana Heap maximum in bytes",
			}),
		heapUsed: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "heap_used_in_bytes",
				Namespace: namespace,
				Help:      "Kibana Heap usage in bytes",
			}),
		load1m: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "os_load1",
				Namespace: namespace,
				Help:      "Kibana load average 1m",
			}),
		load5m: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "os_load5",
				Namespace: namespace,
				Help:      "Kibana load average 5m",
			}),
		load15m: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "os_load15",
				Namespace: namespace,
				Help:      "Kibana load average 15m",
			}),
		respTimeAvg: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "response_average",
				Namespace: namespace,
				Help:      "Kibana average response time in milliseconds",
			}),
		respTimeMax: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "response_max",
				Namespace: namespace,
				Help:      "Kibana maximum response time in milliseconds",
			}),
		reqDisconnects: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "requests_disconnects",
				Namespace: namespace,
				Help:      "Kibana request disconnections count",
			}),
		reqTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "requests_total",
				Namespace: namespace,
				Help:      "Kibana total request count",
			}),
	}

	return exporter
}

func (e *Exporter) parseMetrics(m *KibanaMetrics) error {
	log.Printf("State: %s", m.Status.Overall.State)
	statusVal := 0.0
	if m.Status.Overall.State == "green" {
		statusVal = 1.0
	}

	e.status.Set(statusVal)

	e.concurrentConnections.Set(float64(m.Metrics.ConcurrentConnections))
	e.uptime.Set(float64(m.Metrics.Process.UptimeInMillis))
	e.heapTotal.Set(float64(m.Metrics.Process.Memory.Heap.TotalInBytes))
	e.heapUsed.Set(float64(m.Metrics.Process.Memory.Heap.UsedInBytes))
	e.load1m.Set(m.Metrics.Os.Load.Load1m)
	e.load5m.Set(m.Metrics.Os.Load.Load5m)
	e.load15m.Set(m.Metrics.Os.Load.Load15m)
	e.respTimeAvg.Set(m.Metrics.ResponseTimes.AvgInMillis)
	e.respTimeMax.Set(m.Metrics.ResponseTimes.MaxInMillis)
	e.reqDisconnects.Set(float64(m.Metrics.Requests.Disconnects))
	e.reqTotal.Set(float64(m.Metrics.Requests.Total))

	return nil
}

func (e *Exporter) send(ch chan<- prometheus.Metric) error {
	ch <- e.status
	ch <- e.concurrentConnections
	ch <- e.uptime
	ch <- e.heapTotal
	ch <- e.heapUsed
	ch <- e.load1m
	ch <- e.load5m
	ch <- e.load15m
	ch <- e.respTimeAvg
	ch <- e.respTimeMax
	ch <- e.reqDisconnects
	ch <- e.reqTotal

	return nil
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.status.Desc()
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.lock.Lock()
	defer e.lock.Unlock()

	err, metrics := e.collector.scrape()
	if err != nil {
		log.Printf("error while scraping metrics from Kibana: %s", err)
		return
	}

	err = e.parseMetrics(metrics)
	if err != nil {
		log.Printf("error while parsing metrics from Kibana: %s", err)
		return
	}

	err = e.send(ch)
	if err != nil {
		log.Printf("error while responding to Prometheus with metrics: %s", err)
	}
}

func main() {
	flag.Parse()

	exporter := NewExporter(*kibanaUri, *kibanaUsername, *kibanaPassword, namespace)
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
	log.Printf("starting metrics server at %s,", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
