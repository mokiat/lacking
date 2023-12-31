package std

import (
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/shortcut"
	"github.com/mokiat/lacking/ui/state"
)

const (
	codeAreaHistoryCapacity            = 100
	codeAreaPaddingLeft                = 2
	codeAreaPaddingRight               = 2
	codeAreaPaddingTop                 = 2
	codeAreaPaddingBottom              = 2
	codeAreaTextPaddingLeft            = 5
	codeAreaTextPaddingRight           = 5
	codeAreaRulerPaddingLeft           = 5
	codeAreaRulerPaddingRight          = 5
	codeAreaCursorWidth                = float32(1.0)
	codeAreaBorderSize                 = float32(2.0)
	codeAreaKeyScrollSpeed             = 20
	codeAreaFontSize                   = float32(18.0)
	codeAreaChangeAccumulationDuration = time.Second
)

var CodeArea = co.Define(&codeAreaComponent{})

type CodeAreaData struct {
	ReadOnly bool
	Code     string
}

type CodeAreaCallbackData struct {
	OnChange func(string)
}

var _ ui.ElementHistoryHandler = (*codeAreaComponent)(nil)
var _ ui.ElementClipboardHandler = (*codeAreaComponent)(nil)
var _ ui.ElementResizeHandler = (*codeAreaComponent)(nil)
var _ ui.ElementRenderHandler = (*codeAreaComponent)(nil)
var _ ui.ElementKeyboardHandler = (*codeAreaComponent)(nil)
var _ ui.ElementMouseHandler = (*codeAreaComponent)(nil)

type codeAreaComponent struct {
	co.BaseComponent

	history *state.History

	font     *ui.Font
	fontSize float32

	cursorRow      int
	cursorColumn   int
	selectorRow    int
	selectorColumn int

	isReadOnly bool
	lines      [][]rune
	onChange   func(string)

	textWidth  int
	textHeight int
	rulerWidth int

	offsetX    float32
	offsetY    float32
	maxOffsetX float32
	maxOffsetY float32

	isDragging bool
}

func (c *codeAreaComponent) OnCreate() {
	c.history = state.NewHistory(codeAreaHistoryCapacity)

	c.font = co.OpenFont(c.Scope(), "fonts/roboto-mono-regular.ttf")
	c.fontSize = codeAreaFontSize

	c.cursorRow = 0
	c.cursorColumn = 0

	data := co.GetData[CodeAreaData](c.Properties())
	c.isReadOnly = data.ReadOnly
	c.lines = c.textToLines(data.Code)
	c.refreshTextSize()
}

func (c *codeAreaComponent) OnUpsert() {
	data := co.GetData[CodeAreaData](c.Properties())
	if data.ReadOnly != c.isReadOnly {
		c.isReadOnly = data.ReadOnly
		c.history.Clear()
	}
	if data.Code != c.constructText() {
		c.history.Clear()
		c.lines = c.textToLines(data.Code)
		c.refreshTextSize()
	}

	callbackData := co.GetOptionalCallbackData[CodeAreaCallbackData](c.Properties(), CodeAreaCallbackData{})
	c.onChange = callbackData.OnChange

	c.cursorRow = min(c.cursorRow, len(c.lines)-1)
	c.cursorColumn = min(c.cursorColumn, len(c.lines[c.cursorRow]))
	c.selectorRow = min(c.selectorRow, len(c.lines)-1)
	c.selectorColumn = min(c.selectorColumn, len(c.lines[c.selectorRow]))
}

func (c *codeAreaComponent) Render() co.Instance {
	padding := ui.Spacing{
		Left:   codeAreaPaddingLeft,
		Right:  codeAreaPaddingRight,
		Top:    codeAreaPaddingTop,
		Bottom: codeAreaPaddingBottom,
	}
	return co.New(Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(ElementData{
			Essence:   c,
			Focusable: opt.V(true),
			IdealSize: opt.V(ui.Size{
				Width:  c.textWidth + c.rulerWidth,
				Height: c.textHeight,
			}.Grow(padding.Size())),
		})
	})
}

