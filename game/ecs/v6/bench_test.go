package ecs_test

import (
	"testing"

	"github.com/mokiat/lacking/game/ecs/v6"
)

// go test -test.fullpath=true -cpuprofile=cpu.pprof -memprofile=mem.pprof -benchmem -run=^$ -bench ^Benchmark github.com/mokiat/lacking/game/ecs/v6 -count=1

func BenchmarkCheckEntity(b *testing.B) {
	type Position struct {
		X, Y float64
	}
	type Name struct {
		Value string
	}
	type Age struct {
		Value int
	}

	scope := ecs.NewScope()
	positionType := ecs.Type[Position](scope)
	nameType := ecs.Type[Name](scope)
	ageType := ecs.Type[Age](scope)
	scene := ecs.NewScene(scope)

	id := scene.CreateEntity(nil)
	scene.EditEntity(id, func(op *ecs.EditOperation) {
		ecs.AddComponent(op, positionType, Position{
			X: 1.0,
			Y: 2.0,
		})
		ecs.AddComponent(op, nameType, Name{
			Value: "Alice",
		})
	})

	for b.Loop() {
		ok := scene.CheckEntity(id, ecs.Conditions(
			ecs.HasComponent(positionType),
			ecs.HasComponent(nameType),
			ecs.LacksComponent(ageType),
		))
		if !ok {
			b.Fatal("unexpected failed check")
		}
	}
}

func BenchmarkEditEntity(b *testing.B) {
	type Position struct {
		X, Y float64
	}
	type Name struct {
		Value string
	}
	type Age struct {
		Value int
	}

	scope := ecs.NewScope()
	positionType := ecs.Type[Position](scope)
	nameType := ecs.Type[Name](scope)
	ageType := ecs.Type[Age](scope)
	scene := ecs.NewScene(scope)

	id := scene.CreateEntity(nil)
	scene.EditEntity(id, func(op *ecs.EditOperation) {
		ecs.AddComponent(op, positionType, Position{
			X: 1.0,
			Y: 2.0,
		})
		ecs.AddComponent(op, ageType, Age{
			Value: 30,
		})
	})

	for b.Loop() {
		scene.EditEntity(id, func(op *ecs.EditOperation) {
			ecs.AddComponent(op, nameType, Name{
				Value: "Alice",
			})
		})
		scene.EditEntity(id, func(op *ecs.EditOperation) {
			ecs.RemoveComponent(op, nameType)
		})
	}
}

func BenchmarkQueryEntities(b *testing.B) {
	type Position struct {
		X, Y float64
	}
	type Velocity struct {
		X, Y float64
	}

	const entityCount = 1000

	scope := ecs.NewScope()
	positionType := ecs.Type[Position](scope)
	velocityType := ecs.Type[Velocity](scope)
	scene := ecs.NewScene(scope)

	for range entityCount {
		id := scene.CreateEntity(nil)
		scene.EditEntity(id, func(op *ecs.EditOperation) {
			ecs.AddComponent(op, positionType, Position{X: 1.0, Y: 2.0})
			ecs.AddComponent(op, velocityType, Velocity{X: 0.5, Y: 0.5})
		})
	}

	for b.Loop() {
		count := 0
		scene.QueryEntities(
			ecs.Conditions(
				ecs.HasComponent(positionType),
			),
			func(_ ecs.ID, op *ecs.ReadOperation) bool {
				pos := ecs.GetComponent(op, positionType)
				vel := ecs.GetComponent(op, velocityType)
				pos.X += vel.X
				pos.Y += vel.Y
				count++
				return true
			},
		)
		if count != entityCount {
			b.Fatalf("unexpected entity count: got %d, want %d", count, entityCount)
		}
	}
}

func BenchmarkQueryEntitiesMultiArchetype(b *testing.B) {
	type Position struct {
		X, Y float64
	}
	type Velocity struct {
		X, Y float64
	}
	type Tag struct{}

	const entityCount = 1000

	scope := ecs.NewScope()
	positionType := ecs.Type[Position](scope)
	velocityType := ecs.Type[Velocity](scope)
	tagType := ecs.Type[Tag](scope)
	scene := ecs.NewScene(scope)

	for i := range entityCount {
		id := scene.CreateEntity(nil)
		scene.EditEntity(id, func(op *ecs.EditOperation) {
			ecs.AddComponent(op, positionType, Position{X: 1.0, Y: 2.0})
			ecs.AddComponent(op, velocityType, Velocity{X: 0.5, Y: 0.5})
			if i%2 == 0 {
				ecs.AddComponent(op, tagType, Tag{})
			}
		})
	}

	for b.Loop() {
		count := 0
		scene.QueryEntities(
			ecs.Conditions(
				ecs.HasComponent(positionType),
				ecs.HasComponent(velocityType),
				ecs.HasComponent(tagType),
			),
			func(_ ecs.ID, op *ecs.ReadOperation) bool {
				pos := ecs.GetComponent(op, positionType)
				vel := ecs.GetComponent(op, velocityType)
				pos.X += vel.X
				pos.Y += vel.Y
				count++
				return true
			},
		)
		if count != entityCount/2 {
			b.Fatalf("unexpected entity count: got %d, want %d", count, entityCount/2)
		}
	}
}
