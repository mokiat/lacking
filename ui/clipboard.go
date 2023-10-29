package ui

// ClipboardEvent indicates an event related to a clipboard action.
type ClipboardEvent struct {

	// Action represents the type of clipboard event.
	Action ClipboardAction

	// Text contains the text stored in the clipboard in case of a Paste action.
	Text string
}

const (
	ClipboardActionCut ClipboardAction = 1 + iota
	ClipboardActionCopy
	ClipboardActionPaste
)

// ClipboardAction indicates the type of clipboard operation.
type ClipboardAction int

// String returns a string representation of this action.
func (a ClipboardAction) String() string {
	switch a {
	case ClipboardActionCut:
		return "CUT"
	case ClipboardActionCopy:
		return "COPY"
	case ClipboardActionPaste:
		return "PASTE"
	default:
		return "UNKNOWN"
	}
}
