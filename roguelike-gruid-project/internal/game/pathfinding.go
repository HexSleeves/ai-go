package game

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/sirupsen/logrus"
)

// PathfindingStrategy defines different pathfinding approaches
type PathfindingStrategy int

const (
	StrategyDirect PathfindingStrategy = iota // Direct path to target
	StrategyAvoidEntities                     // Avoid other entities when possible
	StrategyPreferOpen                        // Prefer open areas
	StrategyStealthy                          // Avoid player FOV when possible
)

// PathfindingManager handles all pathfinding operations for the game
type PathfindingManager struct {
	pathRange *paths.PathRange
	game      *Game
	neighbors paths.Neighbors
}

// NewPathfindingManager creates a new pathfinding manager
func NewPathfindingManager(game *Game) *PathfindingManager {
	// Create path range covering the entire map
	mapRange := gruid.NewRange(0, 0, game.dungeon.Width, game.dungeon.Height)
	
	return &PathfindingManager{
		pathRange: paths.NewPathRange(mapRange),
		game:      game,
		neighbors: paths.Neighbors{},
	}
}

// Neighbors implements paths.Astar.Neighbors
// Returns walkable neighboring positions using 4-way movement
func (pm *PathfindingManager) Neighbors(p gruid.Point) []gruid.Point {
	return pm.neighbors.Cardinal(p, func(q gruid.Point) bool {
		return pm.isWalkable(q)
	})
}

// Cost implements paths.Astar.Cost
// Returns the movement cost from one position to an adjacent position
func (pm *PathfindingManager) Cost(from, to gruid.Point) int {
	baseCost := 1

	// Check if there are blocking entities at the destination
	entities := pm.game.ecs.EntitiesAt(to)
	for _, entityID := range entities {
		if pm.game.ecs.HasComponent(entityID, components.CBlocksMovement) {
			// Add extra cost for positions with blocking entities
			// This encourages pathfinding to find alternate routes
			baseCost += 8
		}
	}

	// Future: Add terrain-based costs here
	// if pm.game.dungeon.GetTerrain(to) == TerrainDifficult {
	//     baseCost += 2
	// }

	return baseCost
}

// Estimation implements paths.Astar.Estimation
// Returns the heuristic distance estimate for A* algorithm
func (pm *PathfindingManager) Estimation(from, to gruid.Point) int {
	// Use Manhattan distance for 4-way movement
	return paths.DistanceManhattan(from, to)
}

// isWalkable checks if a position is walkable
func (pm *PathfindingManager) isWalkable(p gruid.Point) bool {
	return pm.game.dungeon.InBounds(p) && pm.game.dungeon.isWalkable(p)
}

// FindPath computes a path from start to goal using the best available algorithm
func (pm *PathfindingManager) FindPath(from, to gruid.Point, strategy PathfindingStrategy) []gruid.Point {
	if !pm.isWalkable(from) || !pm.isWalkable(to) {
		logrus.Debugf("Pathfinding: Invalid start (%v) or goal (%v) position", from, to)
		return nil
	}

	var path []gruid.Point

	// Choose pathfinding algorithm based on distance and strategy
	distance := paths.DistanceManhattan(from, to)

	if distance > 15 && strategy == StrategyDirect {
		// Use JPS for longer distances with direct strategy for better performance
		path = pm.findPathJPS(from, to)
	} else {
		// Use A* for shorter distances or complex strategies
		path = pm.pathRange.AstarPath(pm, from, to)
	}

	if path == nil {
		logrus.Debugf("Pathfinding: No path found from %v to %v", from, to)
		return nil
	}

	// Apply strategy-specific modifications
	path = pm.applyStrategy(path, strategy)

	logrus.Debugf("Pathfinding: Found path of length %d from %v to %v", len(path), from, to)
	return path
}

// findPathJPS uses Jump Point Search for better performance on longer paths
func (pm *PathfindingManager) findPathJPS(from, to gruid.Point) []gruid.Point {
	// Use gruid's JPS implementation
	path := pm.pathRange.JPSPath(nil, from, to, pm.isWalkable, false) // false = 4-way movement
	return path
}

