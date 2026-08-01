package physics

// GlobalAcceleratorID uniquely identifies a global accelerator that was
// created through [GlobalAcceleratorView.Create].
//
// The zero value, also available as [NilGlobalAcceleratorID], does not
// reference a valid global accelerator.
type GlobalAcceleratorID struct {
	index    int32
	revision int32
}

// NilGlobalAcceleratorID is a [GlobalAcceleratorID] that never references a
// valid global accelerator.
var NilGlobalAcceleratorID = GlobalAcceleratorID{}

// GlobalAcceleratorView provides access to the global accelerators that
// belong to a [Scene].
//
// A global accelerator evaluates its [AccelerationSolver] once for every
// body in the scene, on every simulation step, irrespective of the body's
// position. It is intended for scene-wide effects, such as gravity or wind,
// as opposed to a body accelerator, which affects a single, specific body.
type GlobalAcceleratorView struct {
	scene *Scene
}

// Create allocates a new global accelerator that uses the specified solver
// and returns its ID.
//
// The accelerator is enabled by default.
func (v GlobalAcceleratorView) Create(solver AccelerationSolver) GlobalAcceleratorID {
	index := v.scene.allocateGlobalAccelerator()

	accelerator := &v.scene.globalAccelerators[index]
	accelerator.solver = solver
	accelerator.revision++ // progress revision to valid (odd) value
	accelerator.enabled = true

	return GlobalAcceleratorID{
		index:    index,
		revision: accelerator.revision,
	}
}

// Delete removes the global accelerator with the specified ID.
//
// It panics if the ID does not reference a valid global accelerator, be it
// because it was never created, has already been deleted, or belongs to a
// different [Scene]. Use [GlobalAcceleratorView.IsValid] first if the ID's
// validity is not otherwise guaranteed.
func (v GlobalAcceleratorView) Delete(id GlobalAcceleratorID) {
	accelerator := v.resolve(id, true)
	accelerator.solver = nil // allow the solver to be garbage collected
	accelerator.revision++   // progress revision to invalid (even) value

	v.scene.releaseGlobalAccelerator(id.index)
}

// Handle returns a [GlobalAcceleratorHandle] that wraps the specified ID,
// as a more convenient means of repeatedly accessing the same global
// accelerator without having to pass its ID to this view on every call.
func (v GlobalAcceleratorView) Handle(id GlobalAcceleratorID) GlobalAcceleratorHandle {
	return GlobalAcceleratorHandle{
		view: v,
		id:   id,
	}
}

// IsValid returns whether the specified ID references a global accelerator
// that has not been deleted.
func (v GlobalAcceleratorView) IsValid(id GlobalAcceleratorID) bool {
	accelerator := v.resolve(id, false)
	return accelerator != nil
}

// Solver returns the acceleration solver used by the specified global
// accelerator.
//
// It panics if the ID does not reference a valid global accelerator.
func (v GlobalAcceleratorView) Solver(id GlobalAcceleratorID) AccelerationSolver {
	accelerator := v.resolve(id, true)
	return accelerator.solver
}

// SetSolver changes the acceleration solver used by the specified global
// accelerator.
//
// It panics if the ID does not reference a valid global accelerator.
func (v GlobalAcceleratorView) SetSolver(id GlobalAcceleratorID, solver AccelerationSolver) {
	accelerator := v.resolve(id, true)
	accelerator.solver = solver
}

// Enabled returns whether the specified global accelerator is evaluated
// during the simulation. A global accelerator is enabled by default.
//
// It panics if the ID does not reference a valid global accelerator.
func (v GlobalAcceleratorView) Enabled(id GlobalAcceleratorID) bool {
	accelerator := v.resolve(id, true)
	return accelerator.enabled
}

// SetEnabled changes whether the specified global accelerator is evaluated
// during the simulation.
//
// It panics if the ID does not reference a valid global accelerator.
func (v GlobalAcceleratorView) SetEnabled(id GlobalAcceleratorID, enabled bool) {
	accelerator := v.resolve(id, true)
	accelerator.enabled = enabled
}

func (v GlobalAcceleratorView) resolve(id GlobalAcceleratorID, required bool) *globalAccelerator {
	if id.revision == 0 {
		if required {
			panic("invalid global accelerator ID")
		}
		return nil
	}
	accelerator := &v.scene.globalAccelerators[id.index]
	if accelerator.revision != id.revision {
		if required {
			panic("invalid global accelerator ID")
		}
		return nil
	}
	return accelerator
}

// GlobalAcceleratorHandle is a convenience wrapper that binds together a
// [GlobalAcceleratorID] and the [GlobalAcceleratorView] needed to resolve
// it, so that callers that repeatedly act on the same global accelerator
// do not have to keep passing its ID around.
//
// It is created through [GlobalAcceleratorView.Handle].
type GlobalAcceleratorHandle struct {
	view GlobalAcceleratorView
	id   GlobalAcceleratorID
}

// ID returns the [GlobalAcceleratorID] wrapped by this handle.
func (h GlobalAcceleratorHandle) ID() GlobalAcceleratorID {
	return h.id
}

// Delete removes the wrapped global accelerator.
//
// It panics if the handle does not reference a valid global accelerator.
func (h GlobalAcceleratorHandle) Delete() {
	h.view.Delete(h.id)
}

// IsValid returns whether the wrapped global accelerator has not been
// deleted.
func (h GlobalAcceleratorHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

// Solver returns the acceleration solver used by the wrapped global
// accelerator.
//
// It panics if the handle does not reference a valid global accelerator.
func (h GlobalAcceleratorHandle) Solver() AccelerationSolver {
	return h.view.Solver(h.id)
}

// SetSolver changes the acceleration solver used by the wrapped global
// accelerator.
//
// It panics if the handle does not reference a valid global accelerator.
func (h GlobalAcceleratorHandle) SetSolver(solver AccelerationSolver) {
	h.view.SetSolver(h.id, solver)
}

// Enabled returns whether the wrapped global accelerator is evaluated
// during the simulation. A global accelerator is enabled by default.
//
// It panics if the handle does not reference a valid global accelerator.
func (h GlobalAcceleratorHandle) Enabled() bool {
	return h.view.Enabled(h.id)
}

// SetEnabled changes whether the wrapped global accelerator is evaluated
// during the simulation.
//
// It panics if the handle does not reference a valid global accelerator.
func (h GlobalAcceleratorHandle) SetEnabled(enabled bool) {
	h.view.SetEnabled(h.id, enabled)
}

type globalAccelerator struct {
	solver   AccelerationSolver
	revision int32
	enabled  bool
}

func (s *globalAccelerator) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
