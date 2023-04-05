package preset

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/constraint"
)

var (
	CarComponentID = ecs.NewComponentTypeID()
)

type CarDirection int

const (
	CarDirectionNeutral CarDirection = iota
	CarDirectionForward
	CarDirectionReverse
)

type CarComponent struct {
	Car            *Car
	Direction      CarDirection
	SteeringAmount float64
	Acceleration   float64
	Deceleration   float64
	Recover        bool
}

func (*CarComponent) TypeID() ecs.ComponentTypeID {
	return CarComponentID
}

func NewChassisDefinition() *ChassisDefinition {
	return &ChassisDefinition{}
}

type ChassisDefinition struct {
	nodeName string
	bodyDef  *physics.BodyDefinition
}

func (d *ChassisDefinition) WithNodeName(name string) *ChassisDefinition {
	d.nodeName = name
	return d
}

func (d *ChassisDefinition) WithBodyDefinition(def *physics.BodyDefinition) *ChassisDefinition {
	d.bodyDef = def
	return d
}

func NewWheelDefinition() *WheelDefinition {
	return &WheelDefinition{}
}

type WheelDefinition struct {
	nodeName string
	bodyDef  *physics.BodyDefinition
}

func (d *WheelDefinition) WithNodeName(name string) *WheelDefinition {
	d.nodeName = name
	return d
}

func (d *WheelDefinition) WithBodyDefinition(def *physics.BodyDefinition) *WheelDefinition {
	d.bodyDef = def
	return d
}

func NewHubDefinition() *HubDefinition {
	return &HubDefinition{}
}

type HubDefinition struct {
	nodeName string
	bodyDef  *physics.BodyDefinition
}

func (d *HubDefinition) WithNodeName(name string) *HubDefinition {
	d.nodeName = name
	return d
}

func (d *HubDefinition) WithBodyDefinition(def *physics.BodyDefinition) *HubDefinition {
	d.bodyDef = def
	return d
}

func NewAxisDefinition() *AxisDefinition {
	return &AxisDefinition{
		position:         dprec.ZeroVec3(),
		width:            2.0,
		suspensionLength: 0.5,
	}
}

type AxisDefinition struct {
	position         dprec.Vec3
	width            float64
	suspensionLength float64
	springLength     float64
	springFrequency  float64
	springDamping    float64
	maxSteeringAngle dprec.Angle
	maxAcceleration  float64
	maxBraking       float64
	reverseRatio     float64
	leftWheelDef     *WheelDefinition
	rightWheelDef    *WheelDefinition
	leftHubDef       *HubDefinition
	rightHubDef      *HubDefinition
}

func (d *AxisDefinition) WithPosition(position dprec.Vec3) *AxisDefinition {
	d.position = position
	return d
}

func (d *AxisDefinition) WithWidth(width float64) *AxisDefinition {
	d.width = width
	return d
}

func (d *AxisDefinition) WithSuspensionLength(length float64) *AxisDefinition {
	d.suspensionLength = length
	return d
}

func (d *AxisDefinition) WithSpringLength(length float64) *AxisDefinition {
	d.springLength = length
	return d
}

func (d *AxisDefinition) WithSpringFrequency(frequency float64) *AxisDefinition {
	d.springFrequency = frequency
	return d
}

func (d *AxisDefinition) WithSpringDamping(damping float64) *AxisDefinition {
	d.springDamping = damping
	return d
}

func (d *AxisDefinition) WithLeftWheelDefinition(def *WheelDefinition) *AxisDefinition {
	d.leftWheelDef = def
	return d
}

func (d *AxisDefinition) WithRightWheelDefinition(def *WheelDefinition) *AxisDefinition {
	d.rightWheelDef = def
	return d
}

func (d *AxisDefinition) WithLeftHubDefinition(def *HubDefinition) *AxisDefinition {
	d.leftHubDef = def
	return d
}

func (d *AxisDefinition) WithRightHubDefinition(def *HubDefinition) *AxisDefinition {
	d.rightHubDef = def
	return d
}

func (d *AxisDefinition) WithMaxSteeringAngle(maxAngle dprec.Angle) *AxisDefinition {
	d.maxSteeringAngle = maxAngle
	return d
}

