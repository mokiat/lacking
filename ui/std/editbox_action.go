package std

import (
	"slices"

	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/ui/state"
)

func (c *editboxComponent) createActionMoveCursor(column int) state.Action {
	return func() {
		c.cursorColumn = column
	}
}

func (c *editboxComponent) createActionMoveSelector(column int) state.Action {
	return func() {
		c.selectorColumn = column
	}
}

func (c *editboxComponent) createActionInsertSegment(column int, segment []rune) state.Action {
	return func() {
		if len(segment) > 0 {
			prefix := c.line[:column]
			suffix := c.line[column:]
			c.line = gog.Concat(prefix, segment, suffix)
		}
	}
}

func (c *editboxComponent) createActionDeleteSegment(fromColumn, toColumn int) state.Action {
	return func() {
		if fromColumn < toColumn {
			c.line = slices.Delete(c.line, fromColumn, toColumn)
		}
	}
}
