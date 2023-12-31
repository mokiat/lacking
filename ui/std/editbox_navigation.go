package std

func (c *editboxComponent) moveCursorLeft() {
	if c.cursorColumn > 0 {
		c.cursorColumn--
	}
}

func (c *editboxComponent) moveCursorRight() {
	if c.cursorColumn < len(c.line) {
		c.cursorColumn++
	}
}

func (c *editboxComponent) moveCursorToStartOfLine() {
	c.cursorColumn = 0
	c.offsetX = 0.0
}

func (c *editboxComponent) moveCursorToEndOfLine() {
	c.cursorColumn = len(c.line)
	c.offsetX = c.maxOffsetX
}

func (c *editboxComponent) moveCursorToSelectionStart() {
	c.cursorColumn = min(c.cursorColumn, c.selectorColumn)
}

func (c *editboxComponent) moveCursorToSelectionEnd() {
	c.cursorColumn = max(c.cursorColumn, c.selectorColumn)
}

func (c *editboxComponent) scrollLeft() {
	c.offsetX -= editboxKeyScrollSpeed
	c.offsetX = min(max(c.offsetX, 0), c.maxOffsetX)
}

func (c *editboxComponent) scrollRight() {
	c.offsetX += editboxKeyScrollSpeed
	c.offsetX = min(max(c.offsetX, 0), c.maxOffsetX)
}
