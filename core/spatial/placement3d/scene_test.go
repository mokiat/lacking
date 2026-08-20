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

// boxAt builds an axis-aligned box centered at the given coordinates with the
// given half-extent along every axis.
func boxAt(x, y, z, half float64) shape3d.Box {
	return shape3d.NewBox(
		dprec.NewVec3(x, y, z),
		shape3d.IdentityRotation(),
		dprec.NewVec3(half, half, half),
	)
}

// planeMesh builds a mesh made of two triangles forming a quad in the XZ plane
// (at the given y), centered at the given point and spanning halfSize in each
// of the X and Z directions.
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
			Expect(objID).NotTo(Equal(placement3d.NilObjectID))

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

	Describe("terrain management", func() {
		It("creates terrains", func() {
			terrainID := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			Expect(terrainID).NotTo(Equal(placement3d.NilTerrainID))
		})

		It("stores and updates user data", func() {
			terrainID := scene.CreateTerrain(placement3d.TerrainInfo[string]{
				UserData: "first",
			})
			Expect(scene.GetTerrainUserData(terrainID)).To(Equal("first"))

			scene.SetTerrainUserData(terrainID, "second")
			Expect(scene.GetTerrainUserData(terrainID)).To(Equal("second"))
		})

		It("stores and updates terrain shape user data", func() {
			terrainID := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			shapeID := scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh:     planeMesh(0.0, 0.0, 0.0, 5.0),
				UserData: "a",
			})
			Expect(scene.GetTerrainShapeUserData(shapeID)).To(Equal("a"))

			scene.SetTerrainShapeUserData(shapeID, "b")
			Expect(scene.GetTerrainShapeUserData(shapeID)).To(Equal("b"))
		})

		It("maps a terrain shape back to its owning terrain", func() {
			terrainID := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			shapeID := scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			Expect(scene.GetTerrainForShape(shapeID)).To(Equal(terrainID))
		})

		It("reuses the indices of deleted terrains", func() {
			first := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			scene.DeleteTerrain(first)
			second := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			Expect(second).To(Equal(first))
		})

		It("panics when attaching an empty mesh", func() {
			terrainID := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			Expect(func() {
				scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
					Mesh: shape3d.NewMesh(nil),
				})
			}).To(Panic())
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

		It("exposes a box iterator", func() {
			scene.AttachBox(objID, placement3d.BoxInfo[string]{
				Box: boxAt(0.0, 0.0, 0.0, 1.0),
			})

			count := 0
			for range scene.BoxIter(placement3d.Filter{}) {
				count++
			}
			Expect(count).To(Equal(1))
		})

		It("stores and updates shape user data", func() {
			shapeID := scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere:   sphereAt(0.0, 0.0, 0.0, 1.0),
				UserData: "a",
			})
			Expect(scene.GetObjectShapeUserData(shapeID)).To(Equal("a"))

			scene.SetObjectShapeUserData(shapeID, "b")
			Expect(scene.GetObjectShapeUserData(shapeID)).To(Equal("b"))
		})

		It("maps a shape back to its owning object", func() {
			shapeID := scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			Expect(scene.GetObjectForShape(shapeID)).To(Equal(objID))
		})

		It("removes a deleted shape from iteration", func() {
			shapeID := scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.DeleteObjectShape(shapeID)

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
		var objID placement3d.ObjectID

		BeforeEach(func() {
			objID = scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Filtering: placement3d.FilterInfo{
					RejectGroup: 7,
					SourceMask:  opt.V(uint32(0b01)),
				},
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
		})

		countSpheres := func(filter placement3d.Filter) int {
			count := 0
			scene.EachSphere(filter, func(shape3d.Sphere) bool {
				count++
				return true
			})
			return count
		}

		It("yields shapes that occupy a layer of the mask", func() {
			Expect(countSpheres(placement3d.Filter{Mask: 0b01})).To(Equal(1))
		})

		It("skips shapes that occupy no layer of the mask", func() {
			Expect(countSpheres(placement3d.Filter{Mask: 0b10})).To(BeZero())
		})

		It("yields everything for the zero mask", func() {
			Expect(countSpheres(placement3d.Filter{Mask: 0})).To(Equal(1))
		})

		It("yields everything for the full mask", func() {
			Expect(countSpheres(placement3d.Filter{
				Mask: placement3d.FullMask,
			})).To(Equal(1))
		})

		It("skips shapes that share the reject group", func() {
			Expect(countSpheres(placement3d.Filter{RejectGroup: 7})).To(BeZero())
		})

		It("yields shapes that have a different reject group", func() {
			Expect(countSpheres(placement3d.Filter{RejectGroup: 8})).To(Equal(1))
		})

		It("yields everything for the zero filter", func() {
			Expect(countSpheres(placement3d.Filter{})).To(Equal(1))
		})
	})

	Describe("CollectObjectIntersections", func() {
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

		collect := func() placement3d.ObjectContactList {
			var contacts placement3d.ObjectContactList
			scene.CollectObjectIntersections(contacts.AddContact)
			return contacts
		}

		It("reports a single contact between two overlapping shapes", func() {
			first, second := attachOverlappingSpheres()
			contacts := collect()
			Expect(contacts).To(HaveLen(1))
			Expect([]placement3d.ObjectID{
				contacts[0].SourceObjectID,
				contacts[0].TargetObjectID,
			}).To(ConsistOf(first, second))
		})

		It("reports object IDs that agree with the shape IDs", func() {
			attachOverlappingSpheres()
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
				Filtering: placement3d.FilterInfo{RejectGroup: 7},
				Sphere:    sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(second, placement3d.SphereInfo[string]{
				Filtering: placement3d.FilterInfo{RejectGroup: 7},
				Sphere:    sphereAt(1.0, 0.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("does not report shapes whose masks do not overlap", func() {
			first := scene.CreateObject(placement3d.ObjectInfo[string]{})
			second := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(first, placement3d.SphereInfo[string]{
				Filtering: placement3d.FilterInfo{
					SourceMask: opt.V(uint32(0b01)),
					TargetMask: opt.V(uint32(0b01)),
				},
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(second, placement3d.SphereInfo[string]{
				Filtering: placement3d.FilterInfo{
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

		It("reports a contact between two overlapping boxes", func() {
			first := scene.CreateObject(placement3d.ObjectInfo[string]{})
			second := scene.CreateObject(placement3d.ObjectInfo[string]{
				Position: opt.V(dprec.NewVec3(1.5, 0.0, 0.0)),
			})
			scene.AttachBox(first, placement3d.BoxInfo[string]{
				Box: boxAt(0.0, 0.0, 0.0, 2.0),
			})
			scene.AttachBox(second, placement3d.BoxInfo[string]{
				Box: boxAt(0.0, 0.0, 0.0, 2.0),
			})
			Expect(collect()).To(HaveLen(1))
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
				Box: boxAt(0.0, 0.0, 0.0, 1.0),
			})
			Expect(collect()).To(HaveLen(1))
		})

		It("does not report contacts with terrain shapes", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, -0.5, 0.0, 1.0),
			})
			terrainID := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("reports every overlapping pair of two multi-shape objects", func() {
			first := scene.CreateObject(placement3d.ObjectInfo[string]{})
			second := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(first, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(first, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.5, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(second, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.25, 0.0, 0.0, 1.0),
			})
			// Both shapes of the first object overlap the single shape of the
			// second one, while the two shapes of the first object are not
			// tested against each other.
			Expect(collect()).To(HaveLen(2))
		})

		It("does not produce phantom contacts after index reuse", func() {
			first, second := attachOverlappingSpheres()
			Expect(collect()).To(HaveLen(1))

			scene.DeleteObject(first)
			scene.DeleteObject(second)
			Expect(collect()).To(BeEmpty())

			// Recreate, reusing both the object and the shape indices.
			third := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(third, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			Expect(collect()).To(BeEmpty())
		})
	})

	Describe("CollectTerrainIntersections", func() {
		var objID placement3d.ObjectID
		var terrainID placement3d.TerrainID

		BeforeEach(func() {
			objID = scene.CreateObject(placement3d.ObjectInfo[string]{})
			terrainID = scene.CreateTerrain(placement3d.TerrainInfo[string]{})
		})

		collect := func() placement3d.TerrainContactList {
			var contacts placement3d.TerrainContactList
			scene.CollectTerrainIntersections(contacts.AddContact)
			return contacts
		}

		It("reports a contact between a shape and an overlapping mesh", func() {
			// The plane's triangles face -Y, so the sphere is placed just below
			// the plane (on the front side) where it overlaps and is pushed out.
			shapeID := scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, -0.5, 0.0, 1.0),
			})
			meshID := scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})

			contacts := collect()
			Expect(contacts).To(HaveLen(1))
			Expect(contacts[0].SourceObjectID).To(Equal(objID))
			Expect(contacts[0].SourceShapeID).To(Equal(shapeID))
			Expect(contacts[0].TargetTerrainID).To(Equal(terrainID))
			Expect(contacts[0].TargetShapeID).To(Equal(meshID))

			// The contact normal must push the sphere out the front (-Y) side,
			// never inward into the mesh.
			Expect(contacts[0].TargetNormal.Y).To(BeNumerically("<", 0.0))
		})

		It("does not report a shape overlapping a mesh from behind", func() {
			// A sphere on the +Y (back) side of the -Y-facing plane would have
			// to be pushed further inward to separate, which the mesh logic
			// prevents.
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.5, 0.0, 1.0),
			})
			scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("does not report a shape disjoint from a mesh", func() {
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 10.0, 0.0, 1.0),
			})
			scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("reports a single contact even when many triangles overlap", func() {
			// A large box overlaps both triangles of the plane, yet only the
			// deepest contact is reported.
			scene.AttachBox(objID, placement3d.BoxInfo[string]{
				Box: boxAt(0.0, -0.5, 0.0, 2.0),
			})
			scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			Expect(collect()).To(HaveLen(1))
		})

		It("reports a contact per terrain shape", func() {
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, -0.5, 0.0, 1.0),
			})
			scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, -0.1, 0.0, 5.0),
			})
			Expect(collect()).To(HaveLen(2))
		})

		It("does not report shapes that share a reject group", func() {
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Filtering: placement3d.FilterInfo{RejectGroup: 7},
				Sphere:    sphereAt(0.0, -0.5, 0.0, 1.0),
			})
			scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Filtering: placement3d.FilterInfo{RejectGroup: 7},
				Mesh:      planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("does not report shapes whose masks do not overlap", func() {
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Filtering: placement3d.FilterInfo{
					SourceMask: opt.V(uint32(0b01)),
					TargetMask: opt.V(uint32(0b01)),
				},
				Sphere: sphereAt(0.0, -0.5, 0.0, 1.0),
			})
			scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Filtering: placement3d.FilterInfo{
					SourceMask: opt.V(uint32(0b10)),
					TargetMask: opt.V(uint32(0b10)),
				},
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			Expect(collect()).To(BeEmpty())
		})

		It("stops reporting once the terrain is deleted", func() {
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, -0.5, 0.0, 1.0),
			})
			scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			Expect(collect()).To(HaveLen(1))

			scene.DeleteTerrain(terrainID)
			Expect(collect()).To(BeEmpty())
		})

		It("stops reporting once the terrain shape is deleted", func() {
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, -0.5, 0.0, 1.0),
			})
			meshID := scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			Expect(collect()).To(HaveLen(1))

			scene.DeleteTerrainShape(meshID)
			Expect(collect()).To(BeEmpty())
		})

		It("reports a contact per object shape", func() {
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(-1.0, -0.5, 0.0, 1.0),
			})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(1.0, -0.5, 0.0, 1.0),
			})
			scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			Expect(collect()).To(HaveLen(2))
		})

		It("reattaches correctly after terrain shape index reuse", func() {
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, -0.5, 0.0, 1.0),
			})
			meshID := scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			scene.DeleteTerrainShape(meshID)

			other := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			reusedID := scene.AttachMesh(other, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})

			contacts := collect()
			Expect(contacts).To(HaveLen(1))
			Expect(contacts[0].TargetTerrainID).To(Equal(other))
			Expect(contacts[0].TargetShapeID).To(Equal(reusedID))
			Expect(scene.GetTerrainForShape(reusedID)).To(Equal(other))
		})

		It("tracks object movement into and out of terrain contact", func() {
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, -0.5, 0.0, 1.0),
			})
			scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})
			Expect(collect()).To(HaveLen(1))

			scene.SetObjectTransform(objID, shape3d.TranslationTransform(
				dprec.NewVec3(0.0, -20.0, 0.0),
			))
			Expect(collect()).To(BeEmpty())

			scene.SetObjectTransform(objID, shape3d.TranslationTransform(
				dprec.NewVec3(0.0, 0.0, 0.0),
			))
			Expect(collect()).To(HaveLen(1))
		})
	})

	Describe("shape id spaces", func() {
		It("keeps object and terrain shape ids independent", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			shapeID := scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere:   sphereAt(0.0, 0.0, 0.0, 1.0),
				UserData: "object-shape",
			})
			terrainID := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			meshID := scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh:     planeMesh(0.0, 0.0, 0.0, 5.0),
				UserData: "terrain-shape",
			})

			// Both are the first shape of their kind, hence they share the raw
			// index while remaining distinct references.
			Expect(int32(shapeID)).To(Equal(int32(meshID)))
			Expect(scene.GetObjectShapeUserData(shapeID)).To(Equal("object-shape"))
			Expect(scene.GetTerrainShapeUserData(meshID)).To(Equal("terrain-shape"))
		})
	})

	Describe("sphere queries", func() {
		It("reports a sphere overlapping an object shape", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			shapeID := scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			contact, ok := scene.CheckSphereObjectIntersection(
				sphereAt(1.5, 0.0, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceObjectID).To(Equal(placement3d.NilObjectID))
			Expect(contact.SourceShapeID).To(Equal(placement3d.NilObjectShapeID))
			Expect(contact.TargetObjectID).To(Equal(objID))
			Expect(contact.TargetShapeID).To(Equal(shapeID))
		})

		It("returns false for a sphere disjoint from every object shape", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckSphereObjectIntersection(
				sphereAt(10.0, 0.0, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("honors the query mask", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Filtering: placement3d.FilterInfo{
					SourceMask: opt.V(uint32(0b01)),
				},
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckSphereObjectIntersection(
				sphereAt(1.5, 0.0, 0.0, 1.0),
				placement3d.Filter{Mask: 0b10},
			)
			Expect(ok).To(BeFalse())
		})

		It("honors the query reject group", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Filtering: placement3d.FilterInfo{
					RejectGroup: 7,
				},
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckSphereObjectIntersection(
				sphereAt(1.5, 0.0, 0.0, 1.0),
				placement3d.Filter{RejectGroup: 7},
			)
			Expect(ok).To(BeFalse())
		})

		It("reports a sphere overlapping a terrain shape from the front", func() {
			terrainID := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			meshID := scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})

			// The plane faces -Y, so approach it from below (the front side).
			contact, ok := scene.CheckSphereTerrainIntersection(
				sphereAt(0.0, -0.5, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceShapeID).To(Equal(placement3d.NilObjectShapeID))
			Expect(contact.TargetTerrainID).To(Equal(terrainID))
			Expect(contact.TargetShapeID).To(Equal(meshID))
			Expect(contact.TargetNormal.Y).To(BeNumerically("<", 0.0))
		})

		It("does not report a sphere overlapping a terrain shape from behind", func() {
			terrainID := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})

			_, ok := scene.CheckSphereTerrainIntersection(
				sphereAt(0.0, 0.5, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("keeps object and terrain queries separate", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckSphereTerrainIntersection(
				sphereAt(0.5, 0.0, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})
	})

	Describe("box queries", func() {
		It("reports a box overlapping an object shape", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			shapeID := scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			contact, ok := scene.CheckBoxObjectIntersection(
				boxAt(1.5, 0.0, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceShapeID).To(Equal(placement3d.NilObjectShapeID))
			Expect(contact.TargetObjectID).To(Equal(objID))
			Expect(contact.TargetShapeID).To(Equal(shapeID))
		})

		It("returns false for a box disjoint from every object shape", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckBoxObjectIntersection(
				boxAt(10.0, 0.0, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("reports a box overlapping a terrain shape from the front", func() {
			terrainID := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			meshID := scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})

			// The plane faces -Y, so approach it from below (the front side).
			contact, ok := scene.CheckBoxTerrainIntersection(
				boxAt(0.0, -0.5, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetTerrainID).To(Equal(terrainID))
			Expect(contact.TargetShapeID).To(Equal(meshID))
			Expect(contact.TargetNormal.Y).To(BeNumerically("<", 0.0))
		})

		It("does not report a box overlapping a terrain shape from behind", func() {
			terrainID := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})

			_, ok := scene.CheckBoxTerrainIntersection(
				boxAt(0.0, 0.5, 0.0, 1.0),
				placement3d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})
	})

	Describe("segment queries", func() {
		It("collects every object shape a segment passes through", func() {
			near := scene.CreateObject(placement3d.ObjectInfo[string]{})
			far := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(near, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(far, placement3d.SphereInfo[string]{
				Sphere: sphereAt(4.0, 0.0, 0.0, 1.0),
			})

			var contacts placement3d.ObjectContactList
			scene.CollectSegmentObjectIntersections(
				shape3d.NewSegment(
					dprec.NewVec3(-5.0, 0.0, 0.0),
					dprec.NewVec3(9.0, 0.0, 0.0),
				),
				placement3d.Filter{},
				contacts.AddContact,
			)
			Expect(contacts).To(HaveLen(2))
		})

		It("finds an object shape crossed by the segment", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			shapeID := scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			contact, ok := scene.CheckSegmentObjectIntersection(
				shape3d.NewSegment(
					dprec.NewVec3(-5.0, 0.0, 0.0),
					dprec.NewVec3(5.0, 0.0, 0.0),
				),
				placement3d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceShapeID).To(Equal(placement3d.NilObjectShapeID))
			Expect(contact.TargetObjectID).To(Equal(objID))
			Expect(contact.TargetShapeID).To(Equal(shapeID))
		})

		It("finds the nearest of two object shapes crossed by the segment", func() {
			near := scene.CreateObject(placement3d.ObjectInfo[string]{})
			far := scene.CreateObject(placement3d.ObjectInfo[string]{})
			nearShapeID := scene.AttachSphere(near, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})
			scene.AttachSphere(far, placement3d.SphereInfo[string]{
				Sphere: sphereAt(4.0, 0.0, 0.0, 1.0),
			})

			contact, ok := scene.CheckSegmentObjectIntersection(
				shape3d.NewSegment(
					dprec.NewVec3(-5.0, 0.0, 0.0),
					dprec.NewVec3(9.0, 0.0, 0.0),
				),
				placement3d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetShapeID).To(Equal(nearShapeID))
		})

		It("finds a terrain shape crossed by the segment", func() {
			terrainID := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			meshID := scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})

			contact, ok := scene.CheckSegmentTerrainIntersection(
				shape3d.NewSegment(
					dprec.NewVec3(2.0, -5.0, 0.0),
					dprec.NewVec3(2.0, 5.0, 0.0),
				),
				placement3d.Filter{},
			)
			Expect(ok).To(BeTrue())
			Expect(contact.SourceShapeID).To(Equal(placement3d.NilObjectShapeID))
			Expect(contact.TargetTerrainID).To(Equal(terrainID))
			Expect(contact.TargetShapeID).To(Equal(meshID))
		})

		It("returns false when the segment misses everything", func() {
			objID := scene.CreateObject(placement3d.ObjectInfo[string]{})
			scene.AttachSphere(objID, placement3d.SphereInfo[string]{
				Sphere: sphereAt(0.0, 0.0, 0.0, 1.0),
			})

			_, ok := scene.CheckSegmentObjectIntersection(
				shape3d.NewSegment(
					dprec.NewVec3(-5.0, 5.0, 0.0),
					dprec.NewVec3(5.0, 5.0, 0.0),
				),
				placement3d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})

		It("keeps object and terrain queries separate", func() {
			terrainID := scene.CreateTerrain(placement3d.TerrainInfo[string]{})
			scene.AttachMesh(terrainID, placement3d.MeshInfo[string]{
				Mesh: planeMesh(0.0, 0.0, 0.0, 5.0),
			})

			_, ok := scene.CheckSegmentObjectIntersection(
				shape3d.NewSegment(
					dprec.NewVec3(0.0, 5.0, 0.0),
					dprec.NewVec3(0.0, -5.0, 0.0),
				),
				placement3d.Filter{},
			)
			Expect(ok).To(BeFalse())
		})
	})
})
