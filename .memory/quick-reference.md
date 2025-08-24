# Quick Reference Memory Bank

## Essential Commands

### Development Workflow
```bash
# Start development
just run-dev              # Run with development flags
just run-race             # Run with race detection

# Code quality (run before commits)
just fmt                  # Format all Go code
just lint                 # Run go vet analysis
just deps                 # Tidy module dependencies

# Testing
just test                 # Run all tests  
just test-verbose         # Detailed test output
just coverage             # Generate coverage reports

# Building
just build                # Native executable → bin/roguelike
just build-wasm           # WebAssembly → bin/roguelike.wasm
just serve-wasm           # Serve WASM locally
```

## File Locations

### Core Files (Most Important)
| Purpose | Location |
|---------|----------|
| Main Entry | `cmd/roguelike/main.go` |
| Game State | `internal/game/game.go` |
| ECS Core | `internal/ecs/ecs.go` |
| Events | `internal/game/events.go` |
| Config | `assets/config/game_config.json` |

### System Locations
| System | Files |
|--------|-------|
| ECS Components | `internal/ecs/components/*.go` |
| Game Logic | `internal/game/*.go` |
| UI Layer | `internal/ui/*.go` |
| Turn Management | `internal/turn_queue/*.go` |
| I/O Operations | `internal/io/*.go` |
| Utilities | `internal/utils/*.go` |

## Common Code Patterns

### Entity Creation
```go
entityID := ecs.CreateEntity()
ecs.AddComponents(entityID,
    &PositionComponent{X: x, Y: y},
    &HealthComponent{Current: 100, Max: 100},
    &RenderableComponent{Char: '@', Color: color.White},
)
```

### Component Access
```go
// Safe access with existence check
if pos, hasPos := ecs.GetPositionComponent(entityID); hasPos {
    // Use pos safely
}

// Safe update with callback
ecs.UpdateHealthComponent(entityID, func(h *HealthComponent) {
    h.Current -= damage
})
```

### Event Handling
```go
// Trigger event
event := EntityDeathEvent{EntityID: deadEntity}
game.eventQueue <- event

// Process events
for len(game.eventQueue) > 0 {
    event := <-game.eventQueue
    game.processEvent(event)
}
```

## Architecture Quick Map

### Data Flow
```
User Input → Game Actions → Events → State Changes → Rendering
                 ↓
            Turn Queue → AI Processing → Component Updates
```

### Component Access Flow
```
Game Logic → ECS Safe Accessors → Component Maps → Thread-Safe Updates
```

### Event System Flow  
```
Game Action → Event Creation → Event Queue → Event Processing → Consequences
```

## Performance Hotspots

### Critical Areas
1. **Component Access** - Use safe accessors, avoid nested locks
2. **Spatial Queries** - Use SpatialGrid for O(1) lookups
3. **AI Processing** - Batch operations, cache calculations
4. **Rendering** - Sprite atlas, layer-based rendering
5. **Turn Queue** - Priority heap with cleanup

### Optimization Guidelines
- Use Manhattan distance for non-pathfinding
- Cache entity queries when possible
- Batch component operations
- Minimize lock contention in ECS
- Use spatial grid for collision detection

## Testing Quick Reference

### Test Categories
```bash
# Component tests
internal/ecs/components/*_test.go

# Input validation  
internal/game/input_test.go           # 322+ lines

# System integration
internal/game/*_test.go

# Benchmarks
*_benchmark_test.go
```

### Test Patterns
```go
// Component testing
func TestComponent(t *testing.T) {
    ecs := ecs.NewECS()
    entityID := ecs.CreateEntity()
    
    component := &SomeComponent{Value: 42}
    ecs.AddComponent(entityID, component)
    
    retrieved, exists := ecs.GetSomeComponent(entityID)
    require.True(t, exists)
    assert.Equal(t, 42, retrieved.Value)
}

// Input testing
func TestKeyBinding(t *testing.T) {
    action := mapKeyToAction(gruid.Key('i'), ModeNormal)
    assert.Equal(t, ActionOpenInventory, action)
}
```

## Build Targets

### Platform Support
```bash
# Native (development)
GOOS=darwin GOARCH=amd64 go build    # macOS
GOOS=linux GOARCH=amd64 go build     # Linux  
GOOS=windows GOARCH=amd64 go build   # Windows

# WebAssembly
GOOS=js GOARCH=wasm go build         # Browser
```

### Build Tags
```go
//go:build !js        # Native platforms only
//go:build js         # WebAssembly only  
//go:build debug      # Debug builds
```

## Configuration Quick Access

### Config Structure
```json
{
  "gameplay": {
    "dungeon": {"width": 80, "height": 24},
    "difficulty": {"monster_spawn_rate": 0.1}
  },
  "display": {
    "use_tiles": true,
    "tile_size": 16,
    "font_size": 14
  },
  "input": {
    "key_bindings": {"move_north": "ArrowUp"}
  }
}
```

### Config Access
```go
width := config.Gameplay.Dungeon.Width
useTiles := config.Display.UseTiles
keyBindings := config.Input.KeyBindings
```

## Error Handling

### Common Error Patterns
```go
// Configuration errors
if err := config.Validate(); err != nil {
    return fmt.Errorf("config validation failed: %w", err)
}

// Save/load errors
if err := game.SaveGame(filename); err != nil {
    return fmt.Errorf("failed to save game: %w", err)
}

// Component access errors  
if !ecs.EntityExists(entityID) {
    return fmt.Errorf("entity %d does not exist", entityID)
}
```

### Recovery Patterns
```go
// Graceful degradation
defer func() {
    if r := recover(); r != nil {
        log.Error("operation failed", "error", r)
        // Clean up and continue
    }
}()
```

## Dependencies

### Core Dependencies
```go
"codeberg.org/anaseto/gruid"          // Terminal UI framework
"codeberg.org/anaseto/gruid-sdl"      // SDL desktop driver  
"codeberg.org/anaseto/gruid-js"       // WebAssembly driver
"golang.org/x/image"                  // Image processing
```

### Standard Library Usage
```go
"log/slog"          // Structured logging
"encoding/json"     // Configuration and saves
"container/heap"    // Turn queue priority heap
"sync"              // Thread safety (RWMutex)
```