package ecs

import (
	"testing"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

func TestECS_AddEntity(t *testing.T) {
	ecs := NewECS()
	
	id := ecs.AddEntity()
	if id == 0 {
		t.Error("Expected non-zero entity ID")
	}
	
	if !ecs.EntityExists(id) {
		t.Error("Entity should exist after creation")
	}
}

func TestECS_RemoveEntity(t *testing.T) {
	ecs := NewECS()
	
	id := ecs.AddEntity()
	ecs.AddComponent(id, components.CPosition, gruid.Point{X: 5, Y: 5})
	
	ecs.RemoveEntity(id)
	
	if ecs.EntityExists(id) {
		t.Error("Entity should not exist after removal")
	}
	
	if ecs.HasComponent(id, components.CPosition) {
		t.Error("Component should be removed with entity")
	}
}

func TestECS_AddComponent(t *testing.T) {
	ecs := NewECS()
	id := ecs.AddEntity()
	
	pos := gruid.Point{X: 10, Y: 20}
	ecs.AddComponent(id, components.CPosition, pos)
	
	if !ecs.HasComponent(id, components.CPosition) {
		t.Error("Entity should have position component")
	}
	
	retrievedPos, ok := ecs.GetPosition(id)
	if !ok {
		t.Error("Should be able to retrieve position component")
	}
	
	if retrievedPos != pos {
		t.Errorf("Expected position %v, got %v", pos, retrievedPos)
	}
}

func TestECS_GetEntitiesWithComponent(t *testing.T) {
	ecs := NewECS()
	
	// Create entities with and without AI components
	aiEntity1 := ecs.AddEntity()
	aiEntity2 := ecs.AddEntity()
	nonAiEntity := ecs.AddEntity()
	
	ecs.AddComponent(aiEntity1, components.CAITag, components.AITag{})
	ecs.AddComponent(aiEntity2, components.CAITag, components.AITag{})
	ecs.AddComponent(nonAiEntity, components.CPosition, gruid.Point{X: 0, Y: 0})
	
	aiEntities := ecs.GetEntitiesWithComponent(components.CAITag)
	
	if len(aiEntities) != 2 {
		t.Errorf("Expected 2 AI entities, got %d", len(aiEntities))
	}
	
	// Check that the correct entities are returned
	found1, found2 := false, false
	for _, id := range aiEntities {
		if id == aiEntity1 {
			found1 = true
		}
		if id == aiEntity2 {
			found2 = true
		}
		if id == nonAiEntity {
			t.Error("Non-AI entity should not be in AI entities list")
		}
	}
	
	if !found1 || !found2 {
		t.Error("All AI entities should be found")
	}
}

func BenchmarkECS_AddEntity(b *testing.B) {
	ecs := NewECS()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ecs.AddEntity()
	}
}

func BenchmarkECS_AddComponent(b *testing.B) {
	ecs := NewECS()
	entities := make([]EntityID, 1000)
	
	for i := range entities {
		entities[i] = ecs.AddEntity()
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entityID := entities[i%len(entities)]
		ecs.AddComponent(entityID, components.CPosition, gruid.Point{X: i, Y: i})
	}
}

func BenchmarkECS_GetEntitiesWithComponent(b *testing.B) {
	ecs := NewECS()
	
	// Create 1000 entities, half with AI components
	for i := 0; i < 1000; i++ {
		id := ecs.AddEntity()
		if i%2 == 0 {
			ecs.AddComponent(id, components.CAITag, components.AITag{})
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ecs.GetEntitiesWithComponent(components.CAITag)
	}
}
