package physics

import "github.com/mokiat/gomath/dprec"

// FixedDistanceSolverConfig holds the parameters with which a
// [FixedDistanceSolver] is configured, either through
// [NewFixedDistanceSolver] or [FixedDistanceSolver.Configure].
type FixedDistanceSolverConfig struct {

	// FixedPoint is the world-space position at which the target is
	// anchored.
	FixedPoint dprec.Vec3

	// BodyAnchorOffset is the body-local-space offset, relative to the
	// target's center of mass, of the point that is held at Distance
	// from FixedPoint.
	BodyAnchorOffset dprec.Vec3

	// Distance is the distance at which the target is held from
	// FixedPoint.
	Distance float64
}

// FixedDistanceSolver is a [SoloConstraintSolver] that holds a point on a
// target at a fixed distance from a fixed point in world space, acting
// like a rigid rod between the two - it resists the target moving both
// closer to and farther from FixedPoint.
//
// A FixedDistanceSolver must be configured, either through
// [NewFixedDistanceSolver] or [FixedDistanceSolver.Configure], before
// being registered with a [Scene] through [SoloConstraintView.Create].
type FixedDistanceSolver struct {
	fixedPoint       dprec.Vec3
	bodyAnchorOffset dprec.Vec3
	distance         float64

	jacobian Jacobian
	drift    float64
}

var _ SoloConstraintSolver = (*FixedDistanceSolver)(nil)

// NewFixedDistanceSolver creates a new [FixedDistanceSolver] configured
// according to config.
func NewFixedDistanceSolver(config FixedDistanceSolverConfig) *FixedDistanceSolver {
	result := &FixedDistanceSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config.
//
// Configure must be called before this solver is registered with a
// [Scene] through [SoloConstraintView.Create]. Unlike
// [NewFixedDistanceSolver], it can be called on an already-allocated
// solver, which allows solvers to be cached (e.g. in a slice) and
// configured on demand.
func (s *FixedDistanceSolver) Configure(config FixedDistanceSolverConfig) {
	s.fixedPoint = config.FixedPoint
	s.bodyAnchorOffset = config.BodyAnchorOffset
	s.distance = config.Distance
}

// FixedPoint returns the world-space position at which the target is
// anchored.
func (s *FixedDistanceSolver) FixedPoint() dprec.Vec3 {
	return s.fixedPoint
}

// SetFixedPoint changes the world-space position at which the target is
// anchored.
//
// It returns the solver itself, so that calls can be chained.
func (s *FixedDistanceSolver) SetFixedPoint(fixedPoint dprec.Vec3) *FixedDistanceSolver {
	s.fixedPoint = fixedPoint
	return s
}

// BodyAnchorOffset returns the body-local-space offset, relative to the
// target's center of mass, of the point that is held at Distance from
// FixedPoint.
func (s *FixedDistanceSolver) BodyAnchorOffset() dprec.Vec3 {
	return s.bodyAnchorOffset
}

// SetBodyAnchorOffset changes the body-local-space offset, relative to
// the target's center of mass, of the point that is held at Distance
// from FixedPoint.
//
// It returns the solver itself, so that calls can be chained.
func (s *FixedDistanceSolver) SetBodyAnchorOffset(offset dprec.Vec3) *FixedDistanceSolver {
	s.bodyAnchorOffset = offset
	return s
}

// Distance returns the distance at which the target is held from
// FixedPoint.
func (s *FixedDistanceSolver) Distance() float64 {
	return s.distance
}

// SetDistance changes the distance at which the target is held from
// FixedPoint.
//
// It returns the solver itself, so that calls can be chained.
func (s *FixedDistanceSolver) SetDistance(distance float64) *FixedDistanceSolver {
	s.distance = distance
	return s
}

// Reset implements [SoloConstraintSolver.Reset].
//
// It recomputes the constraint's [Jacobian], along with the world-space
// offset from the target's center of mass to its anchor point (derived
// from BodyAnchorOffset and the target's current rotation), and the
// current distance error (drift) between that anchor point and
// FixedPoint, based on the target's current position and rotation.
//
// If the anchor point currently coincides with FixedPoint, the
// constraint direction is undefined; an arbitrary axis is used as a
// fallback in that degenerate case.
func (s *FixedDistanceSolver) Reset(ctx SoloConstraintContext) {
	anchorOffsetWS := dprec.QuatVec3Rotation(ctx.Target.Rotation(), s.bodyAnchorOffset)
	anchorWS := dprec.Vec3Sum(ctx.Target.Position(), anchorOffsetWS)
	delta := dprec.Vec3Diff(anchorWS, s.fixedPoint)

	normal := dprec.BasisXVec3()
	actualDistance := delta.Length()
	if actualDistance > Epsilon {
		normal = dprec.UnitVec3(delta)
	}

	s.jacobian = Jacobian{
		LinearSlope:  normal,
		AngularSlope: dprec.Vec3Cross(anchorOffsetWS, normal),
	}
	s.drift = s.distance - actualDistance
}

// ApplyImpulses implements [SoloConstraintSolver.ApplyImpulses].
//
// It resolves an impulse, without restitution, that drives the target's
// velocity at the anchor point toward closing the distance error
// (drift) computed by [FixedDistanceSolver.Reset], pushing the target
// away from FixedPoint when it is too close and pulling it back when it
// is too far.
func (s *FixedDistanceSolver) ApplyImpulses(ctx SoloConstraintContext) {
	impulse := ctx.ImpulseSolution(s.jacobian, s.drift, 0.0)
	ctx.Target.ApplyImpulse(impulse)
}

// ApplyNudges implements [SoloConstraintSolver.ApplyNudges].
//
// It nudges the target's position and rotation to reduce any remaining
// distance error (drift) between its anchor point and FixedPoint.
func (s *FixedDistanceSolver) ApplyNudges(ctx SoloConstraintContext) {
	nudge := ctx.NudgeSolution(s.jacobian, s.drift)
	ctx.Target.ApplyNudge(nudge)
}
