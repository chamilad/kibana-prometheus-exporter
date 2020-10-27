package exporter

import (
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net/http"
	"strings"
	"sync"
)

// A type that implements the prometheus.Collector interface. This will
// be used to register the metrics with Prometheus.
type Exporter struct {
	lock      sync.RWMutex
	collector *KibanaCollector

	// metrics
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

// NewExporter will create a Exporter struct and initialize the metrics
// that will be scraped by Prometheus. It will use the provided Kibana
// details to populate a KibanaCollector struct.
func NewExporter(kUrl, kUname, kPwd, namespace string, skipTls bool) (error, *Exporter) {
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		return errors.New("namespace cannot be empty"), nil
	}

	collector := &KibanaCollector{}
	collector.url = kUrl

	if strings.HasPrefix(kUrl, "https://") {
		if skipTls {
			log.Printf("skipping TLS verification for Kibana URL %s", kUrl)
		}

		tConf := &tls.Config{
			InsecureSkipVerify: skipTls,
		}

		tr := &http.Transport{
			TLSClientConfig: tConf,
		}

		collector.client = &http.Client{
			Transport: tr,
		}
	} else {
		collector.client = &http.Client{}
		if skipTls {
			log.Printf("kibana.skip-tls is enabled for an http URL, ignoring: %s", kUrl)
		}
	}

	if kUname != "" && kPwd != "" {
		log.Printf("using authenticated requests with Kibana")
		creds := fmt.Sprintf("%s:%s", kUname, kPwd)
		encCreds := base64.StdEncoding.EncodeToString([]byte(creds))
		collector.authHeader = fmt.Sprintf("Basic %s", encCreds)
	} else {
		log.Print("Kibana username or password is not provided, assuming unauthenticated communication")
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
				Name:      "os_load_1m",
				Namespace: namespace,
				Help:      "Kibana load average 1m",
			}),
		load5m: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "os_load_5m",
				Namespace: namespace,
				Help:      "Kibana load average 5m",
			}),
		load15m: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "os_load_15m",
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

	return nil, exporter
}

// parseMetrics will set the metrics values using the KibanaMetrics
// struct, converting values to float64 where needed.
func (e *Exporter) parseMetrics(m *KibanaMetrics) error {
	// any value other than "green" is assumed to be less than 1
	statusVal := 0.0
	if strings.ToLower(m.Status.Overall.State) == "green" {
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

// Describe is the Exporter implementing prometheus.Collector
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.status.Desc()
	ch <- e.concurrentConnections.Desc()
	ch <- e.uptime.Desc()
	ch <- e.heapTotal.Desc()
	ch <- e.heapUsed.Desc()
	ch <- e.load1m.Desc()
	ch <- e.load5m.Desc()
	ch <- e.load15m.Desc()
	ch <- e.respTimeAvg.Desc()
	ch <- e.respTimeMax.Desc()
	ch <- e.reqDisconnects.Desc()
	ch <- e.reqTotal.Desc()
}

// Collect is the Exporter implementing prometheus.Collector
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
