# ECS Component Reference

## Core Component Types (36 total)

### Entity Identification
```go
CPlayerTag        // Marks the player entity
CAITag           // Marks AI-controlled entities  
CCorpseTag       // Marks dead entities
```

### Spatial & Rendering
```go
CPosition        // X, Y coordinates + dungeon level
CRenderable      // Character, color, render layer
CBlocksMovement  // Solid entities that block movement
CFOV             // Field of view component
```

### Character Attributes
```go
CHealth          // Current/max HP
CStats           // Strength, dexterity, constitution, etc.
CExperience      // XP, level, next level threshold
CSkills          // Combat skills, magic schools
CMana            // Current/max mana for magic
CStamina         // Current/max stamina for actions
```

### Inventory & Equipment
```go
CInventory       // List of carried items
CEquipment       // Worn/wielded items by slot
CItemPickup      // Items that can be picked up
```

### Combat & Interaction
```go
CCombat          // Attack power, defense, combat stats
CStatusEffects   // Buffs, debuffs, temporary effects
CName            // Display name for entities
```

### AI & Pathfinding
```go
CAIComponent     // AI state, behavior, targets
CPathfindingComponent // A* pathfinding data
```

### Turn Management
```go
CTurnActor       // Entities that take turns
```

## Component Access Patterns

### Safe Component Access
```go
// Check existence before access
if player, hasPlayer := ecs.GetPlayerComponent(entityID); hasPlayer {
    // Use player component safely
}

// Batch component check
if pos, hasPos := ecs.GetPositionComponent(entityID); hasPos {
    if renderable, hasRender := ecs.GetRenderableComponent(entityID); hasRender {
        // Entity has both position and renderable
    }
}
```

### Component Updates
```go
// Safe component update with callback
ecs.UpdateHealthComponent(entityID, func(health *HealthComponent) {
    health.Current -= damage
    if health.Current <= 0 {
        health.Current = 0
        // Trigger death event
    }
})

// Bulk component addition
ecs.AddComponents(entityID, 
    &PositionComponent{X: x, Y: y},
    &RenderableComponent{Char: '@', Color: color.White},
    &HealthComponent{Current: 100, Max: 100},
)
```

### Entity Queries
```go
// Find all entities with specific component combination
combatants := ecs.GetEntitiesWithComponents(CPosition, CHealth, CCombat)

// Find all visible entities
visibleEntities := ecs.GetEntitiesWithComponents(CPosition, CRenderable, CFOV)

// Find all AI actors ready for turn
aiActors := ecs.GetEntitiesWithComponents(CAIComponent, CTurnActor, CPosition)
```

## Component Data Structures

### Common Component Patterns
```go
type PositionComponent struct {
    X, Y  int
    Level int  // Dungeon level/depth
}

type HealthComponent struct {
    Current, Max int
    Regeneration int  // HP regen per turn
}

type AIComponent struct {
    State      AIState       // Idle, Hunting, Fleeing, etc.
    Target     EntityID      // Current target entity
    LastSeen   PositionComponent // Last known player position
    ActionQueue []AIAction   // Planned actions
}

type InventoryComponent struct {
    Items    []EntityID     // List of item entities
    Capacity int            // Max items
    Weight   int            // Current weight
}
```

### Component Type Constants
```go
const (
    // Core identification
    CPlayerTag = "player_tag"
    CAITag = "ai_tag" 
    CCorpseTag = "corpse_tag"
    
    // Spatial
    CPosition = "position"
    CRenderable = "renderable"
    CBlocksMovement = "blocks_movement"
    CFOV = "fov"
    
    // Character
    CHealth = "health"
    CStats = "stats"
    CExperience = "experience"
    CSkills = "skills"
    CMana = "mana"
    CStamina = "stamina"
    
    // Inventory
    CInventory = "inventory"
    CEquipment = "equipment"
    CItemPickup = "item_pickup"
    
    // Combat
    CCombat = "combat"
    CStatusEffects = "status_effects"
    CName = "name"
    
    // AI & Turns
    CAIComponent = "ai_component"
    CPathfindingComponent = "pathfinding_component"
    CTurnActor = "turn_actor"
)
```

## Thread Safety

### Safe Access Rules
1. **Always use provided accessors** - Don't access component maps directly
2. **Update callbacks are atomic** - Use UpdateXComponent functions
3. **Bulk operations are protected** - AddComponents/RemoveComponents are safe
4. **Read operations are concurrent** - Multiple goroutines can read simultaneously
5. **Write operations are exclusive** - Only one writer at a time

### Performance Considerations
- Component access is O(1) hash map lookup
- Entity queries are O(n) but cached where possible
- Bulk operations minimize lock overhead
- Component removal is lazy to avoid mid-iteration issues