.PHONY: help lint fmt vet test build install clean coverage pre-publish

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

lint: ## Run all linters
	golangci-lint run --timeout=5m

lint-fix: ## Run linters with auto-fix
	golangci-lint run --fix --timeout=5m

fmt: ## Format code
	go fmt ./...
	goimports -w . 2>/dev/null || true

vet: ## Run go vet
	go vet ./...

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-short: ## Run tests without race detector
	go test -v ./...

coverage: test ## Show test coverage
	go tool cover -html=coverage.out

build: ## Build the binary
	go build -v -o mcp-cli .

install: ## Install to GOPATH/bin
	go install

clean: ## Clean build artifacts
	rm -f mcp-cli mcp-cli-go coverage.out
	rm -rf dist/

deps: ## Download dependencies
	go mod download
	go mod tidy
	go mod verify

pre-publish: ## Run all pre-publish checks
	@./lint.sh

# CI targets
ci-test: ## Run CI tests (same as GitHub Actions)
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

ci-lint: ## Run CI linting
	golangci-lint run --timeout=5m

ci: ci-lint ci-test build ## Run all CI checks

# Development helpers
watch: ## Watch for changes and run tests
	@command -v entr > /dev/null || (echo "Install entr: brew install entr or apt install entr" && exit 1)
	find . -name '*.go' | entr -c go test ./...

mod-update: ## Update dependencies
	go get -u ./...
	go mod tidy

# Security
security: ## Run security scan
	gosec ./...

# Documentation
docs: ## Generate documentation
	@echo "📚 Documentation locations:"
	@echo "  - README.md"
	@echo "  - docs/"
	@echo "  - GoDoc: https://pkg.go.dev/github.com/LaurieRhodes/mcp-cli-go"

.DEFAULT_GOAL := help
