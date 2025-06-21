package turn

import "github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs"

// TurnEntry represents an item in the turn queue with time (priority) and entity ID
type TurnEntry struct {
	Time     uint64
	EntityID ecs.EntityID
}

// turnHeap implements a min-heap where the smallest time values are at the top
type turnHeap []TurnEntry

func (h turnHeap) Len() int { return len(h) }

func (h turnHeap) Less(i, j int) bool {
	if h[i].Time != h[j].Time {
		return h[i].Time < h[j].Time
	}
	return h[i].EntityID < h[j].EntityID
}

func (h turnHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }

func (h *turnHeap) Push(x any) {
	*h = append(*h, x.(TurnEntry))
}

func (h *turnHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func (h *turnHeap) FindIndex(entityID ecs.EntityID) int {
	for i, entry := range *h {
		if entry.EntityID == entityID {
			return i
		}
	}

	return -1
}
