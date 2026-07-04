package isec3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("SphereMesh", func() {
	// Two parallel triangles facing +Z, both covering x>=0, y>=0, x+y<=4. The
	// floor lies in z=0 and the upper one in z=0.5, so a sphere centered above
	// them penetrates the upper triangle more deeply than the floor.
	var mesh shape3d.Mesh

	newSphere := func(x, y, z, radius float64) shape3d.Sphere {
		return shape3d.Sphere{
			Center: dprec.NewVec3(x, y, z),
			Radius: radius,
		}
	}

	floorTriangle := shape3d.Triangle{
		A: dprec.NewVec3(0.0, 0.0, 0.0),
		B: dprec.NewVec3(4.0, 0.0, 0.0),
		C: dprec.NewVec3(0.0, 4.0, 0.0),
	}
	upperTriangle := shape3d.Triangle{
		A: dprec.NewVec3(0.0, 0.0, 0.5),
		B: dprec.NewVec3(4.0, 0.0, 0.5),
		C: dprec.NewVec3(0.0, 4.0, 0.5),
	}
	// A triangle off in another region, never reached by the spheres below.
	farTriangle := shape3d.Triangle{
		A: dprec.NewVec3(20.0, 20.0, 0.0),
		B: dprec.NewVec3(24.0, 20.0, 0.0),
		C: dprec.NewVec3(20.0, 24.0, 0.0),
	}

	BeforeEach(func() {
		mesh = shape3d.Mesh{
			Triangles: []shape3d.Triangle{floorTriangle, upperTriangle},
		}
	})

	Describe("CheckSphereMesh", func() {
		It("returns true when the sphere intersects a triangle from the front", func() {
			Expect(isec3d.CheckSphereMesh(newSphere(1.0, 1.0, 1.0, 1.5), mesh)).To(BeTrue())
		})

		It("returns true when only a later triangle in the list is intersected", func() {
			offset := shape3d.Mesh{
				Triangles: []shape3d.Triangle{farTriangle, floorTriangle},
			}
			Expect(isec3d.CheckSphereMesh(newSphere(1.0, 1.0, 1.0, 1.5), offset)).To(BeTrue())
		})

		It("returns false when the sphere misses every triangle", func() {
			Expect(isec3d.CheckSphereMesh(newSphere(20.0, 0.0, 1.0, 1.5), mesh)).To(BeFalse())
		})

		It("returns false when the sphere is behind every triangle", func() {
			Expect(isec3d.CheckSphereMesh(newSphere(1.0, 1.0, -1.0, 1.5), mesh)).To(BeFalse())
		})

		It("returns false for an empty mesh", func() {
			Expect(isec3d.CheckSphereMesh(newSphere(1.0, 1.0, 1.0, 1.5), shape3d.Mesh{})).To(BeFalse())
		})
	})

	Describe("ResolveSphereMesh", func() {
		resolve := func(sphere shape3d.Sphere, m shape3d.Mesh) (shape3d.Contact, bool) {
			var sink shape3d.LastContact
			isec3d.ResolveSphereMesh(sphere, m, sink.AddContact)
			return sink.Contact()
		}

		It("yields the contact for the most deeply penetrated triangle", func() {
			// The sphere penetrates the upper (z=0.5) triangle by 1.0 and the
			// floor (z=0) by 0.5, so the upper triangle wins.
			contact, ok := resolve(newSphere(1.0, 1.0, 1.0, 1.5), mesh)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(1.0, 1.0, 0.5))
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
			Expect(contact.Depth).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("selects the deepest contact regardless of triangle order", func() {
			reordered := shape3d.Mesh{
				Triangles: []shape3d.Triangle{upperTriangle, floorTriangle},
			}
			contact, ok := resolve(newSphere(1.0, 1.0, 1.0, 1.5), reordered)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(1.0, 1.0, 0.5))
			Expect(contact.Depth).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("ignores triangles outside the sphere's reach", func() {
			withFar := shape3d.Mesh{
				Triangles: []shape3d.Triangle{floorTriangle, upperTriangle, farTriangle},
			}
			contact, ok := resolve(newSphere(1.0, 1.0, 1.0, 1.5), withFar)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(1.0, 1.0, 0.5))
			Expect(contact.Depth).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("reports the same contact as resolving the hit triangle directly", func() {
			single := shape3d.Mesh{
				Triangles: []shape3d.Triangle{floorTriangle},
			}
			sphere := newSphere(1.0, 1.0, 1.0, 1.5)

			meshContact, ok := resolve(sphere, single)
			Expect(ok).To(BeTrue())

			var sink shape3d.LastContact
			isec3d.ResolveSphereTriangle(sphere, floorTriangle, sink.AddContact)
			triangleContact, ok := sink.Contact()
			Expect(ok).To(BeTrue())

			Expect(meshContact.TargetPoint).To(dprectest.HaveVec3Coords(
				triangleContact.TargetPoint.X,
				triangleContact.TargetPoint.Y,
				triangleContact.TargetPoint.Z,
			))
			Expect(meshContact.TargetNormal).To(dprectest.HaveVec3Coords(
				triangleContact.TargetNormal.X,
				triangleContact.TargetNormal.Y,
				triangleContact.TargetNormal.Z,
			))
			Expect(meshContact.Depth).To(BeNumerically("~", triangleContact.Depth, 1e-6))
		})

		It("does not yield a contact when the sphere misses every triangle", func() {
			_, ok := resolve(newSphere(20.0, 0.0, 1.0, 1.5), mesh)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact when the sphere is behind every triangle", func() {
			_, ok := resolve(newSphere(1.0, 1.0, -1.0, 1.5), mesh)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact for an empty mesh", func() {
			_, ok := resolve(newSphere(1.0, 1.0, 1.0, 1.5), shape3d.Mesh{})
			Expect(ok).To(BeFalse())
		})

		It("reports a unit normal", func() {
			contact, ok := resolve(newSphere(1.0, 1.0, 1.0, 1.5), mesh)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("removes the overlap when the sphere is moved by Depth along the normal", func() {
			sphere := newSphere(1.0, 1.0, 1.0, 1.5)
			contact, ok := resolve(sphere, mesh)
			Expect(ok).To(BeTrue())

			movedCenter := dprec.Vec3Sum(sphere.Center, dprec.Vec3Prod(contact.TargetNormal, contact.Depth))
			// After the move the contact point sits exactly at radius, so the
			// sphere only just touches the triangle.
			distance := dprec.Vec3Diff(movedCenter, contact.TargetPoint).Length()
			Expect(distance).To(BeNumerically("~", sphere.Radius, 1e-6))
		})
	})
})
