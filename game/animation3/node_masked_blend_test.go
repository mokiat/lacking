package animation3_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("MaskedBlendNode", func() {

	const (
		rootBone  = 0
		spineBone = 1
		armBone   = 2
	)

	var (
		skeleton      *animation3.Skeleton
		basePose      animation3.Transform
		layerPose     animation3.Transform
		halfwayPose   animation3.Transform
		boneWeights   map[string]float64
		slideClip     *animation3.Clip
		liftClip      *animation3.Clip
		basePoseNode  *animation3.FixedPoseNode
		layerPoseNode *animation3.FixedPoseNode
	)

	// translationClip creates a clip that moves the root bone linearly from
	// the zero translation to the specified one over the specified length,
	// without rotating or scaling it.
	translationClip := func(name string, length float64, end dprec.Vec3) *animation3.Clip {
		return animation3.NewClip(name, 0.0, length, map[string]animation3.Channel{
			"root": animation3.NewComponentKeyframeChannel(
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

	// slideNode creates a looping base node whose root bone slides along X
	// from 0.0 to 2.0 over two seconds.
	slideNode := func() *animation3.ClipNode {
		return animation3.NewClipNode(skeleton, slideClip, true)
	}

	// liftNode creates a looping layer node whose root bone lifts along Y
	// from 0.0 to 4.0 over four seconds.
	liftNode := func() *animation3.ClipNode {
		return animation3.NewClipNode(skeleton, liftClip, true)
	}

	// poseNode creates a static node that holds the specified transform on
	// all bones.
	poseNode := func(transform animation3.Transform) *animation3.FixedPoseNode {
		return animation3.NewFixedPoseNode(skeleton, map[string]animation3.Transform{
			"root":  transform,
			"spine": transform,
			"arm":   transform,
		})
	}

	expectTransform := func(actual, expected animation3.Transform) {
		GinkgoHelper()
		expectVec3(actual.Translation, expected.Translation)
		expectRotation(actual.Rotation, expected.Rotation)
		expectVec3(actual.Scale, expected.Scale)
	}

	BeforeEach(func() {
		skeleton = animation3.NewSkeleton([]animation3.SkeletonBone{
			{Name: "root", RestTransform: animation3.IdentityTransform()},
			{Name: "spine", RestTransform: animation3.IdentityTransform()},
			{Name: "arm", RestTransform: animation3.IdentityTransform()},
		})
		basePose = animation3.Transform{
			Translation: dprec.NewVec3(1.0, 0.0, 0.0),
			Rotation:    dprec.IdentityQuat(),
			Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
		}
		layerPose = animation3.Transform{
			Translation: dprec.NewVec3(0.0, 2.0, 0.0),
			Rotation:    yRotation(90.0),
			Scale:       dprec.NewVec3(3.0, 3.0, 3.0),
		}
		halfwayPose = animation3.Transform{
			Translation: dprec.NewVec3(0.5, 1.0, 0.0),
			Rotation:    yRotation(45.0),
			Scale:       dprec.NewVec3(2.0, 2.0, 2.0),
		}
		boneWeights = map[string]float64{
			"spine": 1.0,
			"arm":   0.5,
		}
		slideClip = translationClip("slide", 2.0, dprec.NewVec3(2.0, 0.0, 0.0))
		liftClip = translationClip("lift", 4.0, dprec.NewVec3(0.0, 4.0, 0.0))
		basePoseNode = poseNode(basePose)
		layerPoseNode = poseNode(layerPose)
	})

	Describe("NewMaskedBlendNode", func() {

		It("starts with an overall weight of 1.0", func() {
			node := animation3.NewMaskedBlendNode(skeleton, basePoseNode, layerPoseNode, boneWeights)
			Expect(node.Weight()).To(Equal(1.0))
		})

		It("starts out as not synchronized", func() {
			node := animation3.NewMaskedBlendNode(skeleton, basePoseNode, layerPoseNode, boneWeights)
			Expect(node.IsSynchronized()).To(BeFalse())
		})

		It("ignores weights for bones that are not in the skeleton", func() {
			node := animation3.NewMaskedBlendNode(skeleton, basePoseNode, layerPoseNode, map[string]float64{
				"unknown": 1.0,
			})
			expectTransform(node.Transform(rootBone), basePose)
			expectTransform(node.Transform(spineBone), basePose)
			expectTransform(node.Transform(armBone), basePose)
		})

		It("does not retain the specified map", func() {
			node := animation3.NewMaskedBlendNode(skeleton, basePoseNode, layerPoseNode, boneWeights)
			boneWeights["root"] = 1.0
			boneWeights["spine"] = 0.0
			expectTransform(node.Transform(rootBone), basePose)
			expectTransform(node.Transform(spineBone), layerPose)
		})
	})

	Describe("SetWeight", func() {

		It("changes the overall weight", func() {
			node := animation3.NewMaskedBlendNode(skeleton, basePoseNode, layerPoseNode, boneWeights)
			node.SetWeight(0.25)
			Expect(node.Weight()).To(Equal(0.25))
		})
	})

	Describe("Rate", func() {

		It("returns the rate of the base node", func() {
			node := animation3.NewMaskedBlendNode(skeleton, slideNode(), liftNode(), boneWeights)
			Expect(node.Rate()).To(BeNumerically("~", 0.5, 1e-9))
		})
	})

	Describe("Fraction", func() {

		It("returns the fraction of the base node", func() {
			base := slideNode()
			base.SetFraction(0.25)
			node := animation3.NewMaskedBlendNode(skeleton, base, liftNode(), boneWeights)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})
	})

	Describe("SetFraction", func() {

		It("relocates the base node", func() {
			base := slideNode()
			node := animation3.NewMaskedBlendNode(skeleton, base, liftNode(), boneWeights)
			node.SetFraction(0.5)
			Expect(base.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
		})

		It("relocates a synchronized layer node to the fraction of the base", func() {
			layer := animation3.Synchronized(liftNode())
			node := animation3.NewMaskedBlendNode(skeleton, slideNode(), layer, boneWeights)
			node.SetFraction(1.25) // wrapped by the looping base
			Expect(layer.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("leaves a non-synchronized layer node in place", func() {
			layer := liftNode()
			node := animation3.NewMaskedBlendNode(skeleton, slideNode(), layer, boneWeights)
			node.SetFraction(0.5)
			Expect(layer.Fraction()).To(Equal(0.0))
		})
	})

	Describe("Advance", func() {

		It("advances the base node", func() {
			base := slideNode()
			node := animation3.NewMaskedBlendNode(skeleton, base, liftNode(), boneWeights)
			node.Advance(0.5)
			Expect(base.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("advances a non-synchronized layer node in real time", func() {
			layer := liftNode()
			node := animation3.NewMaskedBlendNode(skeleton, slideNode(), layer, boneWeights)
			node.Advance(0.5)
			Expect(layer.Fraction()).To(BeNumerically("~", 0.125, 1e-9))
		})

		It("keeps a synchronized layer node at the fraction of the base", func() {
			layer := animation3.Synchronized(liftNode())
			node := animation3.NewMaskedBlendNode(skeleton, slideNode(), layer, map[string]float64{
				"root": 1.0,
			})
			node.Advance(0.5)
			Expect(layer.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			expectVec3(node.Transform(rootBone).Translation, dprec.NewVec3(0.0, 1.0, 0.0))
		})
	})

	Describe("Transform", func() {

		var node *animation3.MaskedBlendNode

		BeforeEach(func() {
			node = animation3.NewMaskedBlendNode(skeleton, basePoseNode, layerPoseNode, boneWeights)
		})

		It("shows the base on a bone that is not in the mask", func() {
			expectTransform(node.Transform(rootBone), basePose)
		})

		It("shows the layer on a bone with a full weight", func() {
			expectTransform(node.Transform(spineBone), layerPose)
		})

		It("blends the two on a bone with a partial weight", func() {
			expectTransform(node.Transform(armBone), halfwayPose)
		})

		It("scales the bone weights by the overall weight", func() {
			node.SetWeight(0.5)
			expectTransform(node.Transform(rootBone), basePose)
			expectTransform(node.Transform(spineBone), halfwayPose)
			expectTransform(node.Transform(armBone), animation3.Transform{
				Translation: dprec.NewVec3(0.75, 0.5, 0.0),
				Rotation:    yRotation(quarterDegrees),
				Scale:       dprec.NewVec3(1.5, 1.5, 1.5),
			})
		})

		It("shows the base on all bones for an overall weight of 0.0", func() {
			node.SetWeight(0.0)
			expectTransform(node.Transform(rootBone), basePose)
			expectTransform(node.Transform(spineBone), basePose)
			expectTransform(node.Transform(armBone), basePose)
		})

		It("extrapolates past a full weight", func() {
			node.SetWeight(2.0)
			expectTransform(node.Transform(spineBone), animation3.Transform{
				Translation: dprec.NewVec3(-1.0, 4.0, 0.0),
				Rotation:    yRotation(doubleDegrees),
				Scale:       dprec.NewVec3(5.0, 5.0, 5.0),
			})
			expectTransform(node.Transform(armBone), layerPose)
		})
	})

	Describe("DeltaTransform", func() {

		DescribeTable("blends the motion of the two nodes by the weight of the bone",
			func(weight float64, expected dprec.Vec3) {
				node := animation3.NewMaskedBlendNode(skeleton, slideNode(), liftNode(), map[string]float64{
					"root": weight,
				})
				delta := node.DeltaTransform(rootBone, 0.5)
				expectTransform(delta, animation3.Transform{
					Translation: expected,
					Rotation:    dprec.IdentityQuat(),
					Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
				})
			},
			Entry("root not in the mask", 0.0, dprec.NewVec3(0.5, 0.0, 0.0)),
			Entry("root fully in the mask", 1.0, dprec.NewVec3(0.0, 0.5, 0.0)),
			Entry("root partially in the mask", 0.5, dprec.NewVec3(0.25, 0.25, 0.0)),
		)

		It("scales the bone weight by the overall weight", func() {
			node := animation3.NewMaskedBlendNode(skeleton, slideNode(), liftNode(), map[string]float64{
				"root": 1.0,
			})
			node.SetWeight(0.5)
			expectVec3(node.DeltaTransform(rootBone, 0.5).Translation, dprec.NewVec3(0.25, 0.25, 0.0))
		})

		It("does not change the playback position", func() {
			layer := liftNode()
			node := animation3.NewMaskedBlendNode(skeleton, slideNode(), layer, boneWeights)
			node.SetFraction(0.25)
			node.DeltaTransform(rootBone, 1.0)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			Expect(layer.Fraction()).To(Equal(0.0))
		})
	})

	Describe("SetSynchronized", func() {

		It("configures the node without affecting its children", func() {
			base := slideNode()
			layer := liftNode()
			node := animation3.NewMaskedBlendNode(skeleton, base, layer, boneWeights)

			node.SetSynchronized(true)
			Expect(node.IsSynchronized()).To(BeTrue())
			Expect(base.IsSynchronized()).To(BeFalse())
			Expect(layer.IsSynchronized()).To(BeFalse())

			node.SetSynchronized(false)
			Expect(node.IsSynchronized()).To(BeFalse())
		})
	})
})
