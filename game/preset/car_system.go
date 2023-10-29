package preset

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/ui"
)

func NewCarSystem(ecsScene *ecs.Scene, gfxScene *graphics.Scene) *CarSystem {
	return &CarSystem{
		ecsScene: ecsScene,
		gfxScene: gfxScene,

		keysOfInterest: make(map[ui.KeyCode]struct{}),
		keyStates:      make(map[ui.KeyCode]bool),

		mouseOfInterest:   false,
		mouseButtonStates: make(map[ui.MouseButton]bool),
		mouseAreaWidth:    1.0,
		mouseAreaHeight:   1.0,
	}
}

type CarSystem struct {
	ecsScene *ecs.Scene
	gfxScene *graphics.Scene

	keysOfInterest map[ui.KeyCode]struct{}
	keyStates      map[ui.KeyCode]bool

	mouseOfInterest   bool
	mouseButtonStates map[ui.MouseButton]bool
	mouseAreaWidth    int
	mouseAreaHeight   int
	mouseX            int
	mouseY            int
	mouseScroll       float32
}

func (s *CarSystem) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	if !s.mouseOfInterest {
		return false
	}
	bounds := element.Bounds()
	s.mouseAreaWidth = bounds.Width
	s.mouseAreaHeight = bounds.Height
	switch event.Action {
	case ui.MouseActionDown:
		s.mouseButtonStates[event.Button] = true
	case ui.MouseActionUp:
		s.mouseButtonStates[event.Button] = false
	case ui.MouseActionScroll:
		s.mouseScroll += event.ScrollY
	case ui.MouseActionMove:
		s.mouseX = event.X
		s.mouseY = event.Y
	}
	return true
}

func (s *CarSystem) OnKeyboardEvent(event ui.KeyboardEvent) bool {
	if _, ok := s.keysOfInterest[event.Code]; !ok {
		return false
	}
	switch event.Action {
	case ui.KeyboardActionDown:
		s.keyStates[event.Code] = true
	case ui.KeyboardActionUp:
		s.keyStates[event.Code] = false
	}
	return true
}

func (s *CarSystem) Update(elapsedSeconds float64) {
	s.mouseOfInterest = false

	result := s.ecsScene.Find(ecs.Having(CarComponentID))
	defer result.Close()

	var entity *ecs.Entity
	for result.FetchNext(&entity) {
		var keyboardControl *CarKeyboardControl
		if ecs.FetchComponent(entity, &keyboardControl) {
			s.updateKeyboard(elapsedSeconds, entity)
		}

		var mouseControl *CarMouseControl
		if ecs.FetchComponent(entity, &mouseControl) {
			s.mouseOfInterest = true
			s.updateMouse(elapsedSeconds, entity)
		}

		var gamepadControl *CarGamepadControl
		if ecs.FetchComponent(entity, &gamepadControl) {
			s.updateGamepad(elapsedSeconds, entity)
		}

		s.updateCar(elapsedSeconds, entity)
	}
}

