package physics

import "github.com/mokiat/gomath/dprec"

// DirectionForceSolverConfig holds the parameters with which a
// [DirectionForceSolver] is configured, either through
// [NewDirectionForceSolver] or [DirectionForceSolver.Configure].
type DirectionForceSolverConfig struct {

	// BodyAnchorOffset is the body-local-space offset, relative to the
	// target's center of mass, of the point at which the force is applied.
	//
	// An offset away from the center of mass causes the force to also
	// induce torque on the target, in addition to accelerating it linearly.
	BodyAnchorOffset dprec.Vec3

	// Direction is the world-space direction along which the force acts.
	// It need not be normalized; it is normalized internally, so that
	// Magnitude alone controls the force's magnitude.
	Direction dprec.Vec3

	// Magnitude is the magnitude, in Newtons, of the force applied along
	// Direction.
	Magnitude float64
}

// DirectionForceSolver is an [AccelerationSolver] that applies a constant
// force along a fixed direction in world space, at a fixed offset from the
// target's center of mass - modeling effects such as wind or an external
// jet blowing on the target from one side.
//
// Unlike [AxisForceSolver], whose axis is expressed in the target's local
// space and so rotates together with it, DirectionForceSolver's Direction
// is expressed in world space and stays constant regardless of the
// target's orientation. BodyAnchorOffset, however, is still expressed in
// the target's local space, same as with AxisForceSolver, so the point at
// which the force is applied still moves and rotates together with the
// target, meaning the torque induced by the force still varies as the
// target rotates even though the force's direction does not.
//
// A DirectionForceSolver must be configured, either through
// [NewDirectionForceSolver] or [DirectionForceSolver.Configure], before
// being registered with a [Scene] through [BodyAcceleratorView.Create] or
// [GlobalAcceleratorView.Create].
type DirectionForceSolver struct {
	bodyAnchorOffset dprec.Vec3
	direction        dprec.Vec3
	magnitude        float64
}

var _ AccelerationSolver = (*DirectionForceSolver)(nil)

// NewDirectionForceSolver creates a new [DirectionForceSolver] configured
// according to config.
func NewDirectionForceSolver(config DirectionForceSolverConfig) *DirectionForceSolver {
	result := &DirectionForceSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config.
//
// Configure must be called before this solver is registered with a
// [Scene] through [BodyAcceleratorView.Create] or
// [GlobalAcceleratorView.Create]. Unlike [NewDirectionForceSolver], it can
// be called on an already-allocated solver, which allows solvers to be
// cached (e.g. in a slice) and configured on demand.
func (s *DirectionForceSolver) Configure(config DirectionForceSolverConfig) {
	s.bodyAnchorOffset = config.BodyAnchorOffset
	s.direction = dprec.UnitVec3(config.Direction)
	s.magnitude = config.Magnitude
}

// BodyAnchorOffset returns the body-local-space offset, relative to the
// target's center of mass, of the point at which the force is applied.
func (s *DirectionForceSolver) BodyAnchorOffset() dprec.Vec3 {
	return s.bodyAnchorOffset
}

// SetBodyAnchorOffset changes the body-local-space offset, relative to
// the target's center of mass, of the point at which the force is
// applied.
//
// It returns the solver itself, so that calls can be chained.
func (s *DirectionForceSolver) SetBodyAnchorOffset(offset dprec.Vec3) *DirectionForceSolver {
	s.bodyAnchorOffset = offset
	return s
}

// Direction returns the world-space direction along which the force acts,
// as a unit vector.
func (s *DirectionForceSolver) Direction() dprec.Vec3 {
	return s.direction
}

// SetDirection changes the world-space direction along which the force
// acts. The specified direction need not be normalized.
//
// It returns the solver itself, so that calls can be chained.
func (s *DirectionForceSolver) SetDirection(direction dprec.Vec3) *DirectionForceSolver {
	s.direction = dprec.UnitVec3(direction)
	return s
}

// Magnitude returns the magnitude, in Newtons, of the force applied along
// Direction.
func (s *DirectionForceSolver) Magnitude() float64 {
	return s.magnitude
}

// SetMagnitude changes the magnitude, in Newtons, of the force applied
// along Direction.
//
// It returns the solver itself, so that calls can be chained.
func (s *DirectionForceSolver) SetMagnitude(magnitude float64) *DirectionForceSolver {
	s.magnitude = magnitude
	return s
}

// ApplyAcceleration accumulates on ctx.Target the linear and angular
// acceleration that result from a force of Magnitude, acting along
// Direction, applied at BodyAnchorOffset - with BodyAnchorOffset rotated
// from the target's local space into world space using its current
// orientation, and Direction used as-is, since it is already expressed in
// world space.
func (s *DirectionForceSolver) ApplyAcceleration(ctx AccelerationContext) {
	anchorOffsetWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), s.bodyAnchorOffset)

	forceWS := dprec.Vec3Prod(s.direction, s.magnitude)
	ctx.Target.ApplyOffsetForce(anchorOffsetWS, forceWS)
}
