package internal

type BaseColumn interface {
	StoragePosition(row ArchetypeRow) StoragePosition

	Destroy()
}

func newColumn[T any](storage *ComponentStorage[T]) *Column[T] {
	return &Column[T]{
		storage: storage,
	}
}

type Column[T any] struct {
	storage   *ComponentStorage[T]
	chunkRefs []uint32
}

var _ BaseColumn = (*Column[struct{}])(nil)

func (c *Column[T]) StoragePosition(row ArchetypeRow) StoragePosition {
	return StoragePosition{
		ChunkIndex: uint32(row) / chunkSize,
		Offset:     uint32(row) % chunkSize,
	}
}

func (c *Column[T]) Destroy() {
	// TODO
}
