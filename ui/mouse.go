package ui

import "fmt"

// MouseEvent represents an event related to a mouse action.
type MouseEvent struct {
	// Index indicates which mouse triggered the event. By default
	// the index for a the primary mouse is 0.
	// This is applicable for devices with multiple pointers
	// (mobile) or in case a second mouse is emulated
	// (e.g. with a game controller).
	Index int

	// X specifies the X position relative to the receiver control.
	X int

	// Y specifies the Y position relative to the receiver control.
	Y int

	// Type specifies the mouse event type.
	Type MouseEventType

	// Button specifies the button for which the event is
	// applicable.
	Button MouseButton
}

func (e MouseEvent) String() string {
	return fmt.Sprintf("(%d,(%d,%d),%s,%s)",
		e.Index,
		e.X,
		e.Y,
		e.Type,
		e.Button,
	)
}

// MouseEventType represents the type of mouse event.
type MouseEventType int

const (
	// MouseEventTypeDown indicates that a mouse button
	// was pressed down over the receiver control.
	MouseEventTypeDown MouseEventType = 1 + iota

	// MouseEventTypeUp indicates that a mouse button
	// was released over the receiver control.
	MouseEventTypeUp

	// MouseEventTypeMove indicates that the mouse was
	// moved over the receiver control.
	MouseEventTypeMove

	// MouseEventTypeDrag indicates that the mouse that
	// was previously pressed within the receiver control
	// is being moved.
	// The even could be received for a motion outside the
	// bounds of the control.
	MouseEventTypeDrag

	// MouseEventTypeDragCancel indicates that a drag operation
	// was cancelled by the parent control (other control might
	// have taken over).
	MouseEventTypeDragCancel

	// MouseEventTypeEnter indicates that the mouse has
	// entered the bounds of the control.
	MouseEventTypeEnter

	// MouseEventTypeLeave indicates that the mouse has
	// left the bounds of the control.
	// If the mouse was being dragged, the control may
	// receive further events.
	MouseEventTypeLeave
)

func (t MouseEventType) String() string {
	switch t {
	case MouseEventTypeDown:
		return "down"
	case MouseEventTypeUp:
		return "up"
	case MouseEventTypeMove:
		return "move"
	case MouseEventTypeDrag:
		return "drag"
	case MouseEventTypeDragCancel:
		return "drag_cancel"
	case MouseEventTypeEnter:
		return "enter"
	case MouseEventTypeLeave:
		return "leave"
	default:
		return "unknown"
	}
}

// MouseButton represents the mouse button.
type MouseButton int

const (
	// MouseButtonLeft specifies the left mouse button.
	MouseButtonLeft = 1 + iota

	// MouseButtonMiddle specifies the middle mouse button.
	MouseButtonMiddle

	// MouseButtonRight specifies the right mouse button.
	MouseButtonRight
)

func (b MouseButton) String() string {
	switch b {
	case MouseButtonLeft:
		return "left"
	case MouseButtonMiddle:
		return "middle"
	case MouseButtonRight:
		return "right"
	default:
		return "unknown"
	}
}
