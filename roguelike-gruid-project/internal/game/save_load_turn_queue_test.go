package game

import (
	"testing"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	turn "github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/turn_queue"
)

func TestSaveLoadTurnQueue(t *testing.T) {
	// Create and initialize game
	game := NewGame()
	game.InitLevel()

	playerID := game.PlayerID

	// Spawn several monsters to populate the turn queue
	monsterPositions := []gruid.Point{
		game.GetPlayerPosition().Add(gruid.Point{X: 2, Y: 0}),
		game.GetPlayerPosition().Add(gruid.Point{X: 0, Y: 2}),
		game.GetPlayerPosition().Add(gruid.Point{X: -2, Y: 0}),
		game.GetPlayerPosition().Add(gruid.Point{X: 0, Y: -2}),
	}

	var monsterIDs []ecs.EntityID
	for _, pos := range monsterPositions {
		game.SpawnMonster(pos)
		// Find the monster at this position
		for _, id := range game.ecs.EntitiesAt(pos) {
			if game.ecs.HasComponent(id, components.CAITag) {
				monsterIDs = append(monsterIDs, id)
				break
			}
		}
	}

	if len(monsterIDs) != len(monsterPositions) {
		t.Fatalf("Expected %d monsters, got %d", len(monsterPositions), len(monsterIDs))
	}

	// Advance time and add entities to turn queue with different times
	game.turnQueue.CurrentTime = 1000
	
	// Add player to queue
	game.turnQueue.Add(playerID, 1100)
	
	// Add monsters to queue with different times
	for i, monsterID := range monsterIDs {
		game.turnQueue.Add(monsterID, uint64(1050+i*25)) // 1050, 1075, 1100, 1125
	}

	// Record initial turn queue state
	initialCurrentTime := game.turnQueue.CurrentTime
	initialSnapshot := game.turnQueue.Snapshot()
	initialQueueSize := game.turnQueue.Len()

	if initialQueueSize == 0 {
		t.Fatal("Turn queue should not be empty before saving")
	}

	t.Logf("Initial queue size: %d", initialQueueSize)
	t.Logf("Initial current time: %d", initialCurrentTime)

	// Save the game
	err := game.SaveGame()
	if err != nil {
		t.Fatalf("Failed to save game: %v", err)
	}

	// Create new game and load
	newGame := NewGame()
	newGame.InitLevel()

	// Verify new game starts with different queue state
	newGameInitialSize := newGame.turnQueue.Len()
	t.Logf("New game initial queue size: %d", newGameInitialSize)

	err = newGame.LoadGame()
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Verify turn queue was restored correctly
	loadedCurrentTime := newGame.turnQueue.CurrentTime
	loadedSnapshot := newGame.turnQueue.Snapshot()
	loadedQueueSize := newGame.turnQueue.Len()

	// Check current time
	if loadedCurrentTime != initialCurrentTime {
		t.Errorf("Current time mismatch: expected %d, got %d", initialCurrentTime, loadedCurrentTime)
	}

	// Check queue size
	if loadedQueueSize != initialQueueSize {
		t.Errorf("Queue size mismatch: expected %d, got %d", initialQueueSize, loadedQueueSize)
	}

	// Check that all entries were restored
	if len(loadedSnapshot) != len(initialSnapshot) {
		t.Errorf("Snapshot length mismatch: expected %d, got %d", len(initialSnapshot), len(loadedSnapshot))
	}

	// Create maps for easier comparison (since heap order might differ)
	initialEntries := make(map[ecs.EntityID]uint64)
	for _, entry := range initialSnapshot {
		initialEntries[entry.EntityID] = entry.Time
	}

	loadedEntries := make(map[ecs.EntityID]uint64)
	for _, entry := range loadedSnapshot {
		loadedEntries[entry.EntityID] = entry.Time
	}

	// Verify all entities and their times match
	for entityID, expectedTime := range initialEntries {
		if loadedTime, exists := loadedEntries[entityID]; !exists {
			t.Errorf("Entity %d missing from loaded queue", entityID)
		} else if loadedTime != expectedTime {
			t.Errorf("Entity %d time mismatch: expected %d, got %d", entityID, expectedTime, loadedTime)
		}
	}

	// Verify no extra entities in loaded queue
	for entityID := range loadedEntries {
		if _, exists := initialEntries[entityID]; !exists {
			t.Errorf("Unexpected entity %d in loaded queue", entityID)
		}
	}

	t.Logf("Successfully restored turn queue with %d entries", loadedQueueSize)
}

