package internal

import "github.com/mokiat/gog/ds"

// BaseStorage represents a base interface for storage, which is responsible for
// allocating and managing columns for component storage.
type BaseStorage interface {

	// AllocateColumn allocates a new column for storing component values and
	// returns the allocated column.
	//
	// Make sure to call Release on the column when done with it.
	AllocateColumn(sizeType ColumnSizeType) BaseColumn
}

// NewStorage creates a new storage for columns of type T.
func NewStorage[T any]() *Storage[T] {
	var columnPools [ColumnSizeTypeCount]*ds.Stack[*Column[T]]
	for i := range ColumnSizeTypeCount {
		columnPools[i] = ds.EmptyStack[*Column[T]]()
	}
	return &Storage[T]{
		columnPools: columnPools,
	}
}

// Storage acts as a memory manager for columns by allocating and releasing
// columns and pooling them.
type Storage[T any] struct {
	columnPools [ColumnSizeTypeCount]*ds.Stack[*Column[T]]
	tempValue   T
}

var _ BaseStorage = (*Storage[struct{}])(nil)

// AllocateColumn allocates a new column for storing component values and
// returns the allocated column.
func (s *Storage[T]) AllocateColumn(sizeType ColumnSizeType) BaseColumn {
	if pool := s.columnPools[sizeType]; !pool.IsEmpty() {
		return pool.Pop()
	}
	return NewColumn[T](s, sizeType)
}

// ReleaseColumn releases the specified column, making it available for
// future allocations.
func (s *Storage[T]) ReleaseColumn(column *Column[T]) {
	sizeType := column.SizeType()
	s.columnPools[sizeType].Push(column)
}

// TempValue returns the temporary value stored in the storage.
func (s *Storage[T]) TempValue() T {
	return s.tempValue
}

// SetTempValue sets the temporary value in the storage.
func (s *Storage[T]) SetTempValue(value T) {
	s.tempValue = value
}
