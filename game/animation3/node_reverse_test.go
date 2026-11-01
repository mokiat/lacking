package animation3_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("ReverseNode", func() {

	const bone = 0

	var (
		skeleton *animation3.Skeleton
		clip     *animation3.Clip
	)

	// slideNode creates a clip node whose bone slides along X from 0.0 to
	// 2.0 over two seconds, so the X translation of the bone equals twice the
	// fraction of the node.
	slideNode := func(loop bool) *animation3.ClipNode {
		return animation3.NewClipNode(skeleton, clip, loop)
	}

	BeforeEach(func() {
		skeleton = animation3.NewSkeleton([]animation3.SkeletonBone{
			{Name: "root", RestTransform: animation3.IdentityTransform()},
		})
		clip = animation3.NewClip("slide", 0.0, 2.0, map[string]animation3.Channel{
			"root": animation3.NewComponentKeyframeChannel(
				animation3.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.NewVec3(0.0, 0.0, 0.0)},
					{Timestamp: 2.0, Value: dprec.NewVec3(2.0, 0.0, 0.0)},
				},
				animation3.KeyframeList[dprec.Quat]{
					{Timestamp: 0.0, Value: dprec.IdentityQuat()},
				},
				animation3.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.NewVec3(1.0, 1.0, 1.0)},
				},
			),
		})
	})

	Describe("Reverse", func() {

		It("wraps the node like NewReverseNode", func() {
			delegate := slideNode(false)
			delegate.SetFraction(0.25)
			node := animation3.Reverse(delegate)
			Expect(node.Fraction()).To(BeNumerically("~", 0.75, 1e-9))
		})
	})

	Describe("Rate", func() {

		It("returns the rate of the delegate", func() {
			delegate := slideNode(false)
			node := animation3.NewReverseNode(delegate)
			Expect(node.Rate()).To(Equal(delegate.Rate()))
			Expect(node.Rate()).To(BeNumerically("~", 0.5, 1e-9))
		})
	})

	Describe("Fraction", func() {

		DescribeTable("mirrors the fraction of the delegate",
			func(delegateFraction, expected float64) {
				delegate := slideNode(false)
				delegate.SetFraction(delegateFraction)
				node := animation3.NewReverseNode(delegate)
				Expect(node.Fraction()).To(BeNumerically("~", expected, 1e-9))
			},
			Entry("delegate at the start", 0.0, 1.0),
			Entry("delegate at a quarter", 0.25, 0.75),
			Entry("delegate at the middle", 0.5, 0.5),
			Entry("delegate at the end", 1.0, 0.0),
		)

		It("reports 1.0 for a looping delegate at its start", func() {
			node := animation3.NewReverseNode(slideNode(true))
			Expect(node.Fraction()).To(Equal(1.0))
		})
	})

	Describe("SetFraction", func() {

		DescribeTable("places a looping delegate at the mirrored fraction",
			func(fraction, expectedDelegate, expected float64) {
				delegate := slideNode(true)
				node := animation3.NewReverseNode(delegate)
				node.SetFraction(fraction)
				Expect(delegate.Fraction()).To(BeNumerically("~", expectedDelegate, 1e-9))
				Expect(node.Fraction()).To(BeNumerically("~", expected, 1e-9))
			},
			Entry("in range", 0.25, 0.75, 0.25),
			Entry("past the end", 1.3, 0.7, 0.3),
			Entry("before the start", -0.2, 0.2, 0.8),
		)

		DescribeTable("places a non-looping delegate at the mirrored fraction",
			func(fraction, expectedDelegate, expected float64) {
				delegate := slideNode(false)
				node := animation3.NewReverseNode(delegate)
				node.SetFraction(fraction)
				Expect(delegate.Fraction()).To(BeNumerically("~", expectedDelegate, 1e-9))
				Expect(node.Fraction()).To(BeNumerically("~", expected, 1e-9))
			},
			Entry("in range", 0.25, 0.75, 0.25),
			Entry("past the end", 1.3, 0.0, 1.0),
			Entry("before the start", -0.2, 1.0, 0.0),
		)
	})

	Describe("Advance", func() {

		It("moves the delegate backwards", func() {
			delegate := slideNode(false)
			node := animation3.NewReverseNode(delegate)
			node.SetFraction(0.0)
			node.Advance(0.5)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(delegate.Fraction()).To(BeNumerically("~", 0.75, 1e-9))
		})

		It("moves the delegate forward for a negative delta", func() {
			delegate := slideNode(false)
			node := animation3.NewReverseNode(delegate)
			node.SetFraction(0.5)
			node.Advance(-0.5)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(delegate.Fraction()).To(BeNumerically("~", 0.75, 1e-9))
		})

		It("stops at the start of a non-looping delegate", func() {
			delegate := slideNode(false)
			node := animation3.NewReverseNode(delegate)
			node.SetFraction(0.0)
			node.Advance(5.0)
			Expect(node.Fraction()).To(Equal(1.0))
			Expect(delegate.Fraction()).To(Equal(0.0))
		})

		It("wraps around a looping delegate", func() {
			delegate := slideNode(true)
			node := animation3.NewReverseNode(delegate)
			node.SetFraction(0.0) // places the delegate at its start
			node.Advance(2.5)     // a full loop and a quarter
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(delegate.Fraction()).To(BeNumerically("~", 0.75, 1e-9))
		})
	})

	Describe("Transform", func() {

		DescribeTable("shows the delegate at the mirrored position",
			func(fraction, expectedX float64) {
				node := animation3.NewReverseNode(slideNode(false))
				node.SetFraction(fraction)
				expectVec3(node.Transform(bone).Translation, dprec.NewVec3(expectedX, 0.0, 0.0))
			},
			Entry("at the start", 0.0, 2.0),
			Entry("at a quarter", 0.25, 1.5),
			Entry("at the end", 1.0, 0.0),
		)
	})

	Describe("DeltaTransform", func() {

		It("measures the motion of the upcoming reverse playback", func() {
			node := animation3.NewReverseNode(slideNode(false))
			node.SetFraction(0.25)
			delta := node.DeltaTransform(bone, 1.0)
			expectVec3(delta.Translation, dprec.NewVec3(-1.0, 0.0, 0.0))
			expectRotation(delta.Rotation, dprec.IdentityQuat())
			expectVec3(delta.Scale, dprec.NewVec3(1.0, 1.0, 1.0))
		})

		It("satisfies D * A = B for the poses before and after advancing", func() {
			node := animation3.NewReverseNode(slideNode(true))
			node.SetFraction(0.25)
			before := node.Transform(bone).Matrix()
			delta := node.DeltaTransform(bone, 0.5).Matrix()
			node.Advance(0.5)
			after := node.Transform(bone).Matrix()
			result := animation3.Mat4Transform(dprec.Mat4Prod(delta, before))
			expected := animation3.Mat4Transform(after)
			expectVec3(result.Translation, expected.Translation)
			expectRotation(result.Rotation, expected.Rotation)
			expectVec3(result.Scale, expected.Scale)
		})

		It("does not change the playback position", func() {
			node := animation3.NewReverseNode(slideNode(true))
			node.SetFraction(0.25)
			node.DeltaTransform(bone, 1.0)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})
	})

	Describe("SetSynchronized", func() {

		It("configures the delegate", func() {
			delegate := slideNode(false)
			node := animation3.NewReverseNode(delegate)
			Expect(node.IsSynchronized()).To(BeFalse())

			node.SetSynchronized(true)
			Expect(delegate.IsSynchronized()).To(BeTrue())
			Expect(node.IsSynchronized()).To(BeTrue())

			node.SetSynchronized(false)
			Expect(delegate.IsSynchronized()).To(BeFalse())
			Expect(node.IsSynchronized()).To(BeFalse())
		})

		It("reflects a delegate that is already synchronized", func() {
			node := animation3.NewReverseNode(animation3.Synchronized(slideNode(false)))
			Expect(node.IsSynchronized()).To(BeTrue())
		})
	})
})
