package physics

import "github.com/mokiat/gomath/sprec"

// Constraint represents a restriction enforced on one body on its own
// or on two bodies in conjunction.
type Constraint struct {
	solver ConstraintSolver

	scene *Scene
	prev  *Constraint
	next  *Constraint

	enabled bool

	primary   *Body
	secondary *Body
}

// Solver returns the constraint solver that will be used
// to enforce mathematically this constraint.
func (c *Constraint) Solver() ConstraintSolver {
	return c.solver
}

// Primary returns the primary body on which this constraint
// acts.
func (c *Constraint) PrimaryBody() *Body {
	return c.primary
}

// SecondaryBody returns the secondary body on which this constraint
// acts. If this is a single-body constraint, then this will be nil.
func (c *Constraint) SecondaryBody() *Body {
	return c.secondary
}

// Enabled returns whether this constraint will be enforced.
// By default a constraint is enabled.
func (c *Constraint) Enabled() bool {
	return c.enabled
}

// SetEnabled changes whether this constraint will be enforced.
func (c *Constraint) SetEnabled(enabled bool) {
	if enabled == c.enabled {
		return
	}
	c.enabled = enabled
	switch enabled {
	case true:
		c.scene.appendConstraint(c)
	case false:
		c.scene.removeConstraint(c)
	}
}

// Delete removes this constraint.
func (c *Constraint) Delete() {
	c.scene.removeConstraint(c)
	c.scene = nil
	c.primary = nil
	c.secondary = nil
}

// ConstraintSolver represents the algorithm necessary
// to enforce the constraint.
type ConstraintSolver interface {

	// Reset clears the internal cache state for this constraint solver.
	// This is called at the start of every iteration.
	Reset()

	// CalculateImpulses returns a set of impulses to be applied
	// to the primary and optionally the secondary body.
	CalculateImpulses(primary, secondary *Body, elapsedSeconds float32) ConstraintImpulseSolution

	// CalculateNudges returns a set of nudges to be applied
	// to the primary and optionally the secondary body.
	CalculateNudges(primary, secondary *Body, elapsedSeconds float32) ConstraintNudgeSolution
}

var _ ConstraintSolver = (*NilConstraintSolver)(nil)

// NilConstraintSolver is a ConstraintSolver that does nothing.
type NilConstraintSolver struct{}

func (s *NilConstraintSolver) Reset() {
}

func (s *NilConstraintSolver) CalculateImpulses(primary, secondary *Body, elapsedSeconds float32) ConstraintImpulseSolution {
	return ConstraintImpulseSolution{}
}

func (s *NilConstraintSolver) CalculateNudges(primary, secondary *Body, elapsedSeconds float32) ConstraintNudgeSolution {
	return ConstraintNudgeSolution{}
}

// ConstraintImpulseSolution is a solution to a constraint that
// indicates the impulses that need to be applied to the primary body
// and optionally (if the body is not nil) secondary body.
type ConstraintImpulseSolution struct {
	PrimaryImpulse          sprec.Vec3
	PrimaryAngularImpulse   sprec.Vec3
	SecondaryImpulse        sprec.Vec3
	SecondaryAngularImpulse sprec.Vec3
}

// ConstraintNudgeSolution is a solution to a constraint that
// indicates the nudges that need to be applied to the primary body
// and optionally (if the body is not nil) secondary body.
type ConstraintNudgeSolution struct {
	PrimaryNudge          sprec.Vec3
	PrimaryAngularNudge   sprec.Vec3
	SecondaryNudge        sprec.Vec3
	SecondaryAngularNudge sprec.Vec3
}