func (s *CarSystem) updateKeyboard(elapsedSeconds float64, entity *ecs.Entity) {
	var carComp *CarComponent
	ecs.FetchComponent(entity, &carComp)
	var keyboardComp *CarKeyboardControl
	ecs.FetchComponent(entity, &keyboardComp)

	s.keysOfInterest[keyboardComp.AccelerateKey] = struct{}{}
	s.keysOfInterest[keyboardComp.DecelerateKey] = struct{}{}
	s.keysOfInterest[keyboardComp.TurnLeftKey] = struct{}{}
	s.keysOfInterest[keyboardComp.TurnRightKey] = struct{}{}
	s.keysOfInterest[keyboardComp.ShiftUpKey] = struct{}{}
	s.keysOfInterest[keyboardComp.ShiftDownKey] = struct{}{}
	s.keysOfInterest[keyboardComp.RecoverKey] = struct{}{}

	if s.keyStates[keyboardComp.AccelerateKey] {
		carComp.Acceleration += elapsedSeconds * keyboardComp.AccelerationChangeSpeed
	} else {
		carComp.Acceleration -= elapsedSeconds * keyboardComp.AccelerationChangeSpeed
	}
	carComp.Acceleration = dprec.Clamp(carComp.Acceleration, 0.0, 1.0)

	if s.keyStates[keyboardComp.DecelerateKey] {
		carComp.Deceleration += elapsedSeconds * keyboardComp.DecelerationChangeSpeed
	} else {
		carComp.Deceleration -= elapsedSeconds * keyboardComp.DecelerationChangeSpeed
	}
	carComp.Deceleration = dprec.Clamp(carComp.Deceleration, 0.0, 1.0)

	autoMaxSteeringAmount := 1.0 / (1.0 + 0.05*carComp.Car.Velocity())
	switch {
	case s.keyStates[keyboardComp.TurnLeftKey] == s.keyStates[keyboardComp.TurnRightKey]:
		if keyboardComp.SteeringAmount > 0.0 {
			keyboardComp.SteeringAmount -= elapsedSeconds * keyboardComp.SteeringRestoreSpeed
			keyboardComp.SteeringAmount = dprec.Max(0.0, keyboardComp.SteeringAmount)
		}
		if keyboardComp.SteeringAmount < 0.0 {
			keyboardComp.SteeringAmount += elapsedSeconds * keyboardComp.SteeringRestoreSpeed
			keyboardComp.SteeringAmount = dprec.Min(0.0, keyboardComp.SteeringAmount)
		}
	case s.keyStates[keyboardComp.TurnLeftKey]:
		keyboardComp.SteeringAmount -= elapsedSeconds * keyboardComp.SteeringChangeSpeed
		keyboardComp.SteeringAmount = dprec.Max(keyboardComp.SteeringAmount, -autoMaxSteeringAmount)
	case s.keyStates[keyboardComp.TurnRightKey]:
		keyboardComp.SteeringAmount += elapsedSeconds * keyboardComp.SteeringChangeSpeed
		keyboardComp.SteeringAmount = dprec.Min(keyboardComp.SteeringAmount, autoMaxSteeringAmount)
	}

	maxSteeringAngle := carComp.Car.Axes()[0].MaxSteeringAngle()
	steeringAngle := maxSteeringAngle * dprec.Angle(keyboardComp.SteeringAmount)

	carDirection := carComp.Car.Chassis().Body().Orientation().OrientationZ()
	carDirection.Y = 0.0

	carActualDirection := carComp.Car.Chassis().Body().Velocity()
	carActualDirection.Y = 0.0

	recoverSin := dprec.Vec3Cross(
		dprec.UnitVec3(carDirection),
		dprec.UnitVec3(carActualDirection),
	)
	var recoverAngle dprec.Angle
	if (carActualDirection.Length() > 10.0) && (recoverSin.Length() > 0.000001) {
		recoverAngle = -dprec.Angle(dprec.Sign(recoverSin.Y)) * dprec.Asin(recoverSin.Length())
		recoverAngle = dprec.Clamp(recoverAngle, -maxSteeringAngle, maxSteeringAngle)
	}

	recoverAngle = recoverAngle / 1.5
	carComp.SteeringAmount = float64(dprec.Clamp(steeringAngle+recoverAngle, -maxSteeringAngle, maxSteeringAngle) / maxSteeringAngle)

	if s.keyStates[keyboardComp.ShiftDownKey] {
		carComp.Gear = CarGearReverse
	}
	if s.keyStates[keyboardComp.ShiftUpKey] {
		carComp.Gear = CarGearForward
	}
	carComp.Recover = s.keyStates[keyboardComp.RecoverKey]
}

