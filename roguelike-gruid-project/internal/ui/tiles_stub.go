//go:build !sdl && !js
// +build !sdl,!js

package ui

import (
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
	"github.com/sirupsen/logrus"
)

// ToggleTileMode is a stub for non-SDL builds
func ToggleTileMode() error {
	logrus.Info("Tile mode is not available in this build. Use -tags sdl to enable tile support.")
	return nil
}

// GetCurrentTileMode is a stub for non-SDL builds
func GetCurrentTileMode() bool {
	return false // Always return false for non-SDL builds
}

// ClearTileCache is a stub for non-SDL builds
func ClearTileCache() {
	// No-op for non-SDL builds
}

// UpdateTileConfig is a stub for non-SDL builds
func UpdateTileConfig(newConfig config.TileConfig) error {
	logrus.Info("Tile configuration is not available in this build. Use -tags sdl to enable tile support.")
	return nil
}
