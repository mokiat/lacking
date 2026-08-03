package physics

// SoloConstraintContext contains the information that a [SoloConstraintSolver]
// needs in order to process the single body it acts upon during a physics
// simulation step.
type SoloConstraintContext struct {

	// DeltaSeconds is the amount of time, in seconds, covered by the
	// current physics simulation step.
	DeltaSeconds float64

	// ImpulseBeta is the Baumgarte stabilization factor to be used when
	// correcting positional drift through impulses.
	ImpulseBeta float64

	// NudgeBeta is the Baumgarte stabilization factor to be used when
	// correcting positional drift through nudges.
	NudgeBeta float64

	// Target is the [ConstraintTarget] for the body that is constrained,
	// through which the solver reads its motion state and applies
	// corrective impulses and nudges to it.
	Target ConstraintTarget
}

func (c SoloConstraintContext) ImpulseLambda(jacobian Jacobian, drift, restitutionCoef float64) float64 {
	effMass := jacobian.InverseEffectiveMass(c.Target)
	if effMass < Epsilon {
		return 0.0
	}
	effVelocity := jacobian.EffectiveVelocity(c.Target)
	restitution := 1 + restitutionCoef*RestitutionClamp(effVelocity)
	baumgarte := c.ImpulseBeta * drift / c.DeltaSeconds
	return -(restitution*effVelocity - baumgarte) / effMass
}

func (c SoloConstraintContext) ImpulseLambdaSplit(jacobian Jacobian, drift, restitutionCoef float64) (float64, float64) {
	effMass := jacobian.InverseEffectiveMass(c.Target)
	if effMass < Epsilon {
		return 0.0, 0.0
	}
	effVelocity := jacobian.EffectiveVelocity(c.Target)
	restitution := 1 + restitutionCoef*RestitutionClamp(effVelocity)
	baumgarte := c.ImpulseBeta * drift / c.DeltaSeconds
	return -restitution * effVelocity / effMass, baumgarte / effMass
}

func (c SoloConstraintContext) ImpulseSolution(jacobian Jacobian, drift, restitutionCoef float64) Impulse {
	lambda := c.ImpulseLambda(jacobian, drift, restitutionCoef)
	return jacobian.Impulse(lambda)
}

func (c SoloConstraintContext) NudgeLambda(jacobian Jacobian, drift float64) float64 {
	effMass := jacobian.InverseEffectiveMass(c.Target)
	if effMass < Epsilon {
		return 0.0
	}
	return c.NudgeBeta * drift / effMass
}

func (c SoloConstraintContext) NudgeSolution(jacobian Jacobian, drift float64) Nudge {
	lambda := c.NudgeLambda(jacobian, drift)
	return jacobian.Nudge(lambda)
}

// SoloConstraintSolver implements the mathematical logic that enforces a
// constraint acting on a single body.
//
// Instances are registered with a [Scene] through [SoloConstraintView.Create]
// and are subsequently driven by the physics engine through the methods
// below during each simulation step.
type SoloConstraintSolver interface {

	// Reset clears any internal cache state held by the solver and
	// recomputes any data (e.g. Jacobians) that is derived from the
	// current position and orientation of the target body.
	//
	// This is called once before the first
	// [SoloConstraintSolver.ApplyImpulses] iteration of a step, since the
	// target body's position and orientation remain unchanged throughout
	// that loop.
	//
	// This is also called before every single
	// [SoloConstraintSolver.ApplyNudges] invocation, since nudges
	// reposition the target body and would otherwise leave the solver's
	// cached, position-derived data stale for subsequent iterations.
	Reset(ctx SoloConstraintContext)

	// ApplyImpulses is called by the physics engine to instruct the solver
	// to apply the necessary impulses to its target body, in order to
	// correct its velocity so that the constraint is satisfied.
	//
	// This is called multiple times per step, once for each impulse
	// resolution iteration.
	ApplyImpulses(ctx SoloConstraintContext)

	// ApplyNudges is called by the physics engine to instruct the solver
	// to apply the necessary nudges to its target body, in order to
	// correct its position so that the constraint is satisfied.
	//
	// This is called multiple times per step, once for each nudge
	// resolution iteration.
	ApplyNudges(ctx SoloConstraintContext)
}

