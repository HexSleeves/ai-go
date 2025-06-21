package components

import "codeberg.org/anaseto/gruid"

// PathfindingComponent stores pathfinding state for an entity
type PathfindingComponent struct {
	// CurrentPath is the sequence of points the entity should follow
	CurrentPath []gruid.Point

	// TargetPos is the current pathfinding target
	TargetPos gruid.Point

	// PathValid indicates if the current path is still valid
	PathValid bool

	// LastRecompute tracks when the path was last computed (turn number)
	LastRecompute int

	// Strategy defines the pathfinding approach to use (stored as int)
	Strategy int

	// RecomputeFrequency defines how often to recompute paths (in turns)
	// 0 means only recompute when necessary
	RecomputeFrequency int

	// MaxPathLength limits the maximum path length to prevent excessive computation
	MaxPathLength int

	// UsePathfinding enables/disables pathfinding for this entity
	UsePathfinding bool
}

// NewPathfindingComponent creates a new pathfinding component with default values
func NewPathfindingComponent() PathfindingComponent {
	return PathfindingComponent{
		CurrentPath:        make([]gruid.Point, 0),
		TargetPos:          gruid.Point{},
		PathValid:          false,
		LastRecompute:      0,
		Strategy:           0,  // Default strategy (StrategyDirect)
		RecomputeFrequency: 0,  // Only recompute when necessary
		MaxPathLength:      50, // Reasonable default for most maps
		UsePathfinding:     true,
	}
}

// NewPathfindingComponentWithStrategy creates a pathfinding component with a specific strategy
func NewPathfindingComponentWithStrategy(strategy int) PathfindingComponent {
	comp := NewPathfindingComponent()
	comp.Strategy = strategy
	return comp
}

// HasPath returns true if the component has a valid path
func (pc *PathfindingComponent) HasPath() bool {
	return pc.PathValid && len(pc.CurrentPath) > 0
}

// GetNextPosition returns the next position in the path, or zero point if no path
func (pc *PathfindingComponent) GetNextPosition(currentPos gruid.Point) gruid.Point {
	if !pc.HasPath() {
		return gruid.Point{}
	}

	// Find current position in path
	for i, p := range pc.CurrentPath {
		if p == currentPos {
			if i+1 < len(pc.CurrentPath) {
				return pc.CurrentPath[i+1]
			}
			break
		}
	}

	// If current position not found in path, return first position
	if len(pc.CurrentPath) > 0 {
		return pc.CurrentPath[0]
	}

	return gruid.Point{}
}

// AdvancePath removes the current position from the path
func (pc *PathfindingComponent) AdvancePath(currentPos gruid.Point) {
	if len(pc.CurrentPath) == 0 {
		return
	}

	// If we're at the first position in the path, remove it
	if pc.CurrentPath[0] == currentPos {
		pc.CurrentPath = pc.CurrentPath[1:]
		return
	}

	// Find and remove current position from path
	for i, p := range pc.CurrentPath {
		if p == currentPos {
			pc.CurrentPath = pc.CurrentPath[i+1:]
			return
		}
	}
}

// ClearPath clears the current path and marks it as invalid
func (pc *PathfindingComponent) ClearPath() {
	pc.CurrentPath = pc.CurrentPath[:0] // Clear slice but keep capacity
	pc.PathValid = false
	pc.TargetPos = gruid.Point{}
}

// SetPath sets a new path and marks it as valid
func (pc *PathfindingComponent) SetPath(path []gruid.Point, target gruid.Point) {
	// Limit path length if necessary
	if pc.MaxPathLength > 0 && len(path) > pc.MaxPathLength {
		path = path[:pc.MaxPathLength]
	}

	pc.CurrentPath = make([]gruid.Point, len(path))
	copy(pc.CurrentPath, path)
	pc.TargetPos = target
	pc.PathValid = true
}

// NeedsRecompute checks if the path needs to be recomputed
func (pc *PathfindingComponent) NeedsRecompute(currentTurn int) bool {
	if !pc.UsePathfinding {
		return false
	}

	if !pc.PathValid || len(pc.CurrentPath) == 0 {
		return true
	}

	if pc.RecomputeFrequency > 0 {
		return currentTurn-pc.LastRecompute >= pc.RecomputeFrequency
	}

	return false
}

// MarkRecomputed updates the last recompute time
func (pc *PathfindingComponent) MarkRecomputed(currentTurn int) {
	pc.LastRecompute = currentTurn
}

// GetPathLength returns the current path length
func (pc *PathfindingComponent) GetPathLength() int {
	return len(pc.CurrentPath)
}

// IsAtTarget checks if the entity is at the target position
func (pc *PathfindingComponent) IsAtTarget(currentPos gruid.Point) bool {
	return currentPos == pc.TargetPos
}

// GetRemainingDistance returns the number of steps remaining in the path
func (pc *PathfindingComponent) GetRemainingDistance(currentPos gruid.Point) int {
	if !pc.HasPath() {
		return -1 // No path
	}

	// Find current position in path and return remaining steps
	for i, p := range pc.CurrentPath {
		if p == currentPos {
			return len(pc.CurrentPath) - i - 1
		}
	}

	// If current position not found, return full path length
	return len(pc.CurrentPath)
}

// SetStrategy updates the pathfinding strategy
func (pc *PathfindingComponent) SetStrategy(strategy int) {
	if pc.Strategy != strategy {
		pc.Strategy = strategy
		// Mark path as invalid to force recomputation with new strategy
		pc.PathValid = false
	}
}

// SetMaxPathLength updates the maximum path length
func (pc *PathfindingComponent) SetMaxPathLength(maxLength int) {
	pc.MaxPathLength = maxLength

	// Truncate current path if it exceeds the new limit
	if maxLength > 0 && len(pc.CurrentPath) > maxLength {
		pc.CurrentPath = pc.CurrentPath[:maxLength]
	}
}

// EnablePathfinding enables pathfinding for this entity
func (pc *PathfindingComponent) EnablePathfinding() {
	pc.UsePathfinding = true
}

// DisablePathfinding disables pathfinding for this entity
func (pc *PathfindingComponent) DisablePathfinding() {
	pc.UsePathfinding = false
	pc.ClearPath()
}

// IsPathfindingEnabled returns true if pathfinding is enabled
func (pc *PathfindingComponent) IsPathfindingEnabled() bool {
	return pc.UsePathfinding
}

// GetStrategyName returns a human-readable name for the current strategy
func (pc *PathfindingComponent) GetStrategyName() string {
	switch pc.Strategy {
	case 0: // StrategyDirect
		return "Direct"
	case 1: // StrategyAvoidEntities
		return "Avoid Entities"
	case 2: // StrategyPreferOpen
		return "Prefer Open"
	case 3: // StrategyStealthy
		return "Stealthy"
	default:
		return "Unknown"
	}
}

// Clone creates a deep copy of the pathfinding component
func (pc *PathfindingComponent) Clone() PathfindingComponent {
	clone := *pc
	clone.CurrentPath = make([]gruid.Point, len(pc.CurrentPath))
	copy(clone.CurrentPath, pc.CurrentPath)
	return clone
}
