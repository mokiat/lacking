package ecs

type componentArchetype struct {
	mask componentMask
	size uint32

	lookup     componentLookup
	components []componentChain
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

func (a *componentArchetype) isEmpty() bool {
	return a.size == 0
}

func (a *componentArchetype) getChain(id typeID) componentChain {
	index := a.lookup[id]
	return a.components[index]
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

func getChain[T any](archetype *componentArchetype, compType *ComponentType[T]) *specificComponentChain[T] {
	anyChain := archetype.getChain(compType.id())
	return anyChain.(*specificComponentChain[T])
}
