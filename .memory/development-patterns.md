# Development Patterns Memory Bank

## Common Code Patterns

### Entity Creation Pattern
```go
// Standard entity creation
func CreatePlayer(ecs *ecs.ECS, x, y int) ecs.EntityID {
    entityID := ecs.CreateEntity()
    
    ecs.AddComponents(entityID,
        &PositionComponent{X: x, Y: y, Level: 1},
        &RenderableComponent{Char: '@', Color: color.White, Layer: LayerPlayer},
        &HealthComponent{Current: 100, Max: 100},
        &StatsComponent{Strength: 10, Dexterity: 10, Constitution: 10},
        &PlayerTagComponent{},
        &InventoryComponent{Items: []EntityID{}, Capacity: 20},
        &TurnActorComponent{},
    )
    
    return entityID
}
```

### Event Handling Pattern
```go
// Event creation and dispatch
func TriggerEntityDeath(game *Game, entityID ecs.EntityID) {
    event := EntityDeathEvent{
        EntityID: entityID,
        Cause:    DeathCauseCombat,
        Location: game.ecs.GetPositionComponent(entityID),
    }
    
    game.eventQueue <- event
}

// Event processing
func (g *Game) ProcessEvents() {
    for len(g.eventQueue) > 0 {
        event := <-g.eventQueue
        switch e := event.(type) {
        case EntityDeathEvent:
            g.handleEntityDeath(e)
        case ItemPickupEvent:
            g.handleItemPickup(e)
        }
    }
}
```

### AI Action Pattern
```go
// AI decision making
func (g *Game) ProcessAITurn(entityID ecs.EntityID) {
    ai, hasAI := g.ecs.GetAIComponent(entityID)
    if !hasAI {
        return
    }
    
    // Strategic action sequence
    actions := g.planAIActions(entityID, ai)
    
    for _, action := range actions {
        if g.executeAction(entityID, action) {
            break // Action succeeded, end turn
        }
    }
}

func (g *Game) planAIActions(entityID ecs.EntityID, ai *AIComponent) []AIAction {
    switch ai.State {
    case StateHunting:
        return []AIAction{ActionAttack, ActionMove, ActionWait}
    case StateFleeing:
        return []AIAction{ActionFlee, ActionHide, ActionWait}
    default:
        return []AIAction{ActionWander, ActionWait}
    }
}
```

### Configuration Access Pattern
```go
// Configuration with fallbacks
func GetConfigValue[T any](config *GameConfig, path string, defaultValue T) T {
    if value, exists := config.Get(path); exists {
        if typed, ok := value.(T); ok {
            return typed
        }
    }
    return defaultValue
}

// Usage
dungeonWidth := GetConfigValue(config, "gameplay.dungeon.width", 80)
tileSize := GetConfigValue(config, "display.tile_size", 16)
```

### Save/Load Pattern
```go
// Save game state
func (g *Game) SaveGame(filename string) error {
    saveData := GameSaveData{
        Version:   SaveVersion,
        Timestamp: time.Now(),
        Depth:     g.Depth,
        ECS:       g.ecs.Serialize(),
        TurnQueue: g.turnQueue.Serialize(),
        Stats:     g.stats,
    }
    
    return io.SaveJSON(filename, saveData)
}

// Load game state
func (g *Game) LoadGame(filename string) error {
    var saveData GameSaveData
    if err := io.LoadJSON(filename, &saveData); err != nil {
        return err
    }
    
    // Validate save version and timestamp
    if err := validateSaveData(&saveData); err != nil {
        return err
    }
    
    g.Depth = saveData.Depth
    g.ecs.Deserialize(saveData.ECS)
    g.turnQueue.Deserialize(saveData.TurnQueue)
    g.stats = saveData.Stats
    
    return nil
}
```

## Testing Patterns

### Component Testing
```go
func TestHealthComponent(t *testing.T) {
    ecs := ecs.NewECS()
    entityID := ecs.CreateEntity()
    
    // Add component
    health := &HealthComponent{Current: 50, Max: 100}
    ecs.AddComponent(entityID, health)
    
    // Test retrieval
    retrieved, exists := ecs.GetHealthComponent(entityID)
    require.True(t, exists)
    assert.Equal(t, 50, retrieved.Current)
    
    // Test update
    ecs.UpdateHealthComponent(entityID, func(h *HealthComponent) {
        h.Current += 25
    })
    
    updated, _ := ecs.GetHealthComponent(entityID)
    assert.Equal(t, 75, updated.Current)
}
```

