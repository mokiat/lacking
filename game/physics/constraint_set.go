package physics

import "github.com/mokiat/lacking/game/physics/solver"

// ConstraintSet represents a set of constraints.
//
// This type is useful when multiple constraints need to
// be managed (enabled,disabled,deleted) as a single unit.
type ConstraintSet struct {
	scene         *Scene
	sbConstraints []SBConstraint
	dbConstraints []DBConstraint
}

// CreateSingleBodyConstraint creates a new physics constraint that acts on
// a single body and stores it in this set.
//
// Note: Constraints creates as part of this set should not be deleted
// individually.
func (s *ConstraintSet) CreateSingleBodyConstraint(body Body, solver solver.Constraint) SBConstraint {
	constraint := s.scene.CreateSingleBodyConstraint(body, solver)
	s.sbConstraints = append(s.sbConstraints, constraint)
	return constraint
}

// CreateDoubleBodyConstraint creates a new physics constraint that acts on
// two bodies and enables it for this scene.
//
// Note: Constraints creates as part of this set should not be deleted
// individually.
func (s *ConstraintSet) CreateDoubleBodyConstraint(primary, secondary Body, solver solver.PairConstraint) DBConstraint {
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
