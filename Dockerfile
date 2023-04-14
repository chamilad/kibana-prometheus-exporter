FROM quay.io/prometheus/busybox:latest
LABEL maintainer="Chamila de Alwis <me@chamila.dev>"

ARG BINARY="kibana_exporter-latest-linux-amd64"

COPY build/release/${BINARY} /bin/kibana_exporter

ENTRYPOINT ["/bin/kibana_exporter"]
EXPOSE     9684
