package animation_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation"
)

var _ = Describe("BlendNode", func() {

	const bone = 0

	var (
		skeleton    *animation.Skeleton
		walkClip    *animation.Clip
		runClip     *animation.Clip
		firstPose   animation.Transform
		secondPose  animation.Transform
		halfwayPose animation.Transform
	)

	// translationClip creates a looping clip that moves the bone linearly
	// from the zero translation to the specified one over the specified
	// length, without rotating or scaling it.
	translationClip := func(name string, length float64, end dprec.Vec3) *animation.Clip {
		return animation.NewClip(name, 0.0, length, map[string]animation.Channel{
			"bone": animation.NewComponentKeyframeChannel(
				animation.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.ZeroVec3()},
					{Timestamp: length, Value: end},
				},
				animation.KeyframeList[dprec.Quat]{
					{Timestamp: 0.0, Value: dprec.IdentityQuat()},
				},
				animation.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.NewVec3(1.0, 1.0, 1.0)},
				},
			),
		})
	}

	// walkNode creates a looping node that moves the bone along X from 0.0
	// to 2.0 over two seconds (a rate of 0.5).
	walkNode := func() *animation.ClipNode {
		return animation.NewClipNode(skeleton, walkClip, true)
	}

	// runNode creates a looping node that moves the bone along Y from 0.0 to
	// 1.0 over one second (a rate of 1.0).
	runNode := func() *animation.ClipNode {
		return animation.NewClipNode(skeleton, runClip, true)
	}

	poseNode := func(transform animation.Transform) *animation.FixedPoseNode {
		return animation.NewFixedPoseNode(skeleton, map[string]animation.Transform{
			"bone": transform,
		})
	}

	expectTransform := func(actual, expected animation.Transform) {
		GinkgoHelper()
		expectVec3(actual.Translation, expected.Translation)
		expectRotation(actual.Rotation, expected.Rotation)
		expectVec3(actual.Scale, expected.Scale)
	}

	BeforeEach(func() {
		skeleton = animation.NewSkeleton([]animation.SkeletonBone{
			{Name: "bone", RestTransform: animation.IdentityTransform()},
		})
		walkClip = translationClip("walk", 2.0, dprec.NewVec3(2.0, 0.0, 0.0))
		runClip = translationClip("run", 1.0, dprec.NewVec3(0.0, 1.0, 0.0))
		firstPose = animation.Transform{
			Translation: dprec.NewVec3(1.0, 0.0, 0.0),
			Rotation:    dprec.IdentityQuat(),
			Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
		}
		secondPose = animation.Transform{
			Translation: dprec.NewVec3(0.0, 2.0, 0.0),
			Rotation:    yRotation(90.0),
			Scale:       dprec.NewVec3(3.0, 3.0, 3.0),
		}
		halfwayPose = animation.Transform{
			Translation: dprec.NewVec3(0.5, 1.0, 0.0),
			Rotation:    yRotation(45.0),
			Scale:       dprec.NewVec3(2.0, 2.0, 2.0),
		}
	})

	Describe("NewBlendNode", func() {

		It("starts with a weight of 0.0", func() {
			node := animation.NewBlendNode(walkNode(), runNode())
			Expect(node.Weight()).To(Equal(0.0))
		})

		It("starts at a fraction of 0.0", func() {
			node := animation.NewBlendNode(walkNode(), runNode())
			Expect(node.Fraction()).To(Equal(0.0))
		})

		It("starts out as not synchronized", func() {
			node := animation.NewBlendNode(walkNode(), runNode())
			Expect(node.IsSynchronized()).To(BeFalse())
		})
	})

	Describe("Blend", func() {

		It("combines the nodes like NewBlendNode", func() {
			node := animation.Blend(poseNode(firstPose), poseNode(secondPose))
			Expect(node.Weight()).To(Equal(0.0))
			expectTransform(node.Transform(bone), firstPose)
		})
	})

	Describe("SetWeight", func() {

		DescribeTable("clamps the weight",
			func(weight, expected float64) {
				node := animation.NewBlendNode(walkNode(), runNode())
				node.SetWeight(weight)
				Expect(node.Weight()).To(Equal(expected))
			},
			Entry("in range", 0.25, 0.25),
			Entry("below the range", -1.0, 0.0),
			Entry("above the range", 2.0, 1.0),
		)
	})

	Describe("Rate", func() {

		DescribeTable("blends the cycle lengths of synchronized children",
			func(weight, expected float64) {
				node := animation.NewBlendNode(
					animation.Synchronized(walkNode()),
					animation.Synchronized(runNode()),
				)
				node.SetWeight(weight)
				Expect(node.Rate()).To(BeNumerically("~", expected, 1e-9))
			},
			Entry("first only", 0.0, 0.5),
			Entry("second only", 1.0, 1.0),
			Entry("halfway, a cycle of 1.5 seconds", 0.5, 1.0/1.5),
			Entry("a quarter, a cycle of 1.75 seconds", 0.25, 1.0/1.75),
		)

		It("ignores a child that is not synchronized", func() {
			node := animation.NewBlendNode(animation.Synchronized(walkNode()), runNode())
			node.SetWeight(0.75)
			Expect(node.Rate()).To(BeNumerically("~", 0.5, 1e-9))
		})

		It("ignores a synchronized child with a rate of 0.0", func() {
			node := animation.NewBlendNode(
				animation.Synchronized(poseNode(firstPose)),
				animation.Synchronized(runNode()),
			)
			node.SetWeight(0.25)
			Expect(node.Rate()).To(BeNumerically("~", 1.0, 1e-9))
		})

		It("is 0.0 when no child is synchronized", func() {
			node := animation.NewBlendNode(walkNode(), runNode())
			node.SetWeight(0.5)
			Expect(node.Rate()).To(Equal(0.0))
		})
	})

	Describe("SetFraction", func() {

		It("wraps the fraction", func() {
			node := animation.NewBlendNode(walkNode(), runNode())
			node.SetFraction(1.25)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			node.SetFraction(-0.25)
			Expect(node.Fraction()).To(BeNumerically("~", 0.75, 1e-9))
		})

		It("relocates synchronized children and leaves the others in place", func() {
			first := animation.Synchronized(walkNode())
			second := runNode()
			node := animation.NewBlendNode(first, second)
			node.SetFraction(0.5)
			Expect(first.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			Expect(second.Fraction()).To(Equal(0.0))
		})
	})

	Describe("Advance", func() {

		It("keeps synchronized children of different lengths in phase", func() {
			first := animation.Synchronized(walkNode())
			second := animation.Synchronized(runNode())
			node := animation.NewBlendNode(first, second)
			node.SetWeight(0.5)
			node.Advance(0.75) // half of the blended cycle of 1.5 seconds
			Expect(node.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			Expect(first.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			Expect(second.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
		})

		It("lets a child that is not synchronized play at its own pace", func() {
			first := animation.Synchronized(walkNode())
			second := runNode()
			node := animation.NewBlendNode(first, second)
			node.SetWeight(0.5)
			node.Advance(0.75)
			Expect(node.Fraction()).To(BeNumerically("~", 0.375, 1e-9))
			Expect(first.Fraction()).To(BeNumerically("~", 0.375, 1e-9))
			Expect(second.Fraction()).To(BeNumerically("~", 0.75, 1e-9))
		})

		It("wraps around at the end of the cycle", func() {
			node := animation.NewBlendNode(animation.Synchronized(walkNode()), runNode())
			node.Advance(2.5)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("wraps around at the start of the cycle for a negative delta", func() {
			node := animation.NewBlendNode(animation.Synchronized(walkNode()), runNode())
			node.Advance(-0.5)
			Expect(node.Fraction()).To(BeNumerically("~", 0.75, 1e-9))
		})

		It("keeps its fraction when no child is synchronized", func() {
			first := walkNode()
			second := runNode()
			node := animation.NewBlendNode(first, second)
			node.SetFraction(0.25)
			node.Advance(0.5)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(first.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(second.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
		})
	})

	Describe("Transform", func() {

		DescribeTable("blends the children according to the weight",
			func(weight float64, expected func() animation.Transform) {
				node := animation.NewBlendNode(poseNode(firstPose), poseNode(secondPose))
				node.SetWeight(weight)
				expectTransform(node.Transform(bone), expected())
			},
			Entry("first only", 0.0, func() animation.Transform { return firstPose }),
			Entry("second only", 1.0, func() animation.Transform { return secondPose }),
			Entry("halfway", 0.5, func() animation.Transform { return halfwayPose }),
		)
	})

	Describe("DeltaTransform", func() {

		DescribeTable("blends the motion of the children according to the weight",
			func(weight float64, expected dprec.Vec3) {
				node := animation.NewBlendNode(walkNode(), runNode())
				node.SetWeight(weight)
				expectTransform(node.DeltaTransform(bone, 0.5), animation.Transform{
					Translation: expected,
					Rotation:    dprec.IdentityQuat(),
					Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
				})
			},
			Entry("first only", 0.0, dprec.NewVec3(0.5, 0.0, 0.0)),
			Entry("second only", 1.0, dprec.NewVec3(0.0, 0.5, 0.0)),
			Entry("halfway", 0.5, dprec.NewVec3(0.25, 0.25, 0.0)),
		)

		It("does not change the playback position", func() {
			first := animation.Synchronized(walkNode())
			node := animation.NewBlendNode(first, runNode())
			node.SetFraction(0.25)
			node.DeltaTransform(bone, 1.0)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(first.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})
	})

	Describe("SetSynchronized", func() {

		It("configures the node without affecting its children", func() {
			first := walkNode()
			second := runNode()
			node := animation.NewBlendNode(first, second)

			node.SetSynchronized(true)
			Expect(node.IsSynchronized()).To(BeTrue())
			Expect(first.IsSynchronized()).To(BeFalse())
			Expect(second.IsSynchronized()).To(BeFalse())

			node.SetSynchronized(false)
			Expect(node.IsSynchronized()).To(BeFalse())
		})
	})
})
