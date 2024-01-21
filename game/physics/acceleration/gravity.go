package acceleration

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/physics/solver"
)

// NewGravityDirection creates a new GravityDirection acceleration
// that applies a constant force in a specific direction.
func NewGravityDirection() *GravityDirection {
	return &GravityDirection{
		direction:    dprec.NewVec3(0.0, -1.0, 0.0),
		acceleration: 9.8,
	}
}

var _ solver.Acceleration = (*GravityDirection)(nil)

// GravityDirection represents the solution for an acceleration that
// applies a constant force in a specific direction.
type GravityDirection struct {
	direction    dprec.Vec3
	acceleration float64
}

// Direction returns the direction of the gravity force.
func (d *GravityDirection) Direction() dprec.Vec3 {
	return d.direction
}

// SetDirection changes the direction of the gravity force.
func (d *GravityDirection) SetDirection(direction dprec.Vec3) *GravityDirection {
	d.direction = direction
	return d
}

// Acceleration returns the acceleration of the gravity force.
func (d *GravityDirection) Acceleration() float64 {
	return d.acceleration
}

// SetAcceleration changes the acceleration of the gravity force.
func (d *GravityDirection) SetAcceleration(acceleration float64) *GravityDirection {
	d.acceleration = acceleration
	return d
}

func (d *GravityDirection) ApplyAcceleration(ctx solver.AccelerationContext) {
	ctx.Target.AddLinearAcceleration(
		dprec.Vec3Prod(d.direction, d.acceleration),
	)
}

// NewGravityPosition creates a new GravityPosition acceleration
// that applies a constant force towards a specific position.
func NewGravityPosition() *GravityPosition {
	return &GravityPosition{
		position: dprec.ZeroVec3(),
	}
}

var _ solver.Acceleration = (*GravityPosition)(nil)

// GravityPosition represents the solution for an acceleration that
// applies a constant force towards a specific position.
type GravityPosition struct {
	position     dprec.Vec3
	acceleration float64

	// TODO: Add ability to control distance falloff.
}

// Position returns the position of the gravity force.
func (d *GravityPosition) Position() dprec.Vec3 {
	return d.position
}

// SetPosition changes the position of the gravity force.
func (d *GravityPosition) SetPosition(position dprec.Vec3) *GravityPosition {
	d.position = position
	return d
}

// Acceleration returns the acceleration of the gravity force.
func (d *GravityPosition) Acceleration() float64 {
	return d.acceleration
}

// SetAcceleration changes the acceleration of the gravity force.
func (d *GravityPosition) SetAcceleration(acceleration float64) *GravityPosition {
	d.acceleration = acceleration
	return d
}

func (d *GravityPosition) ApplyAcceleration(ctx solver.AccelerationContext) {
	delta := dprec.Vec3Diff(d.position, ctx.Target.Position())
	if distance := delta.Length(); distance > solver.Epsilon {
		ctx.Target.AddLinearAcceleration(
			dprec.ResizedVec3(delta, d.acceleration),
		)
	}
}
