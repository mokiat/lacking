package ecs

type componentArchetype struct {
	mask componentMask
	size uint32

	// TODO: Consider using an array to pointer mapping.
	components map[typeIndex]componentChain
}

func (a *componentArchetype) allocateOffset() uint32 {
	offset := a.size
	a.size++
	return offset
}

func (a *componentArchetype) releaseOffset(offset uint32) {
	panic("not implemented")
}
