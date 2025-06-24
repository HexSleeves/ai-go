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
