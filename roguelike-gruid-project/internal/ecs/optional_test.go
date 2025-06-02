package ecs

import (
	"testing"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

func TestOption_Some(t *testing.T) {
	opt := Some(42)

	if !opt.IsSome() {
		t.Error("Expected IsSome() to return true")
	}

	if opt.IsNone() {
		t.Error("Expected IsNone() to return false")
	}

	value := opt.Unwrap()
	if value != 42 {
		t.Errorf("Expected value 42, got %d", value)
	}
}

func TestOption_None(t *testing.T) {
	opt := None[int]()

	if opt.IsSome() {
		t.Error("Expected IsSome() to return false")
	}

	if !opt.IsNone() {
		t.Error("Expected IsNone() to return true")
	}
}

func TestOption_UnwrapPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected Unwrap() to panic on None value")
		}
	}()

	opt := None[int]()
	opt.Unwrap()
}

func TestOption_UnwrapOr(t *testing.T) {
	// Test with Some value
	opt := Some(42)
	value := opt.UnwrapOr(100)
	if value != 42 {
		t.Errorf("Expected value 42, got %d", value)
	}

	// Test with None value
	opt = None[int]()
	value = opt.UnwrapOr(100)
	if value != 100 {
		t.Errorf("Expected default value 100, got %d", value)
	}
}

func TestOption_Map(t *testing.T) {
	// Test mapping Some value
	opt := Some(5)
	mapped := Map(opt, func(x int) int { return x * 2 })

	if !mapped.IsSome() {
		t.Error("Expected mapped option to be Some")
	}

	if mapped.Unwrap() != 10 {
		t.Errorf("Expected mapped value 10, got %d", mapped.Unwrap())
	}

	// Test mapping None value
	opt = None[int]()
	mapped = Map(opt, func(x int) int { return x * 2 })

	if !mapped.IsNone() {
		t.Error("Expected mapped option to be None")
	}
}

func TestGetPositionOpt(t *testing.T) {
	ecs := NewECS()
	entityID := ecs.AddEntity()

	// Test with no position component
	posOpt := ecs.GetPositionOpt(entityID)
	if !posOpt.IsNone() {
		t.Error("Expected None for missing position component")
	}

	// Test with position component
	testPos := gruid.Point{X: 10, Y: 20}
	ecs.AddComponent(entityID, components.CPosition, testPos)
	posOpt = ecs.GetPositionOpt(entityID)

	if !posOpt.IsSome() {
		t.Error("Expected Some for existing position component")
	}

	pos := posOpt.Unwrap()
	if pos != testPos {
		t.Errorf("Expected position %v, got %v", testPos, pos)
	}
}

func TestGetHealthOpt(t *testing.T) {
	ecs := NewECS()
	entityID := ecs.AddEntity()

	// Test with no health component
	healthOpt := ecs.GetHealthOpt(entityID)
	if !healthOpt.IsNone() {
		t.Error("Expected None for missing health component")
	}

	// Test with health component
	testHealth := components.NewHealth(100)
	ecs.AddComponent(entityID, components.CHealth, testHealth)
	healthOpt = ecs.GetHealthOpt(entityID)

	if !healthOpt.IsSome() {
		t.Error("Expected Some for existing health component")
	}

	health := healthOpt.Unwrap()
	if health != testHealth {
		t.Errorf("Expected health %v, got %v", testHealth, health)
	}
}

func TestGetRenderableOpt(t *testing.T) {
	ecs := NewECS()
	entityID := ecs.AddEntity()

	// Test with no renderable component
	renderableOpt := ecs.GetRenderableOpt(entityID)
	if !renderableOpt.IsNone() {
		t.Error("Expected None for missing renderable component")
	}

	// Test with renderable component
	testRenderable := components.Renderable{Glyph: '@', Color: gruid.ColorDefault}
	ecs.AddComponent(entityID, components.CRenderable, testRenderable)
	renderableOpt = ecs.GetRenderableOpt(entityID)

	if !renderableOpt.IsSome() {
		t.Error("Expected Some for existing renderable component")
	}

	renderable := renderableOpt.Unwrap()
	if renderable != testRenderable {
		t.Errorf("Expected renderable %v, got %v", testRenderable, renderable)
	}
}

