package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/rl"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/log"
	turn "github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/turn_queue"
	"github.com/sirupsen/logrus"
)

// SaveData represents the complete game state for serialization
type SaveData struct {
	Version   string         `json:"version"`
	Timestamp time.Time      `json:"timestamp"`
	PlayerID  ecs.EntityID   `json:"player_id"`
	Depth     int            `json:"depth"`
	Entities  []SavedEntity  `json:"entities"`
	Map       SavedMap       `json:"map"`
	TurnQueue SavedTurnQueue `json:"turn_queue"`
	Messages  []SavedMessage `json:"messages"`
	GameStats SavedGameStats `json:"game_stats"`
}

// SavedEntity represents an entity and its components
type SavedEntity struct {
	ID         ecs.EntityID           `json:"id"`
	Components map[string]interface{} `json:"components"`
}

// SavedMap represents the map state
type SavedMap struct {
	Width    int      `json:"width"`
	Height   int      `json:"height"`
	Cells    [][]int  `json:"cells"`    // Grid data
	Explored []uint64 `json:"explored"` // Explored bitset
}

// SavedTurnQueue represents the turn queue state
type SavedTurnQueue struct {
	CurrentTime uint64                `json:"current_time"`
	Entries     []SavedTurnQueueEntry `json:"entries"`
}

// SavedTurnQueueEntry represents a turn queue entry
type SavedTurnQueueEntry struct {
	EntityID ecs.EntityID `json:"entity_id"`
	Time     uint64       `json:"time"`
}

// SavedMessage represents a log message
type SavedMessage struct {
	Text      string    `json:"text"`
	Color     uint32    `json:"color"` // Color as uint32
	Timestamp time.Time `json:"timestamp"`
}

// SavedGameStats represents game statistics
type SavedGameStats struct {
	PlayTime       time.Duration `json:"play_time"`
	MonstersKilled int           `json:"monsters_killed"`
	ItemsCollected int           `json:"items_collected"`
	DamageDealt    int           `json:"damage_dealt"`
	DamageTaken    int           `json:"damage_taken"`
}

const (
	SaveVersion = "1.0.0"
	SaveDir     = "saves"
	SaveFile    = "game.save"
)

