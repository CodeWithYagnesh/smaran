BINARY  := didfix
PREFIX  := /usr/local/bin
GOFLAGS := -ldflags="-s -w"

.PHONY: all build install uninstall clean test

all: build

build:
	go build $(GOFLAGS) -o $(BINARY) .

install: build
	install -m 0755 $(BINARY) $(PREFIX)/$(BINARY)

uninstall:
	rm -f $(PREFIX)/$(BINARY)

clean:
	rm -f $(BINARY)

test:
	go test -v ./...
