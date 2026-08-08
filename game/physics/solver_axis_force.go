package physics

import "github.com/mokiat/gomath/dprec"

// AxisForceSolverConfig holds the parameters with which an
// [AxisForceSolver] is configured, either through [NewAxisForceSolver] or
// [AxisForceSolver.Configure].
type AxisForceSolverConfig struct {

	// BodyAnchorOffset is the body-local-space offset, relative to the
	// target's center of mass, of the point at which the force is applied.
	//
	// An offset away from the center of mass causes the force to also
	// induce torque on the target, in addition to accelerating it linearly.
	BodyAnchorOffset dprec.Vec3

	// Axis is the body-local-space direction along which the force acts.
	// It need not be normalized; it is normalized internally, so that
	// Magnitude alone controls the force's magnitude.
	Axis dprec.Vec3

	// Magnitude is the magnitude, in Newtons, of the force applied along
	// Axis.
	Magnitude float64
}

// AxisForceSolver is an [AccelerationSolver] that applies a constant force
// along a body-fixed axis, at a fixed offset from the target's center of
// mass - modeling effects such as a thruster or engine mounted on the
// target.
//
// Since Axis and BodyAnchorOffset are both expressed in the target's
// local space, the force rotates together with the target, unlike
// [GravitySolver], which pulls along a fixed direction in world space.
//
// An AxisForceSolver must be configured, either through
// [NewAxisForceSolver] or [AxisForceSolver.Configure], before being
// registered with a [Scene] through [BodyAcceleratorView.Create] or
// [GlobalAcceleratorView.Create].
type AxisForceSolver struct {
	bodyAnchorOffset dprec.Vec3
	axis             dprec.Vec3
	magnitude        float64
}

var _ AccelerationSolver = (*AxisForceSolver)(nil)

// NewAxisForceSolver creates a new [AxisForceSolver] configured according
// to config.
func NewAxisForceSolver(config AxisForceSolverConfig) *AxisForceSolver {
	result := &AxisForceSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config.
//
// Configure must be called before this solver is registered with a
// [Scene] through [BodyAcceleratorView.Create] or
// [GlobalAcceleratorView.Create]. Unlike [NewAxisForceSolver], it can be
// called on an already-allocated solver, which allows solvers to be
// cached (e.g. in a slice) and configured on demand.
func (s *AxisForceSolver) Configure(config AxisForceSolverConfig) {
	s.bodyAnchorOffset = config.BodyAnchorOffset
	s.axis = dprec.UnitVec3(config.Axis)
	s.magnitude = config.Magnitude
}

// BodyAnchorOffset returns the body-local-space offset, relative to the
// target's center of mass, of the point at which the force is applied.
func (s *AxisForceSolver) BodyAnchorOffset() dprec.Vec3 {
	return s.bodyAnchorOffset
}

// SetBodyAnchorOffset changes the body-local-space offset, relative to
// the target's center of mass, of the point at which the force is
// applied.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisForceSolver) SetBodyAnchorOffset(offset dprec.Vec3) *AxisForceSolver {
	s.bodyAnchorOffset = offset
	return s
}

// Axis returns the body-local-space direction along which the force
// acts, as a unit vector.
func (s *AxisForceSolver) Axis() dprec.Vec3 {
	return s.axis
}

// SetAxis changes the body-local-space direction along which the force
// acts. The specified direction need not be normalized.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisForceSolver) SetAxis(axis dprec.Vec3) *AxisForceSolver {
	s.axis = dprec.UnitVec3(axis)
	return s
}

// Magnitude returns the magnitude, in Newtons, of the force applied along
// Axis.
func (s *AxisForceSolver) Magnitude() float64 {
	return s.magnitude
}

// SetMagnitude changes the magnitude, in Newtons, of the force applied
// along Axis.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisForceSolver) SetMagnitude(magnitude float64) *AxisForceSolver {
	s.magnitude = magnitude
	return s
}

// ApplyAcceleration accumulates on ctx.Target the linear and angular
// acceleration that result from a force of Magnitude, acting along Axis,
// applied at BodyAnchorOffset - all three rotated from the target's local
// space into world space using its current orientation.
func (s *AxisForceSolver) ApplyAcceleration(ctx AccelerationContext) {
	anchorOffsetWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), s.bodyAnchorOffset)

	force := dprec.Vec3Prod(s.axis, s.magnitude)
	forceWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), force)
	ctx.Target.ApplyOffsetForce(anchorOffsetWS, forceWS)
}
