GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

all: build

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
	  -o "_output/bin/${GOOS}-${GOARCH}/sst"

.PHONY: clean
clean:
	rm -rf _tmp _output