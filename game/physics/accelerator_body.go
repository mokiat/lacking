package physics

type BodyAcceleratorID struct {
	index    int32
	revision int32
}

var NilBodyAcceleratorID = BodyAcceleratorID{}

type BodyAcceleratorView struct {
	scene *Scene
}

func (s BodyAcceleratorView) Create(bodyID BodyID, solver AccelerationSolver) BodyAcceleratorID {
	// TODO: Verify that the body exists!
	bodyIndex := bodyID.index

	index := s.scene.allocateBodyAccelerator()

	accelerator := &s.scene.bodyAccelerators[index]
	accelerator.solver = solver
	accelerator.revision++ // progress ID to valid (odd) state
	accelerator.isEnabled = true

	s.scene.attachBodyAccelerator(bodyIndex, index)

	return BodyAcceleratorID{
		index:    index,
		revision: accelerator.revision,
	}
}

func (s BodyAcceleratorView) Delete(id BodyAcceleratorID) {
	// TODO: Should I allow the deletion of an invalid ID (noop)?
	accelerator := s.resolve(id, true)
	accelerator.solver = nil // allow the solver to be garbage collected
	accelerator.revision++   // progress ID to invalid (event) state

	s.scene.detachBodyAccelerator(accelerator.bodyIndex, id.index)
	s.scene.releaseBodyAccelerator(id.index)
}

func (s BodyAcceleratorView) Handle(id BodyAcceleratorID) BodyAcceleratorHandle {
	return BodyAcceleratorHandle{
		view: s,
		id:   id,
	}
}

func (s BodyAcceleratorView) IsValid(id BodyAcceleratorID) bool {
	accelerator := s.resolve(id, false)
	return accelerator != nil
}

func (s BodyAcceleratorView) BodyID(id BodyAcceleratorID) BodyID {
	accelerator := s.resolve(id, true)
	bodyIndex := accelerator.bodyIndex
	bodyRevision := s.scene.bodies[bodyIndex].reference.Revision
	return BodyID{
		index:    bodyIndex,
		revision: int32(bodyRevision),
	}
}

func (s BodyAcceleratorView) Solver(id BodyAcceleratorID) AccelerationSolver {
	accelerator := s.resolve(id, true)
	return accelerator.solver
}

func (s BodyAcceleratorView) SetSolver(id BodyAcceleratorID, solver AccelerationSolver) {
	accelerator := s.resolve(id, true)
	accelerator.solver = solver
}

func (s BodyAcceleratorView) Enabled(id BodyAcceleratorID) bool {
	accelerator := s.resolve(id, true)
	return accelerator.isEnabled
}

func (s BodyAcceleratorView) SetEnabled(id BodyAcceleratorID, enabled bool) {
	accelerator := s.resolve(id, true)
	accelerator.isEnabled = enabled
}

func (s BodyAcceleratorView) resolve(id BodyAcceleratorID, required bool) *bodyAccelerator {
	if id.revision == 0 {
		if required {
			panic("invalid body accelerator ID")
		}
		return nil
	}
	accelerator := &s.scene.bodyAccelerators[id.index]
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
