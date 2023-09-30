package app

import (
	"github.com/mokiat/lacking/audio"
	"github.com/mokiat/lacking/render"
)

// Window represents a native application window.
//
// All methods must be invoked on the UI thread (UI goroutine),
// unless otherwise specified.
type Window interface {

	// Platform returns the information on the platform that is running the app.
	Platform() Platform

	// Title returns this window's title.
	Title() string

	// SetTitle changes the title of this window.
	SetTitle(title string)

	// Size returns the content area of this window.
	Size() (int, int)

	// SetSize changes the content area of this window.
	SetSize(width, height int)

	// FramebufferSize returns the texel size of the underlying framebuffer.
	// This would normally be used as reference when using graphics libraries
	// to draw on the screen as the framebuffer size may differ from the
	// window size.
	FramebufferSize() (int, int)

	// Gamepads returns an array of potentially available gamepad
	// controllers.
	//
	// Use the Connected and Supported methods on the Gamepad object to check
	// whether it is connected and can be used.
	//
	// This API supports up to 4 connected devices.
	Gamepads() [4]Gamepad

	// Schedule queues a function to be called on the main thread
	// when possible. There are no guarantees that it will necessarily
	// be on the next frame iteration.
	Schedule(fn func())

	// Invalidate causes this window to be redrawn when possible.
	Invalidate()

	// CreateCursor creates a new cursor object based on the specified
	// definition.
	CreateCursor(definition CursorDefinition) Cursor

	// UseCursor changes the currently displayed cursor on the screen.
	// Specifying nil returns the default cursor.
	UseCursor(cursor Cursor)

	// CursorVisible returns whether a cursor is to be displayed on the
	// screen. This is determined based on the visibility and lock settings
	// of the cursor
	CursorVisible() bool

	// SetCursorVisible changes whether a cursor is displayed on the
	// screen.
	SetCursorVisible(visible bool)

	// SetCursorLocked traps the cursor within the boundaries of the window
	// and reports relative motion events. This method also hides the cursor.
	SetCursorLocked(locked bool)

	// RenderAPI provides access to a usable Render API based on the current
	// window's screen.
	//
	// If the implementation of this API does not support graphics, then this
	// method returns nil.
	RenderAPI() render.API

	// AudioAPI provides access to a usable Audio API based on the current
	// window.
	//
	// If the implementation of this API does not support graphics, then this
	// method returns nil.
	AudioAPI() audio.API

	// Close disposes of this window.
	Close()
}
