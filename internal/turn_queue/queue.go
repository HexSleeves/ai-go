package turn

import (
	"container/heap"
	"fmt"
	"time"

	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
	"github.com/sirupsen/logrus"
)

// TurnQueue manages entity turns based on time using a min-heap.
type TurnQueue struct {
	queue                  *turnHeap
	CurrentTime            uint64
	OperationsSinceCleanup uint32
	TotalCleanups          uint64
	TotalEntitiesRemoved   uint64
}

func NewTurnQueue() *TurnQueue {
	h := &turnHeap{}
	heap.Init(h)

	return &TurnQueue{
		CurrentTime:            0,
		queue:                  h,
		OperationsSinceCleanup: 0,
		TotalCleanups:          0,
		TotalEntitiesRemoved:   0,
	}
}

func (tq *TurnQueue) Add(entityID ecs.EntityID, time uint64) {
	entry := TurnEntry{Time: time, EntityID: entityID}
	heap.Push(tq.queue, entry)
	logrus.Debugf("Added entity %d to turn queue with time %d", entityID, time)
}

func (tq *TurnQueue) Remove(entityID ecs.EntityID) {
	index := tq.queue.FindIndex(entityID)
	if index == -1 {
		logrus.Debugf("TurnQueue: Entity %d not found in queue", entityID)
		return
	}

	heap.Remove(tq.queue, index)
}

// Next removes and returns the next entity (the one with the smallest time)
// from the queue. Returns the entry and true if the queue is not empty,
// otherwise returns a zero TurnEntry and false.
func (tq *TurnQueue) Next() (TurnEntry, bool) {
	if tq.queue.Len() == 0 {
		return TurnEntry{}, false
	}

	entry := heap.Pop(tq.queue).(TurnEntry)
	logrus.Debugf("Popped entity %d from turn queue (time: %d)", entry.EntityID, entry.Time)
	return entry, true
}

// Peek returns the next entity without removing it from the queue.
// Returns the entry and true if the queue is not empty,
// otherwise returns a zero TurnEntry and false.
func (tq *TurnQueue) Peek() (TurnEntry, bool) {
	if tq.queue.Len() == 0 {
		return TurnEntry{}, false
	}

	entry := (*tq.queue)[0]
	logrus.Debugf("Peeked entity %d from turn queue (time: %d)", entry.EntityID, entry.Time)
	return entry, true
}

func (tq *TurnQueue) Len() int {
	return tq.queue.Len()
}

func (tq *TurnQueue) IsEmpty() bool {
	return tq.queue.Len() == 0
}

// Snapshot returns a copy of all entries in the turn queue for saving
func (tq *TurnQueue) Snapshot() []TurnEntry {
	entries := make([]TurnEntry, len(*tq.queue))
	copy(entries, *tq.queue)
	return entries
}

// RestoreFromSnapshot restores the turn queue from a saved snapshot
func (tq *TurnQueue) RestoreFromSnapshot(entries []TurnEntry) {
	// Clear the current queue
	tq.queue = &turnHeap{}
	heap.Init(tq.queue)

	// Add all entries back to the queue
	for _, entry := range entries {
		heap.Push(tq.queue, entry)
	}
}

// PrintQueue prints the current state of the turn queue for debugging purposes.
func (tq *TurnQueue) PrintQueue() {
	if tq.IsEmpty() {
		logrus.Debug("---- Turn Queue: EMPTY ----")
		return
	}

	logrus.Debug("---- Turn Queue Contents ----")
	logrus.Debugf("Current Game Time: %d\n", tq.CurrentTime)
	logrus.Debugf("Queue Size: %d\n", tq.Len())

	logrus.Debug("Queue (in heap order):")
	for i, entry := range *tq.queue {
		delta := int64(entry.Time) - int64(tq.CurrentTime)
		logrus.Debugf("[%d] EntityID: %d, Time: %d (Δ%d from current)\n",
			i, entry.EntityID, entry.Time, delta)
	}

	logrus.Debug("\nProcessing order (sorted by time):")
	sorted := make([]TurnEntry, len(*tq.queue))
	copy(sorted, *tq.queue)

	tq.sortEntriesByTime(sorted)

	for i, entry := range sorted {
		delta := int64(entry.Time) - int64(tq.CurrentTime)
		logrus.Debugf("%d. EntityID: %d, Time: %d (Δ%d from current)\n",
			i+1, entry.EntityID, entry.Time, delta)
	}

	logrus.Debug("----------------------------")
}

