package components

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/rl"
)

// FOV holds data related to an entity's field of view.
type FOV struct {
	Range   int
	Visible []uint64
	fov     *rl.FOV
}

// NewFOVComponent creates and initializes a new FOV component.
// It requires the map dimensions to correctly size the internal bitset and FOV calculator.
func NewFOVComponent(fovRange, mapWidth, mapHeight int) *FOV {
	bitsetSize := (mapWidth*mapHeight + 63) / 64
	mapRange := gruid.NewRange(0, 0, mapWidth, mapHeight)

	return &FOV{
		Range:   fovRange,
		Visible: make([]uint64, bitsetSize),
		fov:     rl.NewFOV(mapRange),
	}
}

// IsVisible checks if a point is currently visible according to this component's bitset.
// Assumes mapWidth is passed correctly or accessible.
func (f *FOV) IsVisible(p gruid.Point, mapWidth int) bool {
	// Basic bounds check (optional, as FOV calc should handle it, but safer)
	if p.X < 0 || p.Y < 0 || p.X >= mapWidth { // Need mapHeight too ideally, but width is key for index
		return false
	}
	idx := p.Y*mapWidth + p.X
	if idx < 0 { // Should not happen with proper bounds check
		return false
	}
	sliceIdx := idx / 64
	bitIdx := uint(idx % 64)

	// Check if sliceIdx is within the bounds of the Visible slice
	if sliceIdx >= len(f.Visible) {
		return false
	}

	return (f.Visible[sliceIdx] & (1 << bitIdx)) != 0
}

// GetFOVCalculator returns the internal rl.FOV instance for calculations.
func (f *FOV) GetFOVCalculator() *rl.FOV {
	return f.fov
}

// ClearVisible resets the visible bitset.
func (f *FOV) ClearVisible() {
	for i := range f.Visible {
		f.Visible[i] = 0
	}
}

// SetVisible marks a point as visible in this component's bitset.
// Assumes mapWidth is passed correctly or accessible.
func (f *FOV) SetVisible(p gruid.Point, mapWidth int) {
	// Basic bounds check
	if p.X < 0 || p.Y < 0 || p.X >= mapWidth { // Need mapHeight too ideally
		return
	}
	idx := p.Y*mapWidth + p.X
	if idx < 0 {
		return
	}
	sliceIdx := idx / 64
	bitIdx := uint(idx % 64)

	// Check if sliceIdx is within the bounds of the Visible slice
	if sliceIdx < len(f.Visible) {
		f.Visible[sliceIdx] |= (1 << bitIdx)
	}
}

// GetVisiblePoints returns a slice of all points that are currently visible
func (f *FOV) GetVisiblePoints(mapWidth int) []gruid.Point {
	points := make([]gruid.Point, 0)

	// Calculate map height based on visible slice size and width
	mapHeight := (len(f.Visible) * 64) / mapWidth

	// Iterate through all possible positions
	for y := range mapHeight {
		for x := range mapWidth {
			p := gruid.Point{X: x, Y: y}
			if f.IsVisible(p, mapWidth) {
				points = append(points, p)
			}
		}
	}

	return points
}
