//go:build sdl
// +build sdl

package ui

import (
	"codeberg.org/anaseto/gruid"
	sdl "codeberg.org/anaseto/gruid-sdl"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
	"github.com/sirupsen/logrus"
)

var driver gruid.Driver
var currentTileManager sdl.TileManager
var fontTileDrawer sdl.TileManager
var imageTileManager *ImageTileManager

func init() {
	// Get tile configuration
	tileConfig := config.LoadTileConfig()

	// Create font-based tile drawer as fallback
	var err error
	fontTileDrawer, err = GetTileDrawer()
	if err != nil {
		logrus.Fatal(err)
	}

	// Create image tile manager
	EnsureFallbackTilesExist(tileConfig.TilesetPath, tileConfig.TileSize)
	imageTileManager = NewImageTileManager(&tileConfig, fontTileDrawer)

	// Choose initial tile manager based on configuration
	if tileConfig.Enabled {
		currentTileManager = imageTileManager
		logrus.Info("Using image-based tile rendering")
	} else {
		currentTileManager = fontTileDrawer
		logrus.Info("Using font-based tile rendering")
	}

	dr := sdl.NewDriver(sdl.Config{
		TileManager: currentTileManager,
	})

	//dr.SetScale(2.0, 2.0)
	dr.PreventQuit()
	driver = dr
}

// ToggleTileMode switches between image-based and font-based tile rendering
func ToggleTileMode() error {
	// Load current configuration
	tileConfig := config.LoadTileConfig()

	// Toggle the enabled state
	tileConfig.Enabled = !tileConfig.Enabled

	// Save the new configuration
	if err := config.SaveTileConfig(tileConfig); err != nil {
		return err
	}

	// Update the image tile manager configuration
	if imageTileManager != nil {
		imageTileManager.UpdateConfig(&tileConfig)
	}

	// Switch tile manager
	if tileConfig.Enabled {
		currentTileManager = imageTileManager
		logrus.Info("Switched to image-based tile rendering")
	} else {
		currentTileManager = fontTileDrawer
		logrus.Info("Switched to font-based tile rendering")
	}

	// Update the SDL driver with the new tile manager
	if sdlDriver, ok := driver.(*sdl.Driver); ok {
		// Note: gruid-sdl doesn't have a direct way to change TileManager at runtime
		// This would require reinitializing the driver, which is complex
		// For now, we'll just update our internal state and log the change
		// The change will take effect on next restart
		logrus.Info("Tile mode change will take effect on next restart")
		_ = sdlDriver // Avoid unused variable warning
	}

	return nil
}

// GetCurrentTileMode returns whether tile mode is currently enabled
func GetCurrentTileMode() bool {
	tileConfig := config.LoadTileConfig()
	return tileConfig.Enabled
}

// ClearTileCache clears the tile cache (useful when switching tilesets)
func ClearTileCache() {
	if imageTileManager != nil {
		imageTileManager.ClearCache()
		logrus.Info("Tile cache cleared")
	}
}

// UpdateTileConfig updates the tile configuration and applies changes
func UpdateTileConfig(newConfig config.TileConfig) error {
	// Save the new configuration
	if err := config.SaveTileConfig(newConfig); err != nil {
		return err
	}

	// Update the image tile manager
	if imageTileManager != nil {
		imageTileManager.UpdateConfig(&newConfig)
	}

	logrus.Info("Tile configuration updated")
	return nil
}
