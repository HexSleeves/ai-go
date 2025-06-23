package game

import (
	"math"
	"testing"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

// createTestGame creates a minimal game instance for testing
func createTestGame() *Game {
	game := NewGame()
	game.dungeon = NewMap(10, 10)

	// Create a simple test map with walls around the edges
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			if x == 0 || x == 9 || y == 0 || y == 9 {
				game.dungeon.Grid.Set(gruid.Point{X: x, Y: y}, WallCell)
			} else {
				game.dungeon.Grid.Set(gruid.Point{X: x, Y: y}, FloorCell)
			}
		}
	}

	game.pathfindingMgr = NewPathfindingManager(game)
	return game
}

func TestNewPathfindingManager(t *testing.T) {
	game := createTestGame()

	if game.pathfindingMgr == nil {
		t.Fatal("PathfindingManager should not be nil")
	}

	if game.pathfindingMgr.game != game {
		t.Error("PathfindingManager should reference the correct game")
	}

	if game.pathfindingMgr.pathRange == nil {
		t.Error("PathRange should not be nil")
	}
}

func TestPathfindingManagerNeighbors(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	// Test neighbors in the middle of the map (should have 4 neighbors)
	center := gruid.Point{X: 5, Y: 5}
	neighbors := pm.Neighbors(center)

	if len(neighbors) != 4 {
		t.Errorf("Expected 4 neighbors for center position, got %d", len(neighbors))
	}

	// Test neighbors near a wall (should have fewer neighbors)
	nearWall := gruid.Point{X: 1, Y: 1}
	neighbors = pm.Neighbors(nearWall)

	if len(neighbors) != 2 {
		t.Errorf("Expected 2 neighbors for corner position, got %d", len(neighbors))
	}
}

func TestPathfindingManagerCost(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	from := gruid.Point{X: 2, Y: 2}
	to := gruid.Point{X: 3, Y: 2}

	// Test basic cost (should be 1 for empty space)
	cost := pm.Cost(from, to)
	if cost != 1 {
		t.Errorf("Expected cost 1 for empty space, got %d", cost)
	}

	// Add a blocking entity and test increased cost
	entityID := game.ecs.AddEntity()
	game.ecs.AddComponent(entityID, components.CPosition, to)
	game.ecs.AddComponent(entityID, components.CBlocksMovement, components.BlocksMovement{})

	cost = pm.Cost(from, to)
	if cost <= 1 {
		t.Errorf("Expected increased cost for blocked position, got %d", cost)
	}
}

func TestPathfindingManagerEstimation(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	from := gruid.Point{X: 1, Y: 1}
	to := gruid.Point{X: 4, Y: 5}

	estimation := pm.Estimation(from, to)
	expectedManhattan := int(math.Abs(float64(4-1)) + math.Abs(float64(5-1))) // 3 + 4 = 7

	if estimation != expectedManhattan {
		t.Errorf("Expected Manhattan distance %d, got %d", expectedManhattan, estimation)
	}
}

func TestFindPath(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	from := gruid.Point{X: 1, Y: 1}
	to := gruid.Point{X: 3, Y: 3}

	path := pm.FindPath(from, to, StrategyDirect)

	if path == nil {
		t.Fatal("Expected a valid path, got nil")
	}

	if len(path) == 0 {
		t.Fatal("Expected non-empty path")
	}

	// Check that path starts at 'from' and ends at 'to'
	if path[0] != from {
		t.Errorf("Path should start at %v, but starts at %v", from, path[0])
	}

	if path[len(path)-1] != to {
		t.Errorf("Path should end at %v, but ends at %v", to, path[len(path)-1])
	}
}

func TestFindPathBlocked(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	// Try to find path to a wall
	from := gruid.Point{X: 1, Y: 1}
	to := gruid.Point{X: 0, Y: 0} // This is a wall

	path := pm.FindPath(from, to, StrategyDirect)

	if path != nil {
		t.Error("Expected nil path to wall, but got a path")
	}
}

