package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/testing/dprectest"
	"github.com/mokiat/lacking/util/shape2d"
)

var _ = Describe("Edge", func() {

	Specify("#Normal", func() {
		edge := shape2d.NewEdge(
			dprec.NewVec2(0.0, 0.0),
			dprec.NewVec2(0.0, 1.0),
		)
		normal := edge.Normal()
		Expect(normal).To(dprectest.HaveVec2Coords(1.0, 0.0))
	})

})
