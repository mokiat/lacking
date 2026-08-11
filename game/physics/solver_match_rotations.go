package physics

import "github.com/mokiat/gomath/dprec"

// MatchRotationsSolver is a [PairConstraintSolver] that rotates its two
// target bodies so that their local X, Y and Z axes each become
// pairwise parallel to one another - driving the two targets toward a
// shared orientation.
//
// Each axis is matched as a line, without regard for which of its two
// directions it points, so the constraint's stable solutions include
// not just an exact orientation match, but also any relative
// orientation reachable from it by a 180-degree rotation about a shared
// axis - four equilibria in total. In practice, provided the two
// targets start out reasonably close to already being aligned, the
// constraint converges to the exact match rather than to one of the
// 180-degree-off alternatives.
//
// Unlike [CopyRotationSolver], which unconditionally overwrites the
// primary target's rotation and angular velocity with the secondary
// target's on every step, MatchRotationsSolver is a genuine two-way
// dynamic constraint: it only applies corrective impulses and nudges,
// scaled by each target's mass and inertia, so external torques can
// still influence the primary target's motion, and either target can
// push back against the other.
type MatchRotationsSolver struct {
	axisXSolver *MatchAxesSolver
	axisYSolver *MatchAxesSolver
	axisZSolver *MatchAxesSolver
}

var _ PairConstraintSolver = (*MatchRotationsSolver)(nil)

// NewMatchRotationsSolver creates a new [MatchRotationsSolver], ready to
// match the primary and secondary targets' orientations as soon as it
// is registered with a [Scene] through [PairConstraintView.Create].
//
// MatchRotationsSolver holds no configurable state of its own, so
// unlike most other solvers in this package, it has no Config type or
// Configure method; NewMatchRotationsSolver is the only way to obtain
// one.
func NewMatchRotationsSolver() *MatchRotationsSolver {
	return &MatchRotationsSolver{
		axisXSolver: NewMatchAxesSolver(MatchAxesSolverConfig{
			PrimaryBodyAxis:   dprec.BasisXVec3(),
			SecondaryBodyAxis: dprec.BasisXVec3(),
		}),
		axisYSolver: NewMatchAxesSolver(MatchAxesSolverConfig{
			PrimaryBodyAxis:   dprec.BasisYVec3(),
			SecondaryBodyAxis: dprec.BasisYVec3(),
		}),
		axisZSolver: NewMatchAxesSolver(MatchAxesSolverConfig{
			PrimaryBodyAxis:   dprec.BasisZVec3(),
			SecondaryBodyAxis: dprec.BasisZVec3(),
		}),
	}
}

// Reset implements [PairConstraintSolver.Reset].
//
// It recomputes the constraint's internal state - the Jacobians and
// drift needed to correct any remaining misalignment between the
// targets' axes - from the targets' current rotations.
func (s *MatchRotationsSolver) Reset(ctx PairConstraintContext) {
	s.axisXSolver.Reset(ctx)
	s.axisYSolver.Reset(ctx)
	s.axisZSolver.Reset(ctx)
}

// ApplyImpulses implements [PairConstraintSolver.ApplyImpulses].
//
// It resolves impulses, without restitution, that drive the two
// targets' relative angular velocity toward bringing their axes back
// into alignment, based on the state [MatchRotationsSolver.Reset]
// computed.
func (s *MatchRotationsSolver) ApplyImpulses(ctx PairConstraintContext) {
	s.axisXSolver.ApplyImpulses(ctx)
	s.axisYSolver.ApplyImpulses(ctx)
	s.axisZSolver.ApplyImpulses(ctx)
}

// ApplyNudges implements [PairConstraintSolver.ApplyNudges].
//
// It first recomputes the constraint's internal state, the same way
// [MatchRotationsSolver.Reset] does, since a preceding nudge - by this
// solver's own previous iteration, or by another constraint acting on
// either target - may have rotated either target since
// [MatchRotationsSolver.Reset] or the last call to this method. It then
// nudges both targets' rotations to reduce any remaining misalignment
// between their axes.
func (s *MatchRotationsSolver) ApplyNudges(ctx PairConstraintContext) {
	s.axisXSolver.ApplyNudges(ctx)
	s.axisYSolver.ApplyNudges(ctx)
	s.axisZSolver.ApplyNudges(ctx)
}
