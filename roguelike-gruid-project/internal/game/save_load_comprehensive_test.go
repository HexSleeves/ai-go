package game

import (
	"testing"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

func TestSaveLoadAllComponents(t *testing.T) {
	// Create and initialize game
	game := NewGame()
	game.InitLevel()

	playerID := game.PlayerID

	// Add all possible components to the player for comprehensive testing
	expSystem := NewExperienceSystem(game)
	expSystem.AwardExperience(playerID, 150) // Should level up

	// Add status effects
	if game.ecs.HasStatusEffectsSafe(playerID) {
		statusEffects := game.ecs.GetStatusEffectsSafe(playerID)
		statusEffects.Effects = append(statusEffects.Effects, components.StatusEffect{
			Name:        "Test Buff",
			Description: "A test status effect",
			Duration:    10,
			StrengthMod: 2,
			AttackMod:   1,
		})
		game.ecs.AddComponent(playerID, components.CStatusEffects, statusEffects)
	}

	// Spawn a monster with AI to test AI component save/load
	monsterPos := game.GetPlayerPosition().Add(gruid.Point{X: 3, Y: 3})
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

	// Modify monster AI state for testing
	if game.ecs.HasAIComponentSafe(monsterID) {
		err := game.ecs.UpdateAIComponent(monsterID, func(ai *components.AIComponent) error {
			ai.State = components.AIStateChasing
			ai.SearchTurns = 5
			ai.LastKnownPlayerPos = gruid.Point{X: 10, Y: 10}
			return nil
		})
		if err != nil {
			t.Errorf("Failed to update AI component: %v", err)
		}
	}

	// Spawn an item to test item pickup component
	items := CreateBasicItems()
	potion := items["Health Potion"]
	itemPos := game.GetPlayerPosition().Add(gruid.Point{X: 1, Y: 1})
	itemID := game.SpawnItem(potion, 2, itemPos)

	// Record initial state for comparison
	initialPlayerExp := game.ecs.GetExperienceSafe(playerID)
	initialPlayerStats := game.ecs.GetStatsSafe(playerID)
	initialPlayerSkills := game.ecs.GetSkillsSafe(playerID)
	initialPlayerCombat := game.ecs.GetCombatSafe(playerID)
	initialPlayerMana := game.ecs.GetManaSafe(playerID)
	initialPlayerStamina := game.ecs.GetStaminaSafe(playerID)
	initialPlayerStatusEffects := game.ecs.GetStatusEffectsSafe(playerID)
	initialPlayerInventory := game.ecs.GetInventorySafe(playerID)
	initialPlayerEquipment := game.ecs.GetEquipmentSafe(playerID)

	var initialMonsterAI components.AIComponent
	if game.ecs.HasAIComponentSafe(monsterID) {
		initialMonsterAI = game.ecs.GetAIComponentSafe(monsterID)
	}

	var initialItemPickup components.ItemPickup
	if game.ecs.HasItemPickupSafe(itemID) {
		initialItemPickup = game.ecs.GetItemPickupSafe(itemID)
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

	// Verify all components were restored correctly
	loadedPlayerID := newGame.PlayerID

	// Test Experience component
	if !newGame.ecs.HasExperienceSafe(loadedPlayerID) {
		t.Error("Player should have experience component after load")
	} else {
		loadedExp := newGame.ecs.GetExperienceSafe(loadedPlayerID)
		if loadedExp.Level != initialPlayerExp.Level {
			t.Errorf("Experience level mismatch: expected %d, got %d", initialPlayerExp.Level, loadedExp.Level)
		}
		if loadedExp.TotalXP != initialPlayerExp.TotalXP {
			t.Errorf("Total XP mismatch: expected %d, got %d", initialPlayerExp.TotalXP, loadedExp.TotalXP)
		}
		if loadedExp.SkillPoints != initialPlayerExp.SkillPoints {
			t.Errorf("Skill points mismatch: expected %d, got %d", initialPlayerExp.SkillPoints, loadedExp.SkillPoints)
		}
	}

	// Test Stats component
	if !newGame.ecs.HasStatsSafe(loadedPlayerID) {
		t.Error("Player should have stats component after load")
	} else {
		loadedStats := newGame.ecs.GetStatsSafe(loadedPlayerID)
		if loadedStats.Strength != initialPlayerStats.Strength {
			t.Errorf("Strength mismatch: expected %d, got %d", initialPlayerStats.Strength, loadedStats.Strength)
		}
		if loadedStats.Intelligence != initialPlayerStats.Intelligence {
			t.Errorf("Intelligence mismatch: expected %d, got %d", initialPlayerStats.Intelligence, loadedStats.Intelligence)
		}
	}

	// Test Skills component
	if !newGame.ecs.HasSkillsSafe(loadedPlayerID) {
		t.Error("Player should have skills component after load")
	} else {
		loadedSkills := newGame.ecs.GetSkillsSafe(loadedPlayerID)
		if loadedSkills.MeleeWeapons != initialPlayerSkills.MeleeWeapons {
			t.Errorf("Melee weapons skill mismatch: expected %d, got %d", initialPlayerSkills.MeleeWeapons, loadedSkills.MeleeWeapons)
		}
		if loadedSkills.Evocation != initialPlayerSkills.Evocation {
			t.Errorf("Evocation skill mismatch: expected %d, got %d", initialPlayerSkills.Evocation, loadedSkills.Evocation)
		}
	}

	// Test Combat component
	if !newGame.ecs.HasCombatSafe(loadedPlayerID) {
		t.Error("Player should have combat component after load")
	} else {
		loadedCombat := newGame.ecs.GetCombatSafe(loadedPlayerID)
		if loadedCombat.AttackPower != initialPlayerCombat.AttackPower {
			t.Errorf("Attack power mismatch: expected %d, got %d", initialPlayerCombat.AttackPower, loadedCombat.AttackPower)
		}
		if loadedCombat.Defense != initialPlayerCombat.Defense {
			t.Errorf("Defense mismatch: expected %d, got %d", initialPlayerCombat.Defense, loadedCombat.Defense)
		}
	}

	// Test Mana component
	if !newGame.ecs.HasManaSafe(loadedPlayerID) {
		t.Error("Player should have mana component after load")
	} else {
		loadedMana := newGame.ecs.GetManaSafe(loadedPlayerID)
		if loadedMana.MaxMP != initialPlayerMana.MaxMP {
			t.Errorf("Max mana mismatch: expected %d, got %d", initialPlayerMana.MaxMP, loadedMana.MaxMP)
		}
		if loadedMana.CurrentMP != initialPlayerMana.CurrentMP {
			t.Errorf("Current mana mismatch: expected %d, got %d", initialPlayerMana.CurrentMP, loadedMana.CurrentMP)
		}
	}

	// Test Stamina component
	if !newGame.ecs.HasStaminaSafe(loadedPlayerID) {
		t.Error("Player should have stamina component after load")
	} else {
		loadedStamina := newGame.ecs.GetStaminaSafe(loadedPlayerID)
		if loadedStamina.MaxSP != initialPlayerStamina.MaxSP {
			t.Errorf("Max stamina mismatch: expected %d, got %d", initialPlayerStamina.MaxSP, loadedStamina.MaxSP)
		}
		if loadedStamina.CurrentSP != initialPlayerStamina.CurrentSP {
			t.Errorf("Current stamina mismatch: expected %d, got %d", initialPlayerStamina.CurrentSP, loadedStamina.CurrentSP)
		}
	}

	// Test StatusEffects component
	if !newGame.ecs.HasStatusEffectsSafe(loadedPlayerID) {
		t.Error("Player should have status effects component after load")
	} else {
		loadedStatusEffects := newGame.ecs.GetStatusEffectsSafe(loadedPlayerID)
		if len(loadedStatusEffects.Effects) != len(initialPlayerStatusEffects.Effects) {
			t.Errorf("Status effects count mismatch: expected %d, got %d", len(initialPlayerStatusEffects.Effects), len(loadedStatusEffects.Effects))
		} else if len(loadedStatusEffects.Effects) > 0 {
			if loadedStatusEffects.Effects[0].Name != initialPlayerStatusEffects.Effects[0].Name {
				t.Errorf("Status effect name mismatch: expected %s, got %s", initialPlayerStatusEffects.Effects[0].Name, loadedStatusEffects.Effects[0].Name)
			}
			if loadedStatusEffects.Effects[0].StrengthMod != initialPlayerStatusEffects.Effects[0].StrengthMod {
				t.Errorf("Status effect strength mod mismatch: expected %d, got %d", initialPlayerStatusEffects.Effects[0].StrengthMod, loadedStatusEffects.Effects[0].StrengthMod)
			}
		}
	}

	// Test Inventory component
	if !newGame.ecs.HasInventorySafe(loadedPlayerID) {
		t.Error("Player should have inventory component after load")
	} else {
		loadedInventory := newGame.ecs.GetInventorySafe(loadedPlayerID)
		if len(loadedInventory.Items) != len(initialPlayerInventory.Items) {
			t.Errorf("Inventory items count mismatch: expected %d, got %d", len(initialPlayerInventory.Items), len(loadedInventory.Items))
		}
	}

	// Test Equipment component
	if !newGame.ecs.HasEquipmentSafe(loadedPlayerID) {
		t.Error("Player should have equipment component after load")
	} else {
		loadedEquipment := newGame.ecs.GetEquipmentSafe(loadedPlayerID)
		if (loadedEquipment.Weapon == nil) != (initialPlayerEquipment.Weapon == nil) {
			t.Error("Equipment weapon state mismatch")
		}
		if (loadedEquipment.Armor == nil) != (initialPlayerEquipment.Armor == nil) {
			t.Error("Equipment armor state mismatch")
		}
	}

	// Test AI component on monster
	var loadedMonsterID ecs.EntityID
	for _, id := range newGame.ecs.EntitiesAt(monsterPos) {
		if newGame.ecs.HasComponent(id, components.CAITag) {
			loadedMonsterID = id
			break
		}
	}

	if loadedMonsterID == 0 {
		t.Error("Monster should exist after load")
	} else if !newGame.ecs.HasAIComponentSafe(loadedMonsterID) {
		t.Error("Monster should have AI component after load")
	} else {
		loadedMonsterAI := newGame.ecs.GetAIComponentSafe(loadedMonsterID)
		if loadedMonsterAI.State != initialMonsterAI.State {
			t.Errorf("Monster AI state mismatch: expected %v, got %v", initialMonsterAI.State, loadedMonsterAI.State)
		}
		if loadedMonsterAI.SearchTurns != initialMonsterAI.SearchTurns {
			t.Errorf("Monster AI search turns mismatch: expected %d, got %d", initialMonsterAI.SearchTurns, loadedMonsterAI.SearchTurns)
		}
		if loadedMonsterAI.LastKnownPlayerPos != initialMonsterAI.LastKnownPlayerPos {
			t.Errorf("Monster AI last known player pos mismatch: expected %v, got %v", initialMonsterAI.LastKnownPlayerPos, loadedMonsterAI.LastKnownPlayerPos)
		}
	}

	// Test ItemPickup component
	var loadedItemID ecs.EntityID
	for _, id := range newGame.ecs.EntitiesAt(itemPos) {
		if newGame.ecs.HasComponent(id, components.CItemPickup) {
			loadedItemID = id
			break
		}
	}

	if loadedItemID == 0 {
		t.Error("Item should exist after load")
	} else if !newGame.ecs.HasItemPickupSafe(loadedItemID) {
		t.Error("Item should have pickup component after load")
	} else {
		loadedItemPickup := newGame.ecs.GetItemPickupSafe(loadedItemID)
		if loadedItemPickup.Quantity != initialItemPickup.Quantity {
			t.Errorf("Item quantity mismatch: expected %d, got %d", initialItemPickup.Quantity, loadedItemPickup.Quantity)
		}
		if loadedItemPickup.Item.Name != initialItemPickup.Item.Name {
			t.Errorf("Item name mismatch: expected %s, got %s", initialItemPickup.Item.Name, loadedItemPickup.Item.Name)
		}
	}
}
