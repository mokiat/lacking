package physics

import "github.com/mokiat/gomath/dprec"

// FixedPositionSolverConfig holds the parameters with which a
// [FixedPositionSolver] is configured, either through
// [NewFixedPositionSolver] or [FixedPositionSolver.Configure].
type FixedPositionSolverConfig struct {

	// Position is the world-space location at which the target is to be
	// held fixed.
	Position dprec.Vec3
}

// FixedPositionSolver is a [SoloConstraintSolver] that pins a target to a
// fixed position in world space, regardless of any forces or impulses
// acting on it.
//
// Unlike [SoloCollisionSolver], which only nudges its target enough to
// resolve penetration, FixedPositionSolver is a kinematic constraint - it
// unconditionally zeroes the target's linear velocity through
// [FixedPositionSolver.ApplyImpulses] and snaps it to Position through
// [FixedPositionSolver.ApplyNudges] on every step. It does not affect
// angular velocity or orientation.
//
// A FixedPositionSolver must be configured, either through
// [NewFixedPositionSolver] or [FixedPositionSolver.Configure], before
// being registered with a [Scene] through [SoloConstraintView.Create].
type FixedPositionSolver struct {
	position dprec.Vec3
}

var _ SoloConstraintSolver = (*FixedPositionSolver)(nil)

// NewFixedPositionSolver creates a new [FixedPositionSolver] configured
// according to config.
func NewFixedPositionSolver(config FixedPositionSolverConfig) *FixedPositionSolver {
	result := &FixedPositionSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config.
//
// Configure must be called before this solver is registered with a
// [Scene] through [SoloConstraintView.Create]. Unlike
// [NewFixedPositionSolver], it can be called on an already-allocated
// solver, which allows solvers to be cached (e.g. in a slice) and
// configured on demand.
func (s *FixedPositionSolver) Configure(config FixedPositionSolverConfig) {
	s.position = config.Position
}

// Position returns the world-space location at which the target is held
// fixed.
func (s *FixedPositionSolver) Position() dprec.Vec3 {
	return s.position
}

// SetPosition changes the world-space location at which the target is
// held fixed.
//
// It returns the solver itself, so that calls can be chained.
func (s *FixedPositionSolver) SetPosition(position dprec.Vec3) *FixedPositionSolver {
	s.position = position
	return s
}

// Reset implements [SoloConstraintSolver.Reset].
//
// It is a no-op, since this solver holds no per-step state that needs to
// be derived from the target's current position or velocity.
func (s *FixedPositionSolver) Reset(ctx SoloConstraintContext) {}

// ApplyImpulses implements [SoloConstraintSolver.ApplyImpulses].
//
// It unconditionally zeroes the target's linear velocity.
func (s *FixedPositionSolver) ApplyImpulses(ctx SoloConstraintContext) {
	ctx.Target.SetLinearVelocity(dprec.ZeroVec3())
}

// ApplyNudges implements [SoloConstraintSolver.ApplyNudges].
//
// It unconditionally sets the target's position to Position.
func (s *FixedPositionSolver) ApplyNudges(ctx SoloConstraintContext) {
	ctx.Target.SetPosition(s.position)
}
