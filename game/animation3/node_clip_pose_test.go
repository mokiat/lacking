package animation3_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("ClipPoseNode", func() {

	const (
		animatedBone = 0
		restBone     = 1
	)

	var (
		skeleton *animation3.Skeleton
		restPose animation3.Transform
		channels map[string]animation3.Channel
		clip     *animation3.Clip
	)

	// expectSlideX checks that the animated bone is at the specified X
	// translation, which equals the keyframe timestamp that was sampled.
	expectSlideX := func(node *animation3.ClipPoseNode, expectedX float64) {
		GinkgoHelper()
		expectVec3(node.Transform(animatedBone).Translation, dprec.NewVec3(expectedX, 0.0, 0.0))
	}

	BeforeEach(func() {
		restPose = animation3.Transform{
			Translation: dprec.NewVec3(0.0, 5.0, 0.0),
			Rotation:    dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3()),
			Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
		}
		skeleton = animation3.NewSkeleton([]animation3.SkeletonBone{
			{Name: "slider", RestTransform: animation3.IdentityTransform()},
			{Name: "idle", RestTransform: restPose},
		})
		// The slider moves along X from 0.0 to 10.0 over keyframe times
		// 0.0 to 10.0, so its X translation equals the sampled keyframe time.
		channels = map[string]animation3.Channel{
			"slider": animation3.NewComponentKeyframeChannel(
				animation3.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.NewVec3(0.0, 0.0, 0.0)},
					{Timestamp: 10.0, Value: dprec.NewVec3(10.0, 0.0, 0.0)},
				},
				animation3.KeyframeList[dprec.Quat]{
					{Timestamp: 0.0, Value: dprec.IdentityQuat()},
				},
				animation3.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.NewVec3(1.0, 1.0, 1.0)},
				},
			),
		}
		// The clip covers only part of the keyframe data, like a cropped
		// clip would.
		clip = animation3.NewClip("middle", 2.0, 6.0, channels)
	})

	Describe("NewClipPoseNode", func() {

		It("keeps the clip", func() {
			node := animation3.NewClipPoseNode(skeleton, clip, 1.0)
			Expect(node.Clip()).To(BeIdenticalTo(clip))
		})

		DescribeTable("clamps the timestamp into the clip",
			func(timestamp, expected float64) {
				node := animation3.NewClipPoseNode(skeleton, clip, timestamp)
				Expect(node.Timestamp()).To(Equal(expected))
			},
			Entry("in range", 1.5, 1.5),
			Entry("at the start", 0.0, 0.0),
			Entry("at the end", 4.0, 4.0),
			Entry("before the start", -1.0, 0.0),
			Entry("past the end", 99.0, 4.0),
		)
	})

	Describe("SetTimestamp", func() {

		DescribeTable("clamps the timestamp into the clip",
			func(timestamp, expected float64) {
				node := animation3.NewClipPoseNode(skeleton, clip, 1.0)
				node.SetTimestamp(timestamp)
				Expect(node.Timestamp()).To(Equal(expected))
			},
			Entry("in range", 2.5, 2.5),
			Entry("before the start", -1.0, 0.0),
			Entry("past the end", 99.0, 4.0),
		)

		It("changes the sampled pose", func() {
			node := animation3.NewClipPoseNode(skeleton, clip, 1.0)
			expectSlideX(node, 3.0)
			node.SetTimestamp(3.0)
			expectSlideX(node, 5.0)
		})

		It("pins a zero length clip at its start", func() {
			clip := animation3.NewClip("frame", 3.0, 3.0, channels)
			node := animation3.NewClipPoseNode(skeleton, clip, 1.0)
			Expect(node.Timestamp()).To(Equal(0.0))
			expectSlideX(node, 3.0)
		})
	})

	Describe("Transform", func() {

		DescribeTable("samples the clip relative to its start time",
			func(timestamp, expectedX float64) {
				node := animation3.NewClipPoseNode(skeleton, clip, timestamp)
				expectSlideX(node, expectedX)
			},
			Entry("at the start", 0.0, 2.0),
			Entry("in the middle", 1.5, 3.5),
			Entry("at the end", 4.0, 6.0),
			Entry("before the start", -1.0, 2.0),
			Entry("past the end", 99.0, 6.0),
		)

		It("returns the rest transform of a bone without a channel", func() {
			node := animation3.NewClipPoseNode(skeleton, clip, 1.0)
			transform := node.Transform(restBone)
			expectVec3(transform.Translation, restPose.Translation)
			expectRotation(transform.Rotation, restPose.Rotation)
			expectVec3(transform.Scale, restPose.Scale)
		})
	})

	Describe("playback", func() {

		var node *animation3.ClipPoseNode

		BeforeEach(func() {
			node = animation3.NewClipPoseNode(skeleton, clip, 1.0)
		})

		It("has a rate of 0.0", func() {
			Expect(node.Rate()).To(Equal(0.0))
		})

		It("has a fraction of 0.0", func() {
			Expect(node.Fraction()).To(Equal(0.0))
		})

		It("is not affected by SetFraction", func() {
			node.SetFraction(0.5)
			Expect(node.Fraction()).To(Equal(0.0))
			Expect(node.Timestamp()).To(Equal(1.0))
			expectSlideX(node, 3.0)
		})

		It("is not affected by Advance", func() {
			node.Advance(1.5)
			Expect(node.Fraction()).To(Equal(0.0))
			Expect(node.Timestamp()).To(Equal(1.0))
			expectSlideX(node, 3.0)
		})
	})

	Describe("DeltaTransform", func() {

		DescribeTable("returns the identity transform",
			func(bone int, delta float64) {
				node := animation3.NewClipPoseNode(skeleton, clip, 1.0)
				result := node.DeltaTransform(bone, delta)
				identity := animation3.IdentityTransform()
				expectVec3(result.Translation, identity.Translation)
				expectRotation(result.Rotation, identity.Rotation)
				expectVec3(result.Scale, identity.Scale)
			},
			Entry("animated bone forward", animatedBone, 1.0),
			Entry("animated bone backward", animatedBone, -1.0),
			Entry("rest bone", restBone, 1.0),
			Entry("zero delta", animatedBone, 0.0),
		)
	})

	Describe("SetSynchronized", func() {

		It("starts out as not synchronized", func() {
			node := animation3.NewClipPoseNode(skeleton, clip, 1.0)
			Expect(node.IsSynchronized()).To(BeFalse())
		})

		It("stores the setting", func() {
			node := animation3.NewClipPoseNode(skeleton, clip, 1.0)
			node.SetSynchronized(true)
			Expect(node.IsSynchronized()).To(BeTrue())
			node.SetSynchronized(false)
			Expect(node.IsSynchronized()).To(BeFalse())
		})

		It("works with Synchronized", func() {
			node := animation3.Synchronized(animation3.NewClipPoseNode(skeleton, clip, 1.0))
			Expect(node.IsSynchronized()).To(BeTrue())
		})
	})
})
