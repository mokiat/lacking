package physics

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
