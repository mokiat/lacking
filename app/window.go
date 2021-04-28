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

	// Close disposes of this window.
	Close()
}
