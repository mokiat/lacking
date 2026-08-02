package physics

// BodyAcceleratorID uniquely identifies a body accelerator that was created
// through [BodyAcceleratorView.Create].
//
// The zero value, also available as [NilBodyAcceleratorID], does not
// reference a valid body accelerator.
type BodyAcceleratorID struct {
	index    int32
	revision int32
}

// NilBodyAcceleratorID is a [BodyAcceleratorID] that never references a
// valid body accelerator.
var NilBodyAcceleratorID = BodyAcceleratorID{}

// BodyAcceleratorView provides access to the body accelerators that belong
// to a [Scene].
//
// A body accelerator evaluates its [AccelerationSolver] once for a single,
// specific body, on every simulation step. It is intended for effects that
// are tied to a particular body, as opposed to a global accelerator, which
// affects every body in the scene.
type BodyAcceleratorView struct {
	scene *Scene
}

// Create allocates a new body accelerator that uses the specified solver to
// act on the body identified by bodyID, and returns its ID.
//
// The accelerator is enabled by default. It is automatically deleted
// whenever the target body is deleted.
//
// Create panics if bodyID does not reference a valid body.
func (v BodyAcceleratorView) Create(bodyID BodyID, solver AccelerationSolver) BodyAcceleratorID {
	bodyView := v.scene.Bodies()
	body := bodyView.resolve(bodyID, true)

	index, accelerator := v.scene.allocateBodyAccelerator()

	*accelerator = bodyAcceleratorState{
		solver:    solver,
		revision:  accelerator.revision + 1, // progress revision to valid (odd) value
		bodyIndex: bodyID.index,
		nextIndex: body.firstBodyAcceleratorIndex,
		isEnabled: true,
	}
	body.firstBodyAcceleratorIndex = index

	return BodyAcceleratorID{
		index:    index,
		revision: accelerator.revision,
	}
}

// Delete removes the body accelerator with the specified ID, unlinking it
// from its target body and releasing the underlying storage for reuse.
//
// It panics if the ID does not reference a valid body accelerator, be it
// because it was never created, has already been deleted, or belongs to a
// different [Scene]. Use [BodyAcceleratorView.IsValid] first if the ID's
// validity is not otherwise guaranteed.
func (v BodyAcceleratorView) Delete(id BodyAcceleratorID) {
	accelerator := v.resolve(id, true)

	// Unlink the accelerator from its body's singly-linked list of body
	// accelerators, which may require patching up a preceding sibling.
	body := &v.scene.bodies[accelerator.bodyIndex]
	if body.firstBodyAcceleratorIndex == id.index {
		body.firstBodyAcceleratorIndex = accelerator.nextIndex
	} else {
		prevIndex := body.firstBodyAcceleratorIndex
		for prevIndex != nilIndex {
			prev := &v.scene.bodyAccelerators[prevIndex]
			if prev.nextIndex == id.index {
				prev.nextIndex = accelerator.nextIndex
				break
			}
			prevIndex = prev.nextIndex
		}
	}

	*accelerator = bodyAcceleratorState{
		solver:    nil,                      // allow the solver to be garbage collected
		revision:  accelerator.revision + 1, // progress revision to invalid (even) value
		bodyIndex: nilIndex,
		nextIndex: nilIndex,
		isEnabled: false,
	}

	v.scene.releaseBodyAccelerator(id.index)
}

func (v BodyAcceleratorView) Each(cb func(id BodyAcceleratorID)) {
	v.scene.eachBodyAccelerator(func(index int, accelerator *bodyAcceleratorState) {
		cb(BodyAcceleratorID{
			index:    int32(index),
			revision: accelerator.revision,
		})
	})
}

// Handle returns a [BodyAcceleratorHandle] that wraps the specified ID, as
// a more convenient means of repeatedly accessing the same body accelerator
// without having to pass its ID to this view on every call.
func (v BodyAcceleratorView) Handle(id BodyAcceleratorID) BodyAcceleratorHandle {
	return BodyAcceleratorHandle{
		view: v,
		id:   id,
	}
}

// IsValid returns whether the specified ID references a body accelerator
// that has not been deleted.
func (v BodyAcceleratorView) IsValid(id BodyAcceleratorID) bool {
	accelerator := v.resolve(id, false)
	return accelerator != nil
}

// BodyID returns the ID of the body on which the specified body
// accelerator acts.
//
// It panics if the ID does not reference a valid body accelerator.
func (v BodyAcceleratorView) BodyID(id BodyAcceleratorID) BodyID {
	accelerator := v.resolve(id, true)
	bodyIndex := accelerator.bodyIndex
	body := &v.scene.bodies[bodyIndex]
	return BodyID{
		index:    bodyIndex,
		revision: body.revision,
	}
}

// Solver returns the acceleration solver used by the specified body
// accelerator.
//
// It panics if the ID does not reference a valid body accelerator.
func (v BodyAcceleratorView) Solver(id BodyAcceleratorID) AccelerationSolver {
	accelerator := v.resolve(id, true)
	return accelerator.solver
}

