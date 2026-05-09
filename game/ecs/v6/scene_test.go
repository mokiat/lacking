package ecs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/ecs/v6"
)

var _ = Describe("Scene", func() {
	type Position struct {
		X, Y int
	}
	type Age struct {
		Value int
	}
	type Name struct {
		Value string
	}
	type Identification struct {
		ID uint32
	}
	type Unused struct{} // a tag component

	var (
		scope              *ecs.Scope
		positionType       ecs.ComponentType[Position]
		ageType            ecs.ComponentType[Age]
		nameType           ecs.ComponentType[Name]
		identificationType ecs.ComponentType[Identification]
		unusedType         ecs.ComponentType[Unused]
		scene              *ecs.Scene
	)

	BeforeEach(func() {
		scope = ecs.NewScope()
		positionType = ecs.Type[Position](scope)
		ageType = ecs.Type[Age](scope)
		nameType = ecs.Type[Name](scope)
		identificationType = ecs.Type[Identification](scope)
		_ = identificationType // TODO: REMOVE
		unusedType = ecs.Type[Unused](scope)
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

		Specify("can read components from entity", func() {
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

	When("having multiple entities with various component combinations", func() {
		var (
			entityPosName  ecs.ID
			entityPosAge   ecs.ID
			entityNameOnly ecs.ID
		)

		BeforeEach(func() {
			entityPosName = scene.CreateEntity()
			scene.EditEntity(entityPosName, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				ecs.AddComponent(op, nameType, Name{Value: "Alice"})
			})

			entityPosAge = scene.CreateEntity()
			scene.EditEntity(entityPosAge, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, positionType, Position{X: 3, Y: 4})
				ecs.AddComponent(op, ageType, Age{Value: 30})
			})

			entityNameOnly = scene.CreateEntity()
			scene.EditEntity(entityNameOnly, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, nameType, Name{Value: "Bob"})
			})
		})

		Specify("querying by a single condition returns all matching entities", func() {
			var found []ecs.ID
			scene.QueryEntities(ecs.HasComponent(positionType), func(id ecs.ID, _ *ecs.ReadOperation) bool {
				found = append(found, id)
				return true
			})
			Expect(found).To(ConsistOf(entityPosName, entityPosAge))
		})

		Specify("querying returns correct component values for each entity", func() {
			positions := make(map[ecs.ID]Position)
			scene.QueryEntities(ecs.HasComponent(positionType), func(id ecs.ID, op *ecs.ReadOperation) bool {
				positions[id] = *ecs.GetComponent(op, positionType)
				return true
			})
			Expect(positions[entityPosName]).To(Equal(Position{X: 1, Y: 2}))
			Expect(positions[entityPosAge]).To(Equal(Position{X: 3, Y: 4}))
		})

		Specify("querying with a composite condition filters correctly", func() {
			var found []ecs.ID
			scene.QueryEntities(ecs.Conditions(
				ecs.HasComponent(positionType),
				ecs.LacksComponent(ageType),
			), func(id ecs.ID, _ *ecs.ReadOperation) bool {
				found = append(found, id)
				return true
			})
			Expect(found).To(ConsistOf(entityPosName))
		})

		Specify("querying with no matching entities yields nothing", func() {
			var found []ecs.ID
			scene.QueryEntities(ecs.HasComponent(unusedType), func(id ecs.ID, _ *ecs.ReadOperation) bool {
				found = append(found, id)
				return true
			})
			Expect(found).To(BeEmpty())
		})

		Specify("query can be stopped early by returning false", func() {
			count := 0
			scene.QueryEntities(ecs.HasComponent(positionType), func(_ ecs.ID, _ *ecs.ReadOperation) bool {
				count++
				return false
			})
			Expect(count).To(Equal(1))
		})

		Specify("querying via iterator returns all matching entities", func() {
			var found []ecs.ID
			for id := range scene.QueryEntitiesIter(ecs.HasComponent(nameType)) {
				found = append(found, id)
			}
			Expect(found).To(ConsistOf(entityPosName, entityNameOnly))
		})

		Specify("nested query returns correct results", func() {
			pairs := make(map[ecs.ID][]ecs.ID)
			scene.QueryEntities(ecs.HasComponent(positionType), func(outerID ecs.ID, _ *ecs.ReadOperation) bool {
				scene.QueryEntities(ecs.HasComponent(nameType), func(innerID ecs.ID, _ *ecs.ReadOperation) bool {
					pairs[outerID] = append(pairs[outerID], innerID)
					return true
				})
				return true
			})
			Expect(pairs[entityPosName]).To(ConsistOf(entityPosName, entityNameOnly))
			Expect(pairs[entityPosAge]).To(ConsistOf(entityPosName, entityNameOnly))
		})

		Specify("nested query does not corrupt outer read operation", func() {
			scene.QueryEntities(ecs.HasComponent(positionType), func(outerID ecs.ID, outerOp *ecs.ReadOperation) bool {
				outerPos := ecs.GetComponent(outerOp, positionType)

				scene.QueryEntities(ecs.HasComponent(positionType), func(_ ecs.ID, _ *ecs.ReadOperation) bool {
					return true
				})

				Expect(ecs.GetComponent(outerOp, positionType)).To(Equal(outerPos))
				return true
			})
		})

		Specify("editing an entity during a query panics", func() {
			Expect(func() {
				scene.QueryEntities(ecs.HasComponent(positionType), func(id ecs.ID, _ *ecs.ReadOperation) bool {
					scene.EditEntity(id, func(_ *ecs.EditOperation) {})
					return true
				})
			}).To(Panic())
		})

		Specify("creating an entity during a query panics", func() {
			Expect(func() {
				scene.QueryEntities(ecs.HasComponent(positionType), func(_ ecs.ID, _ *ecs.ReadOperation) bool {
					scene.CreateEntity()
					return true
				})
			}).To(Panic())
		})

		Specify("deleting an entity during a query panics", func() {
			Expect(func() {
				scene.QueryEntities(ecs.HasComponent(positionType), func(id ecs.ID, _ *ecs.ReadOperation) bool {
					scene.DeleteEntity(id)
					return true
				})
			}).To(Panic())
		})
	})

})
