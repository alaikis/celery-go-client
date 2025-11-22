.PHONY: test build clean fmt vet lint examples help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build parameters
BINARY_NAME=celery-client
CMD_DIR=./cmd

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

test: ## Run tests
	$(GOTEST) -v -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	$(GOCMD) tool cover -html=coverage.out

build: ## Build the library
	$(GOBUILD) -v ./...

clean: ## Clean build cache
	$(GOCLEAN)
	rm -f coverage.out

fmt: ## Format code
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOCMD) vet ./...

tidy: ## Tidy go modules
	$(GOMOD) tidy

examples: ## Build example programs
	$(GOBUILD) -o bin/redis_example $(CMD_DIR)/redis_example/main.go
	$(GOBUILD) -o bin/amqp_example $(CMD_DIR)/amqp_example/main.go

all: fmt vet test build ## Run fmt, vet, test and build

.DEFAULT_GOAL := help
