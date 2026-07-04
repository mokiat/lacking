package isec3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("SegmentMesh", func() {
	// A mesh of two parallel triangles, both wound counter-clockwise when viewed
	// from +Z so their normals point along +Z. The near triangle lies in z=0 and
	// the far one in z=-2; both cover the region x>=0, y>=0, x+y<=4. The near
	// triangle is listed first.
	var mesh shape3d.Mesh

	newSegment := func(ax, ay, az, bx, by, bz float64) shape3d.Segment {
		return shape3d.Segment{
			A: dprec.NewVec3(ax, ay, az),
			B: dprec.NewVec3(bx, by, bz),
		}
	}

	nearTriangle := shape3d.Triangle{
		A: dprec.NewVec3(0.0, 0.0, 0.0),
		B: dprec.NewVec3(4.0, 0.0, 0.0),
		C: dprec.NewVec3(0.0, 4.0, 0.0),
	}
	farTriangle := shape3d.Triangle{
		A: dprec.NewVec3(0.0, 0.0, -2.0),
		B: dprec.NewVec3(4.0, 0.0, -2.0),
		C: dprec.NewVec3(0.0, 4.0, -2.0),
	}

	BeforeEach(func() {
		mesh = shape3d.Mesh{
			Triangles: []shape3d.Triangle{nearTriangle, farTriangle},
		}
	})

	Describe("CheckSegmentMesh", func() {
		It("returns true when the segment crosses a triangle from the front", func() {
			segment := newSegment(1.0, 1.0, 1.0, 1.0, 1.0, -3.0)
			Expect(isec3d.CheckSegmentMesh(segment, mesh)).To(BeTrue())
		})

		It("returns true when only a later triangle in the list is crossed", func() {
			// The first triangle is off in another region; only the second is hit.
			offset := shape3d.Mesh{
				Triangles: []shape3d.Triangle{
					{
						A: dprec.NewVec3(10.0, 10.0, 0.0),
						B: dprec.NewVec3(14.0, 10.0, 0.0),
						C: dprec.NewVec3(10.0, 14.0, 0.0),
					},
					nearTriangle,
				},
			}
			segment := newSegment(1.0, 1.0, 1.0, 1.0, 1.0, -1.0)
			Expect(isec3d.CheckSegmentMesh(segment, offset)).To(BeTrue())
		})

		It("returns false when the segment misses every triangle", func() {
			segment := newSegment(10.0, 10.0, 1.0, 10.0, 10.0, -3.0)
			Expect(isec3d.CheckSegmentMesh(segment, mesh)).To(BeFalse())
		})

		It("returns false when the segment approaches only back faces", func() {
			// Travelling along +Z, both triangles are reached from behind.
			segment := newSegment(1.0, 1.0, -3.0, 1.0, 1.0, 1.0)
			Expect(isec3d.CheckSegmentMesh(segment, mesh)).To(BeFalse())
		})

		It("returns false for an empty mesh", func() {
			segment := newSegment(1.0, 1.0, 1.0, 1.0, 1.0, -3.0)
			Expect(isec3d.CheckSegmentMesh(segment, shape3d.Mesh{})).To(BeFalse())
		})
	})

	Describe("ResolveSegmentMesh", func() {
		resolve := func(segment shape3d.Segment, m shape3d.Mesh) (shape3d.Contact, bool) {
			var sink shape3d.LastContact
			isec3d.ResolveSegmentMesh(segment, m, sink.AddContact)
			return sink.Contact()
		}

		It("yields the entry on the first triangle the segment reaches", func() {
			// The segment crosses z=0 at (1,1,0) and z=-2 at (1,1,-2); the entry
			// nearest A is the z=0 crossing.
			contact, ok := resolve(newSegment(1.0, 1.0, 1.0, 1.0, 1.0, -3.0), mesh)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(1.0, 1.0, 0.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
			// Crossing is a quarter of the way along the segment, leaving 0.75
			// beyond it.
			Expect(contact.Depth).To(BeNumerically("~", 0.75, 1e-6))
		})

		It("selects the earliest crossing regardless of triangle order", func() {
			// The far triangle is listed first, but the nearer z=0 crossing must
			// still win, as it has the greater Depth.
			reordered := shape3d.Mesh{
				Triangles: []shape3d.Triangle{farTriangle, nearTriangle},
			}
			contact, ok := resolve(newSegment(1.0, 1.0, 1.0, 1.0, 1.0, -3.0), reordered)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec3Coords(1.0, 1.0, 0.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.75, 1e-6))
		})

		It("reports the same contact as resolving the hit triangle directly", func() {
			// A segment that reaches only the near triangle.
			segment := newSegment(1.0, 1.0, 1.0, 1.0, 1.0, -1.0)

			meshContact, ok := resolve(segment, mesh)
			Expect(ok).To(BeTrue())

			var sink shape3d.LastContact
			isec3d.ResolveSegmentTriangle(segment, nearTriangle, sink.AddContact)
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

		It("does not yield a contact when the segment misses every triangle", func() {
			_, ok := resolve(newSegment(10.0, 10.0, 1.0, 10.0, 10.0, -3.0), mesh)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact for a back-facing approach", func() {
			_, ok := resolve(newSegment(1.0, 1.0, -3.0, 1.0, 1.0, 1.0), mesh)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact for an empty mesh", func() {
			_, ok := resolve(newSegment(1.0, 1.0, 1.0, 1.0, 1.0, -3.0), shape3d.Mesh{})
			Expect(ok).To(BeFalse())
		})
	})
})
