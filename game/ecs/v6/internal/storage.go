package internal

import (
	"github.com/mokiat/gog/ds"
)

// AnyStorage represents a base interface for storage, which is responsible for
// allocating and managing columns for component storage.
type AnyStorage interface {

	// AllocateColumn allocates a new column for storing component values and
	// returns the ID of the allocated column.
	AllocateColumn() ColumnID

	// ReleaseColumn reclaims the column with the specified ID, making it
	// available for future allocations.
	ReleaseColumn(id ColumnID)

	// CopyCell copies the component value from the source column and row to the
	// destination column and row.
	CopyCell(dstColumnID ColumnID, dstRow Row, srcColumnID ColumnID, srcRow Row)

	// GrowColumn grows the column with the specified ID by adding an additional
	// row to it.
	GrowColumn(id ColumnID)

	// ShrinkColumn shrinks the column with the specified ID by removing the last
	// row from it.
	ShrinkColumn(id ColumnID)
}

// NewStorage creates a new storage for columns of type T.
func NewStorage[T any]() *Storage[T] {
	return &Storage[T]{
		chunks: ds.EmptyStack[DataChunk[T]](),

		freeColumns: ds.EmptyStack[ColumnID](),
	}
}

// Storage acts as a memory manager for columns by allocating and releasing
// columns and pooling them.
type Storage[T any] struct {
	chunks *ds.Stack[DataChunk[T]]

	freeColumns *ds.Stack[ColumnID]
	columns     []*Column[T]
}

var _ AnyStorage = (*Storage[struct{}])(nil)

// AllocateColumn allocates a new column for storing component values and
// returns the ID of the allocated column.
func (s *Storage[T]) AllocateColumn() ColumnID {
	return s.allocateColumnID()
}

// ReleaseColumn reclaims the column with the specified ID, making it
// available for future allocations.
func (s *Storage[T]) ReleaseColumn(id ColumnID) {
	column := s.Column(id)
	column.Release()
}

// NewColumn allocates a new column for storing component values of type T and
// returns the allocated column.
func (s *Storage[T]) NewColumn() *Column[T] {
	id := s.allocateColumnID()
	return s.columns[id]
}

// Column returns the column with the specified ID from the storage.
func (s *Storage[T]) Column(id ColumnID) *Column[T] {
	return s.columns[id]
}

// CopyCell copies the component value from the source column and row to the
// destination column and row.
func (s *Storage[T]) CopyCell(dstColumnID ColumnID, dstRow Row, srcColumnID ColumnID, srcRow Row) {
	dstColumn := s.Column(dstColumnID)
	srcColumn := s.Column(srcColumnID)
	dstColumn.SetValue(dstRow, srcColumn.Value(srcRow))
}

// GrowColumn grows the column with the specified ID by adding an additional
// row to it.
func (s *Storage[T]) GrowColumn(id ColumnID) {
	column := s.columns[id]
	column.Grow()
}

// ShrinkColumn shrinks the column with the specified ID by removing the last
// row from it.
func (s *Storage[T]) ShrinkColumn(id ColumnID) {
	column := s.columns[id]
	column.Shrink()
}

func (s *Storage[T]) allocateColumnID() ColumnID {
	if s.freeColumns.IsEmpty() {
		id := ColumnID(len(s.columns))
		column := NewColumn(s, ColumnID(id))
		s.columns = append(s.columns, column)
		return id
	}
	return s.freeColumns.Pop()
}

func (s *Storage[T]) releaseColumnID(columnID ColumnID) {
	s.freeColumns.Push(columnID)
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
