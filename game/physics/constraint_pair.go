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

// ImpulseLambda returns the impulse magnitude (lambda) that needs to be
// applied along primaryJacobian to [PairConstraintContext.PrimaryTarget]
// and along secondaryJacobian to
// [PairConstraintContext.SecondaryTarget], in order to correct their
// combined velocity error, taking into account both restitution and
// Baumgarte positional-drift stabilization.
//
// primaryJacobian and secondaryJacobian must describe the same scalar
// constraint from each body's own point of view - typically with
// secondaryJacobian's slopes being the negation of primaryJacobian's -
// so that applying the same lambda to both targets, as
// [PairConstraintContext.ImpulseSolution] does, is consistent with
// Newton's third law.
//
// drift is the current positional error of the constraint (e.g. a
// penetration depth), and restitutionCoef is the coefficient of
// restitution to apply, in the [0.0, 1.0] range.
//
// The result is intended to be passed to [Jacobian.Impulse] for each of
// primaryJacobian and secondaryJacobian.
func (c PairConstraintContext) ImpulseLambda(primaryJacobian, secondaryJacobian Jacobian, drift, restitutionCoef float64) float64 {
	invEffMass := primaryJacobian.InverseEffectiveMass(c.PrimaryTarget) + secondaryJacobian.InverseEffectiveMass(c.SecondaryTarget)
	if invEffMass < Epsilon {
		return 0.0
	}
	effVelocity := primaryJacobian.EffectiveVelocity(c.PrimaryTarget) + secondaryJacobian.EffectiveVelocity(c.SecondaryTarget)
	restitution := 1 + restitutionCoef*RestitutionClamp(effVelocity)
	driftBias := c.ImpulseBeta * drift / c.DeltaSeconds
	return -(restitution*effVelocity - driftBias) / invEffMass
}

// ImpulseLambdaComponents behaves like
// [PairConstraintContext.ImpulseLambda] but returns its two additive
// components separately instead of their sum: bounce is the
// restitution-only component, derived purely from the two jacobians'
// current combined velocity error, while baumgarte is the
// positional-drift-correction component, derived from drift.
//
// bounce + baumgarte is equal to the value that
// [PairConstraintContext.ImpulseLambda] would return for the same
// arguments. Callers that need to reason about the velocity-only portion
// in isolation - for example to detect separating contacts, or to derive
// a friction bound from it - should use this method instead.
func (c PairConstraintContext) ImpulseLambdaComponents(primaryJacobian, secondaryJacobian Jacobian, drift, restitutionCoef float64) (bounce, baumgarte float64) {
	invEffMass := primaryJacobian.InverseEffectiveMass(c.PrimaryTarget) + secondaryJacobian.InverseEffectiveMass(c.SecondaryTarget)
	if invEffMass < Epsilon {
		return 0.0, 0.0
	}
	effVelocity := primaryJacobian.EffectiveVelocity(c.PrimaryTarget) + secondaryJacobian.EffectiveVelocity(c.SecondaryTarget)
	restitution := 1 + restitutionCoef*RestitutionClamp(effVelocity)
	driftBias := c.ImpulseBeta * drift / c.DeltaSeconds
	return -restitution * effVelocity / invEffMass, driftBias / invEffMass
}

// ImpulseSolution returns the [Impulse] that needs to be applied to
// [PairConstraintContext.PrimaryTarget] along primaryJacobian, and the
// [Impulse] that needs to be applied to
// [PairConstraintContext.SecondaryTarget] along secondaryJacobian, in
// order to correct their combined velocity error, combining restitution
// and Baumgarte positional-drift stabilization.
//
// See [PairConstraintContext.ImpulseLambda] for details on
// primaryJacobian, secondaryJacobian, drift and restitutionCoef.
func (c PairConstraintContext) ImpulseSolution(primaryJacobian, secondaryJacobian Jacobian, drift, restitutionCoef float64) (Impulse, Impulse) {
	lambda := c.ImpulseLambda(primaryJacobian, secondaryJacobian, drift, restitutionCoef)
	return primaryJacobian.Impulse(lambda), secondaryJacobian.Impulse(lambda)
}

