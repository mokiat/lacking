package ecs

type componentChain interface {
	// allocateOffset() uint32
	// releaseOffset(offset uint32)
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
