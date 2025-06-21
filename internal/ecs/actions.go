package ecs

import (
	"fmt"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

func (ecs *ECS) MoveEntity(id EntityID, p gruid.Point) error {
	// Use AddComponent which handles locking and existence checks internally
	// Check if entity exists first (AddComponent logs warning but doesn't return error)
	ecs.mu.RLock()
	exists := ecs.entityExists(id)
	ecs.mu.RUnlock()
	if !exists {
		return fmt.Errorf("entity %d not found", id)
	}

	ecs.AddComponent(id, components.CPosition, p)
	return nil
}
