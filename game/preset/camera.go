package preset

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/ui"
)

type GamepadStateProvider interface {
	GamepadState(index int) (app.GamepadState, bool)
}

func NewYawPitchCameraSystem(ecsScene *ecs.Scene, gamepadProvider GamepadStateProvider) *YawPitchCameraSystem {
	return &YawPitchCameraSystem{
		ecsScene:        ecsScene,
		gamepadProvider: gamepadProvider,
	}
}

type YawPitchCameraSystem struct {
	ecsScene        *ecs.Scene
	gamepadProvider GamepadStateProvider

	translationSpeed float64
	rotationSpeed    dprec.Angle

	hasKeyboardConsumer bool

	keyMoveForward  ui.KeyCode
	keyMoveBackward ui.KeyCode
	keyMoveLeft     ui.KeyCode
	keyMoveRight    ui.KeyCode
	keyMoveUp       ui.KeyCode
	keyMoveDown     ui.KeyCode
	keyLookUp       ui.KeyCode
	keyLookDown     ui.KeyCode
	keyLookLeft     ui.KeyCode
	keyLookRight    ui.KeyCode

	isMoveForward  bool
	isMoveBackward bool
	isMoveLeft     bool
	isMoveRight    bool
	isMoveUp       bool
	isMoveDown     bool
	isLookUp       bool
	isLookDown     bool
	isLookLeft     bool
	isLookRight    bool
}

func (s *YawPitchCameraSystem) UseDefaults() {
	s.translationSpeed = 20.0
	s.rotationSpeed = dprec.Degrees(120.0)

	s.keyMoveForward = ui.KeyCodeW
	s.keyMoveBackward = ui.KeyCodeS
	s.keyMoveLeft = ui.KeyCodeA
	s.keyMoveRight = ui.KeyCodeD
	s.keyMoveUp = ui.KeyCodeSpace
	s.keyMoveDown = ui.KeyCodeLeftShift
	s.keyLookUp = ui.KeyCodeArrowUp
	s.keyLookDown = ui.KeyCodeArrowDown
	s.keyLookLeft = ui.KeyCodeArrowLeft
	s.keyLookRight = ui.KeyCodeArrowRight
}

func (s *YawPitchCameraSystem) OnKeyboardEvent(event ui.KeyboardEvent) bool {
	if !s.hasKeyboardConsumer {
		return false
	}
	active := event.Type != ui.KeyboardEventTypeKeyUp
	switch event.Code {
	case s.keyMoveForward:
		s.isMoveForward = active
		return true
	case s.keyMoveBackward:
		s.isMoveBackward = active
		return true
	case s.keyMoveLeft:
		s.isMoveLeft = active
		return true
	case s.keyMoveRight:
		s.isMoveRight = active
		return true
	case s.keyMoveUp:
		s.isMoveUp = active
		return true
	case s.keyMoveDown:
		s.isMoveDown = active
		return true
	case s.keyLookUp:
		s.isLookUp = active
		return true
	case s.keyLookDown:
		s.isLookDown = active
		return true
	case s.keyLookLeft:
		s.isLookLeft = active
		return true
	case s.keyLookRight:
		s.isLookRight = active
		return true
	default:
		return false
	}
}

func (s *YawPitchCameraSystem) Update(elapsedSeconds float64) {
	result := s.ecsScene.Find(ecs.
		Having(NodeComponentID).
		And(YawPitchCameraComponentID).
		And(ControlledComponentID),
	)
	defer result.Close()

	var hasKeyboardConsumer bool
	var entity *ecs.Entity
	for result.FetchNext(&entity) {
		var controlled *ControlledComponent
		ecs.FetchComponent(entity, &controlled)
		if controlled.Inputs.Is(ControlInputKeyboard) {
			hasKeyboardConsumer = true
			s.updateKeyboard(elapsedSeconds, entity)
		}
		if controlled.Inputs.Is(ControlInputGamepad0) {
			if gamepad, ok := s.gamepadProvider.GamepadState(0); ok {
				s.updateGamepad(elapsedSeconds, gamepad, entity)
			}
		}
	}
	s.hasKeyboardConsumer = hasKeyboardConsumer
}

