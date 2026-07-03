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
