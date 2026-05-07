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
}

// NewStorage creates a new storage for columns of type T.
func NewStorage[T any]() *Storage[T] {
	return &Storage[T]{
		chunks:  ds.EmptyStack[DataChunk[T]](),
		columns: ds.EmptyStack[*Column[T]](),
	}
}

// Storage acts as a memory manager for columns by allocating and releasing
// columns and pooling them.
type Storage[T any] struct {
	chunks    *ds.Stack[DataChunk[T]]
	columns   *ds.Stack[*Column[T]]
	tempValue T
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

// TempValue returns the temporary value stored in the storage.
func (s *Storage[T]) TempValue() T {
	return s.tempValue
}

// SetTempValue sets the temporary value in the storage.
func (s *Storage[T]) SetTempValue(value T) {
	s.tempValue = value
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
