package shape3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape3d"
)

var _ = Describe("CompactTree", func() {
	var (
		tree *shape3d.CompactTree[string]
	)

	BeforeEach(func() {
		tree = shape3d.NewCompactTree[string](shape3d.CompactTreeSettings{
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
			firstItemID  shape3d.CompactTreeItemID
			secondItemID shape3d.CompactTreeItemID
			thirdItemID  shape3d.CompactTreeItemID
		)

		BeforeEach(func() {
			firstItemID = tree.Insert(
				shape3d.NewCompactCubeFromSphere(dprec.NewVec3(16.0, 16.0, 16.0), 2.0),
				"First",
			)
			secondItemID = tree.Insert(
				shape3d.NewCompactCubeFromSphere(dprec.NewVec3(48.0, 48.0, 48.0), 2.0),
				"Second",
			)
			thirdItemID = tree.Insert(
				shape3d.NewCompactCubeFromSphere(dprec.NewVec3(-16.0, -48.0, -16.0), 32.0),
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

		It("is possible to segment-search for items", func() {
			from := dprec.NewVec3(1.0, 1.0, 1.0)
			to := dprec.NewVec3(127.0, 127.0, 127.0)
			segment := shape3d.NewCompactQuerySegment(from, to)
			var found []string
			tree.QuerySegment(segment, func(item string) bool {
				found = append(found, item)
				return true
			})
			Expect(found).To(ConsistOf("First", "Second"))
		})

		It("is possible to area-search for items", func() {
			area := shape3d.NewCompactQueryAABBFromSphere(dprec.NewVec3(64.0, 64.0, 64.0), 63.0)
			var found []string
			tree.QueryAABB(area, func(item string) bool {
				found = append(found, item)
				return true
			})
			Expect(found).To(ConsistOf("First", "Second"))
		})

		When("items are searched", func() {
			BeforeEach(func() {
				area := shape3d.NewCompactQueryAABBFromSphere(dprec.NewVec3(64.0, 64.0, 64.0), 63.0)
				tree.QueryAABB(area, func(item string) bool {
					return true
				})
			})

			It("returns the correct visit stats", func() {
				stats := tree.VisitStats()
				Expect(stats.NodeCountVisited).To(Equal(uint32(5)))
				Expect(stats.NodeCountAccepted).To(Equal(uint32(4)))
				Expect(stats.NodeCountRejected).To(Equal(uint32(1)))
				Expect(stats.ItemCountVisited).To(Equal(uint32(2)))
				Expect(stats.ItemCountAccepted).To(Equal(uint32(2)))
				Expect(stats.ItemCountRejected).To(Equal(uint32(0)))
			})
		})

		When("an item is updated", func() {
			BeforeEach(func() {
				tree.Update(secondItemID,
					shape3d.NewCompactCubeFromSphere(dprec.NewVec3(-48.0, 48.0, -48.0), 2.0),
				)
			})

			It("has the correct state", func() {
				state := tree.Stats()
				Expect(state.NodeCount).To(Equal(uint32(6)))
				Expect(state.ItemCount).To(Equal(uint32(3)))
				Expect(state.ItemCountPerDepth).To(Equal([]uint32{
					0, 1, 2,
				}))
			})

			It("is reflected in segment-search for items", func() {
				from := dprec.NewVec3(1.0, 1.0, 1.0)
				to := dprec.NewVec3(127.0, 127.0, 127.0)
				segment := shape3d.NewCompactQuerySegment(from, to)
				var found []string
				tree.QuerySegment(segment, func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(ConsistOf("First"))
			})

			It("is reflected in area-search for items", func() {
				area := shape3d.NewCompactQueryAABBFromSphere(dprec.NewVec3(64.0, 64.0, 64.0), 63.0)
				var found []string
				tree.QueryAABB(area, func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(ConsistOf("First"))
			})
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
					shape3d.NewCompactCubeFromSphere(dprec.NewVec3(48.0, 48.0, 48.0), 2.0),
					"Second",
				)
				Expect(secondItemID).ToNot(Equal(firstItemID))
				Expect(secondItemID).ToNot(Equal(thirdItemID))
			})

			It("is reflected in segment-search for items", func() {
				from := dprec.NewVec3(1.0, 1.0, 1.0)
				to := dprec.NewVec3(127.0, 127.0, 127.0)
				segment := shape3d.NewCompactQuerySegment(from, to)
				var found []string
				tree.QuerySegment(segment, func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(ConsistOf("First"))
			})

			It("is reflected in area-search for items", func() {
				area := shape3d.NewCompactQueryAABBFromSphere(dprec.NewVec3(64.0, 64.0, 64.0), 63.0)
				var found []string
				tree.QueryAABB(area, func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(ConsistOf("First"))
			})
		})
	})
})
