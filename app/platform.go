package app

// Platform contains information on the system that is running the app.
type Platform interface {

	// Environment returns the current environment.
	Environment() Environment

	// OS returns the current operating system.
	OS() OS
}

// Environment represents the environment in which the app runs.
type Environment string

const (
	// EnvironmentNative indicates a native app binary.
	EnvironmentNative Environment = "NATIVE"

	// EnvironmentBrowser indicates a browser page.
	EnvironmentBrowser Environment = "BROWSER"
)

// OS represents the operating system that runs the application
// regarless if native or via an intermediary like a browser.
type OS string

const (
	// OSLinux indicates the Linux operating system.
	OSLinux OS = "LINUX"

	// OSDarwin indicates the MacOS operating system.
	OSDarwin OS = "DARWIN"

	// OSWindows indicates the Windows operating system.
	OSWindows OS = "WINDOWS"

	// OSUnknown indicates that the operating system could not be determined.
	OSUnknown OS = "UNKNOWN"
)