func TestGetNameOpt(t *testing.T) {
	ecs := NewECS()
	entityID := ecs.AddEntity()

	// Test with no name component
	nameOpt := ecs.GetNameOpt(entityID)
	if !nameOpt.IsNone() {
		t.Error("Expected None for missing name component")
	}

	// Test with name component
	testName := "Test Entity"
	ecs.AddComponent(entityID, components.CName, components.Name{Name: testName})
	nameOpt = ecs.GetNameOpt(entityID)

	if !nameOpt.IsSome() {
		t.Error("Expected Some for existing name component")
	}

	name := nameOpt.Unwrap()
	if name != testName {
		t.Errorf("Expected name %q, got %q", testName, name)
	}
}

func TestGetFOVOpt(t *testing.T) {
	ecs := NewECS()
	entityID := ecs.AddEntity()

	// Test with no FOV component
	fovOpt := ecs.GetFOVOpt(entityID)
	if !fovOpt.IsNone() {
		t.Error("Expected None for missing FOV component")
	}

	// Test with FOV component
	testFOV := components.NewFOVComponent(5, 80, 24)
	ecs.AddComponent(entityID, components.CFOV, testFOV)
	fovOpt = ecs.GetFOVOpt(entityID)

	if !fovOpt.IsSome() {
		t.Error("Expected Some for existing FOV component")
	}

	fov := fovOpt.Unwrap()
	if fov != testFOV {
		t.Errorf("Expected FOV %v, got %v", testFOV, fov)
	}
}

func TestOptionalChaining(t *testing.T) {
	ecs := NewECS()
	entityID := ecs.AddEntity()

	// Test chaining with UnwrapOr
	pos := ecs.GetPositionOpt(entityID).UnwrapOr(gruid.Point{X: 0, Y: 0})
	expected := gruid.Point{X: 0, Y: 0}
	if pos != expected {
		t.Errorf("Expected default position %v, got %v", expected, pos)
	}

	// Add position and test again
	testPos := gruid.Point{X: 10, Y: 20}
	ecs.AddComponent(entityID, components.CPosition, testPos)
	pos = ecs.GetPositionOpt(entityID).UnwrapOr(gruid.Point{X: 0, Y: 0})
	if pos != testPos {
		t.Errorf("Expected actual position %v, got %v", testPos, pos)
	}
}

func TestOptionalWithNonExistentEntity(t *testing.T) {
	ecs := NewECS()
	nonExistentID := EntityID(999)

	// All optional accessors should return None for non-existent entities
	posOpt := ecs.GetPositionOpt(nonExistentID)
	if !posOpt.IsNone() {
		t.Error("Expected None for non-existent entity position")
	}

	healthOpt := ecs.GetHealthOpt(nonExistentID)
	if !healthOpt.IsNone() {
		t.Error("Expected None for non-existent entity health")
	}

	nameOpt := ecs.GetNameOpt(nonExistentID)
	if !nameOpt.IsNone() {
		t.Error("Expected None for non-existent entity name")
	}

	fovOpt := ecs.GetFOVOpt(nonExistentID)
	if !fovOpt.IsNone() {
		t.Error("Expected None for non-existent entity FOV")
	}
}

// Benchmark tests for optional pattern
func BenchmarkGetPositionOpt(b *testing.B) {
	ecs := NewECS()
	entityID := ecs.AddEntity()
	ecs.AddComponent(entityID, components.CPosition, gruid.Point{X: 10, Y: 20})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ecs.GetPositionOpt(entityID)
	}
}

func BenchmarkOptionUnwrapOr(b *testing.B) {
	opt := Some(gruid.Point{X: 10, Y: 20})
	defaultPos := gruid.Point{X: 0, Y: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = opt.UnwrapOr(defaultPos)
	}
}

func BenchmarkOptionMap(b *testing.B) {
	opt := Some(5)
	mapFunc := func(x int) int { return x * 2 }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Map(opt, mapFunc)
	}
}
