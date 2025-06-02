package game

import (
	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ui"
	"github.com/sirupsen/logrus"
)

type playerAction int

const (
	ActionNone playerAction = iota
	ActionW
	ActionS
	ActionN
	ActionE
	ActionQuit
	ActionPickup
	ActionDrop
	ActionInventory
	ActionUseItem
	ActionEquip
	ActionWait
	ActionSave
	ActionLoad
	ActionCharacterSheet
	ActionHelp
)

type actionError int

const (
	actionErrorUnknown actionError = iota
)

func (e actionError) Error() string {
	switch e {
	case actionErrorUnknown:
		return "unknown action"
	}
	return ""
}

func (md *Model) normalModeAction(playerAction playerAction) (again bool, eff gruid.Effect, err error) {
	g := md.game

	logrus.Debugf("Normal mode action: %v\n", playerAction)

	switch playerAction {
	case ActionNone:
		again = true
		err = actionErrorUnknown
	case ActionW, ActionS, ActionN, ActionE:
		action := MoveAction{
			Direction: keyToDir(playerAction),
			EntityID:  g.PlayerID,
		}
		actor, _ := g.ecs.GetTurnActor(g.PlayerID)
		actor.AddAction(action)
		return false, eff, nil

	case ActionWait:
		action := WaitAction{EntityID: g.PlayerID}
		actor, _ := g.ecs.GetTurnActor(g.PlayerID)
		actor.AddAction(action)
		return false, eff, nil

	case ActionPickup:
		return md.handlePickupAction()

	case ActionDrop:
		return md.handleDropAction()

	case ActionInventory:
		return md.handleInventoryAction()

	case ActionUseItem:
		return md.handleUseItemAction()

	case ActionEquip:
		return md.handleEquipAction()

	case ActionSave:
		return md.handleSaveAction()

	case ActionLoad:
		return md.handleLoadAction()

	case ActionCharacterSheet:
		return md.handleCharacterSheetAction()

	case ActionHelp:
		return md.handleHelpAction()

	default:
		logrus.Debugf("Unknown action: %v\n", playerAction)
		err = actionErrorUnknown
	}

	if err != nil {
		again = true
	}

	return again, eff, err
}

// handlePickupAction handles picking up items at the player's position
func (md *Model) handlePickupAction() (again bool, eff gruid.Effect, err error) {
	g := md.game
	playerPos := g.GetPlayerPosition()

	// Find items at player position
	entities := g.ecs.EntitiesAt(playerPos)
	var itemEntities []ecs.EntityID

	for _, entityID := range entities {
		if g.ecs.HasItemPickupSafe(entityID) {
			itemEntities = append(itemEntities, entityID)
		}
	}

	if len(itemEntities) == 0 {
		g.log.AddMessagef(ui.ColorStatusBad, "There's nothing here to pick up.")
		return true, eff, nil // Don't consume turn
	}

	// For now, pick up the first item found
	// TODO: Add item selection UI for multiple items
	itemID := itemEntities[0]
	action := PickupAction{EntityID: g.PlayerID, ItemID: itemID}

	actor, _ := g.ecs.GetTurnActor(g.PlayerID)
	actor.AddAction(action)

	return false, eff, nil
}

// handleDropAction handles dropping items (simplified - drops first item)
func (md *Model) handleDropAction() (again bool, eff gruid.Effect, err error) {
	g := md.game
	inventory := g.ecs.GetInventorySafe(g.PlayerID)

	if len(inventory.Items) == 0 {
		g.log.AddMessagef(ui.ColorStatusBad, "Your inventory is empty.")
		return true, eff, nil // Don't consume turn
	}

	// For now, drop the first item
	// TODO: Add item selection UI
	firstItem := inventory.Items[0]
	action := DropAction{
		EntityID: g.PlayerID,
		ItemName: firstItem.Item.Name,
		Quantity: 1,
	}

	actor, _ := g.ecs.GetTurnActor(g.PlayerID)
	actor.AddAction(action)

	return false, eff, nil
}

// handleInventoryAction shows inventory (for now just logs)
func (md *Model) handleInventoryAction() (again bool, eff gruid.Effect, err error) {
	g := md.game
	inventory := g.ecs.GetInventorySafe(g.PlayerID)

	if len(inventory.Items) == 0 {
		g.log.AddMessagef(ui.ColorStatusGood, "Your inventory is empty.")
	} else {
		g.log.AddMessagef(ui.ColorStatusGood, "Inventory (%d/%d):", len(inventory.Items), inventory.Capacity)
		for _, stack := range inventory.Items {
			if stack.Quantity > 1 {
				g.log.AddMessagef(ui.ColorStatusGood, "  %s x%d", stack.Item.Name, stack.Quantity)
			} else {
				g.log.AddMessagef(ui.ColorStatusGood, "  %s", stack.Item.Name)
			}
		}
	}

	return true, eff, nil // Don't consume turn
}

