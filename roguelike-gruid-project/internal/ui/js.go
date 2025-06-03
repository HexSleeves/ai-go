//go:build js
// +build js

package ui

import (
	"context"

	"codeberg.org/anaseto/gruid"
	js "codeberg.org/anaseto/gruid-js"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
)

var driver gruid.Driver

func init() {
	// Initialize JavaScript driver for WebAssembly builds
	dr := js.NewDriver(js.Config{
		TileSize: 24, // Use a reasonable tile size for web
	})
	driver = dr
}

// GetDriver returns the JavaScript driver for WebAssembly builds
func GetDriver() gruid.Driver {
	return driver
}

// Tile-related functions for JavaScript builds (stubs since tiles aren't supported in JS yet)
func ToggleTileMode() error {
	// No-op for JavaScript builds
	return nil
}

func GetCurrentTileMode() bool {
	return false // Always return false for JavaScript builds
}

func ClearTileCache() {
	// No-op for JavaScript builds
}

func UpdateTileConfig(newConfig config.TileConfig) error {
	// No-op for JavaScript builds
	return nil
}

func subSig(ctx context.Context, msgs chan<- gruid.Msg) {
	// do nothing
}
