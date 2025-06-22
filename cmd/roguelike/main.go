//go:build !js
// +build !js

package main

import (
	"context"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/game"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ui"
)

func main() {
	config.Init()
	ui.InitializeSDL()

	slog.Info("Starting roguelike game", "debug_mode", config.Config.Advanced.DebugMode)

	seed := time.Now().UnixNano()
	rand.New(rand.NewSource(seed))

	gd := gruid.NewGrid(config.Config.Gameplay.DungeonWidth, config.Config.Gameplay.DungeonHeight)
	m := game.NewModel(gd)

	driver := ui.GetDriver()
	app := gruid.NewApp(gruid.AppConfig{
		Model:  m,
		Driver: driver,
	})

	// Set up signal handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("Starting game loop")
	if err := app.Start(ctx); err != nil {
		driver.Close()
		slog.Error("Game loop ended with error", "error", err)
	}

	slog.Info("Game loop ended gracefully")
}
