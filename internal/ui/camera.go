package ui

import (
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
)

// Camera represents the viewport for the game map
type Camera struct {
	X, Y int // Camera position (top-left corner of viewport)
}

// NewCamera creates a new camera centered on the given position
func NewCamera(centerX, centerY int) *Camera {
	camera := &Camera{}
	camera.CenterOn(centerX, centerY)
	return camera
}

// CenterOn centers the camera on the given world coordinates
func (c *Camera) CenterOn(worldX, worldY int) {
	// Center the viewport on the target position
	c.X = worldX - config.MapViewportWidth/2
	c.Y = worldY - config.MapViewportHeight/2

	// Clamp camera to map boundaries
	c.clampToMapBounds()
}

// clampToMapBounds ensures the camera doesn't show areas outside the map
func (c *Camera) clampToMapBounds() {
	// Clamp X coordinate
	if c.X < 0 {
		c.X = 0
	}
	if c.X > config.DungeonWidth-config.MapViewportWidth {
		c.X = config.DungeonWidth - config.MapViewportWidth
	}

	// Clamp Y coordinate
	if c.Y < 0 {
		c.Y = 0
	}
	if c.Y > config.DungeonHeight-config.MapViewportHeight {
		c.Y = config.DungeonHeight - config.MapViewportHeight
	}
}

// WorldToScreen converts world coordinates to screen coordinates
func (c *Camera) WorldToScreen(worldX, worldY int) (screenX, screenY int, visible bool) {
	screenX = worldX - c.X + config.MapViewportX
	screenY = worldY - c.Y + config.MapViewportY

	// Check if the position is visible in the viewport
	visible = screenX >= config.MapViewportX &&
		screenX < config.MapViewportX+config.MapViewportWidth &&
		screenY >= config.MapViewportY &&
		screenY < config.MapViewportY+config.MapViewportHeight

	return screenX, screenY, visible
}

// ScreenToWorld converts screen coordinates to world coordinates
func (c *Camera) ScreenToWorld(screenX, screenY int) (worldX, worldY int) {
	worldX = screenX - config.MapViewportX + c.X
	worldY = screenY - config.MapViewportY + c.Y
	return worldX, worldY
}

// GetViewportBounds returns the world coordinates of the viewport bounds
func (c *Camera) GetViewportBounds() (minX, minY, maxX, maxY int) {
	minX = c.X
	minY = c.Y
	maxX = c.X + config.MapViewportWidth - 1
	maxY = c.Y + config.MapViewportHeight - 1
	return minX, minY, maxX, maxY
}

// IsInViewport checks if a world position is visible in the current viewport
func (c *Camera) IsInViewport(worldX, worldY int) bool {
	minX, minY, maxX, maxY := c.GetViewportBounds()
	return worldX >= minX && worldX <= maxX && worldY >= minY && worldY <= maxY
}

// Update updates the camera position to follow a target (usually the player)
func (c *Camera) Update(targetX, targetY int) {
	// Only move camera if target is near the edge of the viewport
	const scrollMargin = 5 // How close to edge before camera starts following

	currentCenterX := c.X + config.MapViewportWidth/2
	currentCenterY := c.Y + config.MapViewportHeight/2

	// Calculate distance from target to viewport center
	deltaX := targetX - currentCenterX
	deltaY := targetY - currentCenterY

	// Check if we need to scroll horizontally
	if deltaX > scrollMargin {
		c.X += deltaX - scrollMargin
	} else if deltaX < -scrollMargin {
		c.X += deltaX + scrollMargin
	}

	// Check if we need to scroll vertically
	if deltaY > scrollMargin {
		c.Y += deltaY - scrollMargin
	} else if deltaY < -scrollMargin {
		c.Y += deltaY + scrollMargin
	}

	// Ensure camera stays within bounds
	c.clampToMapBounds()
}
