package game

import (
	"fmt"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/sirupsen/logrus"
)

// GameState represents the current state of the game update cycle
type GameState int

const (
	StateNormal GameState = iota
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
		logrus.WithError(err).Error("Invalid game state")
		return nil
	}

	g := md.game

	// Log the current game state for debugging
	logrus.WithFields(logrus.Fields{
		"waitingForInput": g.waitingForInput,
		"gameMode":        md.mode,
		"turnQueueSize":   g.turnQueue.Len(),
		"currentTime":     g.turnQueue.CurrentTime,
	}).Debug("Processing game update")

	// Process based on current state
	if g.waitingForInput {
		return md.handlePlayerInput(msg)
	}

	// Process the turn queue and return appropriate effect
	return md.processTurnQueueWithEffect()
}

// handlePlayerInput processes player input when it's the player's turn
func (md *Model) handlePlayerInput(msg gruid.Msg) gruid.Effect {
	logrus.WithFields(logrus.Fields{
		"messageType": fmt.Sprintf("%T", msg),
		"gameMode":    md.mode,
	}).Debug("Handling player input")

	// Validate player state
	if !md.game.ecs.EntityExists(md.game.PlayerID) {
		logrus.Error("Player entity does not exist")
		return nil
	}

	var effect gruid.Effect
	switch md.mode {
	case modeQuit:
		logrus.Debug("In quit mode, ignoring input")
		return nil
	case modeNormal:
		effect = md.processNormalModeInput(msg)
	default:
		logrus.Warnf("Unexpected game mode: %v", md.mode)
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
		logrus.Debugf("Unhandled message type: %T", msg)
		return nil
	}
}

// handleKeyDown processes keyboard input
func (md *Model) handleKeyDown(msg gruid.MsgKeyDown) gruid.Effect {
	again, effect, err := md.normalModeKeyDown(msg.Key, msg.Mod&gruid.ModShift != 0)
	if err != nil {
		logrus.WithError(err).Debug("Error processing key down")
		md.game.Print(err.Error())
	}

	if again {
		return effect
	}

	return md.EndTurn()
}

// handleMouse processes mouse input
func (md *Model) handleMouse(msg gruid.MsgMouse) gruid.Effect {
	// TODO: Implement mouse handling if needed
	return nil
}

// processTurnQueueWithEffect processes the turn queue and returns appropriate UI effect
func (md *Model) processTurnQueueWithEffect() gruid.Effect {
	logrus.Debug("Processing turn queue")

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
