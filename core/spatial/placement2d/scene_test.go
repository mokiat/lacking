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
// y == 0 by default), centered at the given point and spanning halfSize in the
// X direction. The edge is wound so that its normal faces -Y.
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
			Expect(objID).NotTo(Equal(placement2d.InvalidObjectID))

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
				Rectangle: shape2d.NewRectangle(
					dprec.ZeroVec2(),
					shape2d.IdentityRotation(),
					dprec.NewVec2(1.0, 1.0),
				),
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

		It("stores and updates shape user data", func() {
			shapeID := scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle:   circleAt(0.0, 0.0, 1.0),
				UserData: "a",
			})
			Expect(scene.GetShapeUserData(shapeID)).To(Equal("a"))

			scene.SetShapeUserData(shapeID, "b")
			Expect(scene.GetShapeUserData(shapeID)).To(Equal("b"))
		})

		It("maps a shape back to its owning object", func() {
			shapeID := scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			Expect(scene.GetShapeObject(shapeID)).To(Equal(objID))
		})

		It("removes a deleted shape from iteration", func() {
			shapeID := scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.DeleteShape(shapeID)

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
		It("filters by layer mask", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Filtering: placement2d.FilterInfo{
					SourceMask: opt.V(uint32(0b01)),
				},
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			matching := 0
			scene.EachCircle(placement2d.Filter{Mask: opt.V(uint32(0b01))}, func(shape2d.Circle) bool {
				matching++
				return true
			})
			Expect(matching).To(Equal(1))

			nonMatching := 0
			scene.EachCircle(placement2d.Filter{Mask: opt.V(uint32(0b10))}, func(shape2d.Circle) bool {
				nonMatching++
				return true
			})
			Expect(nonMatching).To(BeZero())
		})
	})

	Describe("CollectIntersections", func() {
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

		collect := func() placement2d.ContactList {
			var contacts placement2d.ContactList
			scene.CollectIntersections(contacts.AddContact)
			return contacts
		}

		It("reports a single contact between two overlapping shapes", func() {
			first, second := attachOverlappingCircles()
			contacts := collect()
			Expect(contacts).To(HaveLen(1))
			Expect([]placement2d.ObjectID{
				scene.GetShapeObject(contacts[0].SourceShapeID),
				scene.GetShapeObject(contacts[0].TargetShapeID),
			}).To(ConsistOf(first, second))
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

		It("reports a contact between a shape and an overlapping mesh", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			// The line's edge faces -Y, so the circle is placed just below the
			// line (on the front side) where it overlaps and is pushed out.
			shapeID := scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, -0.5, 1.0),
			})
			meshID := scene.CreateMesh(placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			contacts := collect()
			Expect(contacts).To(HaveLen(1))
			Expect(contacts[0].SourceShapeID).To(Equal(shapeID))
			Expect(contacts[0].TargetShapeID).To(Equal(placement2d.InvalidShapeID))
			Expect(contacts[0].TargetMeshID).To(Equal(meshID))

			contact := contacts[0].Contact
			// The contact normal must push the circle out the front (-Y) side,
			// never inward into the mesh.
			Expect(contact.TargetNormal.Y).To(BeNumerically("<", 0.0))
		})

		It("does not report a shape overlapping a mesh from behind", func() {
			// A circle on the +Y (back) side of the -Y-facing line would have to
			// be pushed further inward to separate, which the mesh logic
			// prevents.
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.5, 1.0),
			})
			scene.CreateMesh(placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("does not report a shape disjoint from a mesh", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 10.0, 1.0),
			})
			scene.CreateMesh(placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})
			Expect(collect()).To(BeEmpty())
		})
	})

	Describe("shape-vs-shape with rectangles", func() {
		It("reports a contact between two overlapping rectangles", func() {
			first := scene.CreateObject(placement2d.ObjectInfo[string]{})
			second := scene.CreateObject(placement2d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec2(1.5, 0.0)),
			})
			scene.AttachRectangle(first, placement2d.RectangleInfo[string]{
				Rectangle: shape2d.NewRectangle(
					dprec.ZeroVec2(),
					shape2d.IdentityRotation(),
					dprec.NewVec2(2.0, 2.0),
				),
			})
			scene.AttachRectangle(second, placement2d.RectangleInfo[string]{
				Rectangle: shape2d.NewRectangle(
					dprec.ZeroVec2(),
					shape2d.IdentityRotation(),
					dprec.NewVec2(2.0, 2.0),
				),
			})

			var contacts placement2d.ContactList
			scene.CollectIntersections(contacts.AddContact)
			Expect(contacts).To(HaveLen(1))
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
				Rectangle: shape2d.NewRectangle(
					dprec.ZeroVec2(),
					shape2d.IdentityRotation(),
					dprec.NewVec2(1.0, 1.0),
				),
			})

			var contacts placement2d.ContactList
			scene.CollectIntersections(contacts.AddContact)
			Expect(contacts).To(HaveLen(1))
		})
	})

	Describe("CheckCircleIntersection", func() {
		It("reports a circle overlapping a scene shape", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			contact, ok := scene.CheckCircleIntersection(
				circleAt(1.5, 0.0, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceShapeID).To(Equal(placement2d.InvalidShapeID))
			Expect(scene.GetShapeObject(contact.TargetShapeID)).To(Equal(objID))
			Expect(contact.TargetMeshID).To(Equal(placement2d.InvalidMeshID))
		})

		It("returns false for a circle disjoint from every shape", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckCircleIntersection(
				circleAt(10.0, 0.0, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("reports a circle overlapping a mesh from the front", func() {
			meshID := scene.CreateMesh(placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			// The line faces -Y, so approach it from below (the front side).
			contact, ok := scene.CheckCircleIntersection(
				circleAt(0.0, -0.5, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetShapeID).To(Equal(placement2d.InvalidShapeID))
			Expect(contact.TargetMeshID).To(Equal(meshID))
			Expect(contact.TargetNormal.Y).To(BeNumerically("<", 0.0))
		})

		It("does not report a circle overlapping a mesh from behind", func() {
			scene.CreateMesh(placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			_, ok := scene.CheckCircleIntersection(
				circleAt(0.0, 0.5, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("skips dynamic shapes when SkipDynamic is set", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckCircleIntersection(
				circleAt(1.5, 0.0, 1.0),
				placement2d.Filter{SkipDynamic: true},
			)
			Expect(ok).To(BeFalse())
		})

		It("skips static meshes when SkipStatic is set", func() {
			scene.CreateMesh(placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			_, ok := scene.CheckCircleIntersection(
				circleAt(0.0, -0.5, 1.0),
				placement2d.Filter{SkipStatic: true},
			)
			Expect(ok).To(BeFalse())
		})
	})

	Describe("CheckRectangleIntersection", func() {
		It("reports a rectangle overlapping a scene shape", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			contact, ok := scene.CheckRectangleIntersection(
				rectangleAt(1.5, 0.0, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceShapeID).To(Equal(placement2d.InvalidShapeID))
			Expect(scene.GetShapeObject(contact.TargetShapeID)).To(Equal(objID))
			Expect(contact.TargetMeshID).To(Equal(placement2d.InvalidMeshID))
		})

		It("returns false for a rectangle disjoint from every shape", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckRectangleIntersection(
				rectangleAt(10.0, 0.0, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("reports a rectangle overlapping a mesh from the front", func() {
			meshID := scene.CreateMesh(placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			// The line faces -Y, so approach it from below (the front side).
			contact, ok := scene.CheckRectangleIntersection(
				rectangleAt(0.0, -0.5, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetShapeID).To(Equal(placement2d.InvalidShapeID))
			Expect(contact.TargetMeshID).To(Equal(meshID))
			Expect(contact.TargetNormal.Y).To(BeNumerically("<", 0.0))
		})

		It("does not report a rectangle overlapping a mesh from behind", func() {
			scene.CreateMesh(placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			_, ok := scene.CheckRectangleIntersection(
				rectangleAt(0.0, 0.5, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("skips dynamic shapes when SkipDynamic is set", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckRectangleIntersection(
				rectangleAt(1.5, 0.0, 1.0),
				placement2d.Filter{SkipDynamic: true},
			)
			Expect(ok).To(BeFalse())
		})

		It("skips static meshes when SkipStatic is set", func() {
			scene.CreateMesh(placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			_, ok := scene.CheckRectangleIntersection(
				rectangleAt(0.0, -0.5, 1.0),
				placement2d.Filter{SkipStatic: true},
			)
			Expect(ok).To(BeFalse())
		})
	})

	Describe("CollectSegmentIntersections", func() {
		It("collects every shape a segment passes through", func() {
			near := scene.CreateObject(placement2d.ObjectInfo[string]{})
			far := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(near, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(far, placement2d.CircleInfo[string]{
				Circle: circleAt(4.0, 0.0, 1.0),
			})

			var contacts placement2d.ContactList
			scene.CollectSegmentIntersections(
				shape2d.NewSegment(
					dprec.NewVec2(-5.0, 0.0),
					dprec.NewVec2(9.0, 0.0),
				),
				placement2d.Filter{},
				contacts.AddContact,
			)
			Expect(contacts).To(HaveLen(2))
		})
	})

	Describe("CheckSegmentIntersection", func() {
		It("finds a circle crossed by the segment", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			contact, ok := scene.CheckSegmentIntersection(
				shape2d.NewSegment(
					dprec.NewVec2(-5.0, 0.0),
					dprec.NewVec2(5.0, 0.0),
				),
				placement2d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceShapeID).To(Equal(placement2d.InvalidShapeID))
			Expect(scene.GetShapeObject(contact.TargetShapeID)).To(Equal(objID))
		})

		It("finds a mesh crossed by the segment", func() {
			meshID := scene.CreateMesh(placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			contact, ok := scene.CheckSegmentIntersection(
				shape2d.NewSegment(
					dprec.NewVec2(2.0, -5.0),
					dprec.NewVec2(2.0, 5.0),
				),
				placement2d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetShapeID).To(Equal(placement2d.InvalidShapeID))
			Expect(contact.TargetMeshID).To(Equal(meshID))
		})

		It("returns false when the segment misses everything", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckSegmentIntersection(
				shape2d.NewSegment(
					dprec.NewVec2(-5.0, 5.0),
					dprec.NewVec2(5.0, 5.0),
				),
				placement2d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("skips dynamic shapes when SkipDynamic is set", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckSegmentIntersection(
				shape2d.NewSegment(
					dprec.NewVec2(-5.0, 0.0),
					dprec.NewVec2(5.0, 0.0),
				),
				placement2d.Filter{SkipDynamic: true},
			)
			Expect(ok).To(BeFalse())
		})

		It("skips static meshes when SkipStatic is set", func() {
			scene.CreateMesh(placement2d.MeshInfo[string]{
				Mesh: lineMesh(0.0, 0.0, 5.0),
			})

			_, ok := scene.CheckSegmentIntersection(
				shape2d.NewSegment(
					dprec.NewVec2(0.0, 5.0),
					dprec.NewVec2(0.0, -5.0),
				),
				placement2d.Filter{SkipStatic: true},
			)
			Expect(ok).To(BeFalse())
		})
	})
})
