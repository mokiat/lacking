package std

import "github.com/mokiat/lacking/ui"

// ClickListener can be used to get notifications about click events.
type ClickListener func()

// ButtonState indicates the state of a Button control.
type ButtonState int

const (
	// ButtonStateUp indicates that the button is in its default state.
	ButtonStateUp ButtonState = iota

	// ButtonStateOver indicates that the cursor is over the button.
	ButtonStateOver

	// ButtonStateDown indicates that the cursor is pressing on the button.
	ButtonStateDown
)

// BaseButtonComponent provides a basic mouse event handling for
// a button component.
//
// Users are expected to compose this structure into a component implementation
// that can do the actual rendering.
type BaseButtonComponent struct {
	state   ButtonState
	onClick ClickListener
}

func (c *BaseButtonComponent) OnClickListener() ClickListener {
	return c.onClick
}

func (c *BaseButtonComponent) SetOnClickListener(onClick ClickListener) {
	c.onClick = onClick
}

func (c *BaseButtonComponent) State() ButtonState {
	return c.state
}

func (c *BaseButtonComponent) OnMouseEvent(element *ui.Element, event ui.MouseEvent) bool {
	switch event.Type {
	case ui.MouseEventTypeEnter:
		c.state = ButtonStateOver
		element.Invalidate()
		return true
	case ui.MouseEventTypeLeave:
		c.state = ButtonStateUp
		element.Invalidate()
		return true
	case ui.MouseEventTypeUp:
		if event.Button == ui.MouseButtonLeft {
			if c.state == ButtonStateDown {
				c.notifyClicked()
			}
			c.state = ButtonStateOver
			element.Invalidate()
			return true
		}
	case ui.MouseEventTypeDown:
		if event.Button == ui.MouseButtonLeft {
			c.state = ButtonStateDown
			element.Invalidate()
			return true
		}
	}
	return false
}

func (c *BaseButtonComponent) notifyClicked() {
	if c.onClick != nil {
		c.onClick()
	}
}
