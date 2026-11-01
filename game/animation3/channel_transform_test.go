package animation3_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("TransformKeyframeChannel", func() {

	var (
		firstTransform  animation3.Transform
		secondTransform animation3.Transform
		thirdTransform  animation3.Transform
		keyframes       animation3.KeyframeList[animation3.Transform]
	)

	BeforeEach(func() {
		firstTransform = animation3.Transform{
			Translation: dprec.NewVec3(0.0, 0.0, 0.0),
			Rotation:    yRotation(0.0),
			Scale:       dprec.NewVec3(1.0, 1.0, 1.0),
		}
		secondTransform = animation3.Transform{
			Translation: dprec.NewVec3(4.0, 0.0, 0.0),
			Rotation:    yRotation(90.0),
			Scale:       dprec.NewVec3(3.0, 3.0, 3.0),
		}
		thirdTransform = animation3.Transform{
			Translation: dprec.NewVec3(4.0, 2.0, 0.0),
			Rotation:    yRotation(90.0),
			Scale:       dprec.NewVec3(1.0, 2.0, 3.0),
		}
		keyframes = animation3.KeyframeList[animation3.Transform]{
			{Timestamp: 1.0, Value: firstTransform},
			{Timestamp: 3.0, Value: secondTransform},
			{Timestamp: 4.0, Value: thirdTransform},
		}
	})

	Describe("NewTransformKeyframeChannel", func() {

		It("panics without keyframes", func() {
			Expect(func() {
				animation3.NewTransformKeyframeChannel(nil)
			}).To(Panic())
		})

		It("panics with an empty keyframe list", func() {
			Expect(func() {
				animation3.NewTransformKeyframeChannel(animation3.KeyframeList[animation3.Transform]{})
			}).To(Panic())
		})
	})

	Describe("Transform", func() {

		var channel *animation3.TransformKeyframeChannel

		BeforeEach(func() {
			channel = animation3.NewTransformKeyframeChannel(keyframes)
		})

		It("interpolates between the surrounding keyframes", func() {
			// At 1.5 the timestamp is a quarter of the way between the first
			// two keyframes.
			result := channel.Transform(1.5)
			expectVec3(result.Translation, dprec.NewVec3(1.0, 0.0, 0.0))
			expectRotation(result.Rotation, yRotation(quarterDegrees))
			expectVec3(result.Scale, dprec.NewVec3(1.5, 1.5, 1.5))
		})

		It("interpolates in the same way as BlendTransforms", func() {
			// At 3.25 the timestamp is a quarter of the way between the last
			// two keyframes.
			result := channel.Transform(3.25)
			expected := animation3.BlendTransforms(secondTransform, thirdTransform, 0.25)
			expectVec3(result.Translation, expected.Translation)
			expectRotation(result.Rotation, expected.Rotation)
			expectVec3(result.Scale, expected.Scale)
		})

		It("returns the keyframe values at keyframe timestamps", func() {
			result := channel.Transform(3.0)
			expectVec3(result.Translation, secondTransform.Translation)
			expectRotation(result.Rotation, secondTransform.Rotation)
			expectVec3(result.Scale, secondTransform.Scale)
		})

		It("holds the first keyframe before the start", func() {
			result := channel.Transform(-1.0)
			expectVec3(result.Translation, firstTransform.Translation)
			expectRotation(result.Rotation, firstTransform.Rotation)
			expectVec3(result.Scale, firstTransform.Scale)
		})

		It("holds the last keyframe after the end", func() {
			result := channel.Transform(10.0)
			expectVec3(result.Translation, thirdTransform.Translation)
			expectRotation(result.Rotation, thirdTransform.Rotation)
			expectVec3(result.Scale, thirdTransform.Scale)
		})

		It("holds the value of a single keyframe", func() {
			constant := animation3.NewTransformKeyframeChannel(
				animation3.KeyframeList[animation3.Transform]{
					{Timestamp: 1.0, Value: secondTransform},
				},
			)
			for _, timestamp := range []float64{-1.0, 1.0, 2.5, 10.0} {
				result := constant.Transform(timestamp)
				expectVec3(result.Translation, secondTransform.Translation)
				expectRotation(result.Rotation, secondTransform.Rotation)
				expectVec3(result.Scale, secondTransform.Scale)
			}
		})

		It("interpolates rotation along the shorter arc", func() {
			// The second keyframe is in the opposite hemisphere, so a direct
			// interpolation would take the long way around.
			flipped := secondTransform
			flipped.Rotation = dprec.NegativeQuat(secondTransform.Rotation)
			shortArc := animation3.NewTransformKeyframeChannel(
				animation3.KeyframeList[animation3.Transform]{
					{Timestamp: 0.0, Value: firstTransform},
					{Timestamp: 2.0, Value: flipped},
				},
			)
			expectRotation(shortArc.Transform(1.0).Rotation, yRotation(45.0))
		})
	})
})
