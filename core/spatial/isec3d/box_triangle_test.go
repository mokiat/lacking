package isec3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/isec3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// haveApproxVec3Coords matches a vector against the given coordinates with a
// tolerance suited to contact points, which are exact only up to the small
// robustness inflation used when deriving the touch region.
func haveApproxVec3Coords(x, y, z float64) types.GomegaMatcher {
	return SatisfyAll(
		WithTransform(func(v dprec.Vec3) float64 { return v.X }, BeNumerically("~", x, 1e-6)),
		WithTransform(func(v dprec.Vec3) float64 { return v.Y }, BeNumerically("~", y, 1e-6)),
		WithTransform(func(v dprec.Vec3) float64 { return v.Z }, BeNumerically("~", z, 1e-6)),
	)
}

var _ = Describe("BoxTriangle", func() {
	// An axis-aligned unit cube centered at the origin, spanning [-1,1] on each
	// axis.
	var box shape3d.Box

	// A helper to build a floor-like triangle lying in a plane parallel to the
	// XY plane at the given height, large enough to cover the box's XY footprint
	// and with its normal pointing along +Z. The box only collides with it from
	// above (the +Z side).
	horizontalTriangle := func(z float64) shape3d.Triangle {
		return shape3d.Triangle{
			A: dprec.NewVec3(-2.0, -2.0, z),
			B: dprec.NewVec3(2.0, -2.0, z),
			C: dprec.NewVec3(0.0, 2.0, z),
		}
	}

	BeforeEach(func() {
		box = shape3d.Box{
			Center:     dprec.NewVec3(0.0, 0.0, 0.0),
			Rotation:   shape3d.IdentityRotation(),
			HalfWidth:  1.0,
			HalfHeight: 1.0,
			HalfLength: 1.0,
		}
	})

	Describe("CheckBoxTriangle", func() {
		It("returns true for a triangle slicing through the box from the front", func() {
			Expect(isec3d.CheckBoxTriangle(box, horizontalTriangle(-0.5))).To(BeTrue())
		})

		It("returns true for a triangle that just touches a face", func() {
			// The triangle lies in the box's bottom face plane z=-1, facing up.
			Expect(isec3d.CheckBoxTriangle(box, horizontalTriangle(-1.0))).To(BeTrue())
		})

		It("returns false for a triangle below the box", func() {
			Expect(isec3d.CheckBoxTriangle(box, horizontalTriangle(-2.0))).To(BeFalse())
		})

		It("returns false when the box center is behind the triangle", func() {
			// The triangle slices the box at z=0.5, but its normal points along +Z
			// while the box center at z=0 is behind its plane, so it is culled.
			Expect(isec3d.CheckBoxTriangle(box, horizontalTriangle(0.5))).To(BeFalse())
		})

		It("returns false when the box center lies exactly on the triangle's plane", func() {
			Expect(isec3d.CheckBoxTriangle(box, horizontalTriangle(0.0))).To(BeFalse())
		})

		It("returns false for a triangle beside the box", func() {
			// The triangle lies in the plane x=2, facing the box (-X), but the box
			// only reaches x=1.
			beside := shape3d.Triangle{
				A: dprec.NewVec3(2.0, -2.0, -2.0),
				B: dprec.NewVec3(2.0, 0.0, 2.0),
				C: dprec.NewVec3(2.0, 2.0, -2.0),
			}
			Expect(isec3d.CheckBoxTriangle(box, beside)).To(BeFalse())
		})

		It("returns false for a triangle whose plane clears a corner", func() {
			// The plane x+y+z=3.2 faces the origin and passes just beyond the
			// (1,1,1) corner, whose projection onto the plane normal is
			// sqrt(3) ~= 1.73 < 3.2/sqrt(3).
			cleared := shape3d.Triangle{
				A: dprec.NewVec3(3.2, 0.0, 0.0),
				B: dprec.NewVec3(0.0, 0.0, 3.2),
				C: dprec.NewVec3(0.0, 3.2, 0.0),
			}
			Expect(isec3d.CheckBoxTriangle(box, cleared)).To(BeFalse())
		})

		It("returns true for a triangle whose plane cuts a corner", func() {
			// The plane x+y+z=1.5 faces the origin and passes through the box near
			// the (1,1,1) corner.
			cut := shape3d.Triangle{
				A: dprec.NewVec3(1.5, 0.0, 0.0),
				B: dprec.NewVec3(0.0, 0.0, 1.5),
				C: dprec.NewVec3(0.0, 1.5, 0.0),
			}
			Expect(isec3d.CheckBoxTriangle(box, cut)).To(BeTrue())
		})
	})

	Describe("ResolveBoxTriangle", func() {
		resolve := func(b shape3d.Box, t shape3d.Triangle) (shape3d.Contact, bool) {
			var sink shape3d.LastContact
			isec3d.ResolveBoxTriangle(b, t, sink.AddContact)
			return sink.Contact()
		}

		It("yields a contact along the shallowest separating axis", func() {
			// The triangle slices the box at z=-0.5, so the box escapes upward
			// (+Z) by 0.5, the smallest of the per-axis overlaps.
			contact, ok := resolve(box, horizontalTriangle(-0.5))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
			Expect(contact.TargetPoint).To(haveApproxVec3Coords(0.0, 0.0, -0.5))
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})

		It("reports a zero-depth contact for a triangle touching a face", func() {
			contact, ok := resolve(box, horizontalTriangle(-1.0))
			Expect(ok).To(BeTrue())
			Expect(contact.Depth).To(BeNumerically("~", 0.0, 1e-6))
		})

		It("does not yield a contact for a disjoint triangle", func() {
			_, ok := resolve(box, horizontalTriangle(-2.0))
			Expect(ok).To(BeFalse())
		})

		It("does not yield a contact when the box center is behind the triangle", func() {
			_, ok := resolve(box, horizontalTriangle(0.5))
			Expect(ok).To(BeFalse())
		})

		It("reports a unit normal", func() {
			contact, ok := resolve(box, horizontalTriangle(-0.5))
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal.Length()).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("places the contact point on the triangle", func() {
			contact, ok := resolve(box, horizontalTriangle(-0.5))
			Expect(ok).To(BeTrue())
			// The contact point lies in the triangle's plane (z=-0.5).
			Expect(contact.TargetPoint.Z).To(BeNumerically("~", -0.5, 1e-6))
		})

		It("separates the box when it is moved by Depth along the normal", func() {
			contact, ok := resolve(box, horizontalTriangle(-0.5))
			Expect(ok).To(BeTrue())

			moved := box
			moved.Center = dprec.Vec3Sum(box.Center, dprec.Vec3Prod(contact.TargetNormal, contact.Depth))
			// After moving out by Depth the box only just touches, so a re-resolve
			// reports essentially zero remaining penetration.
			resolved, ok := resolve(moved, horizontalTriangle(-0.5))
			if ok {
				Expect(resolved.Depth).To(BeNumerically("~", 0.0, 1e-6))
			}
		})

		It("places the contact point under a tilted box's penetrating edge", func() {
			// The box is rotated 30 degrees about X, so its lowest feature is the
			// edge running along X through local corner (y=-1, z=-1), which sits at
			// world y = sin(30)-cos(30) ~= -0.366. The center height is chosen so
			// that this edge dips 0.1 below the floor plane z=0.
			reach := dprec.Cos(dprec.Degrees(30.0)) + dprec.Sin(dprec.Degrees(30.0))
			tilted := shape3d.Box{
				Center:     dprec.NewVec3(0.0, 0.0, reach-0.1),
				Rotation:   shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(30.0), dprec.BasisXVec3())),
				HalfWidth:  1.0,
				HalfHeight: 1.0,
				HalfLength: 1.0,
			}
			floor := shape3d.Triangle{
				A: dprec.NewVec3(-8.0, -8.0, 0.0),
				B: dprec.NewVec3(8.0, -8.0, 0.0),
				C: dprec.NewVec3(0.0, 8.0, 0.0),
			}
			contact, ok := resolve(tilted, floor)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.1, 1e-6))

			edgeY := dprec.Sin(dprec.Degrees(30.0)) - dprec.Cos(dprec.Degrees(30.0))
			Expect(contact.TargetPoint).To(haveApproxVec3Coords(0.0, edgeY, 0.0))
			// The source point is the midpoint of the penetrating edge itself.
			Expect(contact.EvalSourcePoint()).To(haveApproxVec3Coords(0.0, edgeY, -0.1))
		})

		It("places the contact point at a vertex penetrating a face", func() {
			// A triangle whose apex pokes 0.5 into the box's bottom face, with the
			// rest of it well below the box, so the shallowest escape is upward
			// along the box's own Z axis.
			spike := shape3d.Triangle{
				A: dprec.NewVec3(0.0, 0.0, -0.5),
				B: dprec.NewVec3(2.0, 0.5, -2.5),
				C: dprec.NewVec3(-2.0, 0.5, -2.5),
			}
			contact, ok := resolve(box, spike)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
			Expect(contact.TargetPoint).To(haveApproxVec3Coords(0.0, 0.0, -0.5))
			// The source point is the spot on the box face that the apex pierces.
			Expect(contact.EvalSourcePoint()).To(haveApproxVec3Coords(0.0, 0.0, -1.0))
		})

		It("resolves against an oriented box in world space", func() {
			// A box long along its local X, rotated 90 degrees about Z, so in world
			// space it spans +/-2 along Y and +/-1 along X and Z.
			rotated := shape3d.Box{
				Center:     dprec.NewVec3(0.0, 0.0, 0.0),
				Rotation:   shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3())),
				HalfWidth:  2.0,
				HalfHeight: 1.0,
				HalfLength: 1.0,
			}
			// A triangle facing +Y that slices the box at world y=-1.5; the box
			// escapes along +Y by 0.5.
			triangle := shape3d.Triangle{
				A: dprec.NewVec3(-2.0, -1.5, -2.0),
				B: dprec.NewVec3(0.0, -1.5, 2.0),
				C: dprec.NewVec3(2.0, -1.5, -2.0),
			}
			contact, ok := resolve(rotated, triangle)
			Expect(ok).To(BeTrue())
			Expect(contact.TargetNormal).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(contact.TargetPoint).To(haveApproxVec3Coords(0.0, -1.5, 0.0))
			Expect(contact.Depth).To(BeNumerically("~", 0.5, 1e-6))
		})
	})
})
