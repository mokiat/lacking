package isec3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("BoxMesh", func() {
	// An axis-aligned unit cube centered at the origin, spanning [-1,1] per axis.
	var box shape3d.Box

	// A floor-like triangle at the given height, facing up (+Z), so the box
	// only collides with it from above.
	horizontalTriangle := func(z float64) shape3d.Triangle {
		return shape3d.Triangle{
			A: dprec.NewVec3(-2.0, -2.0, z),
			B: dprec.NewVec3(2.0, -2.0, z),
			C: dprec.NewVec3(0.0, 2.0, z),
		}
	}

	// Two triangles slicing the box from below: the deep one (z=-0.5) is
	// penetrated by 0.5 and the shallow one (z=-0.9) only by 0.1.
	deepTriangle := horizontalTriangle(-0.5)
	shallowTriangle := horizontalTriangle(-0.9)
	// A triangle off in another region, never reached by the box.
	farTriangle := shape3d.Triangle{
		A: dprec.NewVec3(18.0, 18.0, -0.5),
		B: dprec.NewVec3(22.0, 18.0, -0.5),
		C: dprec.NewVec3(20.0, 22.0, -0.5),
	}

	var mesh shape3d.Mesh

	BeforeEach(func() {
		box = shape3d.Box{
			Center:     dprec.NewVec3(0.0, 0.0, 0.0),
			Rotation:   shape3d.IdentityRotation(),
			HalfWidth:  1.0,
			HalfHeight: 1.0,
			HalfLength: 1.0,
		}
		mesh = shape3d.Mesh{
			Triangles: []shape3d.Triangle{shallowTriangle, deepTriangle},
		}
	})

	Describe("CheckBoxMesh", func() {
		It("returns true when the box intersects a triangle", func() {
			Expect(isec3d.CheckBoxMesh(box, mesh)).To(BeTrue())
		})

		It("returns true when only a later triangle in the list is intersected", func() {
			offset := shape3d.Mesh{
				Triangles: []shape3d.Triangle{farTriangle, deepTriangle},
			}
			Expect(isec3d.CheckBoxMesh(box, offset)).To(BeTrue())
		})

		It("returns false when the box misses every triangle", func() {
			below := shape3d.Mesh{
				Triangles: []shape3d.Triangle{horizontalTriangle(-2.0)},
			}
			Expect(isec3d.CheckBoxMesh(box, below)).To(BeFalse())
		})

		It("returns false when the box is behind every triangle", func() {
			// The triangle slices the box, but the box center is behind its plane,
			// so the triangle is back-face culled.
			behind := shape3d.Mesh{
				Triangles: []shape3d.Triangle{horizontalTriangle(0.5)},
			}
			Expect(isec3d.CheckBoxMesh(box, behind)).To(BeFalse())
		})

		It("returns false for an empty mesh", func() {
			Expect(isec3d.CheckBoxMesh(box, shape3d.Mesh{})).To(BeFalse())
		})
	})

	Describe("ResolveBoxMesh", func() {
		resolve := func(b shape3d.Box, m shape3d.Mesh) (shape3d.Contact, bool) {
			var sink shape3d.LastContact
			isec3d.ResolveBoxMesh(b, m, sink.AddContact)
			return sink.Contact()
		}

		It("yields the contact for the most deeply penetrated triangle", func() {
			contact, ok := resolve(box, mesh)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
			Expect(contact.TargetPoint).To(haveApproxVec3Coords(0.0, 0.0, -0.5))
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("selects the deepest contact regardless of triangle order", func() {
			reordered := shape3d.Mesh{
				Triangles: []shape3d.Triangle{deepTriangle, shallowTriangle},
			}
			contact, ok := resolve(box, reordered)
			Expect(ok).To(BeTrue())
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
			Expect(contact.TargetPoint).To(haveApproxVec3Coords(0.0, 0.0, -0.5))
		})

		It("ignores triangles outside the box's reach", func() {
			withFar := shape3d.Mesh{
				Triangles: []shape3d.Triangle{deepTriangle, shallowTriangle, farTriangle},
			}
			contact, ok := resolve(box, withFar)
			Expect(ok).To(BeTrue())
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("reports the same contact as resolving the hit triangle directly", func() {
			single := shape3d.Mesh{
				Triangles: []shape3d.Triangle{deepTriangle},
			}

			meshContact, ok := resolve(box, single)
			Expect(ok).To(BeTrue())

			var sink shape3d.LastContact
			isec3d.ResolveBoxTriangle(box, deepTriangle, sink.AddContact)
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

		It("does not yield a contact when the box misses every triangle", func() {
			below := shape3d.Mesh{
				Triangles: []shape3d.Triangle{horizontalTriangle(-2.0)},
			}
			_, ok := resolve(box, below)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact when the box is behind every triangle", func() {
			behind := shape3d.Mesh{
				Triangles: []shape3d.Triangle{horizontalTriangle(0.5)},
			}
			_, ok := resolve(box, behind)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact for an empty mesh", func() {
			_, ok := resolve(box, shape3d.Mesh{})
			Expect(ok).To(BeFalse())
		})

		It("reports a unit normal", func() {
			contact, ok := resolve(box, mesh)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("separates the box when it is moved by Depth along the normal", func() {
			contact, ok := resolve(box, mesh)
			Expect(ok).To(BeTrue())

			moved := box
			moved.Center = dprec.Vec3Sum(box.Center, dprec.Vec3Prod(contact.TargetNormal, contact.Depth))
			// Moving out by Depth resolves the deepest contact; the deep triangle
			// is no longer penetrated more than touching.
			resolved, ok := resolve(moved, shape3d.Mesh{Triangles: []shape3d.Triangle{deepTriangle}})
			if ok {
				Expect(resolved.Depth).To(BeNumerically("~", 0.0, 1e-6))
			}
		})
	})
})
