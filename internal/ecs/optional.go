package ecs

import (
	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

// Option represents an optional value that may or may not be present.
// This provides explicit null handling for component access.
type Option[T any] struct {
	value   T
	present bool
}

// Some creates an Option with a value present.
func Some[T any](value T) Option[T] {
	return Option[T]{value: value, present: true}
}

// None creates an Option with no value present.
func None[T any]() Option[T] {
	var zero T
	return Option[T]{value: zero, present: false}
}

// IsSome returns true if the Option contains a value.
func (o Option[T]) IsSome() bool {
	return o.present
}

// IsNone returns true if the Option contains no value.
func (o Option[T]) IsNone() bool {
	return !o.present
}

// Unwrap returns the contained value. Panics if no value is present.
// Use IsSome() to check before calling Unwrap().
func (o Option[T]) Unwrap() T {
	if !o.present {
		panic("called Unwrap() on None value")
	}
	return o.value
}

// UnwrapOr returns the contained value or the provided default if no value is present.
func (o Option[T]) UnwrapOr(defaultValue T) T {
	if o.present {
		return o.value
	}
	return defaultValue
}

// Map applies a function to the contained value if present, returning a new Option.
func Map[T, U any](o Option[T], f func(T) U) Option[U] {
	if o.present {
		return Some(f(o.value))
	}
	return None[U]()
}

// Optional accessor methods that return Option types for explicit null handling.

// GetPositionOpt returns an Option containing the position component for an entity.
func (ecs *ECS) GetPositionOpt(id EntityID) Option[gruid.Point] {
	pos, ok := ecs.GetPosition(id)
	if ok {
		return Some(pos)
	}
	return None[gruid.Point]()
}

// GetHealthOpt returns an Option containing the health component for an entity.
func (ecs *ECS) GetHealthOpt(id EntityID) Option[components.Health] {
	health, ok := ecs.GetHealth(id)
	if ok {
		return Some(health)
	}
	return None[components.Health]()
}

// GetRenderableOpt returns an Option containing the renderable component for an entity.
func (ecs *ECS) GetRenderableOpt(id EntityID) Option[components.Renderable] {
	renderable, ok := ecs.GetRenderable(id)
	if ok {
		return Some(renderable)
	}
	return None[components.Renderable]()
}

// GetNameOpt returns an Option containing the name component for an entity.
func (ecs *ECS) GetNameOpt(id EntityID) Option[string] {
	name, ok := ecs.GetName(id)
	if ok {
		return Some(name)
	}
	return None[string]()
}

// GetFOVOpt returns an Option containing the FOV component for an entity.
func (ecs *ECS) GetFOVOpt(id EntityID) Option[*components.FOV] {
	fov, ok := ecs.GetFOV(id)
	if ok {
		return Some(fov)
	}
	return None[*components.FOV]()
}

// GetTurnActorOpt returns an Option containing the TurnActor component for an entity.
func (ecs *ECS) GetTurnActorOpt(id EntityID) Option[components.TurnActor] {
	actor, ok := ecs.GetTurnActor(id)
	if ok {
		return Some(actor)
	}
	return None[components.TurnActor]()
}

// GetPlayerTagOpt returns an Option containing the PlayerTag component for an entity.
func (ecs *ECS) GetPlayerTagOpt(id EntityID) Option[components.PlayerTag] {
	tag, ok := ecs.GetPlayerTag(id)
	if ok {
		return Some(tag)
	}
	return None[components.PlayerTag]()
}

// GetAITagOpt returns an Option containing the AITag component for an entity.
func (ecs *ECS) GetAITagOpt(id EntityID) Option[components.AITag] {
	tag, ok := ecs.GetAITag(id)
	if ok {
		return Some(tag)
	}
	return None[components.AITag]()
}

// GetCorpseTagOpt returns an Option containing the CorpseTag component for an entity.
func (ecs *ECS) GetCorpseTagOpt(id EntityID) Option[components.CorpseTag] {
	tag, ok := ecs.GetCorpseTag(id)
	if ok {
		return Some(tag)
	}
	return None[components.CorpseTag]()
}
