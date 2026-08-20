package physics

import "github.com/mokiat/gomath/dprec"

// FixedRotationSolverConfig holds the parameters with which a
// [FixedRotationSolver] is configured, either through
// [NewFixedRotationSolver] or [FixedRotationSolver.Configure].
type FixedRotationSolverConfig struct {

	// Rotation is the world-space orientation at which the target is to
	// be held fixed.
	Rotation dprec.Quat
}

// FixedRotationSolver is a [SoloConstraintSolver] that pins a target to a
// fixed rotation in world space, regardless of any forces or torques
// acting on it.
//
// Unlike [SoloCollisionSolver], which only nudges its target enough to
// resolve penetration, FixedRotationSolver is a kinematic constraint - it
// unconditionally zeroes the target's angular velocity through
// [FixedRotationSolver.ApplyImpulses] and snaps it to Rotation through
// [FixedRotationSolver.ApplyNudges] on every step. It does not affect
// linear velocity or position - see [FixedPositionSolver] for that.
//
// A FixedRotationSolver must be configured, either through
// [NewFixedRotationSolver] or [FixedRotationSolver.Configure], before
// being registered with a [Scene] through [SoloConstraintView.Create].
type FixedRotationSolver struct {
	rotation dprec.Quat
}

var _ SoloConstraintSolver = (*FixedRotationSolver)(nil)

// NewFixedRotationSolver creates a new [FixedRotationSolver] configured
// according to config.
func NewFixedRotationSolver(config FixedRotationSolverConfig) *FixedRotationSolver {
	result := &FixedRotationSolver{}
	result.Configure(config)
	return result
}

// Configure configures this solver according to config.
//
// Configure must be called before this solver is registered with a
// [Scene] through [SoloConstraintView.Create]. Unlike
// [NewFixedRotationSolver], it can be called on an already-allocated
// solver, which allows solvers to be cached (e.g. in a slice) and
// configured on demand.
func (s *FixedRotationSolver) Configure(config FixedRotationSolverConfig) {
	s.rotation = config.Rotation
}

// Rotation returns the world-space orientation at which the target is
// held fixed.
func (s *FixedRotationSolver) Rotation() dprec.Quat {
	return s.rotation
}

// SetRotation changes the world-space orientation at which the target is
// held fixed.
//
// It returns the solver itself, so that calls can be chained.
func (s *FixedRotationSolver) SetRotation(rotation dprec.Quat) *FixedRotationSolver {
	s.rotation = rotation
	return s
}

// Reset implements [SoloConstraintSolver.Reset].
//
// It is a no-op, since this solver holds no per-step state that needs to
// be derived from the target's current position or velocity.
func (s *FixedRotationSolver) Reset(ctx SoloConstraintContext) {}

// ApplyImpulses implements [SoloConstraintSolver.ApplyImpulses].
//
// It unconditionally zeroes the target's angular velocity.
func (s *FixedRotationSolver) ApplyImpulses(ctx SoloConstraintContext) {
	ctx.Target.SetAngularVelocity(dprec.ZeroVec3())
}

// ApplyNudges implements [SoloConstraintSolver.ApplyNudges].
//
// It unconditionally sets the target's rotation to Rotation.
func (s *FixedRotationSolver) ApplyNudges(ctx SoloConstraintContext) {
	ctx.Target.SetRotation(s.rotation)
}
