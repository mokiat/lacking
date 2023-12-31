package std

import (
	"math"
	"strconv"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
)

func (c *codeAreaComponent) drawFrame(element *ui.Element, canvas *ui.Canvas, bounds sprec.Vec2) {
	canvas.Reset()
	canvas.Rectangle(sprec.ZeroVec2(), bounds)
	canvas.Fill(ui.Fill{
		Color: SurfaceColor,
	})
}

func (c *codeAreaComponent) drawFrameBorder(element *ui.Element, canvas *ui.Canvas, bounds sprec.Vec2) {
	canvas.Reset()
	if element.IsFocused() {
		canvas.SetStrokeColor(SecondaryLightColor)
	} else {
		canvas.SetStrokeColor(PrimaryLightColor)
	}
	canvas.SetStrokeSize(codeAreaBorderSize)
	canvas.Rectangle(sprec.ZeroVec2(), bounds)
	canvas.Stroke()
}

func (c *codeAreaComponent) drawContent(element *ui.Element, canvas *ui.Canvas, bounds sprec.Vec2) {
	rulerPosition := sprec.ZeroVec2()
	rulerSize := sprec.Vec2{
		X: float32(c.rulerWidth),
		Y: bounds.Y,
	}
	canvas.Push()
	canvas.ClipRect(rulerPosition, rulerSize)
	canvas.Translate(rulerPosition)
	c.drawRuler(element, canvas, rulerSize)
	canvas.Pop()

	editorPosition := sprec.Vec2{
		X: rulerSize.X,
	}
	editorSize := sprec.Vec2{
		X: bounds.X - rulerSize.X,
		Y: bounds.Y,
	}
	canvas.Push()
	canvas.ClipRect(editorPosition, editorSize)
	canvas.Translate(editorPosition)
	c.drawSelection(element, canvas, editorSize)
	c.drawText(element, canvas, editorSize)
	c.drawCursor(element, canvas, editorSize)
	canvas.Pop()
}

func (c *codeAreaComponent) drawSelection(element *ui.Element, canvas *ui.Canvas, bounds sprec.Vec2) {
	if !c.hasSelection() || !element.IsFocused() {
		return
	}
	fromRow, toRow := c.selectedRows()
	visibleFromRow, visibleToRow := c.visibleRows(bounds)
	fromRow = max(fromRow, visibleFromRow)
	toRow = min(toRow, visibleToRow)
	if fromRow >= toRow {
		return
	}
	lineHeight := c.font.LineHeight(c.fontSize)

	canvas.Push()
	canvas.Translate(sprec.Vec2{
		X: float32(codeAreaTextPaddingLeft) - c.offsetX,
		Y: float32(fromRow)*lineHeight - c.offsetY,
	})
	for row := fromRow; row < toRow; row++ {
		fromColumn, toColumn := c.selectedColumns(row)
		visibleFromColumn, visibleToColumn := c.visibleColumns(row, bounds)
		fromColumn = max(fromColumn, visibleFromColumn)
		toColumn = min(toColumn, visibleToColumn)
		if fromColumn < toColumn {
			line := c.lines[row]
			selectionOffset := c.font.LineWidth(line[:fromColumn], c.fontSize)
			selectionWidth := c.font.LineWidth(line[fromColumn:toColumn], c.fontSize)
			selectionPosition := sprec.Vec2{
				X: selectionOffset,
			}
			selectionSize := sprec.Vec2{
				X: selectionWidth,
				Y: lineHeight,
			}
			canvas.Reset()
			canvas.Rectangle(selectionPosition, selectionSize)
			canvas.Fill(ui.Fill{
				Color: SecondaryLightColor,
			})
		}
		canvas.Translate(sprec.Vec2{
			Y: lineHeight,
		})
	}
	canvas.Pop()
}

