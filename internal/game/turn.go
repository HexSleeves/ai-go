package game

import (
	"log/slog"

	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

// processTurnQueue processes turns for actors until it's the player's turn
// or the queue is exhausted for the current time step.
func (md *Model) processTurnQueue() {
	g := md.game
	slog.Debug("========= processTurnQueue started =========")

	// Periodically clean up the queue
	metrics := g.turnQueue.CleanupDeadEntities(g.ecs)
	if metrics.EntitiesRemoved > 10 {
		slog.Info("Turn queue cleanup", "removed", metrics.EntitiesRemoved, "time", metrics.ProcessingTime)
	}

	g.turnQueue.PrintQueue()

	// Process turns until we need player input or run out of actors
	for i := range 100 { // Limit iterations to prevent infinite loops
		slog.Debug("Turn queue iteration", "iteration", i)

		if g.turnQueue.IsEmpty() {
			slog.Debug("Turn queue is empty.")
			slog.Debug("========= processTurnQueue ended (queue empty) =========")
			return
		}

		turnEntry, ok := g.turnQueue.Next()
		if !ok {
			slog.Debug("Error: Queue does not have any more actors.")
			slog.Debug("========= processTurnQueue ended (Next error) =========")
			return
		}

		slog.Debug("Processing actor", "entityId", turnEntry.EntityID, "time", turnEntry.Time)
		actor := g.ecs.GetTurnActorSafe(turnEntry.EntityID)
		if !g.ecs.HasComponent(turnEntry.EntityID, components.CTurnActor) {
			slog.Debug("Error: Entity is not a valid actor.", "entityId", turnEntry.EntityID)
			continue
		}

		if !actor.IsAlive() {
			slog.Debug("Entity is not alive, skipping turn.", "entityId", turnEntry.EntityID)
			continue
		}

		isPlayer := turnEntry.EntityID == g.PlayerID
		action := actor.NextAction()

		if isPlayer && action == nil {
			g.waitingForInput = true
			slog.Debug("It's the player's turn, waiting for input.")
			g.turnQueue.Add(turnEntry.EntityID, turnEntry.Time)
			slog.Debug("========= processTurnQueue ended (player's turn) =========")
			return
		}

		if action == nil {
			slog.Debug("Entity has no actions, rescheduling turn", "entityId", turnEntry.EntityID, "time", turnEntry.Time)
			g.turnQueue.Add(turnEntry.EntityID, turnEntry.Time)
			continue
		}

		cost, err := action.(GameAction).Execute(g)
		if err != nil {
			slog.Debug("Failed to execute action", "entityId", turnEntry.EntityID, "error", err)

			// On failure, reschedule with appropriate delay
			if isPlayer {
				g.turnQueue.Add(turnEntry.EntityID, turnEntry.Time)
			} else {
				g.turnQueue.Add(turnEntry.EntityID, turnEntry.Time+100)
			}
			continue
		}

		g.FOVSystem()

		slog.Debug("Action executed", "entityId", turnEntry.EntityID, "cost", cost)

		// Update the game time and schedule next turn
		g.turnQueue.CurrentTime = turnEntry.Time + uint64(cost)
		g.turnQueue.Add(turnEntry.EntityID, g.turnQueue.CurrentTime)

		// Increment turn count for statistics
		g.IncrementTurnCount()
	}

	slog.Debug("========= processTurnQueue ended (iteration limit reached) =========")
}
