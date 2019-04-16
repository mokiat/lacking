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

	Specify("Eq32", func() {
		Expect(Eq32(0.000001, 0.000001)).To(BeTrue())
		Expect(Eq32(0.000001, 0.000002)).To(BeFalse())
		Expect(Eq32(0.0000003, 0.0000002)).To(BeTrue()) // outside precision
	})

	Specify("EqEps32", func() {
		Expect(EqEps32(0.01, 0.01, 0.01)).To(BeTrue())
		Expect(EqEps32(0.01, 0.02, 0.01)).To(BeFalse())
		Expect(EqEps32(0.003, 0.002, 0.01)).To(BeTrue()) // outside precision
	})

	Specify("Sqrt32", func() {
		Expect(Sqrt32(17.64)).To(EqualFloat32(4.2))
	})
})
