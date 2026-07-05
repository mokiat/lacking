package shape3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("Triangle", func() {
	var triangle shape3d.Triangle

	BeforeEach(func() {
		// A triangle lying in the XY plane, wound counter-clockwise so that its
		// normal points along the positive Z axis.
		triangle = shape3d.NewTriangle(
			dprec.NewVec3(0.0, 0.0, 0.0),
			dprec.NewVec3(4.0, 0.0, 0.0),
			dprec.NewVec3(0.0, 3.0, 0.0),
		)
	})

	Describe("TransformedTriangle", func() {
		It("applies the transform to all three vertices", func() {
			transform := shape3d.TRTransform(
				dprec.NewVec3(10.0, 20.0, 30.0),
				shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3())),
			)
			result := shape3d.TransformedTriangle(triangle, transform)
			// A (0,0,0) rotated by 90deg around Z stays (0,0,0), then translated to (10,20,30).
			Expect(result.A).To(dprectest.HaveVec3Coords(10.0, 20.0, 30.0))
			// B (4,0,0) rotated by 90deg around Z becomes (0,4,0), then translated to (10,24,30).
			Expect(result.B).To(dprectest.HaveVec3Coords(10.0, 24.0, 30.0))
			// C (0,3,0) rotated by 90deg around Z becomes (-3,0,0), then translated to (7,20,30).
			Expect(result.C).To(dprectest.HaveVec3Coords(7.0, 20.0, 30.0))
		})

		It("leaves the triangle unchanged for the identity transform", func() {
			result := shape3d.TransformedTriangle(triangle, shape3d.IdentityTransform())
			Expect(result.A).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
			Expect(result.B).To(dprectest.HaveVec3Coords(4.0, 0.0, 0.0))
			Expect(result.C).To(dprectest.HaveVec3Coords(0.0, 3.0, 0.0))
		})

		It("does not modify the original triangle", func() {
			_ = shape3d.TransformedTriangle(triangle, shape3d.TranslationTransform(dprec.NewVec3(5.0, 5.0, 5.0)))
			Expect(triangle.A).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
			Expect(triangle.B).To(dprectest.HaveVec3Coords(4.0, 0.0, 0.0))
			Expect(triangle.C).To(dprectest.HaveVec3Coords(0.0, 3.0, 0.0))
		})
	})

	Describe("Centroid", func() {
		It("returns the average of the three vertices", func() {
			Expect(triangle.Centroid()).To(dprectest.HaveVec3Coords(4.0/3.0, 1.0, 0.0))
		})
	})

	Describe("edge lengths", func() {
		It("returns the length of each edge", func() {
			// A 3-4-5 triangle: AB along X is 4, the hypotenuse BC is 5, and CA
			// along Y is 3.
			Expect(triangle.LengthAB()).To(BeNumerically("~", 4.0, 1e-6))
			Expect(triangle.LengthBC()).To(BeNumerically("~", 5.0, 1e-6))
			Expect(triangle.LengthCA()).To(BeNumerically("~", 3.0, 1e-6))
		})
	})

	Describe("Normal", func() {
		It("returns a unit vector facing the side with counter-clockwise winding", func() {
			Expect(triangle.Normal()).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
		})

		It("flips direction when the winding is reversed", func() {
			reversed := shape3d.NewTriangle(triangle.A, triangle.C, triangle.B)
			Expect(reversed.Normal()).To(dprectest.HaveVec3Coords(0.0, 0.0, -1.0))
		})
	})

	Describe("Area", func() {
		It("returns half the magnitude of the edge cross product", func() {
			Expect(triangle.Area()).To(BeNumerically("~", 6.0, 1e-6))
		})
	})

	Describe("FacesTowards", func() {
		It("returns true for a direction aligned with the normal", func() {
			Expect(triangle.FacesTowards(dprec.NewVec3(0.0, 0.0, 1.0))).To(BeTrue())
		})

		It("returns true for a direction in the same hemisphere as the normal", func() {
			Expect(triangle.FacesTowards(dprec.NewVec3(1.0, 1.0, 0.5))).To(BeTrue())
		})

		It("returns false for a direction opposite to the normal", func() {
			Expect(triangle.FacesTowards(dprec.NewVec3(0.0, 0.0, -1.0))).To(BeFalse())
		})

		It("returns false for a direction perpendicular to the normal", func() {
			Expect(triangle.FacesTowards(dprec.NewVec3(1.0, 0.0, 0.0))).To(BeFalse())
		})
	})

	Describe("BoundingSphere", func() {
		It("is centered at the centroid", func() {
			Expect(triangle.BoundingSphere().Center).To(dprectest.HaveVec3Coords(4.0/3.0, 1.0, 0.0))
		})

		It("has a radius equal to the distance to the farthest vertex", func() {
			// Vertex B is the farthest from the centroid.
			expected := dprec.Vec3Diff(triangle.B, triangle.Centroid()).Length()
			Expect(triangle.BoundingSphere().Radius).To(BeNumerically("~", expected, 1e-6))
		})

		It("contains all three vertices", func() {
			bs := triangle.BoundingSphere()
			Expect(bs.ContainsPoint(triangle.A)).To(BeTrue())
			Expect(bs.ContainsPoint(triangle.B)).To(BeTrue())
			Expect(bs.ContainsPoint(triangle.C)).To(BeTrue())
		})
	})
})
