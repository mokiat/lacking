package math_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/mokiat/lacking/internal/testing/mathmatcher"
	. "github.com/mokiat/lacking/math"
)

var _ = Describe("Basic", func() {

	Specify("Abs32", func() {
		Expect(Abs32(-0.1)).To(EqualFloat32(0.1))
		Expect(Abs32(-13.57)).To(EqualFloat32(13.57))
		Expect(Abs32(11.01)).To(EqualFloat32(11.01))
	})

})
