NAME ?= kt
VERSION ?= $(shell git describe --tags || echo "unknown")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse HEAD)
GO_LDFLAGS='-X "github.com/knight42/kt/pkg/version.Version=$(VERSION)" \
		-X "github.com/knight42/kt/pkg/version.BuildDate=$(BUILD_DATE)" \
		-X "github.com/knight42/kt/pkg/version.GitCommit=$(GIT_COMMIT)" \
		-w -s'

build:
	CGO_ENABLED=0 go build -trimpath -ldflags $(GO_LDFLAGS) -o $(NAME)

install:
	CGO_ENABLED=0 go install -trimpath -ldflags $(GO_LDFLAGS)

clean:
	rm -f $(NAME)

.PHONY: build install clean
