package app

import "fmt"

// MouseEvent represents an event related to a mouse action.
type MouseEvent struct {

	// Index indicates which mouse triggered the event. By default
	// the index for a the primary mouse is 0.
	//
	// This is applicable for devices with multiple pointers
	// (mobile) or in case a second mouse is emulated
	// (e.g. with a game controller).
	Index int

	// Action indicates the action performed with the mouse.
	Action MouseAction

	// Button specifies the button for which the event is applicable.
	Button MouseButton

	// X specifies the horizontal position of the mouse.
	X int

	// Y specifies the vertical position of the mouse.
	Y int

	// ScrollX determines the amount of horizontal scroll.
	ScrollX int

	// ScrollY determines the amount of vertical scroll.
	ScrollY int

	// Payload contains any external data associated with the event.
	Payload any
}

// String returns a string representation of this event.
func (e MouseEvent) String() string {
	return fmt.Sprintf("(%d,%s,%s,(%d,%d),(%d,%d))",
		e.Index,
		e.Action,
		e.Button,
		e.X,
		e.Y,
		e.ScrollX,
		e.ScrollY,
	)
}

const (
	// MouseActionDown indicates that a mouse button was pressed down.
	MouseActionDown MouseAction = 1 + iota

	// MouseActionUp indicates that a mouse button was released.
	MouseActionUp

	// MouseActionMove indicates that the mouse was moved.
	MouseActionMove

	// MouseActionEnter indicates that the mouse has entered the window.
	MouseActionEnter

	// MouseActionLeave indicates that the mouse has left the window.
	//
	// If the mouse was being dragged, further events may be sent from outside
	// the window boundary.
	MouseActionLeave

	// MouseActionDrop indicates that some payload was dropped onto the window.
	MouseActionDrop

	// MouseActionScroll indicates that the mouse wheel was scrolled.
	//
	// The ScrollX and ScrollY values of the event indicate the offset in
	// abstract units (comparable to pixels in magnitude).
	MouseActionScroll
)

// MouseAction represents the type of action performed with the mouse.
type MouseAction int

// String returns a string representation of this event type.
func (a MouseAction) String() string {
	switch a {
	case MouseActionDown:
		return "DOWN"
	case MouseActionUp:
		return "UP"
	case MouseActionMove:
		return "MOVE"
	case MouseActionEnter:
		return "ENTER"
	case MouseActionLeave:
		return "LEAVE"
	case MouseActionDrop:
		return "DROP"
	case MouseActionScroll:
		return "SCROLL"
	default:
		return "UNKNOWN"
	}
}

const (
	// MouseButtonLeft specifies the left mouse button.
	MouseButtonLeft MouseButton = 1 + iota

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
// have been dragged and dropped onto the window.
type FilepathPayload struct {

	// Paths contains file paths to the dropped resources.
	Paths []string
}
