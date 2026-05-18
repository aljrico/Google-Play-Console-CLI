BINARY := gpc

.PHONY: build test lint clean

build:
	go build -o bin/$(BINARY) .

test:
	go test ./...

lint:
	gofmt -w .
	go vet ./...

clean:
	rm -rf bin dist coverage.out
