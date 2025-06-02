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

// FindPath computes a path from start to goal using A* algorithm
func (pm *PathfindingManager) FindPath(from, to gruid.Point, strategy PathfindingStrategy) []gruid.Point {
	if !pm.isWalkable(from) || !pm.isWalkable(to) {
		logrus.Debugf("Pathfinding: Invalid start (%v) or goal (%v) position", from, to)
		return nil
	}

	// Use A* pathfinding
	path := pm.pathRange.AstarPath(pm, from, to)
	
	if path == nil {
		logrus.Debugf("Pathfinding: No path found from %v to %v", from, to)
		return nil
	}

	// Apply strategy-specific modifications
	path = pm.applyStrategy(path, strategy)
	
	logrus.Debugf("Pathfinding: Found path of length %d from %v to %v", len(path), from, to)
	return path
}

// applyStrategy modifies the path based on the pathfinding strategy
func (pm *PathfindingManager) applyStrategy(path []gruid.Point, strategy PathfindingStrategy) []gruid.Point {
	switch strategy {
	case StrategyDirect:
		// No modifications needed
		return path
	case StrategyAvoidEntities:
		// Future: Implement entity avoidance path smoothing
		return path
	case StrategyPreferOpen:
		// Future: Implement open area preference
		return path
	case StrategyStealthy:
		// Future: Implement FOV avoidance
		return path
	default:
		return path
	}
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

// UpdatePathfinding updates pathfinding for an entity
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
	needsRecompute := false
	
	if pathComp.TargetPos != targetPos {
		// Target has changed
		needsRecompute = true
		pathComp.TargetPos = targetPos
	}
	
	if !pathComp.PathValid || !pm.IsPathValid(pathComp.CurrentPath) {
		// Path is invalid
		needsRecompute = true
	}
	
	if len(pathComp.CurrentPath) == 0 {
		// No path exists
		needsRecompute = true
	}

	if needsRecompute {
		// Compute new path
		newPath := pm.FindPath(currentPos, targetPos, strategy)
		pathComp.CurrentPath = newPath
		pathComp.PathValid = (newPath != nil)
		pathComp.Strategy = int(strategy)
		
		// Update the component in ECS
		pm.game.ecs.AddComponent(entityID, components.CPathfindingComponent, *pathComp)
	}
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
