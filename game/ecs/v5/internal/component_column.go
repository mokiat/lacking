package internal

import "math"

// BaseColumn represents a base interface for a column in the component storage.
// A column is responsible for managing the storage of component values for a
// specific component type across multiple entities.
type BaseColumn interface {

	// Storage returns the underlying component storage associated with this
	// column.
	Storage() BaseComponentStorage

	// Grow increases the size of the column by one, ensuring that there is enough
	// capacity to store a component value for a new entity.
	Grow()

	// Shrink decreases the size of the column by one, effectively removing the
	// last component value from the column.
	Shrink()

	// StoragePosition returns the storage position for the component value of the
	// entity at the given archetype row.
	StoragePosition(row ArchetypeRow) StoragePosition

	// Destroy releases any resources associated with the column, such as
	// allocated chunks, and resets the column to an empty state.
	Destroy()
}

func newColumn[T any](storage *ComponentStorage[T]) *Column[T] {
	return &Column[T]{
		storage: storage,
	}
}

// Column is a column in the component storage for a specific component type T.
// It manages the storage of component values for multiple entities, using
// chunks to efficiently allocate memory.
type Column[T any] struct {
	storage   *ComponentStorage[T]
	chunkRefs []uint32
	size      uint32
}

var _ BaseColumn = (*Column[struct{}])(nil)

// Storage returns the underlying component storage associated with this
// column.
func (c *Column[T]) Storage() BaseComponentStorage {
	return c.storage
}

// Grow increases the size of the column by one, ensuring that there is enough
// capacity to store a component value for a new entity.
func (c *Column[T]) Grow() {
	chunkPosition := c.size % chunkSize
	if chunkPosition == 0 {
		chunkRef := c.storage.AllocateChunk()
		c.chunkRefs = append(c.chunkRefs, chunkRef)
	}
	c.size++
}

// Shrink decreases the size of the column by one, effectively removing the
// last component value from the column.
func (c *Column[T]) Shrink() {
	chunkPosition := uint32(c.size-1) % chunkSize
	if chunkPosition == 0 {
		last := len(c.chunkRefs) - 1
		c.storage.ReleaseChunk(c.chunkRefs[last])
		c.chunkRefs = c.chunkRefs[:last]
	}
	c.size--
}

// StoragePosition returns the storage position for the component value of the
// entity at the given archetype row.
func (c *Column[T]) StoragePosition(row ArchetypeRow) StoragePosition {
	return StoragePosition{
		ChunkIndex: uint32(row) / chunkSize,
		Offset:     uint32(row) % chunkSize,
	}
}

// Destroy releases any resources associated with the column, such as
// allocated chunks, and resets the column to an empty state.
func (c *Column[T]) Destroy() {
	for _, chunkIndex := range c.chunkRefs {
		c.storage.ReleaseChunk(chunkIndex)
	}
	c.chunkRefs = nil
	c.size = 0
}

const (
	ColumnSizeTypeSmall ColumnSizeType = iota
	ColumnSizeTypeMedium
	ColumnSizeTypeLarge
	ColumnSizeTypeUnbounded

	ColumnSizeTypeCount
)

// ColumnSizeType describes the size capabilities of a column.
type ColumnSizeType uint8

// MaxSize returns the maximum number of component values that a column of this
// size type can hold.
func (t ColumnSizeType) MaxSize() uint32 {
	switch t {
	case ColumnSizeTypeSmall:
		return 2
	case ColumnSizeTypeMedium:
		return 8
	case ColumnSizeTypeLarge:
		return 64
	case ColumnSizeTypeUnbounded:
		return math.MaxUint32
	default:
		panic("invalid column size type")
	}
}
