package app

// Window represents a native application window.
//
// All methods must be invoked on the UI thread (UI goroutine),
// unless specified otherwise.
type Window interface {

	// Title returns this window's title.
	Title() string

	// SetTitle changes the title of this window.
	SetTitle(title string)

	// Size returns the content area of this window.
	Size() (int, int)

	// SetSize changes the content area of this window.
	SetSize(width, height int)

	// GamepadState returns the state of the specified gamepad and
	// whether a gamepad at the specified index is at all connected.
	GamepadState(index int) (GamepadState, bool)

	// Schedule queues a function to be called on the main thread
	// when possible. There are no guarantees that that will necessarily
	// be on the next frame iteration.
	Schedule(fn func() error)

	// Invalidate causes this window to be redrawn.
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

	// Close disposes of this window.
	Close()
}
