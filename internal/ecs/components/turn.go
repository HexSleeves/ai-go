package components

import (
	"container/list"
)

// TurnActor represents an entity that takes turns in the game
type TurnActor struct {
	Speed        uint64
	Alive        bool
	NextTurnTime uint64
	actions      *list.List
}

// NewTurnActor creates a new TurnActor with the given speed
func NewTurnActor(speed uint64) TurnActor {
	return TurnActor{
		Speed:        speed,
		Alive:        true,
		NextTurnTime: 0,
		actions:      list.New(),
	}
}

// QueueAction adds an action to the back of the action queue
func (ta *TurnActor) QueueAction(action any) *TurnActor {
	ta.actions.PushBack(action)
	return ta
}

// AddAction adds an action to the back of the action queue
func (ta *TurnActor) AddAction(action any) {
	ta.actions.PushBack(action)
}

// NextAction removes and returns the next action from the queue
func (ta *TurnActor) NextAction() any {
	if ta.actions.Len() == 0 {
		return nil
	}
	element := ta.actions.Front()
	ta.actions.Remove(element)
	return element.Value
}

// PeekNextAction returns the next action from the queue without removing it
func (ta *TurnActor) PeekNextAction() any {
	if ta.actions.Len() == 0 {
		return nil
	}

	element := ta.actions.Front()
	return element.Value
}

// IsAlive returns whether the actor is alive
func (ta *TurnActor) IsAlive() bool {
	return ta.Alive
}
