# AI-Go Architecture Memory Bank

## Quick Architecture Reference

### Core Systems Map
```
Game State (internal/game/) ─┬─ ECS World (internal/ecs/)
                             ├─ Event System (events.go)
                             ├─ Turn Queue (internal/turn_queue/)
                             ├─ Spatial Grid (spatial_grid.go)
                             └─ I/O Layer (internal/io/)
                             
UI Layer (internal/ui/) ─────┬─ SDL Driver (sdl.go)
                             ├─ WASM Driver (js.go)
                             └─ Gruid Framework
```

### Entity-Component-System Patterns

**Component Access Pattern:**
```go
// Type-safe component access
player, hasPlayer := ecs.GetPlayerComponent(entityID)
position, hasPos := ecs.GetPositionComponent(entityID)

// Bulk component updates
ecs.UpdateAIComponent(entityID, func(ai *AIComponent) {
    ai.State = StateHunting
    ai.Target = playerPos
})
```

**Entity Query Pattern:**
```go
// Find entities with specific components
entities := ecs.GetEntitiesWithComponents(CPosition, CRenderable, CHealth)
for _, entityID := range entities {
    // Process entity
}
```

### Event System Flow
```
Game Action → GameEvent → Event Queue → Event Processing → State Changes
            ↓
    EntityDeathEvent → Experience Award → Level Up Check
    ItemPickupEvent → Auto-equip → Inventory Update
    CombatEvent → Damage → Status Effects
```

### Turn Management
```
TurnQueue (Min-Heap) → Next Actor → Process Actions → Update Time → Queue Next Turn
     ↓
Priority by Time + Entity Speed + Action Cost
```

## File Location Quick Reference

| Component | Location | Purpose |
|-----------|----------|---------|
| Main Entry | `cmd/roguelike/main.go` | Application startup |
| Game State | `internal/game/game.go` | Central game manager |
| ECS Core | `internal/ecs/ecs.go` | Component system |
| Events | `internal/game/events.go` | Event-driven interactions |
| Turn Queue | `internal/turn_queue/queue.go` | Turn-based timing |
| UI Platform | `internal/ui/sdl.go`, `internal/ui/js.go` | Multi-platform rendering |
| Config | `assets/config/game_config.json` | Game configuration |
| Save System | `assets/saves/` + `internal/io/` | State persistence |

## Performance Hotspots

### Critical Performance Areas
1. **Component Access**: Thread-safe with RWMutex - avoid nested locks
2. **Spatial Queries**: SpatialGrid for O(1) position lookups 
3. **Pathfinding**: A* with heuristic caching
4. **Rendering**: Tile-based with sprite atlas optimization
5. **Turn Processing**: Priority queue with dead entity cleanup

### Memory Management
- Components stored in `map[ComponentType]map[EntityID]any`
- Entity recycling prevents ID exhaustion
- Spatial grid prevents O(n²) collision detection
- Event queue bounded to prevent memory leaks

## Testing Strategy

### Test Categories
- **Unit Tests**: Individual component behavior
- **Input Tests**: Key binding validation (322+ lines)
- **Integration Tests**: Full system interaction
- **Save/Load Tests**: State persistence validation
- **Benchmarks**: Performance-critical paths

### Test Commands
```bash
just test                    # Run all tests
just test-package internal/ecs  # Test specific package
just coverage                # Generate coverage reports
just test-verbose            # Detailed output
```

## Build Targets

### Development
```bash
just run-dev      # Go run with development flags
just run-race     # Race condition detection
just build        # Native executable → bin/
```

### Multi-Platform
```bash
just build-wasm   # WebAssembly → bin/
just serve-wasm   # Local web server
```

### Quality
```bash
just fmt          # Format code (required before commit)
just lint         # Go vet analysis
just deps         # Tidy dependencies
```