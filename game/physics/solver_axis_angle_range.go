package physics

import "github.com/mokiat/gomath/dprec"

// AxisAngleRangeSolverConfig holds the parameters with which an
// [AxisAngleRangeSolver] is configured, either through
// [NewAxisAngleRangeSolver] or [AxisAngleRangeSolver.Configure].
type AxisAngleRangeSolverConfig struct {

	// PrimaryBodyAxis is the body-local-space direction, relative to the
	// primary target's rotation, from which the angle is measured. It need
	// not be unit-length; it is normalized when the solver is configured,
	// so it must not be the zero vector, nor may it be collinear with
	// RotationAxis.
	PrimaryBodyAxis dprec.Vec3

	// SecondaryBodyAxis is the body-local-space direction, relative to the
	// secondary target's rotation, to which the angle is measured. It need
	// not be unit-length; it is normalized when the solver is configured,
	// so it must not be the zero vector.
	SecondaryBodyAxis dprec.Vec3

	// RotationAxis is the body-local-space direction, relative to the
	// primary target's rotation, about which the angle between
	// PrimaryBodyAxis and SecondaryBodyAxis is measured. It need not be
	// unit-length; it is normalized when the solver is configured, so it
	// must not be the zero vector, nor may it be collinear with
	// PrimaryBodyAxis.
	RotationAxis dprec.Vec3

	// MinAngle is the lowest permitted signed angle, measured about
	// RotationAxis, from PrimaryBodyAxis to SecondaryBodyAxis. It must lie
	// within the (-180, 180] degrees range.
	MinAngle dprec.Angle

	// MaxAngle is the highest permitted signed angle, measured about
	// RotationAxis, from PrimaryBodyAxis to SecondaryBodyAxis. It must lie
	// within the (-180, 180] degrees range.
	MaxAngle dprec.Angle

	// RestitutionCoefficient is the bounciness applied when the angle
	// reaches MinAngle or MaxAngle. Negative values are clamped to zero.
	RestitutionCoefficient float64
}

// AxisAngleRangeSolver is a [PairConstraintSolver] that keeps the signed
// angle from an axis fixed to the primary body to an axis fixed to the
// secondary body, measured about a rotation axis fixed to the primary
// body, within a [AxisAngleRangeSolver.MinAngle] and
// [AxisAngleRangeSolver.MaxAngle] range - acting like a rigid stop only
// once one of the range's limits is reached, and applying no torque while
// the angle is within range.
//
// It is the angular counterpart to [AxisRangeSolver], and is what turns a
// hinge - a [BallJointSolver] combined with a [MatchAxesSolver], or
// similar - into a hinge with end stops. Only rotation about the rotation
// axis is restricted; the targets' positions and linear velocities are
// left untouched, as are the two rotational degrees of freedom that tilt
// the rotation axis itself.
//
// The angle is measured after both axes are projected onto the plane
// perpendicular to the rotation axis, and is therefore only defined
// within the (-180, 180] degrees range. MinAngle and MaxAngle must lie
// within that range to be reachable, and a target that swings past the
// 180 degrees discontinuity will be driven toward the opposite limit,
// since the measured angle wraps around to the other end of the range.
//
// Whenever the secondary axis becomes collinear with the rotation axis,
// its projection vanishes and the angle is undefined; the solver then
// applies nothing at all until the two separate again.
//
// An AxisAngleRangeSolver must be configured, either through
// [NewAxisAngleRangeSolver] or [AxisAngleRangeSolver.Configure], before
// being registered with a [Scene] through [PairConstraintView.Create].
type AxisAngleRangeSolver struct {
	primaryBodyAxis        dprec.Vec3
	secondaryBodyAxis      dprec.Vec3
	rotationAxis           dprec.Vec3
	minAngle               dprec.Angle
	maxAngle               dprec.Angle
	restitutionCoefficient float64

	primaryJacobian   Jacobian
	secondaryJacobian Jacobian
	drift             float64
}

var _ PairConstraintSolver = (*AxisAngleRangeSolver)(nil)

