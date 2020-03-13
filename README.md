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

### Docker 
The Docker Image `chamilad/kibana-prometheus-exporter` can be used directly to run the exporter in a Dockerized environment. The Container filesystem only contains the statically linked binary, so that it can be run independently. 

```bash
docker run -p 9684:9684 -it chamilad/kibana-prometheus-exporter:v7.5.x.-latest -kibana.username elastic -kibana.password password -kibana.uri https://elasticcloud.kibana.aws.found.io
```

Refer to the [Makefile](Makefile) and the [Dockerfile](Dockerfile) for more details.

### Kubernetes
Refer the artifacts in [`k8s`](k8s) directory. There is a Deployment and a Service that exposes the Deployment. 

```bash
kubectl apply -f k8s/kibana-prometheus-exporter.yaml
```

```bash
$  kubectl get all -l app=kibana-prometheus-exporter
  NAME                                             READY   STATUS    RESTARTS   AGE
  pod/kibana-prometheus-exporter-b8c888bcd-66kvx   1/1     Running   0          16s
  
  NAME                                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
  service/kibana-prometheus-exporter   ClusterIP   10.96.252.18   <none>        9684/TCP   16s
  
  NAME                                         READY   UP-TO-DATE   AVAILABLE   AGE
  deployment.apps/kibana-prometheus-exporter   1/1     1            1           16s
  
  NAME                                                   DESIRED   CURRENT   READY   AGE
  replicaset.apps/kibana-prometheus-exporter-b8c888bcd   1         1         1       16s
```

With these artifacts deployed, the following Prometheus scrape configuration can be used to scrape the metrics.

```yaml
    - job_name: "kibana"
      scrape_interval: 1m
      metrics_path: "/metrics"
      kubernetes_sd_configs:
        - role: service
      relabel_configs:
        - source_labels: [__meta_kubernetes_service_label_app]
          regex: "kibana-exporter"
          action: keep
        - source_labels: [__meta_kubernetes_namespace]
          action: replace
          target_label: kubernetes_namespace
        - source_labels: [__address__, __meta_kubernetes_service_annotation_prometheus_io_port]
          target_label: __address__
          regex: ([^:]+)(?::\d+)?;(\d+)
          replacement: $1:$2
```

##### Things to note about the Prometheus Scrape Config
 
1. The `scrape_interval` for the job is kept to once per minute. This is to reduce the load on the ElasticSearch cluster, by frequent API calls.
2. The port to connect is detected through a K8s Service annotation, `prometheus.io/port`. 
3. The metrics will end up with the label `job: kibana`

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

## TODO
1. Test edge cases more
2. Add more metrics related to the scrape job itself
3. Add a Grafana dashboards with (Prometheus) alerts 
4. Add mTLS to the metrics server

## Contributing
More metrics, useful tweaks, samples, bug fixes, and any other form of contributions are welcome. Please fork, modify, and open a PR. Please open a GitHub issue for observed bugs or feature requests. I will try to attend to them when possible.

## License
The contents of this repository are licensed under Apache V2 License. 

