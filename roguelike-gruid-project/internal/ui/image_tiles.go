//go:build !js
// +build !js

package ui

import (
	"fmt"
	"image"
	"image/draw"
	"path/filepath"
	"sync"

	"codeberg.org/anaseto/gruid"
	sdl "codeberg.org/anaseto/gruid-sdl"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
	"github.com/sirupsen/logrus"
)

// ImageTileManager implements sdl.TileManager for image-based tiles
type ImageTileManager struct {
	tileCache    map[rune]image.Image
	coloredCache map[rune]map[gruid.Color]image.Image
	spriteAtlas  *KenneyRoguelikeAtlas // Sprite atlas for extracting tiles
	config       *config.TileConfig
	mutex        sync.RWMutex
	fontFallback sdl.TileManager // Fallback to font-based rendering
}

// NewImageTileManager creates a new image-based tile manager
func NewImageTileManager(config *config.TileConfig, fontFallback sdl.TileManager) *ImageTileManager {
	itm := &ImageTileManager{
		tileCache:    make(map[rune]image.Image),
		coloredCache: make(map[rune]map[gruid.Color]image.Image),
		config:       config,
		fontFallback: fontFallback,
	}

	// Try to load sprite atlas
	itm.loadSpriteAtlas()

	// Pre-load common tiles
	itm.preloadCommonTiles()

	return itm
}

// loadSpriteAtlas attempts to load the Kenney spritesheet
func (itm *ImageTileManager) loadSpriteAtlas() {
	// Try alternative common names for the spritesheet
	path := filepath.Join(itm.config.TilesetPath, "colored-transparent_packed.png")
	if atlas, err := NewKenneyRoguelikeAtlas(path); err == nil {
		itm.spriteAtlas = atlas
		logrus.Infof("Loaded sprite atlas from: %s", path)
		return
	}

	logrus.Warn("No sprite atlas found. Place Kenney's spritesheet as 'colored-transparent_packed.png' in the tileset directory")
}

// GetImage implements sdl.TileManager.GetImage
func (itm *ImageTileManager) GetImage(c gruid.Cell) image.Image {
	// If tiles are disabled, use font fallback
	if !itm.config.Enabled {
		if itm.fontFallback != nil {
			return itm.fontFallback.GetImage(c)
		}
		return itm.generateFallbackImage(c)
	}

	// If no sprite atlas is available, use font fallback
	if itm.spriteAtlas == nil {
		if itm.fontFallback != nil {
			return itm.fontFallback.GetImage(c)
		}
		return itm.generateFallbackImage(c)
	}

	// Check if this rune should use sprite rendering
	// Only use sprites for runes that are explicitly mapped in the atlas
	if !itm.shouldUseSprite(c.Rune) {
		// Use font fallback for text characters (letters, numbers, punctuation, etc.)
		if itm.fontFallback != nil {
			return itm.fontFallback.GetImage(c)
		}
		return itm.generateFallbackImage(c)
	}

	// Try to get colored version from cache
	if img := itm.getCachedColoredTile(c.Rune, c.Style.Fg, c.Style.Bg); img != nil {
		return img
	}

	// Load base tile from atlas
	baseImg := itm.loadTile(c.Rune)
	if baseImg == nil {
		// Fallback to font rendering if tile loading fails
		if itm.fontFallback != nil {
			return itm.fontFallback.GetImage(c)
		}
		return itm.generateFallbackImage(c)
	}

	// Apply colors and cache result
	coloredImg := itm.applyColors(baseImg, c.Style.Fg, c.Style.Bg, c.Style.Attrs)
	itm.cacheColoredTile(c.Rune, c.Style.Fg, coloredImg)

	return coloredImg
}

// shouldUseSprite determines if a rune should be rendered using sprites or fonts
func (itm *ImageTileManager) shouldUseSprite(r rune) bool {
	// Only use sprites for runes that are explicitly mapped in the sprite atlas
	// This ensures that UI text (letters, numbers, punctuation) uses font rendering
	// while game entities use sprite rendering
	_, exists := RuneToSpriteMapping[r]
	return exists
}

// TileSize implements sdl.TileManager.TileSize
func (itm *ImageTileManager) TileSize() gruid.Point {
	size := itm.config.TileSize
	scale := itm.config.ScaleFactor
	return gruid.Point{
		X: int(float32(size) * scale),
		Y: int(float32(size) * scale),
	}
}

