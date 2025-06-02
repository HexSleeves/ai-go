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

// Game represents the main game state.
type Game struct {
	Depth           int
	State           GameState
	waitingForInput bool

	dungeon     *Map
	ecs         *ecs.ECS
	spatialGrid *SpatialGrid

	PlayerID  ecs.EntityID
	turnQueue *turn.TurnQueue
	log       *log.MessageLog

	rand *rand.Rand
}

func NewGame() *Game {
	return &Game{
		State:       GameStateRunning,
		ecs:         ecs.NewECS(),
		turnQueue:   turn.NewTurnQueue(),
		log:         log.NewMessageLog(),
		spatialGrid: NewSpatialGrid(config.DungeonWidth, config.DungeonHeight),
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
	playerStart := g.dungeon.generateMap(g, config.DungeonWidth, config.DungeonHeight)
	g.SpawnPlayer(playerStart)
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
