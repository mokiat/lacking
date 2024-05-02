package mdl

import "github.com/mokiat/gomath/dprec"

type Animation struct {
	name      string
	startTime float64
	endTime   float64
	bindings  []*AnimationBinding
}

func (a *Animation) Name() string {
	return a.name
}

func (a *Animation) SetName(value string) {
	a.name = value
}

func (a *Animation) StartTime() float64 {
	return a.startTime
}

func (a *Animation) SetStartTime(value float64) {
	a.startTime = value
}

func (a *Animation) EndTime() float64 {
	return a.endTime
}

func (a *Animation) SetEndTime(value float64) {
	a.endTime = value
}

func (a *Animation) Bindings() []*AnimationBinding {
	return a.bindings
}

func (a *Animation) AddBinding(binding *AnimationBinding) {
	a.bindings = append(a.bindings, binding)
}

type AnimationBinding struct {
	nodeName             string
	translationKeyframes []TranslationKeyframe
	rotationKeyframes    []RotationKeyframe
	scaleKeyframes       []ScaleKeyframe
}

func (a *AnimationBinding) NodeName() string {
	return a.nodeName
}

func (a *AnimationBinding) SetNodeName(value string) {
	a.nodeName = value
}

func (a *AnimationBinding) TranslationKeyframes() []TranslationKeyframe {
	return a.translationKeyframes
}

func (a *AnimationBinding) AddTranslationKeyframe(keyframe TranslationKeyframe) {
	a.translationKeyframes = append(a.translationKeyframes, keyframe)
}

func (a *AnimationBinding) RotationKeyframes() []RotationKeyframe {
	return a.rotationKeyframes
}

func (a *AnimationBinding) AddRotationKeyframe(keyframe RotationKeyframe) {
	a.rotationKeyframes = append(a.rotationKeyframes, keyframe)
}

func (a *AnimationBinding) ScaleKeyframes() []ScaleKeyframe {
	return a.scaleKeyframes
}

func (a *AnimationBinding) AddScaleKeyframe(keyframe ScaleKeyframe) {
	a.scaleKeyframes = append(a.scaleKeyframes, keyframe)
}

type Keyframe[T any] struct {
	Timestamp float64
	Value     T
}

type TranslationKeyframe = Keyframe[dprec.Vec3]

type RotationKeyframe = Keyframe[dprec.Quat]

type ScaleKeyframe = Keyframe[dprec.Vec3]