// NudgeLambda returns the nudge magnitude (lambda) that needs to be
// applied along primaryJacobian to [PairConstraintContext.PrimaryTarget]
// and along secondaryJacobian to [PairConstraintContext.SecondaryTarget],
// in order to correct drift, the current combined positional error of
// the constraint (e.g. a penetration depth), through Baumgarte
// positional-drift stabilization.
//
// See [PairConstraintContext.ImpulseLambda] for the expected relationship
// between primaryJacobian and secondaryJacobian.
//
// The result is intended to be passed to [Jacobian.Nudge] for each of
// primaryJacobian and secondaryJacobian.
func (c PairConstraintContext) NudgeLambda(primaryJacobian, secondaryJacobian Jacobian, drift float64) float64 {
	invEffMass := primaryJacobian.InverseEffectiveMass(c.PrimaryTarget) + secondaryJacobian.InverseEffectiveMass(c.SecondaryTarget)
	if invEffMass < Epsilon {
		return 0.0
	}
	return c.NudgeBeta * drift / invEffMass
}

// NudgeSolution returns the [Nudge] that needs to be applied to
// [PairConstraintContext.PrimaryTarget] along primaryJacobian, and the
// [Nudge] that needs to be applied to
// [PairConstraintContext.SecondaryTarget] along secondaryJacobian, in
// order to correct drift, the current combined positional error of the
// constraint (e.g. a penetration depth).
func (c PairConstraintContext) NudgeSolution(primaryJacobian, secondaryJacobian Jacobian, drift float64) (Nudge, Nudge) {
	lambda := c.NudgeLambda(primaryJacobian, secondaryJacobian, drift)
	return primaryJacobian.Nudge(lambda), secondaryJacobian.Nudge(lambda)
}

