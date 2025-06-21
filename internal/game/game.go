package game

import (
	"math/rand"
	"time"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/log"
	turn "github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/turn_queue"
)

// GameState represents the current state of the game
type GameState int

const (
	GameStateRunning GameState = iota
	GameStatePaused
	GameStateGameOver
	GameStateMenu
)

// GameStats tracks various game statistics
type GameStats struct {
	StartTime      time.Time     // When the game started
	PlayTime       time.Duration // Total time played
	TurnCount      int           // Total number of turns taken
	MonstersKilled int           // Number of monsters killed
	ItemsCollected int           // Number of items collected
	DamageDealt    int           // Total damage dealt
	DamageTaken    int           // Total damage taken
}

// Game represents the main game state.
type Game struct {
	Depth           int
	State           GameState
	waitingForInput bool

	dungeon        *Map
	ecs            *ecs.ECS
	spatialGrid    *SpatialGrid
	pathfindingMgr *PathfindingManager

	PlayerID  ecs.EntityID
	turnQueue *turn.TurnQueue
	log       *log.MessageLog
	stats     *GameStats

	rand *rand.Rand
}

func NewGame() *Game {
	return &Game{
		State:       GameStateRunning,
		ecs:         ecs.NewECS(),
		turnQueue:   turn.NewTurnQueue(),
		log:         log.NewMessageLog(),
		spatialGrid: NewSpatialGrid(config.DungeonWidth, config.DungeonHeight),
		stats: &GameStats{
			StartTime: time.Now(),
		},
	}
}

// InitLevel initializes a new game level
func (g *Game) InitLevel() {
	if g.rand == nil {
		g.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	g.Depth = 1

	// Clear the spatial grid for the new level
	g.spatialGrid.Clear()

	g.dungeon = NewMap(config.DungeonWidth, config.DungeonHeight)

	// Initialize pathfinding manager after map is created
	g.pathfindingMgr = NewPathfindingManager(g)

	items := CreateBasicItems()
	playerStart := g.dungeon.generateMap(g, config.DungeonWidth, config.DungeonHeight, items)
	g.SpawnPlayer(playerStart, items)
}

func (g *Game) GetPlayerPosition() gruid.Point {
	// Use safe accessor - no error handling needed!
	return g.ecs.GetPositionSafe(g.PlayerID)
}

// setGameOverState sets the game to game over state
func (g *Game) setGameOverState() {
	g.State = GameStateGameOver
	g.waitingForInput = false
}

// IsGameOver returns true if the game is in game over state
func (g *Game) IsGameOver() bool {
	return g.State == GameStateGameOver
}

// IsRunning returns true if the game is currently running
func (g *Game) IsRunning() bool {
	return g.State == GameStateRunning
}

// UpdatePlayTime updates the total play time
func (g *Game) UpdatePlayTime() {
	if g.stats != nil {
		g.stats.PlayTime = time.Since(g.stats.StartTime)
	}
}

// IncrementMonstersKilled increments the monsters killed counter
func (g *Game) IncrementMonstersKilled() {
	if g.stats != nil {
		g.stats.MonstersKilled++
	}
}

// IncrementItemsCollected increments the items collected counter
func (g *Game) IncrementItemsCollected() {
	if g.stats != nil {
		g.stats.ItemsCollected++
	}
}

// AddDamageDealt adds to the total damage dealt
func (g *Game) AddDamageDealt(damage int) {
	if g.stats != nil {
		g.stats.DamageDealt += damage
	}
}

// AddDamageTaken adds to the total damage taken
func (g *Game) AddDamageTaken(damage int) {
	if g.stats != nil {
		g.stats.DamageTaken += damage
	}
}

// IncrementTurnCount increments the turn counter
func (g *Game) IncrementTurnCount() {
	if g.stats != nil {
		g.stats.TurnCount++
	}
}

// ECS returns the ECS system for UI access
func (g *Game) ECS() *ecs.ECS {
	return g.ecs
}

// Stats returns the game statistics for UI access
func (g *Game) Stats() *GameStats {
	return g.stats
}

// MessageLog returns the message log for UI access
func (g *Game) MessageLog() *log.MessageLog {
	return g.log
}

// GetPlayerID returns the player entity ID for UI access
func (g *Game) GetPlayerID() ecs.EntityID {
	return g.PlayerID
}

// GetDepth returns the current dungeon depth for UI access
func (g *Game) GetDepth() int {
	return g.Depth
}

// GetMonstersKilled returns the number of monsters killed
func (gs *GameStats) GetMonstersKilled() int {
	return gs.MonstersKilled
}
