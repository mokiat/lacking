package preset

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/ui"
)

func NewCarSystem(ecsScene *ecs.Scene, gamepadProvider GamepadStateProvider) *CarSystem {
	return &CarSystem{
		ecsScene:        ecsScene,
		gamepadProvider: gamepadProvider,
	}
}

type CarSystem struct {
	ecsScene        *ecs.Scene
	gamepadProvider GamepadStateProvider

	hasKeyboardConsumer bool

	keyForward   ui.KeyCode
	keyReverse   ui.KeyCode
	keyLeft      ui.KeyCode
	keyRight     ui.KeyCode
	keyHandbrake ui.KeyCode
	keyRecover   ui.KeyCode

	isForward   bool
	isReverse   bool
	isLeft      bool
	isRight     bool
	isHandbrake bool
	isRecover   bool
}

func (s *CarSystem) UseDefaults() {
	s.keyForward = ui.KeyCodeArrowUp
	s.keyReverse = ui.KeyCodeArrowDown
	s.keyLeft = ui.KeyCodeArrowLeft
	s.keyRight = ui.KeyCodeArrowRight
	s.keyHandbrake = ui.KeyCodeLeftControl
	s.keyRecover = ui.KeyCodeEnter
}

func (s *CarSystem) OnKeyboardEvent(event ui.KeyboardEvent) bool {
	if !s.hasKeyboardConsumer {
		return false
	}
	active := event.Type != ui.KeyboardEventTypeKeyUp
	switch event.Code {
	case s.keyForward:
		s.isForward = active
		return true
	case s.keyRecover:
		s.isReverse = active
		return true
	case s.keyLeft:
		s.isLeft = active
		return true
	case s.keyRight:
		s.isRight = active
		return true
	case s.keyHandbrake:
		s.isHandbrake = active
		return true
	case s.keyRecover:
		s.isRecover = active
		return true
	default:
		return false
	}
}

func (s *CarSystem) Update(elapsedSeconds float64) {
	result := s.ecsScene.Find(ecs.Having(CarComponentID))
	defer result.Close()

	var hasKeyboardConsumer bool
	var entity *ecs.Entity
	for result.FetchNext(&entity) {
		var controlled *ControlledComponent
		if ecs.FetchComponent(entity, &controlled) {
			if controlled.Inputs.Is(ControlInputKeyboard) {
				hasKeyboardConsumer = true
				s.updateKeyboard(elapsedSeconds, entity)
			}
			if controlled.Inputs.Is(ControlInputMouse) {
				s.updateMouse(elapsedSeconds, entity)
			}
			if controlled.Inputs.Is(ControlInputGamepad0) {
				if gamepad, ok := s.gamepadProvider.GamepadState(0); ok {
					s.updateGamepad(elapsedSeconds, gamepad, entity)
				}
			}
		}
		s.updateCar(elapsedSeconds, entity)

	}
	s.hasKeyboardConsumer = hasKeyboardConsumer
}

func (s *CarSystem) updateKeyboard(elapsedSeconds float64, entity *ecs.Entity) {
	// TODO
}

func (s *CarSystem) updateMouse(elapsedSeconds float64, entity *ecs.Entity) {
	// TODO
}

func (s *CarSystem) updateGamepad(elapsedSeconds float64, gamepad app.GamepadState, entity *ecs.Entity) {
	var carComp *CarComponent
	ecs.FetchComponent(entity, &carComp)

	carComp.SteeringAmount = gamepad.LeftStickX * gamepad.LeftStickX * gamepad.LeftStickX
	carComp.Acceleration = gamepad.RightTrigger
	carComp.Deceleration = gamepad.LeftTrigger
	carComp.DesiredDirection = CarDirectionNeutral
	if gamepad.SquareButton {
		carComp.DesiredDirection = CarDirectionReverse
	}
	if gamepad.CrossButton {
		carComp.DesiredDirection = CarDirectionForward
	}
	carComp.Recover = gamepad.TriangleButton
}

