package standard

import "github.com/mokiat/lacking/ui"

// ProgressBar in an inidcator Control that shows the
// amount of progress that has been made on a certain
// activity.
type ProgressBar interface {
	ui.Control

	// MinProgress returns the minimum progress that can
	// be made.
	MinProgress() int

	// SetMinProgress sets the minimum amount of progress
	// that can be made.
	SetMinProgress(min int)

	// MaxProgress returns the maximum amount of progress
	// that can be made.
	MaxProgress() int

	// SetMaxProgress sets the maximum amount of progress
	// that can be made.
	SetMaxProgress(max int)

	// CurrentProgress returns the progress that has been
	// made. The value will be between MinProgress and
	// MaxProgress inclusively.
	CurrentProgress() int

	// SetCurrentProgress sets the amount of progress that
	// has been made. The value is clipped between
	// MinProgress and MaxProgress inclusively.
	SetCurrentProgress(value int)
}
