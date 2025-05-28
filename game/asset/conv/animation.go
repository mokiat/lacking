package conv

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/asset/mdl"
	"github.com/mokiat/lacking/storage/chunked"
)

type AnimationSource interface {
	Animations() []*mdl.Animation
}

func NewAnimationConverter() *AnimationConverter {
	return &AnimationConverter{}
}

type AnimationConverter struct{}

func (c *AnimationConverter) Convert(target *ds.List[chunked.Chunk], asset any) error {
	src, ok := asset.(AnimationSource)
	if !ok {
		return nil
	}
	chunk, err := c.CreateAnimationChunk(src)
	if err != nil {
		return err
	}
	target.Add(chunked.FromValue(dto.AnimationChunkID, chunk))
	return nil
}

func (c *AnimationConverter) CreateAnimationChunk(src AnimationSource) (*dto.AnimationChunk, error) {
	dtoAnimations := make([]dto.Animation, len(src.Animations()))
	for i, animation := range src.Animations() {
		dtoAnimation := dto.Animation{
			ID:        animation.ID(),
			Name:      animation.Name(),
			StartTime: animation.StartTime(),
			EndTime:   animation.EndTime(),
			Bindings:  make([]dto.AnimationBinding, len(animation.Bindings())),
		}
		for i, binding := range animation.Bindings() {
			translationKeyframes := make([]dto.AnimationKeyframe[dprec.Vec3], len(binding.TranslationKeyframes()))
			for j, keyframe := range binding.TranslationKeyframes() {
				translationKeyframes[j] = dto.AnimationKeyframe[dprec.Vec3]{
					Timestamp: keyframe.Timestamp,
					Value:     keyframe.Value,
				}
			}
			rotationKeyframes := make([]dto.AnimationKeyframe[dprec.Quat], len(binding.RotationKeyframes()))
			for j, keyframe := range binding.RotationKeyframes() {
				rotationKeyframes[j] = dto.AnimationKeyframe[dprec.Quat]{
					Timestamp: keyframe.Timestamp,
					Value:     keyframe.Value,
				}
			}
			scaleKeyframes := make([]dto.AnimationKeyframe[dprec.Vec3], len(binding.ScaleKeyframes()))
			for j, keyframe := range binding.ScaleKeyframes() {
				scaleKeyframes[j] = dto.AnimationKeyframe[dprec.Vec3]{
					Timestamp: keyframe.Timestamp,
					Value:     keyframe.Value,
				}
			}
			dtoAnimation.Bindings[i] = dto.AnimationBinding{
				NodeName:             binding.NodeName(),
				TranslationKeyframes: translationKeyframes,
				RotationKeyframes:    rotationKeyframes,
				ScaleKeyframes:       scaleKeyframes,
			}
		}
		dtoAnimations[i] = dtoAnimation
	}
	return &dto.AnimationChunk{
		Animations: dtoAnimations,
	}, nil
}
