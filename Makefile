BINARY := cfeed

PACKAGE="github.com/BullionBear/crypto-feed"
VERSION := $(shell git describe --tags --always --abbrev=0 --match='v[0-9]*.[0-9]*.[0-9]*' 2> /dev/null)
COMMIT_HASH := $(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP := $(shell date '+%Y-%m-%dT%H:%M:%S')
LDFLAGS := -X '${PACKAGE}/internal/env.Version=${VERSION}' \
           -X '${PACKAGE}/internal/env.CommitHash=${COMMIT_HASH}' \
           -X '${PACKAGE}/internal/env.BuildTime=${BUILD_TIMESTAMP}'

gen:
	protoc --go_out=. --go-grpc_out=. api/proto/feed.proto

build:
	env GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o ./bin/$(BINARY)-linux-x86 cmd/server/*.go
	env GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o ./bin/$(BINARY)-darwin-arm64 cmd/server/*.go
	env GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o ./bin/playback-linux-x86 cmd/playback/*.go
	env GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o ./bin/playback-darwin-arm64 cmd/playback/*.go

clean:
	rm -rf bin/*
	rm -rf api/gen
    
