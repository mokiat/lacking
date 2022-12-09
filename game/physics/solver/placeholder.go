package solver

import "github.com/mokiat/gomath/dprec"

type PlaceholderState struct {
	Mass            float64
	MomentOfInertia dprec.Mat3

	LinearVelocity  dprec.Vec3
	AngularVelocity dprec.Vec3

	Position dprec.Vec3
	Rotation dprec.Quat
}

type Placeholder struct {
	inverseMass            float64
	inverseMomentOfInertia dprec.Mat3

	linearVelocity  dprec.Vec3
	angularVelocity dprec.Vec3

	position dprec.Vec3
	rotation dprec.Quat
}

func (p *Placeholder) Init(state PlaceholderState) {
	p.inverseMass = 1.0 / state.Mass
	// TODO: First rotate the moment of inertia according to the object's rotation
	p.inverseMomentOfInertia = dprec.InverseMat3(state.MomentOfInertia)

	p.linearVelocity = state.LinearVelocity
	p.angularVelocity = state.AngularVelocity

	p.position = state.Position
	p.rotation = state.Rotation
}

func (p *Placeholder) LinearVelocity() dprec.Vec3 {
	return p.linearVelocity
}

func (p *Placeholder) SetLinearVelocity(velocity dprec.Vec3) {
	p.linearVelocity = velocity
}

func (p *Placeholder) AngularVelocity() dprec.Vec3 {
	return p.angularVelocity
}

func (p *Placeholder) SetAngularVelocity(velocity dprec.Vec3) {
	p.angularVelocity = velocity
}

func (p *Placeholder) ApplyImpulse(impulse Impulse) {
	linearChange := dprec.Vec3Prod(impulse.Linear, p.inverseMass)
	p.linearVelocity = dprec.Vec3Sum(p.linearVelocity, linearChange)

	angularChange := dprec.Mat3Vec3Prod(p.inverseMomentOfInertia, impulse.Angular)
	p.angularVelocity = dprec.Vec3Sum(p.angularVelocity, angularChange)
}

func (p *Placeholder) Position() dprec.Vec3 {
	return p.position
}

func (p *Placeholder) SetPosition(position dprec.Vec3) {
	p.position = position
}

func (p *Placeholder) Rotation() dprec.Quat {
	return p.rotation
}

func (p *Placeholder) SetRotation(rotation dprec.Quat) {
	p.rotation = rotation
}

func (p *Placeholder) ApplyNudge(nudge Nudge) {
	linearChange := dprec.Vec3Prod(nudge.Linear, p.inverseMass)
	p.position = dprec.Vec3Sum(p.position, linearChange)

	angularChange := dprec.Mat3Vec3Prod(p.inverseMomentOfInertia, nudge.Angular)
	if radians := angularChange.Length(); radians > Epsilon {
		rotationChange := dprec.RotationQuat(dprec.Radians(radians), angularChange)
		p.rotation = dprec.UnitQuat(dprec.QuatProd(rotationChange, p.rotation))
	}
}
