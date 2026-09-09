package query3d_test

import (
	"fmt"
	"math/rand/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/query3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// aabbFromSphere builds an AABB enclosing a sphere with the given center
// coordinates and radius.
func aabbFromSphere(x, y, z, radius float64) shape3d.AABB {
	return shape3d.AABBFromSphere(shape3d.Sphere{
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

	It("panics when an item with an empty box is inserted", func() {
		emptyAABB := shape3d.NewAABB(1.0, 1.0, 1.0, -1.0, -1.0, -1.0)
		Expect(func() { tree.Insert(emptyAABB, "Empty") }).To(Panic())
	})

	It("panics when an item is updated to an empty box", func() {
		itemID := tree.Insert(aabbFromSphere(0.0, 0.0, 0.0, 1.0), "Item")
		emptyAABB := shape3d.NewAABB(1.0, 1.0, 1.0, -1.0, -1.0, -1.0)
		Expect(func() { tree.Update(itemID, emptyAABB) }).To(Panic())
	})

	When("an item has a non-cubic box", func() {
		BeforeEach(func() {
			// A rod stretching along the X axis. Its bounding cube would span
			// 40 units in every direction, whereas the box itself is only two
			// units thick along Y and Z.
			tree.Insert(
				shape3d.NewAABB(-40.0, -2.0, -2.0, 40.0, 2.0, 2.0),
				"Rod",
			)
		})

		It("is found through a query that overlaps the box", func() {
			var found []string
			tree.QueryAABB(aabbFromSphere(30.0, 0.0, 0.0, 2.0), func(item string) bool {
				found = append(found, item)
				return true
			})
			Expect(found).To(ConsistOf("Rod"))
		})

		It("is not found through a query that only overlaps its bounding cube", func() {
			var found []string
			tree.QueryAABB(aabbFromSphere(30.0, 20.0, 0.0, 2.0), func(item string) bool {
				found = append(found, item)
				return true
			})
			Expect(found).To(BeEmpty())
		})

		It("is not found through a segment that only crosses its bounding cube", func() {
			segment := shape3d.NewSegment(
				dprec.NewVec3(20.0, 10.0, -30.0),
				dprec.NewVec3(20.0, 10.0, 30.0),
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
			firstItemID  query3d.TreeItemID
			secondItemID query3d.TreeItemID
			thirdItemID  query3d.TreeItemID
		)

		BeforeEach(func() {
			firstItemID = tree.Insert(
				aabbFromSphere(16.0, 16.0, 16.0, 2.0),
				"First",
			)
			secondItemID = tree.Insert(
				aabbFromSphere(48.0, 48.0, 48.0, 2.0),
				"Second",
			)
			thirdItemID = tree.Insert(
				aabbFromSphere(-16.0, -48.0, -16.0, 32.0),
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
			// grandchild along every axis.
			Expect(state.ItemCountPerDepth).To(Equal([]uint32{
				0, 0, 3,
			}))
		})

		It("is possible to segment-search for items", func() {
			from := dprec.NewVec3(1.0, 1.0, 1.0)
			to := dprec.NewVec3(127.0, 127.0, 127.0)
			segment := shape3d.NewSegment(from, to)
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
			segment := shape3d.NewSegment(from, to)
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
					aabbFromSphere(-48.0, 48.0, -48.0, 2.0),
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
				from := dprec.NewVec3(1.0, 1.0, 1.0)
				to := dprec.NewVec3(127.0, 127.0, 127.0)
				segment := shape3d.NewSegment(from, to)
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
				Expect(state.NodeCount).To(Equal(uint32(5)))
				Expect(state.ItemCount).To(Equal(uint32(2)))
				Expect(state.ItemCountPerDepth).To(Equal([]uint32{
					0, 0, 2,
				}))
			})

			It("does not return an active item id on new insert", func() {
				tree.Stats() // forces internal reordering of items (white box testing)
				secondItemID = tree.Insert(
					aabbFromSphere(48.0, 48.0, 48.0, 2.0),
					"Second",
				)
				Expect(secondItemID).ToNot(Equal(firstItemID))
				Expect(secondItemID).ToNot(Equal(thirdItemID))
			})

			It("is reflected in segment-search for items", func() {
				from := dprec.NewVec3(1.0, 1.0, 1.0)
				to := dprec.NewVec3(127.0, 127.0, 127.0)
				segment := shape3d.NewSegment(from, to)
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

	When("an item is thin along one axis", func() {
		// Both boxes have the same center and the same largest extent, and
		// both descend into the same child node. They differ only along Y,
		// where the node they would descend into next has its center 15 units
		// away. The slab is thin enough along Y to still fit; the block, being
		// as tall as it is wide, is not.
		var (
			slabAABB  = shape3d.NewAABB(-14.0, 32.0, 15.0, 46.0, 34.0, 17.0)
			blockAABB = shape3d.NewAABB(-14.0, 3.0, -14.0, 46.0, 63.0, 46.0)
		)

		It("descends deeper than a cubic item of the same largest extent", func() {
			tree.Insert(slabAABB, "Slab")
			state := tree.Stats()
			Expect(state.NodeCount).To(Equal(uint32(3))) // root + child + grandchild
			Expect(state.ItemCountPerDepth).To(Equal([]uint32{
				0, 0, 1,
			}))
		})

		It("keeps a cubic item at the depth its largest extent allows", func() {
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
			tree.QueryAABB(shape3d.NewAABB(15.0, 33.0, 16.0, 17.0, 33.0, 16.0),
				func(item string) bool {
					found = append(found, item)
					return true
				})
			Expect(found).To(ConsistOf("Slab", "Block"))
		})
	})

	When("an item creates a deeply nested branch", func() {
		var deepItemID query3d.TreeItemID

		BeforeEach(func() {
			// A tiny item placed off-center descends to the deepest allowed
			// node, allocating one node per depth level along the way.
			deepItemID = tree.Insert(
				aabbFromSphere(60.0, 60.0, 60.0, 1.0),
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
					aabbFromSphere(0.0, 0.0, 0.0, 60.0),
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
				aabbFromSphere(16.0, 16.0, 16.0, 2.0),
				"Near",
			)
			farItemID = tree.Insert(
				aabbFromSphere(60.0, 60.0, 60.0, 1.0),
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

			positionFor := func(i int) shape3d.AABB {
				x := float64(-60 + (i*7)%120)
				y := float64(-60 + (i*13)%120)
				z := float64(-60 + (i*5)%120)
				return aabbFromSphere(x, y, z, 1.0)
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
	Describe("QueryFrustum", func() {
		// lookingFrustum returns a frustum positioned at the specified point,
		// looking down the negative Z axis, with a 90 degree field of view.
		lookingFrustum := func(position dprec.Vec3, near, far float64) shape3d.Frustum {
			projection := dprec.PerspectiveMat4(-near, near, -near, near, near, far)
			view := dprec.InverseMat4(dprec.TranslationMat4(position.X, position.Y, position.Z))
			return shape3d.FrustumFromProjection(dprec.Mat4Prod(projection, view))
		}

		collect := func(frustum shape3d.Frustum) []string {
			var found []string
			tree.QueryFrustum(frustum, func(item string) bool {
				found = append(found, item)
				return true
			})
			return found
		}

		When("the tree is empty", func() {
			It("finds nothing and rejects the root", func() {
				found := collect(lookingFrustum(dprec.NewVec3(0.0, 0.0, 64.0), 1.0, 200.0))
				Expect(found).To(BeEmpty())

				stats := tree.VisitStats()
				Expect(stats.NodeCountVisited).To(Equal(uint32(1)))
				Expect(stats.NodeCountAccepted).To(Equal(uint32(0)))
				Expect(stats.NodeCountRejected).To(Equal(uint32(1)))
				Expect(stats.ItemCountVisited).To(Equal(uint32(0)))
			})
		})

		When("items are inserted", func() {
			BeforeEach(func() {
				tree.Insert(aabbFromSphere(0.0, 0.0, 0.0, 2.0), "Front")
				tree.Insert(aabbFromSphere(0.0, 0.0, 100.0, 2.0), "Behind")
				tree.Insert(aabbFromSphere(0.0, 0.0, -100.0, 2.0), "Far")
				tree.Insert(aabbFromSphere(50.0, 0.0, 0.0, 2.0), "Side")
				tree.Insert(aabbFromSphere(60.0, 0.0, 0.0, 2.0), "Edge")
			})

			It("finds only the items within the frustum", func() {
				// From 64 units back, at a distance of 64 the frustum is 64 units
				// wide in each direction, so the item at X=60 is within it and the
				// item at X=50 is even more so. The far plane at 120 leaves out
				// the item at Z=-100 (164 units away) and the item behind the
				// camera is not visible at all.
				found := collect(lookingFrustum(dprec.NewVec3(0.0, 0.0, 64.0), 1.0, 120.0))
				Expect(found).To(ConsistOf("Front", "Side", "Edge"))
			})

			It("respects the far plane", func() {
				found := collect(lookingFrustum(dprec.NewVec3(0.0, 0.0, 64.0), 1.0, 200.0))
				Expect(found).To(ConsistOf("Front", "Side", "Edge", "Far"))
			})

			It("respects the side planes", func() {
				// From 8 units back the frustum is only 8 units wide at the origin,
				// which is enough for the item at the origin but not for the ones
				// off to the side.
				found := collect(lookingFrustum(dprec.NewVec3(0.0, 0.0, 8.0), 1.0, 200.0))
				Expect(found).To(ConsistOf("Front", "Far"))
			})

			It("stops after the visitor returns false", func() {
				count := 0
				tree.QueryFrustum(lookingFrustum(dprec.NewVec3(0.0, 0.0, 64.0), 1.0, 200.0), func(item string) bool {
					count++
					return false
				})
				Expect(count).To(Equal(1))
			})

			It("reports consistent visit stats", func() {
				collect(lookingFrustum(dprec.NewVec3(0.0, 0.0, 64.0), 1.0, 120.0))
				stats := tree.VisitStats()
				Expect(stats.NodeCountVisited).To(Equal(stats.NodeCountAccepted + stats.NodeCountRejected))
				Expect(stats.ItemCountVisited).To(Equal(stats.ItemCountAccepted + stats.ItemCountRejected))
				Expect(stats.ItemCountAccepted).To(Equal(uint32(3)))
				Expect(stats.ItemCountRejected).To(Equal(uint32(2)))
				Expect(stats.NodeCountAccepted).To(BeNumerically(">", 0))
			})

			It("matches an AABB query when built from the same box", func() {
				box := shape3d.NewAABB(-10.0, -10.0, -10.0, 55.0, 10.0, 10.0)

				var fromAABB []string
				tree.QueryAABB(box, func(item string) bool {
					fromAABB = append(fromAABB, item)
					return true
				})
				Expect(fromAABB).To(ConsistOf("Front", "Side"))

				fromFrustum := collect(shape3d.FrustumFromAABB(box))
				Expect(fromFrustum).To(ConsistOf(fromAABB))
			})
		})

		When("many random items are inserted", func() {
			var (
				boxes    []shape3d.AABB
				frustum  shape3d.Frustum
				expected []string
			)

			// boxIntersectsFrustum is the reference item-level test: a box is
			// accepted unless it lies fully behind one of the six surfaces.
			boxIntersectsFrustum := func(box shape3d.AABB, frustum shape3d.Frustum) bool {
				for _, surface := range frustum.Surfaces {
					corner := dprec.NewVec3(box.MinX, box.MinY, box.MinZ)
					if surface.Normal.X >= 0.0 {
						corner.X = box.MaxX
					}
					if surface.Normal.Y >= 0.0 {
						corner.Y = box.MaxY
					}
					if surface.Normal.Z >= 0.0 {
						corner.Z = box.MaxZ
					}
					if surface.SignedDistance(corner) < 0.0 {
						return false
					}
				}
				return true
			}

			BeforeEach(func() {
				random := rand.New(rand.NewPCG(7, 13))
				boxes = make([]shape3d.AABB, 300)
				for i := range boxes {
					x := random.Float64()*120.0 - 60.0
					y := random.Float64()*120.0 - 60.0
					z := random.Float64()*120.0 - 60.0
					radius := 0.5 + random.Float64()*8.0
					boxes[i] = aabbFromSphere(x, y, z, radius)
					tree.Insert(boxes[i], fmt.Sprintf("item-%d", i))
				}

				frustum = lookingFrustum(dprec.NewVec3(10.0, -5.0, 40.0), 0.5, 90.0)
				expected = nil
				for i, box := range boxes {
					if boxIntersectsFrustum(box, frustum) {
						expected = append(expected, fmt.Sprintf("item-%d", i))
					}
				}
			})

			It("finds exactly the items that the item-level test accepts", func() {
				Expect(expected).ToNot(BeEmpty())
				Expect(len(expected)).To(BeNumerically("<", len(boxes)))
				Expect(collect(frustum)).To(ConsistOf(expected))
			})
		})
	})
})
