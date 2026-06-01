BINARY  := peekr
VERSION := $(shell grep 'version = ' main.go | grep -oP '"[^"]+"' | tr -d '"')

.PHONY: all build install clean run

all: build

build:
	go build -ldflags="-s -w" -o $(BINARY) .

install:
	go install .

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY)
