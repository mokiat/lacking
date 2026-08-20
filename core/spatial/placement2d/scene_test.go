package placement2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/placement2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

// circleAt builds a circle with the given center coordinates and radius.
func circleAt(x, y, radius float64) shape2d.Circle {
	return shape2d.Circle{
		Center: dprec.NewVec2(x, y),
		Radius: radius,
	}
}

// rectangleAt builds an axis-aligned rectangle centered at the given
// coordinates with the given half-extent along every axis.
func rectangleAt(x, y, half float64) shape2d.Rectangle {
	return shape2d.NewRectangle(
		dprec.NewVec2(x, y),
		shape2d.IdentityRotation(),
		dprec.NewVec2(half, half),
	)
}

// lineMesh builds a mesh made of a single edge forming a horizontal line (at
// the given y), centered at the given point and spanning halfSize in the X
// direction. The edge is wound so that its normal faces -Y.
func lineMesh(x, y, halfSize float64) shape2d.Mesh {
	a := dprec.NewVec2(x-halfSize, y)
	b := dprec.NewVec2(x+halfSize, y)
	return shape2d.NewMesh([]shape2d.Edge{
		shape2d.NewEdge(a, b),
	})
}

var _ = Describe("Scene", func() {
	var scene *placement2d.Scene[string, string, string]

	BeforeEach(func() {
		scene = placement2d.NewScene[string, string, string](placement2d.SceneSettings{
			Size:     opt.V(128.0),
			MaxDepth: opt.V[uint32](3),
		})
	})

	Describe("object management", func() {
		It("creates objects placed at the origin by default", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			Expect(objID).NotTo(Equal(placement2d.NilObjectID))

			transform := scene.GetObjectTransform(objID)
			Expect(transform.Translation).To(dprectest.HaveVec2Coords(0.0, 0.0))
		})

		It("honors the provided position", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec2(1.0, 2.0)),
			})
			transform := scene.GetObjectTransform(objID)
			Expect(transform.Translation).To(dprectest.HaveVec2Coords(1.0, 2.0))
		})

		It("stores and updates user data", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{
				UserData: "first",
			})
			Expect(scene.GetObjectUserData(objID)).To(Equal("first"))

			scene.SetObjectUserData(objID, "second")
			Expect(scene.GetObjectUserData(objID)).To(Equal("second"))
		})

		It("relocates objects via SetObjectTransform", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.SetObjectTransform(objID, shape2d.TranslationTransform(
				dprec.NewVec2(5.0, 6.0),
			))
			transform := scene.GetObjectTransform(objID)
			Expect(transform.Translation).To(dprectest.HaveVec2Coords(5.0, 6.0))
		})

		It("reuses the indices of deleted objects", func() {
			first := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.DeleteObject(first)
			second := scene.CreateObject(placement2d.ObjectInfo[string]{})
			Expect(second).To(Equal(first))
		})
	})

	Describe("terrain management", func() {
		It("creates terrains", func() {
			terrainID := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			Expect(terrainID).NotTo(Equal(placement2d.NilTerrainID))
		})

		It("stores and updates user data", func() {
			terrainID := scene.CreateTerrain(placement2d.TerrainInfo[string]{
				UserData: "first",
			})
			Expect(scene.GetTerrainUserData(terrainID)).To(Equal("first"))

			scene.SetTerrainUserData(terrainID, "second")
			Expect(scene.GetTerrainUserData(terrainID)).To(Equal("second"))
		})

		It("stores and updates terrain shape user data", func() {
			terrainID := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			shapeID := scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh:     lineMesh(0.0, 0.0, 5.0),
				UserData: "a",
			})
			Expect(scene.GetTerrainShapeUserData(shapeID)).To(Equal("a"))

			scene.SetTerrainShapeUserData(shapeID, "b")
			Expect(scene.GetTerrainShapeUserData(shapeID)).To(Equal("b"))
		})

		It("maps a terrain shape back to its owning terrain", func() {
			terrainID := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			shapeID := scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})
			Expect(scene.GetTerrainForShape(shapeID)).To(Equal(terrainID))
		})

		It("reuses the indices of deleted terrains", func() {
			first := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			scene.DeleteTerrain(first)
			second := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			Expect(second).To(Equal(first))
		})

		It("panics when attaching an empty mesh", func() {
			terrainID := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			Expect(func() {
				scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
					Mesh: shape2d.NewMesh(nil),
				})
			}).To(Panic())
		})
	})

	Describe("shape iteration", func() {
		var objID placement2d.ObjectID

		BeforeEach(func() {
			objID = scene.CreateObject(placement2d.ObjectInfo[string]{})
		})

		It("yields attached circles in world space", func() {
			scene.SetObjectTransform(objID, shape2d.TranslationTransform(
				dprec.NewVec2(10.0, 0.0),
			))
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 2.0),
			})

			var found []shape2d.Circle
			scene.EachCircle(placement2d.Filter{}, func(c shape2d.Circle) bool {
				found = append(found, c)
				return true
			})
			Expect(found).To(HaveLen(1))
			Expect(found[0].Center).To(dprectest.HaveVec2Coords(10.0, 0.0))
			Expect(found[0].Radius).To(Equal(2.0))
		})

		It("yields attached rectangles", func() {
			scene.AttachRectangle(objID, placement2d.RectangleInfo[string]{
				Rectangle: rectangleAt(0.0, 0.0, 1.0),
			})

			count := 0
			scene.EachRectangle(placement2d.Filter{}, func(shape2d.Rectangle) bool {
				count++
				return true
			})
			Expect(count).To(Equal(1))
		})

		It("exposes a circle iterator", func() {
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			count := 0
			for range scene.CircleIter(placement2d.Filter{}) {
				count++
			}
			Expect(count).To(Equal(1))
		})

		It("exposes a rectangle iterator", func() {
			scene.AttachRectangle(objID, placement2d.RectangleInfo[string]{
				Rectangle: rectangleAt(0.0, 0.0, 1.0),
			})

			count := 0
			for range scene.RectangleIter(placement2d.Filter{}) {
				count++
			}
			Expect(count).To(Equal(1))
		})

		It("stores and updates shape user data", func() {
			shapeID := scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle:   circleAt(0.0, 0.0, 1.0),
				UserData: "a",
			})
			Expect(scene.GetObjectShapeUserData(shapeID)).To(Equal("a"))

			scene.SetObjectShapeUserData(shapeID, "b")
			Expect(scene.GetObjectShapeUserData(shapeID)).To(Equal("b"))
		})

		It("maps a shape back to its owning object", func() {
			shapeID := scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			Expect(scene.GetObjectForShape(shapeID)).To(Equal(objID))
		})

		It("removes a deleted shape from iteration", func() {
			shapeID := scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.DeleteObjectShape(shapeID)

			count := 0
			scene.EachCircle(placement2d.Filter{}, func(shape2d.Circle) bool {
				count++
				return true
			})
			Expect(count).To(BeZero())
		})

		It("stops iteration when the callback returns false", func() {
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(3.0, 0.0, 1.0),
			})

			count := 0
			scene.EachCircle(placement2d.Filter{}, func(shape2d.Circle) bool {
				count++
				return false
			})
			Expect(count).To(Equal(1))
		})

		It("moves every shape of an object with multiple shapes", func() {
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(2.0, 0.0, 1.0),
			})
			scene.SetObjectTransform(objID, shape2d.TranslationTransform(
				dprec.NewVec2(0.0, 10.0),
			))

			var centers []dprec.Vec2
			scene.EachCircle(placement2d.Filter{}, func(c shape2d.Circle) bool {
				centers = append(centers, c.Center)
				return true
			})
			Expect(centers).To(HaveLen(2))
			for _, center := range centers {
				Expect(center.Y).To(BeNumerically("~", 10.0, 1e-6))
			}
		})
	})

	Describe("shape iteration filters", func() {
		BeforeEach(func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Filtering: placement2d.FilterInfo{
					RejectGroup: 7,
					SourceMask:  opt.V(uint32(0b01)),
				},
				Circle: circleAt(0.0, 0.0, 1.0),
			})
		})

		countCircles := func(filter placement2d.Filter) int {
			count := 0
			scene.EachCircle(filter, func(shape2d.Circle) bool {
				count++
				return true
			})
			return count
		}

		It("yields shapes that occupy a layer of the mask", func() {
			Expect(countCircles(placement2d.Filter{Mask: 0b01})).To(Equal(1))
		})

		It("skips shapes that occupy no layer of the mask", func() {
			Expect(countCircles(placement2d.Filter{Mask: 0b10})).To(BeZero())
		})

		It("yields everything for the zero mask", func() {
			Expect(countCircles(placement2d.Filter{Mask: 0})).To(Equal(1))
		})

		It("yields everything for the full mask", func() {
			Expect(countCircles(placement2d.Filter{
				Mask: placement2d.FullMask,
			})).To(Equal(1))
		})

		It("skips shapes that share the reject group", func() {
			Expect(countCircles(placement2d.Filter{RejectGroup: 7})).To(BeZero())
		})

		It("yields shapes that have a different reject group", func() {
			Expect(countCircles(placement2d.Filter{RejectGroup: 8})).To(Equal(1))
		})

		It("yields everything for the zero filter", func() {
			Expect(countCircles(placement2d.Filter{})).To(Equal(1))
		})
	})

	Describe("CollectObjectIntersections", func() {
		// attachOverlappingCircles places two unit circles 1.5 apart (so they
		// overlap) on freshly created objects and returns the object IDs.
		attachOverlappingCircles := func() (placement2d.ObjectID, placement2d.ObjectID) {
			first := scene.CreateObject(placement2d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec2(0.0, 0.0)),
			})
			second := scene.CreateObject(placement2d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec2(1.5, 0.0)),
			})
			scene.AttachCircle(first, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(second, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			return first, second
		}

		collect := func() placement2d.ObjectContactList {
			var contacts placement2d.ObjectContactList
			scene.CollectObjectIntersections(contacts.AddContact)
			return contacts
		}

		It("reports a single contact between two overlapping shapes", func() {
			first, second := attachOverlappingCircles()
			contacts := collect()
			Expect(contacts).To(HaveLen(1))
			Expect([]placement2d.ObjectID{
				contacts[0].SourceObjectID,
				contacts[0].TargetObjectID,
			}).To(ConsistOf(first, second))
		})

		It("reports object IDs that agree with the shape IDs", func() {
			attachOverlappingCircles()
			contacts := collect()
			Expect(contacts).To(HaveLen(1))

			contact := contacts[0]
			Expect(contact.SourceObjectID).To(Equal(
				scene.GetObjectForShape(contact.SourceShapeID),
			))
			Expect(contact.TargetObjectID).To(Equal(
				scene.GetObjectForShape(contact.TargetShapeID),
			))
		})

		It("does not report contacts between shapes of the same object", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(1.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("does not report disjoint shapes", func() {
			first := scene.CreateObject(placement2d.ObjectInfo[string]{})
			second := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(first, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(second, placement2d.CircleInfo[string]{
				Circle: circleAt(10.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("does not report shapes that share a reject group", func() {
			first := scene.CreateObject(placement2d.ObjectInfo[string]{})
			second := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(first, placement2d.CircleInfo[string]{
				Filtering: placement2d.FilterInfo{RejectGroup: 7},
				Circle:    circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(second, placement2d.CircleInfo[string]{
				Filtering: placement2d.FilterInfo{RejectGroup: 7},
				Circle:    circleAt(1.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("does not report shapes whose masks do not overlap", func() {
			first := scene.CreateObject(placement2d.ObjectInfo[string]{})
			second := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(first, placement2d.CircleInfo[string]{
				Filtering: placement2d.FilterInfo{
					SourceMask: opt.V(uint32(0b01)),
					TargetMask: opt.V(uint32(0b01)),
				},
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(second, placement2d.CircleInfo[string]{
				Filtering: placement2d.FilterInfo{
					SourceMask: opt.V(uint32(0b10)),
					TargetMask: opt.V(uint32(0b10)),
				},
				Circle: circleAt(1.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("stops reporting once a deleted object's shapes are gone", func() {
			first, _ := attachOverlappingCircles()
			Expect(collect()).To(HaveLen(1))

			scene.DeleteObject(first)
			Expect(collect()).To(BeEmpty())
		})

		It("reflects object movement in the broadphase", func() {
			scene.CreateObject(placement2d.ObjectInfo[string]{})
			first := scene.CreateObject(placement2d.ObjectInfo[string]{})
			second := scene.CreateObject(placement2d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec2(10.0, 0.0)),
			})
			scene.AttachCircle(first, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(second, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())

			scene.SetObjectTransform(second, shape2d.TranslationTransform(
				dprec.NewVec2(1.5, 0.0),
			))
			Expect(collect()).To(HaveLen(1))
		})

		It("reports a contact between two overlapping rectangles", func() {
			first := scene.CreateObject(placement2d.ObjectInfo[string]{})
			second := scene.CreateObject(placement2d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec2(1.5, 0.0)),
			})
			scene.AttachRectangle(first, placement2d.RectangleInfo[string]{
				Rectangle: rectangleAt(0.0, 0.0, 2.0),
			})
			scene.AttachRectangle(second, placement2d.RectangleInfo[string]{
				Rectangle: rectangleAt(0.0, 0.0, 2.0),
			})
			Expect(collect()).To(HaveLen(1))
		})

		It("reports a contact between an overlapping circle and rectangle", func() {
			first := scene.CreateObject(placement2d.ObjectInfo[string]{})
			second := scene.CreateObject(placement2d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec2(1.0, 0.0)),
			})
			scene.AttachCircle(first, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachRectangle(second, placement2d.RectangleInfo[string]{
				Rectangle: rectangleAt(0.0, 0.0, 1.0),
			})
			Expect(collect()).To(HaveLen(1))
		})

		It("does not report contacts with terrain shapes", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, -0.5, 1.0),
			})
			terrainID := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("reports every overlapping pair of two multi-shape objects", func() {
			first := scene.CreateObject(placement2d.ObjectInfo[string]{})
			second := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(first, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(first, placement2d.CircleInfo[string]{
				Circle: circleAt(0.5, 0.0, 1.0),
			})
			scene.AttachCircle(second, placement2d.CircleInfo[string]{
				Circle: circleAt(0.25, 0.0, 1.0),
			})
			// Both shapes of the first object overlap the single shape of the
			// second one, while the two shapes of the first object are not
			// tested against each other.
			Expect(collect()).To(HaveLen(2))
		})

		It("does not produce phantom contacts after index reuse", func() {
			first, second := attachOverlappingCircles()
			Expect(collect()).To(HaveLen(1))

			scene.DeleteObject(first)
			scene.DeleteObject(second)
			Expect(collect()).To(BeEmpty())

			// Recreate, reusing both the object and the shape indices.
			third := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(third, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())
		})
	})

	Describe("CollectTerrainIntersections", func() {
		var objID placement2d.ObjectID
		var terrainID placement2d.TerrainID

		BeforeEach(func() {
			objID = scene.CreateObject(placement2d.ObjectInfo[string]{})
			terrainID = scene.CreateTerrain(placement2d.TerrainInfo[string]{})
		})

		collect := func() placement2d.TerrainContactList {
			var contacts placement2d.TerrainContactList
			scene.CollectTerrainIntersections(contacts.AddContact)
			return contacts
		}

		It("reports a contact between a shape and an overlapping mesh", func() {
			// The line's normal faces -Y, so the circle is placed just below
			// the line (on the front side) where it overlaps and is pushed out.
			shapeID := scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, -0.5, 1.0),
			})
			meshID := scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			contacts := collect()
			Expect(contacts).To(HaveLen(1))
			Expect(contacts[0].SourceObjectID).To(Equal(objID))
			Expect(contacts[0].SourceShapeID).To(Equal(shapeID))
			Expect(contacts[0].TargetTerrainID).To(Equal(terrainID))
			Expect(contacts[0].TargetShapeID).To(Equal(meshID))

			// The contact normal must push the circle out the front (-Y) side,
			// never inward into the mesh.
			Expect(contacts[0].TargetNormal.Y).To(BeNumerically("<", 0.0))
		})

		It("does not report a shape overlapping a mesh from behind", func() {
			// A circle on the +Y (back) side of the -Y-facing line would have
			// to be pushed further inward to separate, which the mesh logic
			// prevents.
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.5, 1.0),
			})
			scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("does not report a shape disjoint from a mesh", func() {
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 10.0, 1.0),
			})
			scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("reports a single contact even when many edges overlap", func() {
			// A large rectangle overlaps both edges of the mesh, yet only the
			// deepest contact is reported.
			scene.AttachRectangle(objID, placement2d.RectangleInfo[string]{
				Rectangle: rectangleAt(0.0, -0.5, 2.0),
			})
			scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: shape2d.NewMesh([]shape2d.Edge{
					shape2d.NewEdge(dprec.NewVec2(-5.0, 0.0), dprec.NewVec2(0.0, 0.0)),
					shape2d.NewEdge(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(5.0, 0.0)),
				}),
			})
			Expect(collect()).To(HaveLen(1))
		})

		It("reports a contact per terrain shape", func() {
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, -0.5, 1.0),
			})
			scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})
			scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, -0.1, 5.0),
			})
			Expect(collect()).To(HaveLen(2))
		})

		It("reports a contact per object shape", func() {
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(-1.0, -0.5, 1.0),
			})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(1.0, -0.5, 1.0),
			})
			scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})
			Expect(collect()).To(HaveLen(2))
		})

		It("does not report shapes that share a reject group", func() {
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Filtering: placement2d.FilterInfo{RejectGroup: 7},
				Circle:    circleAt(0.0, -0.5, 1.0),
			})
			scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Filtering: placement2d.FilterInfo{RejectGroup: 7},
				Mesh:      lineMesh(0.0, 0.0, 5.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("does not report shapes whose masks do not overlap", func() {
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Filtering: placement2d.FilterInfo{
					SourceMask: opt.V(uint32(0b01)),
					TargetMask: opt.V(uint32(0b01)),
				},
				Circle: circleAt(0.0, -0.5, 1.0),
			})
			scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Filtering: placement2d.FilterInfo{
					SourceMask: opt.V(uint32(0b10)),
					TargetMask: opt.V(uint32(0b10)),
				},
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("stops reporting once the terrain is deleted", func() {
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, -0.5, 1.0),
			})
			scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})
			Expect(collect()).To(HaveLen(1))

			scene.DeleteTerrain(terrainID)
			Expect(collect()).To(BeEmpty())
		})

		It("stops reporting once the terrain shape is deleted", func() {
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, -0.5, 1.0),
			})
			meshID := scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})
			Expect(collect()).To(HaveLen(1))

			scene.DeleteTerrainShape(meshID)
			Expect(collect()).To(BeEmpty())
		})

		It("reattaches correctly after terrain shape index reuse", func() {
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, -0.5, 1.0),
			})
			meshID := scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})
			scene.DeleteTerrainShape(meshID)

			other := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			reusedID := scene.AttachMesh(other, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			contacts := collect()
			Expect(contacts).To(HaveLen(1))
			Expect(contacts[0].TargetTerrainID).To(Equal(other))
			Expect(contacts[0].TargetShapeID).To(Equal(reusedID))
			Expect(scene.GetTerrainForShape(reusedID)).To(Equal(other))
		})

		It("tracks object movement into and out of terrain contact", func() {
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, -0.5, 1.0),
			})
			scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})
			Expect(collect()).To(HaveLen(1))

			scene.SetObjectTransform(objID, shape2d.TranslationTransform(
				dprec.NewVec2(0.0, -20.0),
			))
			Expect(collect()).To(BeEmpty())

			scene.SetObjectTransform(objID, shape2d.TranslationTransform(
				dprec.NewVec2(0.0, 0.0),
			))
			Expect(collect()).To(HaveLen(1))
		})
	})

	Describe("shape id spaces", func() {
		It("keeps object and terrain shape ids independent", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			shapeID := scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle:   circleAt(0.0, 0.0, 1.0),
				UserData: "object-shape",
			})
			terrainID := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			meshID := scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh:     lineMesh(0.0, 0.0, 5.0),
				UserData: "terrain-shape",
			})

			// Both are the first shape of their kind, hence they share the raw
			// index while remaining distinct references.
			Expect(int32(shapeID)).To(Equal(int32(meshID)))
			Expect(scene.GetObjectShapeUserData(shapeID)).To(Equal("object-shape"))
			Expect(scene.GetTerrainShapeUserData(meshID)).To(Equal("terrain-shape"))
		})
	})

	Describe("circle queries", func() {
		It("reports a circle overlapping an object shape", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			shapeID := scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			contact, ok := scene.CheckCircleObjectIntersection(
				circleAt(1.5, 0.0, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceObjectID).To(Equal(placement2d.NilObjectID))
			Expect(contact.SourceShapeID).To(Equal(placement2d.NilObjectShapeID))
			Expect(contact.TargetObjectID).To(Equal(objID))
			Expect(contact.TargetShapeID).To(Equal(shapeID))
		})

		It("returns false for a circle disjoint from every object shape", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckCircleObjectIntersection(
				circleAt(10.0, 0.0, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("honors the query mask", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Filtering: placement2d.FilterInfo{
					SourceMask: opt.V(uint32(0b01)),
				},
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckCircleObjectIntersection(
				circleAt(1.5, 0.0, 1.0),
				placement2d.Filter{Mask: 0b10},
			)
			Expect(ok).To(BeFalse())
		})

		It("honors the query reject group", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Filtering: placement2d.FilterInfo{
					RejectGroup: 7,
				},
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckCircleObjectIntersection(
				circleAt(1.5, 0.0, 1.0),
				placement2d.Filter{RejectGroup: 7},
			)
			Expect(ok).To(BeFalse())
		})

		It("reports a circle overlapping a terrain shape from the front", func() {
			terrainID := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			meshID := scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			// The line faces -Y, so approach it from below (the front side).
			contact, ok := scene.CheckCircleTerrainIntersection(
				circleAt(0.0, -0.5, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceShapeID).To(Equal(placement2d.NilObjectShapeID))
			Expect(contact.TargetTerrainID).To(Equal(terrainID))
			Expect(contact.TargetShapeID).To(Equal(meshID))
			Expect(contact.TargetNormal.Y).To(BeNumerically("<", 0.0))
		})

		It("does not report a circle overlapping a terrain shape from behind", func() {
			terrainID := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			_, ok := scene.CheckCircleTerrainIntersection(
				circleAt(0.0, 0.5, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("keeps object and terrain queries separate", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckCircleTerrainIntersection(
				circleAt(0.5, 0.0, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})
	})

	Describe("rectangle queries", func() {
		It("reports a rectangle overlapping an object shape", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			shapeID := scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			contact, ok := scene.CheckRectangleObjectIntersection(
				rectangleAt(1.5, 0.0, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceShapeID).To(Equal(placement2d.NilObjectShapeID))
			Expect(contact.TargetObjectID).To(Equal(objID))
			Expect(contact.TargetShapeID).To(Equal(shapeID))
		})

		It("returns false for a rectangle disjoint from every object shape", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckRectangleObjectIntersection(
				rectangleAt(10.0, 0.0, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("reports a rectangle overlapping a terrain shape from the front", func() {
			terrainID := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			meshID := scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			// The line faces -Y, so approach it from below (the front side).
			contact, ok := scene.CheckRectangleTerrainIntersection(
				rectangleAt(0.0, -0.5, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetTerrainID).To(Equal(terrainID))
			Expect(contact.TargetShapeID).To(Equal(meshID))
			Expect(contact.TargetNormal.Y).To(BeNumerically("<", 0.0))
		})

		It("does not report a rectangle overlapping a terrain shape from behind", func() {
			terrainID := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			_, ok := scene.CheckRectangleTerrainIntersection(
				rectangleAt(0.0, 0.5, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})
	})

	Describe("segment queries", func() {
		It("collects every object shape a segment passes through", func() {
			near := scene.CreateObject(placement2d.ObjectInfo[string]{})
			far := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(near, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(far, placement2d.CircleInfo[string]{
				Circle: circleAt(4.0, 0.0, 1.0),
			})

			var contacts placement2d.ObjectContactList
			scene.CollectSegmentObjectIntersections(
				shape2d.NewSegment(
					dprec.NewVec2(-5.0, 0.0),
					dprec.NewVec2(9.0, 0.0),
				),
				placement2d.Filter{},
				contacts.AddContact,
			)
			Expect(contacts).To(HaveLen(2))
		})

		It("finds an object shape crossed by the segment", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			shapeID := scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			contact, ok := scene.CheckSegmentObjectIntersection(
				shape2d.NewSegment(
					dprec.NewVec2(-5.0, 0.0),
					dprec.NewVec2(5.0, 0.0),
				),
				placement2d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceShapeID).To(Equal(placement2d.NilObjectShapeID))
			Expect(contact.TargetObjectID).To(Equal(objID))
			Expect(contact.TargetShapeID).To(Equal(shapeID))
		})

		It("finds the nearest of two object shapes crossed by the segment", func() {
			near := scene.CreateObject(placement2d.ObjectInfo[string]{})
			far := scene.CreateObject(placement2d.ObjectInfo[string]{})
			nearShapeID := scene.AttachCircle(near, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(far, placement2d.CircleInfo[string]{
				Circle: circleAt(4.0, 0.0, 1.0),
			})

			contact, ok := scene.CheckSegmentObjectIntersection(
				shape2d.NewSegment(
					dprec.NewVec2(-5.0, 0.0),
					dprec.NewVec2(9.0, 0.0),
				),
				placement2d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetShapeID).To(Equal(nearShapeID))
		})

		It("finds a terrain shape crossed by the segment", func() {
			terrainID := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			meshID := scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			contact, ok := scene.CheckSegmentTerrainIntersection(
				shape2d.NewSegment(
					dprec.NewVec2(2.0, -5.0),
					dprec.NewVec2(2.0, 5.0),
				),
				placement2d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceShapeID).To(Equal(placement2d.NilObjectShapeID))
			Expect(contact.TargetTerrainID).To(Equal(terrainID))
			Expect(contact.TargetShapeID).To(Equal(meshID))
		})

		It("returns false when the segment misses everything", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckSegmentObjectIntersection(
				shape2d.NewSegment(
					dprec.NewVec2(-5.0, 5.0),
					dprec.NewVec2(5.0, 5.0),
				),
				placement2d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("keeps object and terrain queries separate", func() {
			terrainID := scene.CreateTerrain(placement2d.TerrainInfo[string]{})
			scene.AttachMesh(terrainID, placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			_, ok := scene.CheckSegmentObjectIntersection(
				shape2d.NewSegment(
					dprec.NewVec2(0.0, 5.0),
					dprec.NewVec2(0.0, -5.0),
				),
				placement2d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})
	})
})
