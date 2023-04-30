package preset

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/ui"
)

var (
	CarComponentID       = ecs.NewComponentTypeID()
	CarKeyboardControlID = ecs.NewComponentTypeID()
	CarMouseControlID    = ecs.NewComponentTypeID()
	CarGamepadControlID  = ecs.NewComponentTypeID()
)

type CarGear int

const (
	CarGearNeutral CarGear = iota
	CarGearForward
	CarGearReverse
)

type CarComponent struct {
	Car            *Car
	Gear           CarGear
	SteeringAmount float64
	Acceleration   float64
	Deceleration   float64
	Recover        bool
	LightsOn       bool
}

func (*CarComponent) TypeID() ecs.ComponentTypeID {
	return CarComponentID
}

type CarKeyboardControl struct {
	AccelerateKey ui.KeyCode
	DecelerateKey ui.KeyCode
	TurnLeftKey   ui.KeyCode
	TurnRightKey  ui.KeyCode
	ShiftUpKey    ui.KeyCode
	ShiftDownKey  ui.KeyCode
	RecoverKey    ui.KeyCode

	AccelerationChangeSpeed float64
	DecelerationChangeSpeed float64
	SteeringAmount          float64
	SteeringChangeSpeed     float64
	SteeringRestoreSpeed    float64
}

func (*CarKeyboardControl) TypeID() ecs.ComponentTypeID {
	return CarKeyboardControlID
}

type CarMouseControl struct {
	AccelerationChangeSpeed float64
	DecelerationChangeSpeed float64
	Destination             dprec.Vec3
}

func (*CarMouseControl) TypeID() ecs.ComponentTypeID {
	return CarMouseControlID
}

type CarGamepadControl struct {
	Gamepad app.Gamepad
}

func (*CarGamepadControl) TypeID() ecs.ComponentTypeID {
	return CarGamepadControlID
}
