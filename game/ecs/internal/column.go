package internal

// ColumnID represents a unique identifier for a column in the component
// storage.
type ColumnID uint32

// NewColumn creates a new column for storing component values of type T using
// the provided storage.
func NewColumn[T any](storage *Storage[T], id ColumnID) *Column[T] {
	return &Column[T]{
		storage: storage,
		id:      id,
		chunks:  nil,
	}
}

// Column is a column in the component storage for a specific component type T.
// It manages the storage of component values for multiple entities, using
// chunks to efficiently allocate memory.
//
// NOTE: A huge benefit of using chunks is that they are immutable during
// growth, which means that references to existing chunks are not invalidated,
// unlike using a single slice and appending to it.
type Column[T any] struct {
	storage *Storage[T]
	id      ColumnID
	chunks  []DataChunk[T]
	size    uint32
}

// ID returns the unique identifier of the column in the component storage.
func (c *Column[T]) ID() ColumnID {
	return c.id
}

// Grow appends an additional row to the column. A zero value is placed.
func (c *Column[T]) Grow() {
	if c.size%chunkSize == 0 {
		c.chunks = append(c.chunks, c.storage.allocateChunk())
	}
	c.size++
}

// Shrink removes the last row of the column. The value is lost.
func (c *Column[T]) Shrink() {
	c.size--
	if c.size%chunkSize == 0 {
		lastChunkIndex := len(c.chunks) - 1
		c.storage.releaseChunk(c.chunks[lastChunkIndex])
		c.chunks = c.chunks[:lastChunkIndex]
	}
}

// Copy copies the component values from the source row to the destination
// row.
func (c *Column[T]) Copy(dst, src Row) {
	if dst != src {
		c.SetValue(dst, c.Value(src))
	}
}

// Value returns the value at the specified row in the column.
func (c *Column[T]) Value(row Row) T {
	chunkIndex := row / chunkSize
	cellIndex := row % chunkSize
	return c.chunks[chunkIndex][cellIndex]
}

// SetValue sets the value at the specified row in the column.
func (c *Column[T]) SetValue(row Row, value T) {
	chunkIndex := row / chunkSize
	cellIndex := row % chunkSize
	c.chunks[chunkIndex][cellIndex] = value
}

// RefValue returns a reference to the value at the specified row in the column.
func (c *Column[T]) RefValue(row Row) *T {
	chunkIndex := row / chunkSize
	cellIndex := row % chunkSize
	return &c.chunks[chunkIndex][cellIndex]
}

// Release releases any resources associated with the column, such as
// allocated chunks, and resets the column to an empty state.
func (c *Column[T]) Release() {
	for _, chunk := range c.chunks {
		c.storage.releaseChunk(chunk)
	}
	c.chunks = c.chunks[:0]
	c.size = 0

	c.storage.releaseColumnID(c.id)
}
