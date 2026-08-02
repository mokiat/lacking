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
	panic("TODO")
}

func (v PairConstraintView) Delete(id PairConstraintID) {
	panic("TODO")
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

type pairConstraintState struct {
	solver             PairConstraintSolver
	revision           int32
	primaryBodyIndex   int32
	secondaryBodyIndex int32
	nextIndex          int32
	isEnabled          bool
}
