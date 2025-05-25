package animationconv

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset/dto/animationdto"
	"github.com/mokiat/lacking/game/asset/mdl"
)

type Source interface {
	Animations() []*mdl.Animation
}

func CreateAnimationChunk(src Source) *animationdto.AnimationChunk {
	dtoAnimations := make([]animationdto.Animation, len(src.Animations()))
	for i, animation := range src.Animations() {
		dtoAnimation := animationdto.Animation{
			Name:      animation.Name(),
			StartTime: animation.StartTime(),
			EndTime:   animation.EndTime(),
			Bindings:  make([]animationdto.AnimationBinding, len(animation.Bindings())),
		}
		for i, binding := range animation.Bindings() {
			translationKeyframes := make([]animationdto.AnimationKeyframe[dprec.Vec3], len(binding.TranslationKeyframes()))
			for j, keyframe := range binding.TranslationKeyframes() {
				translationKeyframes[j] = animationdto.AnimationKeyframe[dprec.Vec3]{
					Timestamp: keyframe.Timestamp,
					Value:     keyframe.Value,
				}
			}
			rotationKeyframes := make([]animationdto.AnimationKeyframe[dprec.Quat], len(binding.RotationKeyframes()))
			for j, keyframe := range binding.RotationKeyframes() {
				rotationKeyframes[j] = animationdto.AnimationKeyframe[dprec.Quat]{
					Timestamp: keyframe.Timestamp,
					Value:     keyframe.Value,
				}
			}
			scaleKeyframes := make([]animationdto.AnimationKeyframe[dprec.Vec3], len(binding.ScaleKeyframes()))
			for j, keyframe := range binding.ScaleKeyframes() {
				scaleKeyframes[j] = animationdto.AnimationKeyframe[dprec.Vec3]{
					Timestamp: keyframe.Timestamp,
					Value:     keyframe.Value,
				}
			}
			dtoAnimation.Bindings[i] = animationdto.AnimationBinding{
				NodeName:             binding.NodeName(),
				TranslationKeyframes: translationKeyframes,
				RotationKeyframes:    rotationKeyframes,
				ScaleKeyframes:       scaleKeyframes,
			}
		}
		dtoAnimations[i] = dtoAnimation
	}
	return &animationdto.AnimationChunk{
		Animations: dtoAnimations,
	}
}
