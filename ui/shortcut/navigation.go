package shortcut

import (
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/ui"
)

func IsJumpToLineStart(os app.OS, event ui.KeyboardEvent) bool {
	event.Modifiers = event.Modifiers & ^ui.KeyModifierSet(ui.KeyModifierShift) // remove shift
	switch os {
	case app.OSDarwin:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierSuper), ui.KeyCodeArrowLeft,
		) || IsKeyCombo(event,
			ui.KeyModifiers(), ui.KeyCodeHome,
		)
	case app.OSWindows:
		fallthrough
	case app.OSLinux:
		fallthrough
	default:
		return IsKeyCombo(event,
			ui.KeyModifiers(), ui.KeyCodeHome,
		)
	}
}

func IsJumpToLineEnd(os app.OS, event ui.KeyboardEvent) bool {
	event.Modifiers = event.Modifiers & ^ui.KeyModifierSet(ui.KeyModifierShift) // remove shift
	switch os {
	case app.OSDarwin:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierSuper), ui.KeyCodeArrowRight,
		) || IsKeyCombo(event,
			ui.KeyModifiers(), ui.KeyCodeEnd,
		)
	case app.OSWindows:
		fallthrough
	case app.OSLinux:
		fallthrough
	default:
		return IsKeyCombo(event,
			ui.KeyModifiers(), ui.KeyCodeEnd,
		)
	}
}

func IsJumpToDocumentStart(os app.OS, event ui.KeyboardEvent) bool {
	event.Modifiers = event.Modifiers & ^ui.KeyModifierSet(ui.KeyModifierShift) // remove shift
	switch os {
	case app.OSDarwin:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierSuper), ui.KeyCodeArrowUp,
		)
	case app.OSWindows:
		fallthrough
	case app.OSLinux:
		fallthrough
	default:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierControl), ui.KeyCodeHome,
		)
	}
}

func IsJumpToDocumentEnd(os app.OS, event ui.KeyboardEvent) bool {
	event.Modifiers = event.Modifiers & ^ui.KeyModifierSet(ui.KeyModifierShift) // remove shift
	switch os {
	case app.OSDarwin:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierSuper), ui.KeyCodeArrowDown,
		)
	case app.OSWindows:
		fallthrough
	case app.OSLinux:
		fallthrough
	default:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierControl), ui.KeyCodeEnd,
		)
	}
}
