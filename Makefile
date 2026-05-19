BINARY := gpc
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
INSTALL ?= install

.PHONY: build test lint install clean

build:
	go build -o bin/$(BINARY) .

test:
	go test ./...

lint:
	gofmt -w .
	go vet ./...

install: build
	$(INSTALL) -d $(DESTDIR)$(BINDIR)
	$(INSTALL) -m 0755 bin/$(BINARY) $(DESTDIR)$(BINDIR)/$(BINARY)

clean:
	rm -rf bin dist coverage.out
