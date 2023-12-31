package shortcut

import (
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/ui"
)

func IsClose(os app.OS, event ui.KeyboardEvent) bool {
	switch os {
	case app.OSDarwin:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierSuper), ui.KeyCodeW,
		)
	case app.OSWindows:
		fallthrough
	case app.OSLinux:
		fallthrough
	default:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierControl), ui.KeyCodeW,
		)
	}
}

func IsSave(os app.OS, event ui.KeyboardEvent) bool {
	switch os {
	case app.OSDarwin:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierSuper), ui.KeyCodeS,
		)
	case app.OSWindows:
		fallthrough
	case app.OSLinux:
		fallthrough
	default:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierControl), ui.KeyCodeS,
		)
	}
}

func IsUndo(os app.OS, event ui.KeyboardEvent) bool {
	switch os {
	case app.OSDarwin:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierSuper), ui.KeyCodeZ,
		)
	case app.OSWindows:
		fallthrough
	case app.OSLinux:
		fallthrough
	default:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierControl), ui.KeyCodeZ,
		)
	}
}

func IsRedo(os app.OS, event ui.KeyboardEvent) bool {
	switch os {
	case app.OSDarwin:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierSuper, ui.KeyModifierShift), ui.KeyCodeZ,
		)
	case app.OSWindows:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierControl), ui.KeyCodeY,
		)
	case app.OSLinux:
		fallthrough
	default:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierControl, ui.KeyModifierShift), ui.KeyCodeZ,
		)
	}
}

func IsKeyCombo(event ui.KeyboardEvent, modifiers ui.KeyModifierSet, code ui.KeyCode) bool {
	return event.Modifiers == modifiers && event.Code == code
}
