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
		id := scene.CreateEntity(nil)
		Expect(id).ToNot(Equal(ecs.NilID))
	})

	Specify("entities have unique IDs", func() {
		id1 := scene.CreateEntity(nil)
		Expect(id1).ToNot(Equal(ecs.NilID))

		id2 := scene.CreateEntity(nil)
		Expect(id2).ToNot(Equal(ecs.NilID))

		Expect(id2).ToNot(Equal(id1))
	})

	Specify("can check for entity existence", func() {
		id := scene.CreateEntity(nil)
		Expect(scene.HasEntity(id)).To(BeTrue())

		Expect(scene.HasEntity(ecs.NilID)).To(BeFalse())
	})

	Specify("can delete entity", func() {
		id := scene.CreateEntity(nil)
		Expect(scene.HasEntity(id)).To(BeTrue())

		scene.DeleteEntity(id)
		Expect(scene.HasEntity(id)).To(BeFalse())
	})

	Specify("deleting an entity does not affect other entities", func() {
		id1 := scene.CreateEntity(nil)
		id2 := scene.CreateEntity(nil)

		scene.DeleteEntity(id1)

		Expect(scene.HasEntity(id1)).To(BeFalse())
		Expect(scene.HasEntity(id2)).To(BeTrue())
	})

	Specify("can add components to entity", func() {
		id := scene.CreateEntity(nil)

		pos := Position{X: 1, Y: 2}
		name := Name{Value: "Alice"}

		scene.EditEntity(id, func(op *ecs.EditOperation) {
			ecs.AddComponent(op, positionType, pos)
			ecs.AddComponent(op, nameType, name)
		})
	})

	Specify("can create entity with initial components", func() {
		id := scene.CreateEntity(func(op *ecs.EditOperation) {
			ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
			ecs.AddComponent(op, nameType, Name{Value: "Alice"})
		})

		Expect(scene.CheckEntity(id, ecs.HasComponent(positionType))).To(BeTrue())
		Expect(scene.CheckEntity(id, ecs.HasComponent(nameType))).To(BeTrue())
		Expect(scene.CheckEntity(id, ecs.HasComponent(ageType))).To(BeFalse())

		var pos *Position
		var name *Name
		scene.ReadEntity(id, func(op *ecs.ReadOperation) {
			pos = ecs.GetComponent(op, positionType)
			name = ecs.GetComponent(op, nameType)
		})
		Expect(*pos).To(Equal(Position{X: 1, Y: 2}))
		Expect(*name).To(Equal(Name{Value: "Alice"}))
	})

	Specify("creation callback fires enter subscriptions with initial components", func() {
		var entered []ecs.ID
		scene.SubscribeEnter(ecs.HasComponent(positionType), func(id ecs.ID) {
			entered = append(entered, id)
		})

		id := scene.CreateEntity(func(op *ecs.EditOperation) {
			ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
		})
		Expect(entered).To(ConsistOf(id))
	})

	When("having an entity with components", func() {
		var id ecs.ID

		BeforeEach(func() {
			id = scene.CreateEntity(nil)

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

		Specify("can replace a component value", func() {
			scene.EditEntity(id, func(op *ecs.EditOperation) {
				ecs.ReplaceComponent(op, positionType, Position{X: 10, Y: 20})
			})

			var pos *Position
			scene.ReadEntity(id, func(op *ecs.ReadOperation) {
				pos = ecs.GetComponent(op, positionType)
			})
			Expect(*pos).To(Equal(Position{X: 10, Y: 20}))
		})

		Specify("can remove and re-add a component in the same edit, updating its value", func() {
			scene.EditEntity(id, func(op *ecs.EditOperation) {
				ecs.RemoveComponent(op, positionType)
				ecs.AddComponent(op, positionType, Position{X: 10, Y: 20})
			})

			Expect(scene.CheckEntity(id, ecs.HasComponent(positionType))).To(BeTrue())

			var pos *Position
			scene.ReadEntity(id, func(op *ecs.ReadOperation) {
				pos = ecs.GetComponent(op, positionType)
			})
			Expect(*pos).To(Equal(Position{X: 10, Y: 20}))
		})

		Specify("adding and removing the same component in the same edit is a no-op", func() {
			scene.EditEntity(id, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, ageType, Age{Value: 42})
				ecs.RemoveComponent(op, ageType)
			})

			Expect(scene.CheckEntity(id, ecs.HasComponent(ageType))).To(BeFalse())
		})
	})

	When("having multiple entities with various component combinations", func() {
		var (
			entityPosName  ecs.ID
			entityPosAge   ecs.ID
			entityNameOnly ecs.ID
		)

		BeforeEach(func() {
			entityPosName = scene.CreateEntity(nil)
			scene.EditEntity(entityPosName, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				ecs.AddComponent(op, nameType, Name{Value: "Alice"})
			})

			entityPosAge = scene.CreateEntity(nil)
			scene.EditEntity(entityPosAge, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, positionType, Position{X: 3, Y: 4})
				ecs.AddComponent(op, ageType, Age{Value: 30})
			})

			entityNameOnly = scene.CreateEntity(nil)
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
					scene.CreateEntity(nil)
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

	When("subscribing to entity enter events", func() {
		var entered []ecs.ID

		BeforeEach(func() {
			entered = nil
		})

		When("condition is HasComponent", func() {
			BeforeEach(func() {
				scene.SubscribeEnter(ecs.HasComponent(positionType), func(id ecs.ID) {
					entered = append(entered, id)
				})
			})

			Specify("does not fire when entity is created with no components", func() {
				scene.CreateEntity(nil)
				Expect(entered).To(BeEmpty())
			})

			Specify("fires when entity gains the required component", func() {
				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				})
				Expect(entered).To(ConsistOf(id))
			})

			Specify("does not fire again when another component is added while condition remains satisfied", func() {
				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				})
				entered = nil

				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, ageType, Age{Value: 30})
				})
				Expect(entered).To(BeEmpty())
			})

			Specify("fires again after entity re-gains the component", func() {
				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				})
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.RemoveComponent(op, positionType)
				})
				entered = nil

				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 3, Y: 4})
				})
				Expect(entered).To(ConsistOf(id))
			})

			Specify("does not fire when entity is deleted", func() {
				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				})
				entered = nil

				scene.DeleteEntity(id)
				Expect(entered).To(BeEmpty())
			})

			Specify("all subscribers receive the notification", func() {
				var secondEntered []ecs.ID
				scene.SubscribeEnter(ecs.HasComponent(positionType), func(id ecs.ID) {
					secondEntered = append(secondEntered, id)
				})

				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				})
				Expect(entered).To(ConsistOf(id))
				Expect(secondEntered).To(ConsistOf(id))
			})

			Specify("stops firing after subscription is deleted", func() {
				sub := scene.SubscribeEnter(ecs.HasComponent(ageType), func(id ecs.ID) {
					entered = append(entered, id)
				})
				sub.Delete()

				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, ageType, Age{Value: 25})
				})
				Expect(entered).To(BeEmpty())
			})

			Specify("fires for a composite condition only when all components are present", func() {
				var compositeEntered []ecs.ID
				scene.SubscribeEnter(ecs.Conditions(
					ecs.HasComponent(positionType),
					ecs.HasComponent(nameType),
				), func(id ecs.ID) {
					compositeEntered = append(compositeEntered, id)
				})

				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				})
				Expect(compositeEntered).To(BeEmpty())

				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, nameType, Name{Value: "Alice"})
				})
				Expect(compositeEntered).To(ConsistOf(id))
			})

			Specify("does not fire when a component is replaced in place", func() {
				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				})
				entered = nil

				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.ReplaceComponent(op, positionType, Position{X: 3, Y: 4})
				})
				Expect(entered).To(BeEmpty())
			})

			Specify("does not fire when a component is removed and re-added in the same edit", func() {
				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				})
				entered = nil

				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.RemoveComponent(op, positionType)
					ecs.AddComponent(op, positionType, Position{X: 3, Y: 4})
				})
				Expect(entered).To(BeEmpty())
			})
		})

		When("condition is LacksComponent", func() {
			BeforeEach(func() {
				scene.SubscribeEnter(ecs.LacksComponent(positionType), func(id ecs.ID) {
					entered = append(entered, id)
				})
			})

			Specify("fires when entity is created (starts without the excluded component)", func() {
				id := scene.CreateEntity(nil)
				Expect(entered).To(ConsistOf(id))
			})

			Specify("does not fire again when unrelated component is added", func() {
				id := scene.CreateEntity(nil)
				entered = nil

				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, ageType, Age{Value: 30})
				})
				Expect(entered).To(BeEmpty())
			})

			Specify("fires again when entity re-loses the excluded component", func() {
				id := scene.CreateEntity(nil)
				entered = nil
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				})
				Expect(entered).To(BeEmpty())

				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.RemoveComponent(op, positionType)
				})
				Expect(entered).To(ConsistOf(id))
			})
		})
	})

	When("subscribing to entity exit events", func() {
		var exited []ecs.ID

		BeforeEach(func() {
			exited = nil
		})

		When("condition is HasComponent", func() {
			BeforeEach(func() {
				scene.SubscribeExit(ecs.HasComponent(positionType), func(id ecs.ID) {
					exited = append(exited, id)
				})
			})

			Specify("does not fire when entity without the component is created", func() {
				scene.CreateEntity(nil)
				Expect(exited).To(BeEmpty())
			})

			Specify("fires when entity loses the required component", func() {
				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				})
				Expect(exited).To(BeEmpty())

				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.RemoveComponent(op, positionType)
				})
				Expect(exited).To(ConsistOf(id))
			})

			Specify("fires when entity with the component is deleted", func() {
				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				})
				Expect(exited).To(BeEmpty())

				scene.DeleteEntity(id)
				Expect(exited).To(ConsistOf(id))
			})

			Specify("does not fire when entity without the component is deleted", func() {
				id := scene.CreateEntity(nil)
				scene.DeleteEntity(id)
				Expect(exited).To(BeEmpty())
			})

			Specify("does not fire when an unrelated component is removed", func() {
				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
					ecs.AddComponent(op, ageType, Age{Value: 30})
				})
				exited = nil

				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.RemoveComponent(op, ageType)
				})
				Expect(exited).To(BeEmpty())
			})

			Specify("stops firing after subscription is deleted", func() {
				sub := scene.SubscribeExit(ecs.HasComponent(ageType), func(id ecs.ID) {
					exited = append(exited, id)
				})
				sub.Delete()

				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, ageType, Age{Value: 25})
				})
				scene.DeleteEntity(id)
				Expect(exited).To(BeEmpty())
			})

			Specify("does not fire when a component is replaced in place", func() {
				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				})
				Expect(exited).To(BeEmpty())

				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.ReplaceComponent(op, positionType, Position{X: 3, Y: 4})
				})
				Expect(exited).To(BeEmpty())
			})

			Specify("does not fire when a component is removed and re-added in the same edit", func() {
				id := scene.CreateEntity(nil)
				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				})
				Expect(exited).To(BeEmpty())

				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.RemoveComponent(op, positionType)
					ecs.AddComponent(op, positionType, Position{X: 3, Y: 4})
				})
				Expect(exited).To(BeEmpty())
			})
		})

		When("condition is LacksComponent", func() {
			BeforeEach(func() {
				scene.SubscribeExit(ecs.LacksComponent(positionType), func(id ecs.ID) {
					exited = append(exited, id)
				})
			})

			Specify("fires when entity gains the excluded component", func() {
				id := scene.CreateEntity(nil)
				Expect(exited).To(BeEmpty())

				scene.EditEntity(id, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
				})
				Expect(exited).To(ConsistOf(id))
			})

			Specify("does not fire when a component-less entity is deleted", func() {
				// LacksComponent(pos) is satisfied by the empty archetype, and deletion
				// dispatches exit with EmptyTypeMask as the "new" mask — so the condition
				// remains satisfied and exit does not fire.
				id := scene.CreateEntity(nil)
				Expect(exited).To(BeEmpty())
				scene.DeleteEntity(id)
				Expect(exited).To(BeEmpty())
			})
		})
	})

	When("performing mutations from within notification handlers", func() {
		Specify("entity created in enter notification is accessible after the triggering operation", func() {
			var createdID ecs.ID
			scene.SubscribeEnter(ecs.HasComponent(positionType), func(_ ecs.ID) {
				createdID = scene.CreateEntity(nil)
			})

			id := scene.CreateEntity(nil)
			scene.EditEntity(id, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
			})

			Expect(createdID).ToNot(Equal(ecs.NilID))
			Expect(scene.HasEntity(createdID)).To(BeTrue())
		})

		Specify("entity created in notification fires its own enter notifications", func() {
			var lacksPositionEntered []ecs.ID
			scene.SubscribeEnter(ecs.LacksComponent(positionType), func(id ecs.ID) {
				lacksPositionEntered = append(lacksPositionEntered, id)
			})

			var createdID ecs.ID
			scene.SubscribeEnter(ecs.HasComponent(positionType), func(_ ecs.ID) {
				createdID = scene.CreateEntity(nil)
			})

			triggerID := scene.CreateEntity(nil)
			lacksPositionEntered = nil // reset: clear notification from triggerID's own creation
			scene.EditEntity(triggerID, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
			})

			// createdID is created (deferred) after HasPosition enter fires;
			// its own LacksPosition enter should fire in the same processQueue run.
			Expect(lacksPositionEntered).To(ConsistOf(createdID))
		})

		Specify("entity edited in enter notification is updated after the triggering operation", func() {
			targetID := scene.CreateEntity(nil)

			scene.SubscribeEnter(ecs.HasComponent(positionType), func(_ ecs.ID) {
				scene.EditEntity(targetID, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, ageType, Age{Value: 42})
				})
			})

			triggerID := scene.CreateEntity(nil)
			scene.EditEntity(triggerID, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
			})

			Expect(scene.CheckEntity(targetID, ecs.HasComponent(ageType))).To(BeTrue())
		})

		Specify("edit in notification fires its own enter notifications", func() {
			var ageEntered []ecs.ID
			scene.SubscribeEnter(ecs.HasComponent(ageType), func(id ecs.ID) {
				ageEntered = append(ageEntered, id)
			})

			targetID := scene.CreateEntity(nil)

			scene.SubscribeEnter(ecs.HasComponent(positionType), func(_ ecs.ID) {
				scene.EditEntity(targetID, func(op *ecs.EditOperation) {
					ecs.AddComponent(op, ageType, Age{Value: 99})
				})
			})

			triggerID := scene.CreateEntity(nil)
			scene.EditEntity(triggerID, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
			})

			Expect(ageEntered).To(ConsistOf(targetID))
		})

		Specify("entity deleted in exit notification is removed after the triggering operation", func() {
			sideEffectID := scene.CreateEntity(nil)
			scene.EditEntity(sideEffectID, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, positionType, Position{X: 3, Y: 4})
			})

			scene.SubscribeExit(ecs.HasComponent(positionType), func(id ecs.ID) {
				if id != sideEffectID {
					scene.DeleteEntity(sideEffectID)
				}
			})

			triggerID := scene.CreateEntity(nil)
			scene.EditEntity(triggerID, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, positionType, Position{X: 1, Y: 2})
			})
			scene.DeleteEntity(triggerID)

			Expect(scene.HasEntity(sideEffectID)).To(BeFalse())
		})

		Specify("delete in notification fires its own exit notifications", func() {
			var posExited []ecs.ID
			scene.SubscribeExit(ecs.HasComponent(positionType), func(id ecs.ID) {
				posExited = append(posExited, id)
			})

			targetID := scene.CreateEntity(nil)
			scene.EditEntity(targetID, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, positionType, Position{X: 3, Y: 4})
			})
			posExited = nil // reset

			scene.SubscribeEnter(ecs.HasComponent(ageType), func(_ ecs.ID) {
				scene.DeleteEntity(targetID)
			})

			triggerID := scene.CreateEntity(nil)
			scene.EditEntity(triggerID, func(op *ecs.EditOperation) {
				ecs.AddComponent(op, ageType, Age{Value: 10})
			})

			Expect(posExited).To(ConsistOf(targetID))
		})
	})

})
