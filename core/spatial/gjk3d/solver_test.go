package gjk3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/gjk3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("Solver", func() {
	var solver *gjk3d.Solver

	BeforeEach(func() {
		solver = gjk3d.NewSolver()
	})

	Describe("Intersect", func() {
		Describe("sphere vs sphere", func() {
			It("returns false when spheres are clearly separated", func() {
				a := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(0.0, 0.0, 0.0),
					Radius: 1.0,
				})
				b := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(3.0, 0.0, 0.0),
					Radius: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("returns true when spheres clearly overlap", func() {
				a := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(0.0, 0.0, 0.0),
					Radius: 1.0,
				})
				b := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(1.5, 0.0, 0.0),
					Radius: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns true when one sphere contains the other", func() {
				a := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(0.0, 0.0, 0.0),
					Radius: 3.0,
				})
				b := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(0.5, 0.0, 0.0),
					Radius: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns true when spheres touch exactly at skin radius boundary", func() {
				a := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(0.0, 0.0, 0.0),
					Radius: 1.0,
				})
				b := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(2.0, 0.0, 0.0),
					Radius: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})
		})

		Describe("box vs box", func() {
			It("returns false when boxes are clearly separated", func() {
				a := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(4.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("returns true when boxes clearly overlap", func() {
				a := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(1.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns true when one box deeply contains the other", func() {
				a := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(3.0, 3.0, 3.0),
				))
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.5, 0.3, -0.2),
					shape3d.IdentityRotation(),
					dprec.NewVec3(0.5, 0.5, 0.5),
				))
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns true when a face of one box touches the face of another", func() {
				a := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				// Place b so its left face sits exactly at x=1 (right face of a).
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(2.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns false when rotated boxes pass by each other diagonally", func() {
				rot45 := shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(45.0), dprec.BasisZVec3()))
				a := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(0.5, 0.5, 0.5),
				))
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(2.0, 0.0, 0.0),
					rot45,
					dprec.NewVec3(0.5, 0.5, 0.5),
				))
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("returns false when boxes are separated along a diagonal direction", func() {
				a := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(2.5, 2.5, 2.5),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("returns true when boxes overlap at a corner", func() {
				a := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(1.8, 1.8, 1.8),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})
		})

		Describe("sphere vs box", func() {
			It("returns false when the sphere is clearly outside the box", func() {
				a := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(0.0, 0.0, 0.0),
					Radius: 1.0,
				})
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(4.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("returns true when the sphere overlaps the box", func() {
				a := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(0.0, 0.0, 0.0),
					Radius: 1.5,
				})
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(2.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns true when the sphere is fully inside the box", func() {
				a := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(0.0, 0.0, 0.0),
					Radius: 0.3,
				})
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns false when the sphere misses the box corner diagonally", func() {
				a := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(2.0, 2.0, 2.0),
					Radius: 1.0,
				})
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("returns true when the sphere reaches the box corner diagonally", func() {
				a := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(1.5, 1.5, 1.5),
					Radius: 1.0,
				})
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})
		})

		Describe("capsule vs capsule", func() {
			It("returns false when capsules are clearly separated", func() {
				a := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(-1.0, 0.0, 0.0),
					dprec.NewVec3(1.0, 0.0, 0.0),
				), 0.5)
				b := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(-1.0, 3.0, 0.0),
					dprec.NewVec3(1.0, 3.0, 0.0),
				), 0.5)
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("returns true when capsule end-caps overlap", func() {
				a := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(-1.0, 0.0, 0.0),
					dprec.NewVec3(0.0, 0.0, 0.0),
				), 0.5)
				b := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(0.5, 0.0, 0.0),
					dprec.NewVec3(1.5, 0.0, 0.0),
				), 0.5)
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns true when skew capsules cross", func() {
				a := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(-1.0, 0.0, 0.0),
					dprec.NewVec3(1.0, 0.0, 0.0),
				), 0.1)
				b := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(0.0, -1.0, 0.1),
					dprec.NewVec3(0.0, 1.0, 0.1),
				), 0.1)
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns false when skew capsules pass at a distance", func() {
				a := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(-1.0, 0.0, 0.0),
					dprec.NewVec3(1.0, 0.0, 0.0),
				), 0.1)
				b := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(0.0, -1.0, 0.5),
					dprec.NewVec3(0.0, 1.0, 0.5),
				), 0.1)
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})
		})

		Describe("triangle vs sphere", func() {
			It("returns false when the sphere hovers above the triangle plane", func() {
				a := gjk3d.ShapeFromTriangle(shape3d.NewTriangle(
					dprec.NewVec3(-1.0, 0.0, -1.0),
					dprec.NewVec3(1.0, 0.0, -1.0),
					dprec.NewVec3(0.0, 0.0, 1.0),
				))
				b := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(0.0, 2.0, 0.0),
					Radius: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("returns true when the sphere touches the triangle interior", func() {
				a := gjk3d.ShapeFromTriangle(shape3d.NewTriangle(
					dprec.NewVec3(-1.0, 0.0, -1.0),
					dprec.NewVec3(1.0, 0.0, -1.0),
					dprec.NewVec3(0.0, 0.0, 1.0),
				))
				b := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(0.0, 0.5, 0.0),
					Radius: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})

			It("returns false when the sphere is beside the triangle edge", func() {
				a := gjk3d.ShapeFromTriangle(shape3d.NewTriangle(
					dprec.NewVec3(-1.0, 0.0, -1.0),
					dprec.NewVec3(1.0, 0.0, -1.0),
					dprec.NewVec3(0.0, 0.0, 1.0),
				))
				b := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(0.0, 0.0, 3.0),
					Radius: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})
		})

		Describe("edge cases", func() {
			It("returns false when shape A has no points", func() {
				a := gjk3d.Shape{}
				b := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(0.0, 0.0, 0.0),
					Radius: 1.0,
				})
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("returns false when shape B has no points", func() {
				a := gjk3d.ShapeFromSphere(shape3d.Sphere{
					Center: dprec.NewVec3(0.0, 0.0, 0.0),
					Radius: 1.0,
				})
				b := gjk3d.Shape{}
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("handles collinear segment shapes", func() {
				a := gjk3d.ShapeFromSegment(shape3d.NewSegment(
					dprec.NewVec3(0.0, 5.0, 0.0),
					dprec.NewVec3(1.0, 5.0, 0.0),
				))
				b := gjk3d.ShapeFromSegment(shape3d.NewSegment(
					dprec.NewVec3(0.0, 0.0, 0.0),
					dprec.NewVec3(2.0, 0.0, 0.0),
				))
				Expect(solver.Intersect(a, b)).To(BeFalse())
			})

			It("handles overlapping collinear segment shapes", func() {
				a := gjk3d.ShapeFromSegment(shape3d.NewSegment(
					dprec.NewVec3(0.0, 0.0, 0.0),
					dprec.NewVec3(1.0, 0.0, 0.0),
				))
				b := gjk3d.ShapeFromSegment(shape3d.NewSegment(
					dprec.NewVec3(0.5, 0.0, 0.0),
					dprec.NewVec3(2.0, 0.0, 0.0),
				))
				Expect(solver.Intersect(a, b)).To(BeTrue())
			})
		})
	})
})
