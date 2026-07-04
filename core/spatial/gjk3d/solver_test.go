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

	Describe("Resolve", func() {

		// moveBy returns a copy of the shape translated by the given offset.
		moveBy := func(shape gjk3d.Shape, offset dprec.Vec3) gjk3d.Shape {
			shape.Position = dprec.Vec3Sum(shape.Position, offset)
			return shape
		}

		// expectVec3 asserts that two vectors are approximately equal.
		expectVec3 := func(actual, expected dprec.Vec3) {
			Expect(actual.X).To(BeNumerically("~", expected.X, 1e-6))
			Expect(actual.Y).To(BeNumerically("~", expected.Y, 1e-6))
			Expect(actual.Z).To(BeNumerically("~", expected.Z, 1e-6))
		}

		Describe("sphere vs sphere", func() {
			It("resolves overlapping spheres along the line of centers", func() {
				a := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(0.0, 0.0, 0.0), 1.0))
				b := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(1.5, 0.0, 0.0), 1.0))

				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
				expectVec3(contact.TargetNormal, dprec.NewVec3(-1.0, 0.0, 0.0))
				expectVec3(contact.TargetPoint, dprec.NewVec3(0.5, 0.0, 0.0))
				expectVec3(contact.EvalSourcePoint(), dprec.NewVec3(1.0, 0.0, 0.0))
			})

			It("reports a barely-touching contact with zero depth", func() {
				a := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(0.0, 0.0, 0.0), 1.0))
				b := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(2.0, 0.0, 0.0), 1.0))

				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
				expectVec3(contact.TargetNormal, dprec.NewVec3(-1.0, 0.0, 0.0))
			})

			It("returns false when the spheres are separated", func() {
				a := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(0.0, 0.0, 0.0), 1.0))
				b := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(3.0, 0.0, 0.0), 1.0))

				_, ok := solver.Resolve(a, b)
				Expect(ok).To(BeFalse())
			})
		})

		Describe("box vs box", func() {
			It("resolves along the axis of minimum penetration", func() {
				a := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(1.5, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))

				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
				expectVec3(contact.TargetNormal, dprec.NewVec3(-1.0, 0.0, 0.0))
			})

			It("returns false when the boxes are separated", func() {
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

				_, ok := solver.Resolve(a, b)
				Expect(ok).To(BeFalse())
			})
		})

		Describe("capsule vs capsule", func() {
			// Two capsules crossing perpendicularly within the same z=0 plane.
			// Their Minkowski difference is a flat, zero-thickness rectangle, so
			// the cheapest separation lifts one capsule out of the plane by the
			// combined skin radius (0.6 along z) rather than sliding it in-plane.
			// Guards the flat-containment regime of the EPA solver.
			It("resolves coplanar crossing capsules out of their plane", func() {
				a := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(-1.0, 0.0, 0.0),
					dprec.NewVec3(1.0, 0.0, 0.0),
				), 0.3)
				b := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(0.5, -2.0, 0.0),
					dprec.NewVec3(0.5, 2.0, 0.0),
				), 0.3)

				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				Expect(contact.Depth).To(BeNumerically("~", 0.6, 1e-6))
				Expect(dprec.Abs(contact.TargetNormal.Z)).To(BeNumerically("~", 1.0, 1e-6))

				// The reported contact must be geometrically consistent: moving
				// the source shape along the normal by the depth must separate.
				offset := dprec.Vec3Prod(contact.TargetNormal, contact.Depth+1e-4)
				movedA := a
				movedA.Position = dprec.Vec3Sum(movedA.Position, offset)
				Expect(solver.Intersect(movedA, b)).To(BeFalse())
			})

			// Two capsules crossing with a z offset so their skins overlap only
			// near the crossing point, giving a genuine, shallow 3D penetration.
			It("resolves crossing capsules offset in depth", func() {
				a := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(-1.0, 0.0, 0.0),
					dprec.NewVec3(1.0, 0.0, 0.0),
				), 0.3)
				b := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(0.0, -1.0, 0.4),
					dprec.NewVec3(0.0, 1.0, 0.4),
				), 0.3)

				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				Expect(contact.Depth).To(BeNumerically("~", 0.2, 1e-6))
				// The source (z=0) sits below the target (z=0.4), so it separates
				// downward, away from the target.
				expectVec3(contact.TargetNormal, dprec.NewVec3(0.0, 0.0, -1.0))
			})
		})

		Describe("degenerate features", func() {
			It("resolves exactly coincident spheres", func() {
				a := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(1.0, 2.0, 3.0), 1.0))
				b := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(1.0, 2.0, 3.0), 1.0))

				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				Expect(contact.Depth).To(BeNumerically("~", 2.0, 1e-6))
			})

			It("resolves a sphere centered exactly on a capsule endpoint", func() {
				a := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(0.0, 0.0, 0.0),
					dprec.NewVec3(2.0, 0.0, 0.0),
				), 0.5)
				b := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(0.0, 0.0, 0.0), 0.5))

				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				Expect(contact.Depth).To(BeNumerically("~", 1.0, 1e-6))
			})
		})

		Describe("separation", func() {
			// Moving the source shape along the contact normal by slightly more
			// than the reported depth must end the overlap. This is the key
			// end-to-end check on the EPA normal and depth.
			separates := func(a, b gjk3d.Shape) {
				contact, ok := solver.Resolve(a, b)
				Expect(ok).To(BeTrue())
				offset := dprec.Vec3Prod(contact.TargetNormal, contact.Depth+1e-4)
				Expect(solver.Intersect(moveBy(a, offset), b)).To(BeFalse())
			}

			It("separates overlapping spheres", func() {
				a := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(0.0, 0.0, 0.0), 1.0))
				b := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(1.2, 0.3, -0.4), 1.0))
				separates(a, b)
			})

			It("separates overlapping boxes", func() {
				a := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(1.0, 0.5, 0.3),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				separates(a, b)
			})

			It("separates a rotated box overlapping a box", func() {
				rot := shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(35.0), dprec.BasisYVec3()))
				a := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.0, 0.0, 0.0),
					rot,
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(1.6, 0.2, 0.5),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				separates(a, b)
			})

			It("separates crossing capsules", func() {
				a := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(-1.0, 0.0, 0.0),
					dprec.NewVec3(1.0, 0.0, 0.0),
				), 0.3)
				b := gjk3d.ShapeFromCapsule(shape3d.NewSegment(
					dprec.NewVec3(0.0, -1.0, 0.1),
					dprec.NewVec3(0.0, 1.0, 0.1),
				), 0.3)
				separates(a, b)
			})

			It("separates a sphere overlapping a box", func() {
				a := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(1.3, 0.2, -0.1), 0.6))
				b := gjk3d.ShapeFromBox(shape3d.NewBox(
					dprec.NewVec3(0.0, 0.0, 0.0),
					shape3d.IdentityRotation(),
					dprec.NewVec3(1.0, 1.0, 1.0),
				))
				separates(a, b)
			})
		})

		Describe("consistency", func() {
			It("agrees with Intersect on whether the shapes overlap", func() {
				a := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(0.0, 0.0, 0.0), 1.0))
				b := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(1.5, 0.0, 0.0), 1.0))
				_, ok := solver.Resolve(a, b)
				Expect(ok).To(Equal(solver.Intersect(a, b)))
			})

			It("returns false when a shape has no points", func() {
				a := gjk3d.Shape{}
				b := gjk3d.ShapeFromSphere(shape3d.NewSphere(dprec.NewVec3(0.0, 0.0, 0.0), 1.0))
				_, ok := solver.Resolve(a, b)
				Expect(ok).To(BeFalse())
			})
		})
	})
})
