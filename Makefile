NAME ?= kt
BINDIR = bin
VERSION ?= $(shell git describe --tags || echo "unknown")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse HEAD)
GOBUILD=CGO_ENABLED=0 go build -ldflags '-X "github.com/knight42/kt/pkg/version.Version=$(VERSION)" \
		-X "github.com/knight42/kt/pkg/version.BuildDate=$(BUILD_DATE)" \
		-X "github.com/knight42/kt/pkg/version.GitCommit=$(GIT_COMMIT)" \
		-w -s'

PLATFORM_LIST = \
	darwin-amd64 \
	linux-amd64

WINDOWS_ARCH_LIST = \
	windows-amd64

all: linux-amd64 darwin-amd64 windows-amd64 # Most used

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(NAME)-$(VERSION)-$@/$(NAME)

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(NAME)-$(VERSION)-$@/$(NAME)

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(NAME)-$(VERSION)-$@/$(NAME).exe

all-arch: $(PLATFORM_LIST) $(WINDOWS_ARCH_LIST)

gz_releases=$(addsuffix .tar.gz, $(PLATFORM_LIST))
zip_releases=$(addsuffix .zip, $(WINDOWS_ARCH_LIST))

$(gz_releases): %.tar.gz : %
	tar czf $(NAME)-$(VERSION)-$@ $(NAME)-$(VERSION)-$(subst .tar.gz,,$@)/

$(zip_releases): %.zip : %
	zip -r $(NAME)-$(VERSION)-$@ $(NAME)-$(VERSION)-$(basename $@)/

releases: $(gz_releases) $(zip_releases)
clean:
	rm -r $(NAME)-*