func (c *codeAreaComponent) OnUndo(element *ui.Element) bool {
	canUndo := c.history.CanUndo()
	if canUndo {
		c.history.Undo()
		c.handleChanged()
	}
	return canUndo
}

func (c *codeAreaComponent) OnRedo(element *ui.Element) bool {
	canRedo := c.history.CanRedo()
	if canRedo {
		c.history.Redo()
		c.handleChanged()
	}
	return canRedo
}

func (c *codeAreaComponent) OnClipboardEvent(element *ui.Element, event ui.ClipboardEvent) bool {
	switch event.Action {
	case ui.ClipboardActionCut:
		if c.isReadOnly {
			return false
		}
		if c.hasSelection() {
			text := c.linesToText(c.selectedLines())
			element.Window().RequestCopy(text)
			c.applyChange(c.createChangeDeleteSelection())
			c.handleChanged()
		}
		return true

	case ui.ClipboardActionCopy:
		if c.hasSelection() {
			text := c.linesToText(c.selectedLines())
			element.Window().RequestCopy(text)
		}
		return true

	case ui.ClipboardActionPaste:
		if c.isReadOnly {
			return false
		}
		lines := c.textToLines(event.Text)
		if c.hasSelection() {
			c.applyChange(c.createChangeReplaceSelection(lines))
		} else {
			c.applyChange(c.createChangeInsertLines(lines))
		}
		c.handleChanged()
		return true

	default:
		return false
	}
}

func (c *codeAreaComponent) OnResize(element *ui.Element, bounds ui.Bounds) {
	c.refreshScrollBounds(element)
}

func (c *codeAreaComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
	c.refreshScrollBounds(element)

	bounds := canvas.DrawBounds(element, false)
	contentBounds := canvas.DrawBounds(element, true)

	canvas.Push()
	canvas.ClipRect(bounds.Position, bounds.Size)
	canvas.Translate(bounds.Position)
	c.drawFrame(element, canvas, bounds.Size)
	canvas.Pop()

	canvas.Push()
	canvas.ClipRect(contentBounds.Position, contentBounds.Size)
	canvas.Translate(contentBounds.Position)
	c.drawContent(element, canvas, contentBounds.Size)
	canvas.Pop()

	canvas.Push()
	canvas.ClipRect(bounds.Position, bounds.Size)
	canvas.Translate(bounds.Position)
	c.drawFrameBorder(element, canvas, bounds.Size)
	canvas.Pop()
}

func (c *codeAreaComponent) OnKeyboardEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	switch event.Action {
	case ui.KeyboardActionDown, ui.KeyboardActionRepeat:
		consumed := c.onKeyboardPressEvent(element, event)
		if consumed {
			element.Invalidate()
		}
		return consumed

	case ui.KeyboardActionType:
		consumed := c.onKeyboardTypeEvent(element, event)
		if consumed {
			element.Invalidate()
		}
		return consumed

	default:
		return false
	}
}

func (c *codeAreaComponent) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	switch event.Action {
	case ui.MouseActionScroll:
		if event.Modifiers.Contains(ui.KeyModifierShift) && (event.ScrollX == 0) {
			c.offsetX -= event.ScrollY
		} else {
			c.offsetX -= event.ScrollX
			c.offsetY -= event.ScrollY
		}
		c.offsetX = min(max(c.offsetX, 0), c.maxOffsetX)
		c.offsetY = min(max(c.offsetY, 0), c.maxOffsetY)
		element.Invalidate()
		return true

	case ui.MouseActionDown:
		if event.Button != ui.MouseButtonLeft {
			return false
		}
		c.isDragging = true
		c.cursorRow = c.findCursorRow(element, event.Y)
		c.cursorColumn = c.findCursorColumn(element, event.X)
		extendSelection := event.Modifiers.Contains(ui.KeyModifierShift)
		if !extendSelection {
			c.clearSelection()
		}
		element.Invalidate()
		return true

	case ui.MouseActionMove:
		if c.isDragging {
			c.cursorRow = c.findCursorRow(element, event.Y)
			c.cursorColumn = c.findCursorColumn(element, event.X)
			element.Invalidate()
		}
		return true

	case ui.MouseActionUp:
		if event.Button != ui.MouseButtonLeft {
			return false
		}
		if c.isDragging {
			c.isDragging = false
			c.cursorRow = c.findCursorRow(element, event.Y)
			c.cursorColumn = c.findCursorColumn(element, event.X)
			element.Invalidate()
		}
		return true

	default:
		return false
	}
}

