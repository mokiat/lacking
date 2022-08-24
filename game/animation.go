package game

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/log"
)

type AnimationDefinition struct {
}

// TODO: Split animations into a definition
// and a an instance (that is bound to a hierarchy)
// The definition would contain the keyframes
// and the instance will contain the actual node bindings.

type Animation struct {
	Name      string
	StartTime float64
	EndTime   float64
	Bindings  []AnimationBinding
}

// TODO: Have this method be the constructor for an animation instance.
func (a *Animation) AttachToHierarchy(node *Node) {
	for i := 0; i < len(a.Bindings); i++ {
		binding := &a.Bindings[i]
		if target := node.FindNode(binding.NodeName); target != nil {
			binding.Node = target
		}
	}
}

func (a *Animation) Apply(timestamp float64) {
	// FIXME: This does not work for animation blending
	for _, binding := range a.Bindings {
		if binding.Node == nil {
			log.Warn("Binding is dangling")
			continue
		}
		if len(binding.TranslationKeyframes) > 0 {
			translation := binding.Translation(timestamp)
			binding.Node.SetPosition(translation)
		}
		if len(binding.RotationKeyframes) > 0 {
			rotation := binding.Rotation(timestamp)
			binding.Node.SetRotation(rotation)
		}
		if len(binding.ScaleKeyframes) > 0 {
			scale := binding.Scale(timestamp)
			binding.Node.SetScale(scale)
		}
	}
}

type AnimationBinding struct {
	NodeName             string
	Node                 *Node
	TranslationKeyframes KeyframeList[dprec.Vec3]
	RotationKeyframes    KeyframeList[dprec.Quat]
	ScaleKeyframes       KeyframeList[dprec.Vec3]
}

func (b AnimationBinding) Translation(timestamp float64) dprec.Vec3 {
	left, right, t := b.TranslationKeyframes.Keyframe(timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, t)
}

func (b AnimationBinding) Rotation(timestamp float64) dprec.Quat {
	left, right, t := b.RotationKeyframes.Keyframe(timestamp)
	return dprec.QuatSlerp(left.Value, right.Value, t)
}

func (b AnimationBinding) Scale(timestamp float64) dprec.Vec3 {
	left, right, t := b.ScaleKeyframes.Keyframe(timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, t)
}

type KeyframeList[T any] []Keyframe[T]

func (l KeyframeList[T]) Keyframe(timestamp float64) (Keyframe[T], Keyframe[T], float64) {
	leftIndex := 0
	rightIndex := len(l) - 1
	for leftIndex < rightIndex-1 {
		middleIndex := (leftIndex + rightIndex) / 2
		middle := l[middleIndex]
		if middle.Timestamp <= timestamp {
			leftIndex = middleIndex
		}
		if middle.Timestamp >= timestamp {
			rightIndex = middleIndex
		}
	}
	left := l[leftIndex]
	right := l[rightIndex]
	t := dprec.Clamp((timestamp-left.Timestamp)/(right.Timestamp-left.Timestamp), 0.0, 1.0)
	return left, right, t
}

type Keyframe[T any] struct {
	Timestamp float64
	Value     T
}
