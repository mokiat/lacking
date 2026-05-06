package internal

import "github.com/mokiat/gog/ds"

const chunkSize = 64

// BaseComponentStorage represents a base interface for component storage.
type BaseComponentStorage interface {

	// CreateColumn creates a new column for storing component values.
	CreateColumn() BaseColumn

	// AllocateChunk allocates a new chunk for storing component values and
	// returns the index of the allocated chunk.
	AllocateChunk() uint32

	// ReleaseChunk releases the chunk at the specified index, making it
	// available for future allocations.
	ReleaseChunk(chunkIndex uint32)

	// CopyValue copies the component value from the source storage position to
	// the destination storage position.
	CopyValue(dst, src StoragePosition)

	// ApplyTempValue applies the temporary value to the component at the
	// specified storage position.
	ApplyTempValue(pos StoragePosition)
}

// NewComponentStorage creates a new component storage for components of type T.
func NewComponentStorage[T any]() *ComponentStorage[T] {
	return &ComponentStorage[T]{
		freeChunks: ds.EmptyStack[uint32](),
	}
}

// ComponentStorage is a storage for components of a specific type T.
type ComponentStorage[T any] struct {
	freeChunks *ds.Stack[uint32]
	chunks     []*[chunkSize]T
	tempValue  T
}

var _ BaseComponentStorage = (*ComponentStorage[struct{}])(nil)

// CreateColumn creates a new column for storing component values.
func (s *ComponentStorage[T]) CreateColumn() BaseColumn {
	return newColumn(s)
}

// AllocateChunk allocates a new chunk for storing component values and
// returns the index of the allocated chunk.
func (s *ComponentStorage[T]) AllocateChunk() uint32 {
	if s.freeChunks.IsEmpty() {
		chunkIndex := uint32(len(s.chunks))
		s.chunks = append(s.chunks, new([chunkSize]T))
		return chunkIndex
	}
	return s.freeChunks.Pop()
}

// ReleaseChunk releases the chunk at the specified index, making it
// available for future allocations.
func (s *ComponentStorage[T]) ReleaseChunk(chunkIndex uint32) {
	s.freeChunks.Push(chunkIndex)
}

// GetValue returns the component value at the specified storage position.
func (s *ComponentStorage[T]) GetValue(pos StoragePosition) T {
	return s.chunks[pos.ChunkIndex][pos.Offset]
}

// SetValue sets the component value at the specified storage position.
func (s *ComponentStorage[T]) SetValue(pos StoragePosition, value T) {
	s.chunks[pos.ChunkIndex][pos.Offset] = value
}

// RefValue returns a reference to the component value at the specified storage
// position.
func (s *ComponentStorage[T]) RefValue(pos StoragePosition) *T {
	return &s.chunks[pos.ChunkIndex][pos.Offset]
}

// CopyValue copies the component value from the source storage position to the
// destination storage position.
func (s *ComponentStorage[T]) CopyValue(dst, src StoragePosition) {
	value := s.GetValue(src)
	s.SetValue(dst, value)
}

// GetTempValue returns the temporary value associated with the component
// storage.
func (s *ComponentStorage[T]) GetTempValue() T {
	return s.tempValue
}

// SetTempValue sets the temporary value associated with the component storage.
func (s *ComponentStorage[T]) SetTempValue(value T) {
	s.tempValue = value
}

// ApplyTempValue applies the temporary value to the component at the specified
// storage position.
func (s *ComponentStorage[T]) ApplyTempValue(pos StoragePosition) {
	s.SetValue(pos, s.GetTempValue())
}

// StoragePosition represents the location of a component within the storage.
type StoragePosition struct {

	// ChunkIndex is the index of the chunk where the component is stored.
	ChunkIndex uint32

	// Offset is the index within the chunk where the component is stored.
	Offset uint32
}

// TypePlacementMap is a mapping from component type identifiers to
// storage positions.
type TypePlacementMap [MaxComponentTypes]StoragePosition
