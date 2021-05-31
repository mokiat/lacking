package physics

import "github.com/mokiat/gomath/sprec"

// ConstraintContext contains information related to the
// constraint processing.
type ConstraintContext struct {
	ElapsedSeconds float32
}

// SBConstraint represents a restriction enforced on one body.
type SBConstraint struct {
	solver SBConstraintSolver

	scene *Scene
	prev  *SBConstraint
	next  *SBConstraint

	body *Body
}

// Solver returns the constraint solver that will be used
// to enforce mathematically this constraint.
func (c *SBConstraint) Solver() SBConstraintSolver {
	return c.solver
}

// Body returns the body on which this constraint acts.
func (c *SBConstraint) Body() *Body {
	return c.body
}

// Enabled returns whether this constraint will be enforced.
// By default a constraint is enabled.
func (c *SBConstraint) Enabled() bool {
	if c.prev != nil || c.next != nil {
		return true
	}
	if c.scene != nil {
		return c.scene.firstSBConstraint == c || c.scene.lastSBConstraint == c
	}
	return false
}

// SetEnabled changes whether this constraint will be enforced.
func (c *SBConstraint) SetEnabled(enabled bool) {
	switch enabled {
	case true:
		c.scene.appendSBConstraint(c)
	case false:
		c.scene.removeSBConstraint(c)
	}
}

// Delete removes this constraint.
func (c *SBConstraint) Delete() {
	c.scene.removeSBConstraint(c)
	c.scene.cacheSBConstraint(c)
	c.scene = nil
	c.body = nil
	c.solver = nil
}

// SBConstraintSolver represents the algorithm necessary
// to enforce a single-body constraint.
type SBConstraintSolver interface {

	// Reset clears the internal cache state for this constraint solver.
	// This is called at the start of every iteration.
	Reset()

	// CalculateImpulses returns a set of impulses to be applied
	// to the body.
	CalculateImpulses(body *Body, ctx ConstraintContext) SBImpulseSolution

	// CalculateNudges returns a set of nudges to be applied
	// to the body.
	CalculateNudges(body *Body, ctx ConstraintContext) SBNudgeSolution
}

// SBImpulseSolution is a solution to a single-body constraint that
// contains the impulses that need to be applied to the body.
type SBImpulseSolution struct {
	Impulse        sprec.Vec3
	AngularImpulse sprec.Vec3
}

// SBNudgeSolution is a solution to a single-body constraint that
// contains the nudges that need to be applied to the body.
type SBNudgeSolution struct {
	Nudge        sprec.Vec3
	AngularNudge sprec.Vec3
}

// DBConstraint represents a restriction enforced two bodies in conjunction.
type DBConstraint struct {
	solver DBConstraintSolver

	scene *Scene
	prev  *DBConstraint
	next  *DBConstraint

	primary   *Body
	secondary *Body
}

// Solver returns the constraint solver that will be used
// to enforce mathematically this constraint.
func (c *DBConstraint) Solver() DBConstraintSolver {
	return c.solver
}

// Primary returns the primary body on which this constraint
// acts.
func (c *DBConstraint) PrimaryBody() *Body {
	return c.primary
}

// SecondaryBody returns the secondary body on which this constraint
// acts. If this is a single-body constraint, then this will be nil.
func (c *DBConstraint) SecondaryBody() *Body {
	return c.secondary
}

// Enabled returns whether this constraint will be enforced.
// By default a constraint is enabled.
func (c *DBConstraint) Enabled() bool {
	if c.prev != nil || c.next != nil {
		return true
	}
	if c.scene != nil {
		return c.scene.firstDBConstraint == c || c.scene.lastDBConstraint == c
	}
	return false
}

// SetEnabled changes whether this constraint will be enforced.
func (c *DBConstraint) SetEnabled(enabled bool) {
	switch enabled {
	case true:
		c.scene.appendDBConstraint(c)
	case false:
		c.scene.removeDBConstraint(c)
	}
}

// Delete removes this constraint.
func (c *DBConstraint) Delete() {
	c.scene.removeDBConstraint(c)
	c.scene.cacheDBConstraint(c)
	c.scene = nil
	c.primary = nil
	c.secondary = nil
	c.solver = nil
}

// DBConstraintSolver represents the algorithm necessary to enforce
// a double-body constraint.
type DBConstraintSolver interface {

	// Reset clears the internal cache state for this constraint solver.
	// This is called at the start of every iteration.
	Reset()

	// CalculateImpulses returns a set of impulses to be applied
	// to the two bodies.
	CalculateImpulses(primary, secondary *Body, ctx ConstraintContext) DBImpulseSolution

	// CalculateNudges returns a set of nudges to be applied
	// to the two bodies.
	CalculateNudges(primary, secondary *Body, ctx ConstraintContext) DBNudgeSolution
}

// DBImpulseSolution is a solution to a constraint that
// indicates the impulses that need to be applied to the primary body
// and optionally (if the body is not nil) secondary body.
type DBImpulseSolution struct {
	Primary   SBImpulseSolution
	Secondary SBImpulseSolution
}

// DBNudgeSolution is a solution to a constraint that
// indicates the nudges that need to be applied to the primary body
// and optionally (if the body is not nil) secondary body.
type DBNudgeSolution struct {
	Primary   SBNudgeSolution
	Secondary SBNudgeSolution
}

var _ SBConstraintSolver = (*NilSBConstraintSolver)(nil)

// NilConstraintSolver is a ConstraintSolver that does nothing.
type NilSBConstraintSolver struct{}

func (s *NilSBConstraintSolver) Reset() {}

func (s *NilSBConstraintSolver) CalculateImpulses(body *Body, ctx ConstraintContext) SBImpulseSolution {
	return SBImpulseSolution{}
}

func (s *NilSBConstraintSolver) CalculateNudges(body *Body, ctx ConstraintContext) SBNudgeSolution {
	return SBNudgeSolution{}
}

var _ DBConstraintSolver = (*NilDBConstraintSolver)(nil)

// NilConstraintSolver is a ConstraintSolver that does nothing.
type NilDBConstraintSolver struct{}

func (s *NilDBConstraintSolver) Reset() {}

func (s *NilDBConstraintSolver) CalculateImpulses(primary, secondary *Body, ctx ConstraintContext) DBImpulseSolution {
	return DBImpulseSolution{}
}

func (s *NilDBConstraintSolver) CalculateNudges(primary, secondary *Body, ctx ConstraintContext) DBNudgeSolution {
	return DBNudgeSolution{}
}
