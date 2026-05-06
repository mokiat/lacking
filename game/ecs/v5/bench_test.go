package ecs_test

import (
	"testing"

	"github.com/mokiat/lacking/game/ecs/v5"
)

// go test -test.fullpath=true -cpuprofile=cpu.pprof -memprofile=mem.pprof -benchmem -run=^$ -bench ^Benchmark github.com/mokiat/lacking/game/ecs/v5 -count=1

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

	id := scene.CreateEntity()
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

	id := scene.CreateEntity()
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
