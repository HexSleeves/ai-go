package game

import (
	"log/slog"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ui"
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
	ActionScrollMessagesUp
	ActionScrollMessagesDown
	ActionScrollMessagesBottom
	ActionFullMessageLog
	ActionCloseScreen
	ActionToggleTiles
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

	slog.Debug("Normal mode action", "action", playerAction)

	switch playerAction {
	case ActionNone:
		again = true
		err = actionErrorUnknown
	case ActionW, ActionS, ActionN, ActionE:
		direction := keyToDir(playerAction)
		// Queue player movement action(s)
		md.queuePlayerMovement(direction)
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

	case ActionScrollMessagesUp:
		return md.handleScrollMessagesUpAction()

	case ActionScrollMessagesDown:
		return md.handleScrollMessagesDownAction()

	case ActionScrollMessagesBottom:
		return md.handleScrollMessagesBottomAction()

	case ActionFullMessageLog:
		return md.handleFullMessageLogAction()

	case ActionToggleTiles:
		return md.handleToggleTilesAction()

	default:
		slog.Debug("Unknown action", "action", playerAction)
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

// handleInventoryAction shows inventory screen
func (md *Model) handleInventoryAction() (again bool, eff gruid.Effect, err error) {
	md.mode = modeInventory
	md.inventoryScreen.ResetSelection()
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
		slog.Error("Save failed", "error", err)
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
		slog.Error("Load failed", "error", err)
	} else {
		g.log.AddMessagef(ui.ColorStatusGood, "Game loaded successfully!")
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
	g.log.AddMessagef(ui.ColorStatusGood, "T - Toggle tile/ASCII rendering")
	g.log.AddMessagef(ui.ColorStatusGood, "Q - Quit")
	g.log.AddMessagef(ui.ColorStatusGood, "? - This help")
	g.log.AddMessagef(ui.ColorStatusGood, "")
	g.log.AddMessagef(ui.ColorStatusGood, "=== Messages ===")
	g.log.AddMessagef(ui.ColorStatusGood, "Page Up - Scroll messages up")
	g.log.AddMessagef(ui.ColorStatusGood, "Page Down - Scroll messages down")
	g.log.AddMessagef(ui.ColorStatusGood, "M - Jump to latest messages")

	return true, eff, nil // Don't consume turn
}

// handleCharacterSheetAction displays character information
func (md *Model) handleCharacterSheetAction() (again bool, eff gruid.Effect, err error) {
	md.mode = modeCharacterSheet
	return true, eff, nil // Don't consume turn
}

// handleScrollMessagesUpAction scrolls the message log up
func (md *Model) handleScrollMessagesUpAction() (again bool, eff gruid.Effect, err error) {
	md.messagePanel.ScrollUp(md.game.MessageLog())
	return true, eff, nil // Don't consume turn
}

// handleScrollMessagesDownAction scrolls the message log down
func (md *Model) handleScrollMessagesDownAction() (again bool, eff gruid.Effect, err error) {
	md.messagePanel.ScrollDown()
	return true, eff, nil // Don't consume turn
}

// handleScrollMessagesBottomAction scrolls to the bottom of the message log
func (md *Model) handleScrollMessagesBottomAction() (again bool, eff gruid.Effect, err error) {
	md.messagePanel.ScrollToBottom()
	return true, eff, nil // Don't consume turn
}

// handleFullMessageLogAction shows the full message log screen
func (md *Model) handleFullMessageLogAction() (again bool, eff gruid.Effect, err error) {
	md.mode = modeFullMessageLog
	md.fullMessageScreen.ScrollToBottom()
	return true, eff, nil // Don't consume turn
}

// handleToggleTilesAction toggles between tile-based and ASCII rendering
func (md *Model) handleToggleTilesAction() (again bool, eff gruid.Effect, err error) {
	g := md.game

	// Toggle tile mode
	if err := ui.ToggleTileMode(); err != nil {
		g.log.AddMessagef(ui.ColorStatusBad, "Failed to toggle tile mode: %v", err)
		slog.Error("Toggle tile mode failed", "error", err)
		return true, eff, nil // Don't consume turn
	}

	// Check current state and inform user
	if ui.GetCurrentTileMode() {
		g.log.AddMessagef(ui.ColorStatusGood, "Switched to tile-based rendering! (Restart required)")
	} else {
		g.log.AddMessagef(ui.ColorStatusGood, "Switched to ASCII rendering! (Restart required)")
	}

	return true, eff, nil // Don't consume turn
}

// Player Action Buffering System

// queuePlayerMovement intelligently queues player movement actions
func (md *Model) queuePlayerMovement(direction gruid.Point) {
	g := md.game
	actor, _ := g.ecs.GetTurnActor(g.PlayerID)
	
	// Check if player already has queued actions
	if actor.PeekNextAction() != nil {
		// Player has queued actions, just add one more move
		action := MoveAction{
			Direction: direction,
			EntityID:  g.PlayerID,
		}
		actor.AddAction(action)
		return
	}
	
	// No queued actions, create a strategic movement sequence
	md.generatePlayerMovementSequence(direction)
}

// generatePlayerMovementSequence creates intelligent movement sequences for the player
func (md *Model) generatePlayerMovementSequence(direction gruid.Point) {
	g := md.game
	actor, _ := g.ecs.GetTurnActor(g.PlayerID)
	playerPos := g.GetPlayerPosition()
	
	// Always queue the primary movement action
	primaryAction := MoveAction{
		Direction: direction,
		EntityID:  g.PlayerID,
	}
	actor.AddAction(primaryAction)
	
	// Check if we should queue additional actions based on context
	nextPos := playerPos.Add(direction)
	
	// Auto-pickup behavior: if moving onto an item, queue pickup action
	if md.shouldAutoPickup(nextPos) {
		entities := g.ecs.EntitiesAt(nextPos)
		for _, entityID := range entities {
			if g.ecs.HasItemPickupSafe(entityID) {
				pickupAction := PickupAction{EntityID: g.PlayerID, ItemID: entityID}
				actor.AddAction(pickupAction)
				break // Only pickup one item automatically
			}
		}
	}
	
	// Smart door behavior: if moving toward a door, queue open action (future enhancement)
	// Smart combat: if moving toward an enemy, this will be handled by EntityBump -> AttackAction
}

// queuePlayerAction queues a single action for the player
func (md *Model) queuePlayerAction(action GameAction) {
	g := md.game
	actor, _ := g.ecs.GetTurnActor(g.PlayerID)
	actor.AddAction(action)
}

// queuePlayerActionSequence queues multiple actions for the player
func (md *Model) queuePlayerActionSequence(actions []GameAction) {
	g := md.game
	actor, _ := g.ecs.GetTurnActor(g.PlayerID)
	
	for _, action := range actions {
		actor.AddAction(action)
	}
}

// Helper functions for player action buffering

// shouldAutoPickup determines if the player should automatically pick up items
func (md *Model) shouldAutoPickup(pos gruid.Point) bool {
	// For now, always enable auto-pickup
	// TODO: Add configuration option for auto-pickup behavior
	return true
}

// getPlayerActionQueueSize returns the number of queued actions for the player
func (md *Model) getPlayerActionQueueSize() int {
	g := md.game
	actor, _ := g.ecs.GetTurnActor(g.PlayerID)
	
	// Count queued actions by peeking and consuming them
	count := 0
	for actor.PeekNextAction() != nil {
		actor.NextAction() // Remove action
		count++
	}
	
	return count
}

// clearPlayerActionQueue clears all queued actions for the player
func (md *Model) clearPlayerActionQueue() {
	g := md.game
	actor, _ := g.ecs.GetTurnActor(g.PlayerID)
	
	// Clear all queued actions
	for actor.PeekNextAction() != nil {
		actor.NextAction()
	}
}

// Helper functions for direction conversion
// Note: keyToDir function is defined in input.go
