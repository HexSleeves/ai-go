package config

import (
	"flag"
	"os"

	"github.com/sirupsen/logrus"
)

// GameConfig holds the configuration for the game
type GameConfig struct {
	DebugLogging bool
	Tiles        TileConfig
}

// ParseFlags parses command-line flags and returns a GameConfig
func ParseFlags() *GameConfig {
	config := &GameConfig{
		Tiles: LoadTileConfig(), // Load tile configuration
	}
	flag.BoolVar(&config.DebugLogging, "debug", false, "Enable debug logging")
	flag.BoolVar(&config.DebugLogging, "d", false, "Enable debug logging (shorthand)")
	flag.BoolVar(&config.Tiles.Enabled, "tiles", config.Tiles.Enabled, "Enable tile-based rendering")
	flag.Parse()
	if config.DebugLogging {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.SetOutput(os.Stdout)
		logrus.Debug("Debug logging enabled")
	} else {
		logrus.SetLevel(logrus.InfoLevel)
		logrus.SetOutput(os.Stdout)
	}

	return config
}

// Global configuration instance
var Config *GameConfig

// Init initializes the configuration
func Init() {
	Config = ParseFlags()
}
