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

// planeMesh builds a mesh made of two triangles forming a quad in the XZ plane
// (at y == 0), centered at the given point and spanning halfSize in each of
// the X and Z directions.
func planeMesh(x, y, z, halfSize float64) shape3d.Mesh {
	a := dprec.NewVec3(x-halfSize, y, z-halfSize)
	b := dprec.NewVec3(x+halfSize, y, z-halfSize)
	c := dprec.NewVec3(x+halfSize, y, z+halfSize)
	d := dprec.NewVec3(x-halfSize, y, z+halfSize)
	return shape3d.NewMesh([]shape3d.Triangle{
		shape3d.NewTriangle(a, b, c),
		shape3d.NewTriangle(a, c, d),
	})
}

var _ = Describe("Scene", func() {
	var scene *placement3d.Scene[string, string, string]

	BeforeEach(func() {
		scene = placement3d.NewScene[string, string, string](placement3d.SceneSettings{
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

		It("maps a shape back to its owning object", func() {
			shapeID := scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			Expect(scene.GetShapeObject(shapeID)).To(Equal(objID))
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

		It("moves every shape of an object with multiple shapes", func() {
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(2.0, 0.0, 0.0, 1.0),
			})
			scene.SetObjectTransform(objID, shape3d.TranslationTransform(
				dprec.NewVec3(0.0, 10.0, 0.0),
			))

			var centers []dprec.Vec3
			scene.EachSphere(placement3d.Filter{}, func(s shape3d.Sphere) bool {
				centers = append(centers, s.Center)
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
		attachOverlappingSpheres := func() (placement3d.ObjectID, placement3d.ObjectID) {
			first := scene.CreateObject(placement3d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec3(0.0, 0.0, 0.0)),
			})
			second := scene.CreateObject(placement3d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec3(1.5, 0.0, 0.0)),
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

		It("reports a single contact between two overlapping shapes", func() {
			first, second := attachOverlappingSpheres()
			contacts := collect()
			Expect(contacts).To(HaveLen(1))
			Expect([]placement3d.ObjectID{
				scene.GetShapeObject(contacts[0].SourceShapeID),
				scene.GetShapeObject(contacts[0].TargetShapeID),
			}).To(ConsistOf(first, second))
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
			first, _ := attachOverlappingSpheres()
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

		It("reports a contact between a shape and an overlapping mesh", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			// The plane's triangles face -Y, so the sphere is placed just below
			// the plane (on the front side) where it overlaps and is pushed out.
			shapeID := scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, -0.5, 0.0, 1.0),
			})
			meshID := scene.CreateMesh(placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})

			contacts := collect()
			Expect(contacts).To(HaveLen(1))
			Expect(contacts[0].SourceShapeID).To(Equal(shapeID))
			Expect(contacts[0].TargetShapeID).To(Equal(placement3d.InvalidShapeID))
			Expect(contacts[0].TargetMeshID).To(Equal(meshID))

			contact := contacts[0].Contact
			// The contact normal must push the sphere out the front (-Y) side,
			// never inward into the mesh.
			Expect(contact.TargetNormal.Y).To(BeNumerically("<", 0.0))
		})

		It("does not report a shape overlapping a mesh from behind", func() {
			// A sphere on the +Y (back) side of the -Y-facing plane would have
			// to be pushed further inward to separate, which the mesh logic
			// prevents.
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.5, 0.0, 1.0),
			})
			scene.CreateMesh(placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("does not report a shape disjoint from a mesh", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 10.0, 0.0, 1.0),
			})
			scene.CreateMesh(placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			Expect(collect()).To(BeEmpty())
		})
	})

	Describe("shape-vs-shape with boxes", func() {
		It("reports a contact between two overlapping boxes", func() {
			first := scene.CreateObject(placement3d.ObjectInfo[string]{})
			second := scene.CreateObject(placement3d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec3(1.5, 0.0, 0.0)),
			})
			scene.AttachBox(first, placement3d.BoxInfo[string]{
				Box: shape3d.NewBox(
					dprec.ZeroVec3(),
					shape3d.IdentityRotation(),
					dprec.NewVec3(2.0, 2.0, 2.0),
				),
			})
			scene.AttachBox(second, placement3d.BoxInfo[string]{
				Box: shape3d.NewBox(
					dprec.ZeroVec3(),
					shape3d.IdentityRotation(),
					dprec.NewVec3(2.0, 2.0, 2.0),
				),
			})

			var contacts placement3d.ContactList
			scene.CollectIntersections(contacts.AddContact)
			Expect(contacts).To(HaveLen(1))
		})

		It("reports a contact between an overlapping sphere and box", func() {
			first := scene.CreateObject(placement3d.ObjectInfo[string]{})
			second := scene.CreateObject(placement3d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec3(1.0, 0.0, 0.0)),
			})
			scene.AttachSphere(first, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachBox(second, placement3d.BoxInfo[string]{
				Box: shape3d.NewBox(
					dprec.ZeroVec3(),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				),
			})

			var contacts placement3d.ContactList
			scene.CollectIntersections(contacts.AddContact)
			Expect(contacts).To(HaveLen(1))
		})
	})

	Describe("CheckSphereIntersection", func() {
		It("reports a sphere overlapping a scene shape", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			contact, ok := scene.CheckSphereIntersection(
				sphereAt(1.5, 0.0, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceShapeID).To(Equal(placement3d.InvalidShapeID))
			Expect(scene.GetShapeObject(contact.TargetShapeID)).To(Equal(objID))
			Expect(contact.TargetMeshID).To(Equal(placement3d.InvalidMeshID))
		})

		It("returns false for a sphere disjoint from every shape", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckSphereIntersection(
				sphereAt(10.0, 0.0, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("reports a sphere overlapping a mesh from the front", func() {
			meshID := scene.CreateMesh(placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})

			// The plane faces -Y, so approach it from below (the front side).
			contact, ok := scene.CheckSphereIntersection(
				sphereAt(0.0, -0.5, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetShapeID).To(Equal(placement3d.InvalidShapeID))
			Expect(contact.TargetMeshID).To(Equal(meshID))
			Expect(contact.TargetNormal.Y).To(BeNumerically("<", 0.0))
		})

		It("does not report a sphere overlapping a mesh from behind", func() {
			scene.CreateMesh(placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})

			_, ok := scene.CheckSphereIntersection(
				sphereAt(0.0, 0.5, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("skips dynamic shapes when SkipDynamic is set", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckSphereIntersection(
				sphereAt(1.5, 0.0, 0.0, 1.0),
				placement3d.Filter{SkipDynamic: true},
			)
			Expect(ok).To(BeFalse())
		})

		It("skips static meshes when SkipStatic is set", func() {
			scene.CreateMesh(placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})

			_, ok := scene.CheckSphereIntersection(
				sphereAt(0.0, -0.5, 0.0, 1.0),
				placement3d.Filter{SkipStatic: true},
			)
			Expect(ok).To(BeFalse())
		})
	})

	Describe("CollectSegmentIntersections", func() {
		It("collects every shape a segment passes through", func() {
			near := scene.CreateObject(placement3d.ObjectInfo[string]{})
			far := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(near, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(far, placement3d.SphereInfo[string]{
				Sphere: sphereAt(4.0, 0.0, 0.0, 1.0),
			})

			var contacts placement3d.ContactList
			scene.CollectSegmentIntersections(
				shape3d.NewSegment(
					dprec.NewVec3(-5.0, 0.0, 0.0),
					dprec.NewVec3(9.0, 0.0, 0.0),
				),
				placement3d.Filter{},
				contacts.AddContact,
			)
			Expect(contacts).To(HaveLen(2))
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
			Expect(contact.SourceShapeID).To(Equal(placement3d.InvalidShapeID))
			Expect(scene.GetShapeObject(contact.TargetShapeID)).To(Equal(objID))
		})

		It("finds a mesh crossed by the segment", func() {
			meshID := scene.CreateMesh(placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})

			contact, ok := scene.CheckSegmentIntersection(
				shape3d.NewSegment(
					dprec.NewVec3(2.0, -5.0, 0.0),
					dprec.NewVec3(2.0, 5.0, 0.0),
				),
				placement3d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetShapeID).To(Equal(placement3d.InvalidShapeID))
			Expect(contact.TargetMeshID).To(Equal(meshID))
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

		It("skips dynamic shapes when SkipDynamic is set", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckSegmentIntersection(
				shape3d.NewSegment(
					dprec.NewVec3(-5.0, 0.0, 0.0),
					dprec.NewVec3(5.0, 0.0, 0.0),
				),
				placement3d.Filter{SkipDynamic: true},
			)
			Expect(ok).To(BeFalse())
		})

		It("skips static meshes when SkipStatic is set", func() {
			scene.CreateMesh(placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})

			_, ok := scene.CheckSegmentIntersection(
				shape3d.NewSegment(
					dprec.NewVec3(0.0, 5.0, 0.0),
					dprec.NewVec3(0.0, -5.0, 0.0),
				),
				placement3d.Filter{SkipStatic: true},
			)
			Expect(ok).To(BeFalse())
		})
	})
})
