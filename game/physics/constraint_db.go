package physics

import "github.com/mokiat/lacking/game/physics/solver"

var invalidDBConstraintState = &dbConstraintState{}

// DBConstraint represents a restriction enforced on two bodies in conjunction.
type DBConstraint struct {
	scene     *Scene
	reference indexReference
}

// Enabled returns whether this constraint will be enforced.
func (c DBConstraint) Enabled() bool {
	state := c.state()
	return state.enabled
}

// SetEnabled changes whether this constraint will be enforced.
func (c DBConstraint) SetEnabled(enabled bool) {
	state := c.state()
	state.enabled = enabled
}

// Logic returns the constraint solver that will be used to enforce
// mathematically this constraint.
func (c DBConstraint) Logic() solver.PairConstraint {
	state := c.state()
	return state.logic
}

// PrimaryBody returns the primary body on which this constraint
// acts.
func (c DBConstraint) PrimaryBody() Body {
	state := c.state()
	return state.primary
}

// SecondaryBody returns the secondary body on which this constraint
// acts.
func (c DBConstraint) SecondaryBody() Body {
	state := c.state()
	return state.secondary
}

// Delete removes this constraint.
func (c DBConstraint) Delete() {
	deleteDBConstraint(c.scene, c.reference)
}

func (c DBConstraint) state() *dbConstraintState {
	index := c.reference.Index
	state := &c.scene.dbConstraints[index]
	if state.reference != c.reference {
		return invalidDBConstraintState
	}
	return state
}

type dbConstraintState struct {
	reference indexReference
	logic     solver.PairConstraint
	primary   Body
	secondary Body
	enabled   bool
}

func (s dbConstraintState) IsActive() bool {
	return s.reference.IsValid() && s.enabled
}

func createDBConstraint(scene *Scene, logic solver.PairConstraint, primary, secondary Body) DBConstraint {
	var freeIndex uint32
	if scene.freeDBConstraintIndices.IsEmpty() {
		freeIndex = uint32(len(scene.dbConstraints))
		scene.dbConstraints = append(scene.dbConstraints, dbConstraintState{})
	} else {
		freeIndex = scene.freeDBConstraintIndices.Pop()
	}

	reference := newIndexReference(freeIndex, scene.nextRevision())
	scene.dbConstraints[freeIndex] = dbConstraintState{
		reference: reference,
		logic:     logic,
		primary:   primary,
		secondary: secondary,
		enabled:   true,
	}
	return DBConstraint{
		scene:     scene,
		reference: reference,
	}
}

func deleteDBConstraint(scene *Scene, reference indexReference) {
	index := reference.Index
	state := &scene.dbConstraints[index]
	if state.reference == reference {
		state.reference = newIndexReference(index, 0)
		state.logic = nil
		scene.freeDBConstraintIndices.Push(index)
	}
}
