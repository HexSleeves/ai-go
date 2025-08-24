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
- **Event System**: Game event handling with `GameEvent` interface for deaths, combat, level-ups (`internal/game/events.go`)
- **Turn Queue**: Priority heap-based turn management (`internal/turn_queue/`)
- **Configuration**: JSON-based config with hot-reloading (`internal/config/`)
- **I/O Management**: File operations, save/load utilities, path resolution (`internal/io/`)
- **UI Layer**: Multi-platform UI with SDL/WASM support (`internal/ui/`)

### Project Structure

```bash
cmd/roguelike/          # Main application entry
assets/                 # Game assets and configuration
├── config/            # JSON configuration files (moved from root)
├── saves/             # Save file storage
└── tiles/             # Sprite assets
internal/
├── ecs/               # Entity-Component-System core
│   └── components/    # Component definitions
├── game/              # Core game logic and save system
├── ui/                # Cross-platform UI layer
├── turn_queue/        # Turn management
├── config/            # Configuration management
├── io/                # File I/O utilities (NEW)
└── utils/             # Shared utilities
docs/                   # Development documentation
├── improvement-tasks.md # Comprehensive improvement plan
```

## Development Patterns

### Code Organization

- **Package-by-Feature**: Clear separation (game, ui, ecs, config)
- **Interface-Driven**: Clean abstractions for cross-platform support
- **Configuration-First**: Extensive external JSON configuration

### Testing Strategy

- Unit tests for individual components with comprehensive input testing (322+ lines)
- Focused test suites replacing legacy integration tests
- Comprehensive save/load testing with timestamp validation
- Key binding and UI state validation testing
- Benchmarks for performance-critical code

### Cross-Platform Support

- Build tags for platform-specific code
- Separate drivers for SDL (desktop) and JS (web)
- Shared core logic with platform-specific UI implementations

## Key Dependencies

- `codeberg.org/anaseto/gruid` - Core terminal UI framework
- `codeberg.org/anaseto/gruid-sdl` - SDL driver for desktop
- `codeberg.org/anaseto/gruid-js` - WebAssembly support
- `log/slog` - Structured logging

## Configuration System

The game uses a sophisticated JSON configuration system (`assets/config/game_config.json`) with:

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
- Located in `assets/saves/` with utilities in `internal/io/`
- Cross-platform path resolution with repo root detection
- Enhanced file-based logging with timestamp management

## Recent Architectural Improvements

### Event System (NEW)

- **Event-driven architecture** with `GameEvent` interface extending `gruid.Msg`
- **Event types**: EntityDeathEvent, ItemPickupEvent, CombatEvent, LevelUpEvent, GameOverEvent
- **Enhanced game interactions** with experience rewards, visual feedback, and state changes
- **Auto-pickup and auto-equip** behavior through event consequences

### Enhanced AI System

- **Strategic action sequences** replacing single-action AI
- **State-based monster behavior** with multi-turn planning
- **Optimized distance calculations** using Manhattan distance where appropriate
- **Advanced pathfinding** with obstacle avoidance and goal seeking

### Comprehensive Input System

- **Mode-specific key bindings**: Normal, Inventory, Character, Message screens
- **Extensive input testing** with 322+ lines of validation coverage
- **Enhanced inventory actions** with improved item handling
- **Flexible key mapping system** supporting complex interactions

### Build System Updates

- **Organized binary output** to `bin/` directory structure
- **Enhanced Just commands** with flexible test runners
- **WebAssembly build improvements** with consistent paths
- **Clean dependency management** with stable version locks

## Component Architecture (ECS Details)

### Core Component Types (36 total)

```go
// Entity identification and AI
CAIComponent, CAITag, CPlayerTag, CCorpseTag

// Spatial and rendering
CPosition, CRenderable, CBlocksMovement, CFOV

// Character attributes
CHealth, CStats, CExperience, CSkills, CMana, CStamina

// Inventory and equipment
CInventory, CEquipment, CItemPickup

// Combat and interaction
CCombat, CStatusEffects, CName

// Turn management and pathfinding
CTurnActor, CPathfindingComponent
```

### ECS Performance Features

- **Thread-safe component access** with RWMutex protection
- **Generic component storage** with type-safe getters
- **Batch component operations** with reflection-based type detection
- **Optimized entity queries** with component filtering
- **Memory-efficient storage** using map-based component organization

## Development Workflow

### Code Quality Standards

- **Format before commit**: Always run `just fmt`
- **Comprehensive testing**: Unit tests with input validation
- **Clean architecture**: Package-by-feature organization
- **Performance focus**: Benchmarks for critical paths
- **Cross-platform support**: Build tags and separate drivers
