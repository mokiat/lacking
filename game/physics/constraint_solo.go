package physics

import "github.com/mokiat/lacking/game/physics/solver"

type SoloConstraintContext struct {
	DeltaSeconds float64
	ImpulseBeta  float64
	NudgeBeta    float64
	Target       *solver.Placeholder // TODO: use package-local ImpulseTarget instead of Placeholder
}

type SoloConstraintSolver interface {
	// Reset clears the internal cache state for this constraint solver.
	//
	// This is called at the start of every iteration.
	Reset(ctx SoloConstraintContext)

	// ApplyImpulses is called by the physics engine to instruct the solver
	// to apply the necessary impulses to its object.
	//
	// This is called multiple times per iteration.
	ApplyImpulses(ctx SoloConstraintContext)

	// ApplyNudges is called by the physics engine to instruct the solver to
	// apply the necessary nudges to its object.
	//
	// This is called multiple times per iteration.
	ApplyNudges(ctx SoloConstraintContext)
}

type SoloConstraintID struct {
	index    int32
	revision int32
}

var NilSoloConstraintID = SoloConstraintID{}

type SoloConstraintView struct {
	scene *Scene
}

func (v SoloConstraintView) Create(bodyID BodyID, solver SoloConstraintSolver) SoloConstraintID {
	bodyView := v.scene.Bodies()
	body := bodyView.resolve(bodyID, true)

	index, constraint := v.scene.allocateSoloConstraint()
	*constraint = soloConstraint{
		solver:    solver,
		revision:  constraint.revision + 1, // progress revision to valid (odd) value
		bodyIndex: bodyID.index,
		nextIndex: body.firstSoloConstraintIndex,
		isEnabled: true,
	}
	body.firstSoloConstraintIndex = index

	return SoloConstraintID{
		index:    index,
		revision: constraint.revision,
	}
}

func (v SoloConstraintView) Delete(id SoloConstraintID) {
	constraint := v.resolve(id, true)

	body := &v.scene.bodies[constraint.bodyIndex]
	if body.firstSoloConstraintIndex == id.index {
		body.firstSoloConstraintIndex = constraint.nextIndex
	} else {
		prevIndex := body.firstSoloConstraintIndex
		for prevIndex != nilIndex {
			prev := &v.scene.soloConstraints[prevIndex]
			if prev.nextIndex == id.index {
				prev.nextIndex = constraint.nextIndex
				break
			}
			prevIndex = prev.nextIndex
		}
	}

	*constraint = soloConstraint{
		solver:    nil,                     // allow the solver to be garbage collected
		revision:  constraint.revision + 1, // progress revision to invalid (even) value
		bodyIndex: nilIndex,
		nextIndex: nilIndex,
		isEnabled: false,
	}

	v.scene.releaseSoloConstraint(id.index)
}

func (v SoloConstraintView) Handle(id SoloConstraintID) SoloConstraintHandle {
	return SoloConstraintHandle{
		view: v,
		id:   id,
	}
}

func (v SoloConstraintView) IsValid(id SoloConstraintID) bool {
	constraint := v.resolve(id, false)
	return constraint != nil
}

func (v SoloConstraintView) BodyID(id SoloConstraintID) BodyID {
	constraint := v.resolve(id, true)
	bodyIndex := constraint.bodyIndex
	body := &v.scene.bodies[bodyIndex]
	return BodyID{
		index:    bodyIndex,
		revision: body.revision,
	}
}

func (v SoloConstraintView) Solver(id SoloConstraintID) SoloConstraintSolver {
	constraint := v.resolve(id, true)
	return constraint.solver
}

func (v SoloConstraintView) SetSolver(id SoloConstraintID, solver SoloConstraintSolver) {
	constraint := v.resolve(id, true)
	constraint.solver = solver
}

func (v SoloConstraintView) Enabled(id SoloConstraintID) bool {
	constraint := v.resolve(id, true)
	return constraint.isEnabled
}

func (v SoloConstraintView) SetEnabled(id SoloConstraintID, enabled bool) {
	constraint := v.resolve(id, true)
	constraint.isEnabled = enabled
}

func (v SoloConstraintView) idFromIndex(index int32) SoloConstraintID {
	constraint := &v.scene.soloConstraints[index]
	return SoloConstraintID{
		index:    index,
		revision: constraint.revision,
	}
}

func (v SoloConstraintView) resolve(id SoloConstraintID, required bool) *soloConstraint {
	if id.revision == 0 {
		if required {
			panic("invalid solo constraint ID")
		}
		return nil
	}
	constraint := &v.scene.soloConstraints[id.index]
	if constraint.revision != id.revision {
		if required {
			panic("invalid solo constraint ID")
		}
		return nil
	}
	return constraint
}

type SoloConstraintHandle struct {
	view SoloConstraintView
	id   SoloConstraintID
}

func (h SoloConstraintHandle) ID() SoloConstraintID {
	return h.id
}

func (h SoloConstraintHandle) Delete() {
	h.view.Delete(h.id)
}

func (h SoloConstraintHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

func (h SoloConstraintHandle) BodyID() BodyID {
	return h.view.BodyID(h.id)
}

func (h SoloConstraintHandle) Solver() SoloConstraintSolver {
	return h.view.Solver(h.id)
}

func (h SoloConstraintHandle) SetSolver(solver SoloConstraintSolver) {
	h.view.SetSolver(h.id, solver)
}

func (h SoloConstraintHandle) Enabled() bool {
	return h.view.Enabled(h.id)
}

func (h SoloConstraintHandle) SetEnabled(enabled bool) {
	h.view.SetEnabled(h.id, enabled)
}

type soloConstraint struct {
	solver    SoloConstraintSolver
	revision  int32
	bodyIndex int32
	nextIndex int32
	isEnabled bool
}