func (s *CarSystem) updateMouse(elapsedSeconds float64, entity *ecs.Entity) {
	var carComp *CarComponent
	ecs.FetchComponent(entity, &carComp)
	var mouseComp *CarMouseControl
	ecs.FetchComponent(entity, &mouseComp)

	if s.mouseButtonStates[ui.MouseButtonLeft] {
		carComp.Acceleration += elapsedSeconds * mouseComp.AccelerationChangeSpeed
	} else {
		carComp.Acceleration -= elapsedSeconds * mouseComp.AccelerationChangeSpeed
	}
	carComp.Acceleration = dprec.Clamp(carComp.Acceleration, 0.0, 1.0)

	if s.mouseButtonStates[ui.MouseButtonRight] {
		carComp.Deceleration += elapsedSeconds * mouseComp.DecelerationChangeSpeed
	} else {
		carComp.Deceleration -= elapsedSeconds * mouseComp.DecelerationChangeSpeed
	}
	carComp.Deceleration = dprec.Clamp(carComp.Deceleration, 0.0, 1.0)

	const epsilon = 1.0
	if s.mouseScroll < -epsilon {
		carComp.Gear = CarGearReverse
	}
	if s.mouseScroll > epsilon {
		carComp.Gear = CarGearForward
	}
	s.mouseScroll = 0

	carComp.Recover = s.mouseButtonStates[ui.MouseButtonMiddle]

	chassis := carComp.Car.Chassis()
	position := chassis.Body().Position()

	camera := s.gfxScene.ActiveCamera()
	viewport := graphics.Viewport{
		Width:  s.mouseAreaWidth,
		Height: s.mouseAreaHeight,
	}
	start, end := s.gfxScene.Ray(viewport, camera, s.mouseX, s.mouseY)

	line := collision.NewLine(start, end)

	intersection, ok := collision.LineWithSurfaceIntersectionPoint(line, position, dprec.BasisYVec3())
	if !ok {
		sphere := collision.NewSphere(position, 1000.0)
		_, intersection, ok = collision.LineWithSphereIntersectionPoints(line, sphere)
	}
	if ok {
		delta := dprec.Vec3Diff(intersection, position)
		delta.Y = 0.0

		forward := chassis.Body().Orientation().OrientationZ()
		forward.Y = 0.0

		sin := dprec.Vec3Cross(
			dprec.UnitVec3(forward),
			dprec.UnitVec3(delta),
		)
		angle := dprec.Angle(dprec.Sign(sin.Y)) * dprec.Asin(sin.Length())

		maxSteeringAngle := carComp.Car.Axes()[0].MaxSteeringAngle()
		carComp.SteeringAmount = float64(dprec.Clamp(-dprec.Angle(angle), -maxSteeringAngle, maxSteeringAngle) / maxSteeringAngle)

	} else {
		carComp.SteeringAmount = 0.0
	}
}

func (s *CarSystem) updateGamepad(elapsedSeconds float64, entity *ecs.Entity) {
	var carComp *CarComponent
	ecs.FetchComponent(entity, &carComp)
	var gamepadComp *CarGamepadControl
	ecs.FetchComponent(entity, &gamepadComp)

	gamepad := gamepadComp.Gamepad
	if !gamepad.Connected() || !gamepad.Supported() {
		return
	}

	leftStickX := gamepad.LeftStickX()
	carComp.SteeringAmount = leftStickX * leftStickX * leftStickX
	carComp.Acceleration = gamepad.RightTrigger()
	carComp.Deceleration = gamepad.LeftTrigger()
	if gamepad.ActionLeftButton() {
		carComp.Gear = CarGearReverse
	}
	if gamepad.ActionDownButton() {
		carComp.Gear = CarGearForward
	}
	carComp.Recover = gamepad.ActionUpButton()
}

func (s *CarSystem) updateCar(elapsedSeconds float64, entity *ecs.Entity) {
	// TODO: Run this inside physics loop for smooth operation.

	var carComp *CarComponent
	ecs.FetchComponent(entity, &carComp)
	var (
		car         = carComp.Car
		chassisBody = car.Chassis().Body()
	)

	for _, light := range carComp.Car.Chassis().HeadLights() {
		light.SetActive(carComp.LightsOn)
	}
	for _, light := range carComp.Car.Chassis().BeamLights() {
		light.SetActive(carComp.LightsOn)
	}
	for _, light := range carComp.Car.Chassis().TailLights() {
		light.SetActive(carComp.LightsOn)
	}
	for _, light := range carComp.Car.Chassis().StopLights() {
		light.SetActive(carComp.Deceleration > 0.1)
	}

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
		if carComp.Gear == CarGearForward {
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