// applyStrategy modifies the path based on the pathfinding strategy
func (pm *PathfindingManager) applyStrategy(path []gruid.Point, strategy PathfindingStrategy) []gruid.Point {
	switch strategy {
	case StrategyDirect:
		// No modifications needed
		return path
	case StrategyAvoidEntities:
		return pm.applyEntityAvoidanceStrategy(path)
	case StrategyPreferOpen:
		return pm.applyOpenAreaStrategy(path)
	case StrategyStealthy:
		return pm.applyStealthyStrategy(path)
	default:
		return path
	}
}

// applyEntityAvoidanceStrategy modifies path to avoid other entities when possible
func (pm *PathfindingManager) applyEntityAvoidanceStrategy(path []gruid.Point) []gruid.Point {
	if len(path) <= 2 {
		return path // Too short to optimize
	}

	optimizedPath := make([]gruid.Point, 0, len(path))
	optimizedPath = append(optimizedPath, path[0]) // Always include start

	for i := 1; i < len(path)-1; i++ {
		currentPoint := path[i]

		// Check if there are entities at this position
		entities := pm.game.ecs.EntitiesAt(currentPoint)
		hasBlockingEntity := false

		for _, entityID := range entities {
			if pm.game.ecs.HasComponent(entityID, components.CBlocksMovement) {
				hasBlockingEntity = true
				break
			}
		}

		if hasBlockingEntity {
			// Try to find an alternative route around this point
			alternative := pm.findAlternativePoint(path[i-1], path[i+1], currentPoint)
			if alternative != (gruid.Point{}) {
				optimizedPath = append(optimizedPath, alternative)
				continue
			}
		}

		optimizedPath = append(optimizedPath, currentPoint)
	}

	optimizedPath = append(optimizedPath, path[len(path)-1]) // Always include end
	return optimizedPath
}

// applyOpenAreaStrategy prefers paths through open areas
func (pm *PathfindingManager) applyOpenAreaStrategy(path []gruid.Point) []gruid.Point {
	// For now, return the original path
	// Future enhancement: Analyze surrounding area and prefer wider corridors
	return path
}

// applyStealthyStrategy avoids player FOV when possible
func (pm *PathfindingManager) applyStealthyStrategy(path []gruid.Point) []gruid.Point {
	if pm.game.PlayerID == 0 {
		return path // No player to avoid
	}

	playerFOV := pm.game.ecs.GetFOVSafe(pm.game.PlayerID)
	if playerFOV == nil {
		return path // No FOV to avoid
	}

	// Filter out points that are visible to the player when possible
	stealthyPath := make([]gruid.Point, 0, len(path))
	stealthyPath = append(stealthyPath, path[0]) // Always include start

	for i := 1; i < len(path)-1; i++ {
		point := path[i]

		// If point is visible to player, try to find alternative
		if playerFOV.IsVisible(point, pm.game.dungeon.Width) {
			// Try adjacent points that might not be visible
			alternative := pm.findStealthyAlternative(path[i-1], path[i+1], point, playerFOV)
			if alternative != (gruid.Point{}) {
				stealthyPath = append(stealthyPath, alternative)
				continue
			}
		}

		stealthyPath = append(stealthyPath, point)
	}

	stealthyPath = append(stealthyPath, path[len(path)-1]) // Always include end
	return stealthyPath
}

