package ecs

import (
	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/sirupsen/logrus"
)

// Helper method that doesn't acquire a lock (for use in methods that already have the lock)
func (ecs *ECS) entityExists(id EntityID) bool {
	_, ok := ecs.entities[id]
	return ok
}

// EntitiesAt returns a slice of EntityIDs located at the given point.
func (ecs *ECS) EntitiesAt(p gruid.Point) []EntityID {
	ecs.mu.RLock()
	defer ecs.mu.RUnlock()

	var ids []EntityID
	if posMap, ok := ecs.components[components.CPosition]; ok {
		for id, comp := range posMap {
			if pos, ok := comp.(gruid.Point); ok && pos == p {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func (ecs *ECS) GetEntitiesAtWithComponents(p gruid.Point, compType components.ComponentType) []EntityID {
	ecs.mu.RLock()
	defer ecs.mu.RUnlock()

	ids := ecs.EntitiesAt(p)
	results := make([]EntityID, 0, len(ids))
	for _, id := range ids {
		if ecs.HasComponent(id, compType) {
			results = append(results, id)
		}
	}

	return results
}

// GetAllEntities returns all managed entity IDs.
func (ecs *ECS) GetAllEntities() []EntityID {
	ecs.mu.RLock()
	defer ecs.mu.RUnlock()
	ids := make([]EntityID, 0, len(ecs.entities))
	for id := range ecs.entities {
		ids = append(ids, id)
	}
	return ids
}

// GetEntitiesWithComponent returns all entities that have a specific component.
func (ecs *ECS) GetEntitiesWithComponent(compType components.ComponentType) []EntityID {
	ecs.mu.RLock()
	defer ecs.mu.RUnlock()

	var entities []EntityID
	if components, ok := ecs.components[compType]; ok {
		entities = make([]EntityID, 0, len(components))
		for id := range components {
			entities = append(entities, id)
		}
	}
	return entities
}

// GetEntitiesWithComponents returns entities that have all specified components.
func (ecs *ECS) GetEntitiesWithComponents(compTypes ...components.ComponentType) []EntityID {
	if len(compTypes) == 0 {
		return nil
	}
	entities := ecs.GetEntitiesWithComponent(compTypes[0])
	if len(entities) == 0 {
		return nil
	}
	var result []EntityID
	for _, id := range entities {
		hasAll := true
		for _, ct := range compTypes[1:] {
			if !ecs.HasComponent(id, ct) {
				hasAll = false
				break
			}
		}
		if hasAll {
			result = append(result, id)
		}
	}
	return result
}

////////////////////////////////////////////////////////////
// --- Specific queries --- //
////////////////////////////////////////////////////////////

type PositionedRenderableEntity struct {
	ID         EntityID
	Position   gruid.Point
	Renderable components.Renderable
}

// GetEntitiesWithPositionAndRenderable queries entities having both Position and Renderable components.
func (ecs *ECS) GetEntitiesWithPositionAndRenderable() []PositionedRenderableEntity {
	entities := ecs.GetEntitiesWithComponents(components.CPosition, components.CRenderable)
	result := make([]PositionedRenderableEntity, 0, len(entities))

	for _, id := range entities {
		pos, posOk := ecs.GetPosition(id)
		renderable, renderableOk := ecs.GetRenderable(id)
		if !posOk {
			logrus.Errorf("Entity %d has Position component but no position", id)
		}
		if !renderableOk {
			logrus.Errorf("Entity %d has Renderable component but no renderable", id)
		}

		result = append(result, PositionedRenderableEntity{
			ID:         id,
			Position:   pos,
			Renderable: renderable,
		})
	}

	return result
}

type PositionedFOVEntity struct {
	ID       EntityID
	Position gruid.Point
	FOV      *components.FOV
}

// GetEntitiesWithPositionAndFOV queries entities having both Position and FOV components.
func (ecs *ECS) GetEntitiesWithPositionAndFOV() []PositionedFOVEntity {
	entities := ecs.GetEntitiesWithComponents(components.CPosition, components.CFOV)
	result := make([]PositionedFOVEntity, 0, len(entities))

	for _, id := range entities {
		pos, posOk := ecs.GetPosition(id)
		fov, fovOk := ecs.GetFOV(id)

		if !posOk {
			logrus.Errorf("Entity %d has Position component but no position", id)
		}
		if !fovOk {
			logrus.Errorf("Entity %d has FOV component but no FOV", id)
		}

		result = append(result, PositionedFOVEntity{
			ID:       id,
			Position: pos,
			FOV:      fov,
		})
	}

	return result
}

// GetPosition returns the position component for an entity.
func (ecs *ECS) GetPosition(id EntityID) (gruid.Point, bool) {
	return GetComponentTyped[gruid.Point](ecs, id, components.CPosition)
}

// GetName returns the name component for an entity.
func (ecs *ECS) GetName(id EntityID) (string, bool) {
	name, ok := GetComponentTyped[components.Name](ecs, id, components.CName)
	return name.Name, ok
}

// GetRenderable returns the renderable component for an entity.
func (ecs *ECS) GetRenderable(id EntityID) (components.Renderable, bool) {
	return GetComponentTyped[components.Renderable](ecs, id, components.CRenderable)
}

// GetHealth returns the health component for an entity.
func (ecs *ECS) GetHealth(id EntityID) (components.Health, bool) {
	return GetComponentTyped[components.Health](ecs, id, components.CHealth)
}

// GetFOV returns the FOV component for an entity.
func (ecs *ECS) GetFOV(id EntityID) (*components.FOV, bool) {
	return GetComponentTyped[*components.FOV](ecs, id, components.CFOV)
}

// GetTurnActor returns the TurnActor component for an entity.
func (ecs *ECS) GetTurnActor(id EntityID) (components.TurnActor, bool) {
	return GetComponentTyped[components.TurnActor](ecs, id, components.CTurnActor)
}

// GetInventory returns the Inventory component for an entity.
func (ecs *ECS) GetInventory(id EntityID) (components.Inventory, bool) {
	return GetComponentTyped[components.Inventory](ecs, id, components.CInventory)
}

// GetEquipment returns the Equipment component for an entity.
func (ecs *ECS) GetEquipment(id EntityID) (components.Equipment, bool) {
	return GetComponentTyped[components.Equipment](ecs, id, components.CEquipment)
}

// GetItemPickup returns the ItemPickup component for an entity.
func (ecs *ECS) GetItemPickup(id EntityID) (components.ItemPickup, bool) {
	return GetComponentTyped[components.ItemPickup](ecs, id, components.CItemPickup)
}

// GetAIComponent returns the AIComponent for an entity.
func (ecs *ECS) GetAIComponent(id EntityID) (components.AIComponent, bool) {
	return GetComponentTyped[components.AIComponent](ecs, id, components.CAIComponent)
}

// GetStats returns the Stats component for an entity.
func (ecs *ECS) GetStats(id EntityID) (components.Stats, bool) {
	return GetComponentTyped[components.Stats](ecs, id, components.CStats)
}

// GetExperience returns the Experience component for an entity.
func (ecs *ECS) GetExperience(id EntityID) (components.Experience, bool) {
	return GetComponentTyped[components.Experience](ecs, id, components.CExperience)
}

// GetSkills returns the Skills component for an entity.
func (ecs *ECS) GetSkills(id EntityID) (components.Skills, bool) {
	return GetComponentTyped[components.Skills](ecs, id, components.CSkills)
}

// GetCombat returns the Combat component for an entity.
func (ecs *ECS) GetCombat(id EntityID) (components.Combat, bool) {
	return GetComponentTyped[components.Combat](ecs, id, components.CCombat)
}

// GetMana returns the Mana component for an entity.
func (ecs *ECS) GetMana(id EntityID) (components.Mana, bool) {
	return GetComponentTyped[components.Mana](ecs, id, components.CMana)
}

// GetStamina returns the Stamina component for an entity.
func (ecs *ECS) GetStamina(id EntityID) (components.Stamina, bool) {
	return GetComponentTyped[components.Stamina](ecs, id, components.CStamina)
}

// GetStatusEffects returns the StatusEffects component for an entity.
func (ecs *ECS) GetStatusEffects(id EntityID) (components.StatusEffects, bool) {
	return GetComponentTyped[components.StatusEffects](ecs, id, components.CStatusEffects)
}

// GetPlayerTag returns the PlayerTag component for an entity.
func (ecs *ECS) GetPlayerTag(id EntityID) (components.PlayerTag, bool) {
	return GetComponentTyped[components.PlayerTag](ecs, id, components.CPlayerTag)
}

// GetAITag returns the AITag component for an entity.
func (ecs *ECS) GetAITag(id EntityID) (components.AITag, bool) {
	return GetComponentTyped[components.AITag](ecs, id, components.CAITag)
}

// GetCorpseTag returns the CorpseTag component for an entity.
func (ecs *ECS) GetCorpseTag(id EntityID) (components.CorpseTag, bool) {
	return GetComponentTyped[components.CorpseTag](ecs, id, components.CCorpseTag)
}
