package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk2d/internal"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("GJK", func() {
	var solver *internal.GJKSolver

	runSolver := func(source, target testShape) {
		minkowskiShape := internal.MinkowskiShape{
			Source:     source.Polygon,
			Target:     target.Polygon,
			Offset:     dprec.Vec2Diff(target.Position, source.Position),
			SkinRadius: source.SkinRadius + target.SkinRadius,
		}
		solver.Reset(&minkowskiShape)
		for solver.Next(&minkowskiShape) {
		}
	}

	BeforeEach(func() {
		solver = internal.NewGJKSolver()
	})

	Context("Circle vs Circle", func() {
		Specify("coincident", func() {
			source := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(1.0, 1.0),
				Radius: 2.0,
			})
			target := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(1.0, 1.0),
				Radius: 2.0,
			})
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.PointSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(0.0, 0.0),
					Refs: internal.RefPair{
						SourceIndex: 0,
						TargetIndex: 0,
					},
				},
			)))
		})

		Specify("overlapping", func() {
			source := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(1.0, 1.0),
				Radius: 2.0,
			})
			target := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(1.0, 2.0),
				Radius: 2.0,
			})
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.PointSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(0.0, 1.0),
					Refs: internal.RefPair{
						SourceIndex: 0,
						TargetIndex: 0,
					},
				},
			)))
		})

		Specify("non-overlapping", func() {
			source := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(1.0, 1.0),
				Radius: 2.0,
			})
			target := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(1.0, 6.0),
				Radius: 2.0,
			})
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeFalse())
		})
	})

	Context("Circle vs Capsule", func() {
		Specify("overlapping at Capsule-A", func() {
			source := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(1.0, 1.0),
				Radius: 2.0,
			})
			target := fromCapsule(shape2d.Capsule{
				A:      dprec.NewVec2(1.0, 2.0),
				B:      dprec.NewVec2(1.0, 8.0),
				Radius: 2.0,
			})
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.PointSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(0.0, 1.0),
					Refs: internal.RefPair{
						SourceIndex: 0,
						TargetIndex: 0, // A
					},
				},
			)))
		})

		Specify("overlapping at Capsule-B", func() {
			source := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(1.0, 9.0),
				Radius: 2.0,
			})
			target := fromCapsule(shape2d.Capsule{
				A:      dprec.NewVec2(1.0, 2.0),
				B:      dprec.NewVec2(1.0, 8.0),
				Radius: 2.0,
			})
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.PointSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(0.0, -1.0),
					Refs: internal.RefPair{
						SourceIndex: 0,
						TargetIndex: 1, // B
					},
				},
			)))
		})

		Specify("overlapping at Capsule-Side", func() {
			source := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(2.0, 5.0),
				Radius: 2.0,
			})
			target := fromCapsule(shape2d.Capsule{
				A:      dprec.NewVec2(1.0, 2.0),
				B:      dprec.NewVec2(1.0, 8.0),
				Radius: 2.0,
			})
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.EdgeSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-1.0, 3.0),
					Refs: internal.RefPair{
						SourceIndex: 0,
						TargetIndex: 1, // B
					},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-1.0, -3.0),
					Refs: internal.RefPair{
						SourceIndex: 0,
						TargetIndex: 0, // A
					},
				},
			)))
		})

		Specify("overlapping at Capsule-Core", func() {
			source := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(1.0, 5.0),
				Radius: 2.0,
			})
			target := fromCapsule(shape2d.Capsule{
				A:      dprec.NewVec2(1.0, 2.0),
				B:      dprec.NewVec2(1.0, 8.0),
				Radius: 2.0,
			})
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.EdgeSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(0.0, -3.0),
					Refs: internal.RefPair{
						SourceIndex: 0,
						TargetIndex: 0, // A
					},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(0.0, 3.0),
					Refs: internal.RefPair{
						SourceIndex: 0,
						TargetIndex: 1, // B
					},
				},
			)))
		})
	})

})
