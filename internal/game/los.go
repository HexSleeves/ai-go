package game

import (
	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/utils"
)

// Define the passable function once (reusing Map's IsOpaque)
func (g *Game) passable(p gruid.Point) bool {
	return !g.dungeon.IsOpaque(p)
}

// FOVSystem updates the visibility for all entities with an FOV component.
func (g *Game) FOVSystem() {
	entities := g.ecs.GetEntitiesWithPositionAndFOV()

	for _, entity := range entities {
		id, pos, fov := entity.ID, entity.Position, entity.FOV
		fov.ClearVisible()
		fovCalculator := fov.GetFOVCalculator()
		visibleTiles := fovCalculator.SSCVisionMap(pos, fov.Range, g.passable, false) // Use asserted pos
		visibleTiles = utils.DrawFilledCircle(visibleTiles, fov.Range, pos)           // Use asserted pos

		// Update visibility and explored status for points in the circle
		for _, p := range visibleTiles {
			fov.SetVisible(p, g.dungeon.Width)

			// If this is the player, also update the global explored map
			if id == g.PlayerID {
				g.dungeon.SetExplored(p)
			}
		}
	}
}
