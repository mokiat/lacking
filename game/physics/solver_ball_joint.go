package physics

import "github.com/mokiat/gomath/dprec"

// BallJointSolverConfig holds the parameters with which a
// [BallJointSolver] is configured, either through [NewBallJointSolver]
// or [BallJointSolver.Configure].
type BallJointSolverConfig struct {

	// PrimaryBodyAnchorOffset is the body-local-space offset, relative to
	// the primary target's center of mass, of the point at which the
	// ball joint is anchored on the primary body.
	PrimaryBodyAnchorOffset dprec.Vec3

	// SecondaryBodyAnchorOffset is the body-local-space offset, relative
	// to the secondary target's center of mass, of the point at which
	// the ball joint is anchored on the secondary body.
	SecondaryBodyAnchorOffset dprec.Vec3
}

// BallJointSolver is a [PairConstraintSolver] that holds an anchor point
// on each of its two target bodies coincident in world space, acting
// like a ball-and-socket joint between the two - it pins the anchor
// points together while leaving all relative rotation between the
// targets unconstrained.
//
// Unlike [DistanceSolver], which only resists the anchor points moving
// closer together or farther apart, BallJointSolver fully constrains
// their relative position, leaving only rotation free.
//
// Internally, it is implemented as three [AxisDisplacementSolver]s, one
// for each of the primary body's local X, Y and Z axes, each configured
// with a zero [AxisDisplacementSolver.Displacement]. Together, since
// those axes span the primary body's entire local frame, they pin the
// anchor points coincident.
//
// A BallJointSolver must be configured, either through
// [NewBallJointSolver] or [BallJointSolver.Configure], before being
// registered with a [Scene] through [PairConstraintView.Create].
type BallJointSolver struct {
	solverX AxisDisplacementSolver
	solverY AxisDisplacementSolver
	solverZ AxisDisplacementSolver
}

var _ PairConstraintSolver = (*BallJointSolver)(nil)

// NewBallJointSolver creates a new [BallJointSolver] configured
// according to config.
func NewBallJointSolver(config BallJointSolverConfig) *BallJointSolver {
	result := &BallJointSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config.
//
// Configure must be called before this solver is registered with a
// [Scene] through [PairConstraintView.Create]. Unlike
// [NewBallJointSolver], it can be called on an already-allocated solver,
// which allows solvers to be cached (e.g. in a slice) and configured on
// demand.
func (s *BallJointSolver) Configure(config BallJointSolverConfig) {
	s.solverX.Configure(AxisDisplacementSolverConfig{
		PrimaryBodyAnchorOffset:   config.PrimaryBodyAnchorOffset,
		PrimaryBodyAxis:           dprec.BasisXVec3(),
		SecondaryBodyAnchorOffset: config.SecondaryBodyAnchorOffset,
		Displacement:              0.0,
	})
	s.solverY.Configure(AxisDisplacementSolverConfig{
		PrimaryBodyAnchorOffset:   config.PrimaryBodyAnchorOffset,
		PrimaryBodyAxis:           dprec.BasisYVec3(),
		SecondaryBodyAnchorOffset: config.SecondaryBodyAnchorOffset,
		Displacement:              0.0,
	})
	s.solverZ.Configure(AxisDisplacementSolverConfig{
		PrimaryBodyAnchorOffset:   config.PrimaryBodyAnchorOffset,
		PrimaryBodyAxis:           dprec.BasisZVec3(),
		SecondaryBodyAnchorOffset: config.SecondaryBodyAnchorOffset,
		Displacement:              0.0,
	})
}

// PrimaryBodyAnchorOffset returns the body-local-space offset, relative
// to the primary target's center of mass, of the point at which the
// ball joint is anchored on the primary body.
func (s *BallJointSolver) PrimaryBodyAnchorOffset() dprec.Vec3 {
	return s.solverX.PrimaryBodyAnchorOffset()
}

// SetPrimaryBodyAnchorOffset changes the body-local-space offset,
// relative to the primary target's center of mass, of the point at
// which the ball joint is anchored on the primary body.
//
// It returns the solver itself, so that calls can be chained.
func (s *BallJointSolver) SetPrimaryBodyAnchorOffset(offset dprec.Vec3) *BallJointSolver {
	s.solverX.SetPrimaryBodyAnchorOffset(offset)
	s.solverY.SetPrimaryBodyAnchorOffset(offset)
	s.solverZ.SetPrimaryBodyAnchorOffset(offset)
	return s
}

// SecondaryBodyAnchorOffset returns the body-local-space offset,
// relative to the secondary target's center of mass, of the point at
// which the ball joint is anchored on the secondary body.
func (s *BallJointSolver) SecondaryBodyAnchorOffset() dprec.Vec3 {
	return s.solverX.SecondaryBodyAnchorOffset()
}

// SetSecondaryBodyAnchorOffset changes the body-local-space offset,
// relative to the secondary target's center of mass, of the point at
// which the ball joint is anchored on the secondary body.
//
// It returns the solver itself, so that calls can be chained.
func (s *BallJointSolver) SetSecondaryBodyAnchorOffset(offset dprec.Vec3) *BallJointSolver {
	s.solverX.SetSecondaryBodyAnchorOffset(offset)
	s.solverY.SetSecondaryBodyAnchorOffset(offset)
	s.solverZ.SetSecondaryBodyAnchorOffset(offset)
	return s
}

// Reset implements [PairConstraintSolver.Reset].
//
// It calls [AxisDisplacementSolver.Reset] on each of the three per-axis
// sub-solvers, in X, Y, Z order, so that each recomputes its own
// Jacobians and drift from the targets' current positions and
// rotations.
func (s *BallJointSolver) Reset(ctx PairConstraintContext) {
	s.solverX.Reset(ctx)
	s.solverY.Reset(ctx)
	s.solverZ.Reset(ctx)
}

// ApplyImpulses implements [PairConstraintSolver.ApplyImpulses].
//
// It calls [AxisDisplacementSolver.ApplyImpulses] on each of the three
// per-axis sub-solvers, in X, Y, Z order, so that their combined
// impulses drive the anchor points' relative velocity toward zero along
// all three of the primary body's local axes.
func (s *BallJointSolver) ApplyImpulses(ctx PairConstraintContext) {
	s.solverX.ApplyImpulses(ctx)
	s.solverY.ApplyImpulses(ctx)
	s.solverZ.ApplyImpulses(ctx)
}

// ApplyNudges implements [PairConstraintSolver.ApplyNudges].
//
// It calls [AxisDisplacementSolver.ApplyNudges] on each of the three
// per-axis sub-solvers, in X, Y, Z order. Each sub-solver recomputes its
// own Jacobians and drift before nudging, as required by
// [AxisDisplacementSolver.ApplyNudges], so a nudge applied by an earlier
// sub-solver in this call is correctly observed by the ones that follow.
func (s *BallJointSolver) ApplyNudges(ctx PairConstraintContext) {
	s.solverX.ApplyNudges(ctx)
	s.solverY.ApplyNudges(ctx)
	s.solverZ.ApplyNudges(ctx)
}
