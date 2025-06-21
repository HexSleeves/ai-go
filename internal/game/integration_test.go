package game

import (
	"testing"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

func TestInventorySystem(t *testing.T) {
	game := NewGame()
	game.InitLevel()

	playerID := game.PlayerID

	// Test that player has inventory
	if !game.ecs.HasInventorySafe(playerID) {
		t.Error("Player should have inventory component")
	}

	// Test that player has equipment
	if !game.ecs.HasEquipmentSafe(playerID) {
		t.Error("Player should have equipment component")
	}

	// Test starting items
	inventory := game.ecs.GetInventorySafe(playerID)
	if len(inventory.Items) == 0 {
		t.Error("Player should have starting items")
	}

	// Test equipment
	equipment := game.ecs.GetEquipmentSafe(playerID)
	if equipment.Weapon == nil {
		t.Error("Player should have starting weapon equipped")
	}
	if equipment.Armor == nil {
		t.Error("Player should have starting armor equipped")
	}
}

func TestCharacterProgression(t *testing.T) {
	game := NewGame()
	game.InitLevel()

	playerID := game.PlayerID

	// Test that player has all progression components
	if !game.ecs.HasStatsSafe(playerID) {
		t.Error("Player should have stats component")
	}
	if !game.ecs.HasExperienceSafe(playerID) {
		t.Error("Player should have experience component")
	}
	if !game.ecs.HasSkillsSafe(playerID) {
		t.Error("Player should have skills component")
	}
	if !game.ecs.HasCombatSafe(playerID) {
		t.Error("Player should have combat component")
	}
	if !game.ecs.HasManaSafe(playerID) {
		t.Error("Player should have mana component")
	}
	if !game.ecs.HasStaminaSafe(playerID) {
		t.Error("Player should have stamina component")
	}

	// Test initial values
	experience := game.ecs.GetExperienceSafe(playerID)
	if experience.Level != 1 {
		t.Errorf("Player should start at level 1, got %d", experience.Level)
	}
	if experience.CurrentXP != 0 {
		t.Errorf("Player should start with 0 XP, got %d", experience.CurrentXP)
	}

	stats := game.ecs.GetStatsSafe(playerID)
	if stats.Strength != 10 {
		t.Errorf("Player should start with 10 strength, got %d", stats.Strength)
	}
}

func TestExperienceSystem(t *testing.T) {
	game := NewGame()
	game.InitLevel()

	playerID := game.PlayerID
	expSystem := NewExperienceSystem(game)

	// Test experience gain
	initialExp := game.ecs.GetExperienceSafe(playerID)
	expSystem.AwardExperience(playerID, 100)

	newExp := game.ecs.GetExperienceSafe(playerID)
	if newExp.TotalXP != initialExp.TotalXP+100 {
		t.Errorf("Expected total XP to increase by 100, got %d", newExp.TotalXP-initialExp.TotalXP)
	}

	// Test level up (give enough XP to level up)
	expSystem.AwardExperience(playerID, 1000)
	leveledUpExp := game.ecs.GetExperienceSafe(playerID)
	if leveledUpExp.Level <= initialExp.Level {
		t.Error("Player should have leveled up")
	}
	if leveledUpExp.SkillPoints <= 0 {
		t.Error("Player should have gained skill points from leveling up")
	}
	if leveledUpExp.AttributePoints <= 0 {
		t.Error("Player should have gained attribute points from leveling up")
	}
}

func TestAISystem(t *testing.T) {
	game := NewGame()
	game.InitLevel()

	// Spawn a monster with AI
	pos := game.GetPlayerPosition().Add(gruid.Point{X: 5, Y: 5})
	game.SpawnMonster(pos)

	// Find the monster
	var monsterID ecs.EntityID
	for _, id := range game.ecs.GetAllEntities() {
		if game.ecs.HasComponent(id, components.CAITag) && id != game.PlayerID {
			monsterID = id
			break
		}
	}

	if monsterID == 0 {
		t.Error("Should have spawned a monster with AI")
	}

	// Test that monster has AI component
	if !game.ecs.HasAIComponentSafe(monsterID) {
		t.Error("Monster should have AI component")
	}

	// Test AI action generation
	action := game.AdvancedMonsterAI(monsterID)
	if action == nil {
		t.Error("AI should generate an action")
	}
}

func TestSaveLoadSystem(t *testing.T) {
	game := NewGame()
	game.InitLevel()

	// Modify game state
	playerID := game.PlayerID
	expSystem := NewExperienceSystem(game)
	expSystem.AwardExperience(playerID, 50)

	// Save the game
	err := game.SaveGame()
	if err != nil {
		t.Errorf("Failed to save game: %v", err)
	}

	// Verify save file exists
	if !HasSaveFile() {
		t.Error("Save file should exist after saving")
	}

	// Create new game and load
	newGame := NewGame()
	newGame.InitLevel()

	err = newGame.LoadGame()
	if err != nil {
		t.Errorf("Failed to load game: %v", err)
	}

	// Verify loaded state
	loadedExp := newGame.ecs.GetExperienceSafe(newGame.PlayerID)
	if loadedExp.TotalXP != 50 {
		t.Errorf("Expected loaded XP to be 50, got %d", loadedExp.TotalXP)
	}
}

func TestInventoryActions(t *testing.T) {
	game := NewGame()
	game.InitLevel()

	playerID := game.PlayerID
	playerPos := game.GetPlayerPosition()

	// Create a test item
	items := CreateBasicItems()
	potion := items["Health Potion"]

	// Spawn item near player
	itemID := game.SpawnItem(potion, 1, playerPos)

	// Get initial potion count (player starts with some)
	initialInventory := game.ecs.GetInventorySafe(playerID)
	initialPotionCount := initialInventory.GetItemCount("Health Potion")

	// Test pickup action
	pickupAction := PickupAction{EntityID: playerID, ItemID: itemID}
	_, err := pickupAction.Execute(game)
	if err != nil {
		t.Errorf("Pickup action failed: %v", err)
	}

	// Verify item was picked up
	if game.ecs.EntityExists(itemID) {
		t.Error("Item should be removed from world after pickup")
	}

	afterPickupInventory := game.ecs.GetInventorySafe(playerID)
	afterPickupCount := afterPickupInventory.GetItemCount("Health Potion")
	if afterPickupCount != initialPotionCount+1 {
		t.Errorf("Expected potion count to increase by 1: initial=%d, after=%d", initialPotionCount, afterPickupCount)
	}

	// Test use action
	useAction := UseItemAction{EntityID: playerID, ItemName: "Health Potion"}
	err = useAction.Execute(game)
	if err != nil {
		t.Errorf("Use action failed: %v", err)
	}

	// Verify item was consumed
	afterUseInventory := game.ecs.GetInventorySafe(playerID)
	afterUseCount := afterUseInventory.GetItemCount("Health Potion")
	if afterUseCount != afterPickupCount-1 {
		t.Errorf("Potion count should decrease by 1 after use: before=%d, after=%d", afterPickupCount, afterUseCount)
	}
}

func TestCombatWithExperience(t *testing.T) {
	game := NewGame()
	game.InitLevel()

	playerID := game.PlayerID

	// Spawn a monster
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
		t.Error("Should have spawned a monster")
	}

	// Record initial experience
	initialExp := game.ecs.GetExperienceSafe(playerID)

	// Attack the monster until it dies
	for i := 0; i < 20; i++ { // Safety limit
		if !game.ecs.EntityExists(monsterID) {
			break
		}

		// Check if monster is still alive
		if game.ecs.HasHealthSafe(monsterID) {
			health := game.ecs.GetHealthSafe(monsterID)
			if health.IsDead() {
				break
			}
		}

		attackAction := AttackAction{AttackerID: playerID, TargetID: monsterID}
		_, err := attackAction.Execute(game)
		if err != nil {
			// Monster might have died, break the loop
			break
		}
	}

	// Verify monster is dead
	if game.ecs.EntityExists(monsterID) && game.ecs.HasHealthSafe(monsterID) {
		health := game.ecs.GetHealthSafe(monsterID)
		if !health.IsDead() {
			t.Error("Monster should be dead after multiple attacks")
		}
	}

	// Verify experience was gained
	finalExp := game.ecs.GetExperienceSafe(playerID)
	if finalExp.TotalXP <= initialExp.TotalXP {
		t.Error("Player should have gained experience from killing monster")
	}
}
