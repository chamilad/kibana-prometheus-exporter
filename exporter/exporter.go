package exporter

import (
	"errors"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

var (
	// https://github.com/elastic/kibana/blob/12466d8b17d8557ff0b561c346511bd1760da4c1/packages/core/status/core-status-common/src/service_status.ts
	statusLevels = map[string]float64{
		// everything is working
		"available": 1,
		// some features may not be working
		"degraded": 0.5,
		// the service is unavailble, but other functions that do not depend on this service should work.
		"unavailable": 0.25,
		// block all user functions and display the status page, reserved for Core services only.
		"critical": 0,
	}
)

// Exporter implements the prometheus.Collector interface. This will
// be used to register the metrics with Prometheus.
type Exporter struct {
	lock      sync.RWMutex
	collector *KibanaCollector

	// metrics
	status                prometheus.Gauge
	coreESStatus          prometheus.Gauge
	coreSOStatus          prometheus.Gauge
	concurrentConnections prometheus.Gauge
	uptime                prometheus.Gauge
	heapTotal             prometheus.Gauge
	heapUsed              prometheus.Gauge
	resSetSize            prometheus.Gauge
	eventLoopDelay        prometheus.Gauge
	load1m                prometheus.Gauge
	load5m                prometheus.Gauge
	load15m               prometheus.Gauge
	osMemTotal            prometheus.Gauge
	osMemUsed             prometheus.Gauge
	respTimeAvg           prometheus.Gauge
	respTimeMax           prometheus.Gauge
	reqDisconnects        prometheus.Gauge
	reqTotal              prometheus.Gauge
}

// NewExporter will create a Exporter struct and initialize the metrics
// that will be scraped by Prometheus. It will use the provided Kibana
// details to populate a KibanaCollector struct.
func NewExporter(namespace string, collector *KibanaCollector) (*Exporter, error) {
	namespace = strings.TrimSpace(namespace)
	if namespace == "" {
		return nil, errors.New("namespace cannot be empty")
	}

	exporter := &Exporter{
		collector: collector,

		status: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "status",
				Help:      "Kibana overall status",
				Namespace: namespace,
			}),
		coreESStatus: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "core_es_status",
				Help:      "Kibana Elasticsearch connectivity status",
				Namespace: namespace,
			}),
		coreSOStatus: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "core_savedobjects_status",
				Help:      "Kibana SavedObjects service status",
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
				Help:      "Kibana process Heap maximum in bytes",
			}),
		heapUsed: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "heap_used_in_bytes",
				Namespace: namespace,
				Help:      "Kibana process Heap usage in bytes",
			}),
		resSetSize: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "resident_set_size_in_bytes",
				Namespace: namespace,
				Help:      "Kibana Memory Resident Set Size in bytes",
			}),
		eventLoopDelay: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "event_loop_delay",
				Namespace: namespace,
				Help:      "Kibana NodeJS Event Loop Delay in milliseconds",
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
		osMemTotal: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "os_memory_max_in_bytes",
				Namespace: namespace,
				Help:      "Kibana memory maximum in bytes",
			}),
		osMemUsed: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name:      "os_memory_used_in_bytes",
				Namespace: namespace,
				Help:      "Kibana memory used in bytes",
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

	return exporter, nil
}

