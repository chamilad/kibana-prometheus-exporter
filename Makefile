# Makefile
BINARY="kibana_exporter"
IMG="kibana-prometheus-exporter"
TARGET="build"
VERSION="v8.5.x.1-latest"

# explicitly go mod
export GO111MODULE=on
LDFLAGS=-ldflags "-extldflags '-static' -s -w"

.DEFAULT_GOAL: $(BINARY)

$(BINARY): clean
	mkdir -p ${TARGET}
	go fmt ./...
	go test -v ./...
	go build -o ${TARGET}/${BINARY}
	chmod +x ${TARGET}/${BINARY}

release: clean
	mkdir -p ${TARGET}/release
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -a -o ${TARGET}/release/${BINARY}-${VERSION}-linux-amd64
	chmod -R +x ${TARGET}/release

test: $(BINARY)
	test -d tmp
	cp ${TARGET}/${BINARY} tmp/

docker: clean release
	docker build --build-arg OS=linux --build-arg ARCH=amd64 --build-arg VERSION=${VERSION} -t chamilad/${IMG}:${VERSION} .

docker-release: clean release docker
	docker push chamilad/${IMG}:${VERSION}

clean:
	go clean
	rm -rf ${TARGET}
