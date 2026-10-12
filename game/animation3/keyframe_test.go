package animation3_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("KeyframeList", func() {

	type Keyframe = animation3.Keyframe[float64]
	type KeyframeList = animation3.KeyframeList[float64]

	Describe("Sample", func() {

		When("the list is empty", func() {
			var list KeyframeList

			It("returns zero-value keyframes and a zero factor", func() {
				left, right, t := list.Sample(1.5)
				Expect(left).To(Equal(Keyframe{}))
				Expect(right).To(Equal(Keyframe{}))
				Expect(t).To(Equal(0.0))
			})

			It("handles a nil list", func() {
				left, right, t := KeyframeList(nil).Sample(0.0)
				Expect(left).To(Equal(Keyframe{}))
				Expect(right).To(Equal(Keyframe{}))
				Expect(t).To(Equal(0.0))
			})
		})

		When("the list has a single keyframe", func() {
			var (
				list     KeyframeList
				keyframe Keyframe
			)

			BeforeEach(func() {
				keyframe = Keyframe{Timestamp: 2.0, Value: 7.0}
				list = KeyframeList{keyframe}
			})

			DescribeTable("returns that keyframe on both sides with a zero factor",
				func(timestamp float64) {
					left, right, t := list.Sample(timestamp)
					Expect(left).To(Equal(keyframe))
					Expect(right).To(Equal(keyframe))
					Expect(t).To(Equal(0.0))
				},
				Entry("before the keyframe", 1.0),
				Entry("at the keyframe", 2.0),
				Entry("after the keyframe", 3.0),
			)
		})

		When("the list has two keyframes", func() {
			var (
				list   KeyframeList
				first  Keyframe
				second Keyframe
			)

			BeforeEach(func() {
				first = Keyframe{Timestamp: 1.0, Value: 10.0}
				second = Keyframe{Timestamp: 3.0, Value: 20.0}
				list = KeyframeList{first, second}
			})

			DescribeTable("returns both keyframes and the correct factor",
				func(timestamp, expectedT float64) {
					left, right, t := list.Sample(timestamp)
					Expect(left).To(Equal(first))
					Expect(right).To(Equal(second))
					Expect(t).To(BeNumerically("~", expectedT, 1e-9))
				},
				Entry("before the range clamps to the first", 0.0, 0.0),
				Entry("at the first keyframe", 1.0, 0.0),
				Entry("a quarter of the way", 1.5, 0.25),
				Entry("halfway", 2.0, 0.5),
				Entry("three quarters of the way", 2.5, 0.75),
				Entry("at the second keyframe", 3.0, 1.0),
				Entry("after the range clamps to the second", 10.0, 1.0),
			)
		})

		When("the list has many keyframes", func() {
			var list KeyframeList

			BeforeEach(func() {
				// Non-uniform spacing and an odd count to exercise the
				// search on both halves of the list.
				list = KeyframeList{
					{Timestamp: 0.0, Value: 0.0},
					{Timestamp: 0.5, Value: 5.0},
					{Timestamp: 1.0, Value: -5.0},
					{Timestamp: 2.0, Value: 15.0},
					{Timestamp: 4.0, Value: 25.0},
					{Timestamp: 4.5, Value: 20.0},
					{Timestamp: 8.0, Value: 90.0},
				}
			})

			DescribeTable("returns the surrounding keyframes and factor",
				func(timestamp float64, leftIndex, rightIndex int, expectedT float64) {
					left, right, t := list.Sample(timestamp)
					Expect(left).To(Equal(list[leftIndex]))
					Expect(right).To(Equal(list[rightIndex]))
					Expect(t).To(BeNumerically("~", expectedT, 1e-9))
				},
				Entry("between first and second", 0.25, 0, 1, 0.5),
				Entry("between second and third", 0.6, 1, 2, 0.2),
				Entry("between third and fourth", 1.75, 2, 3, 0.75),
				Entry("between fourth and fifth", 3.0, 3, 4, 0.5),
				Entry("between fifth and sixth", 4.1, 4, 5, 0.2),
				Entry("between sixth and seventh", 7.3, 5, 6, 0.8),
			)

			DescribeTable("resolves to the keyframe at exact timestamps",
				func(index int) {
					keyframe := list[index]
					left, right, t := list.Sample(keyframe.Timestamp)
					switch t {
					case 0.0:
						Expect(left).To(Equal(keyframe))
					case 1.0:
						Expect(right).To(Equal(keyframe))
					default:
						Fail("expected a factor of 0 or 1 at an exact timestamp")
					}
				},
				Entry("first", 0),
				Entry("second", 1),
				Entry("third", 2),
				Entry("fourth", 3),
				Entry("fifth", 4),
				Entry("sixth", 5),
				Entry("seventh", 6),
			)

			It("clamps to the first keyframe before the range", func() {
				left, right, t := list.Sample(-3.0)
				Expect(left).To(Equal(list[0]))
				Expect(right).To(Equal(list[1]))
				Expect(t).To(Equal(0.0))
			})

			It("clamps to the last keyframe after the range", func() {
				left, right, t := list.Sample(100.0)
				Expect(left).To(Equal(list[5]))
				Expect(right).To(Equal(list[6]))
				Expect(t).To(Equal(1.0))
			})

			It("returns a factor within [0, 1] across the whole range", func() {
				for timestamp := -1.0; timestamp <= 9.0; timestamp += 0.01 {
					left, right, t := list.Sample(timestamp)
					Expect(t).To(BeNumerically(">=", 0.0))
					Expect(t).To(BeNumerically("<=", 1.0))
					Expect(left.Timestamp).To(BeNumerically("<=", right.Timestamp))
				}
			})
		})

		When("keyframes share a timestamp", func() {

			It("returns a zero factor instead of dividing by zero", func() {
				list := KeyframeList{
					{Timestamp: 1.0, Value: 10.0},
					{Timestamp: 1.0, Value: 20.0},
				}
				for _, timestamp := range []float64{0.0, 1.0, 2.0} {
					_, _, t := list.Sample(timestamp)
					Expect(t).To(Equal(0.0))
				}
			})

			It("uses the matching duplicate on either side of the step", func() {
				list := KeyframeList{
					{Timestamp: 0.0, Value: 0.0},
					{Timestamp: 1.0, Value: 10.0},
					{Timestamp: 1.0, Value: 50.0},
					{Timestamp: 2.0, Value: 60.0},
				}

				left, right, t := list.Sample(0.5)
				Expect(left).To(Equal(list[0]))
				Expect(right).To(Equal(list[1]))
				Expect(t).To(BeNumerically("~", 0.5, 1e-9))

				left, right, t = list.Sample(1.5)
				Expect(left).To(Equal(list[2]))
				Expect(right).To(Equal(list[3]))
				Expect(t).To(BeNumerically("~", 0.5, 1e-9))
			})
		})

		It("works with non-numeric value types", func() {
			list := animation3.KeyframeList[string]{
				{Timestamp: 0.0, Value: "a"},
				{Timestamp: 1.0, Value: "b"},
			}
			left, right, t := list.Sample(0.25)
			Expect(left.Value).To(Equal("a"))
			Expect(right.Value).To(Equal("b"))
			Expect(t).To(BeNumerically("~", 0.25, 1e-9))
		})
	})

})