// SoloConstraintID uniquely identifies a solo constraint that has been
// created through [SoloConstraintView.Create].
//
// The zero value is not a valid ID; use [NilSoloConstraintID] to represent
// the absence of a solo constraint.
type SoloConstraintID struct {
	index    int32
	revision int32
}

// NilSoloConstraintID is a [SoloConstraintID] that is guaranteed to never
// reference a valid solo constraint.
var NilSoloConstraintID = SoloConstraintID{}

// SoloConstraintView provides access to the solo constraints (i.e.
// constraints that act on a single body) that belong to a [Scene], as
// opposed to a pair constraint, which acts on two bodies simultaneously.
//
// A SoloConstraintView is a lightweight accessor around a [Scene] and can
// be obtained through [Scene.SoloConstraints].
type SoloConstraintView struct {
	scene *Scene
}

// Create registers solver as a new solo constraint that acts on the body
// identified by bodyID, and returns an ID through which the constraint can
// be referenced in the future.
//
// The returned constraint is enabled by default. It is automatically
// deleted whenever the target body is deleted.
//
// Create panics if bodyID does not reference a valid body.
func (v SoloConstraintView) Create(bodyID BodyID, solver SoloConstraintSolver) SoloConstraintID {
	bodyView := v.scene.Bodies()
	body := bodyView.resolve(bodyID, true)

	index, constraint := v.scene.allocateSoloConstraint()

	*constraint = soloConstraintState{
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

// Delete removes the solo constraint identified by id, unlinking it from
// its target body and releasing the underlying storage for reuse.
//
// Delete panics if id does not reference a valid solo constraint.
func (v SoloConstraintView) Delete(id SoloConstraintID) {
	constraint := v.resolve(id, true)

	// Unlink the constraint from its body's singly-linked list of solo
	// constraints, which may require patching up a preceding sibling.
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

	*constraint = soloConstraintState{
		solver:    nil,                     // allow the solver to be garbage collected
		revision:  constraint.revision + 1, // progress revision to invalid (even) value
		bodyIndex: nilIndex,
		nextIndex: nilIndex,
		isEnabled: false,
	}

	v.scene.releaseSoloConstraint(id.index)
}

// Each calls cb once for every solo constraint that is currently alive
// within this Scene, in unspecified order.
func (v SoloConstraintView) Each(cb func(id SoloConstraintID)) {
	v.scene.eachSoloConstraint(func(index int, constraint *soloConstraintState) {
		cb(SoloConstraintID{
			index:    int32(index),
			revision: constraint.revision,
		})
	})
}

// Handle returns a [SoloConstraintHandle] that wraps id, offering a more
// convenient, object-oriented way to interact with the referenced solo
// constraint.
func (v SoloConstraintView) Handle(id SoloConstraintID) SoloConstraintHandle {
	return SoloConstraintHandle{
		view: v,
		id:   id,
	}
}

// IsValid returns whether id references a solo constraint that is still
// alive within the [Scene].
func (v SoloConstraintView) IsValid(id SoloConstraintID) bool {
	constraint := v.resolve(id, false)
	return constraint != nil
}

// BodyID returns the ID of the body on which the solo constraint
// identified by id acts.
//
// BodyID panics if id does not reference a valid solo constraint.
func (v SoloConstraintView) BodyID(id SoloConstraintID) BodyID {
	constraint := v.resolve(id, true)
	bodyIndex := constraint.bodyIndex
	body := &v.scene.bodies[bodyIndex]
	return BodyID{
		index:    bodyIndex,
		revision: body.revision,
	}
}

// Solver returns the [SoloConstraintSolver] that implements the solo
// constraint identified by id.
//
// Solver panics if id does not reference a valid solo constraint.
func (v SoloConstraintView) Solver(id SoloConstraintID) SoloConstraintSolver {
	constraint := v.resolve(id, true)
	return constraint.solver
}

// SetSolver changes the [SoloConstraintSolver] that implements the solo
// constraint identified by id.
//
// SetSolver panics if id does not reference a valid solo constraint.
func (v SoloConstraintView) SetSolver(id SoloConstraintID, solver SoloConstraintSolver) {
	constraint := v.resolve(id, true)
	constraint.solver = solver
}

// Enabled returns whether the solo constraint identified by id is
// currently enforced by the physics engine.
//
// Enabled panics if id does not reference a valid solo constraint.
func (v SoloConstraintView) Enabled(id SoloConstraintID) bool {
	constraint := v.resolve(id, true)
	return constraint.isEnabled
}

// SetEnabled changes whether the solo constraint identified by id is
// enforced by the physics engine.
//
// SetEnabled panics if id does not reference a valid solo constraint.
func (v SoloConstraintView) SetEnabled(id SoloConstraintID, enabled bool) {
	constraint := v.resolve(id, true)
	constraint.isEnabled = enabled
}

// idFromIndex builds the current [SoloConstraintID] for the solo
// constraint stored at the given slice index.
func (v SoloConstraintView) idFromIndex(index int32) SoloConstraintID {
	constraint := &v.scene.soloConstraints[index]
	return SoloConstraintID{
		index:    index,
		revision: constraint.revision,
	}
}

// resolve looks up the soloConstraintState referenced by id. If id is
// stale or otherwise invalid, resolve panics when required is true, or
// returns nil otherwise.
func (v SoloConstraintView) resolve(id SoloConstraintID, required bool) *soloConstraintState {
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

// SoloConstraintHandle is an object-oriented alternative to
// [SoloConstraintView] that is bound to a specific [SoloConstraintID].
//
// It is obtained through [SoloConstraintView.Handle].
type SoloConstraintHandle struct {
	view SoloConstraintView
	id   SoloConstraintID
}

// ID returns the identifier of the solo constraint targeted by this
// handle.
func (h SoloConstraintHandle) ID() SoloConstraintID {
	return h.id
}

// Delete removes the solo constraint targeted by this handle.
//
// See [SoloConstraintView.Delete] for further details.
func (h SoloConstraintHandle) Delete() {
	h.view.Delete(h.id)
}

// IsValid returns whether this handle still references a solo constraint
// that is alive within the [Scene].
func (h SoloConstraintHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

// BodyID returns the ID of the body on which the targeted solo constraint
// acts.
func (h SoloConstraintHandle) BodyID() BodyID {
	return h.view.BodyID(h.id)
}

// Solver returns the [SoloConstraintSolver] that implements the targeted
// solo constraint.
func (h SoloConstraintHandle) Solver() SoloConstraintSolver {
	return h.view.Solver(h.id)
}

// SetSolver changes the [SoloConstraintSolver] that implements the
// targeted solo constraint.
func (h SoloConstraintHandle) SetSolver(solver SoloConstraintSolver) {
	h.view.SetSolver(h.id, solver)
}

// Enabled returns whether the targeted solo constraint is currently
// enforced by the physics engine.
func (h SoloConstraintHandle) Enabled() bool {
	return h.view.Enabled(h.id)
}

// SetEnabled changes whether the targeted solo constraint is enforced by
// the physics engine.
func (h SoloConstraintHandle) SetEnabled(enabled bool) {
	h.view.SetEnabled(h.id, enabled)
}

// soloConstraintState holds the internal state of a single solo constraint, as
// tracked by a [Scene].
//
// Instances form a singly-linked list (through nextIndex) per body, rooted
// at the owning body's firstSoloConstraintIndex, so that all the solo
// constraints acting on a given body can be enumerated or deleted
// together.
type soloConstraintState struct {
	solver    SoloConstraintSolver
	revision  int32
	bodyIndex int32
	nextIndex int32
	isEnabled bool
}

// isValid returns whether this state is currently backing a live solo
// constraint, as opposed to a freed slot awaiting reuse.
func (s *soloConstraintState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
