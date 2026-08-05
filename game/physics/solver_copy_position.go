package physics

// CopyPositionSolver is a [PairConstraintSolver] that makes the primary
// target follow the secondary target's position, regardless of any
// forces or impulses acting on the primary target.
//
// It is a kinematic constraint, in the same vein as
// [FixedPositionSolver], except that it tracks a moving secondary target
// instead of a constant world-space position. Through
// [CopyPositionSolver.ApplyImpulses] it unconditionally overwrites the
// primary target's linear velocity with the secondary target's, and
// through [CopyPositionSolver.ApplyNudges] it unconditionally overwrites
// the primary target's position with the secondary target's, on every
// step. The secondary target itself is never modified. Only linear
// motion is copied - the primary target's rotation and angular velocity
// are left to evolve on their own.
//
// CopyPositionSolver holds no configurable state, so unlike most other
// solvers in this package, it has no Config type or Configure method;
// [NewCopyPositionSolver] is the only way to obtain one, and a single
// instance can safely back any number of constraints.
type CopyPositionSolver struct{}

var _ PairConstraintSolver = (*CopyPositionSolver)(nil)

// NewCopyPositionSolver creates a new [CopyPositionSolver].
func NewCopyPositionSolver() *CopyPositionSolver {
	return &CopyPositionSolver{}
}

// Reset implements [PairConstraintSolver.Reset].
//
// It is a no-op, since this solver holds no per-step state that needs to
// be derived from the targets' current positions or velocities.
func (s *CopyPositionSolver) Reset(ctx PairConstraintContext) {}

// ApplyImpulses implements [PairConstraintSolver.ApplyImpulses].
//
// It unconditionally overwrites the primary target's linear velocity
// with the secondary target's.
func (s *CopyPositionSolver) ApplyImpulses(ctx PairConstraintContext) {
	ctx.PrimaryTarget.SetLinearVelocity(ctx.SecondaryTarget.LinearVelocity())
}

// ApplyNudges implements [PairConstraintSolver.ApplyNudges].
//
// It unconditionally overwrites the primary target's position with the
// secondary target's.
func (s *CopyPositionSolver) ApplyNudges(ctx PairConstraintContext) {
	ctx.PrimaryTarget.SetPosition(ctx.SecondaryTarget.Position())
}
