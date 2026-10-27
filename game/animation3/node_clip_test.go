package animation3_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("ClipNode", func() {

	var (
		skeleton  *animation3.Skeleton
		clip      *animation3.Clip
		emptyClip *animation3.Clip
	)

	BeforeEach(func() {
		skeleton = animation3.NewSkeleton(nil)
		clip = animation3.NewClip("walk", 1.0, 3.0, nil)
		emptyClip = animation3.NewClip("empty", 2.0, 2.0, nil)
	})

	Describe("NewClipNode", func() {

		It("stores the clip", func() {
			node := animation3.NewClipNode(skeleton, clip, false)
			Expect(node.Clip()).To(BeIdenticalTo(clip))
		})

		It("starts at fraction zero", func() {
			node := animation3.NewClipNode(skeleton, clip, true)
			Expect(node.Fraction()).To(Equal(0.0))
		})

		It("starts as not synchronized", func() {
			node := animation3.NewClipNode(skeleton, clip, false)
			Expect(node.IsSynchronized()).To(BeFalse())
		})
	})

	Describe("Rate", func() {

		It("returns the reciprocal of the clip length", func() {
			node := animation3.NewClipNode(skeleton, clip, false)
			Expect(node.Rate()).To(BeNumerically("~", 0.5, 1e-9))
		})

		It("returns zero for a clip with zero length", func() {
			node := animation3.NewClipNode(skeleton, emptyClip, false)
			Expect(node.Rate()).To(Equal(0.0))
		})
	})

	Describe("SetFraction", func() {

		DescribeTable("non-looping node clamps the fraction",
			func(fraction, expected float64) {
				node := animation3.NewClipNode(skeleton, clip, false)
				node.SetFraction(fraction)
				Expect(node.Fraction()).To(BeNumerically("~", expected, 1e-9))
			},
			Entry("start", 0.0, 0.0),
			Entry("middle", 0.25, 0.25),
			Entry("end", 1.0, 1.0),
			Entry("past the end", 1.5, 1.0),
			Entry("before the start", -0.5, 0.0),
		)

		DescribeTable("looping node wraps the fraction",
			func(fraction, expected float64) {
				node := animation3.NewClipNode(skeleton, clip, true)
				node.SetFraction(fraction)
				Expect(node.Fraction()).To(BeNumerically("~", expected, 1e-9))
			},
			Entry("start", 0.0, 0.0),
			Entry("middle", 0.25, 0.25),
			Entry("end", 1.0, 0.0),
			Entry("past the end", 1.75, 0.75),
			Entry("before the start", -0.25, 0.75),
			Entry("multiple loops before the start", -2.25, 0.75),
			Entry("just before the start", -1e-20, 0.0),
		)

		It("keeps a clip with zero length at fraction zero", func() {
			node := animation3.NewClipNode(skeleton, emptyClip, false)
			node.SetFraction(0.5)
			Expect(node.Fraction()).To(Equal(0.0))
		})
	})

	Describe("Advance", func() {

		DescribeTable("non-looping node stops at the ends",
			func(start, seconds, expected float64) {
				node := animation3.NewClipNode(skeleton, clip, false)
				node.SetFraction(start)
				node.Advance(seconds)
				Expect(node.Fraction()).To(BeNumerically("~", expected, 1e-9))
			},
			Entry("forward within range", 0.0, 0.5, 0.25),
			Entry("backward within range", 0.5, -0.5, 0.25),
			Entry("forward to the end", 0.0, 2.0, 1.0),
			Entry("forward past the end", 0.5, 5.0, 1.0),
			Entry("backward past the start", 0.5, -5.0, 0.0),
			Entry("zero delta", 0.5, 0.0, 0.5),
		)

		DescribeTable("looping node wraps around",
			func(start, seconds, expected float64) {
				node := animation3.NewClipNode(skeleton, clip, true)
				node.SetFraction(start)
				node.Advance(seconds)
				Expect(node.Fraction()).To(BeNumerically("~", expected, 1e-9))
			},
			Entry("forward within range", 0.0, 0.5, 0.25),
			Entry("backward within range", 0.5, -0.5, 0.25),
			Entry("forward to the end", 0.0, 2.0, 0.0),
			Entry("forward past the end", 0.5, 2.5, 0.75),
			Entry("forward multiple loops", 0.0, 6.5, 0.25),
			Entry("backward past the start", 0.25, -1.0, 0.75),
			Entry("backward just past the start", 0.0, -1e-20, 0.0),
		)

		It("keeps a looping node below fraction one", func() {
			node := animation3.NewClipNode(skeleton, clip, true)
			node.Advance(-1e-20)
			Expect(node.Fraction()).To(BeNumerically("<", 1.0))
			node.SetFraction(-1e-20)
			Expect(node.Fraction()).To(BeNumerically("<", 1.0))
		})

		It("accumulates over multiple calls", func() {
			node := animation3.NewClipNode(skeleton, clip, true)
			for range 5 {
				node.Advance(0.5)
			}
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("has no effect on a clip with zero length", func() {
			node := animation3.NewClipNode(skeleton, emptyClip, true)
			node.Advance(1.5)
			Expect(node.Fraction()).To(Equal(0.0))
		})
	})

	Describe("Transform", func() {

		// quarterDegrees is the rotation angle that normalized linear
		// interpolation produces a quarter of the way between 0 and 90 degrees:
		// 2 * atan2(sin(45) / 4, 3 / 4 + cos(45) / 4). Constant angular velocity
		// would give 22.5 degrees instead.
		const quarterDegrees = 21.59816098369244

		const (
			hipBone   = 0
			spineBone = 1
			headBone  = 2
		)

		var (
			spineRest    animation3.Transform
			headRest     animation3.Transform
			animClip     *animation3.Clip
			animSkeleton *animation3.Skeleton
		)

		BeforeEach(func() {
			spineRest = animation3.Transform{
				Translation: dprec.NewVec3(0.0, 1.0, 0.0),
				Rotation:    yRotation(10.0),
				Scale:       dprec.NewVec3(2.0, 2.0, 2.0),
			}
			headRest = animation3.Transform{
				Translation: dprec.NewVec3(0.0, 2.0, 0.0),
				Rotation:    yRotation(20.0),
				Scale:       dprec.NewVec3(3.0, 3.0, 3.0),
			}
			animSkeleton = animation3.NewSkeleton([]animation3.SkeletonBone{
				{Name: "hip", RestTransform: animation3.Transform{
					Translation: dprec.NewVec3(5.0, 5.0, 5.0),
					Rotation:    yRotation(30.0),
					Scale:       dprec.NewVec3(4.0, 4.0, 4.0),
				}},
				{Name: "spine", RestTransform: spineRest},
				{Name: "head", RestTransform: headRest},
			})
			// The clip covers the keyframe range [1.0..3.0]. The hip translation
			// keyframes extend beyond that range on both sides.
			animClip = animation3.NewClip("walk", 1.0, 3.0, map[string]animation3.ClipChannel{
				"hip": {
					TranslationKeyframes: animation3.KeyframeList[dprec.Vec3]{
						{Timestamp: 0.0, Value: dprec.NewVec3(0.0, 0.0, 0.0)},
						{Timestamp: 2.0, Value: dprec.NewVec3(2.0, 0.0, 0.0)},
						{Timestamp: 4.0, Value: dprec.NewVec3(4.0, 4.0, 0.0)},
					},
					RotationKeyframes: animation3.KeyframeList[dprec.Quat]{
						{Timestamp: 1.0, Value: yRotation(0.0)},
						{Timestamp: 3.0, Value: yRotation(90.0)},
					},
					ScaleKeyframes: animation3.KeyframeList[dprec.Vec3]{
						{Timestamp: 1.0, Value: dprec.NewVec3(1.0, 1.0, 1.0)},
						{Timestamp: 3.0, Value: dprec.NewVec3(3.0, 3.0, 3.0)},
					},
				},
				"spine": {
					TranslationKeyframes: animation3.KeyframeList[dprec.Vec3]{
						{Timestamp: 1.0, Value: dprec.NewVec3(7.0, 0.0, 0.0)},
					},
				},
				"tail": {
					TranslationKeyframes: animation3.KeyframeList[dprec.Vec3]{
						{Timestamp: 1.0, Value: dprec.NewVec3(9.0, 9.0, 9.0)},
					},
				},
			})
		})

		DescribeTable("interpolates the animated components relative to the clip start",
			func(fraction float64, translation dprec.Vec3, degrees float64, scale dprec.Vec3) {
				node := animation3.NewClipNode(animSkeleton, animClip, false)
				node.SetFraction(fraction)
				result := node.Transform(hipBone)
				expectVec3(result.Translation, translation)
				expectRotation(result.Rotation, yRotation(degrees))
				expectVec3(result.Scale, scale)
			},
			Entry("at the start", 0.0,
				dprec.NewVec3(1.0, 0.0, 0.0), 0.0, dprec.NewVec3(1.0, 1.0, 1.0)),
			Entry("at a quarter", 0.25,
				dprec.NewVec3(1.5, 0.0, 0.0), quarterDegrees, dprec.NewVec3(1.5, 1.5, 1.5)),
			Entry("on a keyframe", 0.5,
				dprec.NewVec3(2.0, 0.0, 0.0), 45.0, dprec.NewVec3(2.0, 2.0, 2.0)),
			Entry("at the end", 1.0,
				dprec.NewVec3(3.0, 2.0, 0.0), 90.0, dprec.NewVec3(3.0, 3.0, 3.0)),
		)

		It("follows the playback position after advancing", func() {
			node := animation3.NewClipNode(animSkeleton, animClip, true)
			node.Advance(2.5) // wraps around to 0.5 seconds into the clip
			result := node.Transform(hipBone)
			expectVec3(result.Translation, dprec.NewVec3(1.5, 0.0, 0.0))
			expectRotation(result.Rotation, yRotation(quarterDegrees))
			expectVec3(result.Scale, dprec.NewVec3(1.5, 1.5, 1.5))
		})

		It("holds the outermost keyframes outside their range", func() {
			wideClip := animClip.Crop("wide", 0.0, 4.0)
			node := animation3.NewClipNode(animSkeleton, wideClip, false)

			node.SetFraction(0.0)
			result := node.Transform(hipBone)
			expectRotation(result.Rotation, yRotation(0.0))
			expectVec3(result.Scale, dprec.NewVec3(1.0, 1.0, 1.0))

			node.SetFraction(1.0)
			result = node.Transform(hipBone)
			expectRotation(result.Rotation, yRotation(90.0))
			expectVec3(result.Scale, dprec.NewVec3(3.0, 3.0, 3.0))
		})

		It("holds the value of a single keyframe", func() {
			node := animation3.NewClipNode(animSkeleton, animClip, false)
			for _, fraction := range []float64{0.0, 0.5, 1.0} {
				node.SetFraction(fraction)
				expectVec3(node.Transform(spineBone).Translation, dprec.NewVec3(7.0, 0.0, 0.0))
			}
		})

		It("takes components that are not animated from the rest transform", func() {
			node := animation3.NewClipNode(animSkeleton, animClip, false)
			node.SetFraction(0.5)
			result := node.Transform(spineBone)
			expectVec3(result.Translation, dprec.NewVec3(7.0, 0.0, 0.0))
			expectRotation(result.Rotation, spineRest.Rotation)
			expectVec3(result.Scale, spineRest.Scale)
		})

		It("returns the rest transform for a bone without a channel", func() {
			node := animation3.NewClipNode(animSkeleton, animClip, false)
			node.SetFraction(0.5)
			Expect(node.Transform(headBone)).To(Equal(headRest))
		})
	})

	Describe("SetSynchronized", func() {

		It("configures the synchronization", func() {
			node := animation3.NewClipNode(skeleton, clip, false)
			node.SetSynchronized(true)
			Expect(node.IsSynchronized()).To(BeTrue())
			node.SetSynchronized(false)
			Expect(node.IsSynchronized()).To(BeFalse())
		})
	})

	Describe("Synchronized", func() {

		It("marks the node as synchronized and returns it", func() {
			node := animation3.NewClipNode(skeleton, clip, false)
			Expect(animation3.Synchronized(node)).To(BeIdenticalTo(node))
			Expect(node.IsSynchronized()).To(BeTrue())
		})
	})
})
