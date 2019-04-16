package math_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/mokiat/lacking/internal/testing/mathmatcher"
	. "github.com/mokiat/lacking/math"
)

var _ = Describe("Vec2", func() {
	var nullVector Vec2
	var firstVector Vec2
	var secondVector Vec2

	BeforeEach(func() {
		nullVector = Vec2{}
		firstVector = Vec2{
			X: 2.0,
			Y: 3.0,
		}
		secondVector = Vec2{
			X: -1.0,
			Y: 2.0,
		}
	})

	Specify("NewVec2", func() {
		vector := NewVec2(9.8, 2.3)
		Expect(vector).To(HaveVec2Coords(9.8, 2.3))
	})

	Specify("Vec2Sum", func() {
		sum := Vec2Sum(firstVector, secondVector)
		Expect(sum).To(HaveVec2Coords(1.0, 5.0))
	})

	Specify("Vec2Diff", func() {
		sum := Vec2Diff(firstVector, secondVector)
		Expect(sum).To(HaveVec2Coords(3.0, 1.0))
	})

	Specify("Vec2Quot", func() {
		sum := Vec2Quot(firstVector, 2.0)
		Expect(sum).To(HaveVec2Coords(1.0, 1.5))
	})

	Specify("Vec2Dot", func() {
		dot := Vec2Dot(firstVector, secondVector)
		Expect(dot).To(EqualFloat32(4.0))
	})

	Specify("UnitVec2", func() {
		unit := UnitVec2(firstVector)
		Expect(unit).To(HaveVec2Coords(0.554700196225229, 0.832050294337844))
	})

	Specify("#IsZero", func() {
		Expect(nullVector.IsZero()).To(BeTrue())
		Expect(firstVector.IsZero()).To(BeFalse())
		Expect(NewVec2(Epsilon32, Epsilon32).IsZero()).To(BeFalse())
	})

	Specify("#Length", func() {
		lng := firstVector.Length()
		Expect(lng).To(EqualFloat32(3.605551275463989))
	})
})
