package game

import (
	"fmt"
	"math/rand"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/sirupsen/logrus"
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
			logrus.Errorf("Monster entity %d missing FOV component in monstersTurn", id)
			continue
		}

		// Check if monster can see player
		playerPos := g.GetPlayerPosition()
		if monsterFOVComp.IsVisible(playerPos, g.dungeon.Width) {
			logrus.Infof("Monster %d can see player, attacking", id)

			// Use safe accessor - no error handling needed!
			pos := g.ecs.GetPositionSafe(id)

			// Skip if entity doesn't have position component
			if !g.ecs.HasPositionSafe(id) {
				logrus.Errorf("Monster entity %d missing position in monstersTurn", id)
				continue
			}

			actor.AddAction(MoveAction{Direction: playerPos.Sub(pos), EntityID: id})
			continue
		}

		moveOrWait := rand.Intn(2)
		if moveOrWait == 0 {
			action, err := moveMonster(g, id)
			if err != nil {
				logrus.Debugf("Failed to move monster %d: %v", id, err)
				continue
			}
			actor.AddAction(action)
		} else {
			actor.AddAction(WaitAction{EntityID: id})
		}
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
		logrus.Debugf("AI entity %d moving in direction %v", id, validMove)
		action := MoveAction{
			Direction: *validMove,
			EntityID:  id,
		}
		return action, nil
	} else {
		logrus.Debugf("AI entity %d has no valid move, waiting", id)
		return WaitAction{EntityID: id}, nil
	}
}
