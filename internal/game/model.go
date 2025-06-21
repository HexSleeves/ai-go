// This file defines the main model of the game: the Update function that
// updates the model state in response to user input, and the Draw function,
// which draws the grid.

package game

import (
	"log/slog"
	"runtime"
	"time"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ui"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/utils"
)

type mode int

const (
	modeNormal mode = iota
	modeQuit
	modeCharacterSheet
	modeInventory
	modeFullMessageLog
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

	// Full-screen UI components
	characterScreen   *ui.CharacterScreen
	inventoryScreen   *ui.InventoryScreen
	fullMessageScreen *ui.FullMessageScreen

	// Debug information
	lastUpdateTime time.Time
	updateCount    uint64
	lastEffect     gruid.Effect

	// Enhanced UI features
	showPathfindingDebug bool
	pathfindingDebugInfo *PathfindingDebugInfo
	eventQueue           []gruid.Msg
	lastInputTime        time.Time

	// Debug system
	debugLevel   DebugLevel
	showFOVDebug bool
	showAIDebug  bool
	aiDebugInfo  *AIDebugInfo
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
		characterScreen:      ui.NewCharacterScreen(),
		inventoryScreen:      ui.NewInventoryScreen(),
		fullMessageScreen:    ui.NewFullMessageScreen(),
		lastUpdateTime:       time.Now(),
		showPathfindingDebug: false,
		eventQueue:           make([]gruid.Msg, 0),
		lastInputTime:        time.Now(),
		debugLevel:           DebugNone,
		showFOVDebug:         false,
		showAIDebug:          false,
	}
}

func (md *Model) init() gruid.Effect {
	slog.Debug("========= Game Initialization Started =========")
	md.game.InitLevel()

	slog.Debug("Level initialized")
	slog.Debug("About to process turn queue for the first time")

	md.game.FOVSystem()
	md.processTurnQueue()

	slog.Debug("Initial turn queue processing completed")
	slog.Debug("========= Game Initialization Completed =========")

	if runtime.GOOS == "js" {
		return nil
	}

	return gruid.Sub(utils.HandleSignals)
}

// EndTurn finalizes player's turn and runs other events until next player
// turn.
func (md *Model) EndTurn() gruid.Effect {
	slog.Debug("EndTurn called - player finished their turn")

	md.mode = modeNormal
	g := md.game
	g.waitingForInput = false

	g.monstersTurn()
	md.processTurnQueue()

	// Track update metrics
	md.updateCount++
	md.lastUpdateTime = time.Now()

	// Update debug information if enabled
	md.UpdatePathfindingDebug()
	md.UpdateAIDebug()

	// Return nil to indicate the screen should be redrawn
	return nil
}

// GetDebugInfo returns current debug information
func (md *Model) GetDebugInfo() map[string]any {
	debugInfo := map[string]any{
		"mode":                 md.mode,
		"updateCount":          md.updateCount,
		"lastUpdateTime":       md.lastUpdateTime,
		"lastEffect":           md.lastEffect,
		"waitingForInput":      md.game.waitingForInput,
		"turnQueueSize":        md.game.turnQueue.Len(),
		"currentTime":          md.game.turnQueue.CurrentTime,
		"showPathfindingDebug": md.showPathfindingDebug,
		"eventQueueSize":       len(md.eventQueue),
		"lastInputTime":        md.lastInputTime,
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
			slog.Info("Pathfinding debug visualization enabled")
		} else {
			md.game.pathfindingMgr.DisablePathfindingDebug()
			md.pathfindingDebugInfo = nil
			slog.Info("Pathfinding debug visualization disabled")
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

// ToggleFOVDebug toggles FOV debug visualization
func (md *Model) ToggleFOVDebug() {
	md.showFOVDebug = !md.showFOVDebug
	slog.Info("FOV debug visualization", "enabled", md.showFOVDebug)
}

// ToggleAIDebug toggles AI debug visualization
func (md *Model) ToggleAIDebug() {
	md.showAIDebug = !md.showAIDebug

	if md.showAIDebug {
		md.aiDebugInfo = md.game.CollectAIDebugInfo()
		slog.Info("AI debug visualization enabled")
	} else {
		md.aiDebugInfo = nil
		slog.Info("AI debug visualization disabled")
	}
}

// CycleDebugLevel cycles through debug levels
func (md *Model) CycleDebugLevel() {
	md.debugLevel = (md.debugLevel + 1) % 5 // 0-4

	// Update individual debug flags based on level
	switch md.debugLevel {
	case DebugNone:
		md.showFOVDebug = false
		md.showPathfindingDebug = false
		md.showAIDebug = false
		if md.game.pathfindingMgr != nil {
			md.game.pathfindingMgr.DisablePathfindingDebug()
		}
		md.pathfindingDebugInfo = nil
		md.aiDebugInfo = nil
	case DebugFOV:
		md.showFOVDebug = true
		md.showPathfindingDebug = false
		md.showAIDebug = false
	case DebugPathfinding:
		md.showFOVDebug = false
		md.showPathfindingDebug = true
		md.showAIDebug = false
		if md.game.pathfindingMgr != nil {
			md.game.pathfindingMgr.EnablePathfindingDebug()
			md.pathfindingDebugInfo = md.game.pathfindingMgr.GetDebugInfo()
		}
	case DebugAI:
		md.showFOVDebug = false
		md.showPathfindingDebug = false
		md.showAIDebug = true
		md.aiDebugInfo = md.game.CollectAIDebugInfo()
	case DebugFull:
		md.showFOVDebug = true
		md.showPathfindingDebug = true
		md.showAIDebug = true
		if md.game.pathfindingMgr != nil {
			md.game.pathfindingMgr.EnablePathfindingDebug()
			md.pathfindingDebugInfo = md.game.pathfindingMgr.GetDebugInfo()
		}
		md.aiDebugInfo = md.game.CollectAIDebugInfo()
	}

	slog.Info("Debug level set to", "level", md.debugLevel.String())
}

// UpdateAIDebug updates AI debug information
func (md *Model) UpdateAIDebug() {
	if md.showAIDebug || md.debugLevel == DebugAI || md.debugLevel == DebugFull {
		md.aiDebugInfo = md.game.CollectAIDebugInfo()
	}
}

// GetAIDebugInfo returns current AI debug information
func (md *Model) GetAIDebugInfo() *AIDebugInfo {
	return md.aiDebugInfo
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
				slog.Info("Pathfinding Statistics", "stats", stats)
			}
			return nil
		case "F3":
			// Toggle FOV debug
			md.ToggleFOVDebug()
			return nil
		case "F4":
			// Toggle AI debug
			md.ToggleAIDebug()
			return nil
		case "F5":
			// Cycle debug levels
			md.CycleDebugLevel()
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