func TestTurnQueueProcessingAfterLoad(t *testing.T) {
	// Create and initialize game
	game := NewGame()
	game.InitLevel()

	playerID := game.PlayerID

	// Spawn a monster
	monsterPos := game.GetPlayerPosition().Add(gruid.Point{X: 1, Y: 0})
	game.SpawnMonster(monsterPos)

	var monsterID ecs.EntityID
	for _, id := range game.ecs.EntitiesAt(monsterPos) {
		if game.ecs.HasComponent(id, components.CAITag) {
			monsterID = id
			break
		}
	}

	if monsterID == 0 {
		t.Fatal("Should have spawned a monster")
	}

	// Clear existing queue and set up with specific times
	game.turnQueue = turn.NewTurnQueue()
	game.turnQueue.CurrentTime = 100
	game.turnQueue.Add(monsterID, 150) // Monster goes first
	game.turnQueue.Add(playerID, 200)  // Player goes second

	// Save and load
	err := game.SaveGame()
	if err != nil {
		t.Fatalf("Failed to save game: %v", err)
	}

	newGame := NewGame()
	newGame.InitLevel()
	err = newGame.LoadGame()
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Verify turn queue processing works correctly after load
	if newGame.turnQueue.IsEmpty() {
		t.Fatal("Turn queue should not be empty after load")
	}

	// Get next turn - should be the monster (time 150)
	nextEntry, hasNext := newGame.turnQueue.Peek()
	if !hasNext {
		t.Fatal("Should have next turn available")
	}

	if nextEntry.EntityID != monsterID {
		t.Errorf("Expected monster to be next, got entity %d", nextEntry.EntityID)
	}

	if nextEntry.Time != 150 {
		t.Errorf("Expected monster turn time to be 150, got %d", nextEntry.Time)
	}

	// Process the turn
	actualEntry, ok := newGame.turnQueue.Next()
	if !ok {
		t.Fatal("Should be able to get next turn")
	}

	if actualEntry.EntityID != nextEntry.EntityID || actualEntry.Time != nextEntry.Time {
		t.Error("Next() should return the same entry as Peek()")
	}

	// Verify player is next
	playerEntry, hasPlayerNext := newGame.turnQueue.Peek()
	if !hasPlayerNext {
		t.Fatal("Should have player turn available")
	}

	if playerEntry.EntityID != playerID {
		t.Errorf("Expected player to be next, got entity %d", playerEntry.EntityID)
	}

	if playerEntry.Time != 200 {
		t.Errorf("Expected player turn time to be 200, got %d", playerEntry.Time)
	}
}

func TestEmptyTurnQueueSaveLoad(t *testing.T) {
	// Test saving and loading with an empty turn queue
	game := NewGame()
	game.InitLevel()

	// Create a completely empty turn queue
	game.turnQueue = turn.NewTurnQueue()
	game.turnQueue.CurrentTime = 500

	if !game.turnQueue.IsEmpty() {
		t.Fatal("Turn queue should be empty for this test")
	}

	// Save and load
	err := game.SaveGame()
	if err != nil {
		t.Fatalf("Failed to save game: %v", err)
	}

	newGame := NewGame()
	newGame.InitLevel()
	err = newGame.LoadGame()
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Verify queue is still empty
	if !newGame.turnQueue.IsEmpty() {
		t.Error("Turn queue should remain empty after load")
	}

	if newGame.turnQueue.Len() != 0 {
		t.Errorf("Expected queue length 0, got %d", newGame.turnQueue.Len())
	}
}

func TestTurnQueueSnapshotMethods(t *testing.T) {
	// Test the new Snapshot and RestoreFromSnapshot methods directly
	game := NewGame()
	game.InitLevel()

	// Create a fresh queue for testing
	game.turnQueue = turn.NewTurnQueue()

	// Add some entries to the queue
	game.turnQueue.Add(1, 100)
	game.turnQueue.Add(2, 150)
	game.turnQueue.Add(3, 125)

	// Take a snapshot
	snapshot := game.turnQueue.Snapshot()

	if len(snapshot) != 3 {
		t.Errorf("Expected snapshot length 3, got %d", len(snapshot))
	}

	// Verify snapshot contains all entries
	entityTimes := make(map[ecs.EntityID]uint64)
	for _, entry := range snapshot {
		entityTimes[entry.EntityID] = entry.Time
	}

	expectedTimes := map[ecs.EntityID]uint64{
		1: 100,
		2: 150,
		3: 125,
	}

	for entityID, expectedTime := range expectedTimes {
		if actualTime, exists := entityTimes[entityID]; !exists {
			t.Errorf("Entity %d missing from snapshot", entityID)
		} else if actualTime != expectedTime {
			t.Errorf("Entity %d time mismatch: expected %d, got %d", entityID, expectedTime, actualTime)
		}
	}

	// Clear the queue and restore from snapshot
	game.turnQueue.RestoreFromSnapshot([]turn.TurnEntry{})
	if !game.turnQueue.IsEmpty() {
		t.Error("Queue should be empty after restoring from empty snapshot")
	}

	// Restore from original snapshot
	game.turnQueue.RestoreFromSnapshot(snapshot)

	if game.turnQueue.Len() != 3 {
		t.Errorf("Expected queue length 3 after restore, got %d", game.turnQueue.Len())
	}

	// Verify entries are restored correctly by processing them in order
	expectedOrder := []struct {
		entityID ecs.EntityID
		time     uint64
	}{
		{1, 100}, // Lowest time first
		{3, 125}, // Middle time
		{2, 150}, // Highest time
	}

	for i, expected := range expectedOrder {
		entry, ok := game.turnQueue.Next()
		if !ok {
			t.Fatalf("Expected entry %d to be available", i)
		}

		if entry.EntityID != expected.entityID {
			t.Errorf("Entry %d: expected entity %d, got %d", i, expected.entityID, entry.EntityID)
		}

		if entry.Time != expected.time {
			t.Errorf("Entry %d: expected time %d, got %d", i, expected.time, entry.Time)
		}
	}

	// Queue should now be empty
	if !game.turnQueue.IsEmpty() {
		t.Error("Queue should be empty after processing all entries")
	}
}