// SetSolver changes the acceleration solver used by the specified body
// accelerator.
//
// It panics if the ID does not reference a valid body accelerator.
func (v BodyAcceleratorView) SetSolver(id BodyAcceleratorID, solver AccelerationSolver) {
	accelerator := v.resolve(id, true)
	accelerator.solver = solver
}

// Enabled returns whether the specified body accelerator is evaluated
// during the simulation. A body accelerator is enabled by default.
//
// It panics if the ID does not reference a valid body accelerator.
func (v BodyAcceleratorView) Enabled(id BodyAcceleratorID) bool {
	accelerator := v.resolve(id, true)
	return accelerator.isEnabled
}

// SetEnabled changes whether the specified body accelerator is evaluated
// during the simulation.
//
// It panics if the ID does not reference a valid body accelerator.
func (v BodyAcceleratorView) SetEnabled(id BodyAcceleratorID, enabled bool) {
	accelerator := v.resolve(id, true)
	accelerator.isEnabled = enabled
}

// idFromIndex builds the current [BodyAcceleratorID] for the body
// accelerator stored at the given slice index.
func (v BodyAcceleratorView) idFromIndex(index int32) BodyAcceleratorID {
	accelerator := &v.scene.bodyAccelerators[index]
	return BodyAcceleratorID{
		index:    index,
		revision: accelerator.revision,
	}
}

// resolve looks up the bodyAcceleratorState referenced by id. If id is
// stale or otherwise invalid, resolve panics when required is true, or
// returns nil otherwise.
func (v BodyAcceleratorView) resolve(id BodyAcceleratorID, required bool) *bodyAcceleratorState {
	if id.revision == 0 {
		if required {
			panic("invalid body accelerator ID")
		}
		return nil
	}
	accelerator := &v.scene.bodyAccelerators[id.index]
	if accelerator.revision != id.revision {
		if required {
			panic("invalid body accelerator ID")
		}
		return nil
	}
	return accelerator
}

// BodyAcceleratorHandle is a convenience wrapper that binds together a
// [BodyAcceleratorID] and the [BodyAcceleratorView] needed to resolve it,
// so that callers that repeatedly act on the same body accelerator do not
// have to keep passing its ID around.
//
// It is created through [BodyAcceleratorView.Handle].
type BodyAcceleratorHandle struct {
	view BodyAcceleratorView
	id   BodyAcceleratorID
}

// ID returns the [BodyAcceleratorID] wrapped by this handle.
func (h BodyAcceleratorHandle) ID() BodyAcceleratorID {
	return h.id
}

// Delete removes the wrapped body accelerator.
//
// It panics if the handle does not reference a valid body accelerator.
func (h BodyAcceleratorHandle) Delete() {
	h.view.Delete(h.id)
}

// IsValid returns whether the wrapped body accelerator has not been
// deleted.
func (h BodyAcceleratorHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

// BodyID returns the ID of the body on which the wrapped body accelerator
// acts.
func (h BodyAcceleratorHandle) BodyID() BodyID {
	return h.view.BodyID(h.id)
}

// Solver returns the acceleration solver used by the wrapped body
// accelerator.
//
// It panics if the handle does not reference a valid body accelerator.
func (h BodyAcceleratorHandle) Solver() AccelerationSolver {
	return h.view.Solver(h.id)
}

// SetSolver changes the acceleration solver used by the wrapped body
// accelerator.
//
// It panics if the handle does not reference a valid body accelerator.
func (h BodyAcceleratorHandle) SetSolver(solver AccelerationSolver) {
	h.view.SetSolver(h.id, solver)
}

// Enabled returns whether the wrapped body accelerator is evaluated during
// the simulation. A body accelerator is enabled by default.
//
// It panics if the handle does not reference a valid body accelerator.
func (h BodyAcceleratorHandle) Enabled() bool {
	return h.view.Enabled(h.id)
}

// SetEnabled changes whether the wrapped body accelerator is evaluated
// during the simulation.
//
// It panics if the handle does not reference a valid body accelerator.
func (h BodyAcceleratorHandle) SetEnabled(enabled bool) {
	h.view.SetEnabled(h.id, enabled)
}

// bodyAcceleratorState holds the internal state of a single body
// accelerator, as tracked by a [Scene].
//
// Instances form a singly-linked list (through nextIndex) per body, rooted
// at the owning body's firstBodyAcceleratorIndex, so that all the body
// accelerators acting on a given body can be enumerated or deleted
// together.
type bodyAcceleratorState struct {
	solver    AccelerationSolver
	revision  int32
	bodyIndex int32
	nextIndex int32
	isEnabled bool
}

// isValid returns whether this state is currently backing a live body
// accelerator, as opposed to a freed slot awaiting reuse.
func (s *bodyAcceleratorState) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
