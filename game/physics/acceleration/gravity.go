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
