package input

import (
	"sync"

	"github.com/go-gl/glfw/v3.3/glfw"
)

func NewGamepadRecorder() *GamepadRecorder {
	return &GamepadRecorder{
		recordGamepad: NewGamepad(),
		freeGamepad:   NewGamepad(),
	}
}

type GamepadRecorder struct {
	gamepadMU     sync.Mutex
	recordGamepad *Gamepad
	freeGamepad   *Gamepad
}

func (r *GamepadRecorder) Record() {
	r.gamepadMU.Lock()
	defer r.gamepadMU.Unlock()

	gamepad := r.recordGamepad
	gamepad.Available = glfw.Joystick1.Present() && glfw.Joystick1.IsGamepad()
	if gamepad.Available {
		state := glfw.Joystick1.GetGamepadState()
		gamepad.LeftStickX = state.Axes[glfw.AxisLeftX]
		gamepad.LeftStickY = -state.Axes[glfw.AxisLeftY]
		gamepad.RightStickX = state.Axes[glfw.AxisRightX]
		gamepad.RightStickY = -state.Axes[glfw.AxisRightY]
		gamepad.LeftTrigger = (state.Axes[glfw.AxisLeftTrigger] + 1.0) / 2.0
		gamepad.RightTrigger = (state.Axes[glfw.AxisRightTrigger] + 1.0) / 2.0
	}
}

func (r *GamepadRecorder) Fetch() *Gamepad {
	r.gamepadMU.Lock()
	defer r.gamepadMU.Unlock()

	returnGamepad := r.recordGamepad
	r.freeGamepad.synchronize(r.recordGamepad)
	r.recordGamepad = r.freeGamepad
	r.freeGamepad = nil
	return returnGamepad
}

func (r *GamepadRecorder) Release(gamepad *Gamepad) {
	r.gamepadMU.Lock()
	defer r.gamepadMU.Unlock()
	r.freeGamepad = gamepad
}

func NewGamepad() *Gamepad {
	return &Gamepad{}
}

type Gamepad struct {
	Available    bool
	LeftStickX   float32
	LeftStickY   float32
	RightStickX  float32
	RightStickY  float32
	LeftTrigger  float32
	RightTrigger float32
}

func (g *Gamepad) synchronize(other *Gamepad) {
	*g = *other
}
