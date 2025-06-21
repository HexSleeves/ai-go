package game

import (
	"fmt"
	"log/slog"
	"math/rand"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

// monstersTurn handles AI turns for all monsters in the game.
func (g *Game) monstersTurn() {
	aiEntities := g.ecs.GetEntitiesWithComponent(components.CAITag)
	for _, id := range aiEntities {

		// Use safe accessor - no error handling needed!
		actor := g.ecs.GetTurnActorSafe(id)

		// Skip if entity doesn't have TurnActor component
		if !g.ecs.HasComponent(id, components.CTurnActor) {
			continue
		}

		if !actor.IsAlive() {
			continue
		}

		if actor.PeekNextAction() != nil {
			continue
		}

		// Check monster's FOV using safe accessor
		monsterFOVComp := g.ecs.GetFOVSafe(id)
		if monsterFOVComp == nil {
			slog.Error("Monster entity missing FOV component in monstersTurn", "id", id)
			continue
		}

		// Try to use advanced AI if available, otherwise fall back to basic AI
		var action GameAction
		if g.ecs.HasAIComponentSafe(id) {
			action = g.AdvancedMonsterAI(id)
		} else {
			// Fallback to basic AI
			// Check if monster can see player
			playerPos := g.GetPlayerPosition()
			if monsterFOVComp.IsVisible(playerPos, g.dungeon.Width) {
				slog.Info("Monster can see player, attacking", "id", id)

				// Use safe accessor - no error handling needed!
				pos := g.ecs.GetPositionSafe(id)

				// Skip if entity doesn't have position component
				if !g.ecs.HasPositionSafe(id) {
					slog.Error("Monster entity missing position in monstersTurn", "id", id)
					continue
				}

				action = MoveAction{Direction: playerPos.Sub(pos), EntityID: id}
			} else {
				moveOrWait := rand.Intn(2)
				if moveOrWait == 0 {
					var err error
					action, err = moveMonster(g, id)
					if err != nil {
						slog.Debug("Failed to move monster", "id", id, "error", err)
						continue
					}
				} else {
					action = WaitAction{EntityID: id}
				}
			}
		}

		actor.AddAction(action)
	}
}

func moveMonster(g *Game, id ecs.EntityID) (GameAction, error) {
	// Use optional pattern for explicit null handling
	posOpt := g.ecs.GetPositionOpt(id)
	if posOpt.IsNone() {
		return nil, fmt.Errorf("entity %d has no position", id)
	}

	pos := posOpt.Unwrap()

	directions := []gruid.Point{
		{X: -1, Y: 0}, // West
		{X: 1, Y: 0},  // East
		{X: 0, Y: -1}, // North
		{X: 0, Y: 1},  // South
	}
	// This is a simple way to randomize the order of directions
	rand.Shuffle(len(directions), func(i, j int) {
		directions[i], directions[j] = directions[j], directions[i]
	})
	var validMove *gruid.Point
	for _, dir := range directions {
		newPos := pos.Add(dir)
		if g.dungeon.isWalkable(newPos) && len(g.ecs.EntitiesAt(newPos)) == 0 { // EntitiesAt already uses new system
			validMove = &dir
			break
		}
	}
	if validMove != nil {
		slog.Debug("AI entity moving in direction", "id", id, "direction", validMove)
		action := MoveAction{
			Direction: *validMove,
			EntityID:  id,
		}
		return action, nil
	} else {
		slog.Debug("AI entity has no valid move, waiting", "id", id)
		return WaitAction{EntityID: id}, nil
	}
}
