package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/util/shape2d"
)

var _ = Describe("Transform", func() {
	var transform shape2d.Transform

	BeforeEach(func() {
		transform = shape2d.Transform{
			Translation: dprec.NewVec2(4.0, 5.0),
			Rotation:    dprec.Degrees(30.0),
		}
	})

	PSpecify("ChainedTransform", func() {
		parentTransform := shape2d.TRTransform(
			dprec.NewVec2(-2.0, -0.5), dprec.Degrees(90.0),
		)
		combined := shape2d.ChainedTransform(parentTransform, transform)
		Expect(combined.Translation).To(dprectest.HaveVec2Coords(3.0, -4.5))
		Expect(combined.Rotation.Degrees()).To(dprectest.EqualFloat64(120.0))
	})

	PSpecify("#Apply", func() {
		vector := dprec.NewVec2(1.0, 0.0)
		result := transform.Apply(vector)
		Expect(result).To(dprectest.HaveVec2Coords(
			4.0+0.8660254037844386, 5.0-0.5,
		))
		vector = dprec.NewVec2(0.0, 1.0)
		result = transform.Apply(vector)
		Expect(result).To(dprectest.HaveVec2Coords(
			4.0+0.5, 5.0+0.8660254037844386,
		))
	})
})
