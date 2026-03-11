.PHONY: dev build test lint clean install-deps setup help

# Default target
.DEFAULT_GOAL := help

# Variables
GO := go
NODE := node
NPM := npm
WAILS := $(HOME)/go/bin/wails
FRONTEND_DIR := frontend

# Colors for output
CYAN := \033[36m
GREEN := \033[32m
YELLOW := \033[33m
RESET := \033[0m

## ============================================================================
## Development
## ============================================================================

dev: ## Start development server with hot reload
	@if command -v wails >/dev/null 2>&1; then \
		wails dev; \
	else \
		$(WAILS) dev; \
	fi

build: ## Build production binary
	@if command -v wails >/dev/null 2>&1; then \
		wails build; \
	else \
		$(WAILS) build; \
	fi

build-all: ## Build for all platforms
	@echo "$(CYAN)Building for all platforms...$(RESET)"
	@if command -v wails >/dev/null 2>&1; then \
		wails build -platform darwin/amd64,darwin/arm64,windows/amd64,linux/amd64; \
	else \
		$(WAILS) build -platform darwin/amd64,darwin/arm64,windows/amd64,linux/amd64; \
	fi

## ============================================================================
## Testing & Quality
## ============================================================================

test: ## Run Go tests
	@echo "$(CYAN)Running Go tests...$(RESET)"
	$(GO) test -v -race ./...

test-cover: ## Run tests with coverage
	@echo "$(CYAN)Running tests with coverage...$(RESET)"
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(RESET)"

lint: ## Run linters
	@echo "$(CYAN)Running linters...$(RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		$(GO) vet ./...; \
	fi
	@echo "$(CYAN)Checking frontend...$(RESET)"
	cd $(FRONTEND_DIR) && $(NPM) run build

fmt: ## Format code
	@echo "$(CYAN)Formatting Go code...$(RESET)"
	$(GO) fmt ./...
	cd $(FRONTEND_DIR) && $(NPM) run lint --fix 2>/dev/null || true

## ============================================================================
## Setup & Installation
## ============================================================================

setup: ## Setup development environment (for non-Nix users)
	@echo "$(CYAN)Setting up development environment...$(RESET)"
	./scripts/setup.sh

install-deps: ## Install all dependencies
	@echo "$(CYAN)Installing Go dependencies...$(RESET)"
	$(GO) mod download
	@echo "$(CYAN)Installing frontend dependencies...$(RESET)"
	cd $(FRONTEND_DIR) && $(NPM) install
	@echo "$(GREEN)Dependencies installed!$(RESET)"

install-wails: ## Install Wails CLI
	@echo "$(CYAN)Installing Wails CLI...$(RESET)"
	$(GO) install github.com/wailsapp/wails/v2/cmd/wails@latest
	@echo "$(GREEN)Wails installed to $(HOME)/go/bin/wails$(RESET)"

## ============================================================================
## Utility
## ============================================================================

clean: ## Clean build artifacts
	@echo "$(CYAN)Cleaning...$(RESET)"
	rm -rf build/bin
	rm -f coverage.out coverage.html
	cd $(FRONTEND_DIR) && rm -rf dist node_modules/.vite

tidy: ## Tidy Go modules
	$(GO) mod tidy

generate: ## Generate Wails bindings
	@if command -v wails >/dev/null 2>&1; then \
		wails generate module; \
	else \
		$(WAILS) generate module; \
	fi

## ============================================================================
## Help
## ============================================================================

help: ## Show this help message
	@echo ""
	@echo "$(CYAN)Agent Orchestrator - Makefile Commands$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-15s$(RESET) %s\n", $$1, $$2}'
	@echo ""
