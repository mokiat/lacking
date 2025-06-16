package ecs5_test

import (
	"strconv"
	"testing"

	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/ecs5"
)

var NameComponentID = ecs.NewComponentTypeID()

type NameComponent interface {
	SetName(string)
	Name() string
}

type BaseNameComponent struct {
	name string
}

func (*BaseNameComponent) TypeID() ecs.ComponentTypeID {
	return NameComponentID
}

func (c *BaseNameComponent) SetName(name string) {
	c.name = name
}

func (c *BaseNameComponent) Name() string {
	return c.name
}

var AgeComponentID = ecs.NewComponentTypeID()

type AgeComponent interface {
	SetAge(int)
	Age() int
}

type BaseAgeComponent struct {
	age int
}

func (*BaseAgeComponent) TypeID() ecs.ComponentTypeID {
	return AgeComponentID
}

func (c *BaseAgeComponent) SetAge(age int) {
	c.age = age
}

func (c *BaseAgeComponent) Age() int {
	return c.age
}

const entityCount = 1024 * 1024

func BenchmarkRawQuery(b *testing.B) {
	type primary struct {
		BaseNameComponent
		BaseAgeComponent
	}

	primarySet := make([]primary, entityCount/2)
	for i := range entityCount / 2 {
		obj := &primarySet[i]
		obj.SetName(strconv.Itoa(i))
		obj.SetAge(i)
	}

	for b.Loop() {
		for i := range primarySet {
			obj := &primarySet[i]
			obj.SetName("test")
			obj.SetAge(i)
		}
	}
}

// func BenchmarkOriginalSetUnset(b *testing.B) {
// 	engine := ecs.NewEngine()
// 	scene := engine.CreateScene()

// 	entities := make([]*ecs.Entity, entityCount)
// 	for i := range entityCount {
// 		entities[i] = scene.CreateEntity()
// 	}

// 	for b.Loop() {
// 		for i := range entityCount {
// 			ecs.AttachComponent(entities[i], &BaseAgeComponent{
// 				age: i,
// 			})
// 		}
// 		for i := range entityCount {
// 			entities[i].DeleteComponent(AgeComponentID)
// 		}
// 	}
// }

func BenchmarkOriginalQuery(b *testing.B) {
	engine := ecs.NewEngine()
	scene := engine.CreateScene()

	for i := range entityCount {
		entity := scene.CreateEntity()
		ecs.AttachComponent(entity, &BaseNameComponent{
			name: strconv.Itoa(i),
		})
		if i%2 == 0 {
			ecs.AttachComponent(entity, &BaseAgeComponent{
				age: i,
			})
		}
	}

	scene.Find(ecs.Having(NameComponentID)).Close() // prepare cache

	type FakeType struct {
		*BaseNameComponent
		*BaseAgeComponent
	}

	for b.Loop() {
		result := scene.Find(ecs.Having(NameComponentID).And(AgeComponentID))
		result.Each(func(entity *ecs.Entity) {
			var obj FakeType
			ecs.FetchComponent(entity, &obj.BaseNameComponent)
			ecs.FetchComponent(entity, &obj.BaseAgeComponent)

			obj.SetName("test")
			obj.SetAge(37)
		})
		result.Close()
	}
}

// func BenchmarkBoilerplateSetUnset(b *testing.B) {
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

func BenchmarkBoilerplateQuery(b *testing.B) {
	engine := ecs5.NewEngine()
	scene := engine.CreateScene()

	nameComponents := ecs5.NewDenseComponentSet[BaseNameComponent](scene)
	ageComponents := ecs5.NewDenseComponentSet[BaseAgeComponent](scene)

	for i := range entityCount {
		entity := scene.CreateEntity()
		nameComponents.Set(entity, BaseNameComponent{
			name: strconv.Itoa(i),
		})
		if i%2 == 0 {
			ageComponents.Set(entity, BaseAgeComponent{
				age: i,
			})
		}
	}

	scene.Query().Release() // prepare cache

	type FakeType struct {
		*BaseNameComponent
		*BaseAgeComponent
	}

	for b.Loop() {
		result := scene.Query(
			ecs5.HasComponent(nameComponents),
			ecs5.HasComponent(ageComponents),
		)
		result.Each(func(entity ecs5.Entity) {
			obj := FakeType{
				BaseNameComponent: nameComponents.Ref(entity),
				BaseAgeComponent:  ageComponents.Ref(entity),
			}
			obj.SetName("test")
			obj.SetAge(37)
		})
		result.Release()
	}
}
