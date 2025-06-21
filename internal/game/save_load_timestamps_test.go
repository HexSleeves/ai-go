package game

import (
	"testing"
	"time"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

func TestSaveLoadMessageTimestamps(t *testing.T) {
	// Create and initialize game
	game := NewGame()
	game.InitLevel()

	// Add some messages with different timestamps
	baseTime := time.Now()
	game.log.AddMessageWithTimestamp("First message", gruid.Color(1), baseTime)
	game.log.AddMessageWithTimestamp("Second message", gruid.Color(2), baseTime.Add(time.Minute))
	game.log.AddMessageWithTimestamp("Third message", gruid.Color(3), baseTime.Add(2*time.Minute))

	// Record initial message state
	initialMessages := make([]struct {
		Text      string
		Color     gruid.Color
		Timestamp time.Time
	}, len(game.log.Messages))

	for i, msg := range game.log.Messages {
		initialMessages[i] = struct {
			Text      string
			Color     gruid.Color
			Timestamp time.Time
		}{
			Text:      msg.Text,
			Color:     msg.Color,
			Timestamp: msg.Timestamp,
		}
	}

	// Save the game
	err := game.SaveGame()
	if err != nil {
		t.Fatalf("Failed to save game: %v", err)
	}

	// Create new game and load
	newGame := NewGame()
	newGame.InitLevel()

	err = newGame.LoadGame()
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Verify message timestamps were restored correctly
	if len(newGame.log.Messages) != len(initialMessages) {
		t.Errorf("Message count mismatch: expected %d, got %d", len(initialMessages), len(newGame.log.Messages))
	}

	for i, expectedMsg := range initialMessages {
		if i >= len(newGame.log.Messages) {
			t.Errorf("Missing message at index %d", i)
			continue
		}

		loadedMsg := newGame.log.Messages[i]

		if loadedMsg.Text != expectedMsg.Text {
			t.Errorf("Message text mismatch at index %d: expected %s, got %s", i, expectedMsg.Text, loadedMsg.Text)
		}

		if loadedMsg.Color != expectedMsg.Color {
			t.Errorf("Message color mismatch at index %d: expected %d, got %d", i, expectedMsg.Color, loadedMsg.Color)
		}

		if !loadedMsg.Timestamp.Equal(expectedMsg.Timestamp) {
			t.Errorf("Message timestamp mismatch at index %d: expected %v, got %v", i, expectedMsg.Timestamp, loadedMsg.Timestamp)
		}
	}
}

func TestSaveLoadGameStatistics(t *testing.T) {
	// Create and initialize game
	game := NewGame()
	game.InitLevel()

	// Simulate some game activity to generate statistics
	// Simulate damage dealt and taken
	game.AddDamageDealt(50)
	game.AddDamageTaken(25)

	// Simulate monster kills
	game.IncrementMonstersKilled()
	game.IncrementMonstersKilled()

	// Simulate item collection
	game.IncrementItemsCollected()
	game.IncrementItemsCollected()
	game.IncrementItemsCollected()

	// Wait a bit to accumulate play time
	time.Sleep(10 * time.Millisecond)
	game.UpdatePlayTime()

	// Record initial statistics
	initialStats := *game.stats

	// Save the game
	err := game.SaveGame()
	if err != nil {
		t.Fatalf("Failed to save game: %v", err)
	}

	// Create new game and load
	newGame := NewGame()
	newGame.InitLevel()

	err = newGame.LoadGame()
	if err != nil {
		t.Fatalf("Failed to load game: %v", err)
	}

	// Verify game statistics were restored correctly
	if newGame.stats == nil {
		t.Fatal("Game statistics should not be nil after load")
	}

	loadedStats := *newGame.stats

	if loadedStats.MonstersKilled != initialStats.MonstersKilled {
		t.Errorf("Monsters killed mismatch: expected %d, got %d", initialStats.MonstersKilled, loadedStats.MonstersKilled)
	}

	if loadedStats.ItemsCollected != initialStats.ItemsCollected {
		t.Errorf("Items collected mismatch: expected %d, got %d", initialStats.ItemsCollected, loadedStats.ItemsCollected)
	}

	if loadedStats.DamageDealt != initialStats.DamageDealt {
		t.Errorf("Damage dealt mismatch: expected %d, got %d", initialStats.DamageDealt, loadedStats.DamageDealt)
	}

	if loadedStats.DamageTaken != initialStats.DamageTaken {
		t.Errorf("Damage taken mismatch: expected %d, got %d", initialStats.DamageTaken, loadedStats.DamageTaken)
	}

	// Play time should be approximately the same (allowing for small differences due to save/load time)
	timeDiff := loadedStats.PlayTime - initialStats.PlayTime
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}
	if timeDiff > time.Second {
		t.Errorf("Play time difference too large: expected ~%v, got %v (diff: %v)", initialStats.PlayTime, loadedStats.PlayTime, timeDiff)
	}
}

