package game

import (
	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ui"
)

// DebugLevel represents different debug visualization levels
type DebugLevel int

const (
	DebugNone DebugLevel = iota
	DebugFOV
	DebugPathfinding
	DebugAI
	DebugFull
)

// String returns a human-readable name for the debug level
func (dl DebugLevel) String() string {
	switch dl {
	case DebugNone:
		return "None"
	case DebugFOV:
		return "FOV Only"
	case DebugPathfinding:
		return "Pathfinding Only"
	case DebugAI:
		return "AI Only"
	case DebugFull:
		return "Full Debug"
	default:
		return "Unknown"
	}
}

// AIDebugInfo contains debug information for AI visualization
type AIDebugInfo struct {
	EntityStates map[ecs.EntityID]AIEntityDebug
	LastUpdate   int // Turn number when last updated
}

// AIEntityDebug contains debug information for a single AI entity
type AIEntityDebug struct {
	EntityID           ecs.EntityID
	Position           gruid.Point
	State              components.AIState
	Behavior           components.AIBehavior
	DistanceToPlayer   int
	CanSeePlayer       bool
	HealthPercent      float64
	LastKnownPlayerPos gruid.Point
	SearchTurns        int
	MaxSearchTurns     int
}

// NewAIDebugInfo creates a new AI debug info structure
func NewAIDebugInfo() *AIDebugInfo {
	return &AIDebugInfo{
		EntityStates: make(map[ecs.EntityID]AIEntityDebug),
		LastUpdate:   0,
	}
}

// CollectAIDebugInfo gathers debug information from all AI entities
func (g *Game) CollectAIDebugInfo() *AIDebugInfo {
	debugInfo := NewAIDebugInfo()
	debugInfo.LastUpdate = g.stats.TurnCount

	// Get all entities with AI components
	aiEntities := g.ecs.GetEntitiesWithComponent(components.CAIComponent)
	playerPos := g.GetPlayerPosition()

	for _, entityID := range aiEntities {
		// Get entity components
		pos := g.ecs.GetPositionSafe(entityID)
		aiComp := g.ecs.GetAIComponentSafe(entityID)
		health := g.ecs.GetHealthSafe(entityID)

		if !g.ecs.HasAIComponentSafe(entityID) {
			continue
		}

		// Calculate debug information
		distanceToPlayer := manhattanDistance(pos, playerPos)
		canSeePlayer := g.canSeePlayer(entityID, playerPos)

		var healthPercent float64 = 1.0
		if g.ecs.HasHealthSafe(entityID) && health.MaxHP > 0 {
			healthPercent = float64(health.CurrentHP) / float64(health.MaxHP)
		}

		// Create debug entry
		debugInfo.EntityStates[entityID] = AIEntityDebug{
			EntityID:           entityID,
			Position:           pos,
			State:              aiComp.State,
			Behavior:           aiComp.Behavior,
			DistanceToPlayer:   distanceToPlayer,
			CanSeePlayer:       canSeePlayer,
			HealthPercent:      healthPercent,
			LastKnownPlayerPos: aiComp.LastKnownPlayerPos,
			SearchTurns:        aiComp.SearchTurns,
			MaxSearchTurns:     aiComp.MaxSearchTurns,
		}
	}

	return debugInfo
}

// GetPathfindingDebugColor returns the appropriate color for a pathfinding debug visualization
func GetPathfindingDebugColor(aiState components.AIState) gruid.Color {
	switch aiState {
	case components.AIStateChasing, components.AIStateAttacking:
		return ui.ColorDebugPathChasing
	case components.AIStateFleeing:
		return ui.ColorDebugPathFleeing
	case components.AIStatePatrolling, components.AIStateIdle:
		return ui.ColorDebugPathPatrolling
	case components.AIStateSearching:
		return ui.ColorDebugPathSearching
	default:
		return ui.ColorForeground
	}
}

// GetFOVDebugColor returns the appropriate color for FOV debug visualization
func GetFOVDebugColor(isVisible, isExplored bool) gruid.Color {
	if isVisible {
		return ui.ColorDebugFOVVisible
	} else if isExplored {
		return ui.ColorDebugFOVExplored
	} else {
		return ui.ColorDebugFOVUnexplored
	}
}

// GetAIStateString returns a human-readable string for an AI state
func GetAIStateString(state components.AIState) string {
	switch state {
	case components.AIStateIdle:
		return "IDLE"
	case components.AIStatePatrolling:
		return "PATROL"
	case components.AIStateChasing:
		return "CHASE"
	case components.AIStateFleeing:
		return "FLEE"
	case components.AIStateAttacking:
		return "ATTACK"
	case components.AIStateSearching:
		return "SEARCH"
	default:
		return "UNKNOWN"
	}
}

// GetAIBehaviorString returns a human-readable string for an AI behavior
func GetAIBehaviorString(behavior components.AIBehavior) string {
	switch behavior {
	case components.AIBehaviorPassive:
		return "Passive"
	case components.AIBehaviorWander:
		return "Wander"
	case components.AIBehaviorGuard:
		return "Guard"
	case components.AIBehaviorHunter:
		return "Hunter"
	case components.AIBehaviorFleeing:
		return "Fleeing"
	case components.AIBehaviorPack:
		return "Pack"
	default:
		return "Unknown"
	}
}

// GetPathCharacter returns the appropriate character for drawing a path segment
func GetPathCharacter(from, to gruid.Point) rune {
	dx := to.X - from.X
	dy := to.Y - from.Y

	// Determine direction and return appropriate character
	if dx == 0 && dy != 0 {
		return '│' // Vertical line
	} else if dy == 0 && dx != 0 {
		return '─' // Horizontal line
	} else if dx != 0 && dy != 0 {
		// Diagonal - use a simple dot
		return '·'
	}

	return '·' // Default waypoint character
}

// IsInViewport checks if a point is within the camera viewport
func (md *Model) IsInViewport(p gruid.Point) bool {
	return md.camera.IsInViewport(p.X, p.Y)
}
