REPO_PATH := gitlab.ricebook.net/platform/eruapp
REVISION := $(shell git rev-parse HEAD || unknown)
BUILTAT := $(shell date +%Y-%m-%dT%H:%M:%S)
GO_LDFLAGS ?= -s -X $(REPO_PATH)/versioninfo.REVISION=$(REVISION) \
			  -X $(REPO_PATH)/versioninfo.BUILTAT=$(BUILTAT)

deps:
	go get -u -v -d github.com/Sirupsen/logrus
	go get -u -v -d github.com/coreos/etcd/client
	go get -u -v -d github.com/deckarep/golang-set
	go get -u -v -d github.com/gin-gonic/gin

build:
	go build -ldflags "$(GO_LDFLAGS)" -a -tags netgo -installsuffix netgo -o eru-stats

test:
	go tool vet .
	go test -v ./...
