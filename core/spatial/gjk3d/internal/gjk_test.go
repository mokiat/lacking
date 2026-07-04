package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk3d/internal"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("GJK", func() {
	var solver *internal.GJKSolver

	runSolver := func(source, target testShape) {
		minkowskiShape := internal.MinkowskiShape{
			Source:     source.Hull,
			Target:     target.Hull,
			Offset:     dprec.Vec3Diff(target.Position, source.Position),
			SkinRadius: source.SkinRadius + target.SkinRadius,
		}
		solver.Reset(&minkowskiShape)
		for solver.Next(&minkowskiShape) {
		}
	}

	BeforeEach(func() {
		solver = internal.NewGJKSolver()
	})

	Context("Sphere vs Sphere", func() {
		Specify("coincident", func() {
			source := fromSphere(shape3d.NewSphere(dprec.NewVec3(1.0, 1.0, 1.0), 2.0))
			target := fromSphere(shape3d.NewSphere(dprec.NewVec3(1.0, 1.0, 1.0), 2.0))
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.PointSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec3(0.0, 0.0, 0.0),
					Refs: internal.RefPair{
						SourceIndex: 0,
						TargetIndex: 0,
					},
				},
			)))
		})

		Specify("overlapping", func() {
			source := fromSphere(shape3d.NewSphere(dprec.NewVec3(0.0, 0.0, 0.0), 2.0))
			target := fromSphere(shape3d.NewSphere(dprec.NewVec3(0.0, 3.0, 0.0), 2.0))
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
		})

		Specify("touching exactly at skin radius", func() {
			source := fromSphere(shape3d.NewSphere(dprec.NewVec3(0.0, 0.0, 0.0), 1.0))
			target := fromSphere(shape3d.NewSphere(dprec.NewVec3(2.0, 0.0, 0.0), 1.0))
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
		})

		Specify("non-overlapping", func() {
			source := fromSphere(shape3d.NewSphere(dprec.NewVec3(0.0, 0.0, 0.0), 2.0))
			target := fromSphere(shape3d.NewSphere(dprec.NewVec3(0.0, 6.0, 0.0), 2.0))
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeFalse())
		})
	})

	Context("Box vs Box", func() {
		Specify("deep overlap encloses the origin", func() {
			source := fromBox(shape3d.NewBox(dprec.ZeroVec3(), shape3d.IdentityRotation(), dprec.NewVec3(1.0, 1.0, 1.0)))
			target := fromBox(shape3d.NewBox(dprec.NewVec3(0.5, 0.3, -0.2), shape3d.IdentityRotation(), dprec.NewVec3(1.0, 1.0, 1.0)))
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeTrue())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex().VertexCount).To(BeNumerically("==", 4))
		})

		Specify("faces touching exactly is treated as containment", func() {
			source := fromBox(shape3d.NewBox(dprec.ZeroVec3(), shape3d.IdentityRotation(), dprec.NewVec3(1.0, 1.0, 1.0)))
			target := fromBox(shape3d.NewBox(dprec.NewVec3(2.0, 0.0, 0.0), shape3d.IdentityRotation(), dprec.NewVec3(1.0, 1.0, 1.0)))
			runSolver(source, target)

			Expect(solver.OverlapsOrigin()).To(BeTrue())
		})

		Specify("separated along a diagonal does not overlap", func() {
			source := fromBox(shape3d.NewBox(dprec.ZeroVec3(), shape3d.IdentityRotation(), dprec.NewVec3(1.0, 1.0, 1.0)))
			target := fromBox(shape3d.NewBox(dprec.NewVec3(2.5, 2.5, 2.5), shape3d.IdentityRotation(), dprec.NewVec3(1.0, 1.0, 1.0)))
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeFalse())
		})
	})

	Context("Capsule vs Capsule", func() {
		Specify("perpendicular skew capsules crossing enclose the origin", func() {
			source := fromCapsule(shape3d.NewSegment(dprec.NewVec3(-1.0, 0.0, 0.0), dprec.NewVec3(1.0, 0.0, 0.0)), 0.3)
			target := fromCapsule(shape3d.NewSegment(dprec.NewVec3(0.0, -1.0, 0.05), dprec.NewVec3(0.0, 1.0, 0.05)), 0.3)
			runSolver(source, target)

			Expect(solver.OverlapsOrigin()).To(BeTrue())
		})

		Specify("parallel capsules just out of skin reach", func() {
			source := fromCapsule(shape3d.NewSegment(dprec.NewVec3(-1.0, 0.0, 0.0), dprec.NewVec3(1.0, 0.0, 0.0)), 0.1)
			target := fromCapsule(shape3d.NewSegment(dprec.NewVec3(-1.0, 3.0, 0.0), dprec.NewVec3(1.0, 3.0, 0.0)), 0.1)
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeFalse())
		})
	})

	Context("degenerate configurations", func() {
		// Two collinear, overlapping segments produce a zero-volume Minkowski
		// difference: a flat segment through the origin. The collinearity guard
		// must detect that the difference is flat and stop before forming a
		// degenerate triangle or tetrahedron, reporting an overlap (touch) but
		// not containment.
		Specify("collinear overlapping segments resolve without containment", func() {
			source := fromCapsule(shape3d.NewSegment(dprec.NewVec3(0.0, 0.0, 0.0), dprec.NewVec3(1.0, 0.0, 0.0)), 0.0)
			target := fromCapsule(shape3d.NewSegment(dprec.NewVec3(0.5, 0.0, 0.0), dprec.NewVec3(2.0, 0.0, 0.0)), 0.0)
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
		})

		// The Minkowski difference of two identical triangles that are offset
		// only along z is a flat, zero-thickness hexagon lying in the z=0.3
		// plane. The origin sits 0.3 below that plane, inside the hexagon
		// outline and within the combined skin radius. The coplanar-tetrahedron
		// guard must prevent a degenerate tetrahedron from being formed while
		// the flat face is discovered, reporting an overlap through the face
		// feature but no containment.
		Specify("flat off-plane difference overlaps through a face feature", func() {
			triangle := shape3d.NewTriangle(
				dprec.NewVec3(-2.0, -2.0, 0.0),
				dprec.NewVec3(2.0, -2.0, 0.0),
				dprec.NewVec3(0.0, 2.0, 0.0),
			)
			source := fromTriangle(triangle)
			target := fromTriangle(triangle)
			target.Position = dprec.NewVec3(0.0, 0.0, 0.3)
			target.SkinRadius = 0.5
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
		})

		// A support vertex that coincides with an existing simplex vertex must
		// not grow the simplex; the solver terminates cleanly.
		Specify("coincident spheres terminate at a point simplex", func() {
			source := fromSphere(shape3d.NewSphere(dprec.ZeroVec3(), 1.0))
			target := fromSphere(shape3d.NewSphere(dprec.ZeroVec3(), 1.0))
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex().VertexCount).To(BeNumerically("==", 1))
		})
	})
})