### Input Testing Pattern
```go
func TestKeyBindings(t *testing.T) {
    tests := []struct {
        name     string
        key      gruid.Key
        mode     UIMode
        expected Action
    }{
        {"Move North", gruid.KeyArrowUp, ModeNormal, ActionMoveNorth},
        {"Open Inventory", gruid.Key('i'), ModeNormal, ActionOpenInventory},
        {"Select Item", gruid.KeyEnter, ModeInventory, ActionSelectItem},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            action := mapKeyToAction(tt.key, tt.mode)
            assert.Equal(t, tt.expected, action)
        })
    }
}
```

### Integration Testing Pattern
```go
func TestCombatSystem(t *testing.T) {
    game := setupTestGame(t)
    
    // Create combatants
    player := createTestPlayer(game.ecs, 5, 5)
    monster := createTestMonster(game.ecs, 6, 5)
    
    // Execute combat
    game.executePlayerAction(player, ActionAttack, Direction{1, 0})
    
    // Verify results
    monsterHealth, _ := game.ecs.GetHealthComponent(monster)
    assert.Less(t, monsterHealth.Current, monsterHealth.Max)
}
```

## Error Handling Patterns

### Graceful Error Recovery
```go
func (g *Game) SafeEntityOperation(entityID ecs.EntityID, operation func() error) error {
    defer func() {
        if r := recover(); r != nil {
            log.Error("Entity operation panic", "entity", entityID, "error", r)
            // Clean up entity if corrupted
            g.ecs.RemoveEntity(entityID)
        }
    }()
    
    return operation()
}
```

### Configuration Validation
```go
func validateGameConfig(config *GameConfig) error {
    errors := []error{}
    
    if config.Gameplay.DungeonWidth < 10 {
        errors = append(errors, fmt.Errorf("dungeon width too small: %d", config.Gameplay.DungeonWidth))
    }
    
    if len(config.Input.KeyBindings) == 0 {
        errors = append(errors, fmt.Errorf("no key bindings configured"))
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("configuration validation failed: %v", errors)
    }
    
    return nil
}
```

## Performance Optimization Patterns

### Spatial Grid Usage
```go
// Efficient collision detection
func (g *Game) GetEntitiesAt(x, y int) []ecs.EntityID {
    // O(1) spatial lookup instead of O(n) component iteration
    return g.spatialGrid.GetEntitiesAt(x, y)
}

// Efficient neighbor search
func (g *Game) GetNearbyEntities(x, y, radius int) []ecs.EntityID {
    entities := []ecs.EntityID{}
    for dx := -radius; dx <= radius; dx++ {
        for dy := -radius; dy <= radius; dy++ {
            entities = append(entities, g.spatialGrid.GetEntitiesAt(x+dx, y+dy)...)
        }
    }
    return entities
}
```

### Component Query Optimization
```go
// Cache frequently used queries
type EntityCache struct {
    visibleEntities []ecs.EntityID
    aiActors       []ecs.EntityID
    lastUpdate     time.Time
}

func (c *EntityCache) GetVisibleEntities(ecs *ecs.ECS) []ecs.EntityID {
    if time.Since(c.lastUpdate) > time.Millisecond*100 {
        c.visibleEntities = ecs.GetEntitiesWithComponents(CPosition, CRenderable)
        c.lastUpdate = time.Now()
    }
    return c.visibleEntities
}
```

## Code Quality Standards

### Function Signatures
```go
// Clear, descriptive function names with context
func (g *Game) MovePlayerToPosition(newX, newY int) bool
func (g *Game) ProcessMonsterAITurn(monsterID ecs.EntityID) error
func (g *Game) CheckLineOfSight(fromX, fromY, toX, toY int) bool

// Consistent error handling
func (g *Game) LoadConfiguration(filename string) (*GameConfig, error)
func (g *Game) SaveGameState(filename string) error
```

### Documentation Standards
```go
// Package-level documentation
// Package game implements the core roguelike game logic using an Entity-Component-System
// architecture with turn-based mechanics and event-driven interactions.

// Function documentation with examples
// MoveEntity moves an entity to a new position, handling collision detection
// and spatial grid updates. Returns true if the move was successful.
//
// Example:
//   success := game.MoveEntity(playerID, newX, newY)
//   if !success {
//       // Handle blocked movement
//   }
func (g *Game) MoveEntity(entityID ecs.EntityID, newX, newY int) bool
```