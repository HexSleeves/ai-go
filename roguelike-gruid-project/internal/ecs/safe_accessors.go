package ecs

import (
	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

// Safe accessor methods that return zero values instead of errors for missing components.
// These methods eliminate the need for error handling in most cases where a default value is acceptable.

// GetPositionSafe returns the position component for an entity, or zero Point if not found.
func (ecs *ECS) GetPositionSafe(id EntityID) gruid.Point {
	pos, _ := ecs.GetPosition(id)
	return pos
}

// GetHealthSafe returns the health component for an entity, or zero Health if not found.
func (ecs *ECS) GetHealthSafe(id EntityID) components.Health {
	health, _ := ecs.GetHealth(id)
	return health
}

// GetRenderableSafe returns the renderable component for an entity, or zero Renderable if not found.
func (ecs *ECS) GetRenderableSafe(id EntityID) components.Renderable {
	renderable, _ := ecs.GetRenderable(id)
	return renderable
}

// GetNameSafe returns the name component for an entity, or empty string if not found.
func (ecs *ECS) GetNameSafe(id EntityID) string {
	name, _ := ecs.GetName(id)
	return name
}

// GetFOVSafe returns the FOV component for an entity, or nil if not found.
func (ecs *ECS) GetFOVSafe(id EntityID) *components.FOV {
	fov, _ := ecs.GetFOV(id)
	return fov
}

// GetTurnActorSafe returns the TurnActor component for an entity, or zero TurnActor if not found.
func (ecs *ECS) GetTurnActorSafe(id EntityID) components.TurnActor {
	actor, _ := ecs.GetTurnActor(id)
	return actor
}

// GetPlayerTagSafe returns the PlayerTag component for an entity, or zero PlayerTag if not found.
func (ecs *ECS) GetPlayerTagSafe(id EntityID) components.PlayerTag {
	tag, _ := ecs.GetPlayerTag(id)
	return tag
}

// GetAITagSafe returns the AITag component for an entity, or zero AITag if not found.
func (ecs *ECS) GetAITagSafe(id EntityID) components.AITag {
	tag, _ := ecs.GetAITag(id)
	return tag
}

// GetCorpseTagSafe returns the CorpseTag component for an entity, or zero CorpseTag if not found.
func (ecs *ECS) GetCorpseTagSafe(id EntityID) components.CorpseTag {
	tag, _ := ecs.GetCorpseTag(id)
	return tag
}

// HasPositionSafe checks if an entity has a position component (convenience method).
func (ecs *ECS) HasPositionSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CPosition)
}

// HasHealthSafe checks if an entity has a health component (convenience method).
func (ecs *ECS) HasHealthSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CHealth)
}

// HasRenderableSafe checks if an entity has a renderable component (convenience method).
func (ecs *ECS) HasRenderableSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CRenderable)
}

// HasFOVSafe checks if an entity has an FOV component (convenience method).
func (ecs *ECS) HasFOVSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CFOV)
}
