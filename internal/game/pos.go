package game

import (
	"fmt"
	"log/slog"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/sirupsen/logrus"
)

// UpdateEntityPosition updates an entity's position and maintains the spatial grid
func (g *Game) UpdateEntityPosition(id ecs.EntityID, oldPos, newPos gruid.Point) {
	// Update the spatial grid
	g.spatialGrid.Move(id, oldPos, newPos)
}

// EntityBump attempts to move the entity with the given ID by the delta.
// It checks for map boundaries and collisions with other entities.
// It returns true if the entity successfully moved, false otherwise (due to wall or collision).
func (g *Game) EntityBump(entityID ecs.EntityID, delta gruid.Point) (moved bool, err error) {
	// Use safe accessor - no error handling needed!
	currentPos := g.ecs.GetPositionSafe(entityID)

	// Check if entity actually has a position (zero value check)
	if currentPos == (gruid.Point{}) && !g.ecs.HasPositionSafe(entityID) {
		return false, fmt.Errorf("entity %d has no position component", entityID)
	}

	newPos := currentPos.Add(delta)

	// Check map bounds and walkability first
	if !g.dungeon.InBounds(newPos) {
		return false, fmt.Errorf("entity %d attempted to move out of bounds to %v", entityID, newPos)
	}

	if !g.dungeon.isWalkable(newPos) {
		return false, fmt.Errorf("entity %d attempted to move into wall at %v", entityID, newPos)
	}

	// Check for collision with other entities at the target position
	for _, otherID := range g.ecs.GetEntitiesAtWithComponents(newPos, components.CBlocksMovement) {
		if otherID == entityID {
			continue // Don't interact with self
		}

		// Check if the target entity has health (i.e., is attackable)
		if g.ecs.HasComponent(otherID, components.CHealth) {
			// Target is attackable. Queue an AttackAction for the bumping entity.
			logrus.Debugf("Entity %d bumping into attackable entity %d. Queuing AttackAction.", entityID, otherID)

			// Use safe accessor - no error handling needed!
			actor := g.ecs.GetTurnActorSafe(entityID)

			// Check if entity actually has a TurnActor component
			if !g.ecs.HasComponent(entityID, components.CTurnActor) {
				return false, fmt.Errorf("entity %d cannot perform actions (missing TurnActor)", entityID)
			}

			// Create and queue the attack action
			attackAction := AttackAction{
				AttackerID: entityID,
				TargetID:   otherID,
			}
			actor.AddAction(attackAction)

			// Return moved=false because the bump resulted in an attack, not movement.
			// The turn cost will be handled by the AttackAction itself when executed.
			return false, nil
		} else {
			// Bumped into a non-attackable entity (e.g., another player, item, scenery)
			slog.Debug("Entity %d bumped into non-attackable entity %d.", entityID, otherID)
			return false, nil // Block movement
		}
	}

	// If no collision, move the entity
	err = g.ecs.MoveEntity(entityID, newPos)
	if err != nil {
		return false, fmt.Errorf("failed to move entity %d: %w", entityID, err)
	}

	// Update the spatial grid
	g.UpdateEntityPosition(entityID, currentPos, newPos)

	// Successfully moved
	return true, nil
}
