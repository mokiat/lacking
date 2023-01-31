package preset

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/constraint"
)

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

func (d *CarDefinition) ApplyToModel(scene *game.Scene, model *game.Model, position dprec.Vec3, rotation dprec.Quat) *Car {
	chassisNode := model.FindNode(d.chassisDef.nodeName)

	chassisPosition := position
	chassisRotation := rotation

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

		leftWheelNode := model.FindNode(axisDef.leftWheelDef.nodeName)
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
			hubNode := model.FindNode(hubDef.nodeName)
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

		rightWheelNode := model.FindNode(axisDef.rightWheelDef.nodeName)
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
			hubNode := model.FindNode(hubDef.nodeName)
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

		axes = append(axes, &Axis{
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

	return &Car{
		chassis: &Chassis{
			node: chassisNode,
			body: chassisBody,
		},
		axes: axes,
	}
}

type Car struct {
	chassis *Chassis
	axes    []*Axis
}

func (c *Car) Chassis() *Chassis {
	return c.chassis
}

func (c *Car) Axes() []*Axis {
	return c.axes
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
	leftWheel  *Wheel
	rightWheel *Wheel
	leftHub    *Hub
	rightHub   *Hub
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
