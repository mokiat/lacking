package animation_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation"
)

var _ = Describe("OneShotNode", func() {

	const (
		rootBone = 0
		armBone  = 1
	)

	var (
		skeleton *animation.Skeleton
		base     *animation.FixedPoseNode
		overlay  *animation.ClipNode
		node     *animation.OneShotNode
	)

	// clipNode creates a non-looping clip node of the specified length that
	// moves all bones linearly along X from the specified start to the
	// specified end.
	clipNode := func(length, start, end float64) *animation.ClipNode {
		channel := func() animation.Channel {
			return animation.NewComponentKeyframeChannel(
				animation.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.NewVec3(start, 0.0, 0.0)},
					{Timestamp: length, Value: dprec.NewVec3(end, 0.0, 0.0)},
				},
				animation.KeyframeList[dprec.Quat]{
					{Timestamp: 0.0, Value: dprec.IdentityQuat()},
				},
				animation.KeyframeList[dprec.Vec3]{
					{Timestamp: 0.0, Value: dprec.NewVec3(1.0, 1.0, 1.0)},
				},
			)
		}
		clip := animation.NewClip("clip", 0.0, length, map[string]animation.Channel{
			"root": channel(),
			"arm":  channel(),
		})
		return animation.NewClipNode(skeleton, clip, false)
	}

	expectX := func(bone int, expected float64) {
		GinkgoHelper()
		expectVec3(node.Transform(bone).Translation, dprec.NewVec3(expected, 0.0, 0.0))
	}

	BeforeEach(func() {
		skeleton = animation.NewSkeleton([]animation.SkeletonBone{
			{Name: "root", RestTransform: animation.IdentityTransform()},
			{Name: "arm", RestTransform: animation.IdentityTransform()},
		})
		base = animation.NewFixedPoseNode(skeleton, nil)
		overlay = clipNode(1.0, 1.0, 1.0) // holds X at 1.0 for one second
		node = animation.NewOneShotNode(skeleton, base, overlay, nil)
	})

	Describe("NewOneShotNode", func() {

		It("is not playing", func() {
			Expect(node.IsPlaying()).To(BeFalse())
			expectX(rootBone, 0.0)
		})

		It("has no fades", func() {
			Expect(node.FadeInDuration()).To(Equal(0.0))
			Expect(node.FadeOutDuration()).To(Equal(0.0))
		})

		It("is not synchronized", func() {
			Expect(node.IsSynchronized()).To(BeFalse())
		})

		It("does not retain the specified map", func() {
			boneWeights := map[string]float64{"arm": 1.0}
			node = animation.NewOneShotNode(skeleton, base, overlay, boneWeights)
			boneWeights["root"] = 1.0
			node.Trigger()
			expectX(rootBone, 0.0)
		})
	})

	Describe("fade durations", func() {

		It("stores the durations", func() {
			node.SetFadeInDuration(0.25)
			node.SetFadeOutDuration(0.5)
			Expect(node.FadeInDuration()).To(Equal(0.25))
			Expect(node.FadeOutDuration()).To(Equal(0.5))
		})

		It("treats negative durations as 0.0", func() {
			node.SetFadeInDuration(-1.0)
			node.SetFadeOutDuration(-1.0)
			Expect(node.FadeInDuration()).To(Equal(0.0))
			Expect(node.FadeOutDuration()).To(Equal(0.0))
		})
	})

	Describe("Trigger", func() {

		It("plays the overlay once", func() {
			node.Trigger()
			Expect(node.IsPlaying()).To(BeTrue())
			expectX(rootBone, 1.0)
			node.Advance(0.75)
			Expect(node.IsPlaying()).To(BeTrue())
			expectX(rootBone, 1.0)
			node.Advance(0.25)
			Expect(node.IsPlaying()).To(BeFalse())
			expectX(rootBone, 0.0)
		})

		It("plays the overlay from its beginning", func() {
			node.Trigger()
			node.Advance(1.0)
			Expect(overlay.Fraction()).To(Equal(1.0))
			node.Trigger()
			Expect(overlay.Fraction()).To(Equal(0.0))
		})

		It("has no effect while the overlay is playing", func() {
			node.Trigger()
			node.Advance(0.5)
			node.Trigger()
			Expect(overlay.Fraction()).To(BeNumerically("~", 0.5, 1e-9))
			node.Advance(0.5)
			Expect(node.IsPlaying()).To(BeFalse())
		})

		It("panics for an overlay without a rate", func() {
			node = animation.NewOneShotNode(skeleton, base, animation.NewFixedPoseNode(skeleton, nil), nil)
			Expect(func() {
				node.Trigger()
			}).To(Panic())
		})

		It("plays the overlay for one cycle at its rate", func() {
			node = animation.NewOneShotNode(skeleton, base, animation.ScaleTime(overlay, 2.0), nil)
			node.Trigger()
			node.Advance(0.25)
			Expect(node.IsPlaying()).To(BeTrue())
			node.Advance(0.25)
			Expect(node.IsPlaying()).To(BeFalse())
		})
	})

	Describe("fades", func() {

		It("fades the overlay in", func() {
			node.SetFadeInDuration(0.5)
			node.Trigger()
			expectX(rootBone, 0.0)
			node.Advance(0.25)
			expectX(rootBone, 0.5)
			node.Advance(0.25)
			expectX(rootBone, 1.0)
		})

		It("fades the overlay out before it ends", func() {
			node.SetFadeOutDuration(0.5)
			node.Trigger()
			node.Advance(0.5)
			expectX(rootBone, 1.0)
			node.Advance(0.25)
			expectX(rootBone, 0.5)
		})

		It("meets the two fades when they overlap", func() {
			node.SetFadeInDuration(1.0)
			node.SetFadeOutDuration(1.0)
			node.Trigger()
			node.Advance(0.5)
			expectX(rootBone, 0.5)
			node.Advance(0.25)
			expectX(rootBone, 0.25)
		})
	})

	Describe("Stop", func() {

		BeforeEach(func() {
			overlay = clipNode(4.0, 1.0, 1.0)
			node = animation.NewOneShotNode(skeleton, base, overlay, nil)
			node.SetFadeInDuration(0.5)
			node.SetFadeOutDuration(0.5)
		})

		It("fades the overlay out from its current weight", func() {
			node.Trigger()
			node.Advance(0.5)
			node.Stop()
			Expect(node.IsPlaying()).To(BeTrue())
			expectX(rootBone, 1.0)
			node.Advance(0.25)
			expectX(rootBone, 0.5)
			node.Advance(0.25)
			Expect(node.IsPlaying()).To(BeFalse())
			expectX(rootBone, 0.0)
		})

		It("fades the overlay out from a partial weight", func() {
			node.Trigger()
			node.Advance(0.25) // weight 0.5
			node.Stop()
			expectX(rootBone, 0.5)
			node.Advance(0.125)
			expectX(rootBone, 0.25)
			node.Advance(0.125)
			Expect(node.IsPlaying()).To(BeFalse())
		})

		It("removes the overlay immediately without a fade out", func() {
			node.SetFadeOutDuration(0.0)
			node.Trigger()
			node.Advance(0.5)
			node.Stop()
			Expect(node.IsPlaying()).To(BeFalse())
			expectX(rootBone, 0.0)
		})

		It("has no effect when the overlay is not playing", func() {
			node.Stop()
			Expect(node.IsPlaying()).To(BeFalse())
			node.Trigger()
			Expect(node.IsPlaying()).To(BeTrue())
		})

		It("allows the overlay to be triggered again while it fades out", func() {
			node.Trigger()
			node.Advance(1.0)
			node.Stop()
			node.Advance(0.25) // weight 0.5
			node.Trigger()
			Expect(node.IsPlaying()).To(BeTrue())
			Expect(overlay.Fraction()).To(Equal(0.0))
			expectX(rootBone, 0.5)
			node.Advance(0.125)
			expectX(rootBone, 0.75)
			node.Advance(1.0)
			expectX(rootBone, 1.0)
		})
	})

	Describe("bone weights", func() {

		BeforeEach(func() {
			node = animation.NewOneShotNode(skeleton, base, overlay, map[string]float64{
				"arm": 0.5,
			})
		})

		It("blends the overlay over the masked bones only", func() {
			node.Trigger()
			expectX(rootBone, 0.0)
			expectX(armBone, 0.5)
		})

		It("scales the bone weights by the fade", func() {
			node.SetFadeInDuration(0.5)
			node.Trigger()
			node.Advance(0.25)
			expectX(armBone, 0.25)
		})
	})

	Describe("playback", func() {

		var cycle *animation.ClipNode

		BeforeEach(func() {
			cycle = clipNode(2.0, 0.0, 2.0)
			node = animation.NewOneShotNode(skeleton, cycle, overlay, nil)
		})

		It("takes the rate and fraction of the base node", func() {
			Expect(node.Rate()).To(BeNumerically("~", 0.5, 1e-9))
			node.Advance(0.5)
			Expect(node.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("relocates only the base node", func() {
			node.Trigger()
			node.Advance(0.25)
			node.SetFraction(0.75)
			Expect(cycle.Fraction()).To(BeNumerically("~", 0.75, 1e-9))
			Expect(overlay.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("does not advance the overlay while it is not playing", func() {
			node.Advance(0.5)
			Expect(overlay.Fraction()).To(Equal(0.0))
		})

		It("ignores the synchronization setting of the overlay", func() {
			overlay.SetSynchronized(true)
			node.Advance(1.0) // base at 0.5
			node.Trigger()
			node.Advance(0.25)
			Expect(overlay.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
			node.SetFraction(0.0)
			Expect(overlay.Fraction()).To(BeNumerically("~", 0.25, 1e-9))
		})

		It("configures its own synchronization without affecting its nodes", func() {
			node.SetSynchronized(true)
			Expect(node.IsSynchronized()).To(BeTrue())
			Expect(cycle.IsSynchronized()).To(BeFalse())
			Expect(overlay.IsSynchronized()).To(BeFalse())
		})
	})

	Describe("DeltaTransform", func() {

		BeforeEach(func() {
			overlay = clipNode(1.0, 0.0, 1.0)
			node = animation.NewOneShotNode(skeleton, base, overlay, map[string]float64{
				"root": 1.0,
			})
		})

		It("returns the motion of the base node while the overlay is not playing", func() {
			expectVec3(node.DeltaTransform(rootBone, 0.5).Translation, dprec.ZeroVec3())
		})

		It("blends the motion of the overlay by its weight", func() {
			node.SetFadeInDuration(0.5)
			node.Trigger()
			node.Advance(0.25) // weight 0.5
			expectVec3(node.DeltaTransform(rootBone, 0.5).Translation, dprec.NewVec3(0.25, 0.0, 0.0))
			expectVec3(node.DeltaTransform(armBone, 0.5).Translation, dprec.ZeroVec3())
		})
	})
})
