package game

import (
	"math/rand"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/sirupsen/logrus"
)



// AdvancedMonsterAI handles more sophisticated monster AI
func (g *Game) AdvancedMonsterAI(entityID ecs.EntityID) GameAction {
	// Get AI component (we'll need to add this to the ECS)
	aiComp := g.getAIComponent(entityID)
	if aiComp == nil {
		// Fallback to basic AI
		return g.basicMonsterAI(entityID)
	}

	pos := g.ecs.GetPositionSafe(entityID)
	health := g.ecs.GetHealthSafe(entityID)
	hasHealth := g.ecs.HasHealthSafe(entityID)
	playerPos := g.GetPlayerPosition()
	distanceToPlayer := manhattanDistance(pos, playerPos)

	// Update AI state based on conditions
	g.updateAIState(entityID, aiComp, health, hasHealth, distanceToPlayer)

	// Save the updated AI component back to ECS
	g.ecs.AddComponent(entityID, components.CAIComponent, *aiComp)

	// Execute behavior based on current state
	switch aiComp.State {
	case components.AIStateChasing:
		return g.chasePlayer(entityID, pos, playerPos)
	case components.AIStateFleeing:
		return g.fleeFromPlayer(entityID, pos, playerPos)
	case components.AIStateSearching:
		return g.searchForPlayer(entityID, aiComp, pos)
	case components.AIStatePatrolling:
		return g.patrolArea(entityID, aiComp, pos)
	case components.AIStateAttacking:
		return g.attackNearbyTarget(entityID, pos)
	default: // AIStateIdle
		return g.idleBehavior(entityID, aiComp, pos)
	}
}

// updateAIState updates the AI state based on current conditions
func (g *Game) updateAIState(entityID ecs.EntityID, aiComp *components.AIComponent, health components.Health, hasHealth bool, distanceToPlayer int) {
	playerPos := g.GetPlayerPosition()
	canSeePlayer := g.canSeePlayer(entityID, playerPos)

	// Check if should flee
	if hasHealth && aiComp.FleeThreshold > 0 {
		healthPercent := float64(health.CurrentHP) / float64(health.MaxHP)
		if healthPercent <= aiComp.FleeThreshold && canSeePlayer {
			aiComp.State = components.AIStateFleeing
			return
		}
	}

	// Check if can see player and should chase
	if canSeePlayer && distanceToPlayer <= aiComp.AggroRange {
		aiComp.State = components.AIStateChasing
		aiComp.LastKnownPlayerPos = playerPos
		aiComp.SearchTurns = 0
		return
	}

	// Check if adjacent to player (attack)
	if distanceToPlayer == 1 {
		aiComp.State = components.AIStateAttacking
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

// chasePlayer moves towards the player
func (g *Game) chasePlayer(entityID ecs.EntityID, pos, playerPos gruid.Point) GameAction {
	direction := getDirectionTowards(pos, playerPos)
	return MoveAction{Direction: direction, EntityID: entityID}
}

// fleeFromPlayer moves away from the player
func (g *Game) fleeFromPlayer(entityID ecs.EntityID, pos, playerPos gruid.Point) GameAction {
	direction := getDirectionAway(pos, playerPos)
	return MoveAction{Direction: direction, EntityID: entityID}
}

// searchForPlayer moves towards last known player position
func (g *Game) searchForPlayer(entityID ecs.EntityID, aiComp *components.AIComponent, pos gruid.Point) GameAction {
	if pos == aiComp.LastKnownPlayerPos {
		// Reached last known position, look around randomly
		directions := []gruid.Point{
			{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
		}
		direction := directions[rand.Intn(len(directions))]
		return MoveAction{Direction: direction, EntityID: entityID}
	}

	direction := getDirectionTowards(pos, aiComp.LastKnownPlayerPos)
	return MoveAction{Direction: direction, EntityID: entityID}
}

// patrolArea moves around the home position
func (g *Game) patrolArea(entityID ecs.EntityID, aiComp *components.AIComponent, pos gruid.Point) GameAction {
	homeDistance := manhattanDistance(pos, aiComp.HomePosition)

	if homeDistance > aiComp.PatrolRadius {
		// Return to home area
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
		logrus.Debugf("Failed to move monster %d: %v", entityID, err)
		return WaitAction{EntityID: entityID}
	}
	return action
}

// getAIComponent retrieves AI component from ECS
func (g *Game) getAIComponent(entityID ecs.EntityID) *components.AIComponent {
	if !g.ecs.HasAIComponentSafe(entityID) {
		return nil
	}
	aiComp := g.ecs.GetAIComponentSafe(entityID)
	return &aiComp
}
