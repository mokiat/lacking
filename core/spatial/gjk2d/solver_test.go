package gjk2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk2d"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Intersect", func() {

	var solver *gjk2d.Solver

	BeforeEach(func() {
		solver = gjk2d.NewSolver()
	})

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
			Expect(gjk2d.Intersect(a, b)).To(BeFalse())
		})
	})

})
