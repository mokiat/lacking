package query2d_test

import (
	"fmt"
	"math/rand/v2"
	"slices"

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

var _ = Describe("BVHTree", func() {
	var (
		tree *query2d.BVHTree[string]
	)

	collectAABB := func(aabb shape2d.AABB) []string {
		var found []string
		tree.QueryAABB(aabb, func(item string) bool {
			found = append(found, item)
			return true
		})
		return found
	}

	collectSegment := func(from, to dprec.Vec2) []string {
		var found []string
		tree.QuerySegment(shape2d.NewSegment(from, to), func(item string) bool {
			found = append(found, item)
			return true
		})
		return found
	}

	collectFrustum := func(minX, minY, maxX, maxY float64) []string {
		frustum := shape2d.FrustumFromProjection(dprec.OrthoMat3(minX, maxX, maxY, minY))
		var found []string
		tree.QueryFrustum(frustum, func(item string) bool {
			found = append(found, item)
			return true
		})
		return found
	}

	BeforeEach(func() {
		tree = query2d.NewBVHTree[string](query2d.BVHTreeSettings{})
	})

	It("has the correct initial state", func() {
		state := tree.Stats()
		Expect(state.NodeCount).To(Equal(uint32(0)))
		Expect(state.ItemCount).To(Equal(uint32(0)))
		Expect(state.ItemCountPerDepth).To(BeEmpty())
	})

	It("is possible to query an empty tree", func() {
		Expect(collectAABB(aabbFromCircle(0.0, 0.0, 100.0))).To(BeEmpty())
		Expect(collectSegment(dprec.NewVec2(-10.0, -10.0), dprec.NewVec2(10.0, 10.0))).To(BeEmpty())
		Expect(collectFrustum(-10.0, -10.0, 10.0, 10.0)).To(BeEmpty())
	})

	It("panics when an item with an empty box is inserted", func() {
		emptyAABB := shape2d.NewAABB(1.0, 1.0, -1.0, -1.0)
		Expect(func() {
			tree.Insert(emptyAABB, "Empty")
		}).To(Panic())
	})

	It("panics when an item is updated to an empty box", func() {
		emptyAABB := shape2d.NewAABB(1.0, 1.0, -1.0, -1.0)
		itemID := tree.Insert(aabbFromCircle(0.0, 0.0, 1.0), "Item")
		Expect(func() {
			tree.Update(itemID, emptyAABB)
		}).To(Panic())
	})

	It("panics when the margin is negative", func() {
		Expect(func() {
			query2d.NewBVHTree[string](query2d.BVHTreeSettings{
				AABBMargin: opt.V(-1.0),
			})
		}).To(Panic())
	})

	When("a single item is inserted", func() {
		var itemID query2d.TreeItemID

		BeforeEach(func() {
			itemID = tree.Insert(aabbFromCircle(5.0, 5.0, 1.0), "Only")
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
				newID := tree.Insert(aabbFromCircle(5.0, 5.0, 1.0), "Again")
				Expect(collectAABB(aabbFromCircle(5.0, 5.0, 1.0))).To(ConsistOf("Again"))
				Expect(newID).To(Equal(itemID)) // the freed slot is reused
			})
		})

		When("the item is moved far away", func() {
			BeforeEach(func() {
				tree.Update(itemID, aabbFromCircle(500.0, 500.0, 1.0))
			})

			It("is reflected in searches", func() {
				Expect(collectAABB(aabbFromCircle(5.0, 5.0, 1.0))).To(BeEmpty())
				Expect(collectAABB(aabbFromCircle(500.0, 500.0, 1.0))).To(ConsistOf("Only"))
			})
		})
	})

	When("items are inserted", func() {
		var (
			firstItemID  query2d.TreeItemID
			secondItemID query2d.TreeItemID
			thirdItemID  query2d.TreeItemID
		)

		BeforeEach(func() {
			firstItemID = tree.Insert(aabbFromCircle(16.0, 16.0, 2.0), "First")
			secondItemID = tree.Insert(aabbFromCircle(48.0, 48.0, 2.0), "Second")
			thirdItemID = tree.Insert(aabbFromCircle(-16.0, -48.0, 32.0), "Third")
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
			Expect(sumOf(state.ItemCountPerDepth)).To(Equal(uint32(3)))
		})

		It("is possible to segment-search for items", func() {
			found := collectSegment(dprec.NewVec2(1.0, 1.0), dprec.NewVec2(127.0, 127.0))
			Expect(found).To(ConsistOf("First", "Second"))
		})

		It("stops QuerySegment after the visitor returns false", func() {
			segment := shape2d.NewSegment(dprec.NewVec2(1.0, 1.0), dprec.NewVec2(127.0, 127.0))
			count := 0
			tree.QuerySegment(segment, func(item string) bool {
				count++
				return false // stop after first item
			})
			Expect(count).To(Equal(1))
		})

		It("is possible to area-search for items", func() {
			Expect(collectAABB(aabbFromCircle(64.0, 64.0, 63.0))).To(ConsistOf("First", "Second"))
		})

		It("stops QueryAABB after the visitor returns false", func() {
			count := 0
			tree.QueryAABB(aabbFromCircle(64.0, 64.0, 63.0), func(item string) bool {
				count++
				return false // stop after first item
			})
			Expect(count).To(Equal(1))
		})

		It("is possible to frustum-search for items", func() {
			Expect(collectFrustum(1.0, 1.0, 127.0, 127.0)).To(ConsistOf("First", "Second"))
			Expect(collectFrustum(-100.0, -100.0, 100.0, 100.0)).To(ConsistOf("First", "Second", "Third"))
			Expect(collectFrustum(200.0, 200.0, 300.0, 300.0)).To(BeEmpty())
		})

		It("stops QueryFrustum after the visitor returns false", func() {
			frustum := shape2d.FrustumFromAABB(shape2d.NewAABB(-100.0, -100.0, 100.0, 100.0))
			count := 0
			tree.QueryFrustum(frustum, func(item string) bool {
				count++
				return false // stop after first item
			})
			Expect(count).To(Equal(1))
		})

		It("supports a query issued from within the visitor", func() {
			var outer, inner []string
			tree.QueryAABB(aabbFromCircle(64.0, 64.0, 63.0), func(item string) bool {
				outer = append(outer, item)
				tree.QueryAABB(aabbFromCircle(16.0, 16.0, 2.0), func(item string) bool {
					inner = append(inner, item)
					return true
				})
				return true
			})
			Expect(outer).To(ConsistOf("First", "Second"))
			Expect(inner).ToNot(BeEmpty())
		})

		It("returns consistent visit stats", func() {
			collectAABB(aabbFromCircle(64.0, 64.0, 63.0))
			stats := tree.VisitStats()
			Expect(stats.NodeCountVisited).To(Equal(stats.NodeCountAccepted + stats.NodeCountRejected))
			Expect(stats.ItemCountVisited).To(Equal(stats.ItemCountAccepted + stats.ItemCountRejected))
			Expect(stats.NodeCountVisited).To(BeNumerically(">", uint32(0)))
			Expect(stats.NodeCountVisited).To(BeNumerically("<=", uint32(5)))
			Expect(stats.ItemCountAccepted).To(Equal(uint32(2)))
		})

		When("an item is updated", func() {
			BeforeEach(func() {
				tree.Update(secondItemID, aabbFromCircle(-48.0, 48.0, 2.0))
			})

			It("has the correct state", func() {
				state := tree.Stats()
				Expect(state.NodeCount).To(Equal(uint32(5)))
				Expect(state.ItemCount).To(Equal(uint32(3)))
			})

			It("is reflected in searches", func() {
				Expect(collectAABB(aabbFromCircle(64.0, 64.0, 63.0))).To(ConsistOf("First"))
				Expect(collectAABB(aabbFromCircle(-48.0, 48.0, 1.0))).To(ConsistOf("Second"))
			})
		})

		When("an item is removed", func() {
			BeforeEach(func() {
				tree.Remove(secondItemID)
			})

			It("panics when the same item is removed again", func() {
				Expect(func() {
					tree.Remove(secondItemID)
				}).To(Panic())
			})

			It("panics when the same item is updated", func() {
				Expect(func() {
					tree.Update(secondItemID, aabbFromCircle(0.0, 0.0, 1.0))
				}).To(Panic())
			})

			It("has the correct state", func() {
				state := tree.Stats()
				Expect(state.NodeCount).To(Equal(uint32(3)))
				Expect(state.ItemCount).To(Equal(uint32(2)))
			})

			It("reuses the id on a new insert", func() {
				newID := tree.Insert(aabbFromCircle(0.0, 0.0, 1.0), "Fourth")
				Expect(newID).To(Equal(secondItemID))
				Expect(newID).ToNot(Equal(firstItemID))
				Expect(newID).ToNot(Equal(thirdItemID))
			})

			It("is reflected in searches", func() {
				Expect(collectAABB(aabbFromCircle(64.0, 64.0, 63.0))).To(ConsistOf("First"))
				found := collectSegment(dprec.NewVec2(1.0, 1.0), dprec.NewVec2(127.0, 127.0))
				Expect(found).To(ConsistOf("First"))
			})
		})
	})

	When("an item moves within its margin", func() {
		var itemID query2d.TreeItemID

		BeforeEach(func() {
			tree = query2d.NewBVHTree[string](query2d.BVHTreeSettings{
				AABBMargin: opt.V(10.0),
			})
			tree.Insert(aabbFromCircle(-100.0, -100.0, 1.0), "Anchor")
			itemID = tree.Insert(aabbFromCircle(0.0, 0.0, 1.0), "Mover")
			tree.Update(itemID, aabbFromCircle(1.0, 1.0, 1.0))
		})

		It("keeps the state of the tree", func() {
			state := tree.Stats()
			Expect(state.NodeCount).To(Equal(uint32(3)))
			Expect(state.ItemCount).To(Equal(uint32(2)))
		})

		It("still reports the exact new position", func() {
			Expect(collectAABB(aabbFromCircle(-3.0, -3.0, 1.0))).To(BeEmpty())
			Expect(collectAABB(aabbFromCircle(3.0, 3.0, 1.0))).To(ConsistOf("Mover"))
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

			for i := range count {
				value := fmt.Sprintf("item-%d", i)
				ids[i] = tree.Insert(positionFor(i), value)
				expected[ids[i]] = value
			}

			// Drop every third item and relocate half of the rest.
			for i := range count {
				switch {
				case i%3 == 0:
					tree.Remove(ids[i])
					delete(expected, ids[i])
				case i%2 == 0:
					tree.Update(ids[i], positionFor(i+1))
				}
			}

			// Insert into the freed slots to exercise item and node reuse.
			for i := 0; i < count; i += 3 {
				value := fmt.Sprintf("reinsert-%d", i)
				id := tree.Insert(positionFor(i), value)
				expected[id] = value
			}

			found := collectAABB(aabbFromCircle(0.0, 0.0, 1000.0))
			Expect(found).To(ConsistOf(slices.Collect(func(yield func(string) bool) {
				for _, value := range expected {
					if !yield(value) {
						return
					}
				}
			})))
			state := tree.Stats()
			Expect(state.ItemCount).To(Equal(uint32(len(expected))))
			Expect(state.NodeCount).To(Equal(uint32(2*len(expected) - 1)))
		})
	})
})

