package isec2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("RectangleMesh", func() {
	// A counter-clockwise-wound square spanning (0,0) to (4,4). Because the
	// winding is counter-clockwise, each edge's right-hand normal faces outward,
	// away from the center at (2,2). The mesh therefore behaves as a one-sided
	// boundary that only collides from the outside.
	var square shape2d.Mesh

	newRectangle := func(x, y, halfWidth, halfHeight float64) shape2d.Rectangle {
		return shape2d.NewRectangle(
			dprec.NewVec2(x, y),
			shape2d.IdentityRotation(),
			dprec.NewVec2(halfWidth, halfHeight),
		)
	}

	BeforeEach(func() {
		square = shape2d.NewMesh([]shape2d.Edge{
			shape2d.NewEdge(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(4.0, 0.0)), // bottom, normal -Y
			shape2d.NewEdge(dprec.NewVec2(4.0, 0.0), dprec.NewVec2(4.0, 4.0)), // right, normal +X
			shape2d.NewEdge(dprec.NewVec2(4.0, 4.0), dprec.NewVec2(0.0, 4.0)), // top, normal +Y
			shape2d.NewEdge(dprec.NewVec2(0.0, 4.0), dprec.NewVec2(0.0, 0.0)), // left, normal -X
		})
	})

	Describe("CheckRectangleMesh", func() {
		It("returns true for a rectangle overlapping an edge from outside", func() {
			Expect(isec2d.CheckRectangleMesh(newRectangle(2.0, -0.5, 1.0, 1.0), square)).To(BeTrue())
		})

		It("returns true for a rectangle that just touches an edge", func() {
			Expect(isec2d.CheckRectangleMesh(newRectangle(2.0, -1.0, 1.0, 1.0), square)).To(BeTrue())
		})

		It("returns false for a rectangle far outside the mesh", func() {
			Expect(isec2d.CheckRectangleMesh(newRectangle(10.0, 10.0, 1.0, 1.0), square)).To(BeFalse())
		})

		It("returns false for a rectangle fully inside the mesh", func() {
			// Every edge is back-face culled, since the center is on the interior
			// side of all of them.
			Expect(isec2d.CheckRectangleMesh(newRectangle(2.0, 2.0, 0.5, 0.5), square)).To(BeFalse())
		})

		It("returns false against an empty mesh", func() {
			Expect(isec2d.CheckRectangleMesh(newRectangle(2.0, -0.5, 1.0, 1.0), shape2d.NewMesh(nil))).To(BeFalse())
		})
	})

	Describe("ResolveRectangleMesh", func() {
		resolve := func(rectangle shape2d.Rectangle, mesh shape2d.Mesh) (shape2d.Contact, bool) {
			var sink shape2d.LastContact
			isec2d.ResolveRectangleMesh(rectangle, mesh, sink.AddContact)
			return sink.Contact()
		}

		It("yields a contact against an overlapping edge", func() {
			contact, ok := resolve(newRectangle(2.0, -0.5, 1.0, 1.0), square)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(2.0, 0.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(0.0, -1.0))
			// Half-height 1.0 minus the 0.5 the center sits in front of the edge.
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("yields only the deepest contact when several edges overlap", func() {
			// Two parallel edges both facing up (+Y), with the second closer to the
			// rectangle so it is penetrated more deeply.
			mesh := shape2d.NewMesh([]shape2d.Edge{
				shape2d.NewEdge(dprec.NewVec2(2.0, 0.0), dprec.NewVec2(-2.0, 0.0)),
				shape2d.NewEdge(dprec.NewVec2(2.0, 0.2), dprec.NewVec2(-2.0, 0.2)),
			})
			contact, ok := resolve(newRectangle(0.0, 0.5, 0.5, 0.5), mesh)
			Expect(ok).To(BeTrue())
			// The closer edge at y=0.2 gives the deeper contact.
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 0.2))
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(0.0, 1.0))
			// Half-height 0.5 minus the 0.3 the center sits in front of the edge.
			Expect(contact.Depth).To(BeNumerically("~", 0.2, 1e-6))
		})

		It("does not yield a contact for a rectangle fully inside the mesh", func() {
			_, ok := resolve(newRectangle(2.0, 2.0, 0.5, 0.5), square)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact for a rectangle far outside the mesh", func() {
			_, ok := resolve(newRectangle(10.0, 10.0, 1.0, 1.0), square)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact against an empty mesh", func() {
			_, ok := resolve(newRectangle(2.0, -0.5, 1.0, 1.0), shape2d.NewMesh(nil))
			Expect(ok).To(BeFalse())
		})
	})
})
