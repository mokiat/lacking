package physics

// CopyRotationSolver is a [PairConstraintSolver] that makes the primary
// target follow the secondary target's rotation, regardless of any
// forces or torques acting on the primary target.
//
// It is a kinematic constraint, in the same vein as
// [FixedRotationSolver], except that it tracks a moving secondary target
// instead of a constant world-space rotation. Through
// [CopyRotationSolver.ApplyImpulses] it unconditionally overwrites the
// primary target's angular velocity with the secondary target's, and
// through [CopyRotationSolver.ApplyNudges] it unconditionally overwrites
// the primary target's rotation with the secondary target's, on every
// step. The secondary target itself is never modified. Only rotational
// motion is copied - see [CopyPositionSolver] for the positional
// counterpart.
//
// CopyRotationSolver holds no configurable state, so unlike most other
// solvers in this package, it has no Config type or Configure method;
// [NewCopyRotationSolver] is the only way to obtain one, and a single
// instance can safely back any number of constraints.
type CopyRotationSolver struct{}

var _ PairConstraintSolver = (*CopyRotationSolver)(nil)

// NewCopyRotationSolver creates a new [CopyRotationSolver].
func NewCopyRotationSolver() *CopyRotationSolver {
	return &CopyRotationSolver{}
}

// Reset implements [PairConstraintSolver.Reset].
//
// It is a no-op, since this solver holds no per-step state that needs to
// be derived from the targets' current rotations or velocities.
func (s *CopyRotationSolver) Reset(ctx PairConstraintContext) {}

// ApplyImpulses implements [PairConstraintSolver.ApplyImpulses].
//
// It unconditionally overwrites the primary target's angular velocity
// with the secondary target's.
func (s *CopyRotationSolver) ApplyImpulses(ctx PairConstraintContext) {
	ctx.PrimaryTarget.SetAngularVelocity(ctx.SecondaryTarget.AngularVelocity())
}

// ApplyNudges implements [PairConstraintSolver.ApplyNudges].
//
// It unconditionally overwrites the primary target's rotation with the
// secondary target's.
func (s *CopyRotationSolver) ApplyNudges(ctx PairConstraintContext) {
	ctx.PrimaryTarget.SetRotation(ctx.SecondaryTarget.Rotation())
}
