package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/game/ecs/internal"
)

var _ = Describe("TypeMask", func() {
	var mask internal.TypeMask

	BeforeEach(func() {
		mask = internal.EmptyTypeMask()
	})

	It("has no types when empty", func() {
		for index := range internal.MaxComponentTypes {
			id := internal.TypeID(index)
			Expect(mask.HasType(id)).To(BeFalse())
		}
	})

	When("types are added", func() {
		BeforeEach(func() {
			mask.AddType(internal.TypeID(0))
			mask.AddType(internal.TypeID(64))
			mask.AddType(internal.TypeID(255))
		})

		It("contains the added types", func() {
			Expect(mask.HasType(internal.TypeID(0))).To(BeTrue())
			Expect(mask.HasType(internal.TypeID(64))).To(BeTrue())
			Expect(mask.HasType(internal.TypeID(255))).To(BeTrue())
		})

		It("does not contain other types", func() {
			Expect(mask.HasType(internal.TypeID(1))).To(BeFalse())
			Expect(mask.HasType(internal.TypeID(36))).To(BeFalse())
			Expect(mask.HasType(internal.TypeID(62))).To(BeFalse())
			Expect(mask.HasType(internal.TypeID(63))).To(BeFalse())
			Expect(mask.HasType(internal.TypeID(254))).To(BeFalse())
		})

		When("cleared", func() {
			BeforeEach(func() {
				mask.Clear()
			})

			It("no longer contains any types", func() {
				for index := range internal.MaxComponentTypes {
					id := internal.TypeID(index)
					Expect(mask.HasType(id)).To(BeFalse())
				}
			})
		})

		When("types are removed", func() {
			BeforeEach(func() {
				mask.RemoveType(internal.TypeID(64))
			})

			It("no longer contains the removed type", func() {
				Expect(mask.HasType(internal.TypeID(64))).To(BeFalse())
			})

			It("still contains the other types", func() {
				Expect(mask.HasType(internal.TypeID(0))).To(BeTrue())
				Expect(mask.HasType(internal.TypeID(255))).To(BeTrue())
			})
		})

		When("inverted", func() {
			var inverted internal.TypeMask

			BeforeEach(func() {
				inverted = mask.Inverted()
			})

			It("does not contain the original types", func() {
				Expect(inverted.HasType(internal.TypeID(0))).To(BeFalse())
				Expect(inverted.HasType(internal.TypeID(64))).To(BeFalse())
				Expect(inverted.HasType(internal.TypeID(255))).To(BeFalse())
			})

			It("contains all other types", func() {
				for index := range internal.MaxComponentTypes {
					id := internal.TypeID(index)
					if gog.IsOneOf(id, 0, 64, 255) {
						continue
					}
					Expect(inverted.HasType(id)).To(BeTrue())
				}
			})
		})

		When("combining masks", func() {
			BeforeEach(func() {
				other := internal.TypeMaskFromTypes(64, 128)
				mask.Combine(other)
			})

			It("contains types from both masks", func() {
				Expect(mask.HasType(internal.TypeID(0))).To(BeTrue())
				Expect(mask.HasType(internal.TypeID(64))).To(BeTrue())
				Expect(mask.HasType(internal.TypeID(128))).To(BeTrue())
				Expect(mask.HasType(internal.TypeID(255))).To(BeTrue())
			})
		})

		It("is possible to check for intersection with another mask", func() {
			other := internal.TypeMaskFromTypes(64, 255)
			Expect(mask.Intersects(other)).To(BeTrue())

			other = internal.TypeMaskFromTypes(255)
			Expect(mask.Intersects(other)).To(BeTrue())

			other = internal.TypeMaskFromTypes(1)
			Expect(mask.Intersects(other)).To(BeFalse())
		})

		It("is possible to check if it contains another mask", func() {
			other := internal.TypeMaskFromTypes(0, 64)
			Expect(mask.Contains(other)).To(BeTrue())

			other = internal.TypeMaskFromTypes(0, 255)
			Expect(mask.Contains(other)).To(BeTrue())

			other = internal.TypeMaskFromTypes(0, 64, 128, 255)
			Expect(mask.Contains(other)).To(BeFalse())
		})

		It("can iterate over its types", func() {
			var types []internal.TypeID
			mask.EachType(func(id internal.TypeID) {
				types = append(types, id)
			})
			Expect(types).To(ConsistOf(
				internal.TypeID(0),
				internal.TypeID(64),
				internal.TypeID(255),
			))
		})
	})
})
