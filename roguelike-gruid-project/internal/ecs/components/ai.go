package components

import "codeberg.org/anaseto/gruid"

// AIBehavior represents different AI behavior types
type AIBehavior int

const (
	AIBehaviorPassive AIBehavior = iota // Doesn't move unless attacked
	AIBehaviorWander                    // Random movement
	AIBehaviorGuard                     // Patrols a specific area
	AIBehaviorHunter                    // Actively seeks player
	AIBehaviorFleeing                   // Runs away from player
	AIBehaviorPack                      // Coordinates with nearby allies
)

// AIState represents the current state of an AI entity
type AIState int

const (
	AIStateIdle AIState = iota
	AIStatePatrolling
	AIStateChasing
	AIStateFleeing
	AIStateAttacking
	AIStateSearching // Lost sight of player, searching last known position
)

// AIComponent extends the basic AITag with more sophisticated behavior
type AIComponent struct {
	Behavior           AIBehavior
	State              AIState
	LastKnownPlayerPos gruid.Point
	HomePosition       gruid.Point
	PatrolRadius       int
	AggroRange         int
	FleeThreshold      float64 // Health percentage to start fleeing
	SearchTurns        int     // Turns spent searching for player
	MaxSearchTurns     int
}

// NewAIComponent creates a new AI component with default values
func NewAIComponent(behavior AIBehavior, homePos gruid.Point) AIComponent {
	return AIComponent{
		Behavior:       behavior,
		State:          AIStateIdle,
		HomePosition:   homePos,
		PatrolRadius:   5,
		AggroRange:     8,
		FleeThreshold:  0.3,
		MaxSearchTurns: 10,
	}
}

// SetState changes the AI state
func (ai *AIComponent) SetState(state AIState) {
	ai.State = state
}

// GetState returns the current AI state
func (ai *AIComponent) GetState() AIState {
	return ai.State
}

// IsAggressive returns true if the AI is in an aggressive state
func (ai *AIComponent) IsAggressive() bool {
	return ai.State == AIStateChasing || ai.State == AIStateAttacking
}

// IsFleeing returns true if the AI is fleeing
func (ai *AIComponent) IsFleeing() bool {
	return ai.State == AIStateFleeing
}

// IsSearching returns true if the AI is searching for the player
func (ai *AIComponent) IsSearching() bool {
	return ai.State == AIStateSearching
}

// ShouldFlee checks if the AI should flee based on health
func (ai *AIComponent) ShouldFlee(currentHP, maxHP int) bool {
	if ai.FleeThreshold <= 0 || maxHP <= 0 {
		return false
	}
	healthPercent := float64(currentHP) / float64(maxHP)
	return healthPercent <= ai.FleeThreshold
}

// IncrementSearchTurns increments the search turn counter
func (ai *AIComponent) IncrementSearchTurns() {
	ai.SearchTurns++
}

// ResetSearchTurns resets the search turn counter
func (ai *AIComponent) ResetSearchTurns() {
	ai.SearchTurns = 0
}

// HasExceededMaxSearchTurns checks if search time limit is exceeded
func (ai *AIComponent) HasExceededMaxSearchTurns() bool {
	return ai.SearchTurns >= ai.MaxSearchTurns
}

// UpdateLastKnownPlayerPos updates the last known player position
func (ai *AIComponent) UpdateLastKnownPlayerPos(pos gruid.Point) {
	ai.LastKnownPlayerPos = pos
}

// GetDistanceFromHome calculates distance from home position
func (ai *AIComponent) GetDistanceFromHome(currentPos gruid.Point) int {
	return manhattanDistance(currentPos, ai.HomePosition)
}

// IsOutsidePatrolArea checks if the AI is outside its patrol area
func (ai *AIComponent) IsOutsidePatrolArea(currentPos gruid.Point) bool {
	return ai.GetDistanceFromHome(currentPos) > ai.PatrolRadius
}

// manhattanDistance calculates Manhattan distance between two points
func manhattanDistance(a, b gruid.Point) int {
	dx := a.X - b.X
	dy := a.Y - b.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}
