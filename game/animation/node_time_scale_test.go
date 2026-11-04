package animation_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation"
)

var _ = Describe("TimeScaleNode", func() {

	const bone = 0

	var (
		skeleton *animation.Skeleton
		clip     *animation.Clip
	)

	// slideNode creates a clip node whose bone slides along X from 0.0 to
	// 2.0 over two seconds, so the X translation of the bone equals twice the
	// fraction of the node.
	slideNode := func(loop bool) *animation.ClipNode {
		return animation.NewClipNode(skeleton, clip, loop)
	}

	BeforeEach(func() {
		skeleton = animation.NewSkeleton([]animation.SkeletonBone{
			{Name: "root", RestTransform: animation.IdentityTransform()},
		})
		clip = animation.NewClip("slide", 0.0, 2.0, map[string]animation.Channel{
			"root": animation.NewComponentKeyframeChannel(
				animation.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.NewVec3(0.0, 0.0, 0.0)},
					{Timestamp: 2.0, Value: dprec.NewVec3(2.0, 0.0, 0.0)},
				},
				animation.KeyframeList[dprec.Quat]{
					{Timestamp: 0.0, Value: dprec.IdentityQuat()},
				},
				animation.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.NewVec3(1.0, 1.0, 1.0)},
				},
			),
		})
	})

	Describe("NewTimeScaleNode", func() {

		It("starts with a time scale of 1.0", func() {
			node := animation.NewTimeScaleNode(slideNode(false))
			Expect(node.TimeScale()).To(Equal(1.0))
		})

		It("leaves the playback of the delegate unchanged", func() {
			delegate := slideNode(false)
			node := animation.NewTimeScaleNode(delegate)
			Expect(node.Rate()).To(Equal(delegate.Rate()))
			node.Advance(0.5)
			Expect(delegate.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})
	})

	Describe("ScaleTime", func() {

		It("wraps the node with the specified time scale", func() {
			delegate := slideNode(false)
			node := animation.ScaleTime(delegate, 2.0)
			Expect(node.TimeScale()).To(Equal(2.0))
			node.Advance(0.5)
			Expect(delegate.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
		})
	})

	Describe("SetTimeScale", func() {

		It("changes the time scale", func() {
			node := animation.NewTimeScaleNode(slideNode(false))
			node.SetTimeScale(0.5)
			Expect(node.TimeScale()).To(Equal(0.5))
		})
	})

	Describe("Rate", func() {

		DescribeTable("scales the rate of the delegate by the magnitude of the time scale",
			func(timeScale, expected float64) {
				node := animation.ScaleTime(slideNode(false), timeScale)
				Expect(node.Rate()).To(BeNumerically("~", expected, 1e-9))
			},
			Entry("faster", 2.0, 1.0),
			Entry("slower", 0.5, 0.25),
			Entry("frozen", 0.0, 0.0),
			Entry("backwards", -2.0, 1.0),
		)
	})

	Describe("Fraction", func() {

		It("returns the fraction of the delegate", func() {
			delegate := slideNode(false)
			delegate.SetFraction(0.25)
			node := animation.ScaleTime(delegate, 2.0)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})
	})

	Describe("SetFraction", func() {

		It("relocates the delegate regardless of the time scale", func() {
			delegate := slideNode(false)
			node := animation.ScaleTime(delegate, 2.0)
			node.SetFraction(0.25)
			Expect(delegate.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})
	})

	Describe("Advance", func() {

		DescribeTable("moves the delegate by the scaled delta",
			func(timeScale, seconds, expected float64) {
				delegate := slideNode(false)
				node := animation.ScaleTime(delegate, timeScale)
				node.SetFraction(0.5)
				node.Advance(seconds)
				Expect(delegate.Fraction()).To(BeNumerically("~", expected, 1e-9))
			},
			Entry("faster", 2.0, 0.25, 0.75),
			Entry("slower", 0.5, 1.0, 0.75),
			Entry("frozen", 0.0, 1.0, 0.5),
			Entry("backwards", -2.0, 0.25, 0.25),
			Entry("backwards for a negative delta", 2.0, -0.25, 0.25),
			Entry("forward for a negative delta and time scale", -2.0, -0.25, 0.75),
		)

		It("stops at the end of a non-looping delegate", func() {
			delegate := slideNode(false)
			node := animation.ScaleTime(delegate, 2.0)
			node.Advance(5.0)
			Expect(node.Fraction()).To(Equal(1.0))
		})

		It("wraps around a looping delegate", func() {
			delegate := slideNode(true)
			node := animation.ScaleTime(delegate, 2.0)
			node.Advance(1.25) // a full loop and a quarter
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})
	})

	Describe("Transform", func() {

		It("shows the delegate at its current position", func() {
			node := animation.ScaleTime(slideNode(false), 2.0)
			node.SetFraction(0.25)
			expectVec3(node.Transform(bone).Translation, dprec.NewVec3(0.5, 0.0, 0.0))
		})
	})

	Describe("DeltaTransform", func() {

		DescribeTable("measures the motion of the upcoming scaled playback",
			func(timeScale, expectedX float64) {
				node := animation.ScaleTime(slideNode(false), timeScale)
				node.SetFraction(0.5)
				delta := node.DeltaTransform(bone, 0.25)
				expectVec3(delta.Translation, dprec.NewVec3(expectedX, 0.0, 0.0))
				expectRotation(delta.Rotation, dprec.IdentityQuat())
				expectVec3(delta.Scale, dprec.NewVec3(1.0, 1.0, 1.0))
			},
			Entry("faster", 2.0, 0.5),
			Entry("slower", 0.5, 0.125),
			Entry("frozen", 0.0, 0.0),
			Entry("backwards", -2.0, -0.5),
		)

		It("satisfies D * A = B for the poses before and after advancing", func() {
			node := animation.ScaleTime(slideNode(true), 1.5)
			node.SetFraction(0.25)
			before := node.Transform(bone).Matrix()
			delta := node.DeltaTransform(bone, 0.5).Matrix()
			node.Advance(0.5)
			after := node.Transform(bone).Matrix()
			result := animation.Mat4Transform(dprec.Mat4Prod(delta, before))
			expected := animation.Mat4Transform(after)
			expectVec3(result.Translation, expected.Translation)
			expectRotation(result.Rotation, expected.Rotation)
			expectVec3(result.Scale, expected.Scale)
		})

		It("does not change the playback position", func() {
			node := animation.ScaleTime(slideNode(true), 2.0)
			node.SetFraction(0.25)
			node.DeltaTransform(bone, 1.0)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})
	})

	Describe("SetSynchronized", func() {

		It("configures the delegate", func() {
			delegate := slideNode(false)
			node := animation.NewTimeScaleNode(delegate)
			Expect(node.IsSynchronized()).To(BeFalse())

			node.SetSynchronized(true)
			Expect(delegate.IsSynchronized()).To(BeTrue())
			Expect(node.IsSynchronized()).To(BeTrue())

			node.SetSynchronized(false)
			Expect(delegate.IsSynchronized()).To(BeFalse())
			Expect(node.IsSynchronized()).To(BeFalse())
		})

		It("reflects a delegate that is already synchronized", func() {
			node := animation.NewTimeScaleNode(animation.Synchronized(slideNode(false)))
			Expect(node.IsSynchronized()).To(BeTrue())
		})
	})
})
