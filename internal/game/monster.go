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

		// Check if entity has queued actions and needs more strategic planning
		if actor.PeekNextAction() != nil {
			continue // Entity already has actions queued
		}

		// Generate strategic action sequences instead of single actions
		g.generateActionSequence(id, &actor)
	}
}

// generateActionSequence creates a strategic sequence of actions for an entity
func (g *Game) generateActionSequence(entityID ecs.EntityID, actor *components.TurnActor) {
	// Check monster's FOV using safe accessor
	monsterFOVComp := g.ecs.GetFOVSafe(entityID)
	if monsterFOVComp == nil {
		slog.Error("Monster entity missing FOV component in generateActionSequence", "id", entityID)
		return
	}

	// Try to use advanced AI if available, otherwise fall back to basic AI
	if g.ecs.HasAIComponentSafe(entityID) {
		g.generateAdvancedActionSequence(entityID, actor)
	} else {
		g.generateBasicActionSequence(entityID, actor, monsterFOVComp)
	}

	// Update the TurnActor component with queued actions
	g.ecs.AddComponent(entityID, components.CTurnActor, *actor)
}

// generateAdvancedActionSequence creates sophisticated action sequences using AI components
func (g *Game) generateAdvancedActionSequence(entityID ecs.EntityID, actor *components.TurnActor) {
	// Get AI component by value
	aiComp, hasAI := g.getAIComponent(entityID)
	if !hasAI {
		// Fallback to basic AI
		monsterFOVComp := g.ecs.GetFOVSafe(entityID)
		if monsterFOVComp != nil {
			g.generateBasicActionSequence(entityID, actor, monsterFOVComp)
		}
		return
	}

	pos := g.ecs.GetPositionSafe(entityID)
	health := g.ecs.GetHealthSafe(entityID)
	hasHealth := g.ecs.HasHealthSafe(entityID)
	playerPos := g.GetPlayerPosition()
	distanceToPlayer := manhattanDistance(pos, playerPos)

	// Update AI state based on conditions
	g.updateAIState(entityID, &aiComp, health, hasHealth, distanceToPlayer)

	// Generate action sequences based on current state
	switch aiComp.State {
	case components.AIStateChasing:
		g.generateChaseSequence(entityID, actor, pos, playerPos)
	case components.AIStateFleeing:
		g.generateFleeSequence(entityID, actor, pos, playerPos)
	case components.AIStateSearching:
		g.generateSearchSequence(entityID, actor, &aiComp, pos)
	case components.AIStatePatrolling:
		g.generatePatrolSequence(entityID, actor, &aiComp, pos)
	case components.AIStateAttacking:
		g.generateAttackSequence(entityID, actor, pos)
	default: // AIStateIdle
		g.generateIdleSequence(entityID, actor, &aiComp, pos)
	}

	// Save the updated AI component back to ECS
	g.ecs.AddComponent(entityID, components.CAIComponent, aiComp)
}

