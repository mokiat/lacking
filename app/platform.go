package app

// Platform contains information on the system that is running the app.
type Platform interface {

	// OS returns the current operating system.
	OS() OS
}

// OS represents the operating system that runs the application
// regarless if native or via an intermediary like a browser.
type OS string

const (
	// OSLinux indicates the Linux operating system.
	OSLinux OS = "linux"

	// OSDarwin indicates the MacOS operating system.
	OSDarwin OS = "darwin"

	// OSWindows indicates the Windows operating system.
	OSWindows OS = "window"

	// OSUnknown indicates that the operating system could not be determined.
	OSUnknown OS = "unknown"
)
