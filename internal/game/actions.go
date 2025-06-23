package game

import (
	"fmt"
	"log/slog"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ui"
)

// GameAction is an interface for actions that can be performed in the game.
type GameAction interface {
	Execute(g *Game) (cost uint, err error)
}

type WaitAction struct {
	EntityID ecs.EntityID
}

func (a WaitAction) Execute(g *Game) (cost uint, err error) {
	return 100, nil // Standard wait cost
}

type MoveAction struct {
	Direction gruid.Point
	EntityID  ecs.EntityID
}

// Execute performs the move action, returning the time cost and any error.
func (a MoveAction) Execute(g *Game) (cost uint, err error) {
	again, err := g.EntityBump(a.EntityID, a.Direction)
	if err != nil {
		return 0, err // No cost if error occurred
	}

	if !again {
		// Bumped into something, action didn't fully succeed in moving
		return 0, nil // No time cost for a bump
	}

	return 100, nil // Standard move cost
}

// AttackAction represents an entity attacking another entity.
type AttackAction struct {
	AttackerID ecs.EntityID
	TargetID   ecs.EntityID
}

// Execute performs the attack action.
func (a AttackAction) Execute(g *Game) (cost uint, err error) {
	// Use safe accessors - no error handling needed!
	attackerName := g.ecs.GetNameSafe(a.AttackerID)
	targetName := g.ecs.GetNameSafe(a.TargetID)

	// Use optional pattern for explicit null handling of health
	targetHealthOpt := g.ecs.GetHealthOpt(a.TargetID)
	if targetHealthOpt.IsNone() {
		// Target might have died between action queuing and execution
		slog.Debug("Attacker tries to attack target, but target has no health component", "attacker", attackerName, "attackerId", a.AttackerID, "target", targetName, "targetId", a.TargetID)
		return 0, fmt.Errorf("target %d has no health", a.TargetID)
	}

	targetHealth := targetHealthOpt.Unwrap()

	// --- Basic Damage Calculation ---
	damage := 1 // Simple fixed damage for now
	targetHealth.CurrentHP -= damage

	// Track damage statistics
	if a.AttackerID == g.PlayerID {
		g.AddDamageDealt(damage)
	} else if a.TargetID == g.PlayerID {
		g.AddDamageTaken(damage)
	}

	// Determine message color based on who is attacking
	var msgColor gruid.Color
	if a.AttackerID == g.PlayerID {
		msgColor = ui.ColorPlayerAttack // Define in ui/color.go
	} else if a.TargetID == g.PlayerID {
		msgColor = ui.ColorEnemyAttack // Define in ui/color.go
	} else {
		msgColor = ui.ColorNeutralAttack // Define in ui/color.go
	}
	g.log.AddMessagef(msgColor, "%s attacks %s for %d damage.", attackerName, targetName, damage)

	slog.Info("Combat action", "attacker", attackerName, "attackerId", a.AttackerID, "target", targetName, "targetId", a.TargetID, "damage", damage, "targetHP", targetHealth.CurrentHP, "targetMaxHP", targetHealth.MaxHP)
	g.ecs.AddComponent(a.TargetID, components.CHealth, targetHealth)

	// Trigger combat event
	g.TriggerCombatEvent(a.AttackerID, a.TargetID, damage, false)

	// Check for death (CurrentHP <= 0) and handle it
	if targetHealth.IsDead() {
		g.handleEntityDeath(a.TargetID, targetName, a.AttackerID)
	}

	return 100, nil // Standard attack cost
}

// handleEntityDeath handles an entity's death, either removing it completely
// or turning it into a corpse (the preferred option)
func (g *Game) handleEntityDeath(entityID ecs.EntityID, entityName string, killerID ecs.EntityID) {
	g.log.AddMessagef(ui.ColorDeath, "%s dies!", entityName)
	slog.Info("Entity has died", "entityName", entityName, "entityId", entityID)

	// Trigger death event
	g.TriggerDeathEvent(entityID, killerID)

	if entityID == g.PlayerID {
		g.log.AddMessagef(ui.ColorCritical, "You died! Game over!")
		slog.Info("Player has died. Game over!")
		g.setGameOverState()
		return
	}

	// Track monster kill statistics
	if killerID == g.PlayerID && g.ecs.HasComponent(entityID, components.CAITag) {
		g.IncrementMonstersKilled()
	}

	// Award experience to the killer if it exists
	if killerID != 0 && g.ecs.EntityExists(killerID) {
		expSystem := NewExperienceSystem(g)
		xpReward := expSystem.GetExperienceForKill(killerID, entityID)
		expSystem.AwardExperience(killerID, xpReward)
	}

	// Turn entity into a corpse
	g.ecs.RemoveComponents(entityID,
		components.CTurnActor,
		components.CAITag,
		components.CBlocksMovement,
		components.CHealth,
	)

	g.ecs.AddComponents(entityID,
		components.Renderable{Glyph: '%', Color: ui.ColorCorpse},
		components.CorpseTag{},
	)

	// Remove from turn queue
	g.turnQueue.Remove(entityID)
	// Remove from spatial grid using safe accessor
	pos := g.ecs.GetPositionSafe(entityID)
	g.spatialGrid.Remove(entityID, pos)
}

// Additional Action Types
// Note: Inventory actions (Pickup, Drop, UseItem, Equip) are defined in inventory_actions.go
