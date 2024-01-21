package physics

import "github.com/mokiat/lacking/game/physics/solver"

type Accelerator interface {
	Delete()
}

// TODO: The GlobalAccelerator can now be made hidden.

type GlobalAccelerator struct {
	scene     *Scene
	reference indexReference
}

func (a *GlobalAccelerator) Delete() {
	deleteGlobalAccelerator(a.scene, a.reference)
}

type globalAcceleratorState struct {
	reference indexReference
	logic     solver.Acceleration
}

func newGlobalAccelerator(scene *Scene, logic solver.Acceleration) *GlobalAccelerator {
	if scene.freeGlobalAcceleratorIndices.IsEmpty() {
		freeIndex := uint32(len(scene.globalAccelerators))
		reference := newIndexReference(freeIndex, scene.nextRevision())
		scene.globalAccelerators = append(scene.globalAccelerators, globalAcceleratorState{
			reference: reference,
			logic:     logic,
		})
		return &GlobalAccelerator{
			scene:     scene,
			reference: reference,
		}
	} else {
		freeIndex := scene.freeGlobalAcceleratorIndices.Pop()
		reference := newIndexReference(freeIndex, scene.nextRevision())
		scene.globalAccelerators[freeIndex] = globalAcceleratorState{
			reference: reference,
			logic:     logic,
		}
		return &GlobalAccelerator{
			scene:     scene,
			reference: reference,
		}
	}
}

func deleteGlobalAccelerator(scene *Scene, reference indexReference) {
	index := reference.Index()
	state := &scene.globalAccelerators[index]
	if state.reference != reference {
		panic("accelerator already deleted")
	}
	state.reference = newIndexReference(index, 0)
	state.logic = nil
	scene.freeGlobalAcceleratorIndices.Push(index)
}