func (c *codeAreaComponent) onKeyboardPressEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	os := element.Window().Platform().OS()
	extendSelection := event.Modifiers.Contains(ui.KeyModifierShift)

	if shortcut.IsClose(os, event) {
		return false // propagate up
	}
	if shortcut.IsSave(os, event) {
		return false // propagate up
	}
	if shortcut.IsCut(os, event) {
		element.Window().Cut()
		return true
	}
	if shortcut.IsCopy(os, event) {
		element.Window().Copy()
		return true
	}
	if shortcut.IsPaste(os, event) {
		element.Window().Paste()
		return true
	}
	if shortcut.IsUndo(os, event) {
		element.Window().Undo()
		return true
	}
	if shortcut.IsRedo(os, event) {
		element.Window().Redo()
		return true
	}
	if shortcut.IsSelectAll(os, event) {
		c.selectAll()
		return true
	}
	if shortcut.IsJumpToLineStart(os, event) {
		if !c.isReadOnly {
			c.moveCursorToStartOfLine()
			if !extendSelection {
				c.clearSelection()
			}
		}
		return true
	}
	if shortcut.IsJumpToLineEnd(os, event) {
		if !c.isReadOnly {
			c.moveCursorToEndOfLine()
			if !extendSelection {
				c.clearSelection()
			}
		}
		return true
	}
	if shortcut.IsJumpToDocumentStart(os, event) {
		if !c.isReadOnly {
			c.moveCursorToStartOfDocument()
			if !extendSelection {
				c.clearSelection()
			}
		}
		return true
	}
	if shortcut.IsJumpToDocumentEnd(os, event) {
		if !c.isReadOnly {
			c.moveCursorToEndOfDocument()
			if !extendSelection {
				c.clearSelection()
			}
		}
		return true
	}

	switch event.Code {

	case ui.KeyCodeEscape:
		c.isDragging = false
		c.clearSelection()
		return true

	case ui.KeyCodeArrowUp:
		if c.isReadOnly {
			c.scrollUp()
		} else {
			c.moveCursorUp()
			if !extendSelection {
				c.clearSelection()
			}
		}
		return true

	case ui.KeyCodeArrowDown:
		if c.isReadOnly {
			c.scrollDown()
		} else {
			c.moveCursorDown()
			if !extendSelection {
				c.clearSelection()
			}
		}
		return true

	case ui.KeyCodeArrowLeft:
		if c.isReadOnly {
			c.scrollLeft()
		} else {
			if extendSelection {
				c.moveCursorLeft()
			} else {
				if c.hasSelection() {
					c.moveCursorToSelectionStart()
				} else {
					c.moveCursorLeft()
				}
				c.clearSelection()
			}
		}
		return true

	case ui.KeyCodeArrowRight:
		if c.isReadOnly {
			c.scrollRight()
		} else {
			if extendSelection {
				c.moveCursorRight()
			} else {
				if c.hasSelection() {
					c.moveCursorToSelectionEnd()
				} else {
					c.moveCursorRight()
				}
				c.clearSelection()
			}
		}
		return true

	case ui.KeyCodeBackspace:
		if c.isReadOnly {
			return false
		}
		if c.hasSelection() {
			c.applyChange(c.createChangeDeleteSelection())
		} else {
			c.applyChange(c.createChangeDeleteCharacterLeft())
		}
		c.handleChanged()
		return true

	case ui.KeyCodeDelete:
		if c.isReadOnly {
			return false
		}
		if c.hasSelection() {
			c.applyChange(c.createChangeDeleteSelection())
		} else {
			c.applyChange(c.createChangeDeleteCharacterRight())
		}
		c.handleChanged()
		return true

	case ui.KeyCodeEnter:
		if c.isReadOnly {
			return false
		}
		lines := [][]rune{
			{},
			{},
		}
		if c.hasSelection() {
			c.applyChange(c.createChangeReplaceSelection(lines))
		} else {
			c.applyChange(c.createChangeInsertLines(lines))
		}
		c.handleChanged()
		return true

	case ui.KeyCodeTab:
		event.Rune = '\t'
		return c.onKeyboardTypeEvent(element, event)

	default:
		return false
	}
}