func (d *AxisDefinition) WithMaxAcceleration(maxAcceleration float64) *AxisDefinition {
	d.maxAcceleration = maxAcceleration
	return d
}

func (d *AxisDefinition) WithMaxBraking(maxBraking float64) *AxisDefinition {
	d.maxBraking = maxBraking
	return d
}

func (d *AxisDefinition) WithReverseRatio(ratio float64) *AxisDefinition {
	d.reverseRatio = ratio
	return d
}

func NewCarDefinition() *CarDefinition {
	return &CarDefinition{}
}

type CarDefinition struct {
	chassisDef *ChassisDefinition
	axesDef    []*AxisDefinition
}

func (d *CarDefinition) WithChassisDefinition(def *ChassisDefinition) *CarDefinition {
	d.chassisDef = def
	return d
}

func (d *CarDefinition) WithAxisDefinition(def *AxisDefinition) *CarDefinition {
	d.axesDef = append(d.axesDef, def)
	return d
}

func (d *CarDefinition) ApplyToModel(scene *game.Scene, info CarApplyInfo) *Car {
	chassisNode := info.Model.FindNode(d.chassisDef.nodeName)
	chassisPosition := info.Position
	chassisRotation := info.Rotation

	chassisBody := scene.Physics().CreateBody(physics.BodyInfo{
		Name:       d.chassisDef.nodeName,
		Definition: d.chassisDef.bodyDef,
		Position:   chassisPosition,
		Rotation:   chassisRotation,
		IsDynamic:  true,
	})
	chassisNode.SetBody(chassisBody)

	var axes []*Axis
	for _, axisDef := range d.axesDef {
		springOffset := dprec.NewVec3(0.0, -axisDef.springLength, 0.0)

		leftWheelRelativePosition := dprec.Vec3Sum(
			axisDef.position,
			dprec.NewVec3(axisDef.width/2.0, 0.0, 0.0),
		)
		leftWheelAbsolutePosition := dprec.Vec3Sum(
			chassisPosition,
			dprec.QuatVec3Rotation(chassisRotation, leftWheelRelativePosition),
		)

		leftWheelNode := info.Model.FindNode(axisDef.leftWheelDef.nodeName)
		leftWheelBody := scene.Physics().CreateBody(physics.BodyInfo{
			Name:       axisDef.leftWheelDef.nodeName,
			Definition: axisDef.leftWheelDef.bodyDef,
			Position:   leftWheelAbsolutePosition,
			Rotation:   chassisRotation,
			IsDynamic:  true,
		})
		leftWheelNode.SetBody(leftWheelBody)

		leftWheelDirection := constraint.NewMatchDirections().
			SetPrimaryDirection(dprec.BasisXVec3()).
			SetSecondaryDirection(dprec.BasisXVec3())
		leftWheelAttachment := scene.Physics().CreateDoubleBodyConstraint(chassisBody, leftWheelBody, constraint.NewPairCombined(
			constraint.NewMatchDirectionOffset().
				SetPrimaryRadius(leftWheelRelativePosition).
				SetSecondaryRadius(dprec.ZeroVec3()).
				SetDirection(dprec.BasisXVec3()).
				SetOffset(0.0),
			constraint.NewMatchDirectionOffset().
				SetPrimaryRadius(leftWheelRelativePosition).
				SetSecondaryRadius(dprec.ZeroVec3()).
				SetDirection(dprec.BasisZVec3()).
				SetOffset(0.0),
			constraint.NewClampDirectionOffset().
				SetDirection(dprec.BasisYVec3()).
				SetMax(axisDef.position.Y).
				SetMin(axisDef.position.Y-axisDef.suspensionLength).
				SetRestitution(0.0),
			constraint.NewCoilover().
				SetPrimaryRadius(dprec.Vec3Sum(leftWheelRelativePosition, springOffset)).
				SetSecondaryRadius(dprec.ZeroVec3()).
				SetFrequency(axisDef.springFrequency).
				SetDamping(axisDef.springDamping),
			leftWheelDirection,
		))

		var leftHub *Hub
		if hubDef := axisDef.leftHubDef; hubDef != nil {
			hubNode := info.Model.FindNode(hubDef.nodeName)
			hubBody := scene.Physics().CreateBody(physics.BodyInfo{
				Name:       hubDef.nodeName,
				Definition: hubDef.bodyDef,
				Position:   leftWheelAbsolutePosition,
				Rotation:   chassisRotation,
				IsDynamic:  true,
			})
			hubNode.SetBody(hubBody)

			scene.Physics().CreateDoubleBodyConstraint(hubBody, leftWheelBody, constraint.NewPairCombined(
				constraint.NewCopyPosition(),
				constraint.NewCopyDirection().
					SetPrimaryDirection(dprec.BasisXVec3()).
					SetSecondaryDirection(dprec.BasisXVec3()),
			))
			scene.Physics().CreateDoubleBodyConstraint(hubBody, chassisBody,
				constraint.NewCopyDirection().
					SetPrimaryDirection(dprec.BasisYVec3()).
					SetSecondaryDirection(dprec.BasisYVec3()),
			)

			leftHub = &Hub{
				node: hubNode,
				body: hubBody,
			}
		}

		rightWheelRelativePosition := dprec.Vec3Sum(
			axisDef.position,
			dprec.NewVec3(-axisDef.width/2.0, 0.0, 0.0),
		)
		rightWheelAbsolutePosition := dprec.Vec3Sum(
			chassisPosition,
			dprec.QuatVec3Rotation(chassisRotation, rightWheelRelativePosition),
		)

		rightWheelNode := info.Model.FindNode(axisDef.rightWheelDef.nodeName)
		rightWheelBody := scene.Physics().CreateBody(physics.BodyInfo{
			Name:       axisDef.rightWheelDef.nodeName,
			Definition: axisDef.rightWheelDef.bodyDef,
			Position:   rightWheelAbsolutePosition,
			Rotation:   chassisRotation,
			IsDynamic:  true,
		})
		rightWheelNode.SetBody(rightWheelBody)

		rightWheelDirection := constraint.NewMatchDirections().
			SetPrimaryDirection(dprec.BasisXVec3()).
			SetSecondaryDirection(dprec.BasisXVec3())
		rightWheelAttachment := scene.Physics().CreateDoubleBodyConstraint(chassisBody, rightWheelBody, constraint.NewPairCombined(
			constraint.NewMatchDirectionOffset().
				SetPrimaryRadius(rightWheelRelativePosition).
				SetSecondaryRadius(dprec.ZeroVec3()).
				SetDirection(dprec.BasisXVec3()).
				SetOffset(0.0),
			constraint.NewMatchDirectionOffset().
				SetPrimaryRadius(rightWheelRelativePosition).
				SetSecondaryRadius(dprec.ZeroVec3()).
				SetDirection(dprec.BasisZVec3()).
				SetOffset(0.0),
			constraint.NewClampDirectionOffset().
				SetDirection(dprec.BasisYVec3()).
				SetMax(axisDef.position.Y).
				SetMin(axisDef.position.Y-axisDef.suspensionLength).
				SetRestitution(0.0),
			constraint.NewCoilover().
				SetPrimaryRadius(dprec.Vec3Sum(rightWheelRelativePosition, springOffset)).
				SetSecondaryRadius(dprec.ZeroVec3()).
				SetFrequency(axisDef.springFrequency).
				SetDamping(axisDef.springDamping),
			rightWheelDirection,
		))

		var rightHub *Hub
		if hubDef := axisDef.rightHubDef; hubDef != nil {
			hubNode := info.Model.FindNode(hubDef.nodeName)
			hubBody := scene.Physics().CreateBody(physics.BodyInfo{
				Name:       hubDef.nodeName,
				Definition: hubDef.bodyDef,
				Position:   rightWheelAbsolutePosition,
				Rotation:   chassisRotation,
				IsDynamic:  true,
			})
			hubNode.SetBody(hubBody)

			scene.Physics().CreateDoubleBodyConstraint(hubBody, rightWheelBody, constraint.NewPairCombined(
				constraint.NewCopyPosition(),
				constraint.NewCopyDirection().
					SetPrimaryDirection(dprec.BasisXVec3()).
					SetSecondaryDirection(dprec.BasisXVec3()),
			))
			scene.Physics().CreateDoubleBodyConstraint(hubBody, chassisBody,
				constraint.NewCopyDirection().
					SetPrimaryDirection(dprec.BasisYVec3()).
					SetSecondaryDirection(dprec.BasisYVec3()),
			)

			rightHub = &Hub{
				node: hubNode,
				body: hubBody,
			}
		}

		scene.Physics().CreateDoubleBodyConstraint(leftWheelBody, rightWheelBody, constraint.NewDifferential())

		axes = append(axes, &Axis{
			maxSteeringAngle: axisDef.maxSteeringAngle,
			maxAcceleration:  axisDef.maxAcceleration,
			maxBraking:       axisDef.maxBraking,
			reverseRatio:     axisDef.reverseRatio,
			leftWheel: &Wheel{
				node:                 leftWheelNode,
				body:                 leftWheelBody,
				directionSolver:      leftWheelDirection,
				attachmentConstraint: leftWheelAttachment,
			},
			rightWheel: &Wheel{
				node:                 rightWheelNode,
				body:                 rightWheelBody,
				directionSolver:      rightWheelDirection,
				attachmentConstraint: rightWheelAttachment,
			},
			leftHub:  leftHub,
			rightHub: rightHub,
		})
	}

	entity := scene.ECS().CreateEntity()
	ecs.AttachComponent(entity, &NodeComponent{
		Node: chassisNode,
	})
	ecs.AttachComponent(entity, &ControlledComponent{
		Inputs: info.Inputs,
	})
	result := &Car{
		chassis: &Chassis{
			node: chassisNode,
			body: chassisBody,
		},
		axes:   axes,
		entity: entity,
	}
	ecs.AttachComponent(entity, &CarComponent{
		Car:            result,
		Direction:      CarDirectionForward,
		SteeringAmount: 0.0,
		Acceleration:   0.0,
		Deceleration:   0.0,
		Recover:        false,
	})
	return result
}

