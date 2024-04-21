package game

import (
	"github.com/mokiat/gomath/dprec"
	asset "github.com/mokiat/lacking/game/newasset"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) convertNode(assetNode asset.Node) nodeDefinition {
	return nodeDefinition{
		ParentIndex: int(assetNode.ParentIndex),
		Name:        assetNode.Name,
		Position:    assetNode.Translation,
		Rotation:    assetNode.Rotation,
		Scale:       assetNode.Scale,
	}
}

func (s *ResourceSet) convertAnimation(assetAnimation asset.Animation) async.Promise[*AnimationDefinition] {
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
	s.gfxWorker.ScheduleVoid(func() {
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

func (s *ResourceSet) convertArmature(assetArmature asset.Armature) armatureDefinition {
	joints := make([]armatureJoint, len(assetArmature.Joints))
	for j, assetJoint := range assetArmature.Joints {
		joints[j] = armatureJoint{
			NodeIndex:         int(assetJoint.NodeIndex),
			InverseBindMatrix: assetJoint.InverseBindMatrix,
		}
	}
	return armatureDefinition{
		Joints: joints,
	}
}