func (s *CarSystem) updateCar(elapsedSeconds float64, entity *ecs.Entity) {
	// TODO: Run this inside physics loop for smooth operation.

	var carComp *CarComponent
	ecs.FetchComponent(entity, &carComp)
	var (
		car         = carComp.Car
		chassisBody = car.Chassis().Body()
	)

	if carComp.Recover {
		rotationVector := dprec.Vec3Cross(
			chassisBody.Orientation().OrientationY(),
			dprec.BasisYVec3(),
		)
		chassisBody.SetAngularVelocity(dprec.Vec3Prod(
			rotationVector, 100*elapsedSeconds,
		))
		velocity := chassisBody.Velocity()
		velocity.Y = 2.0
		chassisBody.SetVelocity(velocity)
	}

	forwardSpeed := dprec.Vec3Dot(
		car.Chassis().Body().Velocity(),
		car.Chassis().Body().Orientation().OrientationZ(),
	)
	isMovingForward := forwardSpeed > 2.0
	isMovingBackward := forwardSpeed < -2.0

	if !isMovingBackward && (carComp.DesiredDirection == CarDirectionForward) {
		carComp.Direction = CarDirectionForward
	}
	if !isMovingForward && (carComp.DesiredDirection == CarDirectionReverse) {
		carComp.Direction = CarDirectionReverse
	}

	for _, axis := range car.Axes() {
		// TODO: Use Ackermann steering. Needs an additional steering offset (intersection line) parameter.
		steeringAngle := -axis.maxSteeringAngle * dprec.Angle(carComp.SteeringAmount)
		steeringQuat := dprec.RotationQuat(steeringAngle, dprec.BasisYVec3())
		direction := dprec.QuatVec3Rotation(steeringQuat, dprec.BasisXVec3())

		leftDirectionSolver := axis.leftWheel.directionSolver
		leftDirectionSolver.SetPrimaryDirection(direction)

		rightDirectionSolver := axis.rightWheel.directionSolver
		rightDirectionSolver.SetPrimaryDirection(direction)

		// Acceleration
		var deltaVelocity float64
		if carComp.Direction == CarDirectionForward {
			deltaVelocity = axis.maxAcceleration * carComp.Acceleration * elapsedSeconds
		} else {
			deltaVelocity = -axis.maxAcceleration * carComp.Acceleration * axis.reverseRatio * elapsedSeconds
		}

		leftWheelBody := axis.LeftWheel().Body()
		rightWheelBody := axis.RightWheel().Body()

		leftWheelBody.SetAngularVelocity(dprec.Vec3Sum(leftWheelBody.AngularVelocity(),
			dprec.Vec3Prod(leftWheelBody.Orientation().OrientationX(), deltaVelocity),
		))
		rightWheelBody.SetAngularVelocity(dprec.Vec3Sum(rightWheelBody.AngularVelocity(),
			dprec.Vec3Prod(rightWheelBody.Orientation().OrientationX(), deltaVelocity),
		))

		// Braking
		if carComp.Deceleration > 0.0 {
			// TODO: Implement ABS

			leftWheelVelocity := dprec.Vec3Dot(
				leftWheelBody.AngularVelocity(),
				leftWheelBody.Orientation().OrientationX(),
			)
			leftWheelCorrection := -dprec.Min(axis.maxBraking*carComp.Deceleration*elapsedSeconds, leftWheelVelocity)
			leftWheelBody.SetAngularVelocity(dprec.Vec3Sum(
				leftWheelBody.AngularVelocity(),
				dprec.Vec3Prod(leftWheelBody.Orientation().OrientationX(), leftWheelCorrection),
			))

			rightWheelVelocity := dprec.Vec3Dot(
				rightWheelBody.AngularVelocity(),
				rightWheelBody.Orientation().OrientationX(),
			)
			rightWheelCorrection := -dprec.Min(axis.maxBraking*carComp.Deceleration*elapsedSeconds, rightWheelVelocity)
			rightWheelBody.SetAngularVelocity(dprec.Vec3Sum(
				rightWheelBody.AngularVelocity(),
				dprec.Vec3Prod(rightWheelBody.Orientation().OrientationX(), rightWheelCorrection),
			))
		}
	}
}
