package asset

import "github.com/mokiat/gomath/sprec"

type Animation struct {
	Name      string
	StartTime float32
	EndTime   float32
	Bindings  []AnimationBinding
}

type AnimationBinding struct {
	NodeIndex            int32
	NodeName             string // alternative in case of isolated animation
	TranslationKeyframes []TranslationKeyframe
	RotationKeyframes    []RotationKeyframe
	ScaleKeyframes       []ScaleKeyframe
}

type TranslationKeyframe struct {
	Timestamp   float32
	Translation sprec.Vec3
}

type RotationKeyframe struct {
	Timestamp float32
	Rotation  sprec.Quat
}

type ScaleKeyframe struct {
	Timestamp float32
	Scale     sprec.Vec3
}