// SaveGame saves the current game state to disk
func (g *Game) SaveGame() error {
	// Create saves directory if it doesn't exist
	if err := os.MkdirAll(SaveDir, 0755); err != nil {
		return fmt.Errorf("failed to create save directory: %w", err)
	}

	saveData := SaveData{
		Version:   SaveVersion,
		Timestamp: time.Now(),
		PlayerID:  g.PlayerID,
		Depth:     g.Depth,
	}

	// Save entities and their components
	entities := g.ecs.GetAllEntities()
	for _, entityID := range entities {
		savedEntity := SavedEntity{
			ID:         entityID,
			Components: make(map[string]interface{}),
		}

		// Save each component type
		if pos, ok := g.ecs.GetPosition(entityID); ok {
			savedEntity.Components["position"] = pos
		}
		if renderable, ok := g.ecs.GetRenderable(entityID); ok {
			savedEntity.Components["renderable"] = renderable
		}
		if health, ok := g.ecs.GetHealth(entityID); ok {
			savedEntity.Components["health"] = health
		}
		if name, ok := g.ecs.GetName(entityID); ok {
			savedEntity.Components["name"] = name
		}
		if g.ecs.HasComponent(entityID, components.CPlayerTag) {
			savedEntity.Components["player_tag"] = true
		}
		if g.ecs.HasComponent(entityID, components.CAITag) {
			savedEntity.Components["ai_tag"] = true
		}
		if g.ecs.HasComponent(entityID, components.CBlocksMovement) {
			savedEntity.Components["blocks_movement"] = true
		}
		if g.ecs.HasComponent(entityID, components.CCorpseTag) {
			savedEntity.Components["corpse_tag"] = true
		}
		if turnActor, ok := g.ecs.GetTurnActor(entityID); ok {
			savedEntity.Components["turn_actor"] = turnActor
		}

		// Save new components
		if inventory, ok := g.ecs.GetInventory(entityID); ok {
			savedEntity.Components["inventory"] = inventory
		}
		if equipment, ok := g.ecs.GetEquipment(entityID); ok {
			savedEntity.Components["equipment"] = equipment
		}
		if itemPickup, ok := g.ecs.GetItemPickup(entityID); ok {
			savedEntity.Components["item_pickup"] = itemPickup
		}
		if aiComponent, ok := g.ecs.GetAIComponent(entityID); ok {
			savedEntity.Components["ai_component"] = aiComponent
		}
		if stats, ok := g.ecs.GetStats(entityID); ok {
			savedEntity.Components["stats"] = stats
		}
		if experience, ok := g.ecs.GetExperience(entityID); ok {
			savedEntity.Components["experience"] = experience
		}
		if skills, ok := g.ecs.GetSkills(entityID); ok {
			savedEntity.Components["skills"] = skills
		}
		if combat, ok := g.ecs.GetCombat(entityID); ok {
			savedEntity.Components["combat"] = combat
		}
		if mana, ok := g.ecs.GetMana(entityID); ok {
			savedEntity.Components["mana"] = mana
		}
		if stamina, ok := g.ecs.GetStamina(entityID); ok {
			savedEntity.Components["stamina"] = stamina
		}
		if statusEffects, ok := g.ecs.GetStatusEffects(entityID); ok {
			savedEntity.Components["status_effects"] = statusEffects
		}

		saveData.Entities = append(saveData.Entities, savedEntity)
	}

	// Save map state
	saveData.Map = SavedMap{
		Width:    g.dungeon.Width,
		Height:   g.dungeon.Height,
		Explored: g.dungeon.Explored,
	}

	// Convert grid to serializable format
	saveData.Map.Cells = make([][]int, g.dungeon.Height)
	for y := 0; y < g.dungeon.Height; y++ {
		saveData.Map.Cells[y] = make([]int, g.dungeon.Width)
		for x := 0; x < g.dungeon.Width; x++ {
			point := gruid.Point{X: x, Y: y}
			cell := g.dungeon.Grid.At(point)
			saveData.Map.Cells[y][x] = int(cell)
		}
	}

	// Save turn queue state with all entries
	queueSnapshot := g.turnQueue.Snapshot()
	savedEntries := make([]SavedTurnQueueEntry, len(queueSnapshot))
	for i, entry := range queueSnapshot {
		savedEntries[i] = SavedTurnQueueEntry{
			EntityID: entry.EntityID,
			Time:     entry.Time,
		}
	}

	saveData.TurnQueue = SavedTurnQueue{
		CurrentTime: g.turnQueue.CurrentTime,
		Entries:     savedEntries,
	}

	// Save messages with timestamps
	for _, msg := range g.log.Messages {
		savedMsg := SavedMessage{
			Text:      msg.Text,
			Color:     uint32(msg.Color),
			Timestamp: msg.Timestamp,
		}
		saveData.Messages = append(saveData.Messages, savedMsg)
	}

	// Save game statistics
	if g.stats != nil {
		// Update play time before saving
		g.UpdatePlayTime()
		saveData.GameStats = SavedGameStats{
			PlayTime:       g.stats.PlayTime,
			MonstersKilled: g.stats.MonstersKilled,
			ItemsCollected: g.stats.ItemsCollected,
			DamageDealt:    g.stats.DamageDealt,
			DamageTaken:    g.stats.DamageTaken,
		}
	}

	// Serialize to JSON
	data, err := json.MarshalIndent(saveData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal save data: %w", err)
	}

	// Write to file
	savePath := filepath.Join(SaveDir, SaveFile)
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write save file: %w", err)
	}

	logrus.Infof("Game saved to %s", savePath)
	return nil
}

