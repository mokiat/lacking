package ecs

type componentArchetype struct {
	mask componentMask
	size uint32

	// lookup     componentLookup
	// components []componentChain
}

func (a *componentArchetype) reset() {
	a.mask = emptyComponentMask()
	a.size = 0

	// for i := range a.lookup {
	// 	a.lookup[i] = -1
	// }
	// // TODO: Pool component chains as well?
	// clear(a.components)
	// a.components = a.components[:0]
}

func (a *componentArchetype) allocateOffset() uint32 {
	// offset := a.size
	// a.size++
	// return offset
	return 0
}

func (a *componentArchetype) releaseOffset(offset uint32) {
	panic("not implemented")
}
