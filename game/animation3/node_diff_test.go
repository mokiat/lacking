package animation3_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("DiffNode", func() {

	const (
		animatedBone = 0
		spareBone    = 1
	)

	var (
		skeleton  *animation3.Skeleton
		spareRest animation3.Transform
		slideClip *animation3.Clip
		liftClip  *animation3.Clip
	)

	// translationClip creates a clip that moves the animated bone linearly
	// from the start translation to the end one over the specified length,
	// without rotating or scaling it.
	translationClip := func(name string, length float64, start, end dprec.Vec3) *animation3.Clip {
		return animation3.NewClip(name, 0.0, length, map[string]animation3.Channel{
			"animated": animation3.NewComponentKeyframeChannel(
				animation3.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: start},
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

	// slideNode creates a looping source node whose animated bone slides
	// along X from 1.0 to 3.0 over two seconds.
	slideNode := func() *animation3.ClipNode {
		return animation3.NewClipNode(skeleton, slideClip, true)
	}

	// liftNode creates a looping reference node whose animated bone lifts
	// along Y from 0.0 to 4.0 over four seconds, so the Y translation of the
	// bone equals four times the fraction of the node.
	liftNode := func() *animation3.ClipNode {
		return animation3.NewClipNode(skeleton, liftClip, true)
	}

	// fixedNode creates a static node that holds the specified transform on
	// the animated bone.
	fixedNode := func(transform animation3.Transform) *animation3.FixedPoseNode {
		return animation3.NewFixedPoseNode(skeleton, map[string]animation3.Transform{
			"animated": transform,
		})
	}

	expectTransform := func(actual, expected animation3.Transform) {
		GinkgoHelper()
		expectVec3(actual.Translation, expected.Translation)
		expectRotation(actual.Rotation, expected.Rotation)
		expectVec3(actual.Scale, expected.Scale)
	}

	BeforeEach(func() {
		spareRest = animation3.Transform{
			Translation: dprec.NewVec3(0.0, 5.0, 0.0),
			Rotation:    rotationAround(30.0, dprec.BasisZVec3()),
			Scale:       dprec.NewVec3(2.0, 2.0, 2.0),
		}
		skeleton = animation3.NewSkeleton([]animation3.SkeletonBone{
			{Name: "animated", RestTransform: animation3.IdentityTransform()},
			{Name: "spare", RestTransform: spareRest},
		})
		slideClip = translationClip("slide", 2.0, dprec.NewVec3(1.0, 0.0, 0.0), dprec.NewVec3(3.0, 0.0, 0.0))
		liftClip = translationClip("lift", 4.0, dprec.ZeroVec3(), dprec.NewVec3(0.0, 4.0, 0.0))
	})

	Describe("NewDiffNode", func() {

		It("starts out as not synchronized", func() {
			node := animation3.NewDiffNode(slideNode(), liftNode())
			Expect(node.IsSynchronized()).To(BeFalse())
		})
	})

	Describe("Diff", func() {

		It("combines the nodes like NewDiffNode", func() {
			source := slideNode()
			node := animation3.Diff(source, animation3.NewClipPoseNode(skeleton, slideClip, 0.0))
			source.SetFraction(0.5)
			expectVec3(node.Transform(animatedBone).Translation, dprec.NewVec3(1.0, 0.0, 0.0))
		})
	})

	Describe("Rate", func() {

		It("returns the rate of the source node", func() {
			node := animation3.NewDiffNode(slideNode(), liftNode())
			Expect(node.Rate()).To(BeNumerically("~", 0.5, 1e-9))
		})
	})

	Describe("Fraction", func() {

		It("returns the fraction of the source node", func() {
			source := slideNode()
			source.SetFraction(0.25)
			node := animation3.NewDiffNode(source, liftNode())
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})
	})

	Describe("SetFraction", func() {

		It("relocates the source node", func() {
			source := slideNode()
			node := animation3.NewDiffNode(source, liftNode())
			node.SetFraction(0.5)
			Expect(source.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
		})

		It("relocates a synchronized reference node to the fraction of the source", func() {
			reference := animation3.Synchronized(liftNode())
			node := animation3.NewDiffNode(slideNode(), reference)
			node.SetFraction(1.25) // wrapped by the looping source
			Expect(reference.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("leaves a non-synchronized reference node in place", func() {
			reference := liftNode()
			node := animation3.NewDiffNode(slideNode(), reference)
			node.SetFraction(0.5)
			Expect(reference.Fraction()).To(Equal(0.0))
		})
	})

	Describe("Advance", func() {

		It("advances the source node", func() {
			source := slideNode()
			node := animation3.NewDiffNode(source, liftNode())
			node.Advance(0.5)
			Expect(source.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("advances a non-synchronized reference node in real time", func() {
			reference := liftNode()
			node := animation3.NewDiffNode(slideNode(), reference)
			node.Advance(0.5)
			Expect(reference.Fraction()).To(BeNumerically("~", 0.125, 1e-9))
		})

		It("keeps a synchronized reference node at the fraction of the source", func() {
			reference := animation3.Synchronized(liftNode())
			node := animation3.NewDiffNode(slideNode(), reference)
			node.Advance(0.5)
			Expect(reference.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			// The source is at X = 1.5 and the reference at Y = 1.0.
			expectVec3(node.Transform(animatedBone).Translation, dprec.NewVec3(1.5, -1.0, 0.0))
		})
	})

	Describe("Transform", func() {

		It("returns the difference of the source relative to the reference", func() {
			source := fixedNode(animation3.Transform{
				Translation: dprec.NewVec3(1.0, 2.0, 3.0),
				Rotation:    yRotation(90.0),
				Scale:       dprec.NewVec3(4.0, 6.0, 8.0),
			})
			reference := fixedNode(animation3.Transform{
				Translation: dprec.NewVec3(1.0, 1.0, 1.0),
				Rotation:    yRotation(30.0),
				Scale:       dprec.NewVec3(2.0, 2.0, 2.0),
			})
			node := animation3.NewDiffNode(source, reference)
			expectTransform(node.Transform(animatedBone), animation3.Transform{
				Translation: dprec.NewVec3(0.0, 1.0, 2.0),
				Rotation:    yRotation(60.0),
				Scale:       dprec.NewVec3(2.0, 3.0, 4.0),
			})
		})

		It("isolates the motion of a clip relative to its first frame", func() {
			source := slideNode()
			node := animation3.NewDiffNode(source, animation3.NewClipPoseNode(skeleton, slideClip, 0.0))
			Expect(node.Transform(animatedBone).Translation.X).To(BeNumerically("~", 0.0, 1e-9))
			node.Advance(0.5)
			expectVec3(node.Transform(animatedBone).Translation, dprec.NewVec3(0.5, 0.0, 0.0))
		})

		It("returns the identity transform for a bone that neither node animates", func() {
			node := animation3.NewDiffNode(slideNode(), liftNode())
			expectTransform(node.Transform(spareBone), animation3.IdentityTransform())
		})

		It("reproduces the source when added on top of the reference", func() {
			source := fixedNode(animation3.Transform{
				Translation: dprec.NewVec3(1.0, 2.0, 3.0),
				Rotation:    rotationAround(70.0, dprec.UnitVec3(dprec.NewVec3(1.0, 1.0, 0.0))),
				Scale:       dprec.NewVec3(4.0, 6.0, 8.0),
			})
			reference := fixedNode(animation3.Transform{
				Translation: dprec.NewVec3(-1.0, 0.5, 1.0),
				Rotation:    rotationAround(40.0, dprec.BasisXVec3()),
				Scale:       dprec.NewVec3(2.0, 0.5, 2.0),
			})
			node := animation3.NewAddNode(reference, animation3.NewDiffNode(source, reference))
			expectTransform(node.Transform(animatedBone), source.Transform(animatedBone))
			expectTransform(node.Transform(spareBone), spareRest)
		})
	})

	Describe("DeltaTransform", func() {

		It("returns the identity transform", func() {
			node := animation3.NewDiffNode(slideNode(), liftNode())
			expectTransform(node.DeltaTransform(animatedBone, 0.5), animation3.IdentityTransform())
			expectTransform(node.DeltaTransform(spareBone, -0.5), animation3.IdentityTransform())
		})

		It("does not change the playback position", func() {
			reference := liftNode()
			node := animation3.NewDiffNode(slideNode(), reference)
			node.SetFraction(0.25)
			node.DeltaTransform(animatedBone, 1.0)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(reference.Fraction()).To(Equal(0.0))
		})
	})

	Describe("SetSynchronized", func() {

		It("configures the node without affecting its children", func() {
			source := slideNode()
			reference := liftNode()
			node := animation3.NewDiffNode(source, reference)

			node.SetSynchronized(true)
			Expect(node.IsSynchronized()).To(BeTrue())
			Expect(source.IsSynchronized()).To(BeFalse())
			Expect(reference.IsSynchronized()).To(BeFalse())

			node.SetSynchronized(false)
			Expect(node.IsSynchronized()).To(BeFalse())
		})
	})
})