// NewAxisAngleRangeSolver creates a new [AxisAngleRangeSolver] configured
// according to config.
func NewAxisAngleRangeSolver(config AxisAngleRangeSolverConfig) *AxisAngleRangeSolver {
	result := &AxisAngleRangeSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config. PrimaryBodyAxis,
// SecondaryBodyAxis and RotationAxis are each normalized to unit length,
// and negative RestitutionCoefficient values are clamped to zero, as with
// [AxisAngleRangeSolver.SetRestitutionCoefficient].
//
// Configure must be called before this solver is registered with a
// [Scene] through [PairConstraintView.Create]. Unlike
// [NewAxisAngleRangeSolver], it can be called on an already-allocated
// solver, which allows solvers to be cached (e.g. in a slice) and
// configured on demand.
func (s *AxisAngleRangeSolver) Configure(config AxisAngleRangeSolverConfig) {
	s.primaryBodyAxis = dprec.UnitVec3(config.PrimaryBodyAxis)
	s.secondaryBodyAxis = dprec.UnitVec3(config.SecondaryBodyAxis)
	s.rotationAxis = dprec.UnitVec3(config.RotationAxis)
	s.minAngle = config.MinAngle
	s.maxAngle = config.MaxAngle
	s.restitutionCoefficient = max(0.0, config.RestitutionCoefficient)
}

// PrimaryBodyAxis returns the body-local-space, unit-length direction,
// relative to the primary target's rotation, from which the angle is
// measured.
func (s *AxisAngleRangeSolver) PrimaryBodyAxis() dprec.Vec3 {
	return s.primaryBodyAxis
}

// SetPrimaryBodyAxis changes the body-local-space direction, relative to
// the primary target's rotation, from which the angle is measured. The
// provided axis need not be unit-length; it is normalized before being
// stored. It must not be collinear with
// [AxisAngleRangeSolver.RotationAxis].
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisAngleRangeSolver) SetPrimaryBodyAxis(axis dprec.Vec3) *AxisAngleRangeSolver {
	s.primaryBodyAxis = dprec.UnitVec3(axis)
	return s
}

// SecondaryBodyAxis returns the body-local-space, unit-length direction,
// relative to the secondary target's rotation, to which the angle is
// measured.
func (s *AxisAngleRangeSolver) SecondaryBodyAxis() dprec.Vec3 {
	return s.secondaryBodyAxis
}

// SetSecondaryBodyAxis changes the body-local-space direction, relative
// to the secondary target's rotation, to which the angle is measured. The
// provided axis need not be unit-length; it is normalized before being
// stored.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisAngleRangeSolver) SetSecondaryBodyAxis(axis dprec.Vec3) *AxisAngleRangeSolver {
	s.secondaryBodyAxis = dprec.UnitVec3(axis)
	return s
}

// RotationAxis returns the body-local-space, unit-length direction,
// relative to the primary target's rotation, about which the angle
// between [AxisAngleRangeSolver.PrimaryBodyAxis] and
// [AxisAngleRangeSolver.SecondaryBodyAxis] is measured.
func (s *AxisAngleRangeSolver) RotationAxis() dprec.Vec3 {
	return s.rotationAxis
}

// SetRotationAxis changes the body-local-space direction, relative to the
// primary target's rotation, about which the angle between
// [AxisAngleRangeSolver.PrimaryBodyAxis] and
// [AxisAngleRangeSolver.SecondaryBodyAxis] is measured. The provided axis
// need not be unit-length; it is normalized before being stored. It must
// not be collinear with [AxisAngleRangeSolver.PrimaryBodyAxis].
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisAngleRangeSolver) SetRotationAxis(axis dprec.Vec3) *AxisAngleRangeSolver {
	s.rotationAxis = dprec.UnitVec3(axis)
	return s
}

