package game

import (
	"codeberg.org/anaseto/gruid"
	"codeberg.org/anaseto/gruid/paths"
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
		for _, p := range fovCalculator.SSCVisionMap(pos, fov.Range, g.passable, false) {
			if paths.DistanceManhattan(p, pos) > fov.Range {
				continue
			}

			fov.SetVisible(p, g.dungeon.Width)

			if id == g.PlayerID && !g.dungeon.IsExplored(p) {
				g.dungeon.SetExplored(p)
			}
		}
	}
}
