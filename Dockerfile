FROM quay.io/prometheus/busybox:latest
LABEL maintainer="Chamila de Alwis <chamila@apache.org>"

ARG OS="linux"
ARG ARCH="amd64"
ARG VERSION="latest"

COPY build/release/kibana_exporter-${VERSION}-${OS}-${ARCH} /bin/kibana_exporter

ENTRYPOINT ["/bin/kibana_exporter"]
EXPOSE     9684

