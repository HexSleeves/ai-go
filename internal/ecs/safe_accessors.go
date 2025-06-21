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

// GetInventorySafe returns the Inventory component for an entity, or zero value if not found.
func (ecs *ECS) GetInventorySafe(id EntityID) components.Inventory {
	comp, _ := ecs.GetInventory(id)
	return comp
}

// GetEquipmentSafe returns the Equipment component for an entity, or zero value if not found.
func (ecs *ECS) GetEquipmentSafe(id EntityID) components.Equipment {
	comp, _ := ecs.GetEquipment(id)
	return comp
}

// GetItemPickupSafe returns the ItemPickup component for an entity, or zero value if not found.
func (ecs *ECS) GetItemPickupSafe(id EntityID) components.ItemPickup {
	comp, _ := ecs.GetItemPickup(id)
	return comp
}

// GetAIComponentSafe returns the AIComponent for an entity, or zero value if not found.
func (ecs *ECS) GetAIComponentSafe(id EntityID) components.AIComponent {
	comp, _ := ecs.GetAIComponent(id)
	return comp
}

// GetStatsSafe returns the Stats component for an entity, or zero value if not found.
func (ecs *ECS) GetStatsSafe(id EntityID) components.Stats {
	comp, _ := ecs.GetStats(id)
	return comp
}

// GetExperienceSafe returns the Experience component for an entity, or zero value if not found.
func (ecs *ECS) GetExperienceSafe(id EntityID) components.Experience {
	comp, _ := ecs.GetExperience(id)
	return comp
}

// GetSkillsSafe returns the Skills component for an entity, or zero value if not found.
func (ecs *ECS) GetSkillsSafe(id EntityID) components.Skills {
	comp, _ := ecs.GetSkills(id)
	return comp
}

// GetCombatSafe returns the Combat component for an entity, or zero value if not found.
func (ecs *ECS) GetCombatSafe(id EntityID) components.Combat {
	comp, _ := ecs.GetCombat(id)
	return comp
}

// GetManaSafe returns the Mana component for an entity, or zero value if not found.
func (ecs *ECS) GetManaSafe(id EntityID) components.Mana {
	comp, _ := ecs.GetMana(id)
	return comp
}

// GetStaminaSafe returns the Stamina component for an entity, or zero value if not found.
func (ecs *ECS) GetStaminaSafe(id EntityID) components.Stamina {
	comp, _ := ecs.GetStamina(id)
	return comp
}

// GetStatusEffectsSafe returns the StatusEffects component for an entity, or zero value if not found.
func (ecs *ECS) GetStatusEffectsSafe(id EntityID) components.StatusEffects {
	comp, _ := ecs.GetStatusEffects(id)
	return comp
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

// HasInventorySafe returns true if the entity has an Inventory component.
func (ecs *ECS) HasInventorySafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CInventory)
}

// HasEquipmentSafe returns true if the entity has an Equipment component.
func (ecs *ECS) HasEquipmentSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CEquipment)
}

// HasItemPickupSafe returns true if the entity has an ItemPickup component.
func (ecs *ECS) HasItemPickupSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CItemPickup)
}

// HasAIComponentSafe returns true if the entity has an AIComponent.
func (ecs *ECS) HasAIComponentSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CAIComponent)
}

// HasStatsSafe returns true if the entity has a Stats component.
func (ecs *ECS) HasStatsSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CStats)
}

// HasExperienceSafe returns true if the entity has an Experience component.
func (ecs *ECS) HasExperienceSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CExperience)
}

// HasSkillsSafe returns true if the entity has a Skills component.
func (ecs *ECS) HasSkillsSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CSkills)
}

// HasCombatSafe returns true if the entity has a Combat component.
func (ecs *ECS) HasCombatSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CCombat)
}

// HasManaSafe returns true if the entity has a Mana component.
func (ecs *ECS) HasManaSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CMana)
}

// HasStaminaSafe returns true if the entity has a Stamina component.
func (ecs *ECS) HasStaminaSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CStamina)
}

// HasStatusEffectsSafe returns true if the entity has a StatusEffects component.
func (ecs *ECS) HasStatusEffectsSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CStatusEffects)
}

// GetPathfindingComponentSafe returns the PathfindingComponent for an entity, or nil if not found.
func (ecs *ECS) GetPathfindingComponentSafe(id EntityID) *components.PathfindingComponent {
	comp, _ := ecs.GetPathfindingComponent(id)
	return comp
}

// HasPathfindingComponentSafe returns true if the entity has a PathfindingComponent.
func (ecs *ECS) HasPathfindingComponentSafe(id EntityID) bool {
	return ecs.HasComponent(id, components.CPathfindingComponent)
}
