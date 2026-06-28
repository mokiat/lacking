package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Edge", func() {
	var edge shape2d.Edge

	BeforeEach(func() {
		// A 3-4-5 edge, so the length and normal come out to clean values.
		edge = shape2d.NewEdge(
			dprec.NewVec2(0.0, 0.0),
			dprec.NewVec2(3.0, 4.0),
		)
	})

	Describe("NewEdge", func() {
		It("sets the start and end points", func() {
			Expect(edge.A).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(edge.B).To(dprectest.HaveVec2Coords(3.0, 4.0))
		})
	})

	Describe("TransformedEdge", func() {
		It("applies the transform to both endpoints", func() {
			transform := shape2d.TRTransform(
				dprec.NewVec2(10.0, 20.0),
				shape2d.RotationFromAngle(dprec.Degrees(90.0)),
			)
			result := shape2d.TransformedEdge(edge, transform)
			// A (0,0) rotated by 90deg stays (0,0), then translated to (10,20).
			Expect(result.A).To(dprectest.HaveVec2Coords(10.0, 20.0))
			// B (3,4) rotated by 90deg becomes (-4,3), then translated to (6,23).
			Expect(result.B).To(dprectest.HaveVec2Coords(6.0, 23.0))
		})

		It("leaves the edge unchanged for the identity transform", func() {
			result := shape2d.TransformedEdge(edge, shape2d.IdentityTransform())
			Expect(result.A).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(result.B).To(dprectest.HaveVec2Coords(3.0, 4.0))
		})

		It("does not modify the original edge", func() {
			shape2d.TransformedEdge(edge, shape2d.TranslationTransform(dprec.NewVec2(5.0, 5.0)))
			Expect(edge.A).To(dprectest.HaveVec2Coords(0.0, 0.0))
			Expect(edge.B).To(dprectest.HaveVec2Coords(3.0, 4.0))
		})
	})

	Describe("Midpoint", func() {
		It("returns the point halfway between the endpoints", func() {
			Expect(edge.Midpoint()).To(dprectest.HaveVec2Coords(1.5, 2.0))
		})
	})

	Describe("Length", func() {
		It("returns the Euclidean distance between A and B", func() {
			Expect(edge.Length()).To(BeNumerically("~", 5.0, 1e-6))
		})

		It("returns zero for a zero-length edge", func() {
			dot := shape2d.NewEdge(
				dprec.NewVec2(1.0, 2.0),
				dprec.NewVec2(1.0, 2.0),
			)
			Expect(dot.Length()).To(BeNumerically("~", 0.0, 1e-6))
		})
	})

	Describe("Normal", func() {
		It("returns a unit vector to the right of the A-to-B direction", func() {
			// Direction (3,4) rotated 90deg clockwise is (4,-3); normalized that
			// is (0.8,-0.6).
			Expect(edge.Normal()).To(dprectest.HaveVec2Coords(0.8, -0.6))
		})

		It("is perpendicular to the edge direction", func() {
			direction := dprec.Vec2Diff(edge.B, edge.A)
			Expect(dprec.Vec2Dot(edge.Normal(), direction)).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("is a unit vector", func() {
			Expect(edge.Normal().Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("points outward for an edge of a counter-clockwise polygon", func() {
			// The bottom edge of a CCW square runs from (0,0) to (1,0); its
			// outward normal points down, away from the interior above it.
			bottom := shape2d.NewEdge(
				dprec.NewVec2(0.0, 0.0),
				dprec.NewVec2(1.0, 0.0),
			)
			Expect(bottom.Normal()).To(dprectest.HaveVec2Coords(0.0, -1.0))
		})
	})

	Describe("BoundingCircle", func() {
		It("is centered at the midpoint of the edge", func() {
			Expect(edge.BoundingCircle().Center).To(dprectest.HaveVec2Coords(1.5, 2.0))
		})

		It("has radius equal to half the edge length", func() {
			Expect(edge.BoundingCircle().Radius).To(BeNumerically("~", 2.5, 1e-6))
		})

		It("contains both endpoints", func() {
			bc := edge.BoundingCircle()
			Expect(bc.ContainsPoint(edge.A)).To(BeTrue())
			Expect(bc.ContainsPoint(edge.B)).To(BeTrue())
		})
	})
})
