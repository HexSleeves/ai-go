package utils

import (
	"slices"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/config"
)

func VisionRange(p gruid.Point, radius int) gruid.Range {
	drg := gruid.NewRange(0, 0, config.DungeonWidth, config.DungeonHeight)
	delta := gruid.Point{X: radius, Y: radius}
	return drg.Intersect(gruid.Range{Min: p.Sub(delta), Max: p.Add(delta).Shift(1, 1)})
}

// DrawFilledCircle calculates all grid cells within a given radius from a center point.
// It uses the distance check method: a cell (x, y) is included if
// (x - centerX)^2 + (y - centerY)^2 <= radius^2.
// This function returns a slice of Point structs representing the cells within the circle.
func DrawFilledCircle(visibles []gruid.Point, radius int, playerPosition gruid.Point) []gruid.Point {
	// Ensure radius is non-negative
	if radius < 0 {
		radius = 0
	}

	var visibleInCircle []gruid.Point
	radiusSq := radius * radius // Calculate radius squared for comparison

	// Define the bounding box around the circle
	minX := playerPosition.X - radius
	maxX := playerPosition.X + radius
	minY := playerPosition.Y - radius
	maxY := playerPosition.Y + radius

	// Iterate through each cell in the bounding box
	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			// Calculate the distance squared from the center to the current cell
			dx := x - playerPosition.X
			dy := y - playerPosition.Y
			distSq := dx*dx + dy*dy

			// If the distance squared is less than or equal to the radius squared,
			// the cell is inside or on the boundary of the circle.
			if distSq <= radiusSq && slices.Contains(visibles, gruid.Point{X: x, Y: y}) {
				visibleInCircle = append(visibleInCircle, gruid.Point{X: x, Y: y})
			}
		}
	}

	return visibleInCircle
}
