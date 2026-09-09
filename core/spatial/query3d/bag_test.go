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

var _ = Describe("Bag", func() {
	var (
		bag *query3d.Bag[string]
	)

	// collectAABB returns every item that the bag reports for the given box.
	collectAABB := func(aabb shape3d.AABB) []string {
		var found []string
		bag.QueryAABB(aabb, func(item string) bool {
			found = append(found, item)
			return true
		})
		return found
	}

	// collectSegment returns every item that the bag reports for the segment.
	collectSegment := func(segment shape3d.Segment) []string {
		var found []string
		bag.QuerySegment(segment, func(item string) bool {
			found = append(found, item)
			return true
		})
		return found
	}

	// everything is a box large enough to cover the whole working area.
	everything := aabbFromSphere(0.0, 0.0, 0.0, 1000.0)

	BeforeEach(func() {
		bag = query3d.NewBag[string](query3d.BagSettings{})
	})

	It("finds nothing when empty", func() {
		Expect(collectAABB(everything)).To(BeEmpty())
	})

	It("panics when an item with an empty box is inserted", func() {
		emptyAABB := shape3d.NewAABB(1.0, 1.0, 1.0, -1.0, -1.0, -1.0)
		Expect(func() { bag.Insert(emptyAABB, "Empty") }).To(Panic())
	})

	It("panics when an item is updated to an empty box", func() {
		itemID := bag.Insert(aabbFromSphere(0.0, 0.0, 0.0, 1.0), "Item")
		emptyAABB := shape3d.NewAABB(1.0, 1.0, 1.0, -1.0, -1.0, -1.0)
		Expect(func() { bag.Update(itemID, emptyAABB) }).To(Panic())
	})

	When("an item has a non-cubic box", func() {
		BeforeEach(func() {
			// A rod stretching along the X axis; it is only two units thick
			// along Y and Z.
			bag.Insert(
				shape3d.NewAABB(-40.0, -2.0, -2.0, 40.0, 2.0, 2.0),
				"Rod",
			)
		})

		It("is found through a query that overlaps the box", func() {
			Expect(collectAABB(aabbFromSphere(30.0, 0.0, 0.0, 2.0))).To(ConsistOf("Rod"))
		})

		It("is not found through a query that misses the box", func() {
			Expect(collectAABB(aabbFromSphere(30.0, 20.0, 0.0, 2.0))).To(BeEmpty())
		})

		It("is not found through a segment that misses the box", func() {
			segment := shape3d.NewSegment(
				dprec.NewVec3(20.0, 10.0, -30.0),
				dprec.NewVec3(20.0, 10.0, 30.0),
			)
			Expect(collectSegment(segment)).To(BeEmpty())
		})

		It("is found through a segment that crosses the box", func() {
			segment := shape3d.NewSegment(
				dprec.NewVec3(20.0, 10.0, 0.0),
				dprec.NewVec3(20.0, -10.0, 0.0),
			)
			Expect(collectSegment(segment)).To(ConsistOf("Rod"))
		})
	})

	When("items are inserted", func() {
		var (
			firstItemID  query3d.BagItemID
			secondItemID query3d.BagItemID
			thirdItemID  query3d.BagItemID
		)

		BeforeEach(func() {
			firstItemID = bag.Insert(aabbFromSphere(16.0, 16.0, 16.0, 2.0), "First")
			secondItemID = bag.Insert(aabbFromSphere(48.0, 48.0, 48.0, 2.0), "Second")
			thirdItemID = bag.Insert(aabbFromSphere(-16.0, -48.0, -16.0, 2.0), "Third")
		})

		It("returns unique ids", func() {
			Expect(firstItemID).ToNot(Equal(secondItemID))
			Expect(firstItemID).ToNot(Equal(thirdItemID))
			Expect(secondItemID).ToNot(Equal(thirdItemID))
		})

		It("finds all of them through a covering query", func() {
			Expect(collectAABB(everything)).To(ConsistOf("First", "Second", "Third"))
		})

		It("is possible to segment-search for items", func() {
			segment := shape3d.NewSegment(
				dprec.NewVec3(1.0, 1.0, 1.0),
				dprec.NewVec3(127.0, 127.0, 127.0),
			)
			Expect(collectSegment(segment)).To(ConsistOf("First", "Second"))
		})

		It("stops QuerySegment after the visitor returns false", func() {
			segment := shape3d.NewSegment(
				dprec.NewVec3(1.0, 1.0, 1.0),
				dprec.NewVec3(127.0, 127.0, 127.0),
			)
			count := 0
			bag.QuerySegment(segment, func(item string) bool {
				count++
				return false // stop after first item
			})
			Expect(count).To(Equal(1))
		})

		It("is possible to area-search for items", func() {
			Expect(collectAABB(aabbFromSphere(64.0, 64.0, 64.0, 63.0))).To(ConsistOf("First", "Second"))
		})

		It("stops QueryAABB after the visitor returns false", func() {
			count := 0
			bag.QueryAABB(aabbFromSphere(64.0, 64.0, 64.0, 63.0), func(item string) bool {
				count++
				return false // stop after first item
			})
			Expect(count).To(Equal(1))
		})

		When("an item is updated", func() {
			BeforeEach(func() {
				bag.Update(secondItemID, aabbFromSphere(-48.0, 48.0, -48.0, 2.0))
			})

			It("no longer finds it at the old location", func() {
				Expect(collectAABB(aabbFromSphere(48.0, 48.0, 48.0, 4.0))).To(BeEmpty())
			})

			It("finds it at the new location", func() {
				Expect(collectAABB(aabbFromSphere(-48.0, 48.0, -48.0, 4.0))).To(ConsistOf("Second"))
			})

			It("is reflected in area-search for items", func() {
				Expect(collectAABB(aabbFromSphere(64.0, 64.0, 64.0, 63.0))).To(ConsistOf("First"))
			})
		})

		When("an item is removed", func() {
			BeforeEach(func() {
				bag.Remove(secondItemID)
			})

			It("panics when the same item is removed again", func() {
				Expect(func() { bag.Remove(secondItemID) }).To(Panic())
			})

			It("panics when the removed item is updated", func() {
				Expect(func() {
					bag.Update(secondItemID, aabbFromSphere(0.0, 0.0, 0.0, 1.0))
				}).To(Panic())
			})

			It("no longer finds the removed item", func() {
				Expect(collectAABB(everything)).To(ConsistOf("First", "Third"))
			})

			It("does not collide a reused id with a still-active id", func() {
				reusedID := bag.Insert(aabbFromSphere(48.0, 48.0, 48.0, 2.0), "Second")
				Expect(reusedID).ToNot(Equal(firstItemID))
				Expect(reusedID).ToNot(Equal(thirdItemID))
				Expect(collectAABB(everything)).To(ConsistOf("First", "Second", "Third"))
			})
		})

		// Removing an item that is not the last one exercises the swap-remove:
		// the last item is moved into the vacated slot and its id mapping must
		// be kept in sync, otherwise it becomes unreachable.
		When("a non-last item is removed", func() {
			BeforeEach(func() {
				bag.Remove(firstItemID) // First is the first-inserted, not the last
			})

			It("still finds the item that was moved into the freed slot", func() {
				Expect(collectAABB(everything)).To(ConsistOf("Second", "Third"))
			})

			It("can still update the moved item through its id", func() {
				bag.Update(thirdItemID, aabbFromSphere(10.0, 10.0, 10.0, 2.0))
				Expect(collectAABB(aabbFromSphere(10.0, 10.0, 10.0, 4.0))).To(ConsistOf("Third"))
				Expect(collectAABB(aabbFromSphere(-16.0, -48.0, -16.0, 4.0))).To(BeEmpty())
			})

			It("can still remove the moved item through its id", func() {
				bag.Remove(thirdItemID)
				Expect(collectAABB(everything)).To(ConsistOf("Second"))
			})
		})
	})

	When("the bag undergoes heavy churn", func() {
		It("keeps queries consistent", func() {
			const count = 200
			bag = query3d.NewBag[string](query3d.BagSettings{
				InitialItemCapacity: opt.V[uint32](4), // force the slice to grow
			})

			ids := make([]query3d.BagItemID, count)
			expected := make(map[query3d.BagItemID]string, count)

			positionFor := func(i int) shape3d.AABB {
				x := float64(-60 + (i*7)%120)
				y := float64(-60 + (i*13)%120)
				z := float64(-60 + (i*5)%120)
				return aabbFromSphere(x, y, z, 1.0)
			}

			// Populate the bag.
			for i := range count {
				value := fmt.Sprintf("item-%d", i)
				ids[i] = bag.Insert(positionFor(i), value)
				expected[ids[i]] = value
			}

			// Churn: drop every third item and relocate half of the rest.
			for i := range count {
				switch {
				case i%3 == 0:
					bag.Remove(ids[i])
					delete(expected, ids[i])
				case i%2 == 0:
					bag.Update(ids[i], positionFor(i+1))
				}
			}

			// Re-insert into the freed slots to exercise id reuse.
			for i := 0; i < count; i += 3 {
				value := fmt.Sprintf("reinsert-%d", i)
				id := bag.Insert(positionFor(i), value)
				expected[id] = value
			}

			found := make(map[string]struct{})
			bag.QueryAABB(everything, func(item string) bool {
				found[item] = struct{}{}
				return true
			})

			Expect(found).To(HaveLen(len(expected)))
			for _, value := range expected {
				Expect(found).To(HaveKey(value))
			}
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
			bag.QueryFrustum(frustum, func(item string) bool {
				found = append(found, item)
				return true
			})
			return found
		}

		When("the bag is empty", func() {
			It("finds nothing", func() {
				Expect(collect(lookingFrustum(dprec.NewVec3(0.0, 0.0, 64.0), 1.0, 200.0))).To(BeEmpty())
			})
		})

		When("items are inserted", func() {
			BeforeEach(func() {
				bag.Insert(aabbFromSphere(0.0, 0.0, 0.0, 2.0), "Front")
				bag.Insert(aabbFromSphere(0.0, 0.0, 100.0, 2.0), "Behind")
				bag.Insert(aabbFromSphere(0.0, 0.0, -100.0, 2.0), "Far")
				bag.Insert(aabbFromSphere(50.0, 0.0, 0.0, 2.0), "Side")
				bag.Insert(aabbFromSphere(60.0, 0.0, 0.0, 2.0), "Edge")
			})

			It("finds only the items within the frustum", func() {
				found := collect(lookingFrustum(dprec.NewVec3(0.0, 0.0, 64.0), 1.0, 120.0))
				Expect(found).To(ConsistOf("Front", "Side", "Edge"))
			})

			It("respects the far plane", func() {
				found := collect(lookingFrustum(dprec.NewVec3(0.0, 0.0, 64.0), 1.0, 200.0))
				Expect(found).To(ConsistOf("Front", "Side", "Edge", "Far"))
			})

			It("respects the side planes", func() {
				found := collect(lookingFrustum(dprec.NewVec3(0.0, 0.0, 8.0), 1.0, 200.0))
				Expect(found).To(ConsistOf("Front", "Far"))
			})

			It("stops after the visitor returns false", func() {
				count := 0
				bag.QueryFrustum(lookingFrustum(dprec.NewVec3(0.0, 0.0, 64.0), 1.0, 200.0), func(item string) bool {
					count++
					return false
				})
				Expect(count).To(Equal(1))
			})

			It("matches an AABB query when built from the same box", func() {
				box := shape3d.NewAABB(-10.0, -10.0, -10.0, 55.0, 10.0, 10.0)

				var fromAABB []string
				bag.QueryAABB(box, func(item string) bool {
					fromAABB = append(fromAABB, item)
					return true
				})
				Expect(fromAABB).To(ConsistOf("Front", "Side"))

				Expect(collect(shape3d.FrustumFromAABB(box))).To(ConsistOf(fromAABB))
			})
		})

		When("many random items are inserted", func() {
			var (
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
				boxes := make([]shape3d.AABB, 300)
				for i := range boxes {
					x := random.Float64()*120.0 - 60.0
					y := random.Float64()*120.0 - 60.0
					z := random.Float64()*120.0 - 60.0
					radius := 0.5 + random.Float64()*8.0
					boxes[i] = aabbFromSphere(x, y, z, radius)
					bag.Insert(boxes[i], fmt.Sprintf("item-%d", i))
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
				Expect(len(expected)).To(BeNumerically("<", 300))
				Expect(collect(frustum)).To(ConsistOf(expected))
			})
		})
	})
})
