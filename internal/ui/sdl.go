//go:build !js
// +build !js

package ui

import (
	"log/slog"
	"os"

	"codeberg.org/anaseto/gruid"
	sdl "codeberg.org/anaseto/gruid-sdl"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
)

var driver gruid.Driver
var currentTileManager sdl.TileManager
var fontTileDrawer sdl.TileManager
var imageTileManager *ImageTileManager

// InitializeSDL initializes the SDL driver and tile managers
func InitializeSDL() {
	// Use the global config instance
	displayConfig := config.Config.Display

	// Create font-based tile drawer as fallback
	var err error
	fontTileDrawer, err = GetTileDrawer(displayConfig)
	if err != nil {
		slog.Error("Failed to create font-based tile drawer", "error", err)
		os.Exit(1)
	}

	// Create image tile manager
	imageTileManager = NewImageTileManager(&displayConfig, fontTileDrawer)

	// Choose initial tile manager based on configuration
	if displayConfig.TilesEnabled {
		currentTileManager = imageTileManager
		slog.Info("Using image-based tile rendering")
	} else {
		currentTileManager = fontTileDrawer
		slog.Info("Using font-based tile rendering")
	}

	// Window size
	slog.Info("Window size: ", displayConfig.WindowWidth, "x", displayConfig.WindowHeight)

	// Log the actual tile size being used
	tileSize := currentTileManager.TileSize()
	slog.Info("Tile size: %dx%d pixels", tileSize.X, tileSize.Y)

	// Font size
	slog.Info("Font size: ", displayConfig.FontSize)

	// Use configured window size
	dr := sdl.NewDriver(sdl.Config{
		TileManager: currentTileManager,
		Width:       int32(displayConfig.WindowWidth),
		Height:      int32(displayConfig.WindowHeight),
		Fullscreen:  displayConfig.Fullscreen,
		WindowTitle: "Roguelike",
		Accelerated: false,
	})

	dr.SetScale(displayConfig.ScaleFactorX, displayConfig.ScaleFactorY)
	dr.PreventQuit()

	driver = dr
}

// ToggleTileMode switches between image-based and font-based tile rendering
func ToggleTileMode() error {
	// Get current configuration
	currentConfig := config.Config

	// Toggle the enabled state
	currentConfig.Display.TilesEnabled = !currentConfig.Display.TilesEnabled

	// Update the image tile manager configuration first
	if imageTileManager != nil {
		imageTileManager.UpdateConfig(&currentConfig.Display)
	}

	// Save the new configuration only after successful update
	if err := config.SaveConfig(currentConfig); err != nil {
		// Revert the tile manager state if save fails
		currentConfig.Display.TilesEnabled = !currentConfig.Display.TilesEnabled
		if imageTileManager != nil {
			imageTileManager.UpdateConfig(&currentConfig.Display)
		}
		return err
	}

	// Switch tile manager
	if currentConfig.Display.TilesEnabled {
		currentTileManager = imageTileManager
		slog.Info("Switched to image-based tile rendering")
	} else {
		currentTileManager = fontTileDrawer
		slog.Info("Switched to font-based tile rendering")
	}

	// Note: gruid-sdl doesn't have a direct way to change TileManager at runtime
	// This would require reinitializing the driver, which is complex
	// For now, we'll just update our internal state and log the change
	// The change will take effect on next restart
	slog.Info("Tile mode change will take effect on next restart")

	return nil
}

// GetCurrentTileMode returns whether tile mode is currently enabled
func GetCurrentTileMode() bool {
	return config.Config.Display.TilesEnabled
}

// ClearTileCache clears the tile cache (useful when switching tilesets)
func ClearTileCache() {
	if imageTileManager != nil {
		imageTileManager.ClearCache()
		slog.Info("Tile cache cleared")
	}
}

// UpdateTileConfig updates the tile configuration and applies changes
func UpdateTileConfig(newConfig config.DisplayConfig) error {
	// Get current full config
	fullConfig := config.Config

	// Update only the Display section
	fullConfig.Display = newConfig

	// Save the new configuration
	if err := config.SaveConfig(fullConfig); err != nil {
		return err
	}

	// Update the image tile manager
	if imageTileManager != nil {
		imageTileManager.UpdateConfig(&newConfig)
	}

	slog.Info("Tile configuration updated")
	return nil
}
