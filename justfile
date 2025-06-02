# Roguelike GRUID Project Justfile
# Run commands with 'just <command>'

# Default recipe to run when just is called without arguments
default:
    @just --list

# Build the game
build:
    cd roguelike-gruid-project && go build -o ../roguelike ./cmd/roguelike

# Run the game
run: build
    ./roguelike

# Run directly with Go run command
run-dev:
    cd roguelike-gruid-project && go run ./cmd/roguelike

# Run with SDL backend
run-sdl:
    cd roguelike-gruid-project && go run --tags sdl ./cmd/roguelike

# Build and run with race detection
run-race:
    cd roguelike-gruid-project && go run -race ./cmd/roguelike

# Clean build artifacts
clean:
    rm -f roguelike

# Run tests
test:
    cd roguelike-gruid-project && go test ./...

# Run tests with verbose output
test-verbose:
    cd roguelike-gruid-project && go test -v ./...

# Run tests for specific package
test-package package:
    cd roguelike-gruid-project && go test -v ./{{package}}

# Format code
fmt:
    cd roguelike-gruid-project && go fmt ./...
    @echo "Go files formatted successfully"

# Check for code issues
lint:
    @echo "Running go vet..."
    cd roguelike-gruid-project && go vet ./...
    @echo "Running go vet completed successfully"

# Install dependencies
deps:
    cd roguelike-gruid-project && go mod tidy

# Build for web (WebAssembly)
build-wasm:
    cd roguelike-gruid-project && GOOS=js GOARCH=wasm go build -o ../roguelike.wasm ./cmd/roguelike

# Serve the WebAssembly build for local testing
serve-wasm: build-wasm
    cd roguelike-gruid-project && go run ./cmd/wasm-server

# Generate documentation
docs:
    cd roguelike-gruid-project && go doc -all ./... > ../docs/api.txt

# Run benchmarks
bench:
    cd roguelike-gruid-project && go test -bench=. ./...

# Run benchmarks with memory profiling
bench-mem:
    cd roguelike-gruid-project && go test -bench=. -benchmem ./...

# Profile CPU usage
profile-cpu:
    cd roguelike-gruid-project && go test -bench=. -cpuprofile=cpu.prof ./...

# Profile memory usage
profile-mem:
    cd roguelike-gruid-project && go test -bench=. -memprofile=mem.prof ./...

# Test coverage
coverage:
    cd roguelike-gruid-project && go test -coverprofile=coverage.out ./...
    cd roguelike-gruid-project && go tool cover -html=coverage.out -o ../docs/coverage.html

# Test coverage with verbose output
coverage-verbose:
    cd roguelike-gruid-project && go test -v -coverprofile=coverage.out ./...
    cd roguelike-gruid-project && go tool cover -func=coverage.out
