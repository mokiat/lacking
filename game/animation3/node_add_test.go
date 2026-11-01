package animation3_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("AddNode", func() {

	const (
		animatedBone = 0
		spareBone    = 1
	)

	var (
		skeleton  *animation3.Skeleton
		slideClip *animation3.Clip
		liftClip  *animation3.Clip
		offset    animation3.Transform
	)

	// translationClip creates a clip that moves the animated bone linearly
	// from the zero translation to the specified one over the specified
	// length, without rotating or scaling it.
	translationClip := func(name string, length float64, end dprec.Vec3) *animation3.Clip {
		return animation3.NewClip(name, 0.0, length, map[string]animation3.Channel{
			"animated": animation3.NewComponentKeyframeChannel(
				animation3.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.ZeroVec3()},
					{Timestamp: length, Value: end},
				},
				animation3.KeyframeList[dprec.Quat]{
					{Timestamp: 0.0, Value: dprec.IdentityQuat()},
				},
				animation3.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.NewVec3(1.0, 1.0, 1.0)},
				},
			),
		})
	}

	// slideNode creates a looping base node whose animated bone slides along
	// X from 0.0 to 2.0 over two seconds, so the X translation of the bone
	// equals twice the fraction of the node.
	slideNode := func() *animation3.ClipNode {
		return animation3.NewClipNode(skeleton, slideClip, true)
	}

	// liftNode creates a looping additive node whose animated bone lifts
	// along Y from 0.0 to 4.0 over four seconds, so the Y translation of the
	// bone equals four times the fraction of the node.
	liftNode := func() *animation3.ClipNode {
		return animation3.NewClipNode(skeleton, liftClip, true)
	}

	// offsetNode creates an additive node that holds the offset transform on
	// the animated bone and the additive neutral on the spare bone.
	offsetNode := func() *animation3.FixedPoseNode {
		return animation3.NewFixedPoseNode(skeleton, map[string]animation3.Transform{
			"animated": offset,
		})
	}

	expectTransform := func(actual, expected animation3.Transform) {
		GinkgoHelper()
		expectVec3(actual.Translation, expected.Translation)
		expectRotation(actual.Rotation, expected.Rotation)
		expectVec3(actual.Scale, expected.Scale)
	}

	BeforeEach(func() {
		// Identity rest transforms make bones that a node does not animate
		// valid as both absolute and additive output.
		skeleton = animation3.NewSkeleton([]animation3.SkeletonBone{
			{Name: "animated", RestTransform: animation3.IdentityTransform()},
			{Name: "spare", RestTransform: animation3.IdentityTransform()},
		})
		slideClip = translationClip("slide", 2.0, dprec.NewVec3(2.0, 0.0, 0.0))
		liftClip = translationClip("lift", 4.0, dprec.NewVec3(0.0, 4.0, 0.0))
		offset = animation3.Transform{
			Translation: dprec.NewVec3(0.0, 1.0, 0.0),
			Rotation:    yRotation(90.0),
			Scale:       dprec.NewVec3(2.0, 2.0, 2.0),
		}
	})

	Describe("NewAddNode", func() {

		It("starts with a weight of 1.0", func() {
			node := animation3.NewAddNode(slideNode(), offsetNode())
			Expect(node.Weight()).To(Equal(1.0))
		})

		It("starts out as not synchronized", func() {
			node := animation3.NewAddNode(slideNode(), offsetNode())
			Expect(node.IsSynchronized()).To(BeFalse())
		})
	})

	Describe("Add", func() {

		It("combines the nodes like NewAddNode", func() {
			node := animation3.Add(slideNode(), offsetNode())
			Expect(node.Weight()).To(Equal(1.0))
			expectVec3(node.Transform(animatedBone).Translation, dprec.NewVec3(0.0, 1.0, 0.0))
		})
	})

	Describe("SetWeight", func() {

		It("changes the weight", func() {
			node := animation3.NewAddNode(slideNode(), offsetNode())
			node.SetWeight(0.25)
			Expect(node.Weight()).To(Equal(0.25))
		})
	})

	Describe("Rate", func() {

		It("returns the rate of the base node", func() {
			node := animation3.NewAddNode(slideNode(), liftNode())
			Expect(node.Rate()).To(BeNumerically("~", 0.5, 1e-9))
		})
	})

	Describe("Fraction", func() {

		It("returns the fraction of the base node", func() {
			base := slideNode()
			base.SetFraction(0.25)
			node := animation3.NewAddNode(base, liftNode())
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})
	})

	Describe("SetFraction", func() {

		It("relocates the base node", func() {
			base := slideNode()
			node := animation3.NewAddNode(base, liftNode())
			node.SetFraction(0.5)
			Expect(base.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
		})

		It("relocates a synchronized additive node to the fraction of the base", func() {
			additive := animation3.Synchronized(liftNode())
			node := animation3.NewAddNode(slideNode(), additive)
			node.SetFraction(1.25) // wrapped by the looping base
			Expect(additive.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("leaves a non-synchronized additive node in place", func() {
			additive := liftNode()
			node := animation3.NewAddNode(slideNode(), additive)
			node.SetFraction(0.5)
			Expect(additive.Fraction()).To(Equal(0.0))
		})
	})

	Describe("Advance", func() {

		It("advances the base node", func() {
			base := slideNode()
			node := animation3.NewAddNode(base, liftNode())
			node.Advance(0.5)
			Expect(base.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("advances a non-synchronized additive node in real time", func() {
			additive := liftNode()
			node := animation3.NewAddNode(slideNode(), additive)
			node.Advance(0.5)
			Expect(additive.Fraction()).To(BeNumerically("~", 0.125, 1e-9))
		})

		It("keeps a synchronized additive node at the fraction of the base", func() {
			additive := animation3.Synchronized(liftNode())
			node := animation3.NewAddNode(slideNode(), additive)
			node.Advance(0.5)
			Expect(additive.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			expectVec3(node.Transform(animatedBone).Translation, dprec.NewVec3(0.5, 1.0, 0.0))
		})
	})

	Describe("Transform", func() {

		var node *animation3.AddNode

		BeforeEach(func() {
			node = animation3.NewAddNode(slideNode(), offsetNode())
			node.SetFraction(0.5) // the base is at X = 1.0
		})

		DescribeTable("applies the additive layer according to the weight",
			func(weight float64, expected animation3.Transform) {
				node.SetWeight(weight)
				expectTransform(node.Transform(animatedBone), expected)
			},
			Entry("in full", 1.0, animation3.Transform{
				Translation: dprec.NewVec3(1.0, 1.0, 0.0),
				Rotation:    yRotation(90.0),
				Scale:       dprec.NewVec3(2.0, 2.0, 2.0),
			}),
			Entry("not at all", 0.0, animation3.Transform{
				Translation: dprec.NewVec3(1.0, 0.0, 0.0),
				Rotation:    dprec.IdentityQuat(),
				Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
			}),
			Entry("by half", 0.5, animation3.Transform{
				Translation: dprec.NewVec3(1.0, 0.5, 0.0),
				Rotation:    yRotation(45.0),
				Scale:       dprec.NewVec3(1.5, 1.5, 1.5),
			}),
			Entry("extrapolated", 2.0, animation3.Transform{
				Translation: dprec.NewVec3(1.0, 2.0, 0.0),
				Rotation:    yRotation(doubleDegrees),
				Scale:       dprec.NewVec3(3.0, 3.0, 3.0),
			}),
		)

		It("applies the additive layer on top of the base rotation", func() {
			base := animation3.NewFixedPoseNode(skeleton, map[string]animation3.Transform{
				"animated": {
					Translation: dprec.ZeroVec3(),
					Rotation:    rotationAround(90.0, dprec.BasisXVec3()),
					Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
				},
			})
			node := animation3.NewAddNode(base, offsetNode())
			expected := dprec.QuatProd(yRotation(90.0), rotationAround(90.0, dprec.BasisXVec3()))
			expectRotation(node.Transform(animatedBone).Rotation, expected)
		})

		It("leaves bones without an additive offset unchanged", func() {
			expectTransform(node.Transform(spareBone), animation3.IdentityTransform())
		})
	})

	Describe("DeltaTransform", func() {

		It("returns the motion of the base node only", func() {
			additive := animation3.Synchronized(liftNode())
			node := animation3.NewAddNode(slideNode(), additive)
			delta := node.DeltaTransform(animatedBone, 0.5)
			expectTransform(delta, animation3.Transform{
				Translation: dprec.NewVec3(0.5, 0.0, 0.0),
				Rotation:    dprec.IdentityQuat(),
				Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
			})
		})

		It("does not change the playback position", func() {
			additive := liftNode()
			node := animation3.NewAddNode(slideNode(), additive)
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
			node := animation3.NewAddNode(base, additive)

			node.SetSynchronized(true)
			Expect(node.IsSynchronized()).To(BeTrue())
			Expect(base.IsSynchronized()).To(BeFalse())
			Expect(additive.IsSynchronized()).To(BeFalse())

			node.SetSynchronized(false)
			Expect(node.IsSynchronized()).To(BeFalse())
		})
	})
})
