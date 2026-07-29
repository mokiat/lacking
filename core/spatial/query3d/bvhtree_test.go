package query3d_test

import (
	"fmt"
	"math/rand/v2"
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/query3d"
)

// tree captures the API contract that an [query3d.Octree] and a
// [query3d.BVHTree] share. The assertions below fail to compile if the two
// ever drift apart, which is what makes one a drop-in replacement for the
// other.
type tree interface {
	Stats() query3d.TreeStats
	VisitStats() query3d.TreeVisitStats
	Insert(area query3d.Area, value string) query3d.TreeItemID
	Update(id query3d.TreeItemID, area query3d.Area)
	Remove(id query3d.TreeItemID)
	QuerySegment(segment query3d.Segment, yield query3d.VisitorFunc[string])
	QueryAABB(aabb query3d.AABB, yield query3d.VisitorFunc[string])
}

var (
	_ tree = (*query3d.Octree[string])(nil)
	_ tree = (*query3d.BVHTree[string])(nil)
)

var _ = Describe("BVHTree", func() {
	var (
		tree *query3d.BVHTree[string]
	)

	BeforeEach(func() {
		tree = query3d.NewBVHTree[string](query3d.BVHTreeSettings{})
	})

	It("has the correct initial state", func() {
		state := tree.Stats()
		Expect(state.NodeCount).To(Equal(uint32(0))) // unlike an octree, there is no root
		Expect(state.ItemCount).To(Equal(uint32(0)))
		Expect(state.ItemCountPerDepth).To(BeEmpty())
	})

	It("is possible to query an empty tree", func() {
		var found []string
		tree.QueryAABB(aabbFromSphere(0.0, 0.0, 0.0, 100.0), func(item string) bool {
			found = append(found, item)
			return true
		})
		Expect(found).To(BeEmpty())

		segment := query3d.NewSegment(
			dprec.NewVec3(-10.0, -10.0, -10.0),
			dprec.NewVec3(10.0, 10.0, 10.0),
		)
		tree.QuerySegment(segment, func(item string) bool {
			found = append(found, item)
			return true
		})
		Expect(found).To(BeEmpty())
	})

	When("a single item is inserted", func() {
		var itemID query3d.TreeItemID

		BeforeEach(func() {
			itemID = tree.Insert(areaFromSphere(5.0, 5.0, 5.0, 1.0), "Only")
		})

		It("has the correct state", func() {
			state := tree.Stats()
			Expect(state.NodeCount).To(Equal(uint32(1))) // the leaf is the root
			Expect(state.ItemCount).To(Equal(uint32(1)))
			Expect(state.ItemCountPerDepth).To(Equal([]uint32{1}))
		})

		When("the item is removed", func() {
			BeforeEach(func() {
				tree.Remove(itemID)
			})

			It("returns to the empty state", func() {
				state := tree.Stats()
				Expect(state.NodeCount).To(Equal(uint32(0)))
				Expect(state.ItemCount).To(Equal(uint32(0)))
			})

			It("is possible to insert again", func() {
				newID := tree.Insert(areaFromSphere(5.0, 5.0, 5.0, 1.0), "Again")
				var found []string
				tree.QueryAABB(aabbFromSphere(5.0, 5.0, 5.0, 1.0), func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(ConsistOf("Again"))
				Expect(newID).To(Equal(itemID)) // the freed slot is reused
			})
		})

		When("the item is moved far away", func() {
			BeforeEach(func() {
				tree.Update(itemID, areaFromSphere(500.0, 500.0, 500.0, 1.0))
			})

			It("is reflected in searches", func() {
				var found []string
				tree.QueryAABB(aabbFromSphere(5.0, 5.0, 5.0, 1.0), func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(BeEmpty())

				tree.QueryAABB(aabbFromSphere(500.0, 500.0, 500.0, 1.0), func(item string) bool {
					found = append(found, item)
					return true
				})
				Expect(found).To(ConsistOf("Only"))
			})
		})
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
			// A tree of N items always has exactly 2*N-1 nodes.
			Expect(state.NodeCount).To(Equal(uint32(5)))
			Expect(state.ItemCount).To(Equal(uint32(3)))
			// The exact distribution depends on the surface area heuristic, so
			// only the total is asserted here.
			Expect(sumOf(state.ItemCountPerDepth)).To(Equal(uint32(3)))
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

		It("supports a query issued from within the visitor", func() {
			var outer, inner []string
			tree.QueryAABB(aabbFromSphere(64.0, 64.0, 64.0, 63.0), func(item string) bool {
				outer = append(outer, item)
				tree.QueryAABB(aabbFromSphere(16.0, 16.0, 16.0, 2.0), func(item string) bool {
					inner = append(inner, item)
					return true
				})
				return true
			})
			Expect(outer).To(ConsistOf("First", "Second"))
			Expect(inner).ToNot(BeEmpty())
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
				Expect(stats.NodeCountVisited).To(Equal(
					stats.NodeCountAccepted + stats.NodeCountRejected,
				))
				Expect(stats.ItemCountVisited).To(Equal(
					stats.ItemCountAccepted + stats.ItemCountRejected,
				))
				Expect(stats.NodeCountVisited).To(BeNumerically(">", uint32(0)))
				Expect(stats.NodeCountVisited).To(BeNumerically("<=", uint32(5)))
				Expect(stats.ItemCountAccepted).To(Equal(uint32(2)))
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
				Expect(state.NodeCount).To(Equal(uint32(5)))
				Expect(state.ItemCount).To(Equal(uint32(3)))
				Expect(sumOf(state.ItemCountPerDepth)).To(Equal(uint32(3)))
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

			It("panics when the same item is updated", func() {
				Expect(func() {
					tree.Update(secondItemID, areaFromSphere(0.0, 0.0, 0.0, 1.0))
				}).To(Panic())
			})

			It("has the correct state", func() {
				state := tree.Stats()
				Expect(state.NodeCount).To(Equal(uint32(3)))
				Expect(state.ItemCount).To(Equal(uint32(2)))
				Expect(sumOf(state.ItemCountPerDepth)).To(Equal(uint32(2)))
			})

			It("does not return an active item id on new insert", func() {
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

	When("an item moves within its margin", func() {
		var itemID query3d.TreeItemID

		BeforeEach(func() {
			tree = query3d.NewBVHTree[string](query3d.BVHTreeSettings{
				AABBMargin: opt.V(10.0),
			})
			tree.Insert(areaFromSphere(-100.0, -100.0, -100.0, 1.0), "Anchor")
			itemID = tree.Insert(areaFromSphere(0.0, 0.0, 0.0, 1.0), "Mover")
			tree.Update(itemID, areaFromSphere(1.0, 1.0, 1.0, 1.0))
		})

		It("leaves the structure of the tree untouched", func() {
			state := tree.Stats()
			Expect(state.NodeCount).To(Equal(uint32(3)))
			Expect(state.ItemCount).To(Equal(uint32(2)))
		})

		It("still reports the exact new position", func() {
			// The grown box of the leaf still covers the old position, but the
			// exact box of the item is always re-tested before it is yielded.
			var found []string
			tree.QueryAABB(aabbFromSphere(-3.0, -3.0, -3.0, 1.0), func(item string) bool {
				found = append(found, item)
				return true
			})
			Expect(found).To(BeEmpty())

			tree.QueryAABB(aabbFromSphere(3.0, 3.0, 3.0, 1.0), func(item string) bool {
				found = append(found, item)
				return true
			})
			Expect(found).To(ConsistOf("Mover"))
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

var _ = Describe("BVHTree and Octree equivalence", func() {
	const (
		itemCount  = 500
		queryCount = 200
		sceneSize  = 1024.0
	)

	// Both trees reduce an item to the exact same box and run the exact same
	// test on it before yielding it, so they must return identical results for
	// any query, despite having completely different internal structures.
	It("return identical results for the same queries", func() {
		random := rand.New(rand.NewPCG(0xBEEF, 0xCAFE))

		randomArea := func() query3d.Area {
			return areaFromSphere(
				random.Float64()*2.0*sceneSize-sceneSize,
				random.Float64()*2.0*sceneSize-sceneSize,
				random.Float64()*2.0*sceneSize-sceneSize,
				random.Float64()*8.0+0.5,
			)
		}
		randomPoint := func() dprec.Vec3 {
			return dprec.NewVec3(
				random.Float64()*2.0*sceneSize-sceneSize,
				random.Float64()*2.0*sceneSize-sceneSize,
				random.Float64()*2.0*sceneSize-sceneSize,
			)
		}

		octree := query3d.NewOctree[int](query3d.OctreeSettings{
			Size:     opt.V(4.0 * sceneSize),
			MaxDepth: opt.V[uint32](8),
		})
		bvhTree := query3d.NewBVHTree[int](query3d.BVHTreeSettings{
			AABBMargin: opt.V(1.0),
		})

		octreeIDs := make([]query3d.TreeItemID, 0, itemCount)
		bvhTreeIDs := make([]query3d.TreeItemID, 0, itemCount)
		for i := range itemCount {
			area := randomArea()
			octreeIDs = append(octreeIDs, octree.Insert(area, i))
			bvhTreeIDs = append(bvhTreeIDs, bvhTree.Insert(area, i))
		}

		collect := func(query func(yield query3d.VisitorFunc[int])) []int {
			var result []int
			query(func(item int) bool {
				result = append(result, item)
				return true
			})
			slices.Sort(result)
			return result
		}

		compare := func(stage string) {
			for range queryCount {
				aabb := aabbFromSphere(
					random.Float64()*2.0*sceneSize-sceneSize,
					random.Float64()*2.0*sceneSize-sceneSize,
					random.Float64()*2.0*sceneSize-sceneSize,
					random.Float64()*128.0+1.0,
				)
				expected := collect(func(yield query3d.VisitorFunc[int]) {
					octree.QueryAABB(aabb, yield)
				})
				actual := collect(func(yield query3d.VisitorFunc[int]) {
					bvhTree.QueryAABB(aabb, yield)
				})
				Expect(actual).To(Equal(expected), "%s: aabb query mismatch", stage)
			}

			for range queryCount {
				segment := query3d.NewSegment(randomPoint(), randomPoint())
				expected := collect(func(yield query3d.VisitorFunc[int]) {
					octree.QuerySegment(segment, yield)
				})
				actual := collect(func(yield query3d.VisitorFunc[int]) {
					bvhTree.QuerySegment(segment, yield)
				})
				Expect(actual).To(Equal(expected), "%s: segment query mismatch", stage)
			}
		}

		compare("after insert")

		// Churn both trees identically and compare again.
		for i := range itemCount {
			switch {
			case i%3 == 0:
				octree.Remove(octreeIDs[i])
				bvhTree.Remove(bvhTreeIDs[i])
			case i%3 == 1:
				area := randomArea()
				octree.Update(octreeIDs[i], area)
				bvhTree.Update(bvhTreeIDs[i], area)
			}
		}

		compare("after churn")
	})
})

func sumOf(values []uint32) uint32 {
	var result uint32
	for _, value := range values {
		result += value
	}
	return result
}
