# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
BINARY_NAME=gnock

# Main package path
MAIN_PACKAGE=./cmd/gnock

install:
	go install $(MAIN_PACKAGE)

# Build the project
all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN_PACKAGE)

# Run tests
test:
	$(GOTEST) -race -v ./...

# Dependencies
deps:
	$(GOGET) -v ./...

# Run golangci-lint
lint:
	golangci-lint run

# Format the code
fmt:
	go fmt ./...

.PHONY: install build test deps lint fmt