// handleUseItemAction uses the first consumable item
func (md *Model) handleUseItemAction() (again bool, eff gruid.Effect, err error) {
	g := md.game
	inventory := g.ecs.GetInventorySafe(g.PlayerID)

	// Find first consumable item
	for _, stack := range inventory.Items {
		if stack.Item.Type == components.ItemTypeConsumable {
			action := UseItemAction{
				EntityID: g.PlayerID,
				ItemName: stack.Item.Name,
			}

			actor, _ := g.ecs.GetTurnActor(g.PlayerID)
			actor.AddAction(action)

			return false, eff, nil
		}
	}

	g.log.AddMessagef(ui.ColorStatusBad, "You have no consumable items.")
	return true, eff, nil // Don't consume turn
}

// handleEquipAction equips the first equippable item
func (md *Model) handleEquipAction() (again bool, eff gruid.Effect, err error) {
	g := md.game
	inventory := g.ecs.GetInventorySafe(g.PlayerID)

	// Find first equippable item
	for _, stack := range inventory.Items {
		if stack.Item.Type == components.ItemTypeWeapon || stack.Item.Type == components.ItemTypeArmor {
			action := EquipAction{
				EntityID: g.PlayerID,
				ItemName: stack.Item.Name,
			}

			actor, _ := g.ecs.GetTurnActor(g.PlayerID)
			actor.AddAction(action)

			return false, eff, nil
		}
	}

	g.log.AddMessagef(ui.ColorStatusBad, "You have no equippable items.")
	return true, eff, nil // Don't consume turn
}

// handleSaveAction saves the game
func (md *Model) handleSaveAction() (again bool, eff gruid.Effect, err error) {
	g := md.game

	if err := g.SaveGame(); err != nil {
		g.log.AddMessagef(ui.ColorStatusBad, "Failed to save game: %v", err)
		logrus.Errorf("Save failed: %v", err)
	} else {
		g.log.AddMessagef(ui.ColorStatusGood, "Game saved successfully!")
	}

	return true, eff, nil // Don't consume turn
}

// handleLoadAction loads the game
func (md *Model) handleLoadAction() (again bool, eff gruid.Effect, err error) {
	g := md.game

	if !HasSaveFile() {
		g.log.AddMessagef(ui.ColorStatusBad, "No save file found!")
		return true, eff, nil // Don't consume turn
	}

	if err := g.LoadGame(); err != nil {
		g.log.AddMessagef(ui.ColorStatusBad, "Failed to load game: %v", err)
		logrus.Errorf("Load failed: %v", err)
	} else {
		g.log.AddMessagef(ui.ColorStatusGood, "Game loaded successfully!")
	}

	return true, eff, nil // Don't consume turn
}

