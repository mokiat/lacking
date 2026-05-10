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

	// CopyCell copies the component value from the source column and row to the
	// destination column and row.
	CopyCell(dstColumnID ColumnID, dstRow Row, srcColumnID ColumnID, srcRow Row)

	// GrowColumn grows the column with the specified ID by adding an additional
	// row to it.
	GrowColumn(id ColumnID)

	// ShrinkColumn shrinks the column with the specified ID by removing the last
	// row from it.
	ShrinkColumn(id ColumnID)

	// ReclaimColumn reclaims the column with the specified ID, making it
	// available for future allocations.
	ReclaimColumn(id ColumnID)
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

// ReclaimColumn reclaims the column with the specified ID, making it
// available for future allocations.
func (s *Storage[T]) ReclaimColumn(id ColumnID) {
	column := s.columns[id]
	column.Release()
}

func (s *Storage[T]) allocateColumn() *Column[T] {
	if s.freeColumns.IsEmpty() {
		id := uint32(len(s.columns))
		column := NewColumn(s, ColumnID(id))
		s.columns = append(s.columns, column)
		return column
	}

	id := s.freeColumns.Pop()
	return s.columns[id]
}

func (s *Storage[T]) releaseColumn(column *Column[T]) {
	s.freeColumns.Push(column.ID())
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
