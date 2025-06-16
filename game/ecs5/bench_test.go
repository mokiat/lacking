package ecs5_test

import (
	"strconv"
	"testing"

	"github.com/mokiat/lacking/game/ecs5"
)

type NameComponent struct {
	name string
}

func (c *NameComponent) SetName(name string) {
	c.name = name
}

func (c *NameComponent) Name() string {
	return c.name
}

type AgeComponent struct {
	age int
}

func (c *AgeComponent) SetAge(age int) {
	c.age = age
}

func (c *AgeComponent) Age() int {
	return c.age
}

const entityCount = 1024 * 1024

// func BenchmarkSetUnset(b *testing.B) {
// 	engine := ecs5.NewEngine()
// 	scene := engine.CreateScene()

// 	_ = ecs5.NewDenseComponentSet[BaseNameComponent](scene)
// 	ageComponents := ecs5.NewDenseComponentSet[BaseAgeComponent](scene)

// 	entities := make([]ecs5.Entity, entityCount)
// 	for i := range entityCount {
// 		entities[i] = scene.CreateEntity()
// 	}

// 	for b.Loop() {
// 		for i := range entityCount {
// 			ageComponents.Set(entities[i], BaseAgeComponent{
// 				age: i,
// 			})
// 		}
// 		for i := range entityCount {
// 			ageComponents.Unset(entities[i])
// 		}
// 	}
// }

func BenchmarkQueryDense(b *testing.B) {
	engine := ecs5.NewEngine()
	scene := engine.CreateScene()

	nameComponents := ecs5.NewDenseComponentSet[NameComponent](scene)
	ageComponents := ecs5.NewDenseComponentSet[AgeComponent](scene)

	for i := range entityCount {
		entity := scene.CreateEntity()
		nameComponents.Set(entity, NameComponent{
			name: strconv.Itoa(i),
		})
		if i%2 == 0 {
			ageComponents.Set(entity, AgeComponent{
				age: i,
			})
		}
	}

	scene.Query().Release() // prepare cache

	type FakeType struct {
		*NameComponent
		*AgeComponent
	}

	for b.Loop() {
		result := scene.Query(
			ecs5.HasComponent(nameComponents),
			ecs5.HasComponent(ageComponents),
		)
		result.Each(func(entity ecs5.Entity) {
			obj := FakeType{
				NameComponent: nameComponents.Ref(entity),
				AgeComponent:  ageComponents.Ref(entity),
			}
			obj.SetName("test")
			obj.SetAge(37)
		})
		result.Release()
	}
}

func BenchmarkQuerySparse(b *testing.B) {
	engine := ecs5.NewEngine()
	scene := engine.CreateScene()

	nameComponents := ecs5.NewSparseComponentSet[NameComponent](scene)
	ageComponents := ecs5.NewSparseComponentSet[AgeComponent](scene)

	for i := range entityCount {
		entity := scene.CreateEntity()
		nameComponents.Set(entity, NameComponent{
			name: strconv.Itoa(i),
		})
		if i%2 == 0 {
			ageComponents.Set(entity, AgeComponent{
				age: i,
			})
		}
	}

	scene.Query().Release() // prepare cache

	type FakeType struct {
		*NameComponent
		*AgeComponent
	}

	for b.Loop() {
		result := scene.Query(
			ecs5.HasComponent(nameComponents),
			ecs5.HasComponent(ageComponents),
		)
		result.Each(func(entity ecs5.Entity) {
			obj := FakeType{
				NameComponent: nameComponents.Ref(entity),
				AgeComponent:  ageComponents.Ref(entity),
			}
			obj.SetName("test")
			obj.SetAge(37)
		})
		result.Release()
	}
}
