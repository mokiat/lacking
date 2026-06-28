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

var _ = Describe("Scene", func() {
	var scene *placement2d.Scene[string, string]

	BeforeEach(func() {
		scene = placement2d.NewScene[string, string](placement2d.SceneSettings{
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

		It("yields attached meshes", func() {
			scene.AttachMesh(objID, placement2d.MeshInfo[string]{
				Mesh: shape2d.NewMesh([]shape2d.Edge{
					shape2d.NewEdge(
						dprec.NewVec2(0.0, 0.0),
						dprec.NewVec2(1.0, 0.0),
					),
				}),
			})

			count := 0
			scene.EachMesh(placement2d.Filter{}, func(shape2d.Mesh) bool {
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

		It("exposes a mesh iterator", func() {
			scene.AttachMesh(objID, placement2d.MeshInfo[string]{
				Mesh: shape2d.NewMesh([]shape2d.Edge{
					shape2d.NewEdge(
						dprec.NewVec2(0.0, 0.0),
						dprec.NewVec2(1.0, 0.0),
					),
				}),
			})

			count := 0
			for range scene.MeshIter(placement2d.Filter{}) {
				count++
			}
			Expect(count).To(Equal(1))
		})

		It("stores and updates shape user data", func() {
			shapeID := scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				ShapeInfo: placement2d.ShapeInfo[string]{UserData: "a"},
				Circle:    circleAt(0.0, 0.0, 1.0),
			})
			Expect(scene.GetShapeUserData(shapeID)).To(Equal("a"))

			scene.SetShapeUserData(shapeID, "b")
			Expect(scene.GetShapeUserData(shapeID)).To(Equal("b"))
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
	})

	Describe("shape iteration filters", func() {
		It("excludes static shapes when SkipStatic is set", func() {
			staticObj := scene.CreateObject(placement2d.ObjectInfo[string]{Static: true})
			dynamicObj := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(staticObj, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(dynamicObj, placement2d.CircleInfo[string]{
				Circle: circleAt(5.0, 0.0, 1.0),
			})

			count := 0
			scene.EachCircle(placement2d.Filter{SkipStatic: true}, func(shape2d.Circle) bool {
				count++
				return true
			})
			Expect(count).To(Equal(1))
		})

		It("excludes dynamic shapes when SkipDynamic is set", func() {
			staticObj := scene.CreateObject(placement2d.ObjectInfo[string]{Static: true})
			dynamicObj := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(staticObj, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(dynamicObj, placement2d.CircleInfo[string]{
				Circle: circleAt(5.0, 0.0, 1.0),
			})

			count := 0
			scene.EachCircle(placement2d.Filter{SkipDynamic: true}, func(shape2d.Circle) bool {
				count++
				return true
			})
			Expect(count).To(Equal(1))
		})

		It("filters by layer mask", func() {
			objID := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				ShapeInfo: placement2d.ShapeInfo[string]{
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
		attachOverlappingCircles := func(firstStatic, secondStatic bool) (placement2d.ObjectID, placement2d.ObjectID) {
			first := scene.CreateObject(placement2d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec2(0.0, 0.0)),
				Static:   firstStatic,
			})
			second := scene.CreateObject(placement2d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec2(1.5, 0.0)),
				Static:   secondStatic,
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

		It("reports a single contact between two overlapping dynamic shapes", func() {
			first, second := attachOverlappingCircles(false, false)
			contacts := collect()
			Expect(contacts).To(HaveLen(1))
			Expect([]placement2d.ObjectID{
				contacts[0].SourceObjectID,
				contacts[0].TargetObjectID,
			}).To(ConsistOf(first, second))
		})

		It("does not report contacts between two static shapes", func() {
			attachOverlappingCircles(true, true)
			Expect(collect()).To(BeEmpty())
		})

		It("reports a contact between a dynamic and a static shape", func() {
			attachOverlappingCircles(false, true)
			Expect(collect()).To(HaveLen(1))
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
				ShapeInfo: placement2d.ShapeInfo[string]{RejectGroup: 7},
				Circle:    circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(second, placement2d.CircleInfo[string]{
				ShapeInfo: placement2d.ShapeInfo[string]{RejectGroup: 7},
				Circle:    circleAt(1.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("does not report shapes whose masks do not overlap", func() {
			first := scene.CreateObject(placement2d.ObjectInfo[string]{})
			second := scene.CreateObject(placement2d.ObjectInfo[string]{})
			scene.AttachCircle(first, placement2d.CircleInfo[string]{
				ShapeInfo: placement2d.ShapeInfo[string]{
					SourceMask: opt.V(uint32(0b01)),
					TargetMask: opt.V(uint32(0b01)),
				},
				Circle: circleAt(0.0, 0.0, 1.0),
			})
			scene.AttachCircle(second, placement2d.CircleInfo[string]{
				ShapeInfo: placement2d.ShapeInfo[string]{
					SourceMask: opt.V(uint32(0b10)),
					TargetMask: opt.V(uint32(0b10)),
				},
				Circle: circleAt(1.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("stops reporting once a deleted object's shapes are gone", func() {
			first, _ := attachOverlappingCircles(false, false)
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
	})

	Describe("CheckCircleIntersection", func() {
		var objID placement2d.ObjectID
		var shapeID placement2d.ShapeID

		BeforeEach(func() {
			objID = scene.CreateObject(placement2d.ObjectInfo[string]{})
			shapeID = scene.AttachCircle(objID, placement2d.CircleInfo[string]{
				Circle: circleAt(0.0, 0.0, 1.0),
			})
		})

		It("finds an overlapping scene circle as the target", func() {
			contact, ok := scene.CheckCircleIntersection(
				circleAt(1.5, 0.0, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceObjectID).To(Equal(placement2d.InvalidObjectID))
			Expect(contact.SourceShapeID).To(Equal(placement2d.InvalidShapeID))
			Expect(contact.TargetObjectID).To(Equal(objID))
			Expect(contact.TargetShapeID).To(Equal(shapeID))
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("returns false when nothing overlaps", func() {
			_, ok := scene.CheckCircleIntersection(
				circleAt(10.0, 0.0, 1.0),
				placement2d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("skips static shapes when SkipStatic is set", func() {
			staticObj := scene.CreateObject(placement2d.ObjectInfo[string]{Static: true})
			scene.AttachCircle(staticObj, placement2d.CircleInfo[string]{
				Circle: circleAt(5.0, 0.0, 1.0),
			})

			_, ok := scene.CheckCircleIntersection(
				circleAt(5.0, 0.0, 1.0),
				placement2d.Filter{SkipStatic: true},
			)
			Expect(ok).To(BeFalse())
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
			Expect(contact.SourceObjectID).To(Equal(placement2d.InvalidObjectID))
			Expect(contact.TargetObjectID).To(Equal(objID))
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
	})
})
