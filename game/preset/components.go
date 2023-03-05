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
	FollowCameraComponentID   = ecs.NewComponentTypeID()
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

type FollowCameraComponent struct {
	Target         *game.Node
	AnchorPosition dprec.Vec3
	AnchorDistance float64
	CameraDistance float64
	PitchAngle     dprec.Angle
	YawAngle       dprec.Angle
	Zoom           float64
}

func (*FollowCameraComponent) TypeID() ecs.ComponentTypeID {
	return FollowCameraComponentID
}
