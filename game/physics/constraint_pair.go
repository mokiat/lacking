package physics

type PairConstraintContext struct {
	DeltaSeconds    float64
	ImpulseBeta     float64
	NudgeBeta       float64
	PrimaryTarget   ConstraintTarget
	SecondaryTarget ConstraintTarget
}

type PairConstraintSolver interface {
	Reset(ctx PairConstraintContext)

	ApplyImpulses(ctx PairConstraintContext)

	ApplyNudges(ctx PairConstraintContext)
}

type PairConstraintID struct {
	index    int32
	revision int32
}

var NilPairConstraintID = PairConstraintID{}

type PairConstraintView struct {
	scene *Scene
}

func (v PairConstraintView) Create(primaryID, secondaryID BodyID, solver PairConstraintSolver) PairConstraintID {
	bodyView := v.scene.Bodies()
	primaryBody := bodyView.resolve(primaryID, true)
	secondaryBody := bodyView.resolve(secondaryID, true)

	index, constraint := v.scene.allocatePairConstraint()

	*constraint = pairConstraintState{
		solver:             solver,
		revision:           constraint.revision + 1, // progress revision to valid (odd) value
		primaryBodyIndex:   primaryID.index,
		secondaryBodyIndex: secondaryID.index,
		primaryNextIndex:   primaryBody.firstSoloConstraintIndex,
		secondaryNextIndex: secondaryBody.firstSoloConstraintIndex,
		isEnabled:          true,
	}
	primaryBody.firstSoloConstraintIndex = index
	secondaryBody.firstSoloConstraintIndex = index

	return PairConstraintID{
		index:    index,
		revision: constraint.revision,
	}
}

func (v PairConstraintView) Delete(id PairConstraintID) {
	constraint := v.resolve(id, true)

	// Unlink the constraint from the primary body.
	primaryBody := &v.scene.bodies[constraint.primaryBodyIndex]
	if primaryBody.firstPairConstraintIndex == constraint.primaryNextIndex {
		primaryBody.firstPairConstraintIndex = constraint.primaryNextIndex
	} else {
		prevIndex := primaryBody.firstPairConstraintIndex
		for prevIndex != -1 {
			prevConstraint := &v.scene.pairConstraints[prevIndex]
			if prevConstraint.primaryNextIndex == constraint.primaryNextIndex {
				prevConstraint.primaryNextIndex = constraint.primaryNextIndex
				break
			}
			prevIndex = prevConstraint.primaryNextIndex
		}
	}

	// Unlink the constraint from the secondary body.
	secondaryBody := &v.scene.bodies[constraint.secondaryBodyIndex]
	if secondaryBody.firstPairConstraintIndex == constraint.secondaryNextIndex {
		secondaryBody.firstPairConstraintIndex = constraint.secondaryNextIndex
	} else {
		prevIndex := secondaryBody.firstPairConstraintIndex
		for prevIndex != -1 {
			prevConstraint := &v.scene.pairConstraints[prevIndex]
			if prevConstraint.secondaryNextIndex == constraint.secondaryNextIndex {
				prevConstraint.secondaryNextIndex = constraint.secondaryNextIndex
				break
			}
			prevIndex = prevConstraint.secondaryNextIndex
		}
	}

	*constraint = pairConstraintState{
		solver:             nil,                     // allow the solver to be garbage collected
		revision:           constraint.revision + 1, // progress revision to invalid (even) value
		primaryBodyIndex:   nilIndex,
		secondaryBodyIndex: nilIndex,
		primaryNextIndex:   nilIndex,
		secondaryNextIndex: nilIndex,
		isEnabled:          false,
	}

	v.scene.releasePairConstraint(id.index)
}

func (v PairConstraintView) Handle(id PairConstraintID) PairConstraintHandle {
	return PairConstraintHandle{
		view: v,
		id:   id,
	}
}

func (v PairConstraintView) IsValid(id PairConstraintID) bool {
	constraint := v.resolve(id, false)
	return constraint != nil
}

func (v PairConstraintView) PrimaryBodyID(id PairConstraintID) BodyID {
	constraint := v.resolve(id, true)
	bodyIndex := constraint.primaryBodyIndex
	body := &v.scene.bodies[bodyIndex]
	return BodyID{
		index:    bodyIndex,
		revision: body.revision,
	}
}

func (v PairConstraintView) SecondaryBodyID(id PairConstraintID) BodyID {
	constraint := v.resolve(id, true)
	bodyIndex := constraint.secondaryBodyIndex
	body := &v.scene.bodies[bodyIndex]
	return BodyID{
		index:    bodyIndex,
		revision: body.revision,
	}
}

func (v PairConstraintView) Solver(id PairConstraintID) PairConstraintSolver {
	constraint := v.resolve(id, true)
	return constraint.solver
}

func (v PairConstraintView) SetSolver(id PairConstraintID, solver PairConstraintSolver) {
	constraint := v.resolve(id, true)
	constraint.solver = solver
}

func (v PairConstraintView) Enabled(id PairConstraintID) bool {
	constraint := v.resolve(id, true)
	return constraint.isEnabled
}

func (v PairConstraintView) SetEnabled(id PairConstraintID, enabled bool) {
	constraint := v.resolve(id, true)
	constraint.isEnabled = enabled
}

func (v PairConstraintView) idFromIndex(index int32) PairConstraintID {
	state := &v.scene.pairConstraints[index]
	return PairConstraintID{
		index:    index,
		revision: state.revision,
	}
}

func (v PairConstraintView) resolve(id PairConstraintID, required bool) *pairConstraintState {
	if id.revision == 0 {
		if required {
			panic("invalid pair constraint ID")
		}
		return nil
	}
	constraint := &v.scene.pairConstraints[id.index]
	if constraint.revision != id.revision {
		if required {
			panic("invalid pair constraint ID")
		}
		return nil
	}
	return constraint
}

type PairConstraintHandle struct {
	view PairConstraintView
	id   PairConstraintID
}

func (h PairConstraintHandle) ID() PairConstraintID {
	return h.id
}

func (h PairConstraintHandle) Delete() {
	h.view.Delete(h.id)
}

func (h PairConstraintHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

func (h PairConstraintHandle) PrimaryBodyID() BodyID {
	return h.view.PrimaryBodyID(h.id)
}

func (h PairConstraintHandle) SecondaryBodyID() BodyID {
	return h.view.SecondaryBodyID(h.id)
}

func (h PairConstraintHandle) Solver() PairConstraintSolver {
	return h.view.Solver(h.id)
}

func (h PairConstraintHandle) SetSolver(solver PairConstraintSolver) {
	h.view.SetSolver(h.id, solver)
}

func (h PairConstraintHandle) Enabled() bool {
	return h.view.Enabled(h.id)
}

func (h PairConstraintHandle) SetEnabled(enabled bool) {
	h.view.SetEnabled(h.id, enabled)
}

type pairConstraintState struct {
	solver             PairConstraintSolver
	revision           int32
	primaryBodyIndex   int32
	secondaryBodyIndex int32
	primaryNextIndex   int32
	secondaryNextIndex int32
	isEnabled          bool
}

func (s *pairConstraintState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
