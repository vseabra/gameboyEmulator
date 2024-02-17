GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOIMPORTS=goimports-reviser
GOLINTCI=golangci-lint

INSTALLDIR=~/.local/bin
BINDIR=bin
BINARY_NAME=dmgo

.PHONY: all build test clean run lint fmt install uninstall

all: test build
build: lint fmt
	$(GOBUILD) -o $(BINDIR)/$(BINARY_NAME) -v ./cmd/gby

test: 
	$(GOTEST) -v ./...

clean: 
	$(GOCLEAN)
	rm -f $(INSTALLDIR)/$(BINARY_NAME)

run:
	$(GOBUILD) -o $(BINDIR)/$(BINARY_NAME) -v ./cmd/gby
	$(INSTALLDIR)/$(BINARY_NAME)

lint:
	$(GOLINTCI) run ./...

fmt:
	$(GOIMPORTS) -apply-to-generated-files -use-cache ./... 

install: build
	cp $(BINDIR)/$(BINARY_NAME) $(INSTALLDIR)

uninstall:
	rm -f $(INSTALLDIR)/$(BINARY_NAME)

