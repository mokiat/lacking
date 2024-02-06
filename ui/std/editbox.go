package std

import (
	"time"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/ui"
	co "github.com/mokiat/lacking/ui/component"
	"github.com/mokiat/lacking/ui/shortcut"
	"github.com/mokiat/lacking/ui/state"
)

const (
	editboxHistoryCapacity            = 100
	editboxPaddingLeft                = 10
	editboxPaddingRight               = 10
	editboxPaddingTop                 = 5
	editboxPaddingBottom              = 5
	editboxTextPaddingLeft            = 2
	editboxTextPaddingRight           = 2
	editboxCursorWidth                = float32(1.0)
	editboxBorderSize                 = float32(2.0)
	editboxBorderRadius               = float32(8.0)
	editboxKeyScrollSpeed             = 20
	editboxFontSize                   = float32(18.0)
	editboxChangeAccumulationDuration = time.Second
)

var EditBox = co.Define(&editboxComponent{})

type EditBoxData struct {
	ReadOnly bool
	Text     string
	Focused  bool
}

type EditBoxCallbackData struct {
	OnChange func(string)
	OnSubmit func(string)
	OnReject func()
}

var _ ui.ElementHistoryHandler = (*editboxComponent)(nil)
var _ ui.ElementClipboardHandler = (*editboxComponent)(nil)
var _ ui.ElementResizeHandler = (*editboxComponent)(nil)
var _ ui.ElementRenderHandler = (*editboxComponent)(nil)
var _ ui.ElementKeyboardHandler = (*editboxComponent)(nil)
var _ ui.ElementMouseHandler = (*editboxComponent)(nil)

type editboxComponent struct {
	co.BaseComponent

	history *state.History

	font     *ui.Font
	fontSize float32

	cursorColumn   int
	selectorColumn int

	isFocused  bool
	isReadOnly bool
	line       []rune
	onChange   func(string)
	onSubmit   func(string)
	onReject   func()

	textWidth  int
	textHeight int

	offsetX    float32
	maxOffsetX float32

	isDragging bool
}

func (c *editboxComponent) OnCreate() {
	// TODO: Make it possible to pass the history from outside so that
	// it can be persisted in a model?
	c.history = state.NewHistory(editboxHistoryCapacity)

	c.font = co.OpenFont(c.Scope(), "ui:///roboto-mono-regular.ttf")
	c.fontSize = editboxFontSize

	c.cursorColumn = 0
	c.selectorColumn = 0

	data := co.GetData[EditBoxData](c.Properties())
	c.isFocused = data.Focused
	c.isReadOnly = data.ReadOnly
	c.line = []rune(data.Text)
	c.refreshTextSize()
}

func (c *editboxComponent) OnUpsert() {
	data := co.GetData[EditBoxData](c.Properties())
	if data.ReadOnly != c.isReadOnly {
		c.isReadOnly = data.ReadOnly
		c.history.Clear()
	}
	if data.Text != string(c.line) {
		c.history.Clear()
		c.line = []rune(data.Text)
		c.refreshTextSize()
	}

	callbackData := co.GetOptionalCallbackData(c.Properties(), EditBoxCallbackData{})
	c.onChange = callbackData.OnChange
	c.onSubmit = callbackData.OnSubmit
	c.onReject = callbackData.OnReject

	c.cursorColumn = min(c.cursorColumn, len(c.line))
	c.selectorColumn = min(c.selectorColumn, len(c.line))
}

func (c *editboxComponent) Render() co.Instance {
	padding := ui.Spacing{
		Left:   editboxPaddingLeft,
		Right:  editboxPaddingRight,
		Top:    editboxPaddingTop,
		Bottom: editboxPaddingBottom,
	}
	textPadding := editboxTextPaddingLeft + editboxTextPaddingRight

	return co.New(Element, func() {
		co.WithLayoutData(c.Properties().LayoutData())
		co.WithData(ElementData{
			Essence:   c,
			Focusable: opt.V(true),
			Focused:   opt.V(c.isFocused),
			IdealSize: opt.V(ui.Size{
				Width:  c.textWidth + textPadding,
				Height: c.textHeight,
			}.Grow(padding.Size())),
			Padding: padding,
		})
	})
}

func (c *editboxComponent) OnUndo(element *ui.Element) bool {
	canUndo := c.history.CanUndo()
	if canUndo {
		c.history.Undo()
		c.handleChanged()
	}
	return canUndo
}

func (c *editboxComponent) OnRedo(element *ui.Element) bool {
	canRedo := c.history.CanRedo()
	if canRedo {
		c.history.Redo()
		c.handleChanged()
	}
	return canRedo
}

