package ecs

type componentChain interface {
	allocateCell() storagePosition
	releaseCell(cell storagePosition)
}

type specificComponentChain[T any] struct {
	compType *ComponentType[T]
	chunks   []uint32
}

func (c *specificComponentChain[T]) getRef(offset uint32) *T {
	pos := storagePosition{
		chunkID:     c.chunks[offset/chunkSize],
		chunkOffset: offset % chunkSize,
	}
	return c.compType.refValue(pos)
}

func (c *specificComponentChain[T]) allocateCell() storagePosition {
	panic("not implemented")
}

func (c *specificComponentChain[T]) releaseCell(cell storagePosition) {
	panic("not implemented")
}
