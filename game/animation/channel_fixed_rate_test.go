package animation_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation"
)

var _ = Describe("FixedRateKeyframeChannel", func() {

	var (
		firstTransform  animation.Transform
		secondTransform animation.Transform
		thirdTransform  animation.Transform
		frames          []animation.Transform
	)

	BeforeEach(func() {
		firstTransform = animation.Transform{
			Translation: dprec.NewVec3(0.0, 0.0, 0.0),
			Rotation:    yRotation(0.0),
			Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
		}
		secondTransform = animation.Transform{
			Translation: dprec.NewVec3(4.0, 0.0, 0.0),
			Rotation:    yRotation(90.0),
			Scale:       dprec.NewVec3(3.0, 3.0, 3.0),
		}
		thirdTransform = animation.Transform{
			Translation: dprec.NewVec3(4.0, 2.0, 0.0),
			Rotation:    yRotation(90.0),
			Scale:       dprec.NewVec3(1.0, 2.0, 3.0),
		}
		// With a start of 1.0 and a rate of 2.0, the frames are located at
		// 1.0, 1.5 and 2.0.
		frames = []animation.Transform{
			firstTransform,
			secondTransform,
			thirdTransform,
		}
	})

	Describe("NewFixedRateKeyframeChannel", func() {

		It("panics without frames", func() {
			Expect(func() {
				animation.NewFixedRateKeyframeChannel(1.0, 2.0, nil)
			}).To(Panic())
		})

		It("panics with an empty frame list", func() {
			Expect(func() {
				animation.NewFixedRateKeyframeChannel(1.0, 2.0, []animation.Transform{})
			}).To(Panic())
		})

		It("panics with a zero frame rate", func() {
			Expect(func() {
				animation.NewFixedRateKeyframeChannel(1.0, 0.0, frames)
			}).To(Panic())
		})

		It("panics with a negative frame rate", func() {
			Expect(func() {
				animation.NewFixedRateKeyframeChannel(1.0, -2.0, frames)
			}).To(Panic())
		})
	})

	Describe("Transform", func() {

		var channel *animation.FixedRateKeyframeChannel

		BeforeEach(func() {
			channel = animation.NewFixedRateKeyframeChannel(1.0, 2.0, frames)
		})

		It("interpolates between the surrounding frames", func() {
			// At 1.125 the timestamp is a quarter of the way between the first
			// two frames.
			result := channel.Transform(1.125)
			expectVec3(result.Translation, dprec.NewVec3(1.0, 0.0, 0.0))
			expectRotation(result.Rotation, yRotation(quarterDegrees))
			expectVec3(result.Scale, dprec.NewVec3(1.5, 1.5, 1.5))
		})

		It("interpolates in the same way as BlendTransforms", func() {
			// At 1.625 the timestamp is a quarter of the way between the last
			// two frames.
			result := channel.Transform(1.625)
			expected := animation.BlendTransforms(secondTransform, thirdTransform, 0.25)
			expectVec3(result.Translation, expected.Translation)
			expectRotation(result.Rotation, expected.Rotation)
			expectVec3(result.Scale, expected.Scale)
		})

		It("returns the frame values at frame timestamps", func() {
			result := channel.Transform(1.5)
			expectVec3(result.Translation, secondTransform.Translation)
			expectRotation(result.Rotation, secondTransform.Rotation)
			expectVec3(result.Scale, secondTransform.Scale)
		})

		It("returns the first frame at the start timestamp", func() {
			result := channel.Transform(1.0)
			expectVec3(result.Translation, firstTransform.Translation)
			expectRotation(result.Rotation, firstTransform.Rotation)
			expectVec3(result.Scale, firstTransform.Scale)
		})

		It("returns the last frame at the end timestamp", func() {
			result := channel.Transform(2.0)
			expectVec3(result.Translation, thirdTransform.Translation)
			expectRotation(result.Rotation, thirdTransform.Rotation)
			expectVec3(result.Scale, thirdTransform.Scale)
		})

		It("holds the first frame before the start", func() {
			result := channel.Transform(-1.0)
			expectVec3(result.Translation, firstTransform.Translation)
			expectRotation(result.Rotation, firstTransform.Rotation)
			expectVec3(result.Scale, firstTransform.Scale)
		})

		It("holds the last frame after the end", func() {
			result := channel.Transform(10.0)
			expectVec3(result.Translation, thirdTransform.Translation)
			expectRotation(result.Rotation, thirdTransform.Rotation)
			expectVec3(result.Scale, thirdTransform.Scale)
		})

		It("holds the value of a single frame", func() {
			constant := animation.NewFixedRateKeyframeChannel(1.0, 2.0, []animation.Transform{
				secondTransform,
			})
			for _, timestamp := range []float64{-1.0, 1.0, 2.5, 10.0} {
				result := constant.Transform(timestamp)
				expectVec3(result.Translation, secondTransform.Translation)
				expectRotation(result.Rotation, secondTransform.Rotation)
				expectVec3(result.Scale, secondTransform.Scale)
			}
		})

		It("locates frames at a typical frame rate", func() {
			// Frame i is translated by i along X, so the X translation reveals
			// the frame position that the timestamp maps to.
			sampled := make([]animation.Transform, 61)
			for i := range sampled {
				sampled[i] = animation.Transform{
					Translation: dprec.NewVec3(float64(i), 0.0, 0.0),
					Rotation:    dprec.IdentityQuat(),
					Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
				}
			}
			channel := animation.NewFixedRateKeyframeChannel(0.5, 30.0, sampled)
			expectVec3(channel.Transform(1.5).Translation, dprec.NewVec3(30.0, 0.0, 0.0))
			expectVec3(channel.Transform(1.5+0.5/30.0).Translation, dprec.NewVec3(30.5, 0.0, 0.0))
			expectVec3(channel.Transform(2.5).Translation, dprec.NewVec3(60.0, 0.0, 0.0))
		})

		It("interpolates rotation along the shorter arc", func() {
			// The second frame is in the opposite hemisphere, so a direct
			// interpolation would take the long way around.
			flipped := secondTransform
			flipped.Rotation = dprec.NegativeQuat(secondTransform.Rotation)
			shortArc := animation.NewFixedRateKeyframeChannel(0.0, 1.0, []animation.Transform{
				firstTransform,
				flipped,
			})
			expectRotation(shortArc.Transform(0.5).Rotation, yRotation(45.0))
		})
	})
})