type CarApplyInfo struct {
	Model    *game.Model
	Position dprec.Vec3
	Rotation dprec.Quat
	Inputs   ControlInput
}

type Car struct {
	chassis *Chassis
	axes    []*Axis
	entity  *ecs.Entity
}

func (c *Car) Chassis() *Chassis {
	return c.chassis
}

func (c *Car) Axes() []*Axis {
	return c.axes
}

func (c *Car) Entity() *ecs.Entity {
	return c.entity
}

func (c *Car) Velocity() float64 {
	return c.Chassis().Body().Velocity().Length()
}

type Chassis struct {
	node *game.Node
	body *physics.Body
}

func (c *Chassis) Node() *game.Node {
	return c.node
}

func (c *Chassis) Body() *physics.Body {
	return c.body
}

type Axis struct {
	maxSteeringAngle dprec.Angle
	maxAcceleration  float64
	maxBraking       float64
	reverseRatio     float64
	leftWheel        *Wheel
	rightWheel       *Wheel
	leftHub          *Hub
	rightHub         *Hub
}

func (a *Axis) MaxSteeringAngle() dprec.Angle {
	return a.maxSteeringAngle
}

func (a *Axis) MaxAcceleration() float64 {
	return a.maxAcceleration
}

func (a *Axis) MaxBraking() float64 {
	return a.maxBraking
}

