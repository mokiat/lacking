package shape3d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

var _ = Describe("Rotation", func() {
	var (
		identity shape3d.Rotation
		rotZ90   shape3d.Rotation
	)

	BeforeEach(func() {
		identity = shape3d.Rotation{
			BasisX: dprec.NewVec3(1.0, 0.0, 0.0),
			BasisY: dprec.NewVec3(0.0, 1.0, 0.0),
			BasisZ: dprec.NewVec3(0.0, 0.0, 1.0),
		}
		// A 90-degree counter-clockwise rotation about the Z axis.
		rotZ90 = shape3d.Rotation{
			BasisX: dprec.NewVec3(0.0, 1.0, 0.0),
			BasisY: dprec.NewVec3(-1.0, 0.0, 0.0),
			BasisZ: dprec.NewVec3(0.0, 0.0, 1.0),
		}
	})

	Describe("IdentityRotation", func() {
		It("maps each axis to itself", func() {
			r := shape3d.IdentityRotation()
			Expect(r.BasisX).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			Expect(r.BasisY).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(r.BasisZ).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
		})
	})

	Describe("RotationFromQuat", func() {
		It("produces the identity rotation for the identity quaternion", func() {
			r := shape3d.RotationFromQuat(dprec.IdentityQuat())
			Expect(r.BasisX).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			Expect(r.BasisY).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(r.BasisZ).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
		})

		It("produces a 90-degree rotation about the Z axis", func() {
			r := shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(90), dprec.BasisZVec3()))
			Expect(r.BasisX).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(r.BasisY).To(dprectest.HaveVec3Coords(-1.0, 0.0, 0.0))
			Expect(r.BasisZ).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
		})

		It("rotates the X axis to the Y axis for a 90-degree Z rotation", func() {
			r := shape3d.RotationFromQuat(dprec.RotationQuat(dprec.Degrees(90), dprec.BasisZVec3()))
			Expect(r.Apply(dprec.NewVec3(1.0, 0.0, 0.0))).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
		})
	})

	Describe("Quat", func() {
		It("returns the identity quaternion for the identity rotation", func() {
			Expect(identity.Quat()).To(dprectest.HaveQuatCoords(1.0, 0.0, 0.0, 0.0))
		})

		It("round-trips with RotationFromQuat", func() {
			quat := dprec.RotationQuat(dprec.Degrees(45), dprec.BasisYVec3())
			r := shape3d.RotationFromQuat(quat)
			result := r.Quat()
			Expect(result).To(dprectest.HaveQuatCoords(quat.W, quat.X, quat.Y, quat.Z))
		})
	})

	Describe("Apply", func() {
		It("leaves points unchanged for the identity rotation", func() {
			Expect(identity.Apply(dprec.NewVec3(3.0, 4.0, 5.0))).To(dprectest.HaveVec3Coords(3.0, 4.0, 5.0))
		})

		It("maps the X axis to BasisX", func() {
			Expect(rotZ90.Apply(dprec.NewVec3(1.0, 0.0, 0.0))).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
		})

		It("maps the Y axis to BasisY", func() {
			Expect(rotZ90.Apply(dprec.NewVec3(0.0, 1.0, 0.0))).To(dprectest.HaveVec3Coords(-1.0, 0.0, 0.0))
		})

		It("maps the Z axis to BasisZ", func() {
			Expect(rotZ90.Apply(dprec.NewVec3(0.0, 0.0, 1.0))).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
		})

		It("rotates a general point 90 degrees about the Z axis", func() {
			Expect(rotZ90.Apply(dprec.NewVec3(3.0, 4.0, 5.0))).To(dprectest.HaveVec3Coords(-4.0, 3.0, 5.0))
		})

		It("returns the zero vector for the zero input", func() {
			Expect(rotZ90.Apply(dprec.NewVec3(0.0, 0.0, 0.0))).To(dprectest.HaveVec3Coords(0.0, 0.0, 0.0))
		})
	})

	Describe("Inverse", func() {
		It("returns the identity when applied to the identity rotation", func() {
			inv := identity.Inverse()
			Expect(inv.BasisX).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			Expect(inv.BasisY).To(dprectest.HaveVec3Coords(0.0, 1.0, 0.0))
			Expect(inv.BasisZ).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
		})

		It("transposes the basis vectors of a 90-degree Z rotation", func() {
			inv := rotZ90.Inverse()
			Expect(inv.BasisX).To(dprectest.HaveVec3Coords(0.0, -1.0, 0.0))
			Expect(inv.BasisY).To(dprectest.HaveVec3Coords(1.0, 0.0, 0.0))
			Expect(inv.BasisZ).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
		})

		It("undoes Apply on the X axis", func() {
			point := dprec.NewVec3(1.0, 0.0, 0.0)
			Expect(rotZ90.Inverse().Apply(rotZ90.Apply(point))).To(dprectest.HaveVec3Coords(point.X, point.Y, point.Z))
		})

		It("undoes Apply on a general point", func() {
			point := dprec.NewVec3(3.0, 4.0, 5.0)
			Expect(rotZ90.Inverse().Apply(rotZ90.Apply(point))).To(dprectest.HaveVec3Coords(point.X, point.Y, point.Z))
		})

		It("is its own inverse for a 180-degree rotation", func() {
			rot180 := shape3d.Rotation{
				BasisX: dprec.NewVec3(-1.0, 0.0, 0.0),
				BasisY: dprec.NewVec3(0.0, -1.0, 0.0),
				BasisZ: dprec.NewVec3(0.0, 0.0, 1.0),
			}
			inv := rot180.Inverse()
			Expect(inv.BasisX).To(dprectest.HaveVec3Coords(-1.0, 0.0, 0.0))
			Expect(inv.BasisY).To(dprectest.HaveVec3Coords(0.0, -1.0, 0.0))
			Expect(inv.BasisZ).To(dprectest.HaveVec3Coords(0.0, 0.0, 1.0))
		})
	})
})
