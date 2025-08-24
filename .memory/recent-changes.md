# Recent Changes Memory Bank

## Major Recent Changes (Last 5 Commits)

### Event System Implementation (NEW)
**File**: `internal/game/events.go` (326 lines)  
**Commit**: b11333f - "refactor: reorganize game structure and implement event system"

**Key Features**:
```go
type GameEvent interface {
    gruid.Msg  // Integrates with gruid message system
}

type EntityDeathEvent struct {
    EntityID ecs.EntityID
    Cause    DeathCause
    Location PositionComponent
}

type ItemPickupEvent struct {
    ItemID   ecs.EntityID
    PlayerID ecs.EntityID
    AutoEquip bool
}
```

**Event Processing Flow**:
1. Game actions trigger events
2. Events queued in game event channel
3. Events processed during update cycle
4. Events trigger consequences (XP, auto-equip, game over)

### Enhanced Input System
**File**: `internal/game/input_test.go` (322+ lines)  
**Commit**: 797749a - "refactor: enhance inventory action handling and key bindings"

**Mode-Specific Key Bindings**:
```go
KEYS_NORMAL = map[gruid.Key]Action{
    gruid.KeyArrowUp:    ActionMoveNorth,
    gruid.Key('i'):      ActionOpenInventory,
    gruid.Key('c'):      ActionOpenCharacter,
}

KEYS_INVENTORY_SCREEN = map[gruid.Key]Action{
    gruid.KeyEnter:      ActionSelectItem,
    gruid.Key('d'):      ActionDropItem,
    gruid.KeyEscape:     ActionCloseScreen,
}
```

**Comprehensive Testing**:
- Key binding validation for all modes
- UI state transition testing
- Input edge case handling

### I/O System Reorganization (NEW)
**Package**: `internal/io/` (136 total lines)  
**Commit**: 022be8d - "refactor: game configuration management and logging"

**Path Management**:
```go
func GetRepositoryRoot() (string, error)
func GetConfigDirectory() string
func GetSaveDirectory() string  
func GetLogDirectory() string
```

**Configuration Migration**:
- `config/game_config.json` → `assets/config/game_config.json`
- Enhanced cross-platform path resolution
- Local vs global configuration support

### AI System Enhancement
**Commits**: Multiple refactoring commits

**Strategic AI Actions**:
```go
type AIAction int
const (
    ActionAttack AIAction = iota
    ActionMove
    ActionFlee
    ActionHide
    ActionWander
    ActionWait
)

// AI now plans action sequences instead of single actions
func (g *Game) planAIActions(entityID ecs.EntityID) []AIAction {
    // Returns prioritized action list
}
```

**Distance Optimization**:
- Manhattan distance for non-pathfinding calculations
- Reduced computational overhead in AI decision making
- Optimized monster targeting and player detection

### Build System Updates
**File**: `justfile` updates  
**Binary Organization**: All builds output to `bin/` directory

**Enhanced Commands**:
```bash
# Flexible test runner
just test [ARGS]         # Pass arguments to go test
just test-package PKG    # Test specific package
just coverage           # Generate coverage reports

# Organized build output
just build              # → bin/roguelike
just build-wasm         # → bin/roguelike.wasm
```

## File Movements and Restructuring

### Configuration Reorganization
```
OLD: config/game_config.json
NEW: assets/config/game_config.json

NEW: assets/saves/          # Save file location
NEW: assets/tiles/          # Sprite assets
```

### New Package Creation
```
NEW: internal/io/           # File I/O utilities
├── io.go                  # Path resolution and directory management
└── saveload.go           # Save/load helper functions
```

### Test Refactoring
```
REMOVED: internal/game/integration_test.go  (295 lines)
ADDED:   internal/game/input_test.go        (322+ lines)

Focus shift: Integration tests → Focused unit tests with better coverage
```

## Breaking Changes to Watch

### Configuration Path Changes
- **Old**: `config.LoadConfig("config/game_config.json")`
- **New**: `config.LoadConfig(io.GetConfigDirectory() + "/game_config.json")`

### Event System Integration
- Game actions now trigger events instead of direct state changes
- Event processing required in main game loop
- Event consequences may have delayed effects

### AI Behavior Changes
- AI now executes action sequences per turn
- More strategic and less predictable behavior
- May require rebalancing of game difficulty

## Performance Improvements

### Computational Optimizations
1. **Distance Calculations**: Manhattan distance where appropriate (O(1) vs pathfinding)
2. **AI Decision Making**: Action sequences reduce per-turn computation
3. **Event Processing**: Batch event processing vs immediate execution
4. **Spatial Queries**: Continued spatial grid optimization

### Memory Management
1. **Component Storage**: No changes to core ECS efficiency
2. **Event Queue**: Bounded queues prevent memory leaks
3. **AI State**: Reduced state complexity with action sequences

## Development Impact

### Testing Strategy Evolution
- **From**: Large integration tests covering multiple systems
- **To**: Focused unit tests with specific validation
- **Result**: Better test maintainability and faster feedback

### Code Organization Improvements
- **Package-by-feature**: Reinforced with new `io` package
- **Clear boundaries**: Event system provides clean interfaces
- **Configuration management**: Centralized path handling

### Build Process Enhancements
- **Consistent output**: All binaries in `bin/` directory
- **Flexible testing**: Parameterized test commands
- **Cross-platform**: Improved WebAssembly build process

## Forward Compatibility

### Event System Extensibility
The new event system is designed for easy extension:
```go
// Adding new event types is straightforward
type PlayerLevelUpEvent struct {
    PlayerID ecs.EntityID
    NewLevel int
    StatPoints int
}

// Event processing is centralized and extensible
func (g *Game) ProcessEvents() {
    for len(g.eventQueue) > 0 {
        event := <-g.eventQueue
        switch e := event.(type) {
        // Easy to add new cases
        case PlayerLevelUpEvent:
            g.handlePlayerLevelUp(e)
        }
    }
}
```

### Configuration System Future-Proofing
- Path resolution supports future config locations
- Local/global config distinction enables user preferences
- JSON validation framework supports schema evolution