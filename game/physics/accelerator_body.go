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
	bodyView := v.scene.Bodies()
	body := bodyView.resolve(bodyID, true)

	index, accelerator := v.scene.allocateBodyAccelerator()
	*accelerator = bodyAccelerator{
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

func (v BodyAcceleratorView) Delete(id BodyAcceleratorID) {
	accelerator := v.resolve(id, true)

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

	*accelerator = bodyAccelerator{
		solver:    nil,                      // allow the solver to be garbage collected
		revision:  accelerator.revision + 1, // progress revision to invalid (even) value
		bodyIndex: nilIndex,
		nextIndex: nilIndex,
		isEnabled: false,
	}

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

func (v BodyAcceleratorView) idFromIndex(index int32) BodyAcceleratorID {
	accelerator := &v.scene.bodyAccelerators[index]
	return BodyAcceleratorID{
		index:    index,
		revision: accelerator.revision,
	}
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
