# Makefile
BINARY="kibana_exporter"
TARGET="build"
VERSION="v7.5.x.1-latest"

# explicitly go mod
export GO111MODULE=on
LDFLAGS=-ldflags "-extldflags '-static' -s -w"

.DEFAULT_GOAL: $(BINARY)

$(BINARY): clean
	mkdir -p ${TARGET}
	go fmt ./...
	go test ./...
	go build -o ${TARGET}/${BINARY}
	chmod +x ${TARGET}/${BINARY}

release: clean
	mkdir -p ${TARGET}/release
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -a -o ${TARGET}/release/${BINARY}-${VERSION}-linux-amd64
	chmod -R +x ${TARGET}/release

docker: clean release
	docker build --build-arg OS=linux --build-arg ARCH=amd64 --build-arg VERSION=${VERSION} -t chamilad/kibana_prometheus_exporter:${VERSION} .

clean:
	go clean
	rm -rf ${TARGET}