func TestGetNextMove(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	// Create a simple path
	path := []gruid.Point{
		{X: 1, Y: 1},
		{X: 2, Y: 1},
		{X: 3, Y: 1},
	}

	currentPos := gruid.Point{X: 1, Y: 1}
	nextMove := pm.GetNextMove(path, currentPos)

	expectedMove := gruid.Point{X: 1, Y: 0} // Move from (1,1) to (2,1)
	if nextMove != expectedMove {
		t.Errorf("Expected move %v, got %v", expectedMove, nextMove)
	}
}

func TestIsPathValid(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	// Test valid path
	validPath := []gruid.Point{
		{X: 1, Y: 1},
		{X: 2, Y: 1},
		{X: 3, Y: 1},
	}

	if !pm.IsPathValid(validPath) {
		t.Error("Expected valid path to be valid")
	}

	// Test invalid path (contains wall)
	invalidPath := []gruid.Point{
		{X: 1, Y: 1},
		{X: 0, Y: 0}, // Wall
		{X: 3, Y: 1},
	}

	if pm.IsPathValid(invalidPath) {
		t.Error("Expected invalid path to be invalid")
	}

	// Test empty path
	emptyPath := []gruid.Point{}
	if pm.IsPathValid(emptyPath) {
		t.Error("Expected empty path to be invalid")
	}
}

func TestUpdatePathfinding(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	// Create an entity
	entityID := game.ecs.AddEntity()
	game.ecs.AddComponent(entityID, components.CPosition, gruid.Point{X: 1, Y: 1})

	targetPos := gruid.Point{X: 3, Y: 3}

	// Update pathfinding
	pm.UpdatePathfinding(entityID, targetPos, StrategyDirect)

	// Check that pathfinding component was created
	pathComp := game.ecs.GetPathfindingComponentSafe(entityID)
	if pathComp == nil {
		t.Fatal("Expected pathfinding component to be created")
	}

	if pathComp.TargetPos != targetPos {
		t.Errorf("Expected target %v, got %v", targetPos, pathComp.TargetPos)
	}

	if !pathComp.PathValid {
		t.Error("Expected path to be valid")
	}

	if len(pathComp.CurrentPath) == 0 {
		t.Error("Expected non-empty path")
	}
}

func TestGetPathfindingMove(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	// Create an entity with a pathfinding component
	entityID := game.ecs.AddEntity()
	currentPos := gruid.Point{X: 1, Y: 1}
	game.ecs.AddComponent(entityID, components.CPosition, currentPos)

	pathComp := components.NewPathfindingComponent()
	pathComp.CurrentPath = []gruid.Point{
		{X: 1, Y: 1},
		{X: 2, Y: 1},
		{X: 3, Y: 1},
	}
	pathComp.PathValid = true
	game.ecs.AddComponent(entityID, components.CPathfindingComponent, pathComp)

	// Get next move
	nextMove := pm.GetPathfindingMove(entityID)

	expectedMove := gruid.Point{X: 1, Y: 0} // Move from (1,1) to (2,1)
	if nextMove != expectedMove {
		t.Errorf("Expected move %v, got %v", expectedMove, nextMove)
	}

	// Check that path was updated (first position removed)
	updatedPathComp := game.ecs.GetPathfindingComponentSafe(entityID)
	if len(updatedPathComp.CurrentPath) != 2 {
		t.Errorf("Expected path length 2 after move, got %d", len(updatedPathComp.CurrentPath))
	}
}

func TestPathfindingComponent(t *testing.T) {
	// Test NewPathfindingComponent
	comp := components.NewPathfindingComponent()

	if comp.UsePathfinding != true {
		t.Error("Expected pathfinding to be enabled by default")
	}

	if comp.MaxPathLength != 50 {
		t.Error("Expected default max path length of 50")
	}

	if comp.Strategy != 0 { // StrategyDirect
		t.Error("Expected default strategy to be Direct")
	}

	// Test HasPath
	if comp.HasPath() {
		t.Error("Expected new component to not have a path")
	}

	// Test SetPath
	testPath := []gruid.Point{{X: 1, Y: 1}, {X: 2, Y: 2}}
	target := gruid.Point{X: 2, Y: 2}
	comp.SetPath(testPath, target)

	if !comp.HasPath() {
		t.Error("Expected component to have a path after SetPath")
	}

	if comp.TargetPos != target {
		t.Errorf("Expected target %v, got %v", target, comp.TargetPos)
	}
}

