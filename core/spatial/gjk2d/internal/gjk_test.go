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

	Context("Circle vs Rectangle", func() {
		rectangle := shape2d.Rectangle{
			Center:     dprec.NewVec2(0.0, 0.0),
			Rotation:   shape2d.IdentityRotation(),
			HalfWidth:  1.0,
			HalfHeight: 1.0,
		}

		Specify("circle center inside rectangle", func() {
			source := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(0.3, 0.2),
				Radius: 0.1,
			})
			runSolver(source, fromRectangle(rectangle))

			// The origin lies strictly inside the Minkowski difference (a
			// rectangle spanning x in [-1.3,0.7], y in [-1.2,0.8]), so the
			// solver builds an enclosing triangle from three of its corners.
			Expect(solver.ContainsOrigin()).To(BeTrue())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.TriangleSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-1.3, 0.8),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 3},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(0.7, -1.2),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 1},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(0.7, 0.8),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 2},
				},
			)))
		})

		Specify("circle poking through a rectangle face", func() {
			source := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(1.5, 0.0),
				Radius: 1.0,
			})
			runSolver(source, fromRectangle(rectangle))

			// The core gap to the right face is 0.5, less than the 1.0 radius.
			// The closest feature is the right face of the difference, at x=-0.5.
			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.EdgeSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-0.5, 1.0),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 2},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-0.5, -1.0),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 1},
				},
			)))
		})

		Specify("circle touching a rectangle face exactly at skin radius", func() {
			source := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(2.0, 0.0),
				Radius: 1.0,
			})
			runSolver(source, fromRectangle(rectangle))

			// The core gap equals the radius, so the shapes touch exactly and
			// the closest edge sits at distance 1.0 from the origin.
			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.EdgeSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-1.0, 1.0),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 2},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-1.0, -1.0),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 1},
				},
			)))
		})

		Specify("circle reaching a rectangle corner within skin radius", func() {
			source := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(1.5, 1.5),
				Radius: 1.0,
			})
			runSolver(source, fromRectangle(rectangle))

			// The nearest feature is the (1,1) corner, which maps to the
			// Minkowski vertex (-0.5,-0.5) at distance sqrt(0.5) ~ 0.707, within
			// the 1.0 radius. This exercises the edge-to-point downgrade of the
			// closest-feature search: the result is a point simplex.
			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.PointSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-0.5, -0.5),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 2},
				},
			)))
		})

		Specify("circle clearly outside the rectangle", func() {
			source := fromCircle(shape2d.Circle{
				Center: dprec.NewVec2(4.0, 0.0),
				Radius: 1.0,
			})
			runSolver(source, fromRectangle(rectangle))

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeFalse())
		})
	})

	Context("Rectangle vs Rectangle", func() {
		unit := func(center dprec.Vec2) testShape {
			return fromRectangle(shape2d.Rectangle{
				Center:     center,
				Rotation:   shape2d.IdentityRotation(),
				HalfWidth:  1.0,
				HalfHeight: 1.0,
			})
		}

		Specify("deep overlap encloses the origin", func() {
			source := unit(dprec.NewVec2(0.0, 0.0))
			target := unit(dprec.NewVec2(0.5, 0.3))
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeTrue())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.TriangleSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-1.5, -1.7),
					Refs:     internal.RefPair{SourceIndex: 2, TargetIndex: 0},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(2.5, 0.3),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 1},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-1.5, 2.3),
					Refs:     internal.RefPair{SourceIndex: 1, TargetIndex: 3},
				},
			)))
		})

		Specify("coincident rectangles enclose the origin", func() {
			source := unit(dprec.NewVec2(1.0, 2.0))
			target := unit(dprec.NewVec2(1.0, 2.0))
			runSolver(source, target)

			// The Minkowski difference is a rectangle centered on the origin,
			// spanning x, y in [-2,2]. The origin sits on the enclosing
			// triangle's first edge.
			Expect(solver.ContainsOrigin()).To(BeTrue())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.TriangleSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(2.0, 0.0),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 1},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-2.0, 0.0),
					Refs:     internal.RefPair{SourceIndex: 1, TargetIndex: 0},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-2.0, -2.0),
					Refs:     internal.RefPair{SourceIndex: 2, TargetIndex: 0},
				},
			)))
		})

		Specify("faces touching exactly is treated as containment", func() {
			source := unit(dprec.NewVec2(0.0, 0.0))
			target := unit(dprec.NewVec2(2.0, 0.0))
			runSolver(source, target)

			// The right face of the source coincides with the left face of the
			// target, so the origin lies on the boundary of the Minkowski
			// difference and is a vertex of the enclosing triangle.
			Expect(solver.ContainsOrigin()).To(BeTrue())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.TriangleSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(0.0, 0.0),
					Refs:     internal.RefPair{SourceIndex: 1, TargetIndex: 0},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(2.0, 0.0),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 0},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(4.0, 2.0),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 2},
				},
			)))
		})

		Specify("corners touching exactly is treated as containment", func() {
			source := unit(dprec.NewVec2(0.0, 0.0))
			target := unit(dprec.NewVec2(2.0, 2.0))
			runSolver(source, target)

			// Only the (1,1) and (-1,-1) corners meet, so the origin is again a
			// vertex of the Minkowski difference and of the enclosing triangle.
			Expect(solver.ContainsOrigin()).To(BeTrue())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.TriangleSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(0.0, 0.0),
					Refs:     internal.RefPair{SourceIndex: 2, TargetIndex: 0},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(2.0, 2.0),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 0},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(0.0, 4.0),
					Refs:     internal.RefPair{SourceIndex: 1, TargetIndex: 3},
				},
			)))
		})

		Specify("just separated does not overlap", func() {
			source := unit(dprec.NewVec2(0.0, 0.0))
			target := unit(dprec.NewVec2(2.0001, 0.0))
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeFalse())
		})
	})

	Context("Capsule vs Capsule", func() {
		Specify("perpendicular capsules crossing enclose the origin", func() {
			source := fromCapsule(shape2d.Capsule{
				A:      dprec.NewVec2(-1.0, 0.0),
				B:      dprec.NewVec2(1.0, 0.0),
				Radius: 0.3,
			})
			target := fromCapsule(shape2d.Capsule{
				A:      dprec.NewVec2(0.0, -1.0),
				B:      dprec.NewVec2(0.0, 1.0),
				Radius: 0.3,
			})
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeTrue())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.TriangleSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(1.0, -1.0),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 0},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-1.0, 1.0),
					Refs:     internal.RefPair{SourceIndex: 1, TargetIndex: 1},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-1.0, -1.0),
					Refs:     internal.RefPair{SourceIndex: 1, TargetIndex: 0},
				},
			)))
		})

		// Two parallel, equal-length spines make the Minkowski difference
		// collapse to a flat horizontal segment (x in [-2,2], y = 0.8). The
		// origin's closest point is that segment's midpoint (0, 0.8), which is
		// itself a support vertex of the difference, so the solver correctly
		// settles on a point simplex rather than an edge, while still
		// reporting the exact 0.8 distance and the skin overlap.
		Specify("parallel capsules overlapping only through their skins", func() {
			source := fromCapsule(shape2d.Capsule{
				A:      dprec.NewVec2(-1.0, 0.0),
				B:      dprec.NewVec2(1.0, 0.0),
				Radius: 0.5,
			})
			target := fromCapsule(shape2d.Capsule{
				A:      dprec.NewVec2(-1.0, 0.8),
				B:      dprec.NewVec2(1.0, 0.8),
				Radius: 0.5,
			})
			runSolver(source, target)

			// The spines are 0.8 apart, less than the combined radius of 1.0.
			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.PointSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(0.0, 0.8),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 0},
				},
			)))
		})

		// Perpendicular spines make the Minkowski difference a genuine 2D
		// rectangle, so the closest feature is a real edge rather than a
		// vertex. The horizontal source and vertical target are held 0.8 apart,
		// within their combined radius of 1.0.
		Specify("perpendicular capsules overlap through an edge feature", func() {
			source := fromCapsule(shape2d.Capsule{
				A:      dprec.NewVec2(-1.0, 0.0),
				B:      dprec.NewVec2(1.0, 0.0),
				Radius: 0.5,
			})
			target := fromCapsule(shape2d.Capsule{
				A:      dprec.NewVec2(0.0, 0.8),
				B:      dprec.NewVec2(0.0, 2.8),
				Radius: 0.5,
			})
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.EdgeSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(1.0, 0.8),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 0},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-1.0, 0.8),
					Refs:     internal.RefPair{SourceIndex: 1, TargetIndex: 0},
				},
			)))
		})

		Specify("parallel capsules just out of skin reach", func() {
			source := fromCapsule(shape2d.Capsule{
				A:      dprec.NewVec2(-1.0, 0.0),
				B:      dprec.NewVec2(1.0, 0.0),
				Radius: 0.5,
			})
			target := fromCapsule(shape2d.Capsule{
				A:      dprec.NewVec2(-1.0, 1.2),
				B:      dprec.NewVec2(1.0, 1.2),
				Radius: 0.5,
			})
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeFalse())
		})

		// A capsule sitting collinear and overlapping another produces a
		// degenerate, zero-area Minkowski difference: a flat segment through
		// the origin (x in [-0.5,3.5], y = 0). The solver deliberately treats
		// the origin as contained so that the downstream EPA can still recover
		// a separation axis. This pins that behavior; the resulting triangle
		// simplex is degenerate (all three vertices lie on the x axis).
		Specify("collinear overlapping capsules are treated as containment", func() {
			source := fromCapsule(shape2d.Capsule{
				A:      dprec.NewVec2(-1.0, 0.0),
				B:      dprec.NewVec2(1.0, 0.0),
				Radius: 0.2,
			})
			target := fromCapsule(shape2d.Capsule{
				A:      dprec.NewVec2(0.5, 0.0),
				B:      dprec.NewVec2(2.5, 0.0),
				Radius: 0.2,
			})
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeTrue())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.TriangleSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(3.5, 0.0),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 1},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-0.5, 0.0),
					Refs:     internal.RefPair{SourceIndex: 1, TargetIndex: 0},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(1.5, 0.0),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 0},
				},
			)))
		})
	})

	Context("Segment vs Segment", func() {
		Specify("crossing segments enclose the origin", func() {
			source := fromPolygon(dprec.ZeroVec2(), shape2d.IdentityRotation(), 0.0,
				dprec.NewVec2(-1.0, 0.0), dprec.NewVec2(1.0, 0.0))
			target := fromPolygon(dprec.ZeroVec2(), shape2d.IdentityRotation(), 0.0,
				dprec.NewVec2(0.0, -1.0), dprec.NewVec2(0.0, 1.0))
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeTrue())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.TriangleSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(1.0, -1.0),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 0},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-1.0, 1.0),
					Refs:     internal.RefPair{SourceIndex: 1, TargetIndex: 1},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-1.0, -1.0),
					Refs:     internal.RefPair{SourceIndex: 1, TargetIndex: 0},
				},
			)))
		})

		Specify("parallel zero-skin segments never overlap", func() {
			source := fromPolygon(dprec.ZeroVec2(), shape2d.IdentityRotation(), 0.0,
				dprec.NewVec2(-1.0, 0.0), dprec.NewVec2(1.0, 0.0))
			target := fromPolygon(dprec.NewVec2(0.0, 0.5), shape2d.IdentityRotation(), 0.0,
				dprec.NewVec2(-1.0, 0.0), dprec.NewVec2(1.0, 0.0))
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeFalse())
		})
	})

	Context("Triangle vs Rectangle", func() {
		triangle := shape2d.NewTriangle(
			dprec.NewVec2(-1.0, -1.0),
			dprec.NewVec2(1.0, -1.0),
			dprec.NewVec2(0.0, 1.0),
		)

		Specify("overlapping shapes enclose the origin", func() {
			source := fromTriangle(triangle)
			target := fromRectangle(shape2d.Rectangle{
				Center:     dprec.NewVec2(0.5, 0.0),
				Rotation:   shape2d.IdentityRotation(),
				HalfWidth:  1.0,
				HalfHeight: 1.0,
			})
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeTrue())
			Expect(solver.OverlapsOrigin()).To(BeTrue())
			Expect(solver.Simplex()).To(Equal(internal.TriangleSimplex(
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(2.5, 0.0),
					Refs:     internal.RefPair{SourceIndex: 0, TargetIndex: 1},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-1.5, 0.0),
					Refs:     internal.RefPair{SourceIndex: 1, TargetIndex: 0},
				},
				internal.MinkowskiVertex{
					Position: dprec.NewVec2(-0.5, -2.0),
					Refs:     internal.RefPair{SourceIndex: 2, TargetIndex: 0},
				},
			)))
		})

		Specify("separated shapes do not overlap", func() {
			source := fromTriangle(triangle)
			target := fromRectangle(shape2d.Rectangle{
				Center:     dprec.NewVec2(5.0, 0.0),
				Rotation:   shape2d.IdentityRotation(),
				HalfWidth:  1.0,
				HalfHeight: 1.0,
			})
			runSolver(source, target)

			Expect(solver.ContainsOrigin()).To(BeFalse())
			Expect(solver.OverlapsOrigin()).To(BeFalse())
		})
	})

})
