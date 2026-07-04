package gjk2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Solver", func() {
	var solver *gjk2d.Solver

	BeforeEach(func() {
		solver = gjk2d.NewSolver()
	})

	Describe("Intersect", func() {
		Describe("circle vs circle", func() {
			It("returns false when circles are clearly separated", func() {
				a := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.0, 0.0),
					Radius: 1.0,
				})
				b := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(3.0, 0.0),
					Radius: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("returns true when circles clearly overlap", func() {
				a := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.0, 0.0),
					Radius: 1.0,
				})
				b := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(1.5, 0.0),
					Radius: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns true when one circle contains the other", func() {
				a := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.0, 0.0),
					Radius: 3.0,
				})
				b := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.5, 0.0),
					Radius: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns true when circles touch exactly at skin radius boundary", func() {
				a := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.0, 0.0),
					Radius: 1.0,
				})
				b := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(2.0, 0.0),
					Radius: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})
		})

		Describe("rectangle vs rectangle", func() {
			It("returns false when rectangles are clearly separated", func() {
				a := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(0.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				})
				b := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(4.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("returns true when rectangles clearly overlap", func() {
				a := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(0.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				})
				b := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(1.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns true when a corner of one rectangle touches the edge of another", func() {
				a := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(0.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				})
				// Place b so its left edge sits exactly at x=1 (right edge of a).
				b := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(2.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns false when rotated rectangles pass by each other diagonally", func() {
				rot45 := shape2d.RotationFromAngle(dprec.Degrees(45.0))
				a := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(0.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  0.5,
					HalfHeight: 0.5,
				})
				b := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(2.0, 0.0),
					Rotation:   rot45,
					HalfWidth:  0.5,
					HalfHeight: 0.5,
				})
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})
		})

		Describe("circle vs rectangle", func() {
			It("returns false when the circle is clearly outside the rectangle", func() {
				a := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.0, 0.0),
					Radius: 1.0,
				})
				b := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(4.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("returns true when the circle overlaps the rectangle", func() {
				a := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.0, 0.0),
					Radius: 1.5,
				})
				b := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(2.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns true when the circle is fully inside the rectangle", func() {
				a := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.0, 0.0),
					Radius: 0.3,
				})
				b := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(0.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})
		})

		Describe("capsule vs capsule", func() {
			It("returns false when capsules are clearly separated", func() {
				a := gjk2d.ShapeFromCapsule(shape2d.Capsule{
					A:      dprec.NewVec2(-1.0, 0.0),
					B:      dprec.NewVec2(1.0, 0.0),
					Radius: 0.5,
				})
				b := gjk2d.ShapeFromCapsule(shape2d.Capsule{
					A:      dprec.NewVec2(-1.0, 3.0),
					B:      dprec.NewVec2(1.0, 3.0),
					Radius: 0.5,
				})
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("returns true when capsule end-caps overlap", func() {
				a := gjk2d.ShapeFromCapsule(shape2d.Capsule{
					A:      dprec.NewVec2(-1.0, 0.0),
					B:      dprec.NewVec2(0.0, 0.0),
					Radius: 0.5,
				})
				b := gjk2d.ShapeFromCapsule(shape2d.Capsule{
					A:      dprec.NewVec2(0.5, 0.0),
					B:      dprec.NewVec2(1.5, 0.0),
					Radius: 0.5,
				})
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns true when perpendicular capsules cross", func() {
				a := gjk2d.ShapeFromCapsule(shape2d.Capsule{
					A:      dprec.NewVec2(-1.0, 0.0),
					B:      dprec.NewVec2(1.0, 0.0),
					Radius: 0.1,
				})
				b := gjk2d.ShapeFromCapsule(shape2d.Capsule{
					A:      dprec.NewVec2(0.0, -1.0),
					B:      dprec.NewVec2(0.0, 1.0),
					Radius: 0.1,
				})
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})
		})

		Describe("edge cases", func() {
			It("returns false when shape A has no points", func() {
				a := gjk2d.Shape{}
				b := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.0, 0.0),
					Radius: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("returns false when shape B has no points", func() {
				a := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.0, 0.0),
					Radius: 1.0,
				})
				b := gjk2d.Shape{}
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})
		})
	})

	Describe("Resolve", func() {

		// moveBy returns a copy of the shape translated by the given offset.
		moveBy := func(shape gjk2d.Shape, offset dprec.Vec2) gjk2d.Shape {
			shape.Position = dprec.Vec2Sum(shape.Position, offset)
			return shape
		}

		// expectVec2 asserts that two vectors are approximately equal.
		expectVec2 := func(actual, expected dprec.Vec2) {
			Expect(actual.X).To(BeNumerically("~", expected.X, 1e-6))
			Expect(actual.Y).To(BeNumerically("~", expected.Y, 1e-6))
		}

		Describe("circle vs circle", func() {
			It("resolves overlapping circles along the line of centers", func() {
				a := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.0, 0.0),
					Radius: 1.0,
				})
				b := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(1.5, 0.0),
					Radius: 1.0,
				})

				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
				expectVec2(contact.TargetNormal, dprec.NewVec2(-1.0, 0.0))
				expectVec2(contact.TargetPoint, dprec.NewVec2(0.5, 0.0))
				expectVec2(contact.EvalSourcePoint(), dprec.NewVec2(1.0, 0.0))
			})

			It("reports a barely-touching contact with zero depth", func() {
				a := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.0, 0.0),
					Radius: 1.0,
				})
				b := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(2.0, 0.0),
					Radius: 1.0,
				})

				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
				expectVec2(contact.TargetNormal, dprec.NewVec2(-1.0, 0.0))
			})

			It("returns false when the circles are separated", func() {
				a := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.0, 0.0),
					Radius: 1.0,
				})
				b := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(3.0, 0.0),
					Radius: 1.0,
				})

				_, ok := solver.Resolve(a, b)
				Expect(ok).To(BeFalse())
			})
		})

		Describe("rectangle vs rectangle", func() {
			It("resolves along the axis of minimum penetration", func() {
				a := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(0.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				})
				b := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(1.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				})

				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				Expect(contact.Depth).To(BeNumerically("~", 1.0, 1e-6))
				expectVec2(contact.TargetNormal, dprec.NewVec2(-1.0, 0.0))
			})

			It("returns false when the rectangles are separated", func() {
				a := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(0.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				})
				b := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(4.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				})

				_, ok := solver.Resolve(a, b)
				Expect(ok).To(BeFalse())
			})

			// The two cores are separated (they touch exactly along the shared
			// edge x=1) but are bridged by the skin radius, so the overlap is
			// purely along the +x axis. The rounded square is rotated 90 degrees
			// so that the GJK support search lands a Minkowski vertex exactly on
			// the origin, leaving EPA with a single-point simplex at the origin.
			// The correct separation is the outward normal of the boundary
			// feature the origin lies on, i.e. (-1, 0) with a depth equal to the
			// combined skin radius.
			It("resolves rounded and plain squares that touch along a shared edge", func() {
				a := gjk2d.ShapeFromRectangleRound(shape2d.Rectangle{
					Center:     dprec.NewVec2(0.0, 0.0),
					Rotation:   shape2d.RotationFromAngle(dprec.Degrees(90)),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				}, 1.45)
				b := gjk2d.ShapeFromRectangle(shape2d.Rectangle{
					Center:     dprec.NewVec2(2.0, 0.0),
					Rotation:   shape2d.IdentityRotation(),
					HalfWidth:  1.0,
					HalfHeight: 1.0,
				})

				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				Expect(contact.Depth).To(BeNumerically("~", 1.45, 1e-6))
				expectVec2(contact.TargetNormal, dprec.NewVec2(-1.0, 0.0))

				// The reported contact must be geometrically consistent: moving
				// the source shape along the normal by the depth must separate.
				offset := dprec.Vec2Prod(contact.TargetNormal, contact.Depth+1e-4)
				Expect(solver.Intersect(moveBy(a, offset), b)).To(BeFalse())
			})
		})

		Describe("capsule vs capsule", func() {
			// Deep overlap of round shapes where the origin is contained in
			// the Minkowski difference core. Guards the containment regime of
			// the EPA solver: the normal must be the outward normal of the
			// closest boundary feature (not its inverse) and the core
			// penetration must add to the combined skin radius (not subtract
			// from it).
			It("resolves deeply crossing capsules", func() {
				a := gjk2d.ShapeFromCapsule(shape2d.Capsule{
					A:      dprec.NewVec2(-1.0, 0.0),
					B:      dprec.NewVec2(1.0, 0.0),
					Radius: 0.3,
				})
				b := gjk2d.ShapeFromCapsule(shape2d.Capsule{
					A:      dprec.NewVec2(0.5, -2.0),
					B:      dprec.NewVec2(0.5, 2.0),
					Radius: 0.3,
				})

				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				Expect(contact.Depth).To(BeNumerically("~", 1.1, 1e-6))
				expectVec2(contact.TargetNormal, dprec.NewVec2(-1.0, 0.0))
				expectVec2(contact.TargetPoint, dprec.NewVec2(0.2, 0.0))
				expectVec2(contact.EvalSourcePoint(), dprec.NewVec2(1.3, 0.0))
			})
		})

		Describe("separation", func() {
			// Moving the source shape along the contact normal by slightly more than
			// the reported depth must end the overlap.
			separates := func(a, b gjk2d.Shape) {
				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				offset := dprec.Vec2Prod(contact.TargetNormal, contact.Depth+1e-4)
				Expect(solver.Intersect(moveBy(a, offset), b)).To(BeFalse())
			}

			It("separates overlapping circles", func() {
				separates(
					gjk2d.ShapeFromCircle(shape2d.Circle{Center: dprec.NewVec2(0.0, 0.0), Radius: 1.0}),
					gjk2d.ShapeFromCircle(shape2d.Circle{Center: dprec.NewVec2(1.2, 0.3), Radius: 1.0}),
				)
			})

			It("separates overlapping rectangles", func() {
				separates(
					gjk2d.ShapeFromRectangle(shape2d.Rectangle{Center: dprec.NewVec2(0.0, 0.0), Rotation: shape2d.IdentityRotation(), HalfWidth: 1.0, HalfHeight: 1.0}),
					gjk2d.ShapeFromRectangle(shape2d.Rectangle{Center: dprec.NewVec2(1.3, 0.5), Rotation: shape2d.IdentityRotation(), HalfWidth: 1.0, HalfHeight: 1.0}),
				)
			})

			It("separates crossing capsules", func() {
				separates(
					gjk2d.ShapeFromCapsule(shape2d.Capsule{A: dprec.NewVec2(-1.0, 0.0), B: dprec.NewVec2(1.0, 0.0), Radius: 0.3}),
					gjk2d.ShapeFromCapsule(shape2d.Capsule{A: dprec.NewVec2(0.0, -1.0), B: dprec.NewVec2(0.0, 1.0), Radius: 0.3}),
				)
			})

			It("separates a circle overlapping a rectangle", func() {
				separates(
					gjk2d.ShapeFromCircle(shape2d.Circle{Center: dprec.NewVec2(0.0, 0.0), Radius: 1.5}),
					gjk2d.ShapeFromRectangle(shape2d.Rectangle{Center: dprec.NewVec2(2.0, 0.0), Rotation: shape2d.IdentityRotation(), HalfWidth: 1.0, HalfHeight: 1.0}),
				)
			})
		})

		// Degenerate configurations of valid shapes, where the Minkowski
		// difference core is flat and/or the origin lies exactly on its
		// boundary. Every sign test in GJK and EPA then runs on values that
		// are either exactly zero or rounding noise, so these scenarios
		// require dedicated handling. Where a scenario admits several valid
		// separation normals, the specs assert properties of the contact
		// rather than exact values.
		Describe("degenerate configurations", func() {

			// resolveAndSeparate resolves the two shapes, asserts that they
			// overlap and that moving the source shape along the contact
			// normal by slightly more than the reported depth ends the
			// overlap, and returns the contact for further assertions.
			resolveAndSeparate := func(a, b gjk2d.Shape) shape2d.Contact {
				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-9))
				offset := dprec.Vec2Prod(contact.TargetNormal, contact.Depth+1e-4)
				Expect(solver.Intersect(moveBy(a, offset), b)).To(BeFalse())
				return contact
			}

			// The Minkowski difference is a single point at the origin, so
			// there is no boundary information at all. Guards the final EPA
			// fallback: any unit direction is a valid normal and the depth is
			// the sum of the radii.
			It("resolves exactly coincident circles", func() {
				a := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(1.0, 2.0),
					Radius: 1.0,
				})
				b := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(1.0, 2.0),
					Radius: 0.5,
				})

				contact := resolveAndSeparate(a, b)
				Expect(contact.Depth).To(BeNumerically("~", 1.5, 1e-6))
			})

			// The Minkowski difference is a flat segment through the origin
			// and the origin is one of its support points. Guards the GJK
			// origin-support termination (a support at the origin must be
			// treated as containment) and the EPA handling of flat polytopes:
			// the normal must be perpendicular to the shared axis, since a
			// normal along the axis does not separate the shapes at the
			// reported depth.
			It("resolves exactly coincident capsules", func() {
				capsule := shape2d.Capsule{
					A:      dprec.NewVec2(-1.0, 0.0),
					B:      dprec.NewVec2(1.0, 0.0),
					Radius: 1.45,
				}
				a := gjk2d.ShapeFromCapsule(capsule)
				b := gjk2d.ShapeFromCapsule(capsule)

				contact := resolveAndSeparate(a, b)
				Expect(contact.Depth).To(BeNumerically("~", 2.9, 1e-6))
				Expect(contact.TargetNormal.X).To(BeNumerically("~", 0.0, 1e-9))
			})

			// The circle's center lies exactly on a capsule core endpoint, so
			// the origin is a vertex of the (flat) Minkowski difference.
			// Guards the GJK origin-support termination together with the EPA
			// zero length edge handling: the padded simplex contains
			// coincident vertices whose polytope edges must be discarded.
			It("resolves a circle centered exactly on a capsule endpoint", func() {
				a := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.0, 2.0),
					Radius: 1.0,
				})
				b := gjk2d.ShapeFromCapsule(shape2d.Capsule{
					A:      dprec.NewVec2(0.0, -2.0),
					B:      dprec.NewVec2(0.0, 2.0),
					Radius: 1.45,
				})

				contact := resolveAndSeparate(a, b)
				Expect(contact.Depth).To(BeNumerically("~", 2.45, 1e-6))
			})

			// The same configuration, except that the capsule points are
			// produced through a rotation, so the near-origin support carries
			// rounding noise and the GJK simplex degenerates to three
			// coincident vertices that provide no boundary information.
			// Guards the EPA support-probe polytope seeding: normalizing the
			// noise vector instead would produce an arbitrary normal.
			It("resolves a circle centered on a rotated capsule endpoint", func() {
				a := gjk2d.ShapeFromCircle(shape2d.Circle{
					Center: dprec.NewVec2(0.0, 2.0),
					Radius: 1.0,
				})
				b := gjk2d.Shape{
					Position: dprec.ZeroVec2(),
					Rotation: shape2d.RotationFromAngle(dprec.Degrees(90.0)),
					Points: []dprec.Vec2{
						dprec.NewVec2(-2.0, 0.0),
						dprec.NewVec2(2.0, 0.0),
					},
					SkinRadius: 1.45,
				}

				contact := resolveAndSeparate(a, b)
				Expect(contact.Depth).To(BeNumerically("~", 2.45, 1e-6))
			})

			// A capsule resting exactly collinear on a longer segment: the
			// Minkowski difference is a flat segment through the origin, but
			// no support lands exactly at the origin, and (because of the
			// rotation) the collinearity tests run on rounding noise. Guards
			// the GJK flat-simplex span handling: the origin must be treated
			// as contained and the contact must be perpendicular to the
			// shared axis.
			It("resolves a capsule resting collinear on a segment", func() {
				rotation := shape2d.RotationFromAngle(dprec.Degrees(45.0))
				a := gjk2d.Shape{
					Position: dprec.NewVec2(1.0, 1.0),
					Rotation: rotation,
					Points: []dprec.Vec2{
						dprec.NewVec2(-1.0, 0.0),
						dprec.NewVec2(1.0, 0.0),
					},
					SkinRadius: 0.5,
				}
				b := gjk2d.Shape{
					Position: dprec.NewVec2(1.0, 1.0),
					Rotation: rotation,
					Points: []dprec.Vec2{
						dprec.NewVec2(-1.5, 0.0),
						dprec.NewVec2(1.5, 0.0),
					},
					SkinRadius: 0.0,
				}

				contact := resolveAndSeparate(a, b)
				Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
				axis := rotation.Apply(dprec.BasisXVec2())
				Expect(dprec.Vec2Dot(contact.TargetNormal, axis)).To(BeNumerically("~", 0.0, 1e-6))
			})

			// Collinear shapes that overlap only through the skin radius,
			// with the origin on the flat difference's line but outside its
			// span. The closest feature is a single Minkowski vertex and the
			// contact is shallow along the shared axis. Guards the collinear
			// vertex-region handling on the non-contained side.
			It("resolves a capsule approaching a collinear segment end to end", func() {
				rotation := shape2d.RotationFromAngle(dprec.Degrees(45.0))
				a := gjk2d.Shape{
					Position: dprec.ZeroVec2(),
					Rotation: rotation,
					Points: []dprec.Vec2{
						dprec.NewVec2(-1.0, 0.0),
						dprec.NewVec2(1.0, 0.0),
					},
					SkinRadius: 0.5,
				}
				b := gjk2d.Shape{
					Position: dprec.ZeroVec2(),
					Rotation: rotation,
					Points: []dprec.Vec2{
						dprec.NewVec2(1.3, 0.0),
						dprec.NewVec2(3.0, 0.0),
					},
					SkinRadius: 0.0,
				}

				contact := resolveAndSeparate(a, b)
				axis := rotation.Apply(dprec.BasisXVec2())
				Expect(contact.Depth).To(BeNumerically("~", 0.2, 1e-6))
				expectVec2(contact.TargetNormal, dprec.InverseVec2(axis))
				expectVec2(contact.TargetPoint, dprec.Vec2Prod(axis, 1.3))
			})
		})

		Describe("consistency with Intersect", func() {
			type pair struct {
				a, b gjk2d.Shape
			}
			cases := []pair{
				{
					a: gjk2d.ShapeFromCircle(shape2d.Circle{Center: dprec.NewVec2(0.0, 0.0), Radius: 1.0}),
					b: gjk2d.ShapeFromCircle(shape2d.Circle{Center: dprec.NewVec2(1.5, 0.0), Radius: 1.0}),
				},
				{
					a: gjk2d.ShapeFromCircle(shape2d.Circle{Center: dprec.NewVec2(0.0, 0.0), Radius: 1.0}),
					b: gjk2d.ShapeFromCircle(shape2d.Circle{Center: dprec.NewVec2(3.0, 0.0), Radius: 1.0}),
				},
				{
					a: gjk2d.ShapeFromRectangle(shape2d.Rectangle{Center: dprec.NewVec2(0.0, 0.0), Rotation: shape2d.IdentityRotation(), HalfWidth: 1.0, HalfHeight: 1.0}),
					b: gjk2d.ShapeFromRectangle(shape2d.Rectangle{Center: dprec.NewVec2(1.0, 0.0), Rotation: shape2d.IdentityRotation(), HalfWidth: 1.0, HalfHeight: 1.0}),
				},
				{
					a: gjk2d.ShapeFromRectangle(shape2d.Rectangle{Center: dprec.NewVec2(0.0, 0.0), Rotation: shape2d.IdentityRotation(), HalfWidth: 1.0, HalfHeight: 1.0}),
					b: gjk2d.ShapeFromRectangle(shape2d.Rectangle{Center: dprec.NewVec2(4.0, 0.0), Rotation: shape2d.IdentityRotation(), HalfWidth: 1.0, HalfHeight: 1.0}),
				},
				{
					a: gjk2d.ShapeFromCapsule(shape2d.Capsule{A: dprec.NewVec2(-1.0, 0.0), B: dprec.NewVec2(1.0, 0.0), Radius: 0.1}),
					b: gjk2d.ShapeFromCapsule(shape2d.Capsule{A: dprec.NewVec2(0.0, -1.0), B: dprec.NewVec2(0.0, 1.0), Radius: 0.1}),
				},
			}

			It("agrees with Intersect on whether the shapes overlap", func() {
				for _, c := range cases {
					_, ok := solver.Resolve(c.a, c.b)
					Expect(ok).To(Equal(solver.Intersect(c.a, c.b)))
				}
			})
		})

		Describe("edge cases", func() {
			It("returns false when a shape has no points", func() {
				a := gjk2d.Shape{}
				b := gjk2d.ShapeFromCircle(shape2d.Circle{Center: dprec.NewVec2(0.0, 0.0), Radius: 1.0})
				_, ok := solver.Resolve(a, b)
				Expect(ok).To(BeFalse())
			})
		})
	})

})
