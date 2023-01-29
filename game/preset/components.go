package preset

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/game/ecs"
)

var (
	NodeComponentID           = ecs.NewComponentTypeID()
	ControlledComponentID     = ecs.NewComponentTypeID()
	YawPitchCameraComponentID = ecs.NewComponentTypeID()
)

type NodeComponent struct {
	Node *game.Node
}

func (*NodeComponent) TypeID() ecs.ComponentTypeID {
	return NodeComponentID
}

const (
	ControlInputKeyboard ControlInput = 1 << iota
	ControlInputMouse
	ControlInputGamepad0
	ControlInputJoystick0
	ControlInputWheel0
)

type ControlInput int

func (i ControlInput) Is(query ControlInput) bool {
	return i&query != 0
}

type ControlledComponent struct {
	Inputs ControlInput
}

func (*ControlledComponent) TypeID() ecs.ComponentTypeID {
	return ControlledComponentID
}

type YawPitchCameraComponent struct {
	YawAngle   dprec.Angle
	PitchAngle dprec.Angle
}

func (*YawPitchCameraComponent) TypeID() ecs.ComponentTypeID {
	return YawPitchCameraComponentID
}