var _ = Describe("BVHTree and Bag equivalence", func() {
	const (
		itemCount  = 500
		queryCount = 200
		sceneSize  = 1024.0
	)

	// A bag tests every single item, so it acts as the reference that the tree
	// has to match exactly, both before and after its structure has changed.
	It("return identical results for the same queries", func() {
		random := rand.New(rand.NewPCG(0xBEEF, 0xCAFE))

		randomCoord := func() float64 {
			return random.Float64()*2.0*sceneSize - sceneSize
		}
		randomBox := func() shape2d.AABB {
			return aabbFromCircle(randomCoord(), randomCoord(), random.Float64()*8.0+0.5)
		}

		bag := query2d.NewBag[int](query2d.BagSettings{})
		tree := query2d.NewBVHTree[int](query2d.BVHTreeSettings{
			AABBMargin: opt.V(1.0),
		})

		bagIDs := make([]query2d.BagItemID, 0, itemCount)
		treeIDs := make([]query2d.TreeItemID, 0, itemCount)
		for i := range itemCount {
			box := randomBox()
			bagIDs = append(bagIDs, bag.Insert(box, i))
			treeIDs = append(treeIDs, tree.Insert(box, i))
		}

		collect := func(query func(yield query2d.VisitorFunc[int])) []int {
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
				aabb := aabbFromCircle(randomCoord(), randomCoord(), random.Float64()*128.0+1.0)
				expected := collect(func(yield query2d.VisitorFunc[int]) {
					bag.QueryAABB(aabb, yield)
				})
				actual := collect(func(yield query2d.VisitorFunc[int]) {
					tree.QueryAABB(aabb, yield)
				})
				Expect(actual).To(Equal(expected), "%s: aabb query mismatch", stage)
			}

			for range queryCount {
				segment := shape2d.NewSegment(
					dprec.NewVec2(randomCoord(), randomCoord()),
					dprec.NewVec2(randomCoord(), randomCoord()),
				)
				expected := collect(func(yield query2d.VisitorFunc[int]) {
					bag.QuerySegment(segment, yield)
				})
				actual := collect(func(yield query2d.VisitorFunc[int]) {
					tree.QuerySegment(segment, yield)
				})
				Expect(actual).To(Equal(expected), "%s: segment query mismatch", stage)
			}

			for range queryCount {
				// A rotated rectangular view somewhere in the scene.
				halfSize := random.Float64()*256.0 + 1.0
				frustum := shape2d.FrustumFromProjection(dprec.Mat3Prod(
					dprec.OrthoMat3(-halfSize, halfSize, halfSize, -halfSize),
					dprec.InverseMat3(dprec.Mat3Prod(
						dprec.TranslationMat3(randomCoord(), randomCoord()),
						dprec.RotationMat3(dprec.Degrees(random.Float64()*360.0)),
					)),
				))
				expected := collect(func(yield query2d.VisitorFunc[int]) {
					bag.QueryFrustum(frustum, yield)
				})
				actual := collect(func(yield query2d.VisitorFunc[int]) {
					tree.QueryFrustum(frustum, yield)
				})
				Expect(actual).To(Equal(expected), "%s: frustum query mismatch", stage)
			}
		}

		compare("after insert")

		// Change both structures identically and compare again.
		for i := range itemCount {
			switch i % 3 {
			case 0:
				bag.Remove(bagIDs[i])
				tree.Remove(treeIDs[i])
			case 1:
				box := randomBox()
				bag.Update(bagIDs[i], box)
				tree.Update(treeIDs[i], box)
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