// handleCharacterSheetAction displays character information
func (md *Model) handleCharacterSheetAction() (again bool, eff gruid.Effect, err error) {
	g := md.game
	playerID := g.PlayerID

	g.log.AddMessagef(ui.ColorStatusGood, "=== Character Sheet ===")

	// Basic info
	if g.ecs.HasExperienceSafe(playerID) {
		exp := g.ecs.GetExperienceSafe(playerID)
		g.log.AddMessagef(ui.ColorStatusGood, "Level: %d | XP: %d/%d | Total XP: %d",
			exp.Level, exp.CurrentXP, exp.XPToNextLevel, exp.TotalXP)
		g.log.AddMessagef(ui.ColorStatusGood, "Skill Points: %d | Attribute Points: %d",
			exp.SkillPoints, exp.AttributePoints)
	}

	// Health, Mana, Stamina
	if g.ecs.HasHealthSafe(playerID) {
		health := g.ecs.GetHealthSafe(playerID)
		g.log.AddMessagef(ui.ColorStatusGood, "Health: %d/%d", health.CurrentHP, health.MaxHP)
	}
	if g.ecs.HasManaSafe(playerID) {
		mana := g.ecs.GetManaSafe(playerID)
		g.log.AddMessagef(ui.ColorStatusGood, "Mana: %d/%d", mana.CurrentMP, mana.MaxMP)
	}
	if g.ecs.HasStaminaSafe(playerID) {
		stamina := g.ecs.GetStaminaSafe(playerID)
		g.log.AddMessagef(ui.ColorStatusGood, "Stamina: %d/%d", stamina.CurrentSP, stamina.MaxSP)
	}

	// Stats
	if g.ecs.HasStatsSafe(playerID) {
		stats := g.ecs.GetStatsSafe(playerID)
		g.log.AddMessagef(ui.ColorStatusGood, "STR: %d | DEX: %d | CON: %d",
			stats.Strength, stats.Dexterity, stats.Constitution)
		g.log.AddMessagef(ui.ColorStatusGood, "INT: %d | WIS: %d | CHA: %d",
			stats.Intelligence, stats.Wisdom, stats.Charisma)
	}

	// Combat stats
	if g.ecs.HasCombatSafe(playerID) {
		combat := g.ecs.GetCombatSafe(playerID)
		g.log.AddMessagef(ui.ColorStatusGood, "Attack: %d | Defense: %d | Accuracy: %d%%",
			combat.AttackPower, combat.Defense, combat.Accuracy)
		g.log.AddMessagef(ui.ColorStatusGood, "Dodge: %d%% | Crit: %d%% | Crit Dmg: %d%%",
			combat.DodgeChance, combat.CriticalChance, combat.CriticalDamage)
	}

	// Skills
	if g.ecs.HasSkillsSafe(playerID) {
		skills := g.ecs.GetSkillsSafe(playerID)
		g.log.AddMessagef(ui.ColorStatusGood, "=== Skills ===")
		g.log.AddMessagef(ui.ColorStatusGood, "Melee: %d | Ranged: %d | Defense: %d",
			skills.MeleeWeapons, skills.RangedWeapons, skills.Defense)
		g.log.AddMessagef(ui.ColorStatusGood, "Stealth: %d | Perception: %d | Medicine: %d",
			skills.Stealth, skills.Perception, skills.Medicine)
		g.log.AddMessagef(ui.ColorStatusGood, "Magic: Evo %d | Con %d | Enc %d | Div %d",
			skills.Evocation, skills.Conjuration, skills.Enchantment, skills.Divination)
	}

	// Equipment
	if g.ecs.HasEquipmentSafe(playerID) {
		equipment := g.ecs.GetEquipmentSafe(playerID)
		g.log.AddMessagef(ui.ColorStatusGood, "=== Equipment ===")
		if equipment.Weapon != nil {
			g.log.AddMessagef(ui.ColorStatusGood, "Weapon: %s", equipment.Weapon.Name)
		} else {
			g.log.AddMessagef(ui.ColorStatusGood, "Weapon: None")
		}
		if equipment.Armor != nil {
			g.log.AddMessagef(ui.ColorStatusGood, "Armor: %s", equipment.Armor.Name)
		} else {
			g.log.AddMessagef(ui.ColorStatusGood, "Armor: None")
		}
	}

	return true, eff, nil // Don't consume turn
}

// handleHelpAction displays help information
func (md *Model) handleHelpAction() (again bool, eff gruid.Effect, err error) {
	g := md.game

	g.log.AddMessagef(ui.ColorStatusGood, "=== HELP ===")
	g.log.AddMessagef(ui.ColorStatusGood, "Movement: Arrow keys, WASD, or hjkl")
	g.log.AddMessagef(ui.ColorStatusGood, "Wait: . (period) or Space")
	g.log.AddMessagef(ui.ColorStatusGood, "")
	g.log.AddMessagef(ui.ColorStatusGood, "=== Inventory ===")
	g.log.AddMessagef(ui.ColorStatusGood, "g - Pick up item")
	g.log.AddMessagef(ui.ColorStatusGood, "D - Drop item")
	g.log.AddMessagef(ui.ColorStatusGood, "i - Show inventory")
	g.log.AddMessagef(ui.ColorStatusGood, "u - Use consumable item")
	g.log.AddMessagef(ui.ColorStatusGood, "e - Equip weapon/armor")
	g.log.AddMessagef(ui.ColorStatusGood, "")
	g.log.AddMessagef(ui.ColorStatusGood, "=== Character ===")
	g.log.AddMessagef(ui.ColorStatusGood, "C - Character sheet")
	g.log.AddMessagef(ui.ColorStatusGood, "")
	g.log.AddMessagef(ui.ColorStatusGood, "=== Game ===")
	g.log.AddMessagef(ui.ColorStatusGood, "S - Save game")
	g.log.AddMessagef(ui.ColorStatusGood, "L - Load game")
	g.log.AddMessagef(ui.ColorStatusGood, "Q - Quit")
	g.log.AddMessagef(ui.ColorStatusGood, "? - This help")

	return true, eff, nil // Don't consume turn
}
