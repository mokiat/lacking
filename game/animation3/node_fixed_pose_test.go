package animation3_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("FixedPoseNode", func() {

	const (
		rootBone  = 0
		childBone = 1
	)

	var (
		skeleton      *animation3.Skeleton
		rootRest      animation3.Transform
		childRest     animation3.Transform
		poseTransform animation3.Transform
		node          *animation3.FixedPoseNode
	)

	BeforeEach(func() {
		rootRest = animation3.Transform{
			Translation: dprec.NewVec3(1.0, 0.0, 0.0),
			Rotation:    dprec.IdentityQuat(),
			Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
		}
		childRest = animation3.Transform{
			Translation: dprec.NewVec3(0.0, 2.0, 0.0),
			Rotation:    dprec.RotationQuat(dprec.Degrees(90.0), dprec.BasisZVec3()),
			Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
		}
		poseTransform = animation3.Transform{
			Translation: dprec.NewVec3(3.0, 4.0, 5.0),
			Rotation:    dprec.RotationQuat(dprec.Degrees(45.0), dprec.BasisYVec3()),
			Scale:       dprec.NewVec3(2.0, 2.0, 2.0),
		}
		skeleton = animation3.NewSkeleton([]animation3.SkeletonBone{
			{Name: "root", RestTransform: rootRest},
			{Name: "child", RestTransform: childRest},
		})
		node = animation3.NewFixedPoseNode(skeleton, map[string]animation3.Transform{
			"child": poseTransform,
		})
	})

	expectTransform := func(actual, expected animation3.Transform) {
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
			node := animation3.NewFixedPoseNode(skeleton, nil)
			expectTransform(node.Transform(rootBone), rootRest)
			expectTransform(node.Transform(childBone), childRest)
		})

		It("ignores entries for bones that are not in the skeleton", func() {
			node := animation3.NewFixedPoseNode(skeleton, map[string]animation3.Transform{
				"unknown": poseTransform,
			})
			expectTransform(node.Transform(rootBone), rootRest)
			expectTransform(node.Transform(childBone), childRest)
		})

		It("does not retain the specified map", func() {
			transforms := map[string]animation3.Transform{
				"child": poseTransform,
			}
			node := animation3.NewFixedPoseNode(skeleton, transforms)
			transforms["child"] = animation3.IdentityTransform()
			transforms["root"] = animation3.IdentityTransform()
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
				expectTransform(node.DeltaTransform(bone, delta), animation3.IdentityTransform())
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
			node := animation3.Synchronized(animation3.NewFixedPoseNode(skeleton, nil))
			Expect(node.IsSynchronized()).To(BeTrue())
		})
	})
})
