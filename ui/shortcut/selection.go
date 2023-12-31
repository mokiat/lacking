package shortcut

import (
	"github.com/mokiat/lacking/app"
	"github.com/mokiat/lacking/ui"
)

func IsSelectAll(os app.OS, event ui.KeyboardEvent) bool {
	switch os {
	case app.OSDarwin:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierSuper), ui.KeyCodeA,
		)
	case app.OSWindows:
		fallthrough
	case app.OSLinux:
		fallthrough
	default:
		return IsKeyCombo(event,
			ui.KeyModifiers(ui.KeyModifierControl), ui.KeyCodeA,
		)
	}
}
