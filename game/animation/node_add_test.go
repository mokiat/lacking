package animation_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation"
)

var _ = Describe("AddNode", func() {

	const (
		animatedBone = 0
		spareBone    = 1
	)

	var (
		skeleton  *animation.Skeleton
		slideClip *animation.Clip
		liftClip  *animation.Clip
		offset    animation.Transform
	)

	// translationClip creates a clip that moves the animated bone linearly
	// from the zero translation to the specified one over the specified
	// length, without rotating or scaling it.
	translationClip := func(name string, length float64, end dprec.Vec3) *animation.Clip {
		return animation.NewClip(name, 0.0, length, map[string]animation.Channel{
			"animated": animation.NewComponentKeyframeChannel(
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

	// slideNode creates a looping base node whose animated bone slides along
	// X from 0.0 to 2.0 over two seconds, so the X translation of the bone
	// equals twice the fraction of the node.
	slideNode := func() *animation.ClipNode {
		return animation.NewClipNode(skeleton, slideClip, true)
	}

	// liftNode creates a looping additive node whose animated bone lifts
	// along Y from 0.0 to 4.0 over four seconds, so the Y translation of the
	// bone equals four times the fraction of the node.
	liftNode := func() *animation.ClipNode {
		return animation.NewClipNode(skeleton, liftClip, true)
	}

	// offsetNode creates an additive node that holds the offset transform on
	// the animated bone and the additive neutral on the spare bone.
	offsetNode := func() *animation.FixedPoseNode {
		return animation.NewFixedPoseNode(skeleton, map[string]animation.Transform{
			"animated": offset,
		})
	}

	expectTransform := func(actual, expected animation.Transform) {
		GinkgoHelper()
		expectVec3(actual.Translation, expected.Translation)
		expectRotation(actual.Rotation, expected.Rotation)
		expectVec3(actual.Scale, expected.Scale)
	}

	BeforeEach(func() {
		// Identity rest transforms make bones that a node does not animate
		// valid as both absolute and additive output.
		skeleton = animation.NewSkeleton([]animation.SkeletonBone{
			{Name: "animated", RestTransform: animation.IdentityTransform()},
			{Name: "spare", RestTransform: animation.IdentityTransform()},
		})
		slideClip = translationClip("slide", 2.0, dprec.NewVec3(2.0, 0.0, 0.0))
		liftClip = translationClip("lift", 4.0, dprec.NewVec3(0.0, 4.0, 0.0))
		offset = animation.Transform{
			Translation: dprec.NewVec3(0.0, 1.0, 0.0),
			Rotation:    yRotation(90.0),
			Scale:       dprec.NewVec3(2.0, 2.0, 2.0),
		}
	})

	Describe("NewAddNode", func() {

		It("starts with a weight of 1.0", func() {
			node := animation.NewAddNode(slideNode(), offsetNode())
			Expect(node.Weight()).To(Equal(1.0))
		})

		It("starts out as not synchronized", func() {
			node := animation.NewAddNode(slideNode(), offsetNode())
			Expect(node.IsSynchronized()).To(BeFalse())
		})
	})

	Describe("Add", func() {

		It("combines the nodes like NewAddNode", func() {
			node := animation.Add(slideNode(), offsetNode())
			Expect(node.Weight()).To(Equal(1.0))
			expectVec3(node.Transform(animatedBone).Translation, dprec.NewVec3(0.0, 1.0, 0.0))
		})
	})

	Describe("SetWeight", func() {

		It("changes the weight", func() {
			node := animation.NewAddNode(slideNode(), offsetNode())
			node.SetWeight(0.25)
			Expect(node.Weight()).To(Equal(0.25))
		})
	})

	Describe("Rate", func() {

		It("returns the rate of the base node", func() {
			node := animation.NewAddNode(slideNode(), liftNode())
			Expect(node.Rate()).To(BeNumerically("~", 0.5, 1e-9))
		})
	})

	Describe("Fraction", func() {

		It("returns the fraction of the base node", func() {
			base := slideNode()
			base.SetFraction(0.25)
			node := animation.NewAddNode(base, liftNode())
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})
	})

	Describe("SetFraction", func() {

		It("relocates the base node", func() {
			base := slideNode()
			node := animation.NewAddNode(base, liftNode())
			node.SetFraction(0.5)
			Expect(base.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
		})

		It("relocates a synchronized additive node to the fraction of the base", func() {
			additive := animation.Synchronized(liftNode())
			node := animation.NewAddNode(slideNode(), additive)
			node.SetFraction(1.25) // wrapped by the looping base
			Expect(additive.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("leaves a non-synchronized additive node in place", func() {
			additive := liftNode()
			node := animation.NewAddNode(slideNode(), additive)
			node.SetFraction(0.5)
			Expect(additive.Fraction()).To(Equal(0.0))
		})
	})

	Describe("Advance", func() {

		It("advances the base node", func() {
			base := slideNode()
			node := animation.NewAddNode(base, liftNode())
			node.Advance(0.5)
			Expect(base.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("advances a non-synchronized additive node in real time", func() {
			additive := liftNode()
			node := animation.NewAddNode(slideNode(), additive)
			node.Advance(0.5)
			Expect(additive.Fraction()).To(BeNumerically("~", 0.125, 1e-9))
		})

		It("keeps a synchronized additive node at the fraction of the base", func() {
			additive := animation.Synchronized(liftNode())
			node := animation.NewAddNode(slideNode(), additive)
			node.Advance(0.5)
			Expect(additive.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			expectVec3(node.Transform(animatedBone).Translation, dprec.NewVec3(0.5, 1.0, 0.0))
		})
	})

	Describe("Transform", func() {

		var node *animation.AddNode

		BeforeEach(func() {
			node = animation.NewAddNode(slideNode(), offsetNode())
			node.SetFraction(0.5) // the base is at X = 1.0
		})

		DescribeTable("applies the additive layer according to the weight",
			func(weight float64, expected animation.Transform) {
				node.SetWeight(weight)
				expectTransform(node.Transform(animatedBone), expected)
			},
			Entry("in full", 1.0, animation.Transform{
				Translation: dprec.NewVec3(1.0, 1.0, 0.0),
				Rotation:    yRotation(90.0),
				Scale:       dprec.NewVec3(2.0, 2.0, 2.0),
			}),
			Entry("not at all", 0.0, animation.Transform{
				Translation: dprec.NewVec3(1.0, 0.0, 0.0),
				Rotation:    dprec.IdentityQuat(),
				Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
			}),
			Entry("by half", 0.5, animation.Transform{
				Translation: dprec.NewVec3(1.0, 0.5, 0.0),
				Rotation:    yRotation(45.0),
				Scale:       dprec.NewVec3(1.5, 1.5, 1.5),
			}),
			Entry("extrapolated", 2.0, animation.Transform{
				Translation: dprec.NewVec3(1.0, 2.0, 0.0),
				Rotation:    yRotation(doubleDegrees),
				Scale:       dprec.NewVec3(3.0, 3.0, 3.0),
			}),
		)

		It("applies the additive layer on top of the base rotation", func() {
			base := animation.NewFixedPoseNode(skeleton, map[string]animation.Transform{
				"animated": {
					Translation: dprec.ZeroVec3(),
					Rotation:    rotationAround(90.0, dprec.BasisXVec3()),
					Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
				},
			})
			node := animation.NewAddNode(base, offsetNode())
			expected := dprec.QuatProd(yRotation(90.0), rotationAround(90.0, dprec.BasisXVec3()))
			expectRotation(node.Transform(animatedBone).Rotation, expected)
		})

		It("leaves bones without an additive offset unchanged", func() {
			expectTransform(node.Transform(spareBone), animation.IdentityTransform())
		})
	})

	Describe("DeltaTransform", func() {

		It("returns the motion of the base node only", func() {
			additive := animation.Synchronized(liftNode())
			node := animation.NewAddNode(slideNode(), additive)
			delta := node.DeltaTransform(animatedBone, 0.5)
			expectTransform(delta, animation.Transform{
				Translation: dprec.NewVec3(0.5, 0.0, 0.0),
				Rotation:    dprec.IdentityQuat(),
				Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
			})
		})

		It("does not change the playback position", func() {
			additive := liftNode()
			node := animation.NewAddNode(slideNode(), additive)
			node.SetFraction(0.25)
			node.DeltaTransform(animatedBone, 1.0)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(additive.Fraction()).To(Equal(0.0))
		})
	})

	Describe("SetSynchronized", func() {

		It("configures the node without affecting its children", func() {
			base := slideNode()
			additive := liftNode()
			node := animation.NewAddNode(base, additive)

			node.SetSynchronized(true)
			Expect(node.IsSynchronized()).To(BeTrue())
			Expect(base.IsSynchronized()).To(BeFalse())
			Expect(additive.IsSynchronized()).To(BeFalse())

			node.SetSynchronized(false)
			Expect(node.IsSynchronized()).To(BeFalse())
		})
	})
})
