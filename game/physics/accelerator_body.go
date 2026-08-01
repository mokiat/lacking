package physics

type BodyAcceleratorID struct {
	index    int32
	revision int32
}

var NilBodyAcceleratorID = BodyAcceleratorID{}

type BodyAcceleratorView struct {
	scene *Scene
}

func (v BodyAcceleratorView) Create(bodyID BodyID, solver AccelerationSolver) BodyAcceleratorID {
	// TODO: Verify that the body exists!
	bodyIndex := bodyID.index

	index := v.scene.allocateBodyAccelerator()

	accelerator := &v.scene.bodyAccelerators[index]
	accelerator.solver = solver
	accelerator.revision++ // progress revision to valid (odd) value
	accelerator.isEnabled = true

	v.scene.attachBodyAccelerator(bodyIndex, index)

	return BodyAcceleratorID{
		index:    index,
		revision: accelerator.revision,
	}
}

func (v BodyAcceleratorView) Delete(id BodyAcceleratorID) {
	accelerator := v.resolve(id, true)
	accelerator.solver = nil // allow the solver to be garbage collected
	accelerator.revision++   // progress revision to invalid (even) value

	v.scene.detachBodyAccelerator(accelerator.bodyIndex, id.index)
	v.scene.releaseBodyAccelerator(id.index)
}

func (v BodyAcceleratorView) Handle(id BodyAcceleratorID) BodyAcceleratorHandle {
	return BodyAcceleratorHandle{
		view: v,
		id:   id,
	}
}

func (v BodyAcceleratorView) IsValid(id BodyAcceleratorID) bool {
	accelerator := v.resolve(id, false)
	return accelerator != nil
}

func (v BodyAcceleratorView) BodyID(id BodyAcceleratorID) BodyID {
	accelerator := v.resolve(id, true)
	bodyIndex := accelerator.bodyIndex
	body := &v.scene.bodies[bodyIndex]
	return BodyID{
		index:    bodyIndex,
		revision: body.revision,
	}
}

func (v BodyAcceleratorView) Solver(id BodyAcceleratorID) AccelerationSolver {
	accelerator := v.resolve(id, true)
	return accelerator.solver
}

func (v BodyAcceleratorView) SetSolver(id BodyAcceleratorID, solver AccelerationSolver) {
	accelerator := v.resolve(id, true)
	accelerator.solver = solver
}

func (v BodyAcceleratorView) Enabled(id BodyAcceleratorID) bool {
	accelerator := v.resolve(id, true)
	return accelerator.isEnabled
}

func (v BodyAcceleratorView) SetEnabled(id BodyAcceleratorID, enabled bool) {
	accelerator := v.resolve(id, true)
	accelerator.isEnabled = enabled
}

func (v BodyAcceleratorView) resolve(id BodyAcceleratorID, required bool) *bodyAccelerator {
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

type BodyAcceleratorHandle struct {
	view BodyAcceleratorView
	id   BodyAcceleratorID
}

func (h BodyAcceleratorHandle) ID() BodyAcceleratorID {
	return h.id
}

func (h BodyAcceleratorHandle) Delete() {
	h.view.Delete(h.id)
}

func (h BodyAcceleratorHandle) IsValid() bool {
	return h.view.IsValid(h.id)
}

func (h BodyAcceleratorHandle) BodyID() BodyID {
	return h.view.BodyID(h.id)
}

func (h BodyAcceleratorHandle) Solver() AccelerationSolver {
	return h.view.Solver(h.id)
}

func (h BodyAcceleratorHandle) SetSolver(solver AccelerationSolver) {
	h.view.SetSolver(h.id, solver)
}

func (h BodyAcceleratorHandle) Enabled() bool {
	return h.view.Enabled(h.id)
}

func (h BodyAcceleratorHandle) SetEnabled(enabled bool) {
	h.view.SetEnabled(h.id, enabled)
}

type bodyAccelerator struct {
	solver    AccelerationSolver
	revision  int32
	bodyIndex int32
	nextIndex int32
	isEnabled bool
}

func (s *bodyAccelerator) isValid() bool {
	return s.revision%2 == 1 // only odd revisions are valid
}
