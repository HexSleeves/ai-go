# ECS Query System Improvement Specification

## Description

Review and improve the current ECS (Entity Component System) structure to provide better query mechanisms that reduce the need for extensive error handling when accessing components. The current system requires frequent error checking when retrieving components, making the code verbose and error-prone.

## Current Issues Analysis

### 1. Verbose Error Handling Patterns

```go
// Current pattern - verbose and repetitive
pos, ok := g.ecs.GetPosition(entityID)
if !ok {
    return false, fmt.Errorf("entity %d position not found", entityID)
}

// Often ignored errors lead to potential bugs
pos, _ := world.GetPosition(id)  // Error ignored!
```

### 2. Inconsistent Error Handling

- Some places check errors: `pos, ok := ecs.GetPosition(id); if !ok { ... }`
- Others ignore them: `pos, _ := ecs.GetPosition(id)`
- Complex queries require multiple error checks

### 3. Type Assertion Complexity

```go
// Current generic function requires type assertions
func GetComponentTyped[T any](ecs *ECS, id EntityID, compType ComponentType) (T, bool) {
    var result T
    comp, ok := ecs.getComponent(id, compType)
    if !ok {
        return result, false
    }
    typedComp, ok := comp.(T)
    if !ok {
        return result, false
    }
    return typedComp, true
}
```

### 4. Lack of Fluent Query Interface

No way to build complex queries like "get all entities with Position, Health, and Renderable components where Health > 0"

## Requirements

### 1. Safe Component Accessors

- [ ] Create `GetPositionSafe(id)` that returns zero value instead of error
- [ ] Implement `GetHealthSafe(id)` with default Health{} return
- [ ] Add `GetRenderableSafe(id)` with default Renderable{} return
- [ ] Provide `GetNameSafe(id)` returning empty string for missing names
- [ ] Create `GetFOVSafe(id)` returning nil for missing FOV

### 2. Optional Component Patterns

- [ ] Implement `Option[T]` type for nullable components
- [ ] Create `GetPositionOpt(id) Option[gruid.Point]` pattern
- [ ] Add `GetHealthOpt(id) Option[Health]` for optional health access
- [ ] Provide fluent methods like `pos.IsSome()`, `pos.Unwrap()`, `pos.UnwrapOr(default)`

### 3. Query Builder Interface

- [ ] Create `Query` struct for building complex queries
- [ ] Implement `ecs.Query().WithPosition().WithHealth().Execute()` pattern
- [ ] Add filtering capabilities: `Query().WithHealth().Where(func(h Health) bool { return h.CurrentHP > 0 })`
- [ ] Support batch operations: `Query().WithPosition().ForEach(func(id EntityID, pos gruid.Point) { ... })`

### 4. Batch Component Access

- [ ] Create `GetComponents(id EntityID) ComponentSet` for accessing multiple components
- [ ] Implement `ComponentSet.Position()`, `ComponentSet.Health()` etc.
- [ ] Add `ComponentSet.Has(ComponentType) bool` for existence checks
- [ ] Provide `ComponentSet.Get[T](ComponentType) (T, bool)` for generic access

### 5. Performance Optimizations

- [ ] Cache frequently accessed components
- [ ] Implement component access pools to reduce allocations
- [ ] Add batch query operations for multiple entities
- [ ] Optimize component lookup with better data structures

## Proposed API Design

### Safe Accessors (Zero Values)

```go
// Returns zero value instead of error - no error handling needed
pos := ecs.GetPositionSafe(id)           // Returns gruid.Point{}
health := ecs.GetHealthSafe(id)          // Returns Health{}
name := ecs.GetNameSafe(id)              // Returns ""
renderable := ecs.GetRenderableSafe(id)  // Returns Renderable{}
```

### Optional Pattern

```go
// Optional pattern for explicit null handling
posOpt := ecs.GetPositionOpt(id)
if posOpt.IsSome() {
    pos := posOpt.Unwrap()
    // Use pos
}

// Or with default
pos := ecs.GetPositionOpt(id).UnwrapOr(gruid.Point{X: 0, Y: 0})
```

### Query Builder

```go
// Fluent query interface
entities := ecs.Query().
    WithPosition().
    WithHealth().
    Where(func(id EntityID, pos gruid.Point, health Health) bool {
        return health.CurrentHP > 0
    }).
    Execute()

// Batch operations
ecs.Query().WithPosition().WithRenderable().
    ForEach(func(id EntityID, pos gruid.Point, r Renderable) {
        // Process each entity
    })
```

### Component Set Access

```go
// Access multiple components at once
components := ecs.GetComponents(id)
if components.Has(CPosition) && components.Has(CHealth) {
    pos := components.Position()
    health := components.Health()
    // Use both components
}
```

## Implementation Plan

### Phase 1: Safe Accessors

- [ ] Implement safe accessor methods for all component types
- [ ] Add comprehensive tests for safe accessors
- [ ] Update documentation with usage examples

### Phase 2: Optional Pattern

- [ ] Create `Option[T]` generic type
- [ ] Implement optional accessor methods
- [ ] Add fluent methods for Option type

### Phase 3: Query Builder

- [ ] Design and implement Query struct
- [ ] Add fluent interface methods
- [ ] Implement filtering and batch operations

### Phase 4: Migration and Optimization

- [ ] Update existing code to use new patterns where beneficial
- [ ] Performance testing and optimization
- [ ] Complete documentation and examples

## Acceptance Criteria

- [ ] New query interface reduces error handling code by at least 50%
- [ ] All existing functionality remains intact (backward compatibility)
- [ ] Performance is maintained or improved
- [ ] Type safety is preserved or enhanced
- [ ] Documentation is comprehensive and clear
- [ ] Migration path is clearly defined
- [ ] All tests pass with new implementation

## Migration Strategy

1. **Additive Changes**: New methods alongside existing ones
2. **Gradual Migration**: Update code incrementally
3. **Deprecation Warnings**: Mark old patterns as deprecated
4. **Performance Testing**: Ensure no regressions

## Notes

- Must maintain Go 1.24.3 compatibility
- Should follow Go best practices for API design
- Consider using generics effectively for type safety
- Evaluate builder pattern vs. functional options pattern
- Preserve thread safety with proper locking
