package ui

import (
	"fmt"

	"github.com/mokiat/lacking/app"
)

// MouseEvent represents an event related to a mouse action.
type MouseEvent struct {

	// Index indicates which mouse triggered the event. By default
	// the index for a the primary mouse is 0.
	// This is applicable for devices with multiple pointers
	// (mobile) or in case a second mouse is emulated
	// (e.g. with a game controller).
	Index int

	// Position specifies the moust position relative to the receiver.
	Position Position

	// Type specifies the mouse event type.
	Type MouseEventType

	// Button specifies the button for which the event is
	// applicable.
	Button MouseButton

	// ScrollX determines the amount of horizontal scroll.
	ScrollX float64

	// ScrollY determines the amount of vertical scroll.
	ScrollY float64

	// Payload contains the data that was dropped.
	Payload interface{}
}

// String returns a string representation for this mouse event.
func (e MouseEvent) String() string {
	return fmt.Sprintf("(%d,%s,%s,%s)",
		e.Index,
		e.Position,
		e.Type,
		e.Button,
	)
}

// MouseEventType represents the type of mouse event.
type MouseEventType = app.MouseEventType

const (
	MouseEventTypeDown       = app.MouseEventTypeDown
	MouseEventTypeUp         = app.MouseEventTypeUp
	MouseEventTypeMove       = app.MouseEventTypeMove
	MouseEventTypeDrag       = app.MouseEventTypeDrag
	MouseEventTypeDragCancel = app.MouseEventTypeDragCancel
	MouseEventTypeDrop       = app.MouseEventTypeDrop
	MouseEventTypeEnter      = app.MouseEventTypeEnter
	MouseEventTypeLeave      = app.MouseEventTypeLeave
	MouseEventTypeScroll     = app.MouseEventTypeScroll
)

// MouseButton represents the mouse button.
type MouseButton = app.MouseButton

const (
	MouseButtonLeft   = app.MouseButtonLeft
	MouseButtonMiddle = app.MouseButtonMiddle
	MouseButtonRight  = app.MouseButtonRight
)

// FilepathPayload is a type of Payload that occurs when files
// have been dragged and dropped into the window.
type FilepathPayload = app.FilepathPayload
