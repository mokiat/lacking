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

	Describe("DeltaTransform", func() {

		const (
			rootBone = 0
			headBone = 1
		)

		var (
			rootSkeleton *animation3.Skeleton
			walkClip     *animation3.Clip
		)

		// expectDelta checks the delta against a translation and a rotation
		// around the Y axis. All deltas in these specs have a unit scale.
		expectDelta := func(delta animation3.Transform, translation dprec.Vec3, degrees float64) {
			GinkgoHelper()
			expectVec3(delta.Translation, translation)
			expectRotation(delta.Rotation, yRotation(degrees))
			expectVec3(delta.Scale, dprec.NewVec3(1.0, 1.0, 1.0))
		}

		BeforeEach(func() {
			rootSkeleton = animation3.NewSkeleton([]animation3.SkeletonBone{
				{Name: "root", RestTransform: animation3.IdentityTransform()},
				{Name: "head", RestTransform: animation3.Transform{
					Translation: dprec.NewVec3(0.0, 2.0, 0.0),
					Rotation:    yRotation(20.0),
					Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
				}},
			})
			// The clip covers the keyframe range [1.0..2.0], so fractions map to
			// keyframe timestamps with an offset of 1.0. Using the notation
			// T(x, y, z) for a translation and R(a) for a rotation of a degrees
			// around the Y axis, the root pose B at each fraction is:
			//
			//   B(0.0) = identity
			//   B(0.4) = T(0, 0, 1)
			//   B(0.6) = T(0, 0, 1) * R(90)
			//   B(1.0) = T(0, 0, 2) * R(90)
			//
			// The root steps forward, turns in place between 0.4 and 0.6 and
			// steps forward again. Each repetition of the loop moves the root by
			// B(1.0) * B(0.0)^-1 = T(0, 0, 2) * R(90). Since R(90) maps +Z onto
			// +X, consecutive loops head in different directions, so the order
			// in which motions are combined is observable.
			walkClip = animation3.NewClip("walk", 1.0, 2.0, map[string]animation3.ClipChannel{
				"root": {
					TranslationKeyframes: animation3.KeyframeList[dprec.Vec3]{
						{Timestamp: 1.0, Value: dprec.NewVec3(0.0, 0.0, 0.0)},
						{Timestamp: 1.4, Value: dprec.NewVec3(0.0, 0.0, 1.0)},
						{Timestamp: 1.6, Value: dprec.NewVec3(0.0, 0.0, 1.0)},
						{Timestamp: 2.0, Value: dprec.NewVec3(0.0, 0.0, 2.0)},
					},
					RotationKeyframes: animation3.KeyframeList[dprec.Quat]{
						{Timestamp: 1.0, Value: yRotation(0.0)},
						{Timestamp: 1.4, Value: yRotation(0.0)},
						{Timestamp: 1.6, Value: yRotation(90.0)},
						{Timestamp: 2.0, Value: yRotation(90.0)},
					},
				},
			})
		})

		It("returns the identity transform for a zero delta", func() {
			node := animation3.NewClipNode(rootSkeleton, walkClip, true)
			node.SetFraction(0.6)
			Expect(node.DeltaTransform(rootBone, 0.0)).To(Equal(animation3.IdentityTransform()))
		})

		It("returns the identity transform for a clip with zero length", func() {
			node := animation3.NewClipNode(rootSkeleton, walkClip.Crop("empty", 1.5, 1.5), true)
			Expect(node.DeltaTransform(rootBone, 1.0)).To(Equal(animation3.IdentityTransform()))
		})

		It("returns no motion for a bone without a channel", func() {
			node := animation3.NewClipNode(rootSkeleton, walkClip, true)
			node.SetFraction(0.6)
			expectDelta(node.DeltaTransform(headBone, 1.8), dprec.ZeroVec3(), 0.0)
		})

		It("does not change the playback position", func() {
			node := animation3.NewClipNode(rootSkeleton, walkClip, true)
			node.SetFraction(0.6)
			node.DeltaTransform(rootBone, 1.8)
			Expect(node.Fraction()).To(BeNumerically("~", 0.6, 1e-9))
		})

		It("satisfies D * A = B for the poses at both positions", func() {
			node := animation3.NewClipNode(rootSkeleton, walkClip, false)
			node.SetFraction(0.2)
			before := node.Transform(rootBone).Matrix()
			delta := node.DeltaTransform(rootBone, 0.6).Matrix()
			node.Advance(0.6)
			after := node.Transform(rootBone).Matrix()

			combined := dprec.Mat4Prod(delta, before)
			expectVec3(combined.Translation(), after.Translation())
			expectRotation(combined.Rotation(), after.Rotation())
		})

		DescribeTable("non-looping node stops measuring at the ends",
			func(fraction, delta float64, translation dprec.Vec3, degrees float64) {
				node := animation3.NewClipNode(rootSkeleton, walkClip, false)
				node.SetFraction(fraction)
				expectDelta(node.DeltaTransform(rootBone, delta), translation, degrees)
			},
			// B(0.4) * B(0.0)^-1 = T(0, 0, 1)
			Entry("forward within range", 0.0, 0.4,
				dprec.NewVec3(0.0, 0.0, 1.0), 0.0),
			// B(0.4) * B(0.6)^-1 = T(0, 0, 1) * R(-90) * T(0, 0, -1) = T(1, 0, 1) * R(-90)
			Entry("backward within range", 0.6, -0.2,
				dprec.NewVec3(1.0, 0.0, 1.0), -90.0),
			// B(1.0) * B(0.6)^-1 = T(0, 0, 2) * T(0, 0, -1) = T(0, 0, 1)
			Entry("forward past the end", 0.6, 5.0,
				dprec.NewVec3(0.0, 0.0, 1.0), 0.0),
			// B(0.0) * B(0.6)^-1 = R(-90) * T(0, 0, -1) = T(1, 0, 0) * R(-90)
			Entry("backward past the start", 0.6, -5.0,
				dprec.NewVec3(1.0, 0.0, 0.0), -90.0),
		)

		DescribeTable("looping node continues each repetition from where the previous one ended",
			func(fraction, delta float64, translation dprec.Vec3, degrees float64) {
				node := animation3.NewClipNode(rootSkeleton, walkClip, true)
				node.SetFraction(fraction)
				expectDelta(node.DeltaTransform(rootBone, delta), translation, degrees)
			},
			// B(0.4) * B(0.0)^-1 = T(0, 0, 1)
			Entry("forward within range", 0.0, 0.4,
				dprec.NewVec3(0.0, 0.0, 1.0), 0.0),
			// B(1.0) * B(0.6)^-1 = T(0, 0, 1)
			Entry("forward exactly to the end", 0.6, 0.4,
				dprec.NewVec3(0.0, 0.0, 1.0), 0.0),
			// B(1.0) * B(0.4) * B(0.6)^-1 = T(1, 0, 2) * R(90) * R(-90) * T(0, 0, -1) = T(1, 0, 1)
			Entry("forward across one wrap", 0.6, 0.8,
				dprec.NewVec3(1.0, 0.0, 1.0), 0.0),
			// From fraction 0.6 to fraction 2.4. The root ends at
			// B(1.0) * B(1.0) * B(0.4) = T(2, 0, 2) * R(180) * T(0, 0, 1) = T(2, 0, 1) * R(180),
			// so the delta is T(2, 0, 1) * R(180) * B(0.6)^-1 = T(1, 0, 1) * R(90).
			// Combining the final segment before the loops instead would give
			// T(-1, 0, 3) * R(90).
			Entry("forward across two wraps", 0.6, 1.8,
				dprec.NewVec3(1.0, 0.0, 1.0), 90.0),
			// B(0.4) * B(0.6)^-1 = T(1, 0, 1) * R(-90)
			Entry("backward within range", 0.6, -0.2,
				dprec.NewVec3(1.0, 0.0, 1.0), -90.0),
			// Each repetition backwards moves the root by
			// B(0.0) * B(1.0)^-1 = T(2, 0, 0) * R(-90), and
			// B(0.6) * B(0.4)^-1 = T(-1, 0, 1) * R(90), so
			// T(2, 0, 0) * R(-90) * T(-1, 0, 1) * R(90) = T(1, 0, -1).
			Entry("backward across one wrap", 0.4, -0.8,
				dprec.NewVec3(1.0, 0.0, -1.0), 0.0),
			// From fraction 0.4 to fraction -1.4, the backward counterpart of the
			// forward case across two wraps:
			// (T(2, 0, 0) * R(-90))^2 * T(-1, 0, 1) * R(90)
			//   = T(2, 0, 2) * R(180) * T(-1, 0, 1) * R(90) = T(3, 0, 1) * R(-90).
			// This is not the inverse of the forward delta, since each delta is
			// relative to the pose that it starts from.
			Entry("backward across two wraps", 0.4, -1.8,
				dprec.NewVec3(3.0, 0.0, 1.0), -90.0),
		)
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
