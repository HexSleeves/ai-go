package ecs

import (
	"testing"

	"codeberg.org/anaseto/gruid"
	"github.com/lecoqjacob/ai-go/roguelike-gruid-project/internal/ecs/components"
)

func TestECS_UpdateComponent(t *testing.T) {
	ecs := NewECS()

	// Create entity with health component (simpler test)
	entityID := ecs.AddEntity()
	health := components.NewHealth(10)
	ecs.AddComponent(entityID, components.CHealth, health)

	// Test successful update - this is a basic test showing the concept
	// In practice, use type-safe methods like UpdateAIComponent
	err := ecs.UpdateComponent(entityID, components.CHealth, func(component interface{}) error {
		// Note: This is a demonstration - the generic UpdateComponent has limitations
		// with Go's type system. Use type-safe methods like UpdateAIComponent instead.
		t.Log("Generic UpdateComponent called - use type-safe methods in practice")
		return nil
	})

	if err != nil {
		t.Errorf("UpdateComponent should succeed, got error: %v", err)
	}
}

func TestECS_UpdateComponent_EntityNotExists(t *testing.T) {
	ecs := NewECS()
	
	// Try to update component on non-existent entity
	err := ecs.UpdateComponent(999, components.CAIComponent, func(component interface{}) error {
		return nil
	})
	
	if err == nil {
		t.Error("Expected error when updating component on non-existent entity")
	}
}

func TestECS_UpdateComponent_ComponentNotExists(t *testing.T) {
	ecs := NewECS()
	
	// Create entity without AI component
	entityID := ecs.AddEntity()
	
	// Try to update non-existent component
	err := ecs.UpdateComponent(entityID, components.CAIComponent, func(component interface{}) error {
		return nil
	})
	
	if err == nil {
		t.Error("Expected error when updating non-existent component")
	}
}

func TestECS_UpdateAIComponent(t *testing.T) {
	ecs := NewECS()
	
	// Create entity with AI component
	entityID := ecs.AddEntity()
	aiComp := components.NewAIComponent(components.AIBehaviorGuard, gruid.Point{X: 10, Y: 10})
	ecs.AddComponent(entityID, components.CAIComponent, aiComp)
	
	// Test type-safe AI component update
	err := ecs.UpdateAIComponent(entityID, func(ai *components.AIComponent) error {
		ai.PatrolRadius = 8
		ai.FleeThreshold = 0.25
		ai.State = components.AIStatePatrolling
		return nil
	})
	
	if err != nil {
		t.Errorf("UpdateAIComponent should succeed, got error: %v", err)
	}
	
	// Verify the update was applied
	updatedAI := ecs.GetAIComponentSafe(entityID)
	if updatedAI.PatrolRadius != 8 {
		t.Errorf("Expected PatrolRadius to be 8, got %d", updatedAI.PatrolRadius)
	}
	if updatedAI.FleeThreshold != 0.25 {
		t.Errorf("Expected FleeThreshold to be 0.25, got %f", updatedAI.FleeThreshold)
	}
	if updatedAI.State != components.AIStatePatrolling {
		t.Errorf("Expected State to be AIStatePatrolling, got %v", updatedAI.State)
	}
}

func TestECS_UpdateAIComponent_NoAIComponent(t *testing.T) {
	ecs := NewECS()
	
	// Create entity without AI component
	entityID := ecs.AddEntity()
	ecs.AddComponent(entityID, components.CHealth, components.NewHealth(10))
	
	// Try to update AI component that doesn't exist
	err := ecs.UpdateAIComponent(entityID, func(ai *components.AIComponent) error {
		ai.State = components.AIStateChasing
		return nil
	})
	
	if err == nil {
		t.Error("Expected error when updating AI component that doesn't exist")
	}
}

func TestECS_UpdateAIComponent_ConcurrentAccess(t *testing.T) {
	ecs := NewECS()
	
	// Create entity with AI component
	entityID := ecs.AddEntity()
	aiComp := components.NewAIComponent(components.AIBehaviorHunter, gruid.Point{X: 0, Y: 0})
	ecs.AddComponent(entityID, components.CAIComponent, aiComp)
	
	// Test that concurrent access is properly synchronized
	done := make(chan bool, 2)
	
	// First goroutine
	go func() {
		for i := 0; i < 100; i++ {
			ecs.UpdateAIComponent(entityID, func(ai *components.AIComponent) error {
				ai.SearchTurns++
				return nil
			})
		}
		done <- true
	}()
	
	// Second goroutine
	go func() {
		for i := 0; i < 100; i++ {
			ecs.UpdateAIComponent(entityID, func(ai *components.AIComponent) error {
				ai.SearchTurns++
				return nil
			})
		}
		done <- true
	}()
	
	// Wait for both goroutines to complete
	<-done
	<-done
	
	// Verify final state
	finalAI := ecs.GetAIComponentSafe(entityID)
	if finalAI.SearchTurns != 200 {
		t.Errorf("Expected SearchTurns to be 200, got %d", finalAI.SearchTurns)
	}
}

func BenchmarkECS_UpdateComponent(b *testing.B) {
	ecs := NewECS()
	entityID := ecs.AddEntity()
	aiComp := components.NewAIComponent(components.AIBehaviorHunter, gruid.Point{X: 0, Y: 0})
	ecs.AddComponent(entityID, components.CAIComponent, aiComp)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ecs.UpdateAIComponent(entityID, func(ai *components.AIComponent) error {
			ai.SearchTurns = i
			return nil
		})
	}
}

func BenchmarkECS_GetSetComponent(b *testing.B) {
	ecs := NewECS()
	entityID := ecs.AddEntity()
	aiComp := components.NewAIComponent(components.AIBehaviorHunter, gruid.Point{X: 0, Y: 0})
	ecs.AddComponent(entityID, components.CAIComponent, aiComp)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ai := ecs.GetAIComponentSafe(entityID)
		ai.SearchTurns = i
		ecs.AddComponent(entityID, components.CAIComponent, ai)
	}
}