// LoadGame loads a saved game state from disk
func (g *Game) LoadGame() error {
	savePath := filepath.Join(SaveDir, SaveFile)

	// Check if save file exists
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		return fmt.Errorf("no save file found at %s", savePath)
	}

	// Read save file
	data, err := os.ReadFile(savePath)
	if err != nil {
		return fmt.Errorf("failed to read save file: %w", err)
	}

	// Parse JSON
	var saveData SaveData
	if err := json.Unmarshal(data, &saveData); err != nil {
		return fmt.Errorf("failed to unmarshal save data: %w", err)
	}

	// Version check
	if saveData.Version != SaveVersion {
		logrus.Warnf("Save file version %s differs from current version %s",
			saveData.Version, SaveVersion)
	}

	// Clear current game state
	g.ecs = ecs.NewECS()
	g.spatialGrid.Clear()

	// Restore basic game state
	g.PlayerID = saveData.PlayerID
	g.Depth = saveData.Depth

	// Restore map
	g.dungeon = NewMap(saveData.Map.Width, saveData.Map.Height)
	g.dungeon.Explored = saveData.Map.Explored

	// Restore grid cells
	for y := 0; y < saveData.Map.Height; y++ {
		for x := 0; x < saveData.Map.Width; x++ {
			if y < len(saveData.Map.Cells) && x < len(saveData.Map.Cells[y]) {
				point := gruid.Point{X: x, Y: y}
				cell := rl.Cell(saveData.Map.Cells[y][x])
				g.dungeon.Grid.Set(point, cell)
			}
		}
	}

	// Restore entities
	for _, savedEntity := range saveData.Entities {
		// Create entity with specific ID
		if err := g.ecs.AddEntityWithID(savedEntity.ID); err != nil {
			logrus.Errorf("Failed to create entity with ID %d: %v", savedEntity.ID, err)
			continue
		}
		entityID := savedEntity.ID

		// Restore components
		for compType, compData := range savedEntity.Components {
			switch compType {
			case "position":
				if pos, ok := compData.(map[string]interface{}); ok {
					point := gruid.Point{
						X: int(pos["X"].(float64)),
						Y: int(pos["Y"].(float64)),
					}
					g.ecs.AddComponent(entityID, components.CPosition, point)
					g.spatialGrid.Add(entityID, point)
				}
			case "health":
				if healthData, ok := compData.(map[string]interface{}); ok {
					health := components.Health{
						CurrentHP: int(healthData["CurrentHP"].(float64)),
						MaxHP:     int(healthData["MaxHP"].(float64)),
					}
					g.ecs.AddComponent(entityID, components.CHealth, health)
				}
			case "player_tag":
				g.ecs.AddComponent(entityID, components.CPlayerTag, components.PlayerTag{})
			case "ai_tag":
				g.ecs.AddComponent(entityID, components.CAITag, components.AITag{})
			case "blocks_movement":
				g.ecs.AddComponent(entityID, components.CBlocksMovement, components.BlocksMovement{})
			case "corpse_tag":
				g.ecs.AddComponent(entityID, components.CCorpseTag, components.CorpseTag{})

			case "renderable":
				if renderData, ok := compData.(map[string]interface{}); ok {
					renderable := components.Renderable{
						Glyph: rune(renderData["Glyph"].(float64)),
						Color: gruid.Color(renderData["Color"].(float64)),
					}
					g.ecs.AddComponent(entityID, components.CRenderable, renderable)
				}

			case "name":
				if nameStr, ok := compData.(string); ok {
					g.ecs.AddComponent(entityID, components.CName, nameStr)
				}

			case "turn_actor":
				if actorData, ok := compData.(map[string]interface{}); ok {
					actor := components.TurnActor{
						Speed:        uint64(actorData["Speed"].(float64)),
						Alive:        actorData["Alive"].(bool),
						NextTurnTime: uint64(actorData["NextTurnTime"].(float64)),
					}
					g.ecs.AddComponent(entityID, components.CTurnActor, actor)
				}

			// New components
			case "inventory":
				if invData, ok := compData.(map[string]interface{}); ok {
					inventory := components.Inventory{
						Capacity: int(invData["Capacity"].(float64)),
						Items:    []components.ItemStack{},
					}
					if itemsData, ok := invData["Items"].([]interface{}); ok {
						for _, itemData := range itemsData {
							if itemMap, ok := itemData.(map[string]interface{}); ok {
								// Reconstruct ItemStack
								stack := components.ItemStack{
									Quantity: int(itemMap["Quantity"].(float64)),
								}
								// Reconstruct Item
								if itemInfo, ok := itemMap["Item"].(map[string]interface{}); ok {
									stack.Item = components.Item{
										Name:        itemInfo["Name"].(string),
										Description: itemInfo["Description"].(string),
										Type:        components.ItemType(itemInfo["Type"].(float64)),
										Glyph:       rune(itemInfo["Glyph"].(float64)),
										Color:       gruid.Color(itemInfo["Color"].(float64)),
										Value:       int(itemInfo["Value"].(float64)),
										Stackable:   itemInfo["Stackable"].(bool),
										MaxStack:    int(itemInfo["MaxStack"].(float64)),
									}
								}
								inventory.Items = append(inventory.Items, stack)
							}
						}
					}
					g.ecs.AddComponent(entityID, components.CInventory, inventory)
				}

			case "equipment":
				if eqData, ok := compData.(map[string]interface{}); ok {
					equipment := components.Equipment{}
					// Restore weapon
					if weaponData, ok := eqData["Weapon"]; ok && weaponData != nil {
						if weaponMap, ok := weaponData.(map[string]interface{}); ok {
							weapon := components.Item{
								Name:        weaponMap["Name"].(string),
								Description: weaponMap["Description"].(string),
								Type:        components.ItemType(weaponMap["Type"].(float64)),
								Glyph:       rune(weaponMap["Glyph"].(float64)),
								Color:       gruid.Color(weaponMap["Color"].(float64)),
								Value:       int(weaponMap["Value"].(float64)),
								Stackable:   weaponMap["Stackable"].(bool),
								MaxStack:    int(weaponMap["MaxStack"].(float64)),
							}
							equipment.Weapon = &weapon
						}
					}
					// Restore armor
					if armorData, ok := eqData["Armor"]; ok && armorData != nil {
						if armorMap, ok := armorData.(map[string]interface{}); ok {
							armor := components.Item{
								Name:        armorMap["Name"].(string),
								Description: armorMap["Description"].(string),
								Type:        components.ItemType(armorMap["Type"].(float64)),
								Glyph:       rune(armorMap["Glyph"].(float64)),
								Color:       gruid.Color(armorMap["Color"].(float64)),
								Value:       int(armorMap["Value"].(float64)),
								Stackable:   armorMap["Stackable"].(bool),
								MaxStack:    int(armorMap["MaxStack"].(float64)),
							}
							equipment.Armor = &armor
						}
					}
					g.ecs.AddComponent(entityID, components.CEquipment, equipment)
				}

			case "experience":
				if expData, ok := compData.(map[string]interface{}); ok {
					experience := components.Experience{
						Level:           int(expData["Level"].(float64)),
						CurrentXP:       int(expData["CurrentXP"].(float64)),
						XPToNextLevel:   int(expData["XPToNextLevel"].(float64)),
						TotalXP:         int(expData["TotalXP"].(float64)),
						SkillPoints:     int(expData["SkillPoints"].(float64)),
						AttributePoints: int(expData["AttributePoints"].(float64)),
					}
					g.ecs.AddComponent(entityID, components.CExperience, experience)
				}

			case "item_pickup":
				if pickupData, ok := compData.(map[string]interface{}); ok {
					pickup := components.ItemPickup{
						Quantity: int(pickupData["Quantity"].(float64)),
					}
					// Restore the Item
					if itemData, ok := pickupData["Item"].(map[string]interface{}); ok {
						pickup.Item = components.Item{
							Name:        itemData["Name"].(string),
							Description: itemData["Description"].(string),
							Type:        components.ItemType(itemData["Type"].(float64)),
							Glyph:       rune(itemData["Glyph"].(float64)),
							Color:       gruid.Color(itemData["Color"].(float64)),
							Value:       int(itemData["Value"].(float64)),
							Stackable:   itemData["Stackable"].(bool),
							MaxStack:    int(itemData["MaxStack"].(float64)),
						}
					}
					g.ecs.AddComponent(entityID, components.CItemPickup, pickup)
				}

			case "ai_component":
				if aiData, ok := compData.(map[string]interface{}); ok {
					aiComponent := components.AIComponent{
						Behavior:       components.AIBehavior(aiData["Behavior"].(float64)),
						State:          components.AIState(aiData["State"].(float64)),
						PatrolRadius:   int(aiData["PatrolRadius"].(float64)),
						AggroRange:     int(aiData["AggroRange"].(float64)),
						FleeThreshold:  aiData["FleeThreshold"].(float64),
						SearchTurns:    int(aiData["SearchTurns"].(float64)),
						MaxSearchTurns: int(aiData["MaxSearchTurns"].(float64)),
					}
					// Restore LastKnownPlayerPos
					if posData, ok := aiData["LastKnownPlayerPos"].(map[string]interface{}); ok {
						aiComponent.LastKnownPlayerPos = gruid.Point{
							X: int(posData["X"].(float64)),
							Y: int(posData["Y"].(float64)),
						}
					}
					// Restore HomePosition
					if homeData, ok := aiData["HomePosition"].(map[string]interface{}); ok {
						aiComponent.HomePosition = gruid.Point{
							X: int(homeData["X"].(float64)),
							Y: int(homeData["Y"].(float64)),
						}
					}
					g.ecs.AddComponent(entityID, components.CAIComponent, aiComponent)
				}

			case "stats":
				if statsData, ok := compData.(map[string]interface{}); ok {
					stats := components.Stats{
						Strength:     int(statsData["Strength"].(float64)),
						Dexterity:    int(statsData["Dexterity"].(float64)),
						Constitution: int(statsData["Constitution"].(float64)),
						Intelligence: int(statsData["Intelligence"].(float64)),
						Wisdom:       int(statsData["Wisdom"].(float64)),
						Charisma:     int(statsData["Charisma"].(float64)),
					}
					g.ecs.AddComponent(entityID, components.CStats, stats)
				}

			case "skills":
				if skillsData, ok := compData.(map[string]interface{}); ok {
					skills := components.Skills{
						MeleeWeapons:  int(skillsData["MeleeWeapons"].(float64)),
						RangedWeapons: int(skillsData["RangedWeapons"].(float64)),
						Defense:       int(skillsData["Defense"].(float64)),
						Stealth:       int(skillsData["Stealth"].(float64)),
						Perception:    int(skillsData["Perception"].(float64)),
						Medicine:      int(skillsData["Medicine"].(float64)),
						Crafting:      int(skillsData["Crafting"].(float64)),
						Evocation:     int(skillsData["Evocation"].(float64)),
						Conjuration:   int(skillsData["Conjuration"].(float64)),
						Enchantment:   int(skillsData["Enchantment"].(float64)),
						Divination:    int(skillsData["Divination"].(float64)),
						Lockpicking:   int(skillsData["Lockpicking"].(float64)),
					}
					g.ecs.AddComponent(entityID, components.CSkills, skills)
				}

			case "combat":
				if combatData, ok := compData.(map[string]interface{}); ok {
					combat := components.Combat{
						AttackPower:    int(combatData["AttackPower"].(float64)),
						Defense:        int(combatData["Defense"].(float64)),
						Accuracy:       int(combatData["Accuracy"].(float64)),
						DodgeChance:    int(combatData["DodgeChance"].(float64)),
						CriticalChance: int(combatData["CriticalChance"].(float64)),
						CriticalDamage: int(combatData["CriticalDamage"].(float64)),
					}
					g.ecs.AddComponent(entityID, components.CCombat, combat)
				}

			case "mana":
				if manaData, ok := compData.(map[string]interface{}); ok {
					mana := components.Mana{
						CurrentMP: int(manaData["CurrentMP"].(float64)),
						MaxMP:     int(manaData["MaxMP"].(float64)),
						RegenRate: int(manaData["RegenRate"].(float64)),
					}
					g.ecs.AddComponent(entityID, components.CMana, mana)
				}

			case "stamina":
				if staminaData, ok := compData.(map[string]interface{}); ok {
					stamina := components.Stamina{
						CurrentSP: int(staminaData["CurrentSP"].(float64)),
						MaxSP:     int(staminaData["MaxSP"].(float64)),
						RegenRate: int(staminaData["RegenRate"].(float64)),
					}
					g.ecs.AddComponent(entityID, components.CStamina, stamina)
				}

			case "status_effects":
				if effectsData, ok := compData.(map[string]interface{}); ok {
					statusEffects := components.StatusEffects{
						Effects: []components.StatusEffect{},
					}
					// Restore effects array
					if effectsArray, ok := effectsData["Effects"].([]interface{}); ok {
						for _, effectData := range effectsArray {
							if effectMap, ok := effectData.(map[string]interface{}); ok {
								effect := components.StatusEffect{
									Name:            effectMap["Name"].(string),
									Description:     effectMap["Description"].(string),
									Duration:        int(effectMap["Duration"].(float64)),
									StrengthMod:     int(effectMap["StrengthMod"].(float64)),
									DexterityMod:    int(effectMap["DexterityMod"].(float64)),
									ConstitutionMod: int(effectMap["ConstitutionMod"].(float64)),
									IntelligenceMod: int(effectMap["IntelligenceMod"].(float64)),
									WisdomMod:       int(effectMap["WisdomMod"].(float64)),
									CharismaMod:     int(effectMap["CharismaMod"].(float64)),
									AttackMod:       int(effectMap["AttackMod"].(float64)),
									DefenseMod:      int(effectMap["DefenseMod"].(float64)),
									AccuracyMod:     int(effectMap["AccuracyMod"].(float64)),
									DodgeMod:        int(effectMap["DodgeMod"].(float64)),
								}
								statusEffects.Effects = append(statusEffects.Effects, effect)
							}
						}
					}
					g.ecs.AddComponent(entityID, components.CStatusEffects, statusEffects)
				}
			}
		}
	}

	// Restore turn queue with all entries
	g.turnQueue.CurrentTime = saveData.TurnQueue.CurrentTime

	// Convert saved entries back to TurnEntry format
	queueEntries := make([]turn.TurnEntry, len(saveData.TurnQueue.Entries))
	for i, savedEntry := range saveData.TurnQueue.Entries {
		queueEntries[i] = turn.TurnEntry{
			EntityID: savedEntry.EntityID,
			Time:     savedEntry.Time,
		}
	}

	// Restore the queue from snapshot
	g.turnQueue.RestoreFromSnapshot(queueEntries)

	// Restore messages with timestamps
	g.log.Messages = []log.Message{}
	for _, savedMsg := range saveData.Messages {
		g.log.AddMessageWithTimestamp(savedMsg.Text, gruid.Color(savedMsg.Color), savedMsg.Timestamp)
	}

	// Restore game statistics
	if g.stats == nil {
		g.stats = &GameStats{}
	}
	g.stats.PlayTime = saveData.GameStats.PlayTime
	g.stats.MonstersKilled = saveData.GameStats.MonstersKilled
	g.stats.ItemsCollected = saveData.GameStats.ItemsCollected
	g.stats.DamageDealt = saveData.GameStats.DamageDealt
	g.stats.DamageTaken = saveData.GameStats.DamageTaken
	// Adjust start time to account for loaded play time
	g.stats.StartTime = time.Now().Add(-g.stats.PlayTime)

	logrus.Infof("Game loaded from %s", savePath)
	return nil
}

// HasSaveFile checks if a save file exists
func HasSaveFile() bool {
	savePath := filepath.Join(SaveDir, SaveFile)
	_, err := os.Stat(savePath)
	return err == nil
}

// DeleteSaveFile removes the save file
func DeleteSaveFile() error {
	savePath := filepath.Join(SaveDir, SaveFile)
	if err := os.Remove(savePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete save file: %w", err)
	}
	return nil
}
