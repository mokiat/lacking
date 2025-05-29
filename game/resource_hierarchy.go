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

func (s *ResourceSet) convertAnimation(assetAnimation dto.Animation) async.Promise[AnimationTemplate] {
	bindings := make(map[string]AnimationKeyframeSet, len(assetAnimation.Bindings))
	for _, assetBinding := range assetAnimation.Bindings {
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
		bindings[assetBinding.NodeName] = AnimationKeyframeSet{
			TranslationKeyframes: translationKeyframes,
			RotationKeyframes:    rotationKeyframes,
			ScaleKeyframes:       scaleKeyframes,
		}
	}

	promise := async.NewPromise[AnimationTemplate]()
	s.gfxWorker.Schedule(func() {
		promise.Deliver(AnimationTemplate{
			ID:        assetAnimation.ID,
			Name:      assetAnimation.Name,
			StartTime: assetAnimation.StartTime,
			EndTime:   assetAnimation.EndTime,
			Loop:      assetAnimation.Loop,
			Bindings:  bindings,
		})
	})
	return promise
}
