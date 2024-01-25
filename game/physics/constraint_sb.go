package physics

import "github.com/mokiat/lacking/game/physics/solver"

var invalidSBConstraintState = &sbConstraintState{}

// SBConstraint represents a restriction enforced on one body.
type SBConstraint struct {
	scene     *Scene
	reference indexReference
}

// Enabled returns whether this constraint will be enforced.
// By default a constraint is enabled.
func (c SBConstraint) Enabled() bool {
	state := c.state()
	return state.enabled
}

// SetEnabled changes whether this constraint will be enforced.
func (c SBConstraint) SetEnabled(enabled bool) {
	state := c.state()
	state.enabled = enabled
}

// Logic returns the constraint solver that will be used to enforce
// mathematically this constraint.
func (c SBConstraint) Logic() solver.Constraint {
	state := c.state()
	return state.logic
}

// Body returns the body on which this constraint acts.
func (c SBConstraint) Body() Body {
	state := c.state()
	return state.body
}

// Delete removes this constraint.
func (c SBConstraint) Delete() {
	deleteSBConstraint(c.scene, c.reference)
}

func (c SBConstraint) state() *sbConstraintState {
	index := c.reference.Index
	state := &c.scene.sbConstraints[index]
	if state.reference != c.reference {
		return invalidSBConstraintState
	}
	return state
}

type sbConstraintState struct {
	reference indexReference
	logic     solver.Constraint
	body      Body
	enabled   bool
}

func (s sbConstraintState) IsActive() bool {
	return s.reference.IsValid() && s.enabled
}

func createSBConstraint(scene *Scene, logic solver.Constraint, body Body) SBConstraint {
	var freeIndex uint32
	if scene.freeSBConstraintIndices.IsEmpty() {
		freeIndex = uint32(len(scene.sbConstraints))
		scene.sbConstraints = append(scene.sbConstraints, sbConstraintState{})
	} else {
		freeIndex = scene.freeSBConstraintIndices.Pop()
	}

	reference := newIndexReference(freeIndex, scene.nextRevision())
	scene.sbConstraints[freeIndex] = sbConstraintState{
		reference: reference,
		logic:     logic,
		body:      body,
		enabled:   true,
	}
	return SBConstraint{
		scene:     scene,
		reference: reference,
	}
}

func deleteSBConstraint(scene *Scene, reference indexReference) {
	index := reference.Index
	state := &scene.sbConstraints[index]
	if state.reference == reference {
		state.reference = newIndexReference(index, 0)
		state.logic = nil
		scene.freeSBConstraintIndices.Push(index)
	}
}
