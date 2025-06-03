//go:build js || sdl
// +build js sdl

package ui

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"sync"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
	"github.com/sirupsen/logrus"
)

// ImageTileManager implements gruid.TileManager for image-based tiles
type ImageTileManager struct {
	tileCache     map[string]image.Image
	coloredCache  map[string]map[gruid.Color]image.Image
	tileMapping   *TileMapping
	config        *config.TileConfig
	mutex         sync.RWMutex
	fontFallback  gruid.TileManager // Fallback to font-based rendering
}

// NewImageTileManager creates a new image-based tile manager
func NewImageTileManager(config *config.TileConfig, fontFallback gruid.TileManager) *ImageTileManager {
	itm := &ImageTileManager{
		tileCache:    make(map[string]image.Image),
		coloredCache: make(map[string]map[gruid.Color]image.Image),
		tileMapping:  NewTileMapping(config.TilesetPath),
		config:       config,
		fontFallback: fontFallback,
	}
	
	// Pre-load common tiles
	itm.preloadCommonTiles()
	
	return itm
}

// GetImage implements gruid.TileManager.GetImage
func (itm *ImageTileManager) GetImage(c gruid.Cell) image.Image {
	// If tiles are disabled, use font fallback
	if !itm.config.Enabled {
		if itm.fontFallback != nil {
			return itm.fontFallback.GetImage(c)
		}
		return itm.generateFallbackImage(c)
	}
	
	// Get tile path for this rune
	tilePath := itm.tileMapping.GetTileForRune(c.Rune)
	
	// Try to get colored version from cache
	if img := itm.getCachedColoredTile(tilePath, c.Style.Fg, c.Style.Bg); img != nil {
		return img
	}
	
	// Load base tile
	baseImg := itm.loadTile(tilePath)
	if baseImg == nil {
		// Fallback to font rendering if tile loading fails
		if itm.fontFallback != nil {
			return itm.fontFallback.GetImage(c)
		}
		return itm.generateFallbackImage(c)
	}
	
	// Apply colors and cache result
	coloredImg := itm.applyColors(baseImg, c.Style.Fg, c.Style.Bg, c.Style.Attrs)
	itm.cacheColoredTile(tilePath, c.Style.Fg, coloredImg)
	
	return coloredImg
}

// TileSize implements gruid.TileManager.TileSize
func (itm *ImageTileManager) TileSize() gruid.Point {
	size := itm.config.TileSize
	scale := itm.config.ScaleFactor
	return gruid.Point{
		X: int(float32(size) * scale),
		Y: int(float32(size) * scale),
	}
}

// loadTile loads a tile image from disk, with caching
func (itm *ImageTileManager) loadTile(tilePath string) image.Image {
	itm.mutex.RLock()
	if img, exists := itm.tileCache[tilePath]; exists {
		itm.mutex.RUnlock()
		return img
	}
	itm.mutex.RUnlock()
	
	// Load image from file
	file, err := os.Open(tilePath)
	if err != nil {
		logrus.Debugf("Failed to open tile file %s: %v", tilePath, err)
		return nil
	}
	defer file.Close()
	
	img, err := png.Decode(file)
	if err != nil {
		logrus.Debugf("Failed to decode PNG tile %s: %v", tilePath, err)
		return nil
	}
	
	// Scale image if necessary
	scaledImg := itm.scaleImage(img)
	
	// Cache the loaded tile
	itm.mutex.Lock()
	itm.tileCache[tilePath] = scaledImg
	
	// Implement simple LRU eviction if cache is too large
	if len(itm.tileCache) > itm.config.CacheSize {
		itm.evictOldestTiles()
	}
	itm.mutex.Unlock()
	
	return scaledImg
}

// getCachedColoredTile retrieves a colored tile from cache
func (itm *ImageTileManager) getCachedColoredTile(tilePath string, fg, bg gruid.Color) image.Image {
	itm.mutex.RLock()
	defer itm.mutex.RUnlock()
	
	if colorMap, exists := itm.coloredCache[tilePath]; exists {
		if img, exists := colorMap[fg]; exists {
			return img
		}
	}
	return nil
}

// cacheColoredTile stores a colored tile in cache
func (itm *ImageTileManager) cacheColoredTile(tilePath string, fg gruid.Color, img image.Image) {
	itm.mutex.Lock()
	defer itm.mutex.Unlock()
	
	if itm.coloredCache[tilePath] == nil {
		itm.coloredCache[tilePath] = make(map[gruid.Color]image.Image)
	}
	itm.coloredCache[tilePath][fg] = img
}

