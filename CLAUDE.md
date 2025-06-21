# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## AI-Go Roguelike

A comprehensive roguelike game built in Go using Entity-Component-System (ECS) architecture with multi-platform support (native SDL and WebAssembly).

## Common Development Commands

This project uses Just (a command runner) for build automation. Key commands:

**Development:**
- `just run-dev` - Run with Go directly (development mode)
- `just run-race` - Run with race detection enabled
- `just build` - Build executable
- `just clean` - Clean build artifacts

**Testing:**
- `just test` - Run all tests
- `just test-verbose` - Verbose test output
- `just test-package <package>` - Test specific package
- `just coverage` - Generate test coverage reports

**Code Quality:**
- `just fmt` - Format Go code (run before commits)
- `just lint` - Run go vet for code analysis
- `just deps` - Tidy module dependencies

**Multi-Platform:**
- `just build-wasm` - Build for WebAssembly
- `just serve-wasm` - Serve WASM build locally

## Architecture Overview

### Entity-Component-System (ECS) Pattern
- **Entities**: Integer IDs (`EntityID`) representing game objects
- **Components**: Data structures stored in type-safe maps (`internal/ecs/components/`)
- **Systems**: Functions operating on entities with specific components
- **World**: Central ECS manager in `internal/ecs/world.go`

### Key Systems
- **Game State**: Central `Game` struct managing world state, turn queue, and spatial systems (`internal/game/`)
- **Turn Queue**: Priority heap-based turn management (`internal/turn_queue/`)
- **Configuration**: JSON-based config with hot-reloading (`internal/config/`)
- **UI Layer**: Multi-platform UI with SDL/WASM support (`internal/ui/`)

### Project Structure
```
cmd/roguelike/          # Main application entry
config/                 # JSON configuration files
internal/
├── ecs/               # Entity-Component-System core
│   └── components/    # Component definitions
├── game/              # Core game logic and save system
├── ui/                # Cross-platform UI layer
├── turn_queue/        # Turn management
├── config/            # Configuration management
└── utils/             # Shared utilities
```

## Development Patterns

### Code Organization
- **Package-by-Feature**: Clear separation (game, ui, ecs, config)
- **Interface-Driven**: Clean abstractions for cross-platform support
- **Configuration-First**: Extensive external JSON configuration

### Testing Strategy
- Unit tests for individual components
- Integration tests for full systems
- Comprehensive save/load testing with timestamp validation
- Benchmarks for performance-critical code

### Cross-Platform Support
- Build tags for platform-specific code
- Separate drivers for SDL (desktop) and JS (web)
- Shared core logic with platform-specific UI implementations

## Key Dependencies
- `codeberg.org/anaseto/gruid` - Core terminal UI framework
- `codeberg.org/anaseto/gruid-sdl` - SDL driver for desktop
- `codeberg.org/anaseto/gruid-js` - WebAssembly support
- `github.com/sirupsen/logrus` - Structured logging

## Configuration System
The game uses a sophisticated JSON configuration system (`config/game_config.json`) with:
- Gameplay settings (difficulty, generation parameters)
- Display options (ASCII vs tiles, colors)
- Input mappings and audio settings
- Advanced developer options

Configuration is validated on load with comprehensive error reporting.

## Save System
Complete game state serialization to JSON with:
- Full ECS world state preservation
- Timestamp-based validation
- Comprehensive error handling for corrupted saves
- Located in `internal/game/saves/`