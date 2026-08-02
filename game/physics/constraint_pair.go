package physics

// PairConstraintContext contains the information that a [PairConstraintSolver]
// needs in order to process the two bodies it acts upon during a physics
// simulation step.
type PairConstraintContext struct {

	// DeltaSeconds is the amount of time, in seconds, covered by the
	// current physics simulation step.
	DeltaSeconds float64

	// ImpulseBeta is the Baumgarte stabilization factor to be used when
	// correcting positional drift through impulses.
	ImpulseBeta float64

	// NudgeBeta is the Baumgarte stabilization factor to be used when
	// correcting positional drift through nudges.
	NudgeBeta float64

	// PrimaryTarget is the [ConstraintTarget] for the primary body that
	// is constrained, through which the solver reads its motion state
	// and applies corrective impulses and nudges to it.
	PrimaryTarget ConstraintTarget

	// SecondaryTarget is the [ConstraintTarget] for the secondary body
	// that is constrained, through which the solver reads its motion
	// state and applies corrective impulses and nudges to it.
	SecondaryTarget ConstraintTarget
}

// PairConstraintSolver implements the mathematical logic that enforces a
// constraint acting on two bodies simultaneously.
//
// Instances are registered with a [Scene] through [PairConstraintView.Create]
// and are subsequently driven by the physics engine through the methods
// below during each simulation step.
type PairConstraintSolver interface {

	// Reset clears any internal cache state held by the solver, in
	// preparation for a new physics simulation step.
	//
	// This is called once at the start of every step, before
	// [PairConstraintSolver.ApplyImpulses] or [PairConstraintSolver.ApplyNudges]
	// are invoked.
	Reset(ctx PairConstraintContext)

	// ApplyImpulses is called by the physics engine to instruct the
	// solver to apply the necessary impulses to its target bodies, in
	// order to correct their velocities so that the constraint is
	// satisfied.
	//
	// This is called multiple times per step, once for each impulse
	// resolution iteration.
	ApplyImpulses(ctx PairConstraintContext)

	// ApplyNudges is called by the physics engine to instruct the solver
	// to apply the necessary nudges to its target bodies, in order to
	// correct their positions so that the constraint is satisfied.
	//
	// This is called multiple times per step, once for each nudge
	// resolution iteration.
	ApplyNudges(ctx PairConstraintContext)
}

// PairConstraintID uniquely identifies a pair constraint that has been
// created through [PairConstraintView.Create].
//
// The zero value is not a valid ID; use [NilPairConstraintID] to represent
// the absence of a pair constraint.
type PairConstraintID struct {
	index    int32
	revision int32
}

// NilPairConstraintID is a [PairConstraintID] that is guaranteed to never
// reference a valid pair constraint.
var NilPairConstraintID = PairConstraintID{}

// PairConstraintView provides access to the pair constraints (i.e.
// constraints that act on two bodies simultaneously) that belong to a
// [Scene], as opposed to a solo constraint, which acts on a single body.
//
// A PairConstraintView is a lightweight accessor around a [Scene] and can
// be obtained through [Scene.PairConstraints].
type PairConstraintView struct {
	scene *Scene
}

// Create registers solver as a new pair constraint that acts on the
// bodies identified by primaryID and secondaryID, and returns an ID
// through which the constraint can be referenced in the future.
//
// The returned constraint is enabled by default. It is automatically
// deleted whenever either of its two target bodies is deleted.
//
// Create panics if primaryID or secondaryID does not reference a valid
// body.
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
		primaryNextIndex:   primaryBody.firstPairConstraintIndex,
		secondaryNextIndex: secondaryBody.firstPairConstraintIndex,
		isEnabled:          true,
	}
	primaryBody.firstPairConstraintIndex = index
	secondaryBody.firstPairConstraintIndex = index

	return PairConstraintID{
		index:    index,
		revision: constraint.revision,
	}
}

