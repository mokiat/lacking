package app

import "fmt"

// MouseEvent represents an event related to a mouse action.
type MouseEvent struct {

	// Index indicates which mouse triggered the event. By default
	// the index for a the primary mouse is 0.
	// This is applicable for devices with multiple pointers
	// (mobile) or in case a second mouse is emulated
	// (e.g. with a game controller).
	Index int

	// X specifies the horizontal position of the mouse.
	X int

	// Y specifies the vertical position of the mouse.
	Y int

	// Type specifies the mouse event type.
	Type MouseEventType

	// Button specifies the button for which the event is
	// applicable.
	Button MouseButton

	// ScrollX determines the amount of horizontal scroll.
	ScrollX float64

	// ScrollY determines the amount of vertical scroll.
	ScrollY float64

	// Payload contains the data that was dropped in case of
	// a MouseEventTypeDrop event.
	Payload interface{}
}

// String returns a string representation of this event.
func (e MouseEvent) String() string {
	return fmt.Sprintf("(%d,(%d,%d),%s,%s,%.2f,%.2f)",
		e.Index,
		e.X,
		e.Y,
		e.Type,
		e.Button,
		e.ScrollX,
		e.ScrollY,
	)
}

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

	// MouseEventTypeDrop indicates that some content was dropped
	// within the receiver.
	MouseEventTypeDrop

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

	// MouseEventTypeScroll indicates that the mouse wheel
	// was scrolled. The X and Y values of the event indicate the
	// offset.
	MouseEventTypeScroll
)

// MouseEventType represents the type of mouse event.
type MouseEventType int

// String returns a string representation of this event type.
func (t MouseEventType) String() string {
	switch t {
	case MouseEventTypeDown:
		return "DOWN"
	case MouseEventTypeUp:
		return "UP"
	case MouseEventTypeMove:
		return "MOVE"
	case MouseEventTypeDrag:
		return "DRAG"
	case MouseEventTypeDragCancel:
		return "DRAGCANCEL"
	case MouseEventTypeEnter:
		return "ENTER"
	case MouseEventTypeLeave:
		return "LEAVE"
	default:
		return "UNKNOWN"
	}
}

const (
	// MouseButtonLeft specifies the left mouse button.
	MouseButtonLeft = 1 + iota

	// MouseButtonMiddle specifies the middle mouse button.
	MouseButtonMiddle

	// MouseButtonRight specifies the right mouse button.
	MouseButtonRight
)

// MouseButton represents the mouse button.
type MouseButton int

// String returns a string representation of this button.
func (b MouseButton) String() string {
	switch b {
	case MouseButtonLeft:
		return "LEFT"
	case MouseButtonMiddle:
		return "MIDDLE"
	case MouseButtonRight:
		return "RIGHT"
	default:
		return "UNKNOWN"
	}
}

// FilepathPayload is a type of Payload that occurs when files
// have been dragged and dropped into the window.
type FilepathPayload struct {

	// Paths contains file paths to the dropped resources.
	Paths []string
}
