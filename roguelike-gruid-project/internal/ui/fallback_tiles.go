//go:build js || sdl
// +build js sdl

package ui

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"

	"codeberg.org/anaseto/gruid"
	"github.com/sirupsen/logrus"
)

// CreateFallbackTiles generates basic fallback tile images when the tileset is missing
func CreateFallbackTiles(tilesetPath string, tileSize int) error {
	// Ensure fallback directory exists
	fallbackDir := filepath.Join(tilesetPath, "fallback")
	if err := os.MkdirAll(fallbackDir, 0755); err != nil {
		return err
	}
	
	// Create basic fallback tiles
	tiles := map[string]func(int) image.Image{
		"unknown.png":     createUnknownTile,
		"player.png":      createPlayerTile,
		"wall.png":        createWallTile,
		"floor.png":       createFloorTile,
		"door.png":        createDoorTile,
		"monster.png":     createMonsterTile,
		"item.png":        createItemTile,
	}
	
	for filename, generator := range tiles {
		img := generator(tileSize)
		if err := saveTileImage(img, filepath.Join(fallbackDir, filename)); err != nil {
			logrus.Warnf("Failed to create fallback tile %s: %v", filename, err)
		}
	}
	
	return nil
}

// createUnknownTile creates a generic "unknown" tile
func createUnknownTile(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	// Create a magenta square with a question mark pattern
	magenta := color.RGBA{255, 0, 255, 255}
	black := color.RGBA{0, 0, 0, 255}
	
	// Fill with magenta
	draw.Draw(img, img.Bounds(), &image.Uniform{magenta}, image.Point{}, draw.Src)
	
	// Draw a simple question mark pattern
	center := size / 2
	quarter := size / 4
	
	// Question mark top curve
	for x := center - quarter; x <= center + quarter; x++ {
		img.Set(x, quarter, black)
		img.Set(x, quarter + 1, black)
	}
	
	// Question mark vertical line
	for y := quarter; y <= center; y++ {
		img.Set(center + quarter, y, black)
	}
	
	// Question mark bottom part
	for y := center; y <= center + quarter/2; y++ {
		img.Set(center, y, black)
	}
	
	// Question mark dot
	img.Set(center, center + quarter, black)
	img.Set(center, center + quarter + 1, black)
	
	return img
}

// createPlayerTile creates a basic player tile
func createPlayerTile(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	blue := color.RGBA{0, 100, 255, 255}
	white := color.RGBA{255, 255, 255, 255}
	
	// Create a simple stick figure
	center := size / 2
	quarter := size / 4
	
	// Head (circle approximation)
	for y := quarter; y <= quarter + quarter/2; y++ {
		for x := center - quarter/4; x <= center + quarter/4; x++ {
			img.Set(x, y, white)
		}
	}
	
	// Body
	for y := quarter + quarter/2; y <= size - quarter; y++ {
		img.Set(center, y, blue)
	}
	
	// Arms
	armY := quarter + quarter/2 + quarter/4
	for x := center - quarter/2; x <= center + quarter/2; x++ {
		img.Set(x, armY, blue)
	}
	
	// Legs
	legStartY := size - quarter
	for y := legStartY; y < size - 2; y++ {
		img.Set(center - quarter/4, y, blue)
		img.Set(center + quarter/4, y, blue)
	}
	
	return img
}

// createWallTile creates a basic wall tile
func createWallTile(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	gray := color.RGBA{128, 128, 128, 255}
	darkGray := color.RGBA{64, 64, 64, 255}
	
	// Fill with gray
	draw.Draw(img, img.Bounds(), &image.Uniform{gray}, image.Point{}, draw.Src)
	
	// Add brick pattern
	brickHeight := size / 4
	for y := 0; y < size; y += brickHeight {
		// Horizontal lines
		for x := 0; x < size; x++ {
			if y < size {
				img.Set(x, y, darkGray)
			}
		}
		
		// Vertical lines (offset every other row)
		offset := 0
		if (y/brickHeight)%2 == 1 {
			offset = size / 2
		}
		
		for x := offset; x < size; x += size/2 {
			for dy := 0; dy < brickHeight && y+dy < size; dy++ {
				if x < size {
					img.Set(x, y+dy, darkGray)
				}
			}
		}
	}
	
	return img
}

