package animation_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation"
)

var _ = Describe("FixedPoseNode", func() {

	const (
		rootBone  = 0
		childBone = 1
	)

	var (
		skeleton      *animation.Skeleton
		rootRest      animation.Transform
		childRest     animation.Transform
		poseTransform animation.Transform
		node          *animation.FixedPoseNode
	)

	BeforeEach(func() {
		rootRest = animation.Transform{
			Translation: dprec.NewVec3(1.0, 0.0, 0.0),
			Rotation:    dprec.IdentityQuat(),
			Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
		}
		childRest = animation.Transform{
			Translation: dprec.NewVec3(0.0, 2.0, 0.0),
			Rotation:    dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3()),
			Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
		}
		poseTransform = animation.Transform{
			Translation: dprec.NewVec3(3.0, 4.0, 5.0),
			Rotation:    dprec.RotationQuat(dprec.Degrees(45.0), dprec.BasisYVec3()),
			Scale:       dprec.NewVec3(2.0, 2.0, 2.0),
		}
		skeleton = animation.NewSkeleton([]animation.SkeletonBone{
			{Name: "root", RestTransform: rootRest},
			{Name: "child", RestTransform: childRest},
		})
		node = animation.NewFixedPoseNode(skeleton, map[string]animation.Transform{
			"child": poseTransform,
		})
	})

	expectTransform := func(actual, expected animation.Transform) {
		GinkgoHelper()
		expectVec3(actual.Translation, expected.Translation)
		expectRotation(actual.Rotation, expected.Rotation)
		expectVec3(actual.Scale, expected.Scale)
	}

	Describe("Transform", func() {

		It("returns the configured transform of a bone", func() {
			expectTransform(node.Transform(childBone), poseTransform)
		})

		It("returns the rest transform of a bone without an entry", func() {
			expectTransform(node.Transform(rootBone), rootRest)
		})

		It("returns the rest transforms for an empty map", func() {
			node := animation.NewFixedPoseNode(skeleton, nil)
			expectTransform(node.Transform(rootBone), rootRest)
			expectTransform(node.Transform(childBone), childRest)
		})

		It("ignores entries for bones that are not in the skeleton", func() {
			node := animation.NewFixedPoseNode(skeleton, map[string]animation.Transform{
				"unknown": poseTransform,
			})
			expectTransform(node.Transform(rootBone), rootRest)
			expectTransform(node.Transform(childBone), childRest)
		})

		It("does not retain the specified map", func() {
			transforms := map[string]animation.Transform{
				"child": poseTransform,
			}
			node := animation.NewFixedPoseNode(skeleton, transforms)
			transforms["child"] = animation.IdentityTransform()
			transforms["root"] = animation.IdentityTransform()
			expectTransform(node.Transform(childBone), poseTransform)
			expectTransform(node.Transform(rootBone), rootRest)
		})
	})

	Describe("playback", func() {

		It("has a rate of 0.0", func() {
			Expect(node.Rate()).To(Equal(0.0))
		})

		It("has a fraction of 0.0", func() {
			Expect(node.Fraction()).To(Equal(0.0))
		})

		It("is not affected by SetFraction", func() {
			node.SetFraction(0.5)
			Expect(node.Fraction()).To(Equal(0.0))
			expectTransform(node.Transform(childBone), poseTransform)
		})

		It("is not affected by Advance", func() {
			node.Advance(1.5)
			Expect(node.Fraction()).To(Equal(0.0))
			expectTransform(node.Transform(childBone), poseTransform)
		})
	})

	Describe("DeltaTransform", func() {

		DescribeTable("returns the identity transform",
			func(bone int, delta float64) {
				expectTransform(node.DeltaTransform(bone, delta), animation.IdentityTransform())
			},
			Entry("configured bone forward", childBone, 1.0),
			Entry("configured bone backward", childBone, -1.0),
			Entry("rest bone", rootBone, 1.0),
			Entry("zero delta", childBone, 0.0),
		)
	})

	Describe("SetSynchronized", func() {

		It("starts out as not synchronized", func() {
			Expect(node.IsSynchronized()).To(BeFalse())
		})

		It("stores the setting", func() {
			node.SetSynchronized(true)
			Expect(node.IsSynchronized()).To(BeTrue())
			node.SetSynchronized(false)
			Expect(node.IsSynchronized()).To(BeFalse())
		})

		It("works with Synchronized", func() {
			node := animation.Synchronized(animation.NewFixedPoseNode(skeleton, nil))
			Expect(node.IsSynchronized()).To(BeTrue())
		})
	})
})
