package std

import "slices"

func (c *codeAreaComponent) hasSelection() bool {
	return c.cursorRow != c.selectorRow || c.cursorColumn != c.selectorColumn
}

func (c *codeAreaComponent) clearSelection() {
	c.selectorRow = c.cursorRow
	c.selectorColumn = c.cursorColumn
}

func (c *codeAreaComponent) selectAll() {
	c.selectorRow = 0
	c.selectorColumn = 0
	c.cursorRow = len(c.lines) - 1
	c.cursorColumn = len(c.lines[c.cursorRow])
}

func (c *codeAreaComponent) selectedRows() (int, int) {
	switch {
	case c.cursorRow < c.selectorRow:
		return c.cursorRow, c.selectorRow + 1
	case c.selectorRow < c.cursorRow:
		return c.selectorRow, c.cursorRow + 1
	default:
		return c.cursorRow, c.cursorRow + 1
	}
}

func (c *codeAreaComponent) selectedColumns(row int) (int, int) {
	if row == c.cursorRow && row == c.selectorRow {
		fromColumn := min(c.cursorColumn, c.selectorColumn)
		toColumn := max(c.cursorColumn, c.selectorColumn)
		return fromColumn, toColumn
	}
	if row < c.cursorRow && row < c.selectorRow {
		return 0, 0
	}
	if row > c.cursorRow && row > c.selectorRow {
		return 0, 0
	}
	switch row {
	case c.cursorRow:
		if c.cursorRow < c.selectorRow {
			return c.cursorColumn, len(c.lines[row])
		} else {
			return 0, c.cursorColumn
		}
	case c.selectorRow:
		if c.selectorRow < c.cursorRow {
			return c.selectorColumn, len(c.lines[row])
		} else {
			return 0, c.selectorColumn
		}
	default:
		return 0, len(c.lines[row])
	}
}

func (c *codeAreaComponent) selectedLines() [][]rune {
	fromRow, toRow := c.selectedRows()
	if fromRow >= toRow {
		return [][]rune{}
	}
	var result [][]rune
	for row := fromRow; row < toRow; row++ {
		fromColumn, toColumn := c.selectedColumns(row)
		segment := slices.Clone(c.lines[row][fromColumn:toColumn])
		result = append(result, segment)
	}
	return result
}
