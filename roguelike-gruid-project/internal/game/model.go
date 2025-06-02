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

	// Enhanced UI features
	showPathfindingDebug bool
	pathfindingDebugInfo *PathfindingDebugInfo
	eventQueue           []gruid.Msg
	lastInputTime        time.Time
}

// NewModel creates a new game model
func NewModel(grid gruid.Grid) *Model {
	return &Model{
		grid:                 grid,
		game:                 NewGame(),
		mode:                 modeNormal,
		camera:               ui.NewCamera(40, 12), // Center of default map
		statsPanel:           ui.NewStatsPanel(),
		messagePanel:         ui.NewMessagePanel(),
		lastUpdateTime:       time.Now(),
		showPathfindingDebug: false,
		eventQueue:           make([]gruid.Msg, 0),
		lastInputTime:        time.Now(),
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

	// Update pathfinding debug information if enabled
	md.UpdatePathfindingDebug()

	// Return nil to indicate the screen should be redrawn
	return nil
}

// GetDebugInfo returns current debug information
func (md *Model) GetDebugInfo() map[string]any {
	debugInfo := map[string]any{
		"mode":                   md.mode,
		"updateCount":            md.updateCount,
		"lastUpdateTime":         md.lastUpdateTime,
		"lastEffect":             md.lastEffect,
		"waitingForInput":        md.game.waitingForInput,
		"turnQueueSize":          md.game.turnQueue.Len(),
		"currentTime":            md.game.turnQueue.CurrentTime,
		"showPathfindingDebug":   md.showPathfindingDebug,
		"eventQueueSize":         len(md.eventQueue),
		"lastInputTime":          md.lastInputTime,
	}

	// Add pathfinding statistics if available
	if md.game.pathfindingMgr != nil {
		debugInfo["pathfindingStats"] = md.game.pathfindingMgr.GetPathfindingStats()
	}

	return debugInfo
}

// TogglePathfindingDebug toggles pathfinding debug visualization
func (md *Model) TogglePathfindingDebug() {
	md.showPathfindingDebug = !md.showPathfindingDebug

	if md.game.pathfindingMgr != nil {
		if md.showPathfindingDebug {
			md.game.pathfindingMgr.EnablePathfindingDebug()
			md.pathfindingDebugInfo = md.game.pathfindingMgr.GetDebugInfo()
			logrus.Info("Pathfinding debug visualization enabled")
		} else {
			md.game.pathfindingMgr.DisablePathfindingDebug()
			md.pathfindingDebugInfo = nil
			logrus.Info("Pathfinding debug visualization disabled")
		}
	}
}

// UpdatePathfindingDebug updates pathfinding debug information
func (md *Model) UpdatePathfindingDebug() {
	if md.showPathfindingDebug && md.game.pathfindingMgr != nil {
		md.pathfindingDebugInfo = md.game.pathfindingMgr.GetDebugInfo()
	}
}

// GetPathfindingDebugInfo returns current pathfinding debug information
func (md *Model) GetPathfindingDebugInfo() *PathfindingDebugInfo {
	return md.pathfindingDebugInfo
}

// QueueEvent adds an event to the event queue for processing
func (md *Model) QueueEvent(msg gruid.Msg) {
	md.eventQueue = append(md.eventQueue, msg)
}

// ProcessEventQueue processes all queued events
func (md *Model) ProcessEventQueue() []gruid.Effect {
	var effects []gruid.Effect

	for _, msg := range md.eventQueue {
		if effect := md.processEvent(msg); effect != nil {
			effects = append(effects, effect)
		}
	}

	// Clear the queue
	md.eventQueue = md.eventQueue[:0]

	return effects
}

// processEvent processes a single event
func (md *Model) processEvent(msg gruid.Msg) gruid.Effect {
	md.lastInputTime = time.Now()

	// Handle debug key combinations first
	if keyMsg, ok := msg.(gruid.MsgKeyDown); ok {
		switch keyMsg.Key {
		case "F1":
			// Toggle pathfinding debug
			md.TogglePathfindingDebug()
			return nil
		case "F2":
			// Print pathfinding statistics
			if md.game.pathfindingMgr != nil {
				stats := md.game.pathfindingMgr.GetPathfindingStats()
				logrus.WithFields(logrus.Fields(stats)).Info("Pathfinding Statistics")
			}
			return nil
		}
	}

	// Process normal game events through the existing Update system
	return md.processGameUpdate(msg)
}

// GetInputResponsiveness returns input responsiveness metrics
func (md *Model) GetInputResponsiveness() time.Duration {
	return time.Since(md.lastInputTime)
}
