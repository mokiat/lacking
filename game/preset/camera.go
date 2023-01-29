package preset

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/ui"
)

func NewYawPitchCameraSystem(ecsScene *ecs.Scene) *YawPitchCameraSystem {
	return &YawPitchCameraSystem{
		ecsScene: ecsScene,
	}
}

type YawPitchCameraSystem struct {
	ecsScene *ecs.Scene

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
	s.translationSpeed = 15.0
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
	active := event.Type != app.KeyboardEventTypeKeyUp
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