// PairConstraintSolver implements the mathematical logic that enforces a
// constraint acting on two bodies simultaneously.
//
// Instances are registered with a [Scene] through [PairConstraintView.Create]
// and are subsequently driven by the physics engine through the methods
// below during each simulation step.
type PairConstraintSolver interface {

	// Reset clears any internal cache state held by the solver and
	// recomputes any data (e.g. Jacobians) that is derived from the
	// current position and orientation of the target bodies.
	//
	// This is called once before the first
	// [PairConstraintSolver.ApplyImpulses] iteration of a step, since the
	// target bodies' positions and orientations remain unchanged
	// throughout that loop.
	//
	// This is also called before every single
	// [PairConstraintSolver.ApplyNudges] invocation, since nudges
	// reposition the target bodies and would otherwise leave the solver's
	// cached, position-derived data stale for subsequent iterations.
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
// body, or if they both reference the same body, since a pair constraint
// cannot act on a single body twice.
func (v PairConstraintView) Create(primaryID, secondaryID BodyID, solver PairConstraintSolver) PairConstraintID {
	if primaryID == secondaryID {
		panic("pair constraint cannot be created between a body and itself")
	}

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

// CreateHandle behaves like [PairConstraintView.Create] but wraps the
// resulting ID in a [PairConstraintHandle], as returned by
// [PairConstraintView.Handle], for callers that want to keep acting on
// the new pair constraint without holding onto its ID separately.
//
// CreateHandle panics if primaryID or secondaryID does not reference a
// valid body, or if they both reference the same body, since a pair
// constraint cannot act on a single body twice.
func (v PairConstraintView) CreateHandle(primaryID, secondaryID BodyID, solver PairConstraintSolver) PairConstraintHandle {
	return v.Handle(v.Create(primaryID, secondaryID, solver))
}

// Delete removes the pair constraint identified by id, unlinking it from
// both of its target bodies and releasing the underlying storage for
// reuse.
//
// Delete panics if id does not reference a valid pair constraint.
func (v PairConstraintView) Delete(id PairConstraintID) {
	constraint := v.resolve(id, true)

	// Unlink the constraint from the primary body.
	primaryBodyIndex := constraint.primaryBodyIndex
	primaryBody := &v.scene.bodies[primaryBodyIndex]
	if primaryBody.firstPairConstraintIndex == id.index {
		primaryBody.firstPairConstraintIndex = constraint.nextIndexForBody(primaryBodyIndex)
	} else {
		prevIndex := primaryBody.firstPairConstraintIndex
		for prevIndex != nilIndex {
			prevConstraint := &v.scene.pairConstraints[prevIndex]
			switch {
			case prevConstraint.primaryBodyIndex == primaryBodyIndex: // body follows primary chain
				if prevConstraint.primaryNextIndex == id.index {
					prevConstraint.primaryNextIndex = constraint.nextIndexForBody(primaryBodyIndex)
					prevIndex = nilIndex // break the loop
				} else {
					prevIndex = prevConstraint.primaryNextIndex
				}
			case prevConstraint.secondaryBodyIndex == primaryBodyIndex: // body follows secondary chain
				if prevConstraint.secondaryNextIndex == id.index {
					prevConstraint.secondaryNextIndex = constraint.nextIndexForBody(primaryBodyIndex)
					prevIndex = nilIndex // break the loop
				} else {
					prevIndex = prevConstraint.secondaryNextIndex
				}
			default:
				panic("body index does not match either primary or secondary body")
			}
		}
	}

	// Unlink the constraint from the secondary body.
	secondaryBodyIndex := constraint.secondaryBodyIndex
	secondaryBody := &v.scene.bodies[secondaryBodyIndex]
	if secondaryBody.firstPairConstraintIndex == id.index {
		secondaryBody.firstPairConstraintIndex = constraint.nextIndexForBody(secondaryBodyIndex)
	} else {
		prevIndex := secondaryBody.firstPairConstraintIndex
		for prevIndex != nilIndex {
			prevConstraint := &v.scene.pairConstraints[prevIndex]
			switch {
			case prevConstraint.primaryBodyIndex == secondaryBodyIndex: // body follows primary chain
				if prevConstraint.primaryNextIndex == id.index {
					prevConstraint.primaryNextIndex = constraint.nextIndexForBody(secondaryBodyIndex)
					prevIndex = nilIndex // break the loop
				} else {
					prevIndex = prevConstraint.primaryNextIndex
				}
			case prevConstraint.secondaryBodyIndex == secondaryBodyIndex: // body follows secondary chain
				if prevConstraint.secondaryNextIndex == id.index {
					prevConstraint.secondaryNextIndex = constraint.nextIndexForBody(secondaryBodyIndex)
					prevIndex = nilIndex // break the loop
				} else {
					prevIndex = prevConstraint.secondaryNextIndex
				}
			default:
				panic("body index does not match either primary or secondary body")
			}
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

// Each calls cb once for every pair constraint that is currently alive
// within this Scene, in unspecified order.
func (v PairConstraintView) Each(cb func(id PairConstraintID)) {
	v.scene.eachPairConstraint(func(index int, constraint *pairConstraintState) {
		cb(PairConstraintID{
			index:    int32(index),
			revision: constraint.revision,
		})
	})
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
//
// Wrapping an ID this way needs no allocation of its own, but unlike a
// plain [PairConstraintID] (which holds no pointers), a Handle keeps an
// internal reference to the owning [Scene]. This keeps that Scene
// reachable for as long as the handle is retained, and adds a pointer
// that the garbage collector must trace wherever the handle is stored.
// Prefer storing IDs over handles in long-lived collections unless the
// convenience is worth that cost.
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

// nextIndexForBody returns the index of the next entry in bodyIndex's
// singly-linked list of pair constraints, following primaryNextIndex or
// secondaryNextIndex depending on whether bodyIndex is this state's
// primary or secondary body.
//
// nextIndexForBody panics if bodyIndex is neither the primary nor the
// secondary body of this state.
func (s *pairConstraintState) nextIndexForBody(bodyIndex int32) int32 {
	switch bodyIndex {
	case s.primaryBodyIndex:
		return s.primaryNextIndex
	case s.secondaryBodyIndex:
		return s.secondaryNextIndex
	default:
		panic("body index does not match either primary or secondary body")
	}
}

// isValid returns whether this state is currently backing a live pair
// constraint, as opposed to a freed slot awaiting reuse.
func (s *pairConstraintState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
