package animation3_test

import (
	"github.com/mokiat/gomath/dprec"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/animation3"
)

var _ = Describe("ComponentKeyframeChannel", func() {

	var (
		translations animation3.KeyframeList[dprec.Vec3]
		rotations    animation3.KeyframeList[dprec.Quat]
		scales       animation3.KeyframeList[dprec.Vec3]
	)

	BeforeEach(func() {
		// Each component is keyed at its own timestamps.
		translations = animation3.KeyframeList[dprec.Vec3]{
			{Timestamp: 0.0, Value: dprec.NewVec3(0.0, 0.0, 0.0)},
			{Timestamp: 2.0, Value: dprec.NewVec3(4.0, 0.0, 0.0)},
			{Timestamp: 3.0, Value: dprec.NewVec3(4.0, 2.0, 0.0)},
		}
		rotations = animation3.KeyframeList[dprec.Quat]{
			{Timestamp: 1.0, Value: yRotation(0.0)},
			{Timestamp: 3.0, Value: yRotation(90.0)},
		}
		scales = animation3.KeyframeList[dprec.Vec3]{
			{Timestamp: 0.0, Value: dprec.NewVec3(1.0, 1.0, 1.0)},
			{Timestamp: 4.0, Value: dprec.NewVec3(3.0, 3.0, 3.0)},
		}
	})

	Describe("NewComponentKeyframeChannel", func() {

		It("panics without translation keyframes", func() {
			Expect(func() {
				animation3.NewComponentKeyframeChannel(nil, rotations, scales)
			}).To(Panic())
		})

		It("panics without rotation keyframes", func() {
			Expect(func() {
				animation3.NewComponentKeyframeChannel(translations, animation3.KeyframeList[dprec.Quat]{}, scales)
			}).To(Panic())
		})

		It("panics without scale keyframes", func() {
			Expect(func() {
				animation3.NewComponentKeyframeChannel(translations, rotations, nil)
			}).To(Panic())
		})
	})

	Describe("Transform", func() {

		var channel *animation3.ComponentKeyframeChannel

		BeforeEach(func() {
			channel = animation3.NewComponentKeyframeChannel(translations, rotations, scales)
		})

		It("interpolates each component within its own timeline", func() {
			// At 1.5 the translation is three quarters of the way between its
			// first two keyframes, the rotation a quarter of the way and the
			// scale three eighths of the way.
			result := channel.Transform(1.5)
			expectVec3(result.Translation, dprec.NewVec3(3.0, 0.0, 0.0))
			expectRotation(result.Rotation, dprec.QuatNLerp(yRotation(0.0), yRotation(90.0), 0.25))
			expectVec3(result.Scale, dprec.NewVec3(1.75, 1.75, 1.75))
		})

		It("returns the keyframe values at keyframe timestamps", func() {
			result := channel.Transform(3.0)
			expectVec3(result.Translation, dprec.NewVec3(4.0, 2.0, 0.0))
			expectRotation(result.Rotation, yRotation(90.0))
			expectVec3(result.Scale, dprec.NewVec3(2.5, 2.5, 2.5))
		})

		It("interpolates rotation halfway at the midpoint", func() {
			// Normalized linear interpolation matches spherical interpolation
			// at the midpoint.
			expectRotation(channel.Transform(2.0).Rotation, yRotation(45.0))
		})

		It("holds the first keyframes before the start", func() {
			result := channel.Transform(-1.0)
			expectVec3(result.Translation, dprec.NewVec3(0.0, 0.0, 0.0))
			expectRotation(result.Rotation, yRotation(0.0))
			expectVec3(result.Scale, dprec.NewVec3(1.0, 1.0, 1.0))
		})

		It("holds the last keyframes after the end", func() {
			result := channel.Transform(10.0)
			expectVec3(result.Translation, dprec.NewVec3(4.0, 2.0, 0.0))
			expectRotation(result.Rotation, yRotation(90.0))
			expectVec3(result.Scale, dprec.NewVec3(3.0, 3.0, 3.0))
		})

		It("holds the value of a component with a single keyframe", func() {
			constant := animation3.NewComponentKeyframeChannel(
				translations,
				animation3.KeyframeList[dprec.Quat]{
					{Timestamp: 1.0, Value: yRotation(30.0)},
				},
				animation3.KeyframeList[dprec.Vec3]{
					{Timestamp: 1.0, Value: dprec.NewVec3(2.0, 2.0, 2.0)},
				},
			)
			for _, timestamp := range []float64{-1.0, 1.0, 2.5, 10.0} {
				result := constant.Transform(timestamp)
				expectRotation(result.Rotation, yRotation(30.0))
				expectVec3(result.Scale, dprec.NewVec3(2.0, 2.0, 2.0))
			}
		})

		It("interpolates rotation along the shorter arc", func() {
			// The second keyframe is in the opposite hemisphere, so a direct
			// interpolation would take the long way around.
			shortArc := animation3.NewComponentKeyframeChannel(
				translations,
				animation3.KeyframeList[dprec.Quat]{
					{Timestamp: 0.0, Value: yRotation(0.0)},
					{Timestamp: 2.0, Value: dprec.NegativeQuat(yRotation(90.0))},
				},
				scales,
			)
			expectRotation(shortArc.Transform(1.0).Rotation, yRotation(45.0))
		})
	})
})