// generateBasicActionSequence creates simple action sequences for monsters without AI components
func (g *Game) generateBasicActionSequence(entityID ecs.EntityID, actor *components.TurnActor, monsterFOVComp *components.FOV) {
	playerPos := g.GetPlayerPosition()
	pos := g.ecs.GetPositionSafe(entityID)

	// Skip if entity doesn't have position component
	if !g.ecs.HasPositionSafe(entityID) {
		slog.Error("Monster entity missing position in generateBasicActionSequence", "id", entityID)
		return
	}

	// Check if monster can see player
	if monsterFOVComp.IsVisible(playerPos, g.dungeon.Width) {
		slog.Info("Monster can see player, generating chase sequence", "id", entityID)

		// Generate a sequence of moves towards the player

		// Queue 2-3 moves in the same direction for persistence
		normalizedDirection := getDirectionTowards(pos, playerPos)
		for i := 0; i < min(3, manhattanDistance(pos, playerPos)); i++ {
			actor.AddAction(MoveAction{Direction: normalizedDirection, EntityID: entityID})
		}
	} else {
		// Generate random movement sequence
		directions := []gruid.Point{
			{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
		}

		// Queue 2-3 random actions
		for i := 0; i < rand.Intn(3)+1; i++ {
			if rand.Intn(2) == 0 {
				direction := directions[rand.Intn(len(directions))]
				actor.AddAction(MoveAction{Direction: direction, EntityID: entityID})
			} else {
				actor.AddAction(WaitAction{EntityID: entityID})
			}
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

// Strategic Action Sequence Generators

// generateChaseSequence creates a sequence of actions for pursuing the player
func (g *Game) generateChaseSequence(entityID ecs.EntityID, actor *components.TurnActor, pos, playerPos gruid.Point) {
	distance := manhattanDistance(pos, playerPos)

	// Generate 2-4 chase actions based on distance
	sequenceLength := min(4, max(2, distance/2))

	for i := 0; i < sequenceLength; i++ {
		// Use pathfinding if available, otherwise fall back to simple movement
		var action GameAction
		if g.pathfindingMgr != nil {
			// Update pathfinding to target the player
			g.pathfindingMgr.UpdatePathfinding(entityID, playerPos, StrategyDirect)

			// Get the next move from pathfinding
			direction := g.pathfindingMgr.GetPathfindingMove(entityID)
			if direction != (gruid.Point{}) {
				action = MoveAction{Direction: direction, EntityID: entityID}
			} else {
				// Fallback to simple directional movement
				direction = getDirectionTowards(pos, playerPos)
				action = MoveAction{Direction: direction, EntityID: entityID}
			}
		} else {
			// Fallback to simple directional movement
			direction := getDirectionTowards(pos, playerPos)
			action = MoveAction{Direction: direction, EntityID: entityID}
		}

		actor.AddAction(action)

		// Update position estimate for next iteration
		if moveAction, ok := action.(MoveAction); ok {
			pos = pos.Add(moveAction.Direction)
		}
	}
}

// generateFleeSequence creates a sequence of actions for fleeing from the player
func (g *Game) generateFleeSequence(entityID ecs.EntityID, actor *components.TurnActor, pos, playerPos gruid.Point) {
	// Generate 3-5 flee actions for sustained escape
	sequenceLength := rand.Intn(3) + 3

	for i := 0; i < sequenceLength; i++ {
		var action GameAction
		if g.pathfindingMgr != nil {
			// Calculate a flee target (opposite direction from player)
			fleeDirection := getDirectionAway(pos, playerPos)
			fleeTarget := pos.Add(fleeDirection.Mul(5)) // Flee 5 steps away

			// Clamp to map bounds
			fleeTarget = g.clampToMapBounds(fleeTarget)

			// Update pathfinding to flee target with entity avoidance
			g.pathfindingMgr.UpdatePathfinding(entityID, fleeTarget, StrategyAvoidEntities)

			// Get the next move from pathfinding
			direction := g.pathfindingMgr.GetPathfindingMove(entityID)
			if direction != (gruid.Point{}) {
				action = MoveAction{Direction: direction, EntityID: entityID}
			} else {
				// Fallback to simple directional movement
				direction = getDirectionAway(pos, playerPos)
				action = MoveAction{Direction: direction, EntityID: entityID}
			}
		} else {
			// Fallback to simple directional movement
			direction := getDirectionAway(pos, playerPos)
			action = MoveAction{Direction: direction, EntityID: entityID}
		}

		actor.AddAction(action)

		// Update position estimate for next iteration
		if moveAction, ok := action.(MoveAction); ok {
			pos = pos.Add(moveAction.Direction)
		}
	}
}

// generateSearchSequence creates a sequence of actions for searching for the player
func (g *Game) generateSearchSequence(entityID ecs.EntityID, actor *components.TurnActor, aiComp *components.AIComponent, pos gruid.Point) {
	// Generate 2-3 search actions
	sequenceLength := rand.Intn(2) + 2

	for i := 0; i < sequenceLength; i++ {
		var action GameAction

		if pos == aiComp.LastKnownPlayerPos {
			// Reached last known position, look around randomly
			directions := []gruid.Point{
				{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
			}
			direction := directions[rand.Intn(len(directions))]
			action = MoveAction{Direction: direction, EntityID: entityID}
		} else {
			// Move towards last known player position
			if g.pathfindingMgr != nil {
				g.pathfindingMgr.UpdatePathfinding(entityID, aiComp.LastKnownPlayerPos, StrategyDirect)
				direction := g.pathfindingMgr.GetPathfindingMove(entityID)
				if direction != (gruid.Point{}) {
					action = MoveAction{Direction: direction, EntityID: entityID}
				} else {
					direction = getDirectionTowards(pos, aiComp.LastKnownPlayerPos)
					action = MoveAction{Direction: direction, EntityID: entityID}
				}
			} else {
				direction := getDirectionTowards(pos, aiComp.LastKnownPlayerPos)
				action = MoveAction{Direction: direction, EntityID: entityID}
			}
		}

		actor.AddAction(action)

		// Update position estimate for next iteration
		if moveAction, ok := action.(MoveAction); ok {
			pos = pos.Add(moveAction.Direction)
		}
	}
}

// generatePatrolSequence creates a sequence of actions for patrolling
func (g *Game) generatePatrolSequence(entityID ecs.EntityID, actor *components.TurnActor, aiComp *components.AIComponent, pos gruid.Point) {
	homeDistance := manhattanDistance(pos, aiComp.HomePosition)

	// Generate 2-4 patrol actions
	sequenceLength := rand.Intn(3) + 2

	for i := 0; i < sequenceLength; i++ {
		var action GameAction

		if homeDistance > aiComp.PatrolRadius {
			// Return to home area
			if g.pathfindingMgr != nil {
				g.pathfindingMgr.UpdatePathfinding(entityID, aiComp.HomePosition, StrategyDirect)
				direction := g.pathfindingMgr.GetPathfindingMove(entityID)
				if direction != (gruid.Point{}) {
					action = MoveAction{Direction: direction, EntityID: entityID}
				} else {
					direction = getDirectionTowards(pos, aiComp.HomePosition)
					action = MoveAction{Direction: direction, EntityID: entityID}
				}
			} else {
				direction := getDirectionTowards(pos, aiComp.HomePosition)
				action = MoveAction{Direction: direction, EntityID: entityID}
			}
		} else {
			// Random movement within patrol area
			directions := []gruid.Point{
				{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
			}
			direction := directions[rand.Intn(len(directions))]
			action = MoveAction{Direction: direction, EntityID: entityID}
		}

		actor.AddAction(action)

		// Update position estimate for next iteration
		if moveAction, ok := action.(MoveAction); ok {
			pos = pos.Add(moveAction.Direction)
			homeDistance = manhattanDistance(pos, aiComp.HomePosition)
		}
	}
}

// generateAttackSequence creates a sequence of actions for attacking
func (g *Game) generateAttackSequence(entityID ecs.EntityID, actor *components.TurnActor, pos gruid.Point) {
	// Check adjacent positions for attackable targets
	directions := []gruid.Point{
		{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
	}

	for _, dir := range directions {
		targetPos := pos.Add(dir)
		entities := g.ecs.EntitiesAt(targetPos)

		for _, targetID := range entities {
			if g.ecs.HasComponent(targetID, components.CHealth) && targetID != entityID {
				// Queue attack action and a follow-up wait
				actor.AddAction(AttackAction{AttackerID: entityID, TargetID: targetID})
				actor.AddAction(WaitAction{EntityID: entityID}) // Recovery time
				return
			}
		}
	}

	// No adjacent targets, queue wait action
	actor.AddAction(WaitAction{EntityID: entityID})
}

// generateIdleSequence creates a sequence of actions for idle behavior
func (g *Game) generateIdleSequence(entityID ecs.EntityID, actor *components.TurnActor, aiComp *components.AIComponent, pos gruid.Point) {
	switch aiComp.Behavior {
	case components.AIBehaviorWander:
		// Generate random movement sequence
		sequenceLength := rand.Intn(3) + 1
		directions := []gruid.Point{
			{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
		}

		for i := 0; i < sequenceLength; i++ {
			if rand.Intn(3) == 0 {
				// Occasionally wait
				actor.AddAction(WaitAction{EntityID: entityID})
			} else {
				direction := directions[rand.Intn(len(directions))]
				actor.AddAction(MoveAction{Direction: direction, EntityID: entityID})
			}
		}
	case components.AIBehaviorGuard:
		// Stay in place, occasionally look around
		actor.AddAction(WaitAction{EntityID: entityID})
		if rand.Intn(4) == 0 {
			// Occasionally "look" in a direction (cosmetic move that might fail)
			directions := []gruid.Point{
				{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
			}
			direction := directions[rand.Intn(len(directions))]
			actor.AddAction(MoveAction{Direction: direction, EntityID: entityID})
		}
	default:
		// Default idle behavior - just wait
		actor.AddAction(WaitAction{EntityID: entityID})
	}
}

// Helper functions

func (g *Game) clampToMapBounds(pos gruid.Point) gruid.Point {
	if pos.X < 0 {
		pos.X = 0
	}
	if pos.Y < 0 {
		pos.Y = 0
	}
	if pos.X >= g.dungeon.Width {
		pos.X = g.dungeon.Width - 1
	}
	if pos.Y >= g.dungeon.Height {
		pos.Y = g.dungeon.Height - 1
	}
	return pos
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
