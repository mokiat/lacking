package std

import "slices"

func (c *editboxComponent) hasSelection() bool {
	return c.cursorColumn != c.selectorColumn
}

func (c *editboxComponent) clearSelection() {
	c.selectorColumn = c.cursorColumn
}

func (c *editboxComponent) selectAll() {
	c.selectorColumn = 0
	c.cursorColumn = len(c.line)
}

func (c *editboxComponent) selectedColumns() (int, int) {
	fromColumn := min(c.cursorColumn, c.selectorColumn)
	toColumn := max(c.cursorColumn, c.selectorColumn)
	return fromColumn, toColumn
}

func (c *editboxComponent) selectedSegment() []rune {
	fromColumn, toColumn := c.selectedColumns()
	return slices.Clone(c.line[fromColumn:toColumn])
}