// MinAngle returns the lowest permitted signed angle, measured about
// [AxisAngleRangeSolver.RotationAxis], from
// [AxisAngleRangeSolver.PrimaryBodyAxis] to
// [AxisAngleRangeSolver.SecondaryBodyAxis].
func (s *AxisAngleRangeSolver) MinAngle() dprec.Angle {
	return s.minAngle
}

// SetMinAngle changes the lowest permitted signed angle, measured about
// [AxisAngleRangeSolver.RotationAxis], from
// [AxisAngleRangeSolver.PrimaryBodyAxis] to
// [AxisAngleRangeSolver.SecondaryBodyAxis]. It must lie within the
// (-180, 180] degrees range to be reachable.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisAngleRangeSolver) SetMinAngle(angle dprec.Angle) *AxisAngleRangeSolver {
	s.minAngle = angle
	return s
}

// MaxAngle returns the highest permitted signed angle, measured about
// [AxisAngleRangeSolver.RotationAxis], from
// [AxisAngleRangeSolver.PrimaryBodyAxis] to
// [AxisAngleRangeSolver.SecondaryBodyAxis].
func (s *AxisAngleRangeSolver) MaxAngle() dprec.Angle {
	return s.maxAngle
}

// SetMaxAngle changes the highest permitted signed angle, measured about
// [AxisAngleRangeSolver.RotationAxis], from
// [AxisAngleRangeSolver.PrimaryBodyAxis] to
// [AxisAngleRangeSolver.SecondaryBodyAxis]. It must lie within the
// (-180, 180] degrees range to be reachable.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisAngleRangeSolver) SetMaxAngle(angle dprec.Angle) *AxisAngleRangeSolver {
	s.maxAngle = angle
	return s
}

// RestitutionCoefficient returns the bounciness applied when the angle
// reaches [AxisAngleRangeSolver.MinAngle] or
// [AxisAngleRangeSolver.MaxAngle].
func (s *AxisAngleRangeSolver) RestitutionCoefficient() float64 {
	return s.restitutionCoefficient
}

// SetRestitutionCoefficient changes the bounciness applied when the angle
// reaches [AxisAngleRangeSolver.MinAngle] or
// [AxisAngleRangeSolver.MaxAngle]. Negative values are clamped to zero.
//
// It returns the solver itself, so that calls can be chained.
func (s *AxisAngleRangeSolver) SetRestitutionCoefficient(coefficient float64) *AxisAngleRangeSolver {
	s.restitutionCoefficient = max(0.0, coefficient)
	return s
}

// Reset implements [PairConstraintSolver.Reset].
//
// It recomputes the constraint's primary and secondary [Jacobian]s and
// current range violation (drift), the same way
// [AxisAngleRangeSolver.recompute] does.
func (s *AxisAngleRangeSolver) Reset(ctx PairConstraintContext) {
	s.recompute(ctx)
}

// ApplyImpulses implements [PairConstraintSolver.ApplyImpulses].
//
// If the angle between the two body axes, as of the last call to
// [AxisAngleRangeSolver.Reset] or [AxisAngleRangeSolver.ApplyNudges], is
// within the
// [AxisAngleRangeSolver.MinAngle]/[AxisAngleRangeSolver.MaxAngle] range,
// it does nothing. Otherwise, it resolves a pair of impulses, combining
// restitution with Baumgarte positional-drift stabilization, that drive
// the targets' relative angular velocity toward bringing the angle back
// within range. If the two targets are already rotating apart (back
// towards the permitted range), it returns without applying anything,
// leaving any remaining violation to [AxisAngleRangeSolver.ApplyNudges].
func (s *AxisAngleRangeSolver) ApplyImpulses(ctx PairConstraintContext) {
	if s.drift == 0.0 {
		return // no constraint violation
	}

	bounceLambda, baumgarteLambda := ctx.ImpulseLambdaComponents(s.primaryJacobian, s.secondaryJacobian, s.drift, s.restitutionCoefficient)
	if bounceLambda < 0.0 {
		return // moving away
	}

	lambda := bounceLambda + baumgarteLambda
	primaryImpulse := s.primaryJacobian.Impulse(lambda)
	secondaryImpulse := s.secondaryJacobian.Impulse(lambda)

	ctx.PrimaryTarget.ApplyImpulse(primaryImpulse)
	ctx.SecondaryTarget.ApplyImpulse(secondaryImpulse)
}

