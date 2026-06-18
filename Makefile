.PHONY: build install test clean

BINARY=owui
PREFIX=$(HOME)/.local

build:
	go build -ldflags "-s -w" -o bin/$(BINARY) ./cmd/owui/

install: build
	install -d $(PREFIX)/bin
	install -m 755 bin/$(BINARY) $(PREFIX)/bin/$(BINARY)

test:
	go test ./...

clean:
	rm -rf bin/