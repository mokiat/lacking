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

func BenchmarkQueryDense(b *testing.B) {
	engine := ecs5.NewEngine()
	scene := engine.CreateScene()

	nameComponents := ecs5.NewDenseComponentSet[NameComponent](scene)
	ageComponents := ecs5.NewDenseComponentSet[AgeComponent](scene)

	for i := range scene.MaxEntityCount() {
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

	for i := range scene.MaxEntityCount() {
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

func BenchmarkQueryTiny(b *testing.B) {
	engine := ecs5.NewEngine()
	scene := engine.CreateScene()

	nameComponents := ecs5.NewTinyComponentSet[NameComponent](scene)
	ageComponents := ecs5.NewTinyComponentSet[AgeComponent](scene)

	for i := range scene.MaxEntityCount() {
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