// sortEntriesByTime sorts a slice of TurnEntry by time, then by EntityID for stable ordering
func (tq *TurnQueue) sortEntriesByTime(entries []TurnEntry) {
	for i := 1; i < len(entries); i++ {
		for j := i; j > 0; j-- {
			if entries[j].Time < entries[j-1].Time {
				entries[j], entries[j-1] = entries[j-1], entries[j]
			} else if entries[j].Time == entries[j-1].Time &&
				entries[j].EntityID < entries[j-1].EntityID {
				entries[j], entries[j-1] = entries[j-1], entries[j]
			} else {
				break
			}
		}
	}
}

type CleanupMetrics struct {
	EntitiesRemoved int
	QueueSizeBefore int
	QueueSizeAfter  int
	ProcessingTime  time.Duration
}

func (m CleanupMetrics) String() string {
	return fmt.Sprintf(
		"CleanupMetrics{Removed: %d, Before: %d, After: %d, Time: %v}",
		m.EntitiesRemoved,
		m.QueueSizeBefore,
		m.QueueSizeAfter,
		m.ProcessingTime,
	)
}

// isValIDTurnActor checks if an entity is valid to remain in the turn queue
func (tq *TurnQueue) isValIDTurnActor(world *ecs.ECS, entityID ecs.EntityID) bool {
	if !world.EntityExists(entityID) {
		return false
	}

	if !world.HasComponent(entityID, components.CTurnActor) {
		return false
	}

	if world.HasComponent(entityID, components.CCorpseTag) {
		return false
	}

	if health, found := world.GetHealth(entityID); found {
		if health.IsDead() {
			return false
		}
	} else {
		logrus.Errorf("TurnQueue: Entity %d has no Health component", entityID)
		return false
	}

	return true
}

func (tq *TurnQueue) getCleanupThreshold(world *ecs.ECS) uint32 {
	base_threshold := 100

	entityCount := len(world.GetAllEntities())
	queueSize := tq.Len()

	if entityCount > 1000 || queueSize > 500 {
		return uint32(base_threshold / 2)
	} else if entityCount < 100 && queueSize < 50 {
		return uint32(base_threshold * 2)
	}

	return uint32(base_threshold)
}

// CleanupDeadEntities removes invalid or dead entities from the queue
func (tq *TurnQueue) CleanupDeadEntities(world *ecs.ECS) CleanupMetrics {
	threshold := tq.getCleanupThreshold(world)
	if tq.OperationsSinceCleanup < threshold {
		tq.OperationsSinceCleanup++
		return CleanupMetrics{}
	}

	logrus.Debug("TurnQueue: Cleaning up dead entities...")

	queueSizeBefore := tq.Len()
	startTime := time.Now()

	newQueueSlice := make(turnHeap, 0, queueSizeBefore)
	removedCount := 0

	originalQueue := tq.queue
	for originalQueue.Len() > 0 {
		entry := heap.Pop(originalQueue).(TurnEntry)

		entityValid := tq.isValIDTurnActor(world, entry.EntityID)

		if entityValid {
			newQueueSlice = append(newQueueSlice, entry)
		} else {
			removedCount++

			name := world.GetNameSafe(entry.EntityID)
			if name == "" {
				name = "Unknown"
			}

			logrus.Debugf("TurnQueue: Removed dead entity from turn queue: %s\n",
				name)
		}
	}

	tq.queue = &newQueueSlice

	heap.Init(tq.queue)

	tq.OperationsSinceCleanup = 0
	tq.TotalCleanups++
	tq.TotalEntitiesRemoved += uint64(removedCount)

	metrics := CleanupMetrics{
		EntitiesRemoved: removedCount,
		QueueSizeBefore: queueSizeBefore,
		QueueSizeAfter:  tq.Len(),
		ProcessingTime:  time.Since(startTime),
	}

	logrus.Debugf("TurnQueue: Cleanup finished. %s\n", metrics)
	return metrics
}
