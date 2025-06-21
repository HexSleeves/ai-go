package game

import (
	"log/slog"
	"math/rand"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ui"
)

func (g *Game) SpawnPlayer(playerStart gruid.Point, items map[string]components.Item) {
	slog.Debug("Spawning player", "position", playerStart)
	playerID := g.ecs.AddEntity()
	g.PlayerID = playerID // Store the player ID in the game struct

	g.ecs.AddComponents(playerID,
		playerStart,
		components.PlayerTag{},
		components.BlocksMovement{},
		components.Name{Name: "Player"},
		components.Renderable{Glyph: '@', Color: ui.ColorPlayer},
		components.NewHealth(10),
		components.NewTurnActor(100),
		components.NewFOVComponent(4, g.dungeon.Width, g.dungeon.Height),
		components.NewInventory(20),   // 20 slot inventory
		components.NewEquipment(),     // Empty equipment slots
		components.NewStats(),         // Starting stats
		components.NewExperience(),    // Level 1, 0 XP
		components.NewSkills(),        // Basic skills
		components.NewCombat(),        // Basic combat stats
		components.NewMana(5),         // 5 mana points
		components.NewStamina(10),     // 10 stamina points
		components.NewStatusEffects(), // No initial effects
	)

	// Add to turn queue
	g.turnQueue.Add(playerID, g.turnQueue.CurrentTime)
	// Add to spatial grid
	g.spatialGrid.Add(playerID, playerStart)

	// Give player starting items
	g.giveStartingItems(playerID, items)

	// Show welcome message
	g.showWelcomeMessage()
}

// SpawnItem creates an item pickup at the specified position
func (g *Game) SpawnItem(item components.Item, quantity int, pos gruid.Point) ecs.EntityID {
	itemID := g.ecs.AddEntity()
	pickup := components.NewItemPickup(item, quantity)

	g.ecs.AddComponents(itemID,
		pos,
		pickup,
		components.Renderable{Glyph: item.Glyph, Color: item.Color},
		components.Name{Name: item.Name},
	)

	// Add to spatial grid
	g.spatialGrid.Add(itemID, pos)

	slog.Debug("Spawned item", "name", item.Name, "quantity", quantity, "position", pos)
	return itemID
}

// giveStartingItems gives the player some starting equipment and items
func (g *Game) giveStartingItems(playerID ecs.EntityID, items map[string]components.Item) {
	if !g.ecs.HasInventorySafe(playerID) {
		return
	}

	inventory := g.ecs.GetInventorySafe(playerID)

	// Give starting items
	startingItems := []struct {
		name     string
		quantity int
	}{
		{"Health Potion", 3},
		{"Iron Sword", 1},
		{"Leather Armor", 1},
		{"Gold Coin", 50},
	}

	for _, startItem := range startingItems {
		if item, exists := items[startItem.name]; exists {
			if inventory.AddItem(item, startItem.quantity) {
				slog.Debug("Gave player item", "name", startItem.name, "quantity", startItem.quantity)
			}
		}
	}

	// Update inventory
	g.ecs.AddComponent(playerID, components.CInventory, inventory)

	// Auto-equip starting weapon and armor
	if g.ecs.HasEquipmentSafe(playerID) {
		equipment := g.ecs.GetEquipmentSafe(playerID)

		// Equip sword
		if sword, exists := items["Iron Sword"]; exists {
			equipment.EquipItem(sword)
			inventory.RemoveItem("Iron Sword", 1)
		}

		// Equip armor
		if armor, exists := items["Leather Armor"]; exists {
			equipment.EquipItem(armor)
			inventory.RemoveItem("Leather Armor", 1)
		}

		// Update components
		g.ecs.AddComponent(playerID, components.CEquipment, equipment)
		g.ecs.AddComponent(playerID, components.CInventory, inventory)

		slog.Debug("Player equipped starting gear")
	}
}

// CreateBasicItems returns common game items
func CreateBasicItems() map[string]components.Item {
	return map[string]components.Item{
		"Health Potion": {
			Name:        "Health Potion",
			Description: "Restores 10 HP",
			Type:        components.ItemTypeConsumable,
			Glyph:       '!',
			Color:       gruid.Color(0xFF0000), // Red
			Value:       50,
			Stackable:   true,
			MaxStack:    10,
		},
		"Iron Sword": {
			Name:        "Iron Sword",
			Description: "A sturdy iron sword",
			Type:        components.ItemTypeWeapon,
			Glyph:       '/',
			Color:       gruid.Color(0xC0C0C0), // Silver
			Value:       100,
			Stackable:   false,
		},
		"Leather Armor": {
			Name:        "Leather Armor",
			Description: "Basic leather protection",
			Type:        components.ItemTypeArmor,
			Glyph:       '[',
			Color:       gruid.Color(0x8B4513), // Brown
			Value:       75,
			Stackable:   false,
		},
		"Gold Coin": {
			Name:        "Gold Coin",
			Description: "Shiny gold currency",
			Type:        components.ItemTypeMisc,
			Glyph:       '$',
			Color:       gruid.Color(0xFFD700), // Gold
			Value:       1,
			Stackable:   true,
			MaxStack:    100,
		},
	}
}

// showWelcomeMessage displays the welcome message and basic instructions
func (g *Game) showWelcomeMessage() {
	g.log.AddMessagef(ui.ColorStatusGood, "Welcome to the Roguelike!")
	g.log.AddMessagef(ui.ColorStatusGood, "You are equipped with basic gear.")
	g.log.AddMessagef(ui.ColorStatusGood, "Press ? for help, C for character sheet.")
	g.log.AddMessagef(ui.ColorStatusGood, "Good luck, adventurer!")
}

func (g *Game) SpawnMonster(pos gruid.Point) {
	monsterID := g.ecs.AddEntity()

	monsterNames := []string{"Orc", "Troll", "Goblin", "Kobold"}
	monsterName := monsterNames[rand.Intn(len(monsterNames))]

	var rune rune
	var speed uint64
	var color gruid.Color = ui.ColorMonster // Default monster color
	var maxHP int

	switch monsterName {
	case "Orc":
		rune = 'o'
		speed = 100
		maxHP = 1
	case "Troll":
		rune = 'T'
		speed = 200
		maxHP = 1
	case "Goblin":
		rune = 'g'
		speed = 100
		color = ui.ColorSleepingMonster // Goblins use a different color
		maxHP = 1
	case "Kobold":
		rune = 'k'
		speed = 150
		maxHP = 1
	}

	// Create AI component with random behavior
	behaviors := []components.AIBehavior{
		components.AIBehaviorWander,
		components.AIBehaviorGuard,
		components.AIBehaviorHunter,
	}
	behavior := behaviors[rand.Intn(len(behaviors))]
	aiComponent := components.NewAIComponent(behavior, pos)

	g.ecs.AddComponents(monsterID,
		pos,
		components.AITag{},
		aiComponent,
		components.BlocksMovement{},
		components.Name{Name: monsterName},
		components.Renderable{Glyph: rune, Color: color},
		components.NewHealth(maxHP),
		components.NewFOVComponent(6, g.dungeon.Width, g.dungeon.Height),
		components.NewTurnActor(speed),
	)

	slog.Debug("Created monster", "id", monsterID, "position", pos, "time", g.turnQueue.CurrentTime+100)

	// Add to turn queue
	g.turnQueue.Add(monsterID, g.turnQueue.CurrentTime+100)
	// Add to spatial grid
	g.spatialGrid.Add(monsterID, pos)
}
