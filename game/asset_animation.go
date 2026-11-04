package game

import (
	"github.com/mokiat/lacking/game/animation"
	"github.com/mokiat/lacking/game/asset/dto"
)

// LoadAnimationRecording loads an animation recording from the given asset.
//
// The rest transforms, keyed by node name, provide the value of the
// components (translation, rotation or scale) that a binding of the
// recording does not animate. Components of nodes without a rest transform
// default to those of the [animation.IdentityTransform].
//
// This is a blocking operation and should be called from a worker thread.
func LoadAnimationRecording(loader *AssetLoader, assetAnimation dto.Animation, restTransforms map[string]animation.Transform) (Identifiable[*animation.Clip], error) {
	channels := make(map[string]animation.Channel, len(assetAnimation.Bindings))
	for _, assetBinding := range assetAnimation.Bindings {
		rest, ok := restTransforms[assetBinding.NodeName]
		if !ok {
			rest = animation.IdentityTransform()
		}
		channels[assetBinding.NodeName] = animation.NewComponentKeyframeChannel(
			convertKeyframes(assetBinding.TranslationKeyframes, rest.Translation),
			convertKeyframes(assetBinding.RotationKeyframes, rest.Rotation),
			convertKeyframes(assetBinding.ScaleKeyframes, rest.Scale),
		)
	}
	clip := animation.NewClip(
		assetAnimation.Name,
		assetAnimation.StartTime,
		assetAnimation.EndTime,
		channels,
	)
	return Identifiable[*animation.Clip]{
		ID:    assetAnimation.ID,
		Value: clip,
	}, nil
}

// LoadAnimationRecordings loads a list of animation recordings from the given
// asset animations. See [LoadAnimationRecording] for the meaning of the rest
// transforms.
//
// This is a blocking operation and should be called from a worker thread.
func LoadAnimationRecordings(loader *AssetLoader, assetAnimations []dto.Animation, restTransforms map[string]animation.Transform) (IdentifiableList[*animation.Clip], error) {
	recordings := make(IdentifiableList[*animation.Clip], len(assetAnimations))
	for i, assetAnimation := range assetAnimations {
		recording, err := LoadAnimationRecording(loader, assetAnimation, restTransforms)
		if err != nil {
			return nil, err
		}
		recordings[i] = recording
	}
	return recordings, nil
}

// UnloadRecording unloads the given animation recording.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadRecording(loader *AssetLoader, idRecording Identifiable[*animation.Clip]) error {
	// At the time being this is a no-op.
	return nil
}

// UnloadRecordings unloads a list of animation recordings.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadRecordings(loader *AssetLoader, idRecordings IdentifiableList[*animation.Clip]) error {
	for _, idRecording := range idRecordings {
		if err := UnloadRecording(loader, idRecording); err != nil {
			return err
		}
	}
	return nil
}

// convertKeyframes converts asset keyframes to an animation keyframe list.
// An empty list of asset keyframes produces a single keyframe that holds the
// specified rest value, since an animation channel needs at least one
// keyframe per component.
func convertKeyframes[T any](assetKeyframes []dto.AnimationKeyframe[T], rest T) animation.KeyframeList[T] {
	if len(assetKeyframes) == 0 {
		return animation.KeyframeList[T]{
			{Timestamp: 0.0, Value: rest},
		}
	}
	result := make(animation.KeyframeList[T], len(assetKeyframes))
	for i, keyframe := range assetKeyframes {
		result[i] = animation.Keyframe[T]{
			Timestamp: keyframe.Timestamp,
			Value:     keyframe.Value,
		}
	}
	return result
}

// restTransformsOf returns the transforms of the specified asset nodes, keyed
// by node name. The first of several nodes that share a name wins.
func restTransformsOf(assetNodes []dto.Node) map[string]animation.Transform {
	result := make(map[string]animation.Transform, len(assetNodes))
	for _, assetNode := range assetNodes {
		if _, ok := result[assetNode.Name]; ok {
			continue
		}
		result[assetNode.Name] = animation.Transform{
			Translation: assetNode.Translation,
			Rotation:    assetNode.Rotation,
			Scale:       assetNode.Scale,
		}
	}
	return result
}
