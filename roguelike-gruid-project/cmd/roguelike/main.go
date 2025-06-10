//go:build !js
// +build !js

package main

import (
	"context"
	"math/rand"
	"time"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/game"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ui"
	"github.com/sirupsen/logrus"
)

func main() {
	config.Init()
	ui.InitializeSDL()

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
		DisableColors: false,
	})

	logrus.Infof("Starting roguelike game - Debug mode: %v", config.Config.Advanced.DebugMode)

	seed := time.Now().UnixNano()
	rand.New(rand.NewSource(seed))

	gd := gruid.NewGrid(config.Config.Gameplay.DungeonWidth, config.Config.Gameplay.DungeonHeight)
	m := game.NewModel(gd)

	driver := ui.GetDriver()
	app := gruid.NewApp(gruid.AppConfig{
		Model:  m,
		Driver: driver,
	})

	logrus.Info("Starting game loop")
	if err := app.Start(context.Background()); err != nil {
		driver.Close()
		logrus.Fatal(err)
	}
}
