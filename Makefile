# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOLINT=golangci-lint-v2

COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html

.PHONY: help test bench coverage coverage-html lint clean tidy fmt

help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run tests
	$(GOTEST) -shuffle=on -race ./...

bench: ## Run benchmarks
	$(GOTEST) -shuffle=on -bench=. -benchmem ./...

coverage: ## Generate test coverage report
	$(GOTEST) -shuffle=on -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -func=$(COVERAGE_FILE)

coverage-html: coverage ## Generate HTML coverage report and open in browser
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	open $(COVERAGE_HTML)

lint: ## Run golangci-lint
	$(GOLINT) run

fmt: ## Format code
	$(GOLINT) fmt ./...

tidy: ## Tidy dependencies
	$(GOMOD) tidy

clean: ## Clean build artifacts and coverage files
	$(GOCLEAN)
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

check: fmt lint test ## Run formatting (via golangci-lint), vetting (also via golangci-lint), linting, tests, benchmarks, and race detection

deps: ## Install development dependencies
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint@latest