// findAlternativePoint tries to find an alternative point between two positions
func (pm *PathfindingManager) findAlternativePoint(from, to, avoid gruid.Point) gruid.Point {
	// Try adjacent points to the avoid position
	directions := []gruid.Point{
		{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
	}

	for _, dir := range directions {
		candidate := avoid.Add(dir)

		// Check if candidate is walkable and doesn't have blocking entities
		if pm.isWalkable(candidate) {
			entities := pm.game.ecs.EntitiesAt(candidate)
			hasBlockingEntity := false

			for _, entityID := range entities {
				if pm.game.ecs.HasComponent(entityID, components.CBlocksMovement) {
					hasBlockingEntity = true
					break
				}
			}

			if !hasBlockingEntity {
				// Check if this creates a valid path segment
				if pm.isValidPathSegment(from, candidate, to) {
					return candidate
				}
			}
		}
	}

	return gruid.Point{} // No alternative found
}

// findStealthyAlternative finds an alternative point that's not visible to player
func (pm *PathfindingManager) findStealthyAlternative(from, to, avoid gruid.Point, playerFOV *components.FOV) gruid.Point {
	directions := []gruid.Point{
		{X: -1, Y: 0}, {X: 1, Y: 0}, {X: 0, Y: -1}, {X: 0, Y: 1},
	}

	for _, dir := range directions {
		candidate := avoid.Add(dir)

		// Check if candidate is walkable and not visible to player
		if pm.isWalkable(candidate) && !playerFOV.IsVisible(candidate, pm.game.dungeon.Width) {
			if pm.isValidPathSegment(from, candidate, to) {
				return candidate
			}
		}
	}

	return gruid.Point{} // No alternative found
}

// isValidPathSegment checks if a path segment is valid (no walls between points)
func (pm *PathfindingManager) isValidPathSegment(from, middle, to gruid.Point) bool {
	// Simple check: ensure both segments are walkable
	return pm.isWalkable(middle) &&
		   paths.DistanceManhattan(from, middle) <= 2 &&
		   paths.DistanceManhattan(middle, to) <= 2
}

// GetNextMove returns the next move direction from a path
func (pm *PathfindingManager) GetNextMove(path []gruid.Point, currentPos gruid.Point) gruid.Point {
	if len(path) == 0 {
		return gruid.Point{} // No movement
	}

	// If we're not at the expected position, recalculate
	if len(path) > 0 && path[0] != currentPos {
		// Find our position in the path
		for i, p := range path {
			if p == currentPos {
				if i+1 < len(path) {
					nextPos := path[i+1]
					return nextPos.Sub(currentPos)
				}
				break
			}
		}
		// If we can't find our position, return no movement
		return gruid.Point{}
	}

	// Normal case: move to next position in path
	if len(path) > 1 {
		nextPos := path[1]
		return nextPos.Sub(currentPos)
	}

	// We're at the destination
	return gruid.Point{}
}

// IsPathValid checks if a cached path is still valid
func (pm *PathfindingManager) IsPathValid(path []gruid.Point) bool {
	if len(path) == 0 {
		return false
	}

	// Check if all positions in the path are still walkable
	for _, p := range path {
		if !pm.isWalkable(p) {
			return false
		}
	}

	// Future: Add more sophisticated validation
	// - Check if blocking entities have moved
	// - Check if map has changed
	
	return true
}

// UpdatePathfinding updates pathfinding for an entity with enhanced logic
func (pm *PathfindingManager) UpdatePathfinding(entityID ecs.EntityID, targetPos gruid.Point, strategy PathfindingStrategy) {
	currentPos := pm.game.ecs.GetPositionSafe(entityID)

	// Get or create pathfinding component
	pathComp := pm.game.ecs.GetPathfindingComponentSafe(entityID)
	if pathComp == nil {
		// Create new pathfinding component
		newPathComp := components.NewPathfindingComponent()
		pm.game.ecs.AddComponent(entityID, components.CPathfindingComponent, newPathComp)
		pathComp = &newPathComp
	}

	// Check if we need to recompute the path
	needsRecompute := pm.shouldRecomputePath(pathComp, targetPos, currentPos, strategy)

	if needsRecompute {
		// Apply group pathfinding considerations
		adjustedStrategy := pm.applyGroupPathfindingStrategy(entityID, strategy, targetPos)

		// Compute new path
		newPath := pm.FindPath(currentPos, targetPos, adjustedStrategy)
		pathComp.CurrentPath = newPath
		pathComp.PathValid = (newPath != nil)
		pathComp.Strategy = int(adjustedStrategy)
		pathComp.TargetPos = targetPos // Set the target position
		pathComp.MarkRecomputed(pm.game.stats.TurnCount)

		// Update the component in ECS
		pm.game.ecs.AddComponent(entityID, components.CPathfindingComponent, *pathComp)

		logrus.Debugf("Pathfinding: Recomputed path for entity %d, strategy: %v", entityID, adjustedStrategy)
	}
}

// shouldRecomputePath determines if a path needs to be recalculated
func (pm *PathfindingManager) shouldRecomputePath(pathComp *components.PathfindingComponent, targetPos, currentPos gruid.Point, strategy PathfindingStrategy) bool {
	// Target has changed
	if pathComp.TargetPos != targetPos {
		return true
	}

	// Strategy has changed
	if pathComp.Strategy != int(strategy) {
		return true
	}

	// Path is invalid or empty
	if !pathComp.PathValid || len(pathComp.CurrentPath) == 0 {
		return true
	}

	// Path is no longer valid (blocked by new obstacles)
	if !pm.IsPathValid(pathComp.CurrentPath) {
		return true
	}

	// Periodic recomputation for dynamic environments
	if pathComp.NeedsRecompute(pm.game.stats.TurnCount) {
		return true
	}

	// Path is significantly longer than optimal (dynamic optimization)
	if len(pathComp.CurrentPath) > 0 {
		directDistance := paths.DistanceManhattan(currentPos, targetPos)
		if len(pathComp.CurrentPath) > directDistance*2 {
			return true // Path is too inefficient
		}
	}

	return false
}

// applyGroupPathfindingStrategy adjusts strategy based on nearby entities
func (pm *PathfindingManager) applyGroupPathfindingStrategy(entityID ecs.EntityID, strategy PathfindingStrategy, targetPos gruid.Point) PathfindingStrategy {
	currentPos := pm.game.ecs.GetPositionSafe(entityID)

	// Count nearby entities with similar targets
	nearbyEntities := pm.findNearbyEntitiesWithSimilarTargets(currentPos, targetPos, 3)

	if len(nearbyEntities) > 2 {
		// Multiple entities targeting same area - use entity avoidance
		if strategy == StrategyDirect {
			logrus.Debugf("Pathfinding: Switching to AvoidEntities strategy due to %d nearby entities", len(nearbyEntities))
			return StrategyAvoidEntities
		}
	}

	return strategy
}

// findNearbyEntitiesWithSimilarTargets finds entities near current position targeting similar area
func (pm *PathfindingManager) findNearbyEntitiesWithSimilarTargets(currentPos, targetPos gruid.Point, radius int) []ecs.EntityID {
	var nearbyEntities []ecs.EntityID

	// Search in radius around current position
	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			searchPos := currentPos.Add(gruid.Point{X: dx, Y: dy})
			if !pm.game.dungeon.InBounds(searchPos) {
				continue
			}

			entities := pm.game.ecs.EntitiesAt(searchPos)
			for _, entityID := range entities {
				pathComp := pm.game.ecs.GetPathfindingComponentSafe(entityID)
				if pathComp != nil && pathComp.PathValid {
					// Check if target is in similar area (within 5 tiles)
					if paths.DistanceManhattan(pathComp.TargetPos, targetPos) <= 5 {
						nearbyEntities = append(nearbyEntities, entityID)
					}
				}
			}
		}
	}

	return nearbyEntities
}

