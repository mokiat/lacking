package preset

import "github.com/mokiat/lacking/app"

type GamepadProvider interface {
	Gamepads() [4]app.Gamepad
}
