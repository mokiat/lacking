package query3d_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/query3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// areaFromSphere builds an Area from sphere center coordinates and radius.
func areaFromSphere(x, y, z, radius float64) query3d.Area {
	return query3d.AreaFromSphere(shape3d.Sphere{
		Center: dprec.NewVec3(x, y, z),
		Radius: radius,
	})
}

// aabbFromSphere builds an AABB enclosing a sphere with the given center
// coordinates and radius.
func aabbFromSphere(x, y, z, radius float64) query3d.AABB {
	return query3d.AABBFromSphere(shape3d.Sphere{
		Center: dprec.NewVec3(x, y, z),
		Radius: radius,
	})
}

var _ = Describe("Octree", func() {
	var (
		tree *query3d.Octree[string]
	)

	BeforeEach(func() {
		tree = query3d.NewOctree[string](query3d.OctreeSettings{
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
			firstItemID  query3d.TreeItemID
			secondItemID query3d.TreeItemID
			thirdItemID  query3d.TreeItemID
		)

		BeforeEach(func() {
			firstItemID = tree.Insert(
				areaFromSphere(16.0, 16.0, 16.0, 2.0),
				"First",
			)
			secondItemID = tree.Insert(
				areaFromSphere(48.0, 48.0, 48.0, 2.0),
				"Second",
			)
			thirdItemID = tree.Insert(
				areaFromSphere(-16.0, -48.0, -16.0, 32.0),
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
			segment := query3d.NewSegment(from, to)
			var found []string
			tree.QuerySegment(segment, func(item string) bool {
				found = append(found, item)
				return true
			})
			Expect(found).To(ConsistOf("First", "Second"))
		})

		It("stops QuerySegment after the visitor returns false", func() {
			from := dprec.NewVec3(1.0, 1.0, 1.0)
			to := dprec.NewVec3(127.0, 127.0, 127.0)
			segment := query3d.NewSegment(from, to)
			count := 0
			tree.QuerySegment(segment, func(item string) bool {
				count++
				return false // stop after first item
			})
			Expect(count).To(Equal(1))
		})

		It("is possible to area-search for items", func() {
			aabb := aabbFromSphere(64.0, 64.0, 64.0, 63.0)
			var found []string
			tree.QueryAABB(aabb, func(item string) bool {
				found = append(found, item)
				return true
			})
			Expect(found).To(ConsistOf("First", "Second"))
		})

		It("stops QueryAABB after the visitor returns false", func() {
			aabb := aabbFromSphere(64.0, 64.0, 64.0, 63.0)
			count := 0
			tree.QueryAABB(aabb, func(item string) bool {
				count++
				return false // stop after first item
			})
			Expect(count).To(Equal(1))
		})

		When("items are searched", func() {
			BeforeEach(func() {
				aabb := aabbFromSphere(64.0, 64.0, 64.0, 63.0)
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
					areaFromSphere(-48.0, 48.0, -48.0, 2.0),
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
				segment := query3d.NewSegment(from, to)
				var found []string
				tree.QuerySegment(segment, func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(ConsistOf("First"))
			})

			It("is reflected in area-search for items", func() {
				aabb := aabbFromSphere(64.0, 64.0, 64.0, 63.0)
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
					areaFromSphere(48.0, 48.0, 48.0, 2.0),
					"Second",
				)
				Expect(secondItemID).ToNot(Equal(firstItemID))
				Expect(secondItemID).ToNot(Equal(thirdItemID))
			})

			It("is reflected in segment-search for items", func() {
				from := dprec.NewVec3(1.0, 1.0, 1.0)
				to := dprec.NewVec3(127.0, 127.0, 127.0)
				segment := query3d.NewSegment(from, to)
				var found []string
				tree.QuerySegment(segment, func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(ConsistOf("First"))
			})

			It("is reflected in area-search for items", func() {
				aabb := aabbFromSphere(64.0, 64.0, 64.0, 63.0)
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
		var deepItemID query3d.TreeItemID

		BeforeEach(func() {
			// A tiny item placed off-center descends to the deepest allowed
			// node, allocating one node per depth level along the way.
			deepItemID = tree.Insert(
				areaFromSphere(60.0, 60.0, 60.0, 1.0),
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
					areaFromSphere(0.0, 0.0, 0.0, 60.0),
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
		var farItemID query3d.TreeItemID

		BeforeEach(func() {
			// Both items descend into the same branch but into different
			// leaves. Removing the far item must collapse its leaf and shrink
			// the cached bounding boxes of the surviving ancestors.
			tree.Insert(
				areaFromSphere(16.0, 16.0, 16.0, 2.0),
				"Near",
			)
			farItemID = tree.Insert(
				areaFromSphere(60.0, 60.0, 60.0, 1.0),
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
				aabbFromSphere(60.0, 60.0, 60.0, 1.0),
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
				aabbFromSphere(16.0, 16.0, 16.0, 2.0),
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
			ids := make([]query3d.TreeItemID, count)
			expected := make(map[query3d.TreeItemID]string, count)

			positionFor := func(i int) query3d.Area {
				x := float64(-60 + (i*7)%120)
				y := float64(-60 + (i*13)%120)
				z := float64(-60 + (i*5)%120)
				return areaFromSphere(x, y, z, 1.0)
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
			tree.QueryAABB(aabbFromSphere(0.0, 0.0, 0.0, 1000.0), func(item string) bool {
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
