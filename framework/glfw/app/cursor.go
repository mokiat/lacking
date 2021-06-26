package app

import "github.com/go-gl/glfw/v3.3/glfw"

type customCursor struct {
	cursor *glfw.Cursor
}

func (c *customCursor) Delete() {
	c.cursor.Destroy()
	c.cursor = nil
}
