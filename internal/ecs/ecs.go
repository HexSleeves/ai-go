package ecs

import (
	"fmt"
	"reflect"
	"sync"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/sirupsen/logrus"
)

// EntityID represents a unique identifier for an entity.
type EntityID int

// ECS manages entities and their components.
type ECS struct {
	nextEntityID EntityID
	mu           sync.RWMutex
	entities     map[EntityID]struct{}                         // Just tracks valid entities
	components   map[components.ComponentType]map[EntityID]any // Generic component storage
}

// NewECS creates and initializes a new ECS.
func NewECS() *ECS {
	return &ECS{
		nextEntityID: 1,
		entities:     make(map[EntityID]struct{}),
		components:   make(map[components.ComponentType]map[EntityID]any),
	}
}

// AddEntity creates a new entity and returns its ID.
func (ecs *ECS) AddEntity() EntityID {
	ecs.mu.Lock()
	defer ecs.mu.Unlock()

	id := ecs.nextEntityID
	ecs.entities[id] = struct{}{}
	ecs.nextEntityID++
	return id
}

// AddEntityWithID creates an entity with a specific ID (used for save/load).
// Returns an error if the ID already exists.
func (ecs *ECS) AddEntityWithID(id EntityID) error {
	ecs.mu.Lock()
	defer ecs.mu.Unlock()

	if _, exists := ecs.entities[id]; exists {
		return fmt.Errorf("entity with ID %d already exists", id)
	}

	ecs.entities[id] = struct{}{}

	// Update nextEntityID if necessary to avoid conflicts
	if id >= ecs.nextEntityID {
		ecs.nextEntityID = id + 1
	}

	return nil
}

// UpdateComponent provides safe concurrent access to a component for mutation.
// The updateFunc receives a pointer to a copy of the component and can modify it.
// The modified component is then stored back to the ECS.
// Returns an error if the entity doesn't exist or doesn't have the component.
//
// NOTE: This generic method has limitations with Go's type system. For better type safety
// and usability, prefer type-specific methods like UpdateAIComponent.
func (ecs *ECS) UpdateComponent(entityID EntityID, componentType components.ComponentType, updateFunc func(component interface{}) error) error {
	ecs.mu.Lock()
	defer ecs.mu.Unlock()

	// Check if entity exists
	if _, exists := ecs.entities[entityID]; !exists {
		return fmt.Errorf("entity %d does not exist", entityID)
	}

	// Check if component exists
	componentMap, exists := ecs.components[componentType]
	if !exists {
		return fmt.Errorf("component type %s not registered", componentType)
	}

	component, exists := componentMap[entityID]
	if !exists {
		return fmt.Errorf("entity %d does not have component %s", entityID, componentType)
	}

	// Create a pointer to the component for mutation
	componentPtr := &component

	// Call the update function with the component pointer
	err := updateFunc(componentPtr)
	if err != nil {
		return err
	}

	// Store the modified component back
	componentMap[entityID] = *componentPtr
	return nil
}

// UpdateAIComponent provides type-safe access to update an AI component.
// The updateFunc receives a pointer to the AIComponent and can modify it directly.
// This method ensures proper concurrency control and automatic persistence.
//
// This is the recommended approach for component mutation as it:
// - Provides type safety
// - Handles concurrency automatically
// - Persists changes automatically
// - Prevents the state loss bug that occurred with pointer-to-copy patterns
func (ecs *ECS) UpdateAIComponent(entityID EntityID, updateFunc func(*components.AIComponent) error) error {
	ecs.mu.Lock()
	defer ecs.mu.Unlock()

	// Check if entity exists
	if _, exists := ecs.entities[entityID]; !exists {
		return fmt.Errorf("entity %d does not exist", entityID)
	}

	// Check if component exists
	componentMap, exists := ecs.components[components.CAIComponent]
	if !exists {
		return fmt.Errorf("AI component type not registered")
	}

	component, exists := componentMap[entityID]
	if !exists {
		return fmt.Errorf("entity %d does not have AI component", entityID)
	}

	// Type assert to AIComponent
	aiComp, ok := component.(components.AIComponent)
	if !ok {
		return fmt.Errorf("component is not an AIComponent")
	}

	// Call the update function with a pointer to the component
	err := updateFunc(&aiComp)
	if err != nil {
		return err
	}

	// Store the modified component back
	componentMap[entityID] = aiComp
	return nil
}

