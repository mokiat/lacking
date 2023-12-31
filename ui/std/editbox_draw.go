package std

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
)

func (c *editboxComponent) drawFrame(element *ui.Element, canvas *ui.Canvas, bounds sprec.Vec2) {
	canvas.Reset()
	canvas.RoundRectangle(
		sprec.ZeroVec2(),
		bounds,
		sprec.NewVec4(
			editboxBorderRadius,
			editboxBorderRadius,
			editboxBorderRadius,
			editboxBorderRadius,
		),
	)
	canvas.Fill(ui.Fill{
		Color: SurfaceColor,
	})
}

func (c *editboxComponent) drawFrameBorder(element *ui.Element, canvas *ui.Canvas, bounds sprec.Vec2) {
	canvas.Reset()
	if element.IsFocused() {
		canvas.SetStrokeColor(SecondaryLightColor)
	} else {
		canvas.SetStrokeColor(PrimaryLightColor)
	}
	canvas.SetStrokeSize(editboxBorderSize)
	canvas.RoundRectangle(
		sprec.ZeroVec2(),
		bounds,
		sprec.NewVec4(
			editboxBorderRadius,
			editboxBorderRadius,
			editboxBorderRadius,
			editboxBorderRadius,
		),
	)
	canvas.Stroke()
}

func (c *editboxComponent) drawContent(element *ui.Element, canvas *ui.Canvas, bounds sprec.Vec2) {
	c.drawSelection(element, canvas, bounds)
	c.drawText(element, canvas, bounds)
	c.drawCursor(element, canvas, bounds)
}

func (c *editboxComponent) drawSelection(element *ui.Element, canvas *ui.Canvas, bounds sprec.Vec2) {
	if !c.hasSelection() || !element.IsFocused() {
		return
	}
	fromColumn, toColumn := c.selectedColumns()
	visibleFromColumn, visibleToColumn := c.visibleColumns(bounds)
	fromColumn = max(fromColumn, visibleFromColumn)
	toColumn = min(toColumn, visibleToColumn)
	if fromColumn >= toColumn {
		return
	}
	selectionOffset := c.font.LineWidth(c.line[:fromColumn], c.fontSize)
	selectionWidth := c.font.LineWidth(c.line[fromColumn:toColumn], c.fontSize)
	selectionHeight := c.font.LineHeight(c.fontSize)
	selectionPosition := sprec.Vec2{
		X: editboxTextPaddingLeft + selectionOffset - c.offsetX,
	}
	selectionSize := sprec.Vec2{
		X: selectionWidth,
		Y: selectionHeight,
	}
	canvas.Reset()
	canvas.Rectangle(selectionPosition, selectionSize)
	canvas.Fill(ui.Fill{
		Color: SecondaryLightColor,
	})
}

func (c *editboxComponent) drawText(element *ui.Element, canvas *ui.Canvas, bounds sprec.Vec2) {
	if len(c.line) == 0 {
		return
	}
	fromColumn, toColumn := c.visibleColumns(bounds)
	if fromColumn >= toColumn {
		return
	}
	visibleTextOffset := c.font.LineWidth(c.line[:fromColumn], c.fontSize)
	visibleTextPosition := sprec.Vec2{
		X: editboxTextPaddingLeft + visibleTextOffset - c.offsetX,
	}
	visibleText := c.line[fromColumn:toColumn]
	canvas.Reset()
	canvas.FillTextLine(visibleText, visibleTextPosition, ui.Typography{
		Font:  c.font,
		Size:  c.fontSize,
		Color: OnSurfaceColor,
	})
}

func (c *editboxComponent) drawCursor(element *ui.Element, canvas *ui.Canvas, bounds sprec.Vec2) {
	if c.isReadOnly || !element.IsFocused() {
		return
	}
	fromColumn, toColumn := c.visibleColumns(bounds)
	if c.cursorColumn < fromColumn || c.cursorColumn > toColumn {
		return
	}
	cursorOffset := c.font.LineWidth(c.line[:c.cursorColumn], c.fontSize)
	cursorHeight := c.font.LineHeight(c.fontSize)
	cursorPosition := sprec.Vec2{
		X: editboxTextPaddingLeft + cursorOffset - c.offsetX,
	}
	cursorSize := sprec.Vec2{
		X: editboxCursorWidth,
		Y: cursorHeight,
	}
	canvas.Reset()
	canvas.Rectangle(cursorPosition, cursorSize)
	canvas.Fill(ui.Fill{
		Color: PrimaryColor,
	})
}

func (c *editboxComponent) visibleColumns(bounds sprec.Vec2) (int, int) {
	fromColumn := len(c.line)
	toColumn := 0
	offset := float32(editboxTextPaddingLeft) - float32(c.offsetX)
	iterator := c.font.LineIterator(c.line, c.fontSize)
	column := 0
	for iterator.Next() {
		character := iterator.Character()
		characterWidth := character.Kern + character.Width
		if offset+characterWidth > 0.0 && offset < bounds.X {
			fromColumn = min(fromColumn, column)
			toColumn = max(toColumn, column+1)
		}
		offset += characterWidth
		column++
	}
	return fromColumn, toColumn
}
