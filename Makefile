BINARY := gpc
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
INSTALL ?= install
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell git log -1 --format=%cI 2>/dev/null || date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X github.com/aljrico/Google-Play-Console-CLI/cmd.version=$(VERSION) \
	-X github.com/aljrico/Google-Play-Console-CLI/cmd.commit=$(COMMIT) \
	-X github.com/aljrico/Google-Play-Console-CLI/cmd.date=$(DATE)

.PHONY: build test lint release-check snapshot install clean

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY) .

test:
	go test ./...

lint:
	gofmt -w .
	go vet ./...

release-check:
	@command -v goreleaser >/dev/null || { echo "goreleaser is required; install it from https://goreleaser.com/install" >&2; exit 127; }
	goreleaser check

snapshot:
	@command -v goreleaser >/dev/null || { echo "goreleaser is required; install it from https://goreleaser.com/install" >&2; exit 127; }
	goreleaser release --snapshot --clean --skip=publish

install: build
	$(INSTALL) -d $(DESTDIR)$(BINDIR)
	$(INSTALL) -m 0755 bin/$(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)

clean:
	rm -rf bin dist coverage.out
