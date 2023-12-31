package std

func (c *codeAreaComponent) moveCursorLeft() {
	if c.cursorColumn > 0 {
		c.cursorColumn--
	} else {
		if c.cursorRow > 0 {
			c.moveCursorUp()
			c.moveCursorToEndOfLine()
		}
	}
}

func (c *codeAreaComponent) moveCursorRight() {
	if c.cursorColumn < len(c.lines[c.cursorRow]) {
		c.cursorColumn++
	} else {
		if c.cursorRow < len(c.lines)-1 {
			c.moveCursorDown()
			c.moveCursorToStartOfLine()
		}
	}
}

func (c *codeAreaComponent) moveCursorUp() {
	if c.cursorRow > 0 {
		c.cursorRow--
		c.cursorColumn = min(c.cursorColumn, len(c.lines[c.cursorRow]))
	} else {
		c.moveCursorToStartOfLine()
	}
}

func (c *codeAreaComponent) moveCursorDown() {
	if c.cursorRow < len(c.lines)-1 {
		c.cursorRow++
		c.cursorColumn = min(c.cursorColumn, len(c.lines[c.cursorRow]))
	} else {
		c.moveCursorToEndOfLine()
	}
}

func (c *codeAreaComponent) moveCursorToStartOfLine() {
	c.cursorColumn = 0
	c.offsetX = 0.0
}

func (c *codeAreaComponent) moveCursorToEndOfLine() {
	c.cursorColumn = len(c.lines[c.cursorRow])
	// NOTE: Moving scroll to end is not always correct in this case.
}

func (c *codeAreaComponent) moveCursorToStartOfDocument() {
	c.cursorRow = 0
	c.offsetY = 0.0
	c.moveCursorToStartOfLine()
}

func (c *codeAreaComponent) moveCursorToEndOfDocument() {
	c.cursorRow = len(c.lines) - 1
	c.offsetY = c.maxOffsetY
	c.moveCursorToEndOfLine()
}

func (c *codeAreaComponent) moveCursorToSelectionStart() {
	fromRow, _ := c.selectedRows()
	fromColumn, _ := c.selectedColumns(fromRow)
	c.cursorRow, c.cursorColumn = fromRow, fromColumn
}

func (c *codeAreaComponent) moveCursorToSelectionEnd() {
	_, toRow := c.selectedRows()
	_, toColumn := c.selectedColumns(toRow - 1)
	c.cursorRow, c.cursorColumn = toRow-1, toColumn
}

func (c *codeAreaComponent) scrollLeft() {
	c.offsetX -= codeAreaKeyScrollSpeed
	c.offsetX = min(max(c.offsetX, 0), c.maxOffsetX)
}

func (c *codeAreaComponent) scrollRight() {
	c.offsetX += codeAreaKeyScrollSpeed
	c.offsetX = min(max(c.offsetX, 0), c.maxOffsetX)
}

func (c *codeAreaComponent) scrollUp() {
	c.offsetY -= codeAreaKeyScrollSpeed
	c.offsetY = min(max(c.offsetY, 0), c.maxOffsetY)
}

func (c *codeAreaComponent) scrollDown() {
	c.offsetY += codeAreaKeyScrollSpeed
	c.offsetY = min(max(c.offsetY, 0), c.maxOffsetY)
}
