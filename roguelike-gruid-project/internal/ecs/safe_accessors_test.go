package ecs

import (
	"testing"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

func TestGetPositionSafe(t *testing.T) {
	ecs := NewECS()
	entityID := ecs.AddEntity()

	// Test with no position component
	pos := ecs.GetPositionSafe(entityID)
	expected := gruid.Point{}
	if pos != expected {
		t.Errorf("Expected zero Point %v, got %v", expected, pos)
	}

	// Test with position component
	testPos := gruid.Point{X: 10, Y: 20}
	ecs.AddComponent(entityID, components.CPosition, testPos)
	pos = ecs.GetPositionSafe(entityID)
	if pos != testPos {
		t.Errorf("Expected position %v, got %v", testPos, pos)
	}
}

func TestGetHealthSafe(t *testing.T) {
	ecs := NewECS()
	entityID := ecs.AddEntity()

	// Test with no health component
	health := ecs.GetHealthSafe(entityID)
	expected := components.Health{}
	if health != expected {
		t.Errorf("Expected zero Health %v, got %v", expected, health)
	}

	// Test with health component
	testHealth := components.NewHealth(100)
	ecs.AddComponent(entityID, components.CHealth, testHealth)
	health = ecs.GetHealthSafe(entityID)
	if health != testHealth {
		t.Errorf("Expected health %v, got %v", testHealth, health)
	}
}

func TestGetRenderableSafe(t *testing.T) {
	ecs := NewECS()
	entityID := ecs.AddEntity()

	// Test with no renderable component
	renderable := ecs.GetRenderableSafe(entityID)
	expected := components.Renderable{}
	if renderable != expected {
		t.Errorf("Expected zero Renderable %v, got %v", expected, renderable)
	}

	// Test with renderable component
	testRenderable := components.Renderable{Glyph: '@', Color: gruid.ColorDefault}
	ecs.AddComponent(entityID, components.CRenderable, testRenderable)
	renderable = ecs.GetRenderableSafe(entityID)
	if renderable != testRenderable {
		t.Errorf("Expected renderable %v, got %v", testRenderable, renderable)
	}
}

func TestGetNameSafe(t *testing.T) {
	ecs := NewECS()
	entityID := ecs.AddEntity()

	// Test with no name component
	name := ecs.GetNameSafe(entityID)
	if name != "" {
		t.Errorf("Expected empty string, got %q", name)
	}

	// Test with name component
	testName := "Test Entity"
	ecs.AddComponent(entityID, components.CName, components.Name{Name: testName})
	name = ecs.GetNameSafe(entityID)
	if name != testName {
		t.Errorf("Expected name %q, got %q", testName, name)
	}
}

func TestGetFOVSafe(t *testing.T) {
	ecs := NewECS()
	entityID := ecs.AddEntity()

	// Test with no FOV component
	fov := ecs.GetFOVSafe(entityID)
	if fov != nil {
		t.Errorf("Expected nil FOV, got %v", fov)
	}

	// Test with FOV component
	testFOV := components.NewFOVComponent(5, 80, 24)
	ecs.AddComponent(entityID, components.CFOV, testFOV)
	fov = ecs.GetFOVSafe(entityID)
	if fov != testFOV {
		t.Errorf("Expected FOV %v, got %v", testFOV, fov)
	}
}

func TestGetTurnActorSafe(t *testing.T) {
	ecs := NewECS()
	entityID := ecs.AddEntity()

	// Test with no TurnActor component
	actor := ecs.GetTurnActorSafe(entityID)
	expected := components.TurnActor{}
	if actor != expected {
		t.Errorf("Expected zero TurnActor %v, got %v", expected, actor)
	}

	// Test with TurnActor component
	testActor := components.NewTurnActor(100)
	ecs.AddComponent(entityID, components.CTurnActor, testActor)
	actor = ecs.GetTurnActorSafe(entityID)
	if actor != testActor {
		t.Errorf("Expected TurnActor %v, got %v", testActor, actor)
	}
}

func TestHasComponentSafeMethods(t *testing.T) {
	ecs := NewECS()
	entityID := ecs.AddEntity()

	// Test all Has*Safe methods with no components
	if ecs.HasPositionSafe(entityID) {
		t.Error("Expected HasPositionSafe to return false")
	}
	if ecs.HasHealthSafe(entityID) {
		t.Error("Expected HasHealthSafe to return false")
	}
	if ecs.HasRenderableSafe(entityID) {
		t.Error("Expected HasRenderableSafe to return false")
	}
	if ecs.HasFOVSafe(entityID) {
		t.Error("Expected HasFOVSafe to return false")
	}

	// Add components and test again
	ecs.AddComponent(entityID, components.CPosition, gruid.Point{X: 1, Y: 1})
	ecs.AddComponent(entityID, components.CHealth, components.NewHealth(10))
	ecs.AddComponent(entityID, components.CRenderable, components.Renderable{Glyph: '@'})
	ecs.AddComponent(entityID, components.CFOV, components.NewFOVComponent(5, 80, 24))

	if !ecs.HasPositionSafe(entityID) {
		t.Error("Expected HasPositionSafe to return true")
	}
	if !ecs.HasHealthSafe(entityID) {
		t.Error("Expected HasHealthSafe to return true")
	}
	if !ecs.HasRenderableSafe(entityID) {
		t.Error("Expected HasRenderableSafe to return true")
	}
	if !ecs.HasFOVSafe(entityID) {
		t.Error("Expected HasFOVSafe to return true")
	}
}

func TestSafeAccessorsWithNonExistentEntity(t *testing.T) {
	ecs := NewECS()
	nonExistentID := EntityID(999)

	// All safe accessors should return zero values for non-existent entities
	pos := ecs.GetPositionSafe(nonExistentID)
	if pos != (gruid.Point{}) {
		t.Errorf("Expected zero Point for non-existent entity, got %v", pos)
	}

	health := ecs.GetHealthSafe(nonExistentID)
	if health != (components.Health{}) {
		t.Errorf("Expected zero Health for non-existent entity, got %v", health)
	}

	name := ecs.GetNameSafe(nonExistentID)
	if name != "" {
		t.Errorf("Expected empty string for non-existent entity, got %q", name)
	}

	fov := ecs.GetFOVSafe(nonExistentID)
	if fov != nil {
		t.Errorf("Expected nil FOV for non-existent entity, got %v", fov)
	}
}

// Benchmark tests to ensure safe accessors don't introduce significant overhead
func BenchmarkGetPositionSafe(b *testing.B) {
	ecs := NewECS()
	entityID := ecs.AddEntity()
	ecs.AddComponent(entityID, components.CPosition, gruid.Point{X: 10, Y: 20})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ecs.GetPositionSafe(entityID)
	}
}

func BenchmarkGetPositionOriginal(b *testing.B) {
	ecs := NewECS()
	entityID := ecs.AddEntity()
	ecs.AddComponent(entityID, components.CPosition, gruid.Point{X: 10, Y: 20})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ecs.GetPosition(entityID)
	}
}

func BenchmarkGetHealthSafe(b *testing.B) {
	ecs := NewECS()
	entityID := ecs.AddEntity()
	ecs.AddComponent(entityID, components.CHealth, components.NewHealth(100))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ecs.GetHealthSafe(entityID)
	}
}

func BenchmarkGetHealthOriginal(b *testing.B) {
	ecs := NewECS()
	entityID := ecs.AddEntity()
	ecs.AddComponent(entityID, components.CHealth, components.NewHealth(100))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ecs.GetHealth(entityID)
	}
}
