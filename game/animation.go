package game

import (
	"github.com/mokiat/gomath/dprec"
)

type AnimationDefinitionInfo struct {
	Name      string
	StartTime float64
	EndTime   float64
	Bindings  []AnimationBindingDefinitionInfo
}

type AnimationBindingDefinitionInfo struct {
	NodeIndex            int
	NodeName             string //alternative in case of isolated animation
	TranslationKeyframes KeyframeList[dprec.Vec3]
	RotationKeyframes    KeyframeList[dprec.Quat]
	ScaleKeyframes       KeyframeList[dprec.Vec3]
}

type AnimationDefinition struct {
	name      string
	startTime float64
	endTime   float64
	bindings  []AnimationBindingDefinitionInfo
}

type AnimationInfo struct {
	Model      *Model
	Definition *AnimationDefinition
}

type Animation struct {
	name       string
	definition *AnimationDefinition
	bindings   []animationBinding
}

func (a *Animation) Name() string {
	return a.name
}

func (a *Animation) StartTime() float64 {
	return a.definition.startTime
}

func (a *Animation) EndTime() float64 {
	return a.definition.endTime
}

func (a *Animation) Apply(timestamp float64) {
	// FIXME: This does not work for animation blending
	for _, binding := range a.bindings {
		if binding.node == nil {
			continue
		}
		if len(binding.translationKeyframes) > 0 {
			translation := binding.Translation(timestamp)
			binding.node.SetPosition(translation)
		}
		if len(binding.rotationKeyframes) > 0 {
			rotation := binding.Rotation(timestamp)
			binding.node.SetRotation(rotation)
		}
		if len(binding.scaleKeyframes) > 0 {
			scale := binding.Scale(timestamp)
			binding.node.SetScale(scale)
		}
	}
}

type animationBinding struct {
	node                 *Node
	translationKeyframes KeyframeList[dprec.Vec3]
	rotationKeyframes    KeyframeList[dprec.Quat]
	scaleKeyframes       KeyframeList[dprec.Vec3]
}

func (b animationBinding) Translation(timestamp float64) dprec.Vec3 {
	left, right, t := b.translationKeyframes.Keyframe(timestamp)
	return dprec.Vec3Lerp(left.Value, right.Value, t)
}

func (b animationBinding) Rotation(timestamp float64) dprec.Quat {
	left, right, t := b.rotationKeyframes.Keyframe(timestamp)
	return dprec.QuatSlerp(left.Value, right.Value, t)
}

func (b animationBinding) Scale(timestamp float64) dprec.Vec3 {
	left, right, t := b.scaleKeyframes.Keyframe(timestamp)
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
