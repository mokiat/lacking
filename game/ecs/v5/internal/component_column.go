package internal

type BaseColumn interface {
	Grow()

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
	size      uint32
}

var _ BaseColumn = (*Column[struct{}])(nil)

func (c *Column[T]) Grow() {
	chunkIndex := uint32(c.size) / chunkSize
	if chunkIndex >= uint32(len(c.chunkRefs)) {
		c.chunkRefs = append(c.chunkRefs, c.storage.AllocateChunk())
	}
}

func (c *Column[T]) StoragePosition(row ArchetypeRow) StoragePosition {
	return StoragePosition{
		ChunkIndex: uint32(row) / chunkSize,
		Offset:     uint32(row) % chunkSize,
	}
}

func (c *Column[T]) Destroy() {
	// TODO: Release chunks back to the storage's free chunk pool.
}
