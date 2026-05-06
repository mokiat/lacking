package ecs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/ecs/v5"
)

var _ = Describe("Scene", func() {
	type Position struct {
		X, Y float64
	}
	type Age struct {
		Value int
	}
	type Name struct {
		Value string
	}

	var (
		scope        *ecs.Scope
		positionType ecs.ComponentType[Position]
		ageType      ecs.ComponentType[Age]
		nameType     ecs.ComponentType[Name]
		scene        *ecs.Scene
	)

	BeforeEach(func() {
		scope = ecs.NewScope()
		positionType = ecs.RegisterType[Position](scope)
		ageType = ecs.RegisterType[Age](scope)
		nameType = ecs.RegisterType[Name](scope)
		scene = ecs.NewScene(scope)
	})

	Specify("can create entity", func() {
		id := scene.CreateEntity()
		Expect(id).ToNot(Equal(ecs.NilID))
	})

	Specify("entities have unique IDs", func() {
		id1 := scene.CreateEntity()
		Expect(id1).ToNot(Equal(ecs.NilID))

		id2 := scene.CreateEntity()
		Expect(id2).ToNot(Equal(ecs.NilID))

		Expect(id2).ToNot(Equal(id1))
	})

	Specify("can check for entity existence", func() {
		id := scene.CreateEntity()
		Expect(scene.HasEntity(id)).To(BeTrue())

		Expect(scene.HasEntity(ecs.NilID)).To(BeFalse())
	})

	Specify("can delete entity", func() {
		id := scene.CreateEntity()
		Expect(scene.HasEntity(id)).To(BeTrue())

		scene.DeleteEntity(id)
		Expect(scene.HasEntity(id)).To(BeFalse())
	})

	Specify("deleting an entity does not affect other entities", func() {
		id1 := scene.CreateEntity()
		id2 := scene.CreateEntity()

		scene.DeleteEntity(id1)

		Expect(scene.HasEntity(id1)).To(BeFalse())
		Expect(scene.HasEntity(id2)).To(BeTrue())
	})

	Specify("can add components to entity", func() {
		id := scene.CreateEntity()

		pos := Position{X: 1, Y: 2}
		name := Name{Value: "Alice"}

		scene.EditEntity(id, func(op *ecs.EditOperation) {
			ecs.AddComponent(op, positionType, pos)
			ecs.AddComponent(op, nameType, name)
		})
	})

	When("having an entity with components", func() {
		var id ecs.ID

		BeforeEach(func() {
			id = scene.CreateEntity()

			pos := Position{X: 1, Y: 2}
			name := Name{Value: "Alice"}

			scene.EditEntity(id, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, positionType, pos)
				ecs.AddComponent(op, nameType, name)
			})
		})

		Specify("can check whether it satisfies a positive condition", func() {
			ok := scene.CheckEntity(id, ecs.HasComponent(positionType))
			Expect(ok).To(BeTrue())

			ok = scene.CheckEntity(id, ecs.HasComponent(nameType))
			Expect(ok).To(BeTrue())

			ok = scene.CheckEntity(id, ecs.HasComponent(ageType))
			Expect(ok).To(BeFalse())
		})

		Specify("can check whether it satisfies a negative condition", func() {
			ok := scene.CheckEntity(id, ecs.LacksComponent(positionType))
			Expect(ok).To(BeFalse())

			ok = scene.CheckEntity(id, ecs.LacksComponent(nameType))
			Expect(ok).To(BeFalse())

			ok = scene.CheckEntity(id, ecs.LacksComponent(ageType))
			Expect(ok).To(BeTrue())
		})

		Specify("can check whether it satisfies a composite condition", func() {
			ok := scene.CheckEntity(id, ecs.Conditions(
				ecs.HasComponent(positionType),
				ecs.HasComponent(nameType),
				ecs.LacksComponent(ageType),
			))
			Expect(ok).To(BeTrue())

			ok = scene.CheckEntity(id, ecs.Conditions(
				ecs.LacksComponent(positionType),
				ecs.HasComponent(nameType),
				ecs.LacksComponent(ageType),
			))
			Expect(ok).To(BeFalse())

			ok = scene.CheckEntity(id, ecs.Conditions(
				ecs.HasComponent(positionType),
				ecs.LacksComponent(nameType),
				ecs.LacksComponent(ageType),
			))
			Expect(ok).To(BeFalse())

			ok = scene.CheckEntity(id, ecs.Conditions(
				ecs.HasComponent(positionType),
				ecs.HasComponent(nameType),
				ecs.HasComponent(ageType),
			))
			Expect(ok).To(BeFalse())
		})

		When("a component is removed", func() {
			BeforeEach(func() {
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.RemoveComponent(op, positionType)
				})
			})

			Specify("the entity no longer satisfies conditions requiring that component", func() {
				ok := scene.CheckEntity(id, ecs.HasComponent(positionType))
				Expect(ok).To(BeFalse())

				ok = scene.CheckEntity(id, ecs.Conditions(
					ecs.HasComponent(positionType),
					ecs.HasComponent(nameType),
				))
				Expect(ok).To(BeFalse())
			})
		})

		When("the entity is deleted", func() {
			BeforeEach(func() {
				scene.DeleteEntity(id)
			})

			Specify("the entity no longer satisfies any conditions", func() {
				ok := scene.CheckEntity(id, ecs.HasComponent(positionType))
				Expect(ok).To(BeFalse())

				ok = scene.CheckEntity(id, ecs.HasComponent(nameType))
				Expect(ok).To(BeFalse())

				ok = scene.CheckEntity(id, ecs.HasComponent(ageType))
				Expect(ok).To(BeFalse())
			})
		})

		Specify("can read components", func() {
			var (
				pos  *Position
				name *Name
				age  *Age
			)
			scene.ReadEntity(id, func(op *ecs.ReadOperation) {
				pos = ecs.GetComponent(op, positionType)
				name = ecs.GetComponent(op, nameType)
				age = ecs.GetComponent(op, ageType)
			})

			Expect(pos).ToNot(BeNil())
			Expect(*pos).To(Equal(Position{X: 1, Y: 2}))

			Expect(name).ToNot(BeNil())
			Expect(*name).To(Equal(Name{Value: "Alice"}))

			Expect(age).To(BeNil())
		})
	})

})
