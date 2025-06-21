package game

import (
	"log/slog"
	"math/rand"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

// AdvancedMonsterAI handles more sophisticated monster AI
func (g *Game) AdvancedMonsterAI(entityID ecs.EntityID) GameAction {
	// Get AI component by value
	aiComp, hasAI := g.getAIComponent(entityID)
	if !hasAI {
		// Fallback to basic AI
		return g.basicMonsterAI(entityID)
	}

	pos := g.ecs.GetPositionSafe(entityID)
	health := g.ecs.GetHealthSafe(entityID)
	hasHealth := g.ecs.HasHealthSafe(entityID)
	playerPos := g.GetPlayerPosition()
	distanceToPlayer := manhattanDistance(pos, playerPos)

	// Update AI state based on conditions
	g.updateAIState(entityID, &aiComp, health, hasHealth, distanceToPlayer)

	// Execute behavior based on current state
	var action GameAction
	switch aiComp.State {
	case components.AIStateChasing:
		action = g.chasePlayer(entityID, pos, playerPos)
	case components.AIStateFleeing:
		action = g.fleeFromPlayer(entityID, pos, playerPos)
	case components.AIStateSearching:
		action = g.searchForPlayer(entityID, &aiComp, pos)
	case components.AIStatePatrolling:
		action = g.patrolArea(entityID, &aiComp, pos)
	case components.AIStateAttacking:
		action = g.attackNearbyTarget(entityID, pos)
	default: // AIStateIdle
		action = g.idleBehavior(entityID, &aiComp, pos)
	}

	// Save the updated AI component back to ECS
	g.ecs.AddComponent(entityID, components.CAIComponent, aiComp)

	return action
}

// updateAIState updates the AI state based on current conditions
func (g *Game) updateAIState(entityID ecs.EntityID, aiComp *components.AIComponent, health components.Health, hasHealth bool, distanceToPlayer int) {
	playerPos := g.GetPlayerPosition()
	canSeePlayer := g.canSeePlayer(entityID, playerPos)

	// Check if should flee
	if hasHealth && aiComp.FleeThreshold > 0 {
		if health.MaxHP == 0 { // corrupt component – bail out
			return
		}

		healthPercent := float64(health.CurrentHP) / float64(health.MaxHP)
		if healthPercent <= aiComp.FleeThreshold && canSeePlayer {
			aiComp.State = components.AIStateFleeing
			return
		}
	}

	// If adjacent, always attack first
	if distanceToPlayer == 1 {
		aiComp.State = components.AIStateAttacking
		return
	}

	// Otherwise, chase when in aggro range and player is visible
	if canSeePlayer && distanceToPlayer <= aiComp.AggroRange {
		aiComp.State = components.AIStateChasing
		aiComp.LastKnownPlayerPos = playerPos
		aiComp.SearchTurns = 0
		return
	}

	// If was chasing but lost sight, start searching
	if aiComp.State == components.AIStateChasing && !canSeePlayer {
		aiComp.State = components.AIStateSearching
		aiComp.SearchTurns = 0
		return
	}

	// Continue searching if not exceeded max turns
	if aiComp.State == components.AIStateSearching {
		aiComp.SearchTurns++
		if aiComp.SearchTurns >= aiComp.MaxSearchTurns {
			aiComp.State = components.AIStateIdle
		}
		return
	}

	// Default behavior based on AI type
	switch aiComp.Behavior {
	case components.AIBehaviorGuard:
		aiComp.State = components.AIStatePatrolling
	case components.AIBehaviorWander:
		aiComp.State = components.AIStatePatrolling
	default:
		aiComp.State = components.AIStateIdle
	}
}

// chasePlayer moves towards the player using pathfinding
func (g *Game) chasePlayer(entityID ecs.EntityID, pos, playerPos gruid.Point) GameAction {
	// Use pathfinding if available, otherwise fall back to simple movement
	if g.pathfindingMgr != nil {
		// Update pathfinding to target the player
		g.pathfindingMgr.UpdatePathfinding(entityID, playerPos, StrategyDirect)

		// Get the next move from pathfinding
		direction := g.pathfindingMgr.GetPathfindingMove(entityID)
		if direction != (gruid.Point{}) {
			return MoveAction{Direction: direction, EntityID: entityID}
		}
	}

	// Fallback to simple directional movement
	direction := getDirectionTowards(pos, playerPos)
	return MoveAction{Direction: direction, EntityID: entityID}
}

// fleeFromPlayer moves away from the player using pathfinding
func (g *Game) fleeFromPlayer(entityID ecs.EntityID, pos, playerPos gruid.Point) GameAction {
	// Use pathfinding if available, otherwise fall back to simple movement
	if g.pathfindingMgr != nil {
		// Calculate a flee target (opposite direction from player)
		fleeDirection := getDirectionAway(pos, playerPos)
		fleeTarget := pos.Add(fleeDirection.Mul(5)) // Flee 5 steps away

		// Clamp to map bounds
		if fleeTarget.X < 0 {
			fleeTarget.X = 0
		}
		if fleeTarget.Y < 0 {
			fleeTarget.Y = 0
		}
		if fleeTarget.X >= g.dungeon.Width {
			fleeTarget.X = g.dungeon.Width - 1
		}
		if fleeTarget.Y >= g.dungeon.Height {
			fleeTarget.Y = g.dungeon.Height - 1
		}

		// Update pathfinding to flee target with entity avoidance
		g.pathfindingMgr.UpdatePathfinding(entityID, fleeTarget, StrategyAvoidEntities)

		// Get the next move from pathfinding
		direction := g.pathfindingMgr.GetPathfindingMove(entityID)
		if direction != (gruid.Point{}) {
			return MoveAction{Direction: direction, EntityID: entityID}
		}
	}

	// Fallback to simple directional movement
	direction := getDirectionAway(pos, playerPos)
	return MoveAction{Direction: direction, EntityID: entityID}
}

// searchForPlayer moves towards last known player position using pathfinding
func (g *Game) searchForPlayer(entityID ecs.EntityID, aiComp *components.AIComponent, pos gruid.Point) GameAction {
	if pos == aiComp.LastKnownPlayerPos {
		// Reached last known position, look around randomly
		directions := []gruid.Point{
			{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
		}
		direction := directions[rand.Intn(len(directions))]
		return MoveAction{Direction: direction, EntityID: entityID}
	}

	// Use pathfinding if available, otherwise fall back to simple movement
	if g.pathfindingMgr != nil {
		// Update pathfinding to target the last known player position
		g.pathfindingMgr.UpdatePathfinding(entityID, aiComp.LastKnownPlayerPos, StrategyDirect)

		// Get the next move from pathfinding
		direction := g.pathfindingMgr.GetPathfindingMove(entityID)
		if direction != (gruid.Point{}) {
			return MoveAction{Direction: direction, EntityID: entityID}
		}
	}

	// Fallback to simple directional movement
	direction := getDirectionTowards(pos, aiComp.LastKnownPlayerPos)
	return MoveAction{Direction: direction, EntityID: entityID}
}

// patrolArea moves around the home position using pathfinding
func (g *Game) patrolArea(entityID ecs.EntityID, aiComp *components.AIComponent, pos gruid.Point) GameAction {
	homeDistance := manhattanDistance(pos, aiComp.HomePosition)

	if homeDistance > aiComp.PatrolRadius {
		// Return to home area using pathfinding
		if g.pathfindingMgr != nil {
			g.pathfindingMgr.UpdatePathfinding(entityID, aiComp.HomePosition, StrategyDirect)

			direction := g.pathfindingMgr.GetPathfindingMove(entityID)
			if direction != (gruid.Point{}) {
				return MoveAction{Direction: direction, EntityID: entityID}
			}
		}

		// Fallback to simple directional movement
		direction := getDirectionTowards(pos, aiComp.HomePosition)
		return MoveAction{Direction: direction, EntityID: entityID}
	}

	// Random movement within patrol area
	directions := []gruid.Point{
		{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
	}
	direction := directions[rand.Intn(len(directions))]
	return MoveAction{Direction: direction, EntityID: entityID}
}

// attackNearbyTarget attacks adjacent enemies
func (g *Game) attackNearbyTarget(entityID ecs.EntityID, pos gruid.Point) GameAction {
	// Check adjacent positions for attackable targets
	directions := []gruid.Point{
		{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
	}

	for _, dir := range directions {
		targetPos := pos.Add(dir)
		entities := g.ecs.EntitiesAt(targetPos)

		for _, targetID := range entities {
			if g.ecs.HasComponent(targetID, components.CHealth) && targetID != entityID {
				return AttackAction{AttackerID: entityID, TargetID: targetID}
			}
		}
	}

	// No adjacent targets, wait
	return WaitAction{EntityID: entityID}
}

// idleBehavior default idle behavior
func (g *Game) idleBehavior(entityID ecs.EntityID, aiComp *components.AIComponent, pos gruid.Point) GameAction {
	switch aiComp.Behavior {
	case components.AIBehaviorWander:
		if rand.Intn(3) == 0 { // 33% chance to move
			directions := []gruid.Point{
				{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
			}
			direction := directions[rand.Intn(len(directions))]
			return MoveAction{Direction: direction, EntityID: entityID}
		}
	}

	return WaitAction{EntityID: entityID}
}

// Helper functions

func manhattanDistance(a, b gruid.Point) int {
	return abs(a.X-b.X) + abs(a.Y-b.Y)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func getDirectionTowards(from, to gruid.Point) gruid.Point {
	dx := to.X - from.X
	dy := to.Y - from.Y

	// Normalize to unit direction
	if dx > 0 {
		dx = 1
	} else if dx < 0 {
		dx = -1
	}

	if dy > 0 {
		dy = 1
	} else if dy < 0 {
		dy = -1
	}

	return gruid.Point{X: dx, Y: dy}
}

func getDirectionAway(from, away gruid.Point) gruid.Point {
	direction := getDirectionTowards(from, away)
	return gruid.Point{X: -direction.X, Y: -direction.Y}
}

// canSeePlayer checks if the entity can see the player using FOV
func (g *Game) canSeePlayer(entityID ecs.EntityID, playerPos gruid.Point) bool {
	fov := g.ecs.GetFOVSafe(entityID)
	if fov == nil {
		return false
	}

	return fov.IsVisible(playerPos, g.dungeon.Width)
}

// basicMonsterAI fallback to the original simple AI
func (g *Game) basicMonsterAI(entityID ecs.EntityID) GameAction {
	// This would be the existing monster AI logic
	action, err := moveMonster(g, entityID)
	if err != nil {
		slog.Debug("Failed to move monster", "entityID", entityID, "error", err)
		return WaitAction{EntityID: entityID}
	}
	return action
}

// getAIComponent retrieves AI component from ECS by value.
// Callers must explicitly call AddComponent to persist any mutations.
//
// IMPORTANT: This function was fixed to prevent state loss. Previously, it returned
// a pointer to a local copy (*components.AIComponent), which caused mutations to be
// lost since they weren't persisted back to the ECS.
//
// Two solutions were implemented:
// 1. Return by value (this function) - requires explicit AddComponent calls
// 2. UpdateAIComponent method - provides safe concurrent access with automatic persistence
func (g *Game) getAIComponent(entityID ecs.EntityID) (components.AIComponent, bool) {
	if !g.ecs.HasAIComponentSafe(entityID) {
		return components.AIComponent{}, false
	}
	aiComp := g.ecs.GetAIComponentSafe(entityID)
	return aiComp, true
}

// AdvancedMonsterAIWithUpdate demonstrates the alternative approach using UpdateAIComponent
// for direct mutation with proper concurrency control.
func (g *Game) AdvancedMonsterAIWithUpdate(entityID ecs.EntityID) GameAction {
	if !g.ecs.HasAIComponentSafe(entityID) {
		// Fallback to basic AI
		return g.basicMonsterAI(entityID)
	}

	pos := g.ecs.GetPositionSafe(entityID)
	health := g.ecs.GetHealthSafe(entityID)
	hasHealth := g.ecs.HasHealthSafe(entityID)
	playerPos := g.GetPlayerPosition()
	distanceToPlayer := manhattanDistance(pos, playerPos)

	var action GameAction

	// Update AI component with direct mutation
	err := g.ecs.UpdateAIComponent(entityID, func(aiComp *components.AIComponent) error {
		// Update AI state based on conditions
		g.updateAIState(entityID, aiComp, health, hasHealth, distanceToPlayer)

		// Execute behavior based on current state
		switch aiComp.State {
		case components.AIStateChasing:
			action = g.chasePlayer(entityID, pos, playerPos)
		case components.AIStateFleeing:
			action = g.fleeFromPlayer(entityID, pos, playerPos)
		case components.AIStateSearching:
			action = g.searchForPlayer(entityID, aiComp, pos)
		case components.AIStatePatrolling:
			action = g.patrolArea(entityID, aiComp, pos)
		case components.AIStateAttacking:
			action = g.attackNearbyTarget(entityID, pos)
		default: // AIStateIdle
			action = g.idleBehavior(entityID, aiComp, pos)
		}

		return nil
	})

	if err != nil {
		slog.Error("Failed to update AI component for entity", "entityID", entityID, "error", err)
		return g.basicMonsterAI(entityID)
	}

	return action
}
