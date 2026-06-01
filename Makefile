BINARY  := peekr
CMD     := ./cmd/peekr
VERSION := $(shell grep 'version = ' $(CMD)/main.go | grep -oP '"[^"]+"' | tr -d '"')

.PHONY: all build install clean run

all: build

build:
	go build -ldflags="-s -w" -o $(BINARY) $(CMD)

install:
	go install $(CMD)

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)