// RemoveEntity removes an entity and all its components.
func (ecs *ECS) RemoveEntity(id EntityID) {
	ecs.mu.Lock()
	defer ecs.mu.Unlock()
	delete(ecs.entities, id)
	for _, components := range ecs.components {
		delete(components, id)
	}
}

// EntityExists checks if an entity exists.
func (ecs *ECS) EntityExists(id EntityID) bool {
	ecs.mu.RLock()
	defer ecs.mu.RUnlock()
	_, ok := ecs.entities[id]
	return ok
}

// HasComponent checks if an entity has a specific component.
func (ecs *ECS) HasComponent(id EntityID, compType components.ComponentType) bool {
	_, exists := ecs.getComponent(id, compType)
	return exists
}

// AddComponent adds or updates a component for an entity.
func (ecs *ECS) AddComponent(id EntityID, compType components.ComponentType, component any) {
	if !ecs.EntityExists(id) {
		logrus.Debugf("Warning: Attempted to add component %s to non-existent entity %d", compType, id)
		return
	}

	ecs.mu.Lock()
	defer ecs.mu.Unlock()
	if ecs.components[compType] == nil {
		ecs.components[compType] = make(map[EntityID]any)
	}

	ecs.components[compType][id] = component
}

// AddComponents adds multiple components to an entity at once.
func (ecs *ECS) AddComponents(id EntityID, comps ...any) {
	if !ecs.EntityExists(id) {
		logrus.Debugf("Warning: Attempted to add components to non-existent entity %d", id)
		return
	}

	for _, comp := range comps {
		compType := reflect.TypeOf(comp)
		if compType.Kind() == reflect.Ptr && compType.Elem().Kind() != reflect.Interface {
			// Keep as pointer for FOV which is stored as *FOV
			if compType == reflect.TypeOf((*components.FOV)(nil)) {
				ecs.AddComponent(id, components.CFOV, comp)
				continue
			}
		}

		// Position component
		if compType == reflect.TypeOf(gruid.Point{}) {
			ecs.AddComponent(id, components.CPosition, comp)
			continue
		}

		// Name component
		if compType == reflect.TypeOf(components.Name{}) {
			ecs.AddComponent(id, components.CName, comp)
			continue
		}

		// All other components
		found := false
		for ct, t := range components.TypeToComponent {
			if t == compType {
				ecs.AddComponent(id, ct, comp)
				found = true
				break
			}
		}

		if !found {
			logrus.Warnf("Unknown component type %T for entity %d", comp, id)
		}
	}
}

// RemoveComponent removes a component from an entity.
func (ecs *ECS) RemoveComponent(id EntityID, compType components.ComponentType) {
	ecs.mu.Lock()
	defer ecs.mu.Unlock()

	if components, ok := ecs.components[compType]; ok {
		delete(components, id)
	}
}

// RemoveComponents removes multiple components from an entity.
func (ecs *ECS) RemoveComponents(id EntityID, compTypes ...components.ComponentType) {
	for _, compType := range compTypes {
		ecs.RemoveComponent(id, compType)
	}
}

// GetComponentTyped retrieves a component for an entity and returns it as the concrete type T.
// Returns the zero value of T and false if the component doesn't exist or if type assertion fails.
func GetComponentTyped[T any](ecs *ECS, id EntityID, compType components.ComponentType) (T, bool) {
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

// --- Helper Functions ---

// GetComponent retrieves a component for an entity.
func (ecs *ECS) getComponent(id EntityID, compType components.ComponentType) (any, bool) {
	ecs.mu.RLock()
	defer ecs.mu.RUnlock()

	if componentMap, ok := ecs.components[compType]; ok {
		if component, exists := componentMap[id]; exists {
			return component, true
		}
	}

	return nil, false
}
