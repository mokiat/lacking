package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape2d"
)

var _ = Describe("CompactTree", func() {
	var (
		tree *shape2d.CompactTree[string]
	)

	BeforeEach(func() {
		tree = shape2d.NewCompactTree[string](shape2d.CompactTreeSettings{
			Size:     opt.V(128.0),
			MaxDepth: opt.V[uint32](3),
		})
	})

	It("has the correct initial state", func() {
		state := tree.Stats()
		Expect(state.NodeCount).To(Equal(uint32(1))) // only root node
		Expect(state.ItemCount).To(Equal(uint32(0)))
	})

	When("items are inserted", func() {
		var (
			firstItemID  shape2d.CompactTreeItemID
			secondItemID shape2d.CompactTreeItemID
			thirdItemID  shape2d.CompactTreeItemID
		)

		BeforeEach(func() {
			firstItemID = tree.Insert(
				shape2d.SquareAreaFromCircle(dprec.NewVec2(16.0, 16.0), 2.0),
				"First",
			)
			secondItemID = tree.Insert(
				shape2d.SquareAreaFromCircle(dprec.NewVec2(48.0, 48.0), 2.0),
				"Second",
			)
			thirdItemID = tree.Insert(
				shape2d.SquareAreaFromCircle(dprec.NewVec2(-16.0, -48.0), 32.0),
				"Third",
			)
		})

		It("returns unique ids", func() {
			Expect(firstItemID).ToNot(Equal(secondItemID))
			Expect(firstItemID).ToNot(Equal(thirdItemID))
			Expect(secondItemID).ToNot(Equal(thirdItemID))
		})

		It("has the correct state", func() {
			state := tree.Stats()
			Expect(state.NodeCount).To(Equal(uint32(5)))
			Expect(state.ItemCount).To(Equal(uint32(3)))
			Expect(state.ItemCountPerDepth).To(Equal([]uint32{
				0, 1, 2,
			}))
		})

		When("an item is removed", func() {
			BeforeEach(func() {
				tree.Remove(secondItemID)
			})

			It("has the correct state", func() {
				state := tree.Stats()
				Expect(state.NodeCount).To(Equal(uint32(4)))
				Expect(state.ItemCount).To(Equal(uint32(2)))
				Expect(state.ItemCountPerDepth).To(Equal([]uint32{
					0, 1, 1,
				}))
			})

			It("does not return an active item id on new insert", func() {
				tree.GC() // forces internal reordering of items (white box testing)
				secondItemID = tree.Insert(
					shape2d.SquareAreaFromCircle(dprec.NewVec2(48.0, 48.0), 2.0),
					"Second",
				)
				Expect(secondItemID).ToNot(Equal(firstItemID))
				Expect(secondItemID).ToNot(Equal(thirdItemID))
			})
		})
	})
})
