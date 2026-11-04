package animation_test

import (
	"slices"

	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation"
)

var _ = Describe("Clip", func() {

	var (
		hipChannel  animation.Channel
		headChannel animation.Channel
		clip        *animation.Clip
	)

	BeforeEach(func() {
		hipChannel = animation.NewComponentKeyframeChannel(
			animation.KeyframeList[dprec.Vec3]{
				{Timestamp: 0.0, Value: dprec.NewVec3(0.0, 1.0, 0.0)},
				{Timestamp: 4.0, Value: dprec.NewVec3(2.0, 1.0, 0.0)},
			},
			animation.KeyframeList[dprec.Quat]{
				{Timestamp: 1.0, Value: dprec.IdentityQuat()},
			},
			animation.KeyframeList[dprec.Vec3]{
				{Timestamp: 1.0, Value: dprec.NewVec3(1.0, 1.0, 1.0)},
			},
		)
		headChannel = animation.NewTransformKeyframeChannel(
			animation.KeyframeList[animation.Transform]{
				{Timestamp: 0.5, Value: animation.IdentityTransform()},
				{Timestamp: 3.5, Value: animation.Transform{
					Translation: dprec.ZeroVec3(),
					Rotation:    dprec.IdentityQuat(),
					Scale:       dprec.NewVec3(2.0, 2.0, 2.0),
				}},
			},
		)
		clip = animation.NewClip("walk", 1.0, 3.0, map[string]animation.Channel{
			"hip":  hipChannel,
			"head": headChannel,
		})
	})

	Describe("NewClip", func() {

		It("stores the name", func() {
			Expect(clip.Name()).To(Equal("walk"))
		})

		It("stores the time range", func() {
			Expect(clip.StartTime()).To(Equal(1.0))
			Expect(clip.EndTime()).To(Equal(3.0))
		})

		It("accepts a nil channels map", func() {
			empty := animation.NewClip("empty", 0.0, 1.0, nil)
			Expect(empty.HasChannel("hip")).To(BeFalse())
			_, ok := empty.Channel("hip")
			Expect(ok).To(BeFalse())
			Expect(slices.Collect(empty.ChannelNamesIter())).To(BeEmpty())
		})
	})

	Describe("Length", func() {

		DescribeTable("returns the non-negative duration of the time range",
			func(startTime, endTime, expected float64) {
				c := animation.NewClip("clip", startTime, endTime, nil)
				Expect(c.Length()).To(BeNumerically("~", expected, 1e-9))
			},
			Entry("range starting at zero", 0.0, 2.5, 2.5),
			Entry("range with an offset start", 1.0, 3.0, 2.0),
			Entry("range with negative timestamps", -2.0, -0.5, 1.5),
			Entry("empty range", 2.0, 2.0, 0.0),
			Entry("inverted range", 3.0, 1.0, 0.0),
		)
	})

	Describe("Channel", func() {

		It("returns an existing channel", func() {
			channel, ok := clip.Channel("hip")
			Expect(ok).To(BeTrue())
			Expect(channel).To(BeIdenticalTo(hipChannel))

			channel, ok = clip.Channel("head")
			Expect(ok).To(BeTrue())
			Expect(channel).To(BeIdenticalTo(headChannel))
		})

		It("reports a missing channel", func() {
			channel, ok := clip.Channel("tail")
			Expect(ok).To(BeFalse())
			Expect(channel).To(BeNil())
		})
	})

	Describe("HasChannel", func() {

		It("returns true for existing channels", func() {
			Expect(clip.HasChannel("hip")).To(BeTrue())
			Expect(clip.HasChannel("head")).To(BeTrue())
		})

		It("returns false for missing channels", func() {
			Expect(clip.HasChannel("tail")).To(BeFalse())
			Expect(clip.HasChannel("")).To(BeFalse())
		})
	})

	Describe("ChannelNamesIter", func() {

		It("yields the name of every channel", func() {
			Expect(slices.Collect(clip.ChannelNamesIter())).To(
				ConsistOf("hip", "head"),
			)
		})

		It("supports stopping early", func() {
			count := 0
			for range clip.ChannelNamesIter() {
				count++
				break
			}
			Expect(count).To(Equal(1))
		})
	})

	Describe("Crop", func() {
		var cropped *animation.Clip

		BeforeEach(func() {
			cropped = clip.Crop("walk-start", 1.5, 2.0)
		})

		It("returns a clip with the specified name and time range", func() {
			Expect(cropped.Name()).To(Equal("walk-start"))
			Expect(cropped.StartTime()).To(Equal(1.5))
			Expect(cropped.EndTime()).To(Equal(2.0))
			Expect(cropped.Length()).To(BeNumerically("~", 0.5, 1e-9))
		})

		It("keeps the same channels", func() {
			Expect(slices.Collect(cropped.ChannelNamesIter())).To(
				ConsistOf("hip", "head"),
			)
			channel, ok := cropped.Channel("hip")
			Expect(ok).To(BeTrue())
			Expect(channel).To(BeIdenticalTo(hipChannel))
		})

		It("keeps animation data outside the new range", func() {
			channel, ok := cropped.Channel("head")
			Expect(ok).To(BeTrue())
			expectVec3(channel.Transform(0.5).Scale, dprec.NewVec3(1.0, 1.0, 1.0))
			expectVec3(channel.Transform(3.5).Scale, dprec.NewVec3(2.0, 2.0, 2.0))
		})

		It("leaves the original clip unchanged", func() {
			Expect(clip.Name()).To(Equal("walk"))
			Expect(clip.StartTime()).To(Equal(1.0))
			Expect(clip.EndTime()).To(Equal(3.0))
			Expect(slices.Collect(clip.ChannelNamesIter())).To(
				ConsistOf("hip", "head"),
			)
		})
	})

})
