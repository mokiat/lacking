package animation3_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("RootMotion", func() {

	var (
		skeleton *animation3.Skeleton
		clip     *animation3.Clip
		node     *animation3.ClipNode
	)

	// translationChannel creates a channel that moves a bone linearly from
	// the zero translation to the specified one over two seconds, without
	// rotating or scaling it.
	translationChannel := func(end dprec.Vec3) animation3.Channel {
		return animation3.NewComponentKeyframeChannel(
			animation3.KeyframeList[dprec.Vec3]{
				{Timestamp: 0.0, Value: dprec.ZeroVec3()},
				{Timestamp: 2.0, Value: end},
			},
			animation3.KeyframeList[dprec.Quat]{
				{Timestamp: 0.0, Value: dprec.IdentityQuat()},
			},
			animation3.KeyframeList[dprec.Vec3]{
				{Timestamp: 0.0, Value: dprec.NewVec3(1.0, 1.0, 1.0)},
			},
		)
	}

	expectTransform := func(actual, expected animation3.Transform) {
		GinkgoHelper()
		expectVec3(actual.Translation, expected.Translation)
		expectRotation(actual.Rotation, expected.Rotation)
		expectVec3(actual.Scale, expected.Scale)
	}

	translation := func(x, y, z float64) animation3.Transform {
		return animation3.Transform{
			Translation: dprec.NewVec3(x, y, z),
			Rotation:    dprec.IdentityQuat(),
			Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
		}
	}

	BeforeEach(func() {
		skeleton = animation3.NewSkeleton([]animation3.SkeletonBone{
			{Name: "root", RestTransform: animation3.IdentityTransform()},
			{Name: "hips", RestTransform: animation3.IdentityTransform()},
		})
		// The root slides along X and the hips lift along Y, both at one
		// unit per second.
		clip = animation3.NewClip("walk", 0.0, 2.0, map[string]animation3.Channel{
			"root": translationChannel(dprec.NewVec3(2.0, 0.0, 0.0)),
			"hips": translationChannel(dprec.NewVec3(0.0, 2.0, 0.0)),
		})
		node = animation3.NewClipNode(skeleton, clip, true)
	})

	Describe("DeltaTransform", func() {

		It("returns the upcoming motion of the named bone", func() {
			motion := animation3.NewRootMotion(skeleton, node, "root")
			expectTransform(motion.DeltaTransform(0.5), translation(0.5, 0.0, 0.0))
		})

		It("resolves the bone by name", func() {
			motion := animation3.NewRootMotion(skeleton, node, "hips")
			expectTransform(motion.DeltaTransform(0.5), translation(0.0, 0.5, 0.0))
		})

		It("measures the motion backwards for a negative delta", func() {
			node.SetFraction(0.5)
			motion := animation3.NewRootMotion(skeleton, node, "root")
			expectTransform(motion.DeltaTransform(-0.5), translation(-0.5, 0.0, 0.0))
		})

		It("does not change the playback position", func() {
			node.SetFraction(0.25)
			motion := animation3.NewRootMotion(skeleton, node, "root")
			motion.DeltaTransform(1.0)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("follows the playback position of the node", func() {
			motion := animation3.NewRootMotion(skeleton, node, "root")
			node.Advance(1.5)
			// The motion continues across the wrap around of the loop.
			expectTransform(motion.DeltaTransform(1.0), translation(1.0, 0.0, 0.0))
		})

		It("reads the motion through the node tree", func() {
			tree := animation3.ScaleTime(node, 2.0)
			motion := animation3.NewRootMotion(skeleton, tree, "root")
			expectTransform(motion.DeltaTransform(0.5), translation(1.0, 0.0, 0.0))
		})

		It("returns the identity transform for an unknown bone", func() {
			motion := animation3.NewRootMotion(skeleton, node, "unknown")
			expectTransform(motion.DeltaTransform(0.5), animation3.IdentityTransform())
		})
	})
})
