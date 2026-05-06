package internal

import (
	"math"

	"github.com/mokiat/gog"
)

// BaseColumn represents a base interface for a column in the component storage.
// A column is responsible for managing the storage of component values for a
// specific component type across multiple entities.
type BaseColumn interface {

	// BaseStorage returns the underlying storage associated with this column.
	BaseStorage() BaseStorage

	// SizeType returns the size type of the column, which indicates the maximum
	// number of component values that the column can hold.
	SizeType() ColumnSizeType

	// CanGrow returns whether the column can grow to accommodate more component
	// values. This is determined by the column's size type and current size.
	CanGrow() bool

	// Grow appends an additional row to the column. A zero value is placed.
	Grow()

	// Shrink removes the last row of the column. The value is lost.
	Shrink()

	// Copy copies the component values from the source row to the destination
	// row.
	Copy(dst, src Row)

	// CopyFromStorage copies the temporary value from the storage into the
	// cell at the specified row.
	CopyFromStorage(dst Row)

	// CopyFromColumn copies the component values from the source column and row to
	// the destination row.
	CopyFromColumn(dst Row, srcColumn BaseColumn, src Row)

	// Release releases any resources associated with the column.
	Release()
}

// NewColumn creates a new column for storing values of type T.
func NewColumn[T any](storage *Storage[T], sizeType ColumnSizeType) *Column[T] {
	return &Column[T]{
		storage:  storage,
		sizeType: sizeType,
	}
}

// Column is a column in the component storage for a specific component type T.
// It manages the storage of component values for multiple entities, using
// chunks to efficiently allocate memory.
type Column[T any] struct {
	storage  *Storage[T]
	sizeType ColumnSizeType
	values   []T
}

var _ BaseColumn = (*Column[struct{}])(nil)

// BaseStorage returns the underlying storage associated with this column.
func (c *Column[T]) BaseStorage() BaseStorage {
	return c.storage
}

// Storage returns the storage associated with this column.
func (c *Column[T]) Storage() *Storage[T] {
	return c.storage
}

// SizeType returns the size type of the column, which indicates the maximum
// number of component values that the column can hold.
func (c *Column[T]) SizeType() ColumnSizeType {
	return c.sizeType
}

// CanGrow returns whether the column can grow to accommodate more component
// values. This is determined by the column's size type and current size.
func (c *Column[T]) CanGrow() bool {
	return len(c.values) < c.sizeType.MaxSize()
}

// Grow appends an additional row to the column. A zero value is placed.
func (c *Column[T]) Grow() {
	c.values = append(c.values, gog.Zero[T]())
}

// Shrink removes the last row of the column. The value is lost.
func (c *Column[T]) Shrink() {
	c.values = c.values[:len(c.values)-1]
}

// Copy copies the component values from the source row to the destination
// row.
func (c *Column[T]) Copy(dst, src Row) {
	c.values[dst] = c.values[src]
}

// CopyFromStorage copies the temporary value from the storage into the
// cell at the specified row.
func (c *Column[T]) CopyFromStorage(dst Row) {
	c.values[dst] = c.storage.TempValue()
}

// CopyFromColumn copies the component values from the source column and row to
// the destination row.
func (c *Column[T]) CopyFromColumn(dst Row, srcColumn BaseColumn, src Row) {
	srcCol := srcColumn.(*Column[T])
	c.values[dst] = srcCol.values[src]
}

// Value returns the value at the specified row in the column.
func (c *Column[T]) Value(row Row) T {
	return c.values[row]
}

// SetValue sets the value at the specified row in the column.
func (c *Column[T]) SetValue(row Row, value T) {
	c.values[row] = value
}

// RefValue returns a reference to the value at the specified row in the column.
func (c *Column[T]) RefValue(row Row) *T {
	return &c.values[row]
}

// Destroy releases any resources associated with the column, such as
// allocated chunks, and resets the column to an empty state.
func (c *Column[T]) Release() {
	c.values = c.values[:0]
	c.storage.ReleaseColumn(c)
}

const (
	// ColumnSizeTypeSmall represents a column that can hold up to 2 component
	// values.
	ColumnSizeTypeSmall ColumnSizeType = iota

	// ColumnSizeTypeMedium represents a column that can hold up to 8 component
	// values.
	ColumnSizeTypeMedium

	// ColumnSizeTypeLarge represents a column that can hold up to 64 component
	// values.
	ColumnSizeTypeLarge

	// ColumnSizeTypeUnbounded represents a column that can hold an unlimited
	// number of component values, up to the maximum uint32 value.
	ColumnSizeTypeUnbounded

	// ColumnSizeTypeCount is the total number of column size types defined.
	//
	// NOTE: This is not a valid enum value!
	ColumnSizeTypeCount
)

// ColumnSizeType describes the size capabilities of a column.
type ColumnSizeType uint8

// MaxSize returns the maximum number of component values that a column of this
// size type can hold.
func (t ColumnSizeType) MaxSize() int {
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
