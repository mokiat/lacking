package game

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) convertHierarchyNode(assetNode dto.Node) HierarchyNodeTemplate {
	return HierarchyNodeTemplate{
		ID:       assetNode.ID,
		ParentID: assetNode.ParentID,
		Name:     assetNode.Name,
		Position: assetNode.Translation,
		Rotation: assetNode.Rotation,
		Scale:    assetNode.Scale,
	}
}

func (s *ResourceSet) convertAnimation(assetAnimation dto.Animation) async.Promise[*AnimationDefinition] {
	bindings := make([]AnimationBindingDefinitionInfo, len(assetAnimation.Bindings))
	for j, assetBinding := range assetAnimation.Bindings {
		translationKeyframes := make([]Keyframe[dprec.Vec3], len(assetBinding.TranslationKeyframes))
		for k, keyframe := range assetBinding.TranslationKeyframes {
			translationKeyframes[k] = Keyframe[dprec.Vec3]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Value,
			}
		}
		rotationKeyframes := make([]Keyframe[dprec.Quat], len(assetBinding.RotationKeyframes))
		for k, keyframe := range assetBinding.RotationKeyframes {
			rotationKeyframes[k] = Keyframe[dprec.Quat]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Value,
			}
		}
		scaleKeyframes := make([]Keyframe[dprec.Vec3], len(assetBinding.ScaleKeyframes))
		for k, keyframe := range assetBinding.ScaleKeyframes {
			scaleKeyframes[k] = Keyframe[dprec.Vec3]{
				Timestamp: keyframe.Timestamp,
				Value:     keyframe.Value,
			}
		}
		bindings[j] = AnimationBindingDefinitionInfo{
			NodeName:             assetBinding.NodeName,
			TranslationKeyframes: translationKeyframes,
			RotationKeyframes:    rotationKeyframes,
			ScaleKeyframes:       scaleKeyframes,
		}
	}

	promise := async.NewPromise[*AnimationDefinition]()
	s.gfxWorker.Schedule(func() {
		animation := s.engine.CreateAnimationDefinition(AnimationDefinitionInfo{
			Name:      assetAnimation.Name,
			StartTime: assetAnimation.StartTime,
			EndTime:   assetAnimation.EndTime,
			Bindings:  bindings,
		})
		promise.Deliver(animation)
	})
	return promise
}
