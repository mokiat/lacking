package asset

import "github.com/mokiat/gomath/dprec"

type Animation struct {
	Name      string
	StartTime float64
	EndTime   float64
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
	Timestamp   float64
	Translation dprec.Vec3
}

type RotationKeyframe struct {
	Timestamp float64
	Rotation  dprec.Quat
}

type ScaleKeyframe struct {
	Timestamp float64
	Scale     dprec.Vec3
}