// applyColors applies foreground and background colors to a tile image
func (itm *ImageTileManager) applyColors(baseImg image.Image, fg, bg gruid.Color, attrs gruid.AttrMask) image.Image {
	bounds := baseImg.Bounds()
	coloredImg := image.NewRGBA(bounds)
	
	fgColor := ColorToRGBA(fg, true)
	bgColor := ColorToRGBA(bg, false)
	
	// Handle reverse attribute
	if attrs&AttrReverse != 0 {
		fgColor, bgColor = bgColor, fgColor
	}
	
	// Apply colors based on the original image
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := baseImg.At(x, y)
			r, g, b, a := originalColor.RGBA()
			
			// If pixel is transparent, use background color
			if a == 0 {
				coloredImg.Set(x, y, bgColor)
			} else {
				// Use grayscale value to determine how much foreground color to apply
				gray := (r + g + b) / 3
				if gray > 32768 { // Threshold for foreground vs background
					coloredImg.Set(x, y, fgColor)
				} else {
					coloredImg.Set(x, y, bgColor)
				}
			}
		}
	}
	
	return coloredImg
}

// scaleImage scales an image according to the configured scale factor
func (itm *ImageTileManager) scaleImage(img image.Image) image.Image {
	if itm.config.ScaleFactor == 1.0 {
		return img
	}
	
	bounds := img.Bounds()
	newWidth := int(float32(bounds.Dx()) * itm.config.ScaleFactor)
	newHeight := int(float32(bounds.Dy()) * itm.config.ScaleFactor)
	
	scaledImg := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	
	// Simple nearest-neighbor scaling
	for y := 0; y < newHeight; y++ {
		for x := 0; x < newWidth; x++ {
			srcX := int(float32(x) / itm.config.ScaleFactor)
			srcY := int(float32(y) / itm.config.ScaleFactor)
			
			if srcX < bounds.Max.X && srcY < bounds.Max.Y {
				scaledImg.Set(x, y, img.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
			}
		}
	}
	
	return scaledImg
}

// generateFallbackImage creates a simple colored rectangle as fallback
func (itm *ImageTileManager) generateFallbackImage(c gruid.Cell) image.Image {
	size := itm.TileSize()
	img := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
	
	fgColor := ColorToRGBA(c.Style.Fg, true)
	bgColor := ColorToRGBA(c.Style.Bg, false)
	
	// Fill with background color
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)
	
	// Draw a simple representation of the character
	// This is a very basic fallback - in practice you might want something more sophisticated
	centerX, centerY := size.X/2, size.Y/2
	for y := centerY - 2; y <= centerY + 2; y++ {
		for x := centerX - 2; x <= centerX + 2; x++ {
			if x >= 0 && x < size.X && y >= 0 && y < size.Y {
				img.Set(x, y, fgColor)
			}
		}
	}
	
	return img
}

// preloadCommonTiles loads frequently used tiles into cache
func (itm *ImageTileManager) preloadCommonTiles() {
	commonRunes := []rune{'@', '#', '.', '+', '!', '?', 'o', 'g', 's'}
	
	for _, r := range commonRunes {
		tilePath := itm.tileMapping.GetTileForRune(r)
		itm.loadTile(tilePath) // This will cache the tile
	}
}

// evictOldestTiles removes some tiles from cache to free memory
func (itm *ImageTileManager) evictOldestTiles() {
	// Simple eviction: remove 25% of cached tiles
	// In a more sophisticated implementation, you'd use LRU
	targetSize := itm.config.CacheSize * 3 / 4
	
	count := 0
	for tilePath := range itm.tileCache {
		if count >= len(itm.tileCache) - targetSize {
			break
		}
		delete(itm.tileCache, tilePath)
		delete(itm.coloredCache, tilePath)
		count++
	}
}

// ClearCache clears all cached tiles (useful when switching tilesets)
func (itm *ImageTileManager) ClearCache() {
	itm.mutex.Lock()
	defer itm.mutex.Unlock()
	
	itm.tileCache = make(map[string]image.Image)
	itm.coloredCache = make(map[string]map[gruid.Color]image.Image)
}

// UpdateConfig updates the tile manager configuration
func (itm *ImageTileManager) UpdateConfig(newConfig *config.TileConfig) {
	itm.mutex.Lock()
	defer itm.mutex.Unlock()
	
	itm.config = newConfig
	itm.tileMapping = NewTileMapping(newConfig.TilesetPath)
	
	// Clear cache if tileset path changed
	itm.ClearCache()
}

// GetTileMapping returns the tile mapping for external use
func (itm *ImageTileManager) GetTileMapping() *TileMapping {
	return itm.tileMapping
}

// String returns a string representation for debugging
func (itm *ImageTileManager) String() string {
	itm.mutex.RLock()
	defer itm.mutex.RUnlock()
	
	return fmt.Sprintf("ImageTileManager{CachedTiles: %d, ColoredTiles: %d, Enabled: %v}", 
		len(itm.tileCache), len(itm.coloredCache), itm.config.Enabled)
}
