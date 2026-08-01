package physics

import "github.com/mokiat/lacking/game/physics/solver"

type SoloConstraintContext struct {
	DeltaSeconds float64
	ImpulseBeta  float64
	NudgeBeta    float64

	Target *solver.Placeholder // TODO: use package-local ImpulseTarget instead of Placeholder
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
	// TODO: verify that bodyID is valid and belongs to this scene

	index := v.scene.allocateSoloConstraint()

	constraint := &v.scene.soloConstraints[index]
	constraint.solver = solver
	constraint.revision++ // progress revision to valid (odd) value
	constraint.bodyIndex = bodyID.index
	constraint.enabled = true

	return SoloConstraintID{
		index:    index,
		revision: constraint.revision,
	}
}

func (v SoloConstraintView) Delete(id SoloConstraintID) {
	constraint := v.resolve(id, true)
	constraint.solver = nil // allow the solver to be garbage collected
	constraint.revision++   // progress revision to invalid (even) value

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
	body := &v.scene.bodies[constraint.bodyIndex]
	return BodyID{
		index:    constraint.bodyIndex,
		revision: int32(body.reference.Revision),
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
	return constraint.enabled
}

func (v SoloConstraintView) SetEnabled(id SoloConstraintID, enabled bool) {
	constraint := v.resolve(id, true)
	constraint.enabled = enabled
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

// TODO: Add methods to SoloConstraintHandle.

type soloConstraint struct {
	solver    SoloConstraintSolver
	revision  int32
	bodyIndex int32
	enabled   bool
}
