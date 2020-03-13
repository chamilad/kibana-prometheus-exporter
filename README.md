# Prometheus Exporter for Kibana (7.5.*)

A standalone Prometheus exporter for Kibana metrics inspired by the [Kibana Prometheus Exporter Plugin](https://github.com/pjhampton/kibana-prometheus-exporter/). 

This makes use of the `/api/status` endpoint to gather and convert metrics to the Prometheus OpenMetrics format.

## Usage

> **NOTE**: Currently only tested against Kibana 7.5 versions. 

```bash
# expose metrics from the local Kibana instance using the provided username and password
kibana-exporter -kibana.uri http://localhost:5601 -kibana.username elastic -kibana.password password
```

By default, the Exporter exposes the `/metrics` endpoint at port `9684`. If needed this port (and the endpoint) can be overridden.

```bash
# expose the /metrics endpoint at port 8080
kibana-exporter -kibana.uri http://localhost:5601 -web.listen-address 8080 
```

```bash
# expose metrics using /scrape endpint
kibana-exporter -kibana.uri http://localhost:5601 -web.telemetry-path "/scrape"
```

### Docker Container
The Docker Image `chamilad/kibana-prometheus-exporter` can be used directly to run the exporter in a Dockerized environment. The Container filesystem only contains the statically linked binary, so that it can be run independently. 

```bash
docker run -p 9684:9684 -it chamilad/kibana-prometheus-exporter:v7.5.x.1 -kibana.username elastic -kibana.password password -kibana.uri https://elasticcloud.kibana.aws.found.io
```

Refer to the [Makefile](Makefile) and the [Dockerfile](Dockerfile) for more details.

### Kubernetes
The K8s sample will soon be added.

## Metrics
The metrics exposed by this Exporter are the following.

| Metric | Description | Type |
|------- | ----------- | ---- |
| `kibana_status` | Kibana overall status | Gauge |
| `kibana_concurrent_connections` | Kibana Concurrent Connections | Gauge |
| `kibana_millis_uptime` | Kibana uptime in milliseconds | Gauge |
| `kibana_heap_max_in_bytes` | Kibana Heap maximum in bytes | Gauge |
| `kibana_heap_used_in_bytes` | Kibana Heap usage in bytes | Gauge |
| `kibana_os_load_1m` | Kibana load average 1m | Gauge |
| `kibana_os_load_5m` | Kibana load average 5m | Gauge |
| `kibana_os_load_15m` | Kibana load average 15m | Gauge |
| `kibana_response_average` | Kibana average response time in milliseconds | Gauge |
| `kibana_response_max` | Kibana maximum response time in milliseconds | Gauge |
| `kibana_requests_disconnects` | Kibana request disconnections count | Gauge |
| `kibana_requests_total` | Kibana total request count | Gauge |

## Contributing
More metrics, useful tweaks, samples, bug fixes, and any other form of contributions are welcome. Please fork, modify, and open a PR. Please open a GitHub issue for observed bugs or feature requests. I will try to attend to them when possible.

## License
The contents of this repository are licensed under Apache V2 License. 