// Phase 2 Tests

func TestJPSPathfinding(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	// Test JPS for longer distances
	from := gruid.Point{X: 1, Y: 1}
	to := gruid.Point{X: 8, Y: 8}

	path := pm.FindPath(from, to, StrategyDirect)

	if path == nil {
		t.Fatal("Expected JPS path to be found")
	}

	if len(path) == 0 {
		t.Fatal("Expected non-empty JPS path")
	}

	// Verify path endpoints
	if path[0] != from {
		t.Errorf("JPS path should start at %v, but starts at %v", from, path[0])
	}

	if path[len(path)-1] != to {
		t.Errorf("JPS path should end at %v, but ends at %v", to, path[len(path)-1])
	}
}

func TestEntityAvoidanceStrategy(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	// Create a blocking entity
	blockingEntity := game.ecs.AddEntity()
	blockingPos := gruid.Point{X: 3, Y: 3}
	game.ecs.AddComponent(blockingEntity, components.CPosition, blockingPos)
	game.ecs.AddComponent(blockingEntity, components.CBlocksMovement, components.BlocksMovement{})

	from := gruid.Point{X: 2, Y: 3}
	to := gruid.Point{X: 4, Y: 3}

	// Test with entity avoidance strategy
	path := pm.FindPath(from, to, StrategyAvoidEntities)

	if path == nil {
		t.Fatal("Expected path with entity avoidance")
	}

	// Check that the path doesn't go through the blocking entity
	for _, point := range path {
		if point == blockingPos {
			t.Error("Path should avoid blocking entity but goes through it")
		}
	}
}

func TestGroupPathfindingStrategy(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	// Create multiple entities targeting the same area
	entity1 := game.ecs.AddEntity()
	entity2 := game.ecs.AddEntity()
	entity3 := game.ecs.AddEntity()

	pos1 := gruid.Point{X: 2, Y: 2}
	pos2 := gruid.Point{X: 3, Y: 2}
	pos3 := gruid.Point{X: 4, Y: 2}
	target := gruid.Point{X: 6, Y: 6}

	game.ecs.AddComponent(entity1, components.CPosition, pos1)
	game.ecs.AddComponent(entity2, components.CPosition, pos2)
	game.ecs.AddComponent(entity3, components.CPosition, pos3)

	// Add pathfinding components targeting the same area
	pathComp1 := components.NewPathfindingComponent()
	pathComp1.TargetPos = target
	pathComp1.PathValid = true
	game.ecs.AddComponent(entity1, components.CPathfindingComponent, pathComp1)

	pathComp2 := components.NewPathfindingComponent()
	pathComp2.TargetPos = target
	pathComp2.PathValid = true
	game.ecs.AddComponent(entity2, components.CPathfindingComponent, pathComp2)

	// Test group pathfinding strategy adjustment
	adjustedStrategy := pm.applyGroupPathfindingStrategy(entity3, StrategyDirect, target)

	// Should switch to AvoidEntities due to nearby entities with similar targets
	// Note: The algorithm requires more than 2 nearby entities, so let's add one more
	entity4 := game.ecs.AddEntity()
	pos4 := gruid.Point{X: 2, Y: 3}
	game.ecs.AddComponent(entity4, components.CPosition, pos4)

	pathComp4 := components.NewPathfindingComponent()
	pathComp4.TargetPos = target
	pathComp4.PathValid = true
	game.ecs.AddComponent(entity4, components.CPathfindingComponent, pathComp4)

	// Now test again with more entities
	adjustedStrategy = pm.applyGroupPathfindingStrategy(entity3, StrategyDirect, target)

	if adjustedStrategy != StrategyAvoidEntities {
		t.Errorf("Expected StrategyAvoidEntities due to group pathfinding, got %v", adjustedStrategy)
	}
}

