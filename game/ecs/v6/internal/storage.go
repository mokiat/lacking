package internal

import (
	"github.com/mokiat/gog/ds"
)

// AnyStorage represents a base interface for storage, which is responsible for
// allocating and managing columns for component storage.
type AnyStorage interface {

	// NewAnyColumn allocates a new column for storing component values and
	// returns the allocated column.
	NewAnyColumn() AnyColumn

	// ResetBuffer clears the storage buffer, allowing it to be reused for new
	// values.
	ResetBuffer()
}

// NewStorage creates a new storage for columns of type T.
func NewStorage[T any]() *Storage[T] {
	return &Storage[T]{
		chunks:  ds.EmptyStack[DataChunk[T]](),
		columns: ds.EmptyStack[*Column[T]](),
		buffer:  make([]T, 0, 4),
	}
}

// Storage acts as a memory manager for columns by allocating and releasing
// columns and pooling them.
type Storage[T any] struct {
	chunks  *ds.Stack[DataChunk[T]]
	columns *ds.Stack[*Column[T]]
	buffer  []T
}

var _ AnyStorage = (*Storage[struct{}])(nil)

// NewAnyColumn allocates a new column for storing component values and
// returns the allocated column.
func (s *Storage[T]) NewAnyColumn() AnyColumn {
	return s.allocateColumn()
}

// NewColumn allocates a new column for storing component values of type T and
// returns the allocated column.
func (s *Storage[T]) NewColumn() *Column[T] {
	return s.allocateColumn()
}

// ResetBuffer clears the storage buffer, allowing it to be reused for new
// values.
func (s *Storage[T]) ResetBuffer() {
	clear(s.buffer)
	s.buffer = s.buffer[:0]
}

// WriteBuffer adds a new value to the storage buffer and returns the index of
// the appended value.
func (s *Storage[T]) WriteBuffer(value T) uint32 {
	s.buffer = append(s.buffer, value)
	return uint32(len(s.buffer) - 1)
}

// ReadBuffer reads a value from the storage buffer at the specified index.
func (s *Storage[T]) ReadBuffer(index uint32) T {
	return s.buffer[index]
}

func (s *Storage[T]) allocateColumn() *Column[T] {
	if s.columns.IsEmpty() {
		return NewColumn(s)
	}
	return s.columns.Pop()
}

func (s *Storage[T]) releaseColumn(column *Column[T]) {
	s.columns.Push(column)
}

func (s *Storage[T]) allocateChunk() DataChunk[T] {
	if s.chunks.IsEmpty() {
		return new([chunkSize]T)
	} else {
		return s.chunks.Pop()
	}
}

func (s *Storage[T]) releaseChunk(chunk DataChunk[T]) {
	s.chunks.Push(chunk)
}
