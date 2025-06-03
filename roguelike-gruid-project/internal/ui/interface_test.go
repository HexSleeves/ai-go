//go:build js || sdl
// +build js sdl

package ui

import (
	"image"
	"image/color"
	"testing"

	"codeberg.org/anaseto/gruid"
	sdl "codeberg.org/anaseto/gruid-sdl"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
)

// TestTileManagerInterface verifies that our implementations satisfy the sdl.TileManager interface
func TestTileManagerInterface(t *testing.T) {
	// Test that TileDrawer implements sdl.TileManager
	var _ sdl.TileManager = (*TileDrawer)(nil)
	
	// Test that ImageTileManager implements sdl.TileManager
	var _ sdl.TileManager = (*ImageTileManager)(nil)
	
	t.Log("Both TileDrawer and ImageTileManager implement sdl.TileManager interface")
}

// TestImageTileManagerCreation tests creating an ImageTileManager
func TestImageTileManagerCreation(t *testing.T) {
	// Create a mock font tile manager
	mockFontTileManager := &MockTileManager{}
	
	// Create tile config
	config := &config.TileConfig{
		Enabled:      true,
		TileSize:     16,
		ScaleFactor:  1.0,
		TilesetPath:  t.TempDir(),
		UseSmoothing: true,
		CacheSize:    100,
	}
	
	// Create ImageTileManager
	itm := NewImageTileManager(config, mockFontTileManager)
	
	if itm == nil {
		t.Fatal("NewImageTileManager returned nil")
	}
	
	// Test TileSize method
	size := itm.TileSize()
	expectedSize := gruid.Point{X: 16, Y: 16}
	if size != expectedSize {
		t.Errorf("Expected tile size %v, got %v", expectedSize, size)
	}
	
	// Test GetImage method with a simple cell
	cell := gruid.Cell{
		Rune:  '@',
		Style: gruid.Style{Fg: gruid.ColorDefault, Bg: gruid.ColorDefault},
	}
	
	img := itm.GetImage(cell)
	if img == nil {
		t.Error("GetImage returned nil")
	}
}

// MockTileManager is a simple mock implementation for testing
type MockTileManager struct{}

func (m *MockTileManager) GetImage(c gruid.Cell) image.Image {
	// Return a simple 16x16 image
	return createSimpleTestImage(16, 16)
}

func (m *MockTileManager) TileSize() gruid.Point {
	return gruid.Point{X: 16, Y: 16}
}

// createSimpleTestImage creates a simple test image
func createSimpleTestImage(width, height int) image.Image {
	// This function is defined in fallback_tiles.go, but we'll create a simple version here
	// to avoid dependencies in the test
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a simple pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if (x+y)%2 == 0 {
				img.Set(x, y, color.RGBA{255, 255, 255, 255}) // White
			} else {
				img.Set(x, y, color.RGBA{0, 0, 0, 255}) // Black
			}
		}
	}
	return img
}

// TestTileManagerCompatibility tests that our tile managers work with gruid-sdl
func TestTileManagerCompatibility(t *testing.T) {
	// Create a mock config
	cfg := sdl.Config{
		TileManager: &MockTileManager{},
		Width:       80,
		Height:      24,
		WindowTitle: "Test",
	}
	
	// This should not panic - it tests that our interface is compatible
	driver := sdl.NewDriver(cfg)
	if driver == nil {
		t.Error("Failed to create SDL driver with our TileManager")
	}
	
	// Test with ImageTileManager
	tileConfig := &config.TileConfig{
		Enabled:      true,
		TileSize:     16,
		ScaleFactor:  1.0,
		TilesetPath:  t.TempDir(),
		UseSmoothing: true,
		CacheSize:    100,
	}
	
	itm := NewImageTileManager(tileConfig, &MockTileManager{})
	
	cfg2 := sdl.Config{
		TileManager: itm,
		Width:       80,
		Height:      24,
		WindowTitle: "Test",
	}
	
	driver2 := sdl.NewDriver(cfg2)
	if driver2 == nil {
		t.Error("Failed to create SDL driver with ImageTileManager")
	}
}