func (a *Axis) ReverseRatio() float64 {
	return a.reverseRatio
}

func (a *Axis) LeftWheel() *Wheel {
	return a.leftWheel
}

func (a *Axis) RightWheel() *Wheel {
	return a.rightWheel
}

func (a *Axis) LeftHub() *Hub {
	return a.leftHub
}

func (a *Axis) RightHub() *Hub {
	return a.rightHub
}

type Wheel struct {
	node                 *game.Node
	body                 *physics.Body
	directionSolver      *constraint.MatchDirections
	attachmentConstraint *physics.DBConstraint
}

func (w *Wheel) Velocity() float64 {
	return dprec.Vec3Dot(w.body.AngularVelocity(), w.body.Orientation().OrientationX())
}

func (w *Wheel) Node() *game.Node {
	return w.node
}

func (w *Wheel) Body() *physics.Body {
	return w.body
}

func (w *Wheel) DirectionSolver() *constraint.MatchDirections {
	return w.directionSolver
}

func (w *Wheel) AttachmentConstraint() *physics.DBConstraint {
	return w.attachmentConstraint
}

type Hub struct {
	node *game.Node
	body *physics.Body
}

func (h *Hub) Node() *game.Node {
	return h.node
}

func (h *Hub) Body() *physics.Body {
	return h.body
}