func (s *YawPitchCameraSystem) updateKeyboard(elapsedSeconds float64, entity *ecs.Entity) {
	var nodeComp *NodeComponent
	ecs.FetchComponent(entity, &nodeComp)
	var cameraComp *YawPitchCameraComponent
	ecs.FetchComponent(entity, &cameraComp)

	oldTranslation, _, oldScale := nodeComp.Node.AbsoluteMatrix().TRS()

	var deltaPitch dprec.Angle
	if s.isLookUp {
		deltaPitch += dprec.Angle(elapsedSeconds)
	}
	if s.isLookDown {
		deltaPitch -= dprec.Angle(elapsedSeconds)
	}
	cameraComp.PitchAngle += deltaPitch * s.rotationSpeed
	pitchRotation := dprec.RotationQuat(cameraComp.PitchAngle, dprec.BasisXVec3())

	var deltaYaw dprec.Angle
	if s.isLookLeft {
		deltaYaw += dprec.Angle(elapsedSeconds)
	}
	if s.isLookRight {
		deltaYaw -= dprec.Angle(elapsedSeconds)
	}
	cameraComp.YawAngle += deltaYaw * s.rotationSpeed
	yawRotation := dprec.RotationQuat(cameraComp.YawAngle, dprec.BasisYVec3())

	var deltaTranslation dprec.Vec3
	if s.isMoveForward {
		deltaTranslation.Z -= 1.0
	}
	if s.isMoveBackward {
		deltaTranslation.Z += 1.0
	}
	if s.isMoveLeft {
		deltaTranslation.X -= 1.0
	}
	if s.isMoveRight {
		deltaTranslation.X += 1.0
	}
	if s.isMoveUp {
		deltaTranslation.Y += 1.0
	}
	if s.isMoveDown {
		deltaTranslation.Y -= 1.0
	}
	if deltaTranslation.Length() < 0.1 {
		deltaTranslation = dprec.ZeroVec3()
	} else {
		deltaTranslation = dprec.ResizedVec3(deltaTranslation, s.translationSpeed*elapsedSeconds)
	}
	deltaTranslation = dprec.QuatVec3Rotation(yawRotation, deltaTranslation)

	nodeComp.Node.SetAbsoluteMatrix(dprec.TRSMat4(
		dprec.Vec3Sum(oldTranslation, deltaTranslation),
		dprec.QuatProd(yawRotation, pitchRotation),
		oldScale,
	))
}

func (s *YawPitchCameraSystem) updateGamepad(elapsedSeconds float64, gamepad app.GamepadState, entity *ecs.Entity) {
	var nodeComp *NodeComponent
	ecs.FetchComponent(entity, &nodeComp)
	var cameraComp *YawPitchCameraComponent
	ecs.FetchComponent(entity, &cameraComp)

	oldTranslation, _, oldScale := nodeComp.Node.AbsoluteMatrix().TRS()

	deltaRotation := dprec.Vec2{
		X: gamepad.RightStickY,
		Y: gamepad.RightStickX,
	}
	if deltaRotation.Length() < 0.1 {
		deltaRotation = dprec.ZeroVec2()
	}
	cameraComp.PitchAngle += dprec.Angle(deltaRotation.X*elapsedSeconds) * s.rotationSpeed
	pitchRotation := dprec.RotationQuat(cameraComp.PitchAngle, dprec.BasisXVec3())
	cameraComp.YawAngle -= dprec.Angle(deltaRotation.Y*elapsedSeconds) * s.rotationSpeed
	yawRotation := dprec.RotationQuat(cameraComp.YawAngle, dprec.BasisYVec3())

	deltaTranslation := dprec.Vec3{
		X: gamepad.LeftStickX,
		Y: gamepad.RightTrigger - gamepad.LeftTrigger,
		Z: -gamepad.LeftStickY,
	}
	if deltaTranslation.Length() < 0.1 {
		deltaTranslation = dprec.ZeroVec3()
	} else {
		deltaTranslation = dprec.Vec3Prod(deltaTranslation, s.translationSpeed*elapsedSeconds)
	}
	deltaTranslation = dprec.QuatVec3Rotation(yawRotation, deltaTranslation)

	nodeComp.Node.SetAbsoluteMatrix(dprec.TRSMat4(
		dprec.Vec3Sum(oldTranslation, deltaTranslation),
		dprec.QuatProd(yawRotation, pitchRotation),
		oldScale,
	))
}

