// This file defines the main model of the game: the Update function that
// updates the model state in response to user input, and the Draw function,
// which draws the grid.

package game

import (
	"runtime"
	"time"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ui"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/utils"
	"github.com/sirupsen/logrus"
)

type mode int

const (
	modeNormal mode = iota
	modeQuit
)

// Model represents the game model that implements gruid.Model
type Model struct {
	grid gruid.Grid
	game *Game
	mode mode

	// UI components
	camera       *ui.Camera
	statsPanel   *ui.StatsPanel
	messagePanel *ui.MessagePanel

	// Debug information
	lastUpdateTime time.Time
	updateCount    uint64
	lastEffect     gruid.Effect
}

// NewModel creates a new game model
func NewModel(grid gruid.Grid) *Model {
	return &Model{
		grid:           grid,
		game:           NewGame(),
		mode:           modeNormal,
		camera:         ui.NewCamera(40, 12), // Center of default map
		statsPanel:     ui.NewStatsPanel(),
		messagePanel:   ui.NewMessagePanel(),
		lastUpdateTime: time.Now(),
	}
}

func (md *Model) init() gruid.Effect {
	logrus.Debug("========= Game Initialization Started =========")
	md.game.InitLevel()

	logrus.Debug("Level initialized")
	logrus.Debug("About to process turn queue for the first time")

	md.game.FOVSystem()
	md.processTurnQueue()

	logrus.Debug("Initial turn queue processing completed")
	logrus.Debug("========= Game Initialization Completed =========")

	if runtime.GOOS == "js" {
		return nil
	}

	return gruid.Sub(utils.HandleSignals)
}

// EndTurn finalizes player's turn and runs other events until next player
// turn.
func (md *Model) EndTurn() gruid.Effect {
	logrus.Debug("EndTurn called - player finished their turn")

	md.mode = modeNormal
	g := md.game
	g.waitingForInput = false

	g.monstersTurn()
	md.processTurnQueue()

	// Track update metrics
	md.updateCount++
	md.lastUpdateTime = time.Now()

	// Return nil to indicate the screen should be redrawn
	return nil
}

// GetDebugInfo returns current debug information
func (md *Model) GetDebugInfo() map[string]any {
	return map[string]any{
		"mode":            md.mode,
		"updateCount":     md.updateCount,
		"lastUpdateTime":  md.lastUpdateTime,
		"lastEffect":      md.lastEffect,
		"waitingForInput": md.game.waitingForInput,
		"turnQueueSize":   md.game.turnQueue.Len(),
		"currentTime":     md.game.turnQueue.CurrentTime,
	}
}
