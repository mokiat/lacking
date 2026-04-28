package ecs

type componentArchetype struct {
	mask componentMask
	size uint32

	// TODO: Consider using an array to pointer mapping.
	components map[typeIndex]componentChain
}

type componentChain interface{}

type specificComponentChain[T any] struct {
	storage *specificComponentStorage[T]

	chunks    [][]T
	chunkSize uint32
}

func (c *specificComponentChain[T]) getRef(offset uint32) *T {
	chunkIndex := offset / c.chunkSize
	indexInChunk := offset % c.chunkSize
	return &c.chunks[chunkIndex][indexInChunk]
}
