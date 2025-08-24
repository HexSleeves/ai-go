# Roguelike GRUID Project Justfile
# Run commands with 'just <command>'

# Default recipe to run when just is called without arguments
default:
    @just --list

# Build the game
build:
    go build -o ./bin/roguelike ./cmd/roguelike

# Run the game
run: build
    ./bin/roguelike

# Run directly with Go run command (uses SDL by default)
run-dev:
    go run ./cmd/roguelike

# Build and run with race detection
run-race:
    go run -race ./cmd/roguelike

# Clean build artifacts
clean:
    rm -f bin/roguelike

# Run tests
test *ARGS:
    go test ./... {{ARGS}}

# Run tests for specific package
test-package package:
    go test -v ./{{package}}

# Format code
fmt:
    go fmt ./...
    @echo "Go files formatted successfully"

# Check for code issues
lint:
    @echo "Running go vet..."
    go vet ./...
    @echo "Running go vet completed successfully"

# Install dependencies
deps:
    go mod tidy

# Build for web (WebAssembly)
build-wasm:
    GOOS=js GOARCH=wasm go build -o ../bin/roguelike.wasm ./cmd/roguelike

# Serve the WebAssembly build for local testing
serve-wasm: build-wasm
    go run ./cmd/wasm-server

# Generate documentation
docs:
    go doc -all ./... > ../docs/api.txt

# Run benchmarks
bench:
    go test -bench=. ./...

# Run benchmarks with memory profiling
bench-mem:
    go test -bench=. -benchmem ./...

# Profile CPU usage
profile-cpu:
    go test -bench=. -cpuprofile=cpu.prof ./...

# Profile memory usage
profile-mem:
    go test -bench=. -memprofile=mem.prof ./...

# Test coverage
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o ../docs/coverage.html

# Test coverage with verbose output
coverage-verbose:
    go test -v -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out
# Workflow Enhancement Commands
# =============================

# Build the workflow CLI
build-workflow:
    go build -o ./bin/workflow ./cmd/workflow

# Initialize workflow configuration
workflow-init: build-workflow
    ./bin/workflow config init

# Show workflow status
workflow-status: build-workflow
    ./bin/workflow status

# Validate workflow configuration
workflow-config-validate: build-workflow
    ./bin/workflow config validate

# Show workflow configuration
workflow-config-show: build-workflow
    ./bin/workflow config show

# Spec-driven development commands
workflow-spec-validate: build-workflow
    ./bin/workflow spec validate

workflow-spec-generate: build-workflow
    ./bin/workflow spec generate

workflow-spec-status: build-workflow
    ./bin/workflow spec status

# Content pipeline commands
workflow-content-generate: build-workflow
    ./bin/workflow content generate

workflow-content-validate: build-workflow
    ./bin/workflow content validate

workflow-content-reload: build-workflow
    ./bin/workflow content reload

# Game balance commands
workflow-balance-analyze: build-workflow
    ./bin/workflow balance analyze

workflow-balance-simulate: build-workflow
    ./bin/workflow balance simulate

workflow-balance-recommend: build-workflow
    ./bin/workflow balance recommend

# Testing framework commands
workflow-test-run: build-workflow
    ./bin/workflow test run

workflow-test-generate: build-workflow
    ./bin/workflow test generate

workflow-test-coverage: build-workflow
    ./bin/workflow test coverage

# Performance optimization commands
workflow-perf-profile: build-workflow
    ./bin/workflow performance profile

workflow-perf-monitor: build-workflow
    ./bin/workflow performance monitor

workflow-perf-optimize: build-workflow
    ./bin/workflow performance optimize

# Plugin system commands
workflow-plugin-list: build-workflow
    ./bin/workflow plugin list

workflow-plugin-install plugin: build-workflow
    ./bin/workflow plugin install {{plugin}}

workflow-plugin-remove plugin: build-workflow
    ./bin/workflow plugin remove {{plugin}}

