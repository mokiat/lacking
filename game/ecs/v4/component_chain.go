package ecs

type componentChain interface {
	// allocateOffset() uint32
	// releaseOffset(offset uint32)
}

type specificComponentChain[T any] struct {
	storage *specificComponentStorage[T]

	chunks    []uint32
	chunkSize uint32
}

func (c *specificComponentChain[T]) getRef(offset uint32) *T {
	chunkIndex := offset / c.chunkSize
	indexInChunk := offset % c.chunkSize
	return &c.chunks[chunkIndex][indexInChunk]
}