// GetPathfindingMove returns the next move for an entity using pathfinding
func (pm *PathfindingManager) GetPathfindingMove(entityID ecs.EntityID) gruid.Point {
	currentPos := pm.game.ecs.GetPositionSafe(entityID)
	pathComp := pm.game.ecs.GetPathfindingComponentSafe(entityID)
	
	if pathComp == nil || !pathComp.PathValid || len(pathComp.CurrentPath) == 0 {
		return gruid.Point{} // No movement
	}

	// Get next move from path
	nextMove := pm.GetNextMove(pathComp.CurrentPath, currentPos)
	
	// Update path by removing the current position
	if len(pathComp.CurrentPath) > 0 && pathComp.CurrentPath[0] == currentPos {
		pathComp.CurrentPath = pathComp.CurrentPath[1:]
		// Update the component in ECS
		pm.game.ecs.AddComponent(entityID, components.CPathfindingComponent, *pathComp)
	}
	
	return nextMove
}

// PathfindingDebugInfo contains debug information for pathfinding visualization
type PathfindingDebugInfo struct {
	EntityPaths    map[ecs.EntityID][]gruid.Point
	FailedPaths    map[ecs.EntityID]gruid.Point // Entity -> failed target
	PathStrategies map[ecs.EntityID]PathfindingStrategy
	AIStates       map[ecs.EntityID]components.AIState // AI state for color coding
	LastUpdate     int // Turn number
}

