package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/core/spatial/shape2d"
)

var _ = Describe("Rotation", func() {
	var (
		identity shape2d.Rotation
		rot90    shape2d.Rotation
	)

	BeforeEach(func() {
		identity = shape2d.Rotation{
			BasisX: dprec.NewVec2(1.0, 0.0),
			BasisY: dprec.NewVec2(0.0, 1.0),
		}
		rot90 = shape2d.Rotation{
			BasisX: dprec.NewVec2(0.0, 1.0),
			BasisY: dprec.NewVec2(-1.0, 0.0),
		}
	})

	Describe("RotationFromCosSin", func() {
		It("produces the identity rotation for cos=1, sin=0", func() {
			r := shape2d.RotationFromCosSin(1.0, 0.0)
			Expect(r.BasisX).To(dprectest.HaveVec2Coords(1.0, 0.0))
			Expect(r.BasisY).To(dprectest.HaveVec2Coords(0.0, 1.0))
		})

		It("produces a 90-degree CCW rotation for cos=0, sin=1", func() {
			r := shape2d.RotationFromCosSin(0.0, 1.0)
			Expect(r.BasisX).To(dprectest.HaveVec2Coords(0.0, 1.0))
			Expect(r.BasisY).To(dprectest.HaveVec2Coords(-1.0, 0.0))
		})

		It("produces a 90-degree CW rotation for cos=0, sin=-1", func() {
			r := shape2d.RotationFromCosSin(0.0, -1.0)
			Expect(r.BasisX).To(dprectest.HaveVec2Coords(0.0, -1.0))
			Expect(r.BasisY).To(dprectest.HaveVec2Coords(1.0, 0.0))
		})
	})

	Describe("RotationFromAngle", func() {
		It("produces the identity rotation for 0 degrees", func() {
			r := shape2d.RotationFromAngle(dprec.Degrees(0))
			Expect(r.Apply(dprec.NewVec2(3.0, 4.0))).To(dprectest.HaveVec2Coords(3.0, 4.0))
		})

		It("rotates the X axis to the Y axis for 90 degrees", func() {
			r := shape2d.RotationFromAngle(dprec.Degrees(90))
			result := r.Apply(dprec.NewVec2(1.0, 0.0))
			Expect(result.X).To(BeNumerically("~", 0.0, 1e-6))
			Expect(result.Y).To(BeNumerically("~", 1.0, 1e-6))
		})

		It("rotates the X axis to negative Y for -90 degrees", func() {
			r := shape2d.RotationFromAngle(dprec.Degrees(-90))
			result := r.Apply(dprec.NewVec2(1.0, 0.0))
			Expect(result.X).To(BeNumerically("~", 0.0, 1e-6))
			Expect(result.Y).To(BeNumerically("~", -1.0, 1e-6))
		})

		It("rotates 180 degrees into a point reflection", func() {
			r := shape2d.RotationFromAngle(dprec.Degrees(180))
			result := r.Apply(dprec.NewVec2(3.0, 4.0))
			Expect(result.X).To(BeNumerically("~", -3.0, 1e-5))
			Expect(result.Y).To(BeNumerically("~", -4.0, 1e-5))
		})
	})

	Describe("Angle", func() {
		It("returns zero for the identity rotation", func() {
			Expect(identity.Angle()).To(BeNumerically("~", dprec.Degrees(0), 1e-6))
		})

		It("returns 90 degrees for a 90-degree CCW rotation", func() {
			Expect(rot90.Angle()).To(BeNumerically("~", dprec.Degrees(90), 1e-6))
		})

		It("round-trips with RotationFromAngle for a positive angle", func() {
			angle := dprec.Degrees(45)
			Expect(shape2d.RotationFromAngle(angle).Angle()).To(BeNumerically("~", angle, 1e-6))
		})

		It("round-trips with RotationFromAngle for a negative angle", func() {
			angle := dprec.Degrees(-135)
			Expect(shape2d.RotationFromAngle(angle).Angle()).To(BeNumerically("~", angle, 1e-6))
		})
	})

	Describe("Apply", func() {
		It("leaves points unchanged for the identity rotation", func() {
			Expect(identity.Apply(dprec.NewVec2(3.0, 4.0))).To(dprectest.HaveVec2Coords(3.0, 4.0))
		})

		It("maps the X axis to BasisX", func() {
			Expect(rot90.Apply(dprec.NewVec2(1.0, 0.0))).To(dprectest.HaveVec2Coords(0.0, 1.0))
		})

		It("maps the Y axis to BasisY", func() {
			Expect(rot90.Apply(dprec.NewVec2(0.0, 1.0))).To(dprectest.HaveVec2Coords(-1.0, 0.0))
		})

		It("rotates a general point 90 degrees counter-clockwise", func() {
			Expect(rot90.Apply(dprec.NewVec2(3.0, 4.0))).To(dprectest.HaveVec2Coords(-4.0, 3.0))
		})

		It("returns the zero vector for the zero input", func() {
			Expect(rot90.Apply(dprec.NewVec2(0.0, 0.0))).To(dprectest.HaveVec2Coords(0.0, 0.0))
		})
	})

	Describe("Inverse", func() {
		It("returns the identity when applied to the identity rotation", func() {
			inv := identity.Inverse()
			Expect(inv.BasisX).To(dprectest.HaveVec2Coords(1.0, 0.0))
			Expect(inv.BasisY).To(dprectest.HaveVec2Coords(0.0, 1.0))
		})

		It("returns the 90-degree clockwise rotation for a 90-degree counter-clockwise rotation", func() {
			inv := rot90.Inverse()
			Expect(inv.BasisX).To(dprectest.HaveVec2Coords(0.0, -1.0))
			Expect(inv.BasisY).To(dprectest.HaveVec2Coords(1.0, 0.0))
		})

		It("undoes Apply on the X axis", func() {
			point := dprec.NewVec2(1.0, 0.0)
			Expect(rot90.Inverse().Apply(rot90.Apply(point))).To(dprectest.HaveVec2Coords(point.X, point.Y))
		})

		It("undoes Apply on a general point", func() {
			point := dprec.NewVec2(3.0, 4.0)
			Expect(rot90.Inverse().Apply(rot90.Apply(point))).To(dprectest.HaveVec2Coords(point.X, point.Y))
		})

		It("is its own inverse for a 180-degree rotation", func() {
			rot180 := shape2d.Rotation{
				BasisX: dprec.NewVec2(-1.0, 0.0),
				BasisY: dprec.NewVec2(0.0, -1.0),
			}
			inv := rot180.Inverse()
			Expect(inv.BasisX).To(dprectest.HaveVec2Coords(-1.0, 0.0))
			Expect(inv.BasisY).To(dprectest.HaveVec2Coords(0.0, -1.0))
		})
	})
})
