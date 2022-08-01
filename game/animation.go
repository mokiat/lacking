package game

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/gomath/stod"
	"github.com/mokiat/lacking/log"
)

type Animation struct {
	Name      string
	StartTime float32
	EndTime   float32
	Bindings  []AnimationBinding
}

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
			translation := binding.Translation(float32(timestamp))
			binding.Node.SetPosition(stod.Vec3(translation))
		}
		if len(binding.RotationKeyframes) > 0 {
			rotation := binding.Rotation(float32(timestamp))
			binding.Node.SetRotation(stod.Quat(rotation))
		}
		if len(binding.ScaleKeyframes) > 0 {
			scale := binding.Scale(float32(timestamp))
			binding.Node.SetScale(stod.Vec3(scale))
		}
	}
}

type AnimationBinding struct {
	NodeName             string
	Node                 *Node
	TranslationKeyframes KeyframeList[sprec.Vec3]
	RotationKeyframes    KeyframeList[sprec.Quat]
	ScaleKeyframes       KeyframeList[sprec.Vec3]
}

func (b AnimationBinding) Translation(timestamp float32) sprec.Vec3 {
	left, right, t := b.TranslationKeyframes.Keyframe(timestamp)
	return sprec.Vec3Lerp(left.Value, right.Value, t)
}

func (b AnimationBinding) Rotation(timestamp float32) sprec.Quat {
	left, right, t := b.RotationKeyframes.Keyframe(timestamp)
	return sprec.QuatSlerp(left.Value, right.Value, t)
}

func (b AnimationBinding) Scale(timestamp float32) sprec.Vec3 {
	left, right, t := b.ScaleKeyframes.Keyframe(timestamp)
	return sprec.Vec3Lerp(left.Value, right.Value, t)
}

type KeyframeList[T any] []Keyframe[T]

func (l KeyframeList[T]) Keyframe(timestamp float32) (Keyframe[T], Keyframe[T], float32) {
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
	t := sprec.Clamp((timestamp-left.Timestamp)/(right.Timestamp-left.Timestamp), 0.0, 1.0)
	return left, right, t
}

type Keyframe[T any] struct {
	Timestamp float32
	Value     T
}
