package isec2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("SegmentMesh", func() {
	// A counter-clockwise-wound square spanning (0,0) to (4,4). Because the
	// winding is counter-clockwise, each edge's right-hand normal faces outward,
	// away from the center at (2,2). The mesh therefore behaves as a one-sided
	// boundary that is only crossed from the outside.
	var square shape2d.Mesh

	newSegment := func(ax, ay, bx, by float64) shape2d.Segment {
		return shape2d.NewSegment(dprec.NewVec2(ax, ay), dprec.NewVec2(bx, by))
	}

	BeforeEach(func() {
		square = shape2d.NewMesh([]shape2d.Edge{
			shape2d.NewEdge(dprec.NewVec2(0.0, 0.0), dprec.NewVec2(4.0, 0.0)), // bottom, normal -Y
			shape2d.NewEdge(dprec.NewVec2(4.0, 0.0), dprec.NewVec2(4.0, 4.0)), // right, normal +X
			shape2d.NewEdge(dprec.NewVec2(4.0, 4.0), dprec.NewVec2(0.0, 4.0)), // top, normal +Y
			shape2d.NewEdge(dprec.NewVec2(0.0, 4.0), dprec.NewVec2(0.0, 0.0)), // left, normal -X
		})
	})

	Describe("CheckSegmentMesh", func() {
		It("returns true for a segment passing through the mesh", func() {
			Expect(isec2d.CheckSegmentMesh(newSegment(-2.0, 2.0, 6.0, 2.0), square)).To(BeTrue())
		})

		It("returns false for a segment lying entirely inside the mesh", func() {
			// Never reaches an edge from its front side.
			Expect(isec2d.CheckSegmentMesh(newSegment(1.0, 2.0, 3.0, 2.0), square)).To(BeFalse())
		})

		It("returns false for a segment that misses the mesh", func() {
			Expect(isec2d.CheckSegmentMesh(newSegment(-2.0, 6.0, 6.0, 6.0), square)).To(BeFalse())
		})

		It("returns false against an empty mesh", func() {
			Expect(isec2d.CheckSegmentMesh(newSegment(-2.0, 2.0, 6.0, 2.0), shape2d.NewMesh(nil))).To(BeFalse())
		})
	})

	Describe("ResolveSegmentMesh", func() {
		resolve := func(segment shape2d.Segment, mesh shape2d.Mesh) (shape2d.Contact, bool) {
			var sink shape2d.LastContact
			isec2d.ResolveSegmentMesh(segment, mesh, sink.AddContact)
			return sink.Contact()
		}

		It("yields a contact at the entry edge of the mesh", func() {
			contact, ok := resolve(newSegment(-2.0, 2.0, 6.0, 2.0), square)
			Expect(ok).To(BeTrue())
			// Enters through the left edge; the exit through the right edge is
			// back-face culled.
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(0.0, 2.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(-1.0, 0.0))
			// Crossing is a quarter of the way along the segment, so 0.75 remains.
			Expect(contact.Depth).To(BeNumerically("~", 0.75, 1e-6))
		})

		It("yields the earliest crossing when several edges are crossed", func() {
			// Two parallel edges both facing -X (front side to their left), with the
			// first at x=1 and the second at x=3.
			mesh := shape2d.NewMesh([]shape2d.Edge{
				shape2d.NewEdge(dprec.NewVec2(1.0, 4.0), dprec.NewVec2(1.0, 0.0)),
				shape2d.NewEdge(dprec.NewVec2(3.0, 4.0), dprec.NewVec2(3.0, 0.0)),
			})
			contact, ok := resolve(newSegment(-1.0, 2.0, 5.0, 2.0), mesh)
			Expect(ok).To(BeTrue())
			// The nearer edge at x=1 is crossed first.
			Expect(contact.TargetPoint).To(dprectest.HaveVec2Coords(1.0, 2.0))
			Expect(contact.TargetNormal).To(dprectest.HaveVec2Coords(-1.0, 0.0))
			Expect(contact.Depth).To(BeNumerically("~", 2.0/3.0, 1e-6))
		})

		It("does not yield a contact for a segment inside the mesh", func() {
			_, ok := resolve(newSegment(1.0, 2.0, 3.0, 2.0), square)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact for a segment that misses the mesh", func() {
			_, ok := resolve(newSegment(-2.0, 6.0, 6.0, 6.0), square)
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact against an empty mesh", func() {
			_, ok := resolve(newSegment(-2.0, 2.0, 6.0, 2.0), shape2d.NewMesh(nil))
			Expect(ok).To(BeFalse())
		})
	})
})
