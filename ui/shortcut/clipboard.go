package shortcut

import (
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/ui"
)

func IsCut(os app.OS, event ui.KeyboardEvent) bool {
	switch os {
	case app.OSDarwin:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierSuper), ui.KeyCodeX,
		)
	case app.OSWindows:
		fallthrough
	case app.OSLinux:
		fallthrough
	default:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierControl), ui.KeyCodeX,
		)
	}
}

func IsCopy(os app.OS, event ui.KeyboardEvent) bool {
	switch os {
	case app.OSDarwin:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierSuper), ui.KeyCodeC,
		)
	case app.OSWindows:
		fallthrough
	case app.OSLinux:
		fallthrough
	default:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierControl), ui.KeyCodeC,
		)
	}
}

func IsPaste(os app.OS, event ui.KeyboardEvent) bool {
	switch os {
	case app.OSDarwin:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierSuper), ui.KeyCodeV,
		)
	case app.OSWindows:
		fallthrough
	case app.OSLinux:
		fallthrough
	default:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierControl), ui.KeyCodeV,
		)
	}
}