func (c *codeAreaComponent) drawText(element *ui.Element, canvas *ui.Canvas, bounds sprec.Vec2) {
	fromRow, toRow := c.visibleRows(bounds)
	if fromRow >= toRow {
		return
	}
	lineHeight := c.font.LineHeight(c.fontSize)

	canvas.Push()
	canvas.Translate(sprec.Vec2{
		X: float32(codeAreaTextPaddingLeft) - c.offsetX,
		Y: float32(fromRow)*lineHeight - c.offsetY,
	})
	for row := fromRow; row < toRow; row++ {
		fromColumn, toColumn := c.visibleColumns(row, bounds)
		if fromColumn < toColumn {
			line := c.lines[row]
			visibleTextOffset := c.font.LineWidth(line[:fromColumn], c.fontSize)
			visibleTextPosition := sprec.Vec2{
				X: visibleTextOffset,
			}
			canvas.Reset()
			canvas.FillTextLine(line[fromColumn:toColumn], visibleTextPosition, ui.Typography{
				Font:  c.font,
				Size:  c.fontSize,
				Color: OnSurfaceColor,
			})
		}
		canvas.Translate(sprec.Vec2{
			Y: lineHeight,
		})
	}
	canvas.Pop()
}

func (c *codeAreaComponent) drawCursor(element *ui.Element, canvas *ui.Canvas, bounds sprec.Vec2) {
	if c.isReadOnly || !element.IsFocused() {
		return
	}
	fromRow, toRow := c.visibleRows(bounds)
	if c.cursorRow < fromRow || c.cursorRow >= toRow {
		return
	}
	fromColumn, toColumn := c.visibleColumns(c.cursorRow, bounds)
	if c.cursorColumn < fromColumn || c.cursorColumn > toColumn {
		return
	}
	line := c.lines[c.cursorRow]
	cursorOffset := c.font.LineWidth(line[:c.cursorColumn], c.fontSize)
	lineHeight := c.font.LineHeight(c.fontSize)
	cursorPosition := sprec.Vec2{
		X: float32(codeAreaTextPaddingLeft) + cursorOffset - c.offsetX,
		Y: float32(c.cursorRow)*lineHeight - c.offsetY,
	}
	cursorSize := sprec.Vec2{
		X: codeAreaCursorWidth,
		Y: lineHeight,
	}
	canvas.Reset()
	canvas.Rectangle(cursorPosition, cursorSize)
	canvas.Fill(ui.Fill{
		Color: PrimaryColor,
	})
}

func (c *codeAreaComponent) drawRuler(element *ui.Element, canvas *ui.Canvas, bounds sprec.Vec2) {
	canvas.Reset()
	canvas.Rectangle(
		sprec.ZeroVec2(),
		bounds,
	)
	canvas.Fill(ui.Fill{
		Color: PrimaryLightColor,
	})

	fromRow, toRow := c.visibleRows(bounds)
	if fromRow >= toRow {
		return
	}
	lineHeight := c.font.LineHeight(c.fontSize)
	canvas.Push()
	canvas.Translate(sprec.Vec2{
		X: codeAreaRulerPaddingLeft,
		Y: float32(fromRow)*lineHeight - c.offsetY,
	})
	for row := fromRow; row < toRow; row++ {
		text := []rune(strconv.Itoa(row + 1))
		canvas.Reset()
		canvas.FillTextLine(text, sprec.ZeroVec2(), ui.Typography{
			Font:  c.font,
			Size:  c.fontSize,
			Color: OnSurfaceColor,
		})
		canvas.Translate(sprec.Vec2{
			Y: lineHeight,
		})
	}
	canvas.Pop()
}

func (c *codeAreaComponent) visibleRows(bounds sprec.Vec2) (int, int) {
	lineHeight := c.font.LineHeight(c.fontSize)
	fromRow := int(c.offsetY / lineHeight)
	toRow := fromRow + int(math.Ceil(float64(bounds.Y/lineHeight))) + 1
	return max(fromRow, 0), min(toRow, len(c.lines))
}

func (c *codeAreaComponent) visibleColumns(row int, bounds sprec.Vec2) (int, int) {
	line := c.lines[row]
	minVisible := len(line)
	maxVisible := 0
	offset := float32(codeAreaTextPaddingLeft) - c.offsetX
	iterator := c.font.LineIterator(line, c.fontSize)
	column := 0
	for iterator.Next() {
		character := iterator.Character()
		characterWidth := character.Kern + character.Width
		if offset+characterWidth > 0.0 && offset < bounds.X {
			minVisible = min(minVisible, column)
			maxVisible = max(maxVisible, column+1)
		}
		offset += characterWidth
		column++
	}
	return minVisible, maxVisible
}
