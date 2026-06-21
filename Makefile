BINARY := zmk-keymap-spacer
PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin

.PHONY: build install uninstall test

build:
	go build -o $(BINARY) .

install:
	go install .

uninstall:
	rm -f $(BINDIR)/$(BINARY)

test:
	go test ./...