// Delete removes the pair constraint identified by id, unlinking it from
// both of its target bodies and releasing the underlying storage for
// reuse.
//
// Delete panics if id does not reference a valid pair constraint.
func (v PairConstraintView) Delete(id PairConstraintID) {
	constraint := v.resolve(id, true)

	// Unlink the constraint from the primary body.
	primaryBody := &v.scene.bodies[constraint.primaryBodyIndex]
	if primaryBody.firstPairConstraintIndex == id.index {
		primaryBody.firstPairConstraintIndex = constraint.primaryNextIndex
	} else {
		prevIndex := primaryBody.firstPairConstraintIndex
		for prevIndex != nilIndex {
			prevConstraint := &v.scene.pairConstraints[prevIndex]
			if prevConstraint.primaryNextIndex == id.index {
				prevConstraint.primaryNextIndex = constraint.primaryNextIndex
				break
			}
			prevIndex = prevConstraint.primaryNextIndex
		}
	}

	// Unlink the constraint from the secondary body.
	secondaryBody := &v.scene.bodies[constraint.secondaryBodyIndex]
	if secondaryBody.firstPairConstraintIndex == id.index {
		secondaryBody.firstPairConstraintIndex = constraint.secondaryNextIndex
	} else {
		prevIndex := secondaryBody.firstPairConstraintIndex
		for prevIndex != nilIndex {
			prevConstraint := &v.scene.pairConstraints[prevIndex]
			if prevConstraint.secondaryNextIndex == id.index {
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

// Handle returns a [PairConstraintHandle] that wraps id, offering a more
// convenient, object-oriented way to interact with the referenced pair
// constraint.
func (v PairConstraintView) Handle(id PairConstraintID) PairConstraintHandle {
	return PairConstraintHandle{
		view: v,
		id:   id,
	}
}

// IsValid returns whether id references a pair constraint that is still
// alive within the [Scene].
func (v PairConstraintView) IsValid(id PairConstraintID) bool {
	constraint := v.resolve(id, false)
	return constraint != nil
}

// PrimaryBodyID returns the ID of the primary body on which the pair
// constraint identified by id acts.
//
// PrimaryBodyID panics if id does not reference a valid pair constraint.
func (v PairConstraintView) PrimaryBodyID(id PairConstraintID) BodyID {
	constraint := v.resolve(id, true)
	bodyIndex := constraint.primaryBodyIndex
	body := &v.scene.bodies[bodyIndex]
	return BodyID{
		index:    bodyIndex,
		revision: body.revision,
	}
}

// SecondaryBodyID returns the ID of the secondary body on which the pair
// constraint identified by id acts.
//
// SecondaryBodyID panics if id does not reference a valid pair
// constraint.
func (v PairConstraintView) SecondaryBodyID(id PairConstraintID) BodyID {
	constraint := v.resolve(id, true)
	bodyIndex := constraint.secondaryBodyIndex
	body := &v.scene.bodies[bodyIndex]
	return BodyID{
		index:    bodyIndex,
		revision: body.revision,
	}
}

// Solver returns the [PairConstraintSolver] that implements the pair
// constraint identified by id.
//
// Solver panics if id does not reference a valid pair constraint.
func (v PairConstraintView) Solver(id PairConstraintID) PairConstraintSolver {
	constraint := v.resolve(id, true)
	return constraint.solver
}

// SetSolver changes the [PairConstraintSolver] that implements the pair
// constraint identified by id.
//
// SetSolver panics if id does not reference a valid pair constraint.
func (v PairConstraintView) SetSolver(id PairConstraintID, solver PairConstraintSolver) {
	constraint := v.resolve(id, true)
	constraint.solver = solver
}

// Enabled returns whether the pair constraint identified by id is
// currently enforced by the physics engine.
//
// Enabled panics if id does not reference a valid pair constraint.
func (v PairConstraintView) Enabled(id PairConstraintID) bool {
	constraint := v.resolve(id, true)
	return constraint.isEnabled
}

// SetEnabled changes whether the pair constraint identified by id is
// enforced by the physics engine.
//
// SetEnabled panics if id does not reference a valid pair constraint.
func (v PairConstraintView) SetEnabled(id PairConstraintID, enabled bool) {
	constraint := v.resolve(id, true)
	constraint.isEnabled = enabled
}

// idFromIndex builds the current [PairConstraintID] for the pair
// constraint stored at the given slice index.
func (v PairConstraintView) idFromIndex(index int32) PairConstraintID {
	state := &v.scene.pairConstraints[index]
	return PairConstraintID{
		index:    index,
		revision: state.revision,
	}
}

// resolve looks up the pairConstraintState referenced by id. If id is
// stale or otherwise invalid, resolve panics when required is true, or
// returns nil otherwise.
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

// PairConstraintHandle is an object-oriented alternative to
// [PairConstraintView] that is bound to a specific [PairConstraintID].
//
// It is obtained through [PairConstraintView.Handle].
type PairConstraintHandle struct {
	view PairConstraintView
	id   PairConstraintID
}

// ID returns the identifier of the pair constraint targeted by this
// handle.
func (h PairConstraintHandle) ID() PairConstraintID {
	return h.id
}

// Delete removes the pair constraint targeted by this handle.
//
// See [PairConstraintView.Delete] for further details.
func (h PairConstraintHandle) Delete() {
	h.view.Delete(h.id)
}

// IsValid returns whether this handle still references a pair constraint
// that is alive within the [Scene].
func (h PairConstraintHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

// PrimaryBodyID returns the ID of the primary body on which the targeted
// pair constraint acts.
func (h PairConstraintHandle) PrimaryBodyID() BodyID {
	return h.view.PrimaryBodyID(h.id)
}

// SecondaryBodyID returns the ID of the secondary body on which the
// targeted pair constraint acts.
func (h PairConstraintHandle) SecondaryBodyID() BodyID {
	return h.view.SecondaryBodyID(h.id)
}

// Solver returns the [PairConstraintSolver] that implements the targeted
// pair constraint.
func (h PairConstraintHandle) Solver() PairConstraintSolver {
	return h.view.Solver(h.id)
}

// SetSolver changes the [PairConstraintSolver] that implements the
// targeted pair constraint.
func (h PairConstraintHandle) SetSolver(solver PairConstraintSolver) {
	h.view.SetSolver(h.id, solver)
}

// Enabled returns whether the targeted pair constraint is currently
// enforced by the physics engine.
func (h PairConstraintHandle) Enabled() bool {
	return h.view.Enabled(h.id)
}

// SetEnabled changes whether the targeted pair constraint is enforced by
// the physics engine.
func (h PairConstraintHandle) SetEnabled(enabled bool) {
	h.view.SetEnabled(h.id, enabled)
}

// pairConstraintState holds the internal state of a single pair
// constraint, as tracked by a [Scene].
//
// Instances are threaded into two independent singly-linked lists, one
// per participating body, each rooted at that body's
// firstPairConstraintIndex. Within a given body's list, a node's link to
// the next entry is held in primaryNextIndex if that body is the node's
// primary body, or in secondaryNextIndex if it is the node's secondary
// body, since the same body can be the primary of one constraint and the
// secondary of another.
type pairConstraintState struct {
	solver             PairConstraintSolver
	revision           int32
	primaryBodyIndex   int32
	secondaryBodyIndex int32
	primaryNextIndex   int32
	secondaryNextIndex int32
	isEnabled          bool
}

// isValid returns whether this state is currently backing a live pair
// constraint, as opposed to a freed slot awaiting reuse.
func (s *pairConstraintState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
