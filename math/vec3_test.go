package math_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/mokiat/lacking/internal/testing/mathmatcher"
	. "github.com/mokiat/lacking/math"
)

var _ = Describe("Vec3", func() {

	Specify("ZeroVec3", func() {
		Expect(ZeroVec3()).To(HaveVec3Coords(0.0, 0.0, 0.0))
	})

})
