package placement3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/placement3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// sphereAt builds a sphere with the given center coordinates and radius.
func sphereAt(x, y, z, radius float64) shape3d.Sphere {
	return shape3d.Sphere{
		Center: dprec.NewVec3(x, y, z),
		Radius: radius,
	}
}

var _ = Describe("Scene", func() {
	var scene *placement3d.Scene[string, string]

	BeforeEach(func() {
		scene = placement3d.NewScene[string, string](placement3d.SceneSettings{
			Size:     opt.V(128.0),
			MaxDepth: opt.V[uint32](3),
		})
	})

	Describe("object management", func() {
		It("creates objects placed at the origin by default", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			Expect(objID).NotTo(Equal(placement3d.InvalidObjectID))

			transform := scene.GetObjectTransform(objID)
			Expect(transform.Translation).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
		})

		It("honors the provided position", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec3(1.0, 2.0, 3.0)),
			})
			transform := scene.GetObjectTransform(objID)
			Expect(transform.Translation).To(dprectest.HaveVec3Coords(1.0, 2.0, 3.0))
		})

		It("stores and updates user data", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{
				UserData: "first",
			})
			Expect(scene.GetObjectUserData(objID)).To(Equal("first"))

			scene.SetObjectUserData(objID, "second")
			Expect(scene.GetObjectUserData(objID)).To(Equal("second"))
		})

		It("relocates objects via SetObjectTransform", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.SetObjectTransform(objID, shape3d.TranslationTransform(
				dprec.NewVec3(5.0, 6.0, 7.0),
			))
			transform := scene.GetObjectTransform(objID)
			Expect(transform.Translation).To(dprectest.HaveVec3Coords(5.0, 6.0, 7.0))
		})

		It("reuses the indices of deleted objects", func() {
			first := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.DeleteObject(first)
			second := scene.CreateObject(placement3d.ObjectInfo[string]{})
			Expect(second).To(Equal(first))
		})
	})

	Describe("shape iteration", func() {
		var objID placement3d.ObjectID

		BeforeEach(func() {
			objID = scene.CreateObject(placement3d.ObjectInfo[string]{})
		})

		It("yields attached spheres in world space", func() {
			scene.SetObjectTransform(objID, shape3d.TranslationTransform(
				dprec.NewVec3(10.0, 0.0, 0.0),
			))
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 2.0),
			})

			var found []shape3d.Sphere
			scene.EachSphere(placement3d.Filter{}, func(s shape3d.Sphere) bool {
				found = append(found, s)
				return true
			})
			Expect(found).To(HaveLen(1))
			Expect(found[0].Center).To(dprectest.HaveVec3Coords(10.0, 0.0, 0.0))
			Expect(found[0].Radius).To(Equal(2.0))
		})

		It("yields attached boxes", func() {
			scene.AttachBox(objID, placement3d.BoxInfo[string]{
				Box: shape3d.NewBox(
					dprec.ZeroVec3(),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				),
			})

			count := 0
			scene.EachBox(placement3d.Filter{}, func(shape3d.Box) bool {
				count++
				return true
			})
			Expect(count).To(Equal(1))
		})

		It("yields attached meshes", func() {
			scene.AttachMesh(objID, placement3d.MeshInfo[string]{
				Mesh: shape3d.NewMesh([]shape3d.Triangle{
					shape3d.NewTriangle(
						dprec.NewVec3(0.0, 0.0, 0.0),
						dprec.NewVec3(1.0, 0.0, 0.0),
						dprec.NewVec3(0.0, 1.0, 0.0),
					),
				}),
			})

			count := 0
			scene.EachMesh(placement3d.Filter{}, func(shape3d.Mesh) bool {
				count++
				return true
			})
			Expect(count).To(Equal(1))
		})

		It("exposes a sphere iterator", func() {
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			count := 0
			for range scene.SphereIter(placement3d.Filter{}) {
				count++
			}
			Expect(count).To(Equal(1))
		})

		It("stores and updates shape user data", func() {
			shapeID := scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				ShapeInfo: placement3d.ShapeInfo[string]{UserData: "a"},
				Sphere:    sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			Expect(scene.GetShapeUserData(shapeID)).To(Equal("a"))

			scene.SetShapeUserData(shapeID, "b")
			Expect(scene.GetShapeUserData(shapeID)).To(Equal("b"))
		})

		It("removes a deleted shape from iteration", func() {
			shapeID := scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.DeleteShape(shapeID)

			count := 0
			scene.EachSphere(placement3d.Filter{}, func(shape3d.Sphere) bool {
				count++
				return true
			})
			Expect(count).To(BeZero())
		})

		It("stops iteration when the callback returns false", func() {
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(3.0, 0.0, 0.0, 1.0),
			})

			count := 0
			scene.EachSphere(placement3d.Filter{}, func(shape3d.Sphere) bool {
				count++
				return false
			})
			Expect(count).To(Equal(1))
		})
	})

	Describe("shape iteration filters", func() {
		It("excludes static shapes when SkipStatic is set", func() {
			staticObj := scene.CreateObject(placement3d.ObjectInfo[string]{Static: true})
			dynamicObj := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(staticObj, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(dynamicObj, placement3d.SphereInfo[string]{
				Sphere: sphereAt(5.0, 0.0, 0.0, 1.0),
			})

			count := 0
			scene.EachSphere(placement3d.Filter{SkipStatic: true}, func(shape3d.Sphere) bool {
				count++
				return true
			})
			Expect(count).To(Equal(1))
		})

		It("excludes dynamic shapes when SkipDynamic is set", func() {
			staticObj := scene.CreateObject(placement3d.ObjectInfo[string]{Static: true})
			dynamicObj := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(staticObj, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(dynamicObj, placement3d.SphereInfo[string]{
				Sphere: sphereAt(5.0, 0.0, 0.0, 1.0),
			})

			count := 0
			scene.EachSphere(placement3d.Filter{SkipDynamic: true}, func(shape3d.Sphere) bool {
				count++
				return true
			})
			Expect(count).To(Equal(1))
		})

		It("filters by layer mask", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				ShapeInfo: placement3d.ShapeInfo[string]{
					SourceMask: opt.V(uint32(0b01)),
				},
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			matching := 0
			scene.EachSphere(placement3d.Filter{Mask: opt.V(uint32(0b01))}, func(shape3d.Sphere) bool {
				matching++
				return true
			})
			Expect(matching).To(Equal(1))

			nonMatching := 0
			scene.EachSphere(placement3d.Filter{Mask: opt.V(uint32(0b10))}, func(shape3d.Sphere) bool {
				nonMatching++
				return true
			})
			Expect(nonMatching).To(BeZero())
		})
	})

	Describe("CollectIntersections", func() {
		// attachOverlappingSpheres places two unit spheres 1.5 apart (so they
		// overlap) on freshly created objects and returns the object IDs.
		attachOverlappingSpheres := func(firstStatic, secondStatic bool) (placement3d.ObjectID, placement3d.ObjectID) {
			first := scene.CreateObject(placement3d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec3(0.0, 0.0, 0.0)),
				Static:   firstStatic,
			})
			second := scene.CreateObject(placement3d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec3(1.5, 0.0, 0.0)),
				Static:   secondStatic,
			})
			scene.AttachSphere(first, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(second, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			return first, second
		}

		collect := func() placement3d.ContactList {
			var contacts placement3d.ContactList
			scene.CollectIntersections(contacts.AddContact)
			return contacts
		}

		It("reports a single contact between two overlapping dynamic shapes", func() {
			first, second := attachOverlappingSpheres(false, false)
			contacts := collect()
			Expect(contacts).To(HaveLen(1))
			Expect([]placement3d.ObjectID{
				contacts[0].SourceObjectID,
				contacts[0].TargetObjectID,
			}).To(ConsistOf(first, second))
		})

		It("does not report contacts between two static shapes", func() {
			attachOverlappingSpheres(true, true)
			Expect(collect()).To(BeEmpty())
		})

		It("reports a contact between a dynamic and a static shape", func() {
			attachOverlappingSpheres(false, true)
			Expect(collect()).To(HaveLen(1))
		})

		It("does not report contacts between shapes of the same object", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(1.0, 0.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("does not report disjoint shapes", func() {
			first := scene.CreateObject(placement3d.ObjectInfo[string]{})
			second := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(first, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(second, placement3d.SphereInfo[string]{
				Sphere: sphereAt(10.0, 0.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("does not report shapes that share a reject group", func() {
			first := scene.CreateObject(placement3d.ObjectInfo[string]{})
			second := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(first, placement3d.SphereInfo[string]{
				ShapeInfo: placement3d.ShapeInfo[string]{RejectGroup: 7},
				Sphere:    sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(second, placement3d.SphereInfo[string]{
				ShapeInfo: placement3d.ShapeInfo[string]{RejectGroup: 7},
				Sphere:    sphereAt(1.0, 0.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("does not report shapes whose masks do not overlap", func() {
			first := scene.CreateObject(placement3d.ObjectInfo[string]{})
			second := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(first, placement3d.SphereInfo[string]{
				ShapeInfo: placement3d.ShapeInfo[string]{
					SourceMask: opt.V(uint32(0b01)),
					TargetMask: opt.V(uint32(0b01)),
				},
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(second, placement3d.SphereInfo[string]{
				ShapeInfo: placement3d.ShapeInfo[string]{
					SourceMask: opt.V(uint32(0b10)),
					TargetMask: opt.V(uint32(0b10)),
				},
				Sphere: sphereAt(1.0, 0.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("stops reporting once a deleted object's shapes are gone", func() {
			first, _ := attachOverlappingSpheres(false, false)
			Expect(collect()).To(HaveLen(1))

			scene.DeleteObject(first)
			Expect(collect()).To(BeEmpty())
		})

		It("reflects object movement in the broadphase", func() {
			scene.CreateObject(placement3d.ObjectInfo[string]{})
			first := scene.CreateObject(placement3d.ObjectInfo[string]{})
			second := scene.CreateObject(placement3d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec3(10.0, 0.0, 0.0)),
			})
			scene.AttachSphere(first, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(second, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())

			scene.SetObjectTransform(second, shape3d.TranslationTransform(
				dprec.NewVec3(1.5, 0.0, 0.0),
			))
			Expect(collect()).To(HaveLen(1))
		})
	})

	Describe("CheckSphereIntersection", func() {
		var objID placement3d.ObjectID
		var shapeID placement3d.ShapeID

		BeforeEach(func() {
			objID = scene.CreateObject(placement3d.ObjectInfo[string]{})
			shapeID = scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
		})

		It("finds an overlapping scene sphere as the target", func() {
			contact, ok := scene.CheckSphereIntersection(
				sphereAt(1.5, 0.0, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceObjectID).To(Equal(placement3d.InvalidObjectID))
			Expect(contact.SourceShapeID).To(Equal(placement3d.InvalidShapeID))
			Expect(contact.TargetObjectID).To(Equal(objID))
			Expect(contact.TargetShapeID).To(Equal(shapeID))
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("returns false when nothing overlaps", func() {
			_, ok := scene.CheckSphereIntersection(
				sphereAt(10.0, 0.0, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("skips static shapes when SkipStatic is set", func() {
			staticObj := scene.CreateObject(placement3d.ObjectInfo[string]{Static: true})
			scene.AttachSphere(staticObj, placement3d.SphereInfo[string]{
				Sphere: sphereAt(5.0, 0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckSphereIntersection(
				sphereAt(5.0, 0.0, 0.0, 1.0),
				placement3d.Filter{SkipStatic: true},
			)
			Expect(ok).To(BeFalse())
		})
	})

	Describe("CheckSegmentIntersection", func() {
		It("finds a sphere crossed by the segment", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			contact, ok := scene.CheckSegmentIntersection(
				shape3d.NewSegment(
					dprec.NewVec3(-5.0, 0.0, 0.0),
					dprec.NewVec3(5.0, 0.0, 0.0),
				),
				placement3d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceObjectID).To(Equal(placement3d.InvalidObjectID))
			Expect(contact.TargetObjectID).To(Equal(objID))
		})

		It("returns false when the segment misses everything", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckSegmentIntersection(
				shape3d.NewSegment(
					dprec.NewVec3(-5.0, 5.0, 0.0),
					dprec.NewVec3(5.0, 5.0, 0.0),
				),
				placement3d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})
	})
})