func NewFollowCameraSystem(ecsScene *ecs.Scene, gamepadProvider GamepadStateProvider) *FollowCameraSystem {
	return &FollowCameraSystem{
		ecsScene:        ecsScene,
		gamepadProvider: gamepadProvider,
	}
}

type FollowCameraSystem struct {
	ecsScene        *ecs.Scene
	gamepadProvider GamepadStateProvider

	rotationSpeed    dprec.Angle
	rotationStrength dprec.Angle
	zoomSpeed        float64

	hasKeyboardConsumer bool

	keyRotateLeft  ui.KeyCode
	keyRotateRight ui.KeyCode
	keyRotateUp    ui.KeyCode
	keyRotateDown  ui.KeyCode
	keyZoomIn      ui.KeyCode
	keyZoomOut     ui.KeyCode

	isRotateLeft  bool
	isRotateRight bool
	isRotateUp    bool
	isRotateDown  bool
	isZoomIn      bool
	isZoomOut     bool
}

func (s *FollowCameraSystem) UseDefaults() {
	s.rotationSpeed = dprec.Degrees(100)
	s.rotationStrength = dprec.Degrees(300.0)
	s.zoomSpeed = 1.0

	s.keyRotateUp = ui.KeyCodeW
	s.keyRotateDown = ui.KeyCodeS
	s.keyRotateLeft = ui.KeyCodeA
	s.keyRotateRight = ui.KeyCodeD
	s.keyZoomIn = ui.KeyCodeE
	s.keyZoomOut = ui.KeyCodeQ
}

func (s *FollowCameraSystem) OnKeyboardEvent(event ui.KeyboardEvent) bool {
	if !s.hasKeyboardConsumer {
		return false
	}
	active := event.Type != ui.KeyboardEventTypeKeyUp
	switch event.Code {
	case s.keyRotateUp:
		s.isRotateUp = active
		return true
	case s.keyRotateDown:
		s.isRotateDown = active
		return true
	case s.keyRotateLeft:
		s.isRotateLeft = active
		return true
	case s.keyRotateRight:
		s.isRotateRight = active
		return true
	case s.keyZoomIn:
		s.isZoomIn = active
		return true
	case s.keyZoomOut:
		s.isZoomOut = active
		return true
	default:
		return false
	}
}

func (s *FollowCameraSystem) Update(elapsedSeconds float64) {
	result := s.ecsScene.Find(ecs.
		Having(NodeComponentID).
		And(FollowCameraComponentID),
	)
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
			if controlled.Inputs.Is(ControlInputGamepad0) {
				if gamepad, ok := s.gamepadProvider.GamepadState(0); ok {
					s.updateGamepad(elapsedSeconds, gamepad, entity)
				}
			}
		}
		s.updateCamera(elapsedSeconds, entity)
	}
	s.hasKeyboardConsumer = hasKeyboardConsumer
}

func (s *FollowCameraSystem) updateKeyboard(elapsedSeconds float64, entity *ecs.Entity) {
	var cameraComp *FollowCameraComponent
	ecs.FetchComponent(entity, &cameraComp)

	if s.isRotateUp {
		cameraComp.PitchAngle -= s.rotationSpeed * dprec.Angle(elapsedSeconds)
	}
	if s.isRotateDown {
		cameraComp.PitchAngle += s.rotationSpeed * dprec.Angle(elapsedSeconds)
	}

	if s.isRotateLeft {
		cameraComp.YawAngle -= s.rotationSpeed * dprec.Angle(elapsedSeconds)
	}
	if s.isRotateRight {
		cameraComp.YawAngle += s.rotationSpeed * dprec.Angle(elapsedSeconds)
	}

	if s.isZoomIn {
		cameraComp.Zoom -= cameraComp.Zoom * elapsedSeconds
	}
	if s.isZoomOut {
		cameraComp.Zoom += cameraComp.Zoom * elapsedSeconds
	}
}