// createFloorTile creates a basic floor tile
func createFloorTile(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	lightBrown := color.RGBA{139, 119, 101, 255}
	darkBrown := color.RGBA{101, 87, 74, 255}
	
	// Fill with light brown
	draw.Draw(img, img.Bounds(), &image.Uniform{lightBrown}, image.Point{}, draw.Src)
	
	// Add some texture with random dark spots
	for y := 0; y < size; y += 4 {
		for x := 0; x < size; x += 4 {
			if (x+y)%8 == 0 {
				img.Set(x, y, darkBrown)
				if x+1 < size {
					img.Set(x+1, y, darkBrown)
				}
			}
		}
	}
	
	return img
}

// createDoorTile creates a basic door tile
func createDoorTile(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	brown := color.RGBA{139, 69, 19, 255}
	darkBrown := color.RGBA{101, 50, 14, 255}
	brass := color.RGBA{181, 166, 66, 255}
	
	// Fill with brown
	draw.Draw(img, img.Bounds(), &image.Uniform{brown}, image.Point{}, draw.Src)
	
	// Door frame
	for i := 0; i < 2; i++ {
		// Top and bottom
		for x := 0; x < size; x++ {
			img.Set(x, i, darkBrown)
			img.Set(x, size-1-i, darkBrown)
		}
		// Left and right
		for y := 0; y < size; y++ {
			img.Set(i, y, darkBrown)
			img.Set(size-1-i, y, darkBrown)
		}
	}
	
	// Door handle
	handleX := size - size/4
	handleY := size / 2
	img.Set(handleX, handleY, brass)
	img.Set(handleX, handleY+1, brass)
	
	return img
}

// createMonsterTile creates a basic monster tile
func createMonsterTile(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	red := color.RGBA{255, 0, 0, 255}
	darkRed := color.RGBA{128, 0, 0, 255}
	white := color.RGBA{255, 255, 255, 255}
	
	center := size / 2
	quarter := size / 4
	
	// Body (oval shape)
	for y := quarter; y < size-quarter; y++ {
		for x := quarter; x < size-quarter; x++ {
			img.Set(x, y, red)
		}
	}
	
	// Eyes
	eyeY := center - quarter/2
	img.Set(center-quarter/2, eyeY, white)
	img.Set(center+quarter/2, eyeY, white)
	img.Set(center-quarter/2, eyeY+1, darkRed)
	img.Set(center+quarter/2, eyeY+1, darkRed)
	
	// Mouth (jagged)
	mouthY := center + quarter/4
	for x := center - quarter/2; x <= center + quarter/2; x++ {
		if x%2 == 0 {
			img.Set(x, mouthY, darkRed)
			img.Set(x, mouthY+1, darkRed)
		}
	}
	
	return img
}

// createItemTile creates a basic item tile
func createItemTile(size int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	
	yellow := color.RGBA{255, 255, 0, 255}
	orange := color.RGBA{255, 165, 0, 255}
	
	center := size / 2
	quarter := size / 4
	
	// Create a simple gem/treasure shape
	// Diamond shape
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			// Distance from center
			dx := abs(x - center)
			dy := abs(y - center)
			
			if dx + dy <= quarter + quarter/2 {
				if dx + dy <= quarter {
					img.Set(x, y, yellow)
				} else {
					img.Set(x, y, orange)
				}
			}
		}
	}
	
	return img
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// saveTileImage saves an image as a PNG file
func saveTileImage(img image.Image, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	return png.Encode(file, img)
}

// EnsureFallbackTilesExist checks if fallback tiles exist and creates them if needed
func EnsureFallbackTilesExist(tilesetPath string, tileSize int) {
	fallbackDir := filepath.Join(tilesetPath, "fallback")
	unknownTile := filepath.Join(fallbackDir, "unknown.png")
	
	// Check if unknown tile exists
	if _, err := os.Stat(unknownTile); os.IsNotExist(err) {
		logrus.Info("Creating fallback tiles...")
		if err := CreateFallbackTiles(tilesetPath, tileSize); err != nil {
			logrus.Warnf("Failed to create fallback tiles: %v", err)
		}
	}
}