func TestGameStatisticsIntegration(t *testing.T) {
	// Create and initialize game
	game := NewGame()
	game.InitLevel()

	playerID := game.PlayerID

	// Test damage tracking through combat
	monsterPos := game.GetPlayerPosition().Add(gruid.Point{X: 1, Y: 0})
	game.SpawnMonster(monsterPos)

	// Find the monster
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

	// Record initial stats
	initialDamageDealt := game.stats.DamageDealt
	initialMonstersKilled := game.stats.MonstersKilled

	// Attack the monster
	attackAction := AttackAction{AttackerID: playerID, TargetID: monsterID}
	_, err := attackAction.Execute(game)
	if err != nil {
		t.Errorf("Attack action failed: %v", err)
	}

	// Verify damage was tracked
	if game.stats.DamageDealt <= initialDamageDealt {
		t.Error("Damage dealt should have increased after attack")
	}

	// Kill the monster (attack until dead)
	for i := 0; i < 20; i++ { // Safety limit
		if !game.ecs.EntityExists(monsterID) {
			break
		}

		if game.ecs.HasHealthSafe(monsterID) {
			health := game.ecs.GetHealthSafe(monsterID)
			if health.IsDead() {
				break
			}
		}

		attackAction := AttackAction{AttackerID: playerID, TargetID: monsterID}
		_, err := attackAction.Execute(game)
		if err != nil {
			break
		}
	}

	// Verify monster kill was tracked
	if game.stats.MonstersKilled <= initialMonstersKilled {
		t.Error("Monsters killed should have increased after killing monster")
	}

	// Test item collection tracking
	items := CreateBasicItems()
	potion := items["Health Potion"]
	itemPos := game.GetPlayerPosition().Add(gruid.Point{X: 2, Y: 0})
	itemID := game.SpawnItem(potion, 1, itemPos)

	initialItemsCollected := game.stats.ItemsCollected

	// Pick up the item
	pickupAction := PickupAction{EntityID: playerID, ItemID: itemID}
	_, err = pickupAction.Execute(game)
	if err != nil {
		t.Errorf("Pickup action failed: %v", err)
	}

	// Verify item collection was tracked
	if game.stats.ItemsCollected <= initialItemsCollected {
		t.Error("Items collected should have increased after pickup")
	}
}

func TestPlayTimeTracking(t *testing.T) {
	// Create and initialize game
	game := NewGame()
	game.InitLevel()

	// Record start time
	startTime := game.stats.StartTime

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Update play time
	game.UpdatePlayTime()

	// Verify play time increased
	if game.stats.PlayTime <= 0 {
		t.Error("Play time should be greater than 0 after waiting")
	}

	// Verify start time hasn't changed
	if !game.stats.StartTime.Equal(startTime) {
		t.Error("Start time should not change during play")
	}

	// Test that play time continues to increase
	firstPlayTime := game.stats.PlayTime
	time.Sleep(50 * time.Millisecond)
	game.UpdatePlayTime()

	if game.stats.PlayTime <= firstPlayTime {
		t.Error("Play time should continue to increase")
	}
}