func (s *FollowCameraSystem) updateGamepad(elapsedSeconds float64, gamepad app.GamepadState, entity *ecs.Entity) {
	var cameraComp *FollowCameraComponent
	ecs.FetchComponent(entity, &cameraComp)

	if gamepad.DpadUpButton {
		cameraComp.PitchAngle -= s.rotationSpeed * dprec.Angle(elapsedSeconds)
	}
	if gamepad.DpadDownButton {
		cameraComp.PitchAngle += s.rotationSpeed * dprec.Angle(elapsedSeconds)
	}

	if gamepad.DpadLeftButton {
		cameraComp.YawAngle -= s.rotationSpeed * dprec.Angle(elapsedSeconds)
	}
	if gamepad.DpadRightButton {
		cameraComp.YawAngle += s.rotationSpeed * dprec.Angle(elapsedSeconds)
	}

	if gamepad.RightBumper {
		cameraComp.Zoom -= cameraComp.Zoom * elapsedSeconds
	}
	if gamepad.LeftBumper {
		cameraComp.Zoom += cameraComp.Zoom * elapsedSeconds
	}

	if dprec.Abs(gamepad.RightStickX) > 0.1 || dprec.Abs(gamepad.RightStickY) > 0.1 {
		target := cameraComp.Target
		targetPosition := target.AbsoluteMatrix().Translation()
		anchorVector := dprec.Vec3Diff(cameraComp.AnchorPosition, targetPosition)

		cameraVectorZ := anchorVector
		cameraVectorX := dprec.Vec3Cross(dprec.BasisYVec3(), cameraVectorZ)
		cameraVectorY := dprec.Vec3Cross(cameraVectorZ, cameraVectorX)

		if dprec.Abs(gamepad.RightStickY) > 0.1 {
			angle := s.rotationStrength * dprec.Angle(gamepad.RightStickY*elapsedSeconds)
			anchorVector = dprec.QuatVec3Rotation(dprec.RotationQuat(angle, cameraVectorX), anchorVector)
		}
		if dprec.Abs(gamepad.RightStickX) > 0.1 {
			angle := -s.rotationStrength * dprec.Angle(gamepad.RightStickX*elapsedSeconds)
			rotation := dprec.RotationQuat(angle, cameraVectorY)
			anchorVector = dprec.QuatVec3Rotation(rotation, anchorVector)
		}

		cameraComp.AnchorPosition = dprec.Vec3Sum(targetPosition, anchorVector)
	}
}

func (s *FollowCameraSystem) updateCamera(elapsedSeconds float64, entity *ecs.Entity) {
	var nodeComp *NodeComponent
	ecs.FetchComponent(entity, &nodeComp)
	var cameraComp *FollowCameraComponent
	ecs.FetchComponent(entity, &cameraComp)

	target := cameraComp.Target
	targetPosition := target.AbsoluteMatrix().Translation()

	// We use a camera anchor to achieve the smooth effect of a
	// camera following the target.
	anchorVector := dprec.Vec3Diff(cameraComp.AnchorPosition, targetPosition)
	anchorVector = dprec.ResizedVec3(anchorVector, cameraComp.AnchorDistance)
	cameraComp.AnchorPosition = dprec.Vec3Sum(targetPosition, anchorVector)

	cameraVectorZ := anchorVector
	cameraVectorX := dprec.Vec3Cross(dprec.BasisYVec3(), cameraVectorZ)
	cameraVectorY := dprec.Vec3Cross(cameraVectorZ, cameraVectorX)

	matrix := dprec.Mat4MultiProd(
		dprec.TransformationMat4(
			dprec.UnitVec3(cameraVectorX),
			dprec.UnitVec3(cameraVectorY),
			dprec.UnitVec3(cameraVectorZ),
			targetPosition,
		),
		dprec.RotationMat4(cameraComp.YawAngle, 0.0, 1.0, 0.0),
		dprec.RotationMat4(cameraComp.PitchAngle, 1.0, 0.0, 0.0),
		dprec.TranslationMat4(0.0, 0.0, cameraComp.CameraDistance*cameraComp.Zoom),
	)
	nodeComp.Node.SetAbsoluteMatrix(matrix)
}