// parseMetrics will set the metrics values using the KibanaMetrics
// struct, converting values to float64 where needed.
func (e *Exporter) parseMetrics(m *KibanaMetrics) error {
	log.Trace().
		Msg("parsing received metrics from kibana")

	if val, ok := statusLevels[strings.ToLower(m.Status.Overall.Level)]; !ok {
		// absence of this metric will default to critical
		// initialising is also considered 0
		e.status.Set(0.0)
	} else {
		e.status.Set(val)
	}

	if val, ok := statusLevels[strings.ToLower(m.Status.Core.Elasticsearch.Level)]; !ok {
		// absence of this metric will default to critical
		// initialising is also considered 0
		e.coreESStatus.Set(0.0)
	} else {
		e.coreESStatus.Set(val)
	}

	if val, ok := statusLevels[strings.ToLower(m.Status.Core.SavedObjects.Level)]; !ok {
		// absence of this metric will default to critical
		// initialising is also considered 0
		e.coreSOStatus.Set(0.0)
	} else {
		e.coreSOStatus.Set(val)
	}

	e.concurrentConnections.Set(float64(m.Metrics.ConcurrentConnections))
	e.uptime.Set(float64(m.Metrics.Process.UptimeInMillis))
	e.heapTotal.Set(float64(m.Metrics.Process.Memory.Heap.TotalInBytes))
	e.heapUsed.Set(float64(m.Metrics.Process.Memory.Heap.UsedInBytes))
	e.resSetSize.Set(float64(m.Metrics.Process.Memory.ResidentSetSizeInBytes))
	e.eventLoopDelay.Set(m.Metrics.Process.EventLoopDelayInMillis)
	e.load1m.Set(m.Metrics.Os.Load.Load1m)
	e.load5m.Set(m.Metrics.Os.Load.Load5m)
	e.load15m.Set(m.Metrics.Os.Load.Load15m)
	e.osMemTotal.Set(float64(m.Metrics.Os.Memory.TotalInBytes))
	e.osMemUsed.Set(float64(m.Metrics.Os.Memory.UsedInBytes))
	e.respTimeAvg.Set(m.Metrics.ResponseTimes.AvgInMillis)
	e.respTimeMax.Set(m.Metrics.ResponseTimes.MaxInMillis)
	e.reqDisconnects.Set(float64(m.Metrics.Requests.Disconnects))
	e.reqTotal.Set(float64(m.Metrics.Requests.Total))

	return nil
}

func (e *Exporter) send(ch chan<- prometheus.Metric) error {
	ch <- e.status
	ch <- e.coreESStatus
	ch <- e.coreSOStatus
	ch <- e.concurrentConnections
	ch <- e.uptime
	ch <- e.heapTotal
	ch <- e.heapUsed
	ch <- e.resSetSize
	ch <- e.eventLoopDelay
	ch <- e.load1m
	ch <- e.load5m
	ch <- e.load15m
	ch <- e.osMemTotal
	ch <- e.osMemUsed
	ch <- e.respTimeAvg
	ch <- e.respTimeMax
	ch <- e.reqDisconnects
	ch <- e.reqTotal

	return nil
}

// Describe is the Exporter implementing prometheus.Collector
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.status.Desc()
	ch <- e.coreESStatus.Desc()
	ch <- e.coreSOStatus.Desc()
	ch <- e.concurrentConnections.Desc()
	ch <- e.uptime.Desc()
	ch <- e.heapTotal.Desc()
	ch <- e.heapUsed.Desc()
	ch <- e.resSetSize.Desc()
	ch <- e.eventLoopDelay.Desc()
	ch <- e.load1m.Desc()
	ch <- e.load5m.Desc()
	ch <- e.load15m.Desc()
	ch <- e.osMemTotal.Desc()
	ch <- e.osMemUsed.Desc()
	ch <- e.respTimeAvg.Desc()
	ch <- e.respTimeMax.Desc()
	ch <- e.reqDisconnects.Desc()
	ch <- e.reqTotal.Desc()
}

// Collect is the Exporter implementing prometheus.Collector
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	log.Trace().
		Msg("a Collect() call received")

	e.lock.Lock()
	defer e.lock.Unlock()

	log.Trace().
		Msg("issueing a scrape() call to the collector")

	metrics, err := e.collector.scrape()
	if err != nil {
		log.Error().
			Msgf("error while scraping metrics from Kibana: %s", err)
		return
	}

	// output for debugging
	log.Debug().
		Interface("metrics", metrics).
		Msg("returned metrics content")

	err = e.parseMetrics(metrics)
	if err != nil {
		log.Error().
			Msgf("error while parsing metrics from Kibana: %s", err)
		return
	}

	err = e.send(ch)
	if err != nil {
		log.Error().
			Msgf("error while responding to Prometheus with metrics: %s", err)
	}
}