func TestPathfindingDebugInfo(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	// Create an entity with pathfinding
	entityID := game.ecs.AddEntity()
	game.ecs.AddComponent(entityID, components.CPosition, gruid.Point{X: 1, Y: 1})

	pm.UpdatePathfinding(entityID, gruid.Point{X: 3, Y: 3}, StrategyDirect)

	// Get debug info
	debugInfo := pm.GetDebugInfo()

	if debugInfo == nil {
		t.Fatal("Expected debug info to be non-nil")
	}

	if len(debugInfo.EntityPaths) == 0 {
		t.Error("Expected entity paths in debug info")
	}

	if _, exists := debugInfo.EntityPaths[entityID]; !exists {
		t.Error("Expected entity to have path in debug info")
	}

	if strategy, exists := debugInfo.PathStrategies[entityID]; !exists || strategy != StrategyDirect {
		t.Errorf("Expected entity to have StrategyDirect in debug info, got %v", strategy)
	}
}

func TestPathfindingStats(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	// Create entities with pathfinding
	for i := 0; i < 3; i++ {
		entityID := game.ecs.AddEntity()
		game.ecs.AddComponent(entityID, components.CPosition, gruid.Point{X: i + 1, Y: 1})
		pm.UpdatePathfinding(entityID, gruid.Point{X: i + 5, Y: 5}, StrategyDirect)
	}

	stats := pm.GetPathfindingStats()

	if stats["total_entities_with_pathfinding"] != 3 {
		t.Errorf("Expected 3 entities with pathfinding, got %v", stats["total_entities_with_pathfinding"])
	}

	if stats["entities_with_valid_paths"] != 3 {
		t.Errorf("Expected 3 entities with valid paths, got %v", stats["entities_with_valid_paths"])
	}

	if avgLength, ok := stats["average_path_length"].(float64); !ok || avgLength <= 0 {
		t.Errorf("Expected positive average path length, got %v", avgLength)
	}
}

func TestDynamicPathRecalculation(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	entityID := game.ecs.AddEntity()
	game.ecs.AddComponent(entityID, components.CPosition, gruid.Point{X: 1, Y: 1})

	// Initial pathfinding
	target1 := gruid.Point{X: 3, Y: 3}
	pm.UpdatePathfinding(entityID, target1, StrategyDirect)

	pathComp := game.ecs.GetPathfindingComponentSafe(entityID)
	if pathComp == nil || !pathComp.PathValid {
		t.Fatal("Expected valid initial path")
	}

	initialPathLength := len(pathComp.CurrentPath)

	// Change target - should trigger recalculation
	target2 := gruid.Point{X: 7, Y: 7}
	pm.UpdatePathfinding(entityID, target2, StrategyDirect)

	pathComp = game.ecs.GetPathfindingComponentSafe(entityID)
	if pathComp == nil || !pathComp.PathValid {
		t.Fatal("Expected valid recalculated path")
	}

	if pathComp.TargetPos != target2 {
		t.Errorf("Expected target to be updated to %v, got %v", target2, pathComp.TargetPos)
	}

	// Path length should be different for different target
	newPathLength := len(pathComp.CurrentPath)
	if newPathLength == initialPathLength {
		t.Error("Expected different path length after target change")
	}
}

func TestStealthyStrategy(t *testing.T) {
	game := createTestGame()
	pm := game.pathfindingMgr

	// Create player with FOV
	game.PlayerID = game.ecs.AddEntity()
	playerPos := gruid.Point{X: 5, Y: 5}
	game.ecs.AddComponent(game.PlayerID, components.CPosition, playerPos)

	// Create FOV component for player
	fovComp := components.NewFOVComponent(6, game.dungeon.Width, game.dungeon.Height)
	game.ecs.AddComponent(game.PlayerID, components.CFOV, fovComp)

	// Compute FOV using the game's FOV system
	game.FOVSystem()

	// Test stealthy pathfinding
	from := gruid.Point{X: 2, Y: 2}
	to := gruid.Point{X: 8, Y: 8}

	path := pm.FindPath(from, to, StrategyStealthy)

	if path == nil {
		t.Fatal("Expected stealthy path to be found")
	}

	// Check that the path tries to avoid player FOV when possible
	// (This is a basic test - in practice, the path might still go through FOV if necessary)
	if len(path) == 0 {
		t.Error("Expected non-empty stealthy path")
	}
}
