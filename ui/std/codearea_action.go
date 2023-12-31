package std

import (
	"slices"

	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/ui/state"
)

func (c *codeAreaComponent) createActionMoveCursor(row, column int) state.Action {
	return func() {
		c.cursorRow = row
		c.cursorColumn = column
	}
}

func (c *codeAreaComponent) createActionMoveSelector(row, column int) state.Action {
	return func() {
		c.selectorRow = row
		c.selectorColumn = column
	}
}

func (c *codeAreaComponent) createActionInsertSegment(row, column int, segment []rune) state.Action {
	return func() {
		if len(segment) > 0 {
			line := c.lines[row]
			prefix := line[:column]
			suffix := line[column:]
			c.lines[row] = gog.Concat(prefix, segment, suffix)
		}
	}
}

func (c *codeAreaComponent) createActionDeleteSegment(row, fromColumn, toColumn int) state.Action {
	return func() {
		if fromColumn < toColumn {
			c.lines[row] = slices.Delete(c.lines[row], fromColumn, toColumn)
		}
	}
}

func (c *codeAreaComponent) createActionInsertLines(row int, lines [][]rune) state.Action {
	return func() {
		if len(lines) > 0 {
			c.lines = slices.Insert(c.lines, row, slices.Clone(lines)...)
		}
	}
}

func (c *codeAreaComponent) createActionDeleteLines(fromRow, toRow int) state.Action {
	return func() {
		if fromRow < toRow {
			c.lines = slices.Delete(c.lines, fromRow, toRow)
		}
	}
}

func (c *codeAreaComponent) createActionInsertSpan(row, column int, span [][]rune) state.Action {
	if len(span) == 0 {
		return func() {}
	}
	if len(span) == 1 {
		return c.createActionInsertSegment(row, column, span[0])
	}
	return func() {
		prefix := slices.Clone(c.lines[row][:column])
		suffix := slices.Clone(c.lines[row][column:])
		c.lines[row] = append(prefix, span[0]...)
		c.lines = slices.Insert(c.lines, row+1, span[1:]...)
		c.lines[row+len(span)-1] = append(c.lines[row+len(span)-1], suffix...)
	}
}

func (c *codeAreaComponent) createActionDeleteSpan(fromRow, fromColumn, toRow, toColumn int) state.Action {
	if fromRow >= toRow {
		return func() {}
	}
	if fromRow == toRow-1 {
		return c.createActionDeleteSegment(fromRow, fromColumn, toColumn)
	}
	return func() {
		c.lines[fromRow] = c.lines[fromRow][:fromColumn]
		c.lines[toRow-1] = c.lines[toRow-1][toColumn:]
		c.lines[fromRow] = append(c.lines[fromRow], c.lines[toRow-1]...)
		c.lines = slices.Delete(c.lines, fromRow+1, toRow)
	}
}
