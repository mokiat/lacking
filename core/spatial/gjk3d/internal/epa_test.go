package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk3d/internal"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("EPA", func() {
	var (
		gjkSolver *internal.GJKSolver
		epaSolver *internal.EPASolver
	)

	// resolve runs GJK to completion and, when the shapes overlap, runs EPA and
	// returns the resulting solution together with a true flag.
	resolve := func(source, target testShape) (internal.EPASolution, bool) {
		minkowskiShape := internal.MinkowskiShape{
			Source:     source.Hull,
			Target:     target.Hull,
			Offset:     dprec.Vec3Diff(target.Position, source.Position),
			SkinRadius: source.SkinRadius + target.SkinRadius,
		}
		gjkSolver.Reset(&minkowskiShape)
		for gjkSolver.Next(&minkowskiShape) {
		}
		if !gjkSolver.OverlapsOrigin() {
			return internal.EPASolution{}, false
		}
		epaSolver.Reset(&minkowskiShape, gjkSolver.Simplex(), gjkSolver.ContainsOrigin())
		for epaSolver.Next(&minkowskiShape) {
		}
		return epaSolver.Solution(), true
	}

	expectVec3 := func(actual, expected dprec.Vec3) {
		Expect(actual.X).To(BeNumerically("~", expected.X, 1e-6))
		Expect(actual.Y).To(BeNumerically("~", expected.Y, 1e-6))
		Expect(actual.Z).To(BeNumerically("~", expected.Z, 1e-6))
	}

	BeforeEach(func() {
		gjkSolver = internal.NewGJKSolver()
		epaSolver = internal.NewEPASolver()
	})

	// Deep overlap of two boxes: the origin is contained in the Minkowski
	// difference core, so EPA expands the polytope and reports the outward
	// normal of the closest boundary face and the core penetration depth.
	Specify("contained box overlap reports the minimum-translation face", func() {
		source := fromBox(shape3d.NewBox(dprec.ZeroVec3(), shape3d.IdentityRotation(), dprec.NewVec3(1.0, 1.0, 1.0)))
		target := fromBox(shape3d.NewBox(dprec.NewVec3(0.5, 0.0, 0.0), shape3d.IdentityRotation(), dprec.NewVec3(1.0, 1.0, 1.0)))

		solution, ok := resolve(source, target)
		Expect(ok).To(BeTrue())
		expectVec3(solution.Normal, dprec.NewVec3(-1.0, 0.0, 0.0))
		Expect(solution.Depth).To(BeNumerically("~", 1.5, 1e-6))
	})

	// Two spheres overlapping only through their skins: the core difference is a
	// single point, so EPA terminates on a point feature. The normal points from
	// the point toward the origin and the depth is the skin bridge.
	Specify("point feature from overlapping spheres", func() {
		source := fromSphere(shape3d.NewSphere(dprec.ZeroVec3(), 1.0))
		target := fromSphere(shape3d.NewSphere(dprec.NewVec3(1.5, 0.0, 0.0), 1.0))

		solution, ok := resolve(source, target)
		Expect(ok).To(BeTrue())
		expectVec3(solution.Normal, dprec.NewVec3(-1.0, 0.0, 0.0))
		Expect(solution.Depth).To(BeNumerically("~", 0.5, 1e-6))
		Expect(solution.BaryA).To(BeNumerically("~", 1.0, 1e-6))
	})

	// A capsule (segment core) against a sphere (point core) offset sideways:
	// the core difference is a segment and the origin projects onto its middle,
	// so EPA terminates on an edge feature.
	Specify("edge feature from a capsule beside a sphere", func() {
		source := fromCapsule(shape3d.NewSegment(dprec.NewVec3(-1.0, 0.0, 0.0), dprec.NewVec3(1.0, 0.0, 0.0)), 0.3)
		target := fromSphere(shape3d.NewSphere(dprec.NewVec3(0.0, 0.5, 0.0), 0.3))

		solution, ok := resolve(source, target)
		Expect(ok).To(BeTrue())
		expectVec3(solution.Normal, dprec.NewVec3(0.0, -1.0, 0.0))
		Expect(solution.Depth).To(BeNumerically("~", 0.1, 1e-6))
		// The closest point is the middle of the segment feature.
		Expect(solution.BaryA + solution.BaryB).To(BeNumerically("~", 1.0, 1e-6))
		Expect(solution.BaryC).To(BeNumerically("~", 0.0, 1e-6))
	})

	// A box against a sphere pressed against one of its faces: the core
	// difference is a box and the origin projects onto a face, so EPA terminates
	// on a triangular face feature.
	Specify("face feature from a sphere pressed onto a box face", func() {
		source := fromBox(shape3d.NewBox(dprec.ZeroVec3(), shape3d.IdentityRotation(), dprec.NewVec3(1.0, 1.0, 1.0)))
		target := fromSphere(shape3d.NewSphere(dprec.NewVec3(1.4, 0.0, 0.0), 0.5))

		solution, ok := resolve(source, target)
		Expect(ok).To(BeTrue())
		expectVec3(solution.Normal, dprec.NewVec3(-1.0, 0.0, 0.0))
		Expect(solution.Depth).To(BeNumerically("~", 0.1, 1e-6))
	})
})
