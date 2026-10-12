package animation3_test

import (
	"slices"

	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("Clip", func() {

	var (
		hipChannel  animation3.ClipChannel
		headChannel animation3.ClipChannel
		clip        *animation3.Clip
	)

	BeforeEach(func() {
		hipChannel = animation3.ClipChannel{
			TranslationKeyframes: animation3.KeyframeList[dprec.Vec3]{
				{Timestamp: 0.0, Value: dprec.NewVec3(0.0, 1.0, 0.0)},
				{Timestamp: 4.0, Value: dprec.NewVec3(2.0, 1.0, 0.0)},
			},
			RotationKeyframes: animation3.KeyframeList[dprec.Quat]{
				{Timestamp: 1.0, Value: dprec.IdentityQuat()},
			},
		}
		headChannel = animation3.ClipChannel{
			ScaleKeyframes: animation3.KeyframeList[dprec.Vec3]{
				{Timestamp: 0.5, Value: dprec.NewVec3(1.0, 1.0, 1.0)},
				{Timestamp: 3.5, Value: dprec.NewVec3(2.0, 2.0, 2.0)},
			},
		}
		clip = animation3.NewClip("walk", 1.0, 3.0, map[string]animation3.ClipChannel{
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
			empty := animation3.NewClip("empty", 0.0, 1.0, nil)
			Expect(empty.HasChannel("hip")).To(BeFalse())
			_, ok := empty.Channel("hip")
			Expect(ok).To(BeFalse())
			Expect(slices.Collect(empty.ChannelNamesIter())).To(BeEmpty())
		})
	})

	Describe("Length", func() {

		DescribeTable("returns the non-negative duration of the time range",
			func(startTime, endTime, expected float64) {
				c := animation3.NewClip("clip", startTime, endTime, nil)
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
			Expect(channel).To(Equal(hipChannel))

			channel, ok = clip.Channel("head")
			Expect(ok).To(BeTrue())
			Expect(channel).To(Equal(headChannel))
		})

		It("reports a missing channel", func() {
			channel, ok := clip.Channel("tail")
			Expect(ok).To(BeFalse())
			Expect(channel).To(Equal(animation3.ClipChannel{}))
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
		var cropped *animation3.Clip

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
			Expect(channel).To(Equal(hipChannel))
		})

		It("does not trim keyframes outside the new range", func() {
			channel, ok := cropped.Channel("head")
			Expect(ok).To(BeTrue())
			Expect(channel.ScaleKeyframes).To(HaveLen(2))
			Expect(channel.ScaleKeyframes[0].Timestamp).To(Equal(0.5))
			Expect(channel.ScaleKeyframes[1].Timestamp).To(Equal(3.5))
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
