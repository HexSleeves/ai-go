package game

import (
	"fmt"
	"log/slog"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

// UpdateState represents the current state of the game update cycle
type UpdateState int

const (
	StateNormal UpdateState = iota
	StateQuit
	StateGameOver
)

// Update is the main entry point for handling game updates
func (md *Model) Update(msg gruid.Msg) gruid.Effect {
	if _, ok := msg.(gruid.MsgInit); ok {
		return md.init()
	}

	// Handle quit command
	if key, ok := msg.(gruid.MsgKeyDown); ok && key.Key == "q" {
		return gruid.End()
	}

	return md.processGameUpdate(msg)
}

// validateGameState checks for valid game state
func (md *Model) validateGameState() error {
	g := md.game

	// Validate game exists
	if g == nil {
		return fmt.Errorf("game state is nil")
	}

	// Validate turn queue
	if g.turnQueue == nil {
		return fmt.Errorf("turn queue is nil")
	}

	// Validate player exists and has required components
	if g.PlayerID != 0 {
		if !g.ecs.EntityExists(g.PlayerID) {
			return fmt.Errorf("player entity %d does not exist", g.PlayerID)
		}
		if !g.ecs.HasComponent(g.PlayerID, components.CTurnActor) {
			return fmt.Errorf("player entity %d missing TurnActor component", g.PlayerID)
		}
	}

	return nil
}

// processGameUpdate handles the main game update logic with clear state management
func (md *Model) processGameUpdate(msg gruid.Msg) gruid.Effect {
	// Validate game state
	if err := md.validateGameState(); err != nil {
		slog.Debug("Invalid game state", "error", err)
		return nil
	}

	g := md.game

	// Log the current game state for debugging
	slog.Debug("Processing game update",
		"waitingForInput", g.waitingForInput,
		"gameMode", md.mode,
		"turnQueueSize", g.turnQueue.Len(),
		"currentTime", g.turnQueue.CurrentTime)

	// Process based on current state
	if g.waitingForInput {
		return md.handlePlayerInput(msg)
	}

	// Process the turn queue and return appropriate effect
	return md.processTurnQueueWithEffect()
}

// handlePlayerInput processes player input when it's the player's turn
func (md *Model) handlePlayerInput(msg gruid.Msg) gruid.Effect {
	slog.Debug("Handling player input",
		"messageType", fmt.Sprintf("%T", msg),
		"gameMode", md.mode)

	// Validate player state
	if !md.game.ecs.EntityExists(md.game.PlayerID) {
		slog.Debug("Player entity does not exist")
		return nil
	}

	var effect gruid.Effect
	switch md.mode {
	case modeQuit:
		slog.Debug("In quit mode, ignoring input")
		return nil
	case modeNormal:
		effect = md.processNormalModeInput(msg)
	case modeCharacterSheet, modeInventory, modeFullMessageLog:
		effect = md.processScreenModeInput(msg)
	default:
		slog.Debug("Unexpected game mode", "mode", md.mode)
		return nil
	}

	return effect
}

// processNormalModeInput handles input during normal gameplay
func (md *Model) processNormalModeInput(msg gruid.Msg) gruid.Effect {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		return md.handleKeyDown(msg)
	case gruid.MsgMouse:
		return md.handleMouse(msg)
	default:
		slog.Debug("Unhandled message type", "type", fmt.Sprintf("%T", msg))
		return nil
	}
}

// processScreenModeInput handles input during full-screen UI modes
func (md *Model) processScreenModeInput(msg gruid.Msg) gruid.Effect {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		return md.handleScreenKeyDown(msg)
	default:
		slog.Debug("Unhandled message type in screen mode", "type", fmt.Sprintf("%T", msg))
		return nil
	}
}

// handleKeyDown processes keyboard input
func (md *Model) handleKeyDown(msg gruid.MsgKeyDown) gruid.Effect {
	again, effect, err := md.normalModeKeyDown(msg.Key, msg.Mod&gruid.ModShift != 0)
	if err != nil {
		slog.Debug("Error processing key down", "error", err)
	}

	if again {
		return effect
	}

	return md.EndTurn()
}

// handleScreenKeyDown processes keyboard input in screen modes
func (md *Model) handleScreenKeyDown(msg gruid.MsgKeyDown) gruid.Effect {
	again, effect, err := md.screenModeKeyDown(msg.Key)
	if err != nil {
		slog.Debug("Error processing screen key down", "error", err)
	}

	if again {
		return effect
	}

	// Screen modes don't end turns, just return the effect
	return effect
}

// handleMouse processes mouse input
func (md *Model) handleMouse(msg gruid.MsgMouse) gruid.Effect {
	// TODO: Implement mouse handling if needed
	return nil
}

// processTurnQueueWithEffect processes the turn queue and returns appropriate UI effect
func (md *Model) processTurnQueueWithEffect() gruid.Effect {
	slog.Debug("Processing turn queue")

	// Process the turn queue
	md.processTurnQueue()

	// Return nil to trigger a redraw
	// This ensures the screen updates after monster moves
	return nil
}

// normalModeKeyDown processes a key press in normal mode
func (md *Model) normalModeKeyDown(key gruid.Key, shift bool) (again bool, effect gruid.Effect, err error) {
	action := KEYS_NORMAL[key]
	again, effect, err = md.normalModeAction(action)
	if _, ok := err.(actionError); ok {
		err = fmt.Errorf("key '%s' does nothing. Type ? for help", key)
	}
	return again, effect, err
}

// screenModeKeyDown processes a key press in screen modes
func (md *Model) screenModeKeyDown(key gruid.Key) (again bool, effect gruid.Effect, err error) {
	action := KEYS_SCREEN[key]
	again, effect, err = md.screenModeAction(action)
	if _, ok := err.(actionError); ok {
		err = fmt.Errorf("key '%s' does nothing in this screen", key)
	}
	return again, effect, err
}

// screenModeAction processes actions in screen modes
func (md *Model) screenModeAction(action playerAction) (again bool, effect gruid.Effect, err error) {
	switch action {
	case ActionCloseScreen:
		md.mode = modeNormal
		return true, effect, nil

	case ActionScrollMessagesUp:
		switch md.mode {
		case modeFullMessageLog:
			md.fullMessageScreen.ScrollUp(md.game.MessageLog(), 1)
		case modeInventory:
			md.inventoryScreen.ScrollUp()
		case modeCharacterSheet:
			md.characterScreen.ScrollUp(1)
		}
		return true, effect, nil

	case ActionScrollMessagesDown:
		switch md.mode {
		case modeFullMessageLog:
			md.fullMessageScreen.ScrollDown(1)
		case modeInventory:
			inventory := md.game.ecs.GetInventorySafe(md.game.PlayerID)
			md.inventoryScreen.ScrollDown(len(inventory.Items))
		case modeCharacterSheet:
			md.characterScreen.ScrollDown(1, &gameDataAdapter{md.game})
		}
		return true, effect, nil

	default:
		return true, effect, actionErrorUnknown
	}
}