// ApplyNudges implements [PairConstraintSolver.ApplyNudges].
//
// It first recomputes the constraint's primary and secondary [Jacobian]s
// and current range violation (drift), the same way
// [AxisAngleRangeSolver.recompute] does, since a preceding nudge - by
// this solver's own previous iteration, or by another constraint acting
// on either target - may have rotated either target since
// [AxisAngleRangeSolver.Reset] or the last call to this method. If the
// angle is within range, it does nothing; otherwise, it nudges both
// targets' rotations to bring the angle back within the
// [AxisAngleRangeSolver.MinAngle]/[AxisAngleRangeSolver.MaxAngle] range.
func (s *AxisAngleRangeSolver) ApplyNudges(ctx PairConstraintContext) {
	s.recompute(ctx)

	if s.drift > 0.0 {
		primaryNudge, secondaryNudge := ctx.NudgeSolution(
			s.primaryJacobian, s.secondaryJacobian, s.drift,
		)
		ctx.PrimaryTarget.ApplyNudge(primaryNudge)
		ctx.SecondaryTarget.ApplyNudge(secondaryNudge)
	}
}

// recompute recalculates the constraint's primary and secondary
// [Jacobian]s, along with the current range violation (drift), based on
// the targets' current rotations.
//
// It measures the signed angle from PrimaryBodyAxis to
// SecondaryBodyAxis, each transformed into world space through its own
// target's current rotation, about RotationAxis, transformed into world
// space through the primary target's current rotation. If that angle is
// below MinAngle, the Jacobians and drift are set up to rotate the two
// axes apart about the rotation axis; if it is above MaxAngle, they are
// set up to rotate them together; otherwise both Jacobians and the drift
// are reset to zero, so that [AxisAngleRangeSolver.ApplyImpulses] and
// [AxisAngleRangeSolver.ApplyNudges] apply no correction.
//
// The Jacobians and drift are similarly reset to zero whenever the
// secondary axis is collinear with the rotation axis, since the secondary
// axis then has no projection onto the plane in which the angle is
// measured, leaving the angle undefined.
func (s *AxisAngleRangeSolver) recompute(ctx PairConstraintContext) {
	primaryAxisWS := dprec.QuatVec3Rotation(ctx.PrimaryTarget.Rotation(), s.primaryBodyAxis)
	secondaryAxisWS := dprec.QuatVec3Rotation(ctx.SecondaryTarget.Rotation(), s.secondaryBodyAxis)
	axisWS := dprec.QuatVec3Rotation(ctx.PrimaryTarget.Rotation(), s.rotationAxis)

	if dprec.Abs(dprec.Vec3Dot(axisWS, secondaryAxisWS)) > 1.0-Epsilon {
		s.primaryJacobian = Jacobian{}
		s.secondaryJacobian = Jacobian{}
		s.drift = 0.0
		return // secondary direction is parallel to axis
	}
	angle := dprec.Vec3ProjectionAngle(primaryAxisWS, secondaryAxisWS, axisWS)

	switch {
	case angle < s.minAngle:
		s.primaryJacobian = Jacobian{
			AngularSlope: dprec.InverseVec3(axisWS),
		}
		s.secondaryJacobian = Jacobian{
			AngularSlope: axisWS,
		}
		s.drift = (s.minAngle - angle).Radians()

	case angle > s.maxAngle:
		s.primaryJacobian = Jacobian{
			AngularSlope: axisWS,
		}
		s.secondaryJacobian = Jacobian{
			AngularSlope: dprec.InverseVec3(axisWS),
		}
		s.drift = (angle - s.maxAngle).Radians()

	default:
		s.primaryJacobian = Jacobian{}
		s.secondaryJacobian = Jacobian{}
		s.drift = 0.0
	}
}