func (c *codeAreaComponent) onKeyboardTypeEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	if c.isReadOnly {
		return false
	}
	lines := [][]rune{
		{event.Rune},
	}
	if c.hasSelection() {
		c.applyChange(c.createChangeReplaceSelection(lines))
	} else {
		c.applyChange(c.createChangeInsertLines(lines))
	}
	c.handleChanged()
	return true
}

func (c *codeAreaComponent) findCursorRow(element *ui.Element, y int) int {
	y += int(c.offsetY)
	y -= element.Padding().Top

	lineHeight := c.font.LineHeight(c.fontSize)
	row := y / int(lineHeight)
	return min(max(0, row), len(c.lines)-1)
}

func (c *codeAreaComponent) findCursorColumn(element *ui.Element, x int) int {
	x += int(c.offsetX)
	x -= element.Padding().Left
	x -= c.rulerWidth
	x -= codeAreaTextPaddingLeft

	bestColumn := 0
	bestDistance := sprec.Abs(float32(x))

	column := 1
	offset := float32(0.0)
	iterator := c.font.LineIterator(c.lines[c.cursorRow], c.fontSize)
	for iterator.Next() {
		character := iterator.Character()
		offset += character.Kern + character.Width
		if distance := sprec.Abs(float32(x) - offset); distance < bestDistance {
			bestColumn = column
			bestDistance = distance
		}
		column++
	}
	return bestColumn
}

func (c *codeAreaComponent) refreshTextSize() {
	txtWidth := float32(0.0)
	for _, line := range c.lines {
		lineWidth := c.font.LineWidth(line, c.fontSize)
		txtWidth = max(txtWidth, lineWidth)
	}
	txtHeight := c.font.LineHeight(c.fontSize) * float32(len(c.lines))

	c.textWidth = codeAreaTextPaddingLeft + int(math.Ceil(float64(txtWidth))) + codeAreaTextPaddingRight
	c.textHeight = int(math.Ceil(float64(txtHeight)))

	rulerText := strconv.Itoa(len(c.lines))
	digitSize := c.font.LineWidth([]rune(rulerText), c.fontSize)
	rulerTextWidth := int(math.Ceil(float64(digitSize)))
	c.rulerWidth = codeAreaRulerPaddingLeft + rulerTextWidth + codeAreaRulerPaddingRight
}

func (c *codeAreaComponent) refreshScrollBounds(element *ui.Element) {
	bounds := element.ContentBounds()

	textPadding := codeAreaTextPaddingLeft + codeAreaTextPaddingRight
	availableTextWidth := bounds.Width - c.rulerWidth - textPadding
	availableTextHeight := bounds.Height
	c.maxOffsetX = float32(max(c.textWidth-availableTextWidth, 0))
	c.maxOffsetY = float32(max(c.textHeight-availableTextHeight, 0))
	c.offsetX = min(max(c.offsetX, 0), c.maxOffsetX)
	c.offsetY = min(max(c.offsetY, 0), c.maxOffsetY)
}

func (c *codeAreaComponent) constructText() string {
	return c.linesToText(c.lines)
}

func (c *codeAreaComponent) linesToText(lines [][]rune) string {
	return strings.Join(gog.Map(lines, func(line []rune) string {
		return string(line)
	}), "\n")
}

func (c *codeAreaComponent) textToLines(text string) [][]rune {
	return gog.Map(strings.Split(text, "\n"), func(line string) []rune {
		return []rune(line)
	})
}

func (c *codeAreaComponent) handleChanged() {
	c.refreshTextSize()
	if c.onChange != nil {
		c.onChange(c.constructText())
	}
}
