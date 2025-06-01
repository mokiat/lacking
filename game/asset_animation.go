package game

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/animation"
	"github.com/mokiat/lacking/game/asset/dto"
)

func (l *AssetLoader) ResolveAnimationRecording(assetAnimation dto.Animation) (Identifiable[*animation.Recording], error) {
	recording := animation.NewRecording()
	recording.SetName(assetAnimation.Name)
	recording.SetStartTime(assetAnimation.StartTime)
	recording.SetEndTime(assetAnimation.EndTime)
	recording.SetLoop(assetAnimation.Loop)

	for _, assetBinding := range assetAnimation.Bindings {
		translationKeyframes := make([]animation.Keyframe[dprec.Vec3], len(assetBinding.TranslationKeyframes))
		for k, keyframe := range assetBinding.TranslationKeyframes {
			translationKeyframes[k] = animation.Keyframe[dprec.Vec3]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Value,
			}
		}
		rotationKeyframes := make([]animation.Keyframe[dprec.Quat], len(assetBinding.RotationKeyframes))
		for k, keyframe := range assetBinding.RotationKeyframes {
			rotationKeyframes[k] = animation.Keyframe[dprec.Quat]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Value,
			}
		}
		scaleKeyframes := make([]animation.Keyframe[dprec.Vec3], len(assetBinding.ScaleKeyframes))
		for k, keyframe := range assetBinding.ScaleKeyframes {
			scaleKeyframes[k] = animation.Keyframe[dprec.Vec3]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Value,
			}
		}
		recording.SetBinding(assetBinding.NodeName, animation.KeyframeSet{
			TranslationKeyframes: translationKeyframes,
			RotationKeyframes:    rotationKeyframes,
			ScaleKeyframes:       scaleKeyframes,
		})
	}
	return Identifiable[*animation.Recording]{
		ID:    assetAnimation.ID,
		Value: recording,
	}, nil
}

func (l *AssetLoader) ResolveAnimationRecordings(assetAnimations []dto.Animation) (IdentifiableList[*animation.Recording], error) {
	recordings := make(IdentifiableList[*animation.Recording], len(assetAnimations))
	for i, assetAnimation := range assetAnimations {
		recording, err := l.ResolveAnimationRecording(assetAnimation)
		if err != nil {
			return nil, err
		}
		recordings[i] = recording
	}
	return recordings, nil
}