// loadTile loads a tile image from the sprite atlas
func (itm *ImageTileManager) loadTile(r rune) image.Image {
	itm.mutex.RLock()
	if img, exists := itm.tileCache[r]; exists {
		itm.mutex.RUnlock()
		return img
	}
	itm.mutex.RUnlock()

	// Get tile from sprite atlas
	if itm.spriteAtlas == nil {
		return nil
	}

	img := itm.spriteAtlas.GetSpriteForRune(r)
	if img == nil {
		return nil
	}

	// Scale image if necessary
	scaledImg := itm.scaleImage(img)

	// Cache the loaded tile
	itm.mutex.Lock()
	itm.tileCache[r] = scaledImg

	// Implement simple LRU eviction if cache is too large
	if len(itm.tileCache) > itm.config.CacheSize {
		itm.evictOldestTiles()
	}
	itm.mutex.Unlock()

	return scaledImg
}

// getCachedColoredTile retrieves a colored tile from cache
func (itm *ImageTileManager) getCachedColoredTile(r rune, fg, bg gruid.Color) image.Image {
	itm.mutex.RLock()
	defer itm.mutex.RUnlock()

	if colorMap, exists := itm.coloredCache[r]; exists {
		if img, exists := colorMap[fg]; exists {
			return img
		}
	}
	return nil
}

// cacheColoredTile stores a colored tile in cache
func (itm *ImageTileManager) cacheColoredTile(r rune, fg gruid.Color, img image.Image) {
	itm.mutex.Lock()
	defer itm.mutex.Unlock()

	if itm.coloredCache[r] == nil {
		itm.coloredCache[r] = make(map[gruid.Color]image.Image)
	}
	itm.coloredCache[r][fg] = img
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
	logrus.Warn("Generating fallback image")

	size := itm.TileSize()
	img := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))

	fgColor := ColorToRGBA(c.Style.Fg, true)
	bgColor := ColorToRGBA(c.Style.Bg, false)

	// Fill with background color
	draw.Draw(img, img.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// Draw a simple representation of the character
	// This is a very basic fallback - in practice you might want something more sophisticated
	centerX, centerY := size.X/2, size.Y/2
	for y := centerY - 2; y <= centerY+2; y++ {
		for x := centerX - 2; x <= centerX+2; x++ {
			if x >= 0 && x < size.X && y >= 0 && y < size.Y {
				img.Set(x, y, fgColor)
			}
		}
	}

	return img
}

// preloadCommonTiles loads frequently used tiles into cache
func (itm *ImageTileManager) preloadCommonTiles() {
	commonRunes := []rune{'@', '#', '.', '+', '!', '?', 'o', 'g', 's', 'D', '/', '\\', ')', '[', ']', '$', '*', '%', '=', '"'}

	for _, r := range commonRunes {
		itm.loadTile(r) // This will cache the tile
	}
}

// evictOldestTiles removes some tiles from cache to free memory
func (itm *ImageTileManager) evictOldestTiles() {
	// Simple eviction: remove 25% of cached tiles
	// In a more sophisticated implementation, you'd use LRU
	targetSize := itm.config.CacheSize * 3 / 4

	count := 0
	for r := range itm.tileCache {
		if count >= len(itm.tileCache)-targetSize {
			break
		}
		delete(itm.tileCache, r)
		delete(itm.coloredCache, r)
		count++
	}
}

// ClearCache clears all cached tiles (useful when switching tilesets)
func (itm *ImageTileManager) ClearCache() {
	itm.mutex.Lock()
	defer itm.mutex.Unlock()

	itm.tileCache = make(map[rune]image.Image)
	itm.coloredCache = make(map[rune]map[gruid.Color]image.Image)
}

// UpdateConfig updates the tile manager configuration
func (itm *ImageTileManager) UpdateConfig(newConfig *config.TileConfig) {
	itm.mutex.Lock()
	defer itm.mutex.Unlock()

	oldPath := itm.config.TilesetPath
	itm.config = newConfig

	// Reload sprite atlas if tileset path changed
	if oldPath != newConfig.TilesetPath {
		itm.loadSpriteAtlas()
		itm.ClearCache()
	}
}

// String returns a string representation for debugging
func (itm *ImageTileManager) String() string {
	itm.mutex.RLock()
	defer itm.mutex.RUnlock()

	return fmt.Sprintf("ImageTileManager{CachedTiles: %d, ColoredTiles: %d, Enabled: %v}",
		len(itm.tileCache), len(itm.coloredCache), itm.config.Enabled)
}
