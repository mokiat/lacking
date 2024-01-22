package physics

import "github.com/mokiat/lacking/game/physics/solver"

var invalidGlobalAcceleratorState = &globalAcceleratorState{}

type GlobalAccelerator struct {
	scene     *Scene
	reference indexReference
}

func (a GlobalAccelerator) Enabled() bool {
	state := a.state()
	return state.enabled
}

func (a GlobalAccelerator) SetEnabled(enabled bool) {
	state := a.state()
	state.enabled = enabled
}

func (a GlobalAccelerator) Delete() {
	deleteGlobalAccelerator(a.scene, a.reference)
}

func (a GlobalAccelerator) state() *globalAcceleratorState {
	index := a.reference.Index()
	state := &a.scene.globalAccelerators[index]
	if state.reference != a.reference {
		return invalidGlobalAcceleratorState
	}
	return state
}

type globalAcceleratorState struct {
	reference indexReference
	logic     solver.Acceleration
	enabled   bool
}

func newGlobalAccelerator(scene *Scene, logic solver.Acceleration) GlobalAccelerator {
	var freeIndex uint32
	if scene.freeGlobalAcceleratorIndices.IsEmpty() {
		freeIndex = uint32(len(scene.globalAccelerators))
		scene.globalAccelerators = append(scene.globalAccelerators, globalAcceleratorState{})
	} else {
		freeIndex = scene.freeGlobalAcceleratorIndices.Pop()
	}

	reference := newIndexReference(freeIndex, scene.nextRevision())
	scene.globalAccelerators[freeIndex] = globalAcceleratorState{
		reference: reference,
		logic:     logic,
		enabled:   true,
	}
	return GlobalAccelerator{
		scene:     scene,
		reference: reference,
	}
}

func deleteGlobalAccelerator(scene *Scene, reference indexReference) {
	index := reference.Index()
	state := &scene.globalAccelerators[index]
	if state.reference == reference {
		state.reference = newIndexReference(index, 0)
		state.logic = nil
		scene.freeGlobalAcceleratorIndices.Push(index)
	}
}
