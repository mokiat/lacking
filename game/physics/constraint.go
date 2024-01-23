package physics

import "github.com/mokiat/lacking/game/physics/solver"

// DBConstraint represents a restriction enforced two bodies in conjunction.
type DBConstraint struct {
	scene *Scene
	prev  *DBConstraint
	next  *DBConstraint

	primary   *Body
	secondary *Body

	solution solver.PairConstraint
}

// Solution returns the constraint solver that will be used
// to enforce mathematically this constraint.
func (c *DBConstraint) Solution() solver.PairConstraint {
	return c.solution
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
	c.solution = nil
}

// ConstraintSet represents a set of constraints.
// This type is useful when multiple constraints need to
// be managed (enabled,disabled,deleted) in parallel.
type ConstraintSet struct {
	scene         *Scene
	sbConstraints []SBConstraint
	dbConstraints []*DBConstraint
}

// CreateSingleBodyConstraint creates a new physics constraint that acts on
// a single body and stores it in this set.
//
// Note: Constraints creates as part of this set should not be deleted
// individually.
func (s *ConstraintSet) CreateSingleBodyConstraint(body *Body, solver solver.Constraint) SBConstraint {
	constraint := s.scene.CreateSingleBodyConstraint(body, solver)
	s.sbConstraints = append(s.sbConstraints, constraint)
	return constraint
}

// CreateDoubleBodyConstraint creates a new physics constraint that acts on
// two bodies and enables it for this scene.
//
// Note: Constraints creates as part of this set should not be deleted
// individually.
func (s *ConstraintSet) CreateDoubleBodyConstraint(primary, secondary *Body, solver solver.PairConstraint) *DBConstraint {
	constraint := s.scene.CreateDoubleBodyConstraint(primary, secondary, solver)
	s.dbConstraints = append(s.dbConstraints, constraint)
	return constraint
}

// Enabled returns whether at least one of the constraints
// in this set is enabled.
func (s *ConstraintSet) Enabled() bool {
	for _, constraint := range s.sbConstraints {
		if constraint.Enabled() {
			return true
		}
	}
	for _, constraint := range s.dbConstraints {
		if constraint.Enabled() {
			return true
		}
	}
	return false
}

// SetEnabled changes the enabled state of all
// constraints in this set.
func (s *ConstraintSet) SetEnabled(enabled bool) {
	for _, constraint := range s.sbConstraints {
		constraint.SetEnabled(enabled)
	}
	for _, constraint := range s.dbConstraints {
		constraint.SetEnabled(enabled)
	}
}

// Delete deletes all contained constraints and this
// set.
func (s *ConstraintSet) Delete() {
	for _, constraint := range s.sbConstraints {
		constraint.Delete()
	}
	for _, constraint := range s.dbConstraints {
		constraint.Delete()
	}
	s.scene = nil
}
