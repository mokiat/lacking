package physics

type GlobalAcceleratorID struct {
	index    int32
	revision int32
}

var NilGlobalAcceleratorID = GlobalAcceleratorID{}

type GlobalAcceleratorView struct {
	scene *Scene
}

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

func (v GlobalAcceleratorView) Delete(id GlobalAcceleratorID) {
	// TODO: Should I allow the deletion of an invalid ID (noop)?
	accelerator := v.resolve(id, true)
	accelerator.solver = nil // allow the solver to be garbage collected
	accelerator.revision++   // progress revision to invalid (even) value

	v.scene.releaseGlobalAccelerator(id.index)
}

func (v GlobalAcceleratorView) Handle(id GlobalAcceleratorID) GlobalAcceleratorHandle {
	return GlobalAcceleratorHandle{
		view: v,
		id:   id,
	}
}

func (v GlobalAcceleratorView) IsValid(id GlobalAcceleratorID) bool {
	accelerator := v.resolve(id, false)
	return accelerator != nil
}

func (v GlobalAcceleratorView) Solver(id GlobalAcceleratorID) AccelerationSolver {
	accelerator := v.resolve(id, true)
	return accelerator.solver
}

func (v GlobalAcceleratorView) SetSolver(id GlobalAcceleratorID, solver AccelerationSolver) {
	accelerator := v.resolve(id, true)
	accelerator.solver = solver
}

func (v GlobalAcceleratorView) Enabled(id GlobalAcceleratorID) bool {
	accelerator := v.resolve(id, true)
	return accelerator.enabled
}

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

type GlobalAcceleratorHandle struct {
	view GlobalAcceleratorView
	id   GlobalAcceleratorID
}

func (h GlobalAcceleratorHandle) ID() GlobalAcceleratorID {
	return h.id
}

func (h GlobalAcceleratorHandle) Delete() {
	h.view.Delete(h.id)
}

func (h GlobalAcceleratorHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

func (h GlobalAcceleratorHandle) Solver() AccelerationSolver {
	return h.view.Solver(h.id)
}

func (h GlobalAcceleratorHandle) SetSolver(solver AccelerationSolver) {
	h.view.SetSolver(h.id, solver)
}

func (h GlobalAcceleratorHandle) Enabled() bool {
	return h.view.Enabled(h.id)
}

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
