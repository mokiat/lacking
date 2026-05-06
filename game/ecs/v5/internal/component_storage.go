package internal

const chunkSize = 64

// BaseComponentStorage represents a base interface for component storage.
type BaseComponentStorage interface {

	// CopyValue copies the component value from the source storage position to
	// the destination storage position.
	CopyValue(dst, src StoragePosition)

	// ApplyTempValue applies the temporary value to the component at the
	// specified storage position.
	ApplyTempValue(pos StoragePosition)
}

// NewComponentStorage creates a new component storage for components of type T.
func NewComponentStorage[T any]() *ComponentStorage[T] {
	return &ComponentStorage[T]{}
}

// ComponentStorage is a storage for components of a specific type T.
type ComponentStorage[T any] struct {
	chunks    [][chunkSize]T // TODO: Maybe array of "Chunk" types.
	tempValue T
}

var _ BaseComponentStorage = (*ComponentStorage[struct{}])(nil)

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
