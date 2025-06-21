package game

import (
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/sirupsen/logrus"
)

// processTurnQueue processes turns for actors until it's the player's turn
// or the queue is exhausted for the current time step.
func (md *Model) processTurnQueue() {
	g := md.game
	logrus.Debug("========= processTurnQueue started =========")

	// Periodically clean up the queue
	metrics := g.turnQueue.CleanupDeadEntities(g.ecs)
	if metrics.EntitiesRemoved > 10 {
		logrus.Infof(
			"Turn queue cleanup: removed %d entities in %v",
			metrics.EntitiesRemoved,
			metrics.ProcessingTime,
		)
	}

	g.turnQueue.PrintQueue()

	// Process turns until we need player input or run out of actors
	for i := range 100 { // Limit iterations to prevent infinite loops
		logrus.Debugf("Turn queue iteration %d", i)

		if g.turnQueue.IsEmpty() {
			logrus.Debug("Turn queue is empty.")
			logrus.Debug("========= processTurnQueue ended (queue empty) =========")
			return
		}

		turnEntry, ok := g.turnQueue.Next()
		if !ok {
			logrus.Debug("Error: Queue does not have any more actors.")
			logrus.Debug("========= processTurnQueue ended (Next error) =========")
			return
		}

		logrus.Debugf("Processing actor: EntityID=%d, Time=%d", turnEntry.EntityID, turnEntry.Time)
		actor := g.ecs.GetTurnActorSafe(turnEntry.EntityID)
		if !g.ecs.HasComponent(turnEntry.EntityID, components.CTurnActor) {
			logrus.Debugf("Error: Entity %d is not a valid actor.", turnEntry.EntityID)
			continue
		}

		if !actor.IsAlive() {
			logrus.Debugf("Entity %d is not alive, skipping turn.", turnEntry.EntityID)
			continue
		}

		isPlayer := turnEntry.EntityID == g.PlayerID
		action := actor.NextAction()

		if isPlayer && action == nil {
			g.waitingForInput = true
			logrus.Debug("It's the player's turn, waiting for input.")
			g.turnQueue.Add(turnEntry.EntityID, turnEntry.Time)
			logrus.Debug("========= processTurnQueue ended (player's turn) =========")
			return
		}

		if action == nil {
			logrus.Debugf("Entity %d has no actions, rescheduling turn at time %d", turnEntry.EntityID, turnEntry.Time)
			g.turnQueue.Add(turnEntry.EntityID, turnEntry.Time)
			continue
		}

		cost, err := action.(GameAction).Execute(g)
		if err != nil {
			logrus.Debugf("Failed to execute action for entity %d: %v", turnEntry.EntityID, err)

			// On failure, reschedule with appropriate delay
			if isPlayer {
				g.turnQueue.Add(turnEntry.EntityID, turnEntry.Time)
			} else {
				g.turnQueue.Add(turnEntry.EntityID, turnEntry.Time+100)
			}
			continue
		}

		g.FOVSystem()

		logrus.Debugf("Action executed for entity %d, cost: %d", turnEntry.EntityID, cost)

		// Update the game time and schedule next turn
		g.turnQueue.CurrentTime = turnEntry.Time + uint64(cost)
		g.turnQueue.Add(turnEntry.EntityID, g.turnQueue.CurrentTime)

		// Increment turn count for statistics
		g.IncrementTurnCount()
	}

	logrus.Debug("========= processTurnQueue ended (iteration limit reached) =========")
}
