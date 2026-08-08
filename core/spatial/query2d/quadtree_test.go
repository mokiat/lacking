package query2d_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/query2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// aabbFromCircle builds an AABB enclosing a circle with the given center
// coordinates and radius.
func aabbFromCircle(x, y, radius float64) shape2d.AABB {
	return shape2d.AABBFromCircle(shape2d.Circle{
		Center: dprec.NewVec2(x, y),
		Radius: radius,
	})
}

var _ = Describe("Quadtree", func() {
	var (
		tree *query2d.Quadtree[string]
	)

	BeforeEach(func() {
		tree = query2d.NewQuadtree[string](query2d.QuadtreeSettings{
			Size:     opt.V(128.0),
			MaxDepth: opt.V[uint32](3),
		})
	})

	It("has the correct initial state", func() {
		state := tree.Stats()
		Expect(state.NodeCount).To(Equal(uint32(1))) // only root node
		Expect(state.ItemCount).To(Equal(uint32(0)))
	})

	It("panics when an item with an empty box is inserted", func() {
		emptyAABB := shape2d.NewAABB(1.0, 1.0, -1.0, -1.0)
		Expect(func() { tree.Insert(emptyAABB, "Empty") }).To(Panic())
	})

	It("panics when an item is updated to an empty box", func() {
		itemID := tree.Insert(aabbFromCircle(0.0, 0.0, 1.0), "Item")
		emptyAABB := shape2d.NewAABB(1.0, 1.0, -1.0, -1.0)
		Expect(func() { tree.Update(itemID, emptyAABB) }).To(Panic())
	})

	When("an item has a non-square box", func() {
		BeforeEach(func() {
			// A bar stretching along the X axis. Its bounding square would span
			// 40 units in every direction, whereas the box itself is only two
			// units thick along Y.
			tree.Insert(
				shape2d.NewAABB(-40.0, -2.0, 40.0, 2.0),
				"Bar",
			)
		})

		It("is found through a query that overlaps the box", func() {
			var found []string
			tree.QueryAABB(aabbFromCircle(30.0, 0.0, 2.0), func(item string) bool {
				found = append(found, item)
				return true
			})
			Expect(found).To(ConsistOf("Bar"))
		})

		It("is not found through a query that only overlaps its bounding square", func() {
			var found []string
			tree.QueryAABB(aabbFromCircle(30.0, 20.0, 2.0), func(item string) bool {
				found = append(found, item)
				return true
			})
			Expect(found).To(BeEmpty())
		})

		It("is not found through a segment that only crosses its bounding square", func() {
			segment := shape2d.NewSegment(
				dprec.NewVec2(-30.0, 10.0),
				dprec.NewVec2(30.0, 10.0),
			)
			var found []string
			tree.QuerySegment(segment, func(item string) bool {
				found = append(found, item)
				return true
			})
			Expect(found).To(BeEmpty())
		})
	})

	When("items are inserted", func() {
		var (
			firstItemID  query2d.TreeItemID
			secondItemID query2d.TreeItemID
			thirdItemID  query2d.TreeItemID
		)

		BeforeEach(func() {
			firstItemID = tree.Insert(
				aabbFromCircle(16.0, 16.0, 2.0),
				"First",
			)
			secondItemID = tree.Insert(
				aabbFromCircle(48.0, 48.0, 2.0),
				"Second",
			)
			thirdItemID = tree.Insert(
				aabbFromCircle(-16.0, -48.0, 32.0),
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
			Expect(state.NodeCount).To(Equal(uint32(6)))
			Expect(state.ItemCount).To(Equal(uint32(3)))
			// The third item is as wide as a whole child node, but it is
			// positioned so that it still fits within the loose area of a
			// grandchild along both axes.
			Expect(state.ItemCountPerDepth).To(Equal([]uint32{
				0, 0, 3,
			}))
		})

		It("is possible to segment-search for items", func() {
			from := dprec.NewVec2(1.0, 1.0)
			to := dprec.NewVec2(127.0, 127.0)
			segment := shape2d.NewSegment(from, to)
			var found []string
			tree.QuerySegment(segment, func(item string) bool {
				found = append(found, item)
				return true
			})
			Expect(found).To(ConsistOf("First", "Second"))
		})

		It("stops QuerySegment after the visitor returns false", func() {
			from := dprec.NewVec2(1.0, 1.0)
			to := dprec.NewVec2(127.0, 127.0)
			segment := shape2d.NewSegment(from, to)
			count := 0
			tree.QuerySegment(segment, func(item string) bool {
				count++
				return false // stop after first item
			})
			Expect(count).To(Equal(1))
		})

		It("is possible to area-search for items", func() {
			aabb := aabbFromCircle(64.0, 64.0, 63.0)
			var found []string
			tree.QueryAABB(aabb, func(item string) bool {
				found = append(found, item)
				return true
			})
			Expect(found).To(ConsistOf("First", "Second"))
		})

		It("stops QueryAABB after the visitor returns false", func() {
			aabb := aabbFromCircle(64.0, 64.0, 63.0)
			count := 0
			tree.QueryAABB(aabb, func(item string) bool {
				count++
				return false // stop after first item
			})
			Expect(count).To(Equal(1))
		})

		When("items are searched", func() {
			BeforeEach(func() {
				aabb := aabbFromCircle(64.0, 64.0, 63.0)
				tree.QueryAABB(aabb, func(item string) bool {
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
					aabbFromCircle(-48.0, 48.0, 2.0),
				)
			})

			It("has the correct state", func() {
				state := tree.Stats()
				Expect(state.NodeCount).To(Equal(uint32(7)))
				Expect(state.ItemCount).To(Equal(uint32(3)))
				Expect(state.ItemCountPerDepth).To(Equal([]uint32{
					0, 0, 3,
				}))
			})

			It("is reflected in segment-search for items", func() {
				from := dprec.NewVec2(1.0, 1.0)
				to := dprec.NewVec2(127.0, 127.0)
				segment := shape2d.NewSegment(from, to)
				var found []string
				tree.QuerySegment(segment, func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(ConsistOf("First"))
			})

			It("is reflected in area-search for items", func() {
				aabb := aabbFromCircle(64.0, 64.0, 63.0)
				var found []string
				tree.QueryAABB(aabb, func(item string) bool {
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

			It("panics when the same item is removed again", func() {
				Expect(func() { tree.Remove(secondItemID) }).To(Panic())
			})

			It("has the correct state", func() {
				state := tree.Stats()
				Expect(state.NodeCount).To(Equal(uint32(5)))
				Expect(state.ItemCount).To(Equal(uint32(2)))
				Expect(state.ItemCountPerDepth).To(Equal([]uint32{
					0, 0, 2,
				}))
			})

			It("does not return an active item id on new insert", func() {
				tree.Stats() // forces internal reordering of items (white box testing)
				secondItemID = tree.Insert(
					aabbFromCircle(48.0, 48.0, 2.0),
					"Second",
				)
				Expect(secondItemID).ToNot(Equal(firstItemID))
				Expect(secondItemID).ToNot(Equal(thirdItemID))
			})

			It("is reflected in segment-search for items", func() {
				from := dprec.NewVec2(1.0, 1.0)
				to := dprec.NewVec2(127.0, 127.0)
				segment := shape2d.NewSegment(from, to)
				var found []string
				tree.QuerySegment(segment, func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(ConsistOf("First"))
			})

			It("is reflected in area-search for items", func() {
				aabb := aabbFromCircle(64.0, 64.0, 63.0)
				var found []string
				tree.QueryAABB(aabb, func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(ConsistOf("First"))
			})
		})
	})

	When("an item is thin along one axis", func() {
		// Both boxes have the same center and the same largest extent, and
		// both descend into the same child node. They differ only along Y,
		// where the node they would descend into next has its center 15 units
		// away. The slab is thin enough along Y to still fit; the block, being
		// as tall as it is wide, is not.
		var (
			slabAABB  = shape2d.NewAABB(-14.0, 32.0, 46.0, 34.0)
			blockAABB = shape2d.NewAABB(-14.0, 3.0, 46.0, 63.0)
		)

		It("descends deeper than a square item of the same largest extent", func() {
			tree.Insert(slabAABB, "Slab")
			state := tree.Stats()
			Expect(state.NodeCount).To(Equal(uint32(3))) // root + child + grandchild
			Expect(state.ItemCountPerDepth).To(Equal([]uint32{
				0, 0, 1,
			}))
		})

		It("keeps a square item at the depth its largest extent allows", func() {
			tree.Insert(blockAABB, "Block")
			state := tree.Stats()
			Expect(state.NodeCount).To(Equal(uint32(2))) // root + child
			Expect(state.ItemCountPerDepth).To(Equal([]uint32{
				0, 1, 0,
			}))
		})

		It("finds both items regardless of the depth they settle at", func() {
			tree.Insert(slabAABB, "Slab")
			tree.Insert(blockAABB, "Block")
			var found []string
			tree.QueryAABB(shape2d.NewAABB(15.0, 33.0, 17.0, 33.0),
				func(item string) bool {
					found = append(found, item)
					return true
				})
			Expect(found).To(ConsistOf("Slab", "Block"))
		})
	})

	When("an item creates a deeply nested branch", func() {
		var deepItemID query2d.TreeItemID

		BeforeEach(func() {
			// A tiny item placed off-center descends to the deepest allowed
			// node, allocating one node per depth level along the way.
			deepItemID = tree.Insert(
				aabbFromCircle(60.0, 60.0, 1.0),
				"Deep",
			)
		})

		It("allocates a node for each depth level", func() {
			state := tree.Stats()
			Expect(state.NodeCount).To(Equal(uint32(3))) // root + child + grandchild
			Expect(state.ItemCount).To(Equal(uint32(1)))
		})

		When("the item is removed", func() {
			BeforeEach(func() {
				tree.Remove(deepItemID)
			})

			It("collapses the whole branch back to the root", func() {
				state := tree.Stats()
				Expect(state.NodeCount).To(Equal(uint32(1)))
				Expect(state.ItemCount).To(Equal(uint32(0)))
			})
		})

		When("the item is moved out of the branch", func() {
			BeforeEach(func() {
				// A large item can no longer fit in any child, so it lands on
				// the root and the vacated branch must collapse.
				tree.Update(deepItemID,
					aabbFromCircle(0.0, 0.0, 60.0),
				)
			})

			It("collapses the vacated branch back to the root", func() {
				state := tree.Stats()
				Expect(state.NodeCount).To(Equal(uint32(1)))
				Expect(state.ItemCount).To(Equal(uint32(1)))
			})
		})
	})

	When("a surviving item shares a branch with a removed item", func() {
		var farItemID query2d.TreeItemID

		BeforeEach(func() {
			// Both items descend into the same branch but into different
			// leaves. Removing the far item must collapse its leaf and shrink
			// the cached bounding boxes of the surviving ancestors.
			tree.Insert(
				aabbFromCircle(16.0, 16.0, 2.0),
				"Near",
			)
			farItemID = tree.Insert(
				aabbFromCircle(60.0, 60.0, 1.0),
				"Far",
			)
			// Settle the tree so every cached box is clean. Only the collapse
			// triggered by the removal below may dirty the surviving ancestors.
			tree.Stats()
			tree.Remove(farItemID)
		})

		It("collapses the cached bounding boxes towards the surviving item", func() {
			// The query targets the location the removed item used to occupy.
			// If the ancestor boxes were left stale, traversal would be accepted
			// into them; with the boxes collapsed, it is rejected at the root.
			var found []string
			tree.QueryAABB(
				aabbFromCircle(60.0, 60.0, 1.0),
				func(item string) bool {
					found = append(found, item)
					return true
				},
			)
			Expect(found).To(BeEmpty())

			stats := tree.VisitStats()
			Expect(stats.NodeCountAccepted).To(Equal(uint32(0)))
			Expect(stats.NodeCountRejected).To(Equal(uint32(1)))
		})

		It("still finds the surviving item", func() {
			var found []string
			tree.QueryAABB(
				aabbFromCircle(16.0, 16.0, 2.0),
				func(item string) bool {
					found = append(found, item)
					return true
				},
			)
			Expect(found).To(ConsistOf("Near"))
		})
	})

	When("the tree undergoes heavy churn", func() {
		It("keeps queries and stats consistent", func() {
			const count = 200
			ids := make([]query2d.TreeItemID, count)
			expected := make(map[query2d.TreeItemID]string, count)

			positionFor := func(i int) shape2d.AABB {
				x := float64(-60 + (i*7)%120)
				y := float64(-60 + (i*13)%120)
				return aabbFromCircle(x, y, 1.0)
			}

			// Populate the tree.
			for i := range count {
				value := fmt.Sprintf("item-%d", i)
				ids[i] = tree.Insert(positionFor(i), value)
				expected[ids[i]] = value
			}

			// Churn: drop every third item and relocate half of the rest.
			for i := range count {
				switch {
				case i%3 == 0:
					tree.Remove(ids[i])
					delete(expected, ids[i])
				case i%2 == 0:
					tree.Update(ids[i], positionFor(i+1))
				}
			}

			// Re-insert into the freed slots to exercise item/node reuse.
			for i := 0; i < count; i += 3 {
				value := fmt.Sprintf("reinsert-%d", i)
				id := tree.Insert(positionFor(i), value)
				expected[id] = value
			}

			// A query covering the whole tree must return exactly the items
			// we expect to still be present.
			found := make(map[string]struct{})
			tree.QueryAABB(aabbFromCircle(0.0, 0.0, 1000.0), func(item string) bool {
				found[item] = struct{}{}
				return true
			})

			Expect(found).To(HaveLen(len(expected)))
			for _, value := range expected {
				Expect(found).To(HaveKey(value))
			}
			Expect(tree.Stats().ItemCount).To(Equal(uint32(len(expected))))
		})
	})
})