func (c *editboxComponent) OnClipboardEvent(element *ui.Element, event ui.ClipboardEvent) bool {
	switch event.Action {
	case ui.ClipboardActionCut:
		if c.isReadOnly {
			return false
		}
		if c.hasSelection() {
			text := string(c.selectedSegment())
			element.Window().RequestCopy(text)
			c.applyChange(c.createChangeDeleteSelection())
			c.handleChanged()
		}
		return true

	case ui.ClipboardActionCopy:
		if c.hasSelection() {
			text := string(c.selectedSegment())
			element.Window().RequestCopy(text)
		}
		return true

	case ui.ClipboardActionPaste:
		if c.isReadOnly {
			return false
		}
		if c.hasSelection() {
			c.applyChange(c.createChangeReplaceSelection([]rune(event.Text)))
		} else {
			c.applyChange(c.createChangeInsertSegment([]rune(event.Text)))
		}
		c.handleChanged()
		return true

	default:
		return false
	}
}

func (c *editboxComponent) OnResize(element *ui.Element, bounds ui.Bounds) {
	c.refreshScrollBounds(element)
}

func (c *editboxComponent) OnRender(element *ui.Element, canvas *ui.Canvas) {
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

func (c *editboxComponent) OnKeyboardEvent(element *ui.Element, event ui.KeyboardEvent) bool {
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

func (c *editboxComponent) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	switch event.Action {
	case ui.MouseActionScroll:
		if event.Modifiers.Contains(ui.KeyModifierShift) && (event.ScrollX == 0) {
			c.offsetX -= event.ScrollY
		} else {
			c.offsetX -= event.ScrollX
		}
		c.offsetX = min(max(c.offsetX, 0), c.maxOffsetX)
		element.Invalidate()
		return true

	case ui.MouseActionDown:
		if event.Button != ui.MouseButtonLeft {
			return false
		}
		c.isDragging = true
		c.cursorColumn = c.findCursorColumn(element, event.X)
		extendSelection := event.Modifiers.Contains(ui.KeyModifierShift)
		if !extendSelection {
			c.clearSelection()
		}
		element.Invalidate()
		return true

	case ui.MouseActionMove:
		if c.isDragging {
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
			c.cursorColumn = c.findCursorColumn(element, event.X)
			element.Invalidate()
		}
		return true

	default:
		return false
	}
}

func (c *editboxComponent) onKeyboardPressEvent(element *ui.Element, event ui.KeyboardEvent) bool {
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

	switch event.Code {

	case ui.KeyCodeEscape:
		if c.isDragging || c.hasSelection() {
			c.isDragging = false
			c.clearSelection()
		} else {
			c.handleRejected()
		}
		return true

	case ui.KeyCodeArrowUp:
		if !c.isReadOnly {
			c.moveCursorToStartOfLine()
			if !extendSelection {
				c.clearSelection()
			}
		}
		return true

	case ui.KeyCodeArrowDown:
		if !c.isReadOnly {
			c.moveCursorToEndOfLine()
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
		c.handleSubmitted()
		return true

	case ui.KeyCodeTab:
		return false

	default:
		return false
	}
}

func (c *editboxComponent) onKeyboardTypeEvent(element *ui.Element, event ui.KeyboardEvent) bool {
	if c.isReadOnly {
		return false
	}
	if c.hasSelection() {
		c.applyChange(c.createChangeReplaceSelection([]rune{event.Rune}))
	} else {
		c.applyChange(c.createChangeInsertSegment([]rune{event.Rune}))
	}
	c.handleChanged()
	return true
}

func (c *editboxComponent) findCursorColumn(element *ui.Element, x int) int {
	x += int(c.offsetX)
	x -= element.Padding().Left
	x -= editboxTextPaddingLeft

	bestColumn := 0
	bestDistance := sprec.Abs(float32(x))

	column := 1
	offset := float32(0.0)
	iterator := c.font.LineIterator(c.line, c.fontSize)
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

func (c *editboxComponent) refreshTextSize() {
	c.textWidth = editboxTextPaddingLeft + int(c.font.LineWidth(c.line, c.fontSize)) + editboxTextPaddingRight
	c.textHeight = int(c.font.LineHeight(c.fontSize))
}

func (c *editboxComponent) refreshScrollBounds(element *ui.Element) {
	bounds := element.ContentBounds()
	availableTextWidth := bounds.Width - editboxTextPaddingLeft - editboxTextPaddingRight
	c.maxOffsetX = float32(max(c.textWidth-availableTextWidth, 0))
	c.offsetX = min(max(c.offsetX, 0), c.maxOffsetX)
}

func (c *editboxComponent) handleChanged() {
	c.refreshTextSize()
	if c.onChange != nil {
		c.onChange(string(c.line))
	}
}

func (c *editboxComponent) handleSubmitted() {
	if c.onSubmit != nil {
		c.onSubmit(string(c.line))
	}
}

func (c *editboxComponent) handleRejected() {
	if c.onReject != nil {
		c.onReject()
	}
}