# Documentation commands
workflow-docs-generate: build-workflow
    ./bin/workflow docs generate

workflow-docs-serve: build-workflow
    ./bin/workflow docs serve

workflow-docs-update: build-workflow
    ./bin/workflow docs update

# CI/CD pipeline commands
workflow-ci-setup: build-workflow
    ./bin/workflow ci setup

workflow-ci-status: build-workflow
    ./bin/workflow ci status

workflow-ci-deploy: build-workflow
    ./bin/workflow ci deploy

# Debug and development tools
workflow-debug-visualize: build-workflow
    ./bin/workflow debug visualize

workflow-debug-replay: build-workflow
    ./bin/workflow debug replay

workflow-debug-metrics: build-workflow
    ./bin/workflow debug metrics

# Workflow development and testing
workflow-test: build-workflow
    go test ./internal/workflow/...

workflow-test-verbose: build-workflow
    go test -v ./internal/workflow/...

workflow-coverage: build-workflow
    go test -coverprofile=workflow-coverage.out ./internal/workflow/...
    go tool cover -html=workflow-coverage.out -o docs/workflow-coverage.html

# Clean workflow artifacts
workflow-clean:
    rm -f bin/workflow
    rm -f workflow-coverage.out
    rm -f docs/workflow-coverage.html

# Show all workflow commands
workflow-help:
    @echo "Workflow Enhancement Commands:"
    @echo "=============================="
    @echo ""
    @echo "Configuration:"
    @echo "  workflow-init              Initialize workflow configuration"
    @echo "  workflow-status            Show workflow status"
    @echo "  workflow-config-validate   Validate configuration"
    @echo "  workflow-config-show       Show configuration"
    @echo ""
    @echo "Spec-driven Development:"
    @echo "  workflow-spec-validate     Validate specifications"
    @echo "  workflow-spec-generate     Generate specifications"
    @echo "  workflow-spec-status       Show spec status"
    @echo ""
    @echo "Content Pipeline:"
    @echo "  workflow-content-generate  Generate content"
    @echo "  workflow-content-validate  Validate content"
    @echo "  workflow-content-reload    Hot-reload content"
    @echo ""
    @echo "Game Balance:"
    @echo "  workflow-balance-analyze   Analyze game balance"
    @echo "  workflow-balance-simulate  Simulate gameplay"
    @echo "  workflow-balance-recommend Generate balance recommendations"
    @echo ""
    @echo "Testing:"
    @echo "  workflow-test-run          Run tests"
    @echo "  workflow-test-generate     Generate tests"
    @echo "  workflow-test-coverage     Run coverage analysis"
    @echo ""
    @echo "Performance:"
    @echo "  workflow-perf-profile      Profile performance"
    @echo "  workflow-perf-monitor      Monitor performance"
    @echo "  workflow-perf-optimize     Optimize performance"
    @echo ""
    @echo "Plugins:"
    @echo "  workflow-plugin-list       List plugins"
    @echo "  workflow-plugin-install    Install plugin"
    @echo "  workflow-plugin-remove     Remove plugin"
    @echo ""
    @echo "Documentation:"
    @echo "  workflow-docs-generate     Generate documentation"
    @echo "  workflow-docs-serve        Serve documentation"
    @echo "  workflow-docs-update       Update documentation"
    @echo ""
    @echo "CI/CD:"
    @echo "  workflow-ci-setup          Setup CI/CD"
    @echo "  workflow-ci-status         Show CI/CD status"
    @echo "  workflow-ci-deploy         Deploy application"
    @echo ""
    @echo "Debug Tools:"
    @echo "  workflow-debug-visualize   Visualize debug info"
    @echo "  workflow-debug-replay      Replay debug session"
    @echo "  workflow-debug-metrics     Show debug metrics"
    @echo ""
    @echo "Development:"
    @echo "  workflow-test              Run workflow tests"
    @echo "  workflow-coverage          Generate workflow coverage"
    @echo "  workflow-clean             Clean workflow artifacts"