package mat

import (
	"github.com/mokiat/lacking/ui"
)

// NewButtonBaseEssence creates a new ButtonBaseEssence instance.
func NewButtonBaseEssence(onClick ClickListener) *ButtonBaseEssence {
	return &ButtonBaseEssence{
		state:   ButtonStateUp,
		onClick: onClick,
	}
}

var _ ui.ElementMouseHandler = (*ButtonBaseEssence)(nil)

// ButtonBaseEssence provides a basic mouse event handling for
// a button control.
// You are expected to compose this structure into an essence that
// can do the actual rendering.
type ButtonBaseEssence struct {
	state   ButtonState
	onClick ClickListener
}

// SetOnClick changes the ClickListener.
func (e *ButtonBaseEssence) SetOnClick(onClick ClickListener) {
	e.onClick = onClick
}

// State returns the current state of the button.
func (e *ButtonBaseEssence) State() ButtonState {
	return e.state
}

func (e *ButtonBaseEssence) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	switch event.Type {
	case ui.MouseEventTypeEnter:
		e.state = ButtonStateOver
		element.Invalidate()
		return true
	case ui.MouseEventTypeLeave:
		e.state = ButtonStateUp
		element.Invalidate()
		return true
	case ui.MouseEventTypeUp:
		if event.Button == ui.MouseButtonLeft {
			if e.state == ButtonStateDown {
				e.onClick()
			}
			e.state = ButtonStateOver
			element.Invalidate()
			return true
		}
	case ui.MouseEventTypeDown:
		if event.Button == ui.MouseButtonLeft {
			e.state = ButtonStateDown
			element.Invalidate()
			return true
		}
	}
	return false
}
