GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOIMPORTS=goimports # Assuming goimports is installed
GOLINTCI=golangci-lint # Assuming golangci-lint is installed

INSTALLDIR=~/.local/bin
BINDIR=bin
BINARY_NAME=getris

.PHONY: all build test clean run lint fmt install uninstall

all: test build
build: lint fmt
	$(GOBUILD) -o $(BINDIR)/$(BINARY_NAME) -v

test: 
	$(GOTEST) -v ./...

clean: 
	$(GOCLEAN)
	rm -f $(INSTALLDIR)/$(BINARY_NAME)

run:
	$(GOBUILD) -o $(INSTALLDIR)/$(BINARY_NAME) -v .
	$(INSTALLDIR)/$(BINARY_NAME)

lint:
	$(GOLINTCI) run ./...

fmt:
	$(GOIMPORTS) -w .

install: build
	cp $(BINDIR)/$(BINARY_NAME) $(INSTALLDIR)

uninstall:
	rm -f $(INSTALLDIR)/$(BINARY_NAME)

