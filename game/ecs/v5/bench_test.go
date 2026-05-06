package ecs_test

import (
	"testing"

	"github.com/mokiat/lacking/game/ecs/v5"
)

// go test -test.fullpath=true -cpuprofile=cpu.pprof -memprofile=mem.pprof -benchmem -run=^$ -bench ^BenchmarkCheckEntity$ github.com/mokiat/lacking/game/ecs/v5 -count=1

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

// func BenchmarkAddComponents(b *testing.B) {
// 	scene := ecs.NewScene()
// 	scene.StoreComponent(ecs.NewExperiment(PositionType, Position{}))
// 	scene.StoreComponent(ecs.NewExperiment(stringType, "hello"))
// 	scene.StoreComponent(ecs.NewExperiment(intType, 13))

// 	var (
// 		pos *Position
// 		str *string
// 	)

// 	for b.Loop() {
// 		str = nil
// 		scene.UpdateEntity(ecs.NilEntityID,
// 			ecs.AddComponentAndFetch(PositionType, &pos),
// 			ecs.AddComponentAndFetch(stringType, &str),
// 			ecs.AddComponent(intType),
// 		)
// 		if str == nil || *str == "" {
// 			b.Fatal("unexpected empty string")
// 		}
// 	}
// }

// func BenchmarkFindTypeMap(b *testing.B) {
// 	type Type01 struct{}
// 	type Type02 struct{}
// 	type Type03 struct{}
// 	type Type04 struct{}
// 	type Type05 struct{}
// 	type Type06 struct{}
// 	type Type07 struct{}
// 	type Type08 struct{}
// 	type Type09 struct{}
// 	type Type10 struct{}
// 	type Type11 struct{}
// 	type Type12 struct{}
// 	type Type13 struct{}
// 	type Type14 struct{}
// 	type Type15 struct{}
// 	type Type16 struct{}
// 	type Type17 struct{}
// 	type Type18 struct{}
// 	type Type19 struct{}
// 	type Type20 struct{}

// 	registry := make(map[reflect.Type]int)

// 	registry[reflect.TypeFor[Type01]()] = 1
// 	registry[reflect.TypeFor[Type02]()] = 2
// 	registry[reflect.TypeFor[Type03]()] = 3
// 	registry[reflect.TypeFor[Type04]()] = 4
// 	registry[reflect.TypeFor[Type05]()] = 5
// 	registry[reflect.TypeFor[Type06]()] = 6
// 	registry[reflect.TypeFor[Type07]()] = 7
// 	registry[reflect.TypeFor[Type08]()] = 8
// 	registry[reflect.TypeFor[Type09]()] = 9
// 	registry[reflect.TypeFor[Type10]()] = 10
// 	registry[reflect.TypeFor[Type11]()] = 11
// 	registry[reflect.TypeFor[Type12]()] = 12
// 	registry[reflect.TypeFor[Type13]()] = 13
// 	registry[reflect.TypeFor[Type14]()] = 14
// 	registry[reflect.TypeFor[Type15]()] = 15
// 	registry[reflect.TypeFor[Type16]()] = 16
// 	registry[reflect.TypeFor[Type17]()] = 17
// 	registry[reflect.TypeFor[Type18]()] = 18
// 	registry[reflect.TypeFor[Type19]()] = 19
// 	registry[reflect.TypeFor[Type20]()] = 20

// 	t := reflect.TypeFor[Type15]()
// 	for b.Loop() {
// 		v, ok := registry[t]
// 		if !ok || v != 15 {
// 			b.Fatal("unexpected type lookup result")
// 		}
// 	}
// }

// func BenchmarkFindTypeSlice(b *testing.B) {
// 	type Type01 struct{}
// 	type Type02 struct{}
// 	type Type03 struct{}
// 	type Type04 struct{}
// 	type Type05 struct{}
// 	type Type06 struct{}
// 	type Type07 struct{}
// 	type Type08 struct{}
// 	type Type09 struct{}
// 	type Type10 struct{}
// 	type Type11 struct{}
// 	type Type12 struct{}
// 	type Type13 struct{}
// 	type Type14 struct{}
// 	type Type15 struct{}
// 	type Type16 struct{}
// 	type Type17 struct{}
// 	type Type18 struct{}
// 	type Type19 struct{}
// 	type Type20 struct{}

// 	type registryEntry struct {
// 		t reflect.Type
// 		v int
// 	}
// 	registry := make([]registryEntry, 0)
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type01](), 1})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type02](), 2})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type03](), 3})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type04](), 4})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type05](), 5})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type06](), 6})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type07](), 7})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type08](), 8})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type09](), 9})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type10](), 10})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type11](), 11})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type12](), 12})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type13](), 13})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type14](), 14})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type15](), 15})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type16](), 16})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type17](), 17})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type18](), 18})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type19](), 19})
// 	registry = append(registry, registryEntry{reflect.TypeFor[Type20](), 20})

// 	t := reflect.TypeFor[Type15]()
// 	for b.Loop() {
// 		var v int
// 		var ok bool
// 		for _, entry := range registry {
// 			if entry.t == t {
// 				v = entry.v
// 				ok = true
// 				break
// 			}
// 		}
// 		if !ok || v != 15 {
// 			b.Fatal("unexpected type lookup result")
// 		}
// 	}
// }

// func BenchmarkSqrt(b *testing.B) {
// 	for b.Loop() {
// 		v := math.Sqrt(12345.6789)
// 		if v == 0 {
// 			b.Fatal("unexpected zero result")
// 		}
// 	}
// }
