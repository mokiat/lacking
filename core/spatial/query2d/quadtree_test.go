package query2d_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/query2d"
)

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

	When("items are inserted", func() {
		var (
			firstItemID  query2d.TreeItemID
			secondItemID query2d.TreeItemID
			thirdItemID  query2d.TreeItemID
		)

		BeforeEach(func() {
			firstItemID = tree.Insert(
				query2d.AreaFromCircle(16.0, 16.0, 2.0),
				"First",
			)
			secondItemID = tree.Insert(
				query2d.AreaFromCircle(48.0, 48.0, 2.0),
				"Second",
			)
			thirdItemID = tree.Insert(
				query2d.AreaFromCircle(-16.0, -48.0, 32.0),
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
			from := dprec.NewVec2(1.0, 1.0)
			to := dprec.NewVec2(127.0, 127.0)
			segment := query2d.NewSegment(from, to)
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
			segment := query2d.NewSegment(from, to)
			count := 0
			tree.QuerySegment(segment, func(item string) bool {
				count++
				return false // stop after first item
			})
			Expect(count).To(Equal(1))
		})

		It("is possible to area-search for items", func() {
			aabb := query2d.AABBFromCircle(64.0, 64.0, 63.0)
			var found []string
			tree.QueryAABB(aabb, func(item string) bool {
				found = append(found, item)
				return true
			})
			Expect(found).To(ConsistOf("First", "Second"))
		})

		It("stops QueryAABB after the visitor returns false", func() {
			aabb := query2d.AABBFromCircle(64.0, 64.0, 63.0)
			count := 0
			tree.QueryAABB(aabb, func(item string) bool {
				count++
				return false // stop after first item
			})
			Expect(count).To(Equal(1))
		})

		When("items are searched", func() {
			BeforeEach(func() {
				aabb := query2d.AABBFromCircle(64.0, 64.0, 63.0)
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
					query2d.AreaFromCircle(-48.0, 48.0, 2.0),
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
				from := dprec.NewVec2(1.0, 1.0)
				to := dprec.NewVec2(127.0, 127.0)
				segment := query2d.NewSegment(from, to)
				var found []string
				tree.QuerySegment(segment, func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(ConsistOf("First"))
			})

			It("is reflected in area-search for items", func() {
				aabb := query2d.AABBFromCircle(64.0, 64.0, 63.0)
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
				Expect(state.NodeCount).To(Equal(uint32(4)))
				Expect(state.ItemCount).To(Equal(uint32(2)))
				Expect(state.ItemCountPerDepth).To(Equal([]uint32{
					0, 1, 1,
				}))
			})

			It("does not return an active item id on new insert", func() {
				tree.Stats() // forces internal reordering of items (white box testing)
				secondItemID = tree.Insert(
					query2d.AreaFromCircle(48.0, 48.0, 2.0),
					"Second",
				)
				Expect(secondItemID).ToNot(Equal(firstItemID))
				Expect(secondItemID).ToNot(Equal(thirdItemID))
			})

			It("is reflected in segment-search for items", func() {
				from := dprec.NewVec2(1.0, 1.0)
				to := dprec.NewVec2(127.0, 127.0)
				segment := query2d.NewSegment(from, to)
				var found []string
				tree.QuerySegment(segment, func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(ConsistOf("First"))
			})

			It("is reflected in area-search for items", func() {
				aabb := query2d.AABBFromCircle(64.0, 64.0, 63.0)
				var found []string
				tree.QueryAABB(aabb, func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(ConsistOf("First"))
			})
		})
	})

	When("an item creates a deeply nested branch", func() {
		var deepItemID query2d.TreeItemID

		BeforeEach(func() {
			// A tiny item placed off-center descends to the deepest allowed
			// node, allocating one node per depth level along the way.
			deepItemID = tree.Insert(
				query2d.AreaFromCircle(60.0, 60.0, 1.0),
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
					query2d.AreaFromCircle(0.0, 0.0, 60.0),
				)
			})

			It("collapses the vacated branch back to the root", func() {
				state := tree.Stats()
				Expect(state.NodeCount).To(Equal(uint32(1)))
				Expect(state.ItemCount).To(Equal(uint32(1)))
			})
		})
	})

	When("the tree undergoes heavy churn", func() {
		It("keeps queries and stats consistent", func() {
			const count = 200
			ids := make([]query2d.TreeItemID, count)
			expected := make(map[query2d.TreeItemID]string, count)

			positionFor := func(i int) query2d.Area {
				x := float64(-60 + (i*7)%120)
				y := float64(-60 + (i*13)%120)
				return query2d.AreaFromCircle(x, y, 1.0)
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
			tree.QueryAABB(query2d.AABBFromCircle(0.0, 0.0, 1000.0), func(item string) bool {
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