// GetDebugInfo returns current pathfinding debug information
func (pm *PathfindingManager) GetDebugInfo() *PathfindingDebugInfo {
	debug := &PathfindingDebugInfo{
		EntityPaths:    make(map[ecs.EntityID][]gruid.Point),
		FailedPaths:    make(map[ecs.EntityID]gruid.Point),
		PathStrategies: make(map[ecs.EntityID]PathfindingStrategy),
		AIStates:       make(map[ecs.EntityID]components.AIState),
		LastUpdate:     pm.game.stats.TurnCount,
	}

	// Collect pathfinding information from all entities
	entities := pm.game.ecs.GetEntitiesWithComponent(components.CPathfindingComponent)
	for _, entityID := range entities {
		pathComp := pm.game.ecs.GetPathfindingComponentSafe(entityID)
		if pathComp != nil {
			// Get AI state for color coding
			if pm.game.ecs.HasAIComponentSafe(entityID) {
				aiComp := pm.game.ecs.GetAIComponentSafe(entityID)
				debug.AIStates[entityID] = aiComp.State
			}

			if pathComp.PathValid && len(pathComp.CurrentPath) > 0 {
				// Copy the path for debug display
				debug.EntityPaths[entityID] = make([]gruid.Point, len(pathComp.CurrentPath))
				copy(debug.EntityPaths[entityID], pathComp.CurrentPath)
				debug.PathStrategies[entityID] = PathfindingStrategy(pathComp.Strategy)
			} else if pathComp.TargetPos != (gruid.Point{}) {
				// Record failed pathfinding attempts
				debug.FailedPaths[entityID] = pathComp.TargetPos
			}
		}
	}

	return debug
}

// EnablePathfindingDebug enables debug mode for pathfinding
func (pm *PathfindingManager) EnablePathfindingDebug() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Debug("Pathfinding debug mode enabled")
}

// DisablePathfindingDebug disables debug mode for pathfinding
func (pm *PathfindingManager) DisablePathfindingDebug() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.Info("Pathfinding debug mode disabled")
}

// GetPathfindingStats returns performance statistics for pathfinding
func (pm *PathfindingManager) GetPathfindingStats() map[string]interface{} {
	stats := make(map[string]interface{})

	entities := pm.game.ecs.GetEntitiesWithComponent(components.CPathfindingComponent)

	totalPaths := 0
	validPaths := 0
	totalPathLength := 0
	strategyCounts := make(map[PathfindingStrategy]int)

	for _, entityID := range entities {
		pathComp := pm.game.ecs.GetPathfindingComponentSafe(entityID)
		if pathComp != nil {
			totalPaths++
			if pathComp.PathValid {
				validPaths++
				totalPathLength += len(pathComp.CurrentPath)
			}
			strategy := PathfindingStrategy(pathComp.Strategy)
			strategyCounts[strategy]++
		}
	}

	stats["total_entities_with_pathfinding"] = totalPaths
	stats["entities_with_valid_paths"] = validPaths
	stats["average_path_length"] = 0.0
	if validPaths > 0 {
		stats["average_path_length"] = float64(totalPathLength) / float64(validPaths)
	}
	stats["strategy_distribution"] = strategyCounts

	return stats
}
