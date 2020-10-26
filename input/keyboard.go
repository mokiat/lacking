package input

import (
	"sync"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type Key int

const (
	KeyEscape = Key(glfw.KeyEscape)
	KeyUp     = Key(glfw.KeyUp)
	KeyDown   = Key(glfw.KeyDown)
	KeyLeft   = Key(glfw.KeyLeft)
	KeyRight  = Key(glfw.KeyRight)
	KeyEnter  = Key(glfw.KeyEnter)
	KeyA      = Key(glfw.KeyA)
	KeyB      = Key(glfw.KeyB)
	KeyC      = Key(glfw.KeyC)
	KeyD      = Key(glfw.KeyD)
	KeyE      = Key(glfw.KeyE)
	KeyF      = Key(glfw.KeyF)
)

const keyCount = 512

func NewKeyboardRecorder(window *glfw.Window) *KeyboardRecorder {
	recorder := &KeyboardRecorder{
		recordKeyboard: NewKeyboard(),
		freeKeyboard:   NewKeyboard(),
	}
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		recorder.keyboardMU.Lock()
		defer recorder.keyboardMU.Unlock()

		inKey := Key(key)
		if inKey < 0 || inKey >= keyCount {
			return
		}
		switch action {
		case glfw.Press:
			recorder.recordKeyboard.pressedKeys[inKey] = true
		case glfw.Release:
			recorder.recordKeyboard.pressedKeys[inKey] = false
		}
	})
	return recorder
}

type KeyboardRecorder struct {
	keyboardMU     sync.Mutex
	recordKeyboard *Keyboard
	freeKeyboard   *Keyboard
}

func (r *KeyboardRecorder) Record() {
}

func (r *KeyboardRecorder) Fetch() *Keyboard {
	r.keyboardMU.Lock()
	defer r.keyboardMU.Unlock()

	returnKeyboard := r.recordKeyboard
	r.freeKeyboard.synchronize(r.recordKeyboard)
	r.recordKeyboard = r.freeKeyboard
	r.freeKeyboard = nil
	return returnKeyboard
}

func (r *KeyboardRecorder) Release(keyboard *Keyboard) {
	r.keyboardMU.Lock()
	defer r.keyboardMU.Unlock()
	r.freeKeyboard = keyboard
}

func NewKeyboard() *Keyboard {
	return &Keyboard{}
}

type Keyboard struct {
	pressedKeys [keyCount]bool
}

func (k *Keyboard) IsPressed(key Key) bool {
	if key < 0 || key >= keyCount {
		return false
	}
	return k.pressedKeys[key]
}

func (k *Keyboard) synchronize(other *Keyboard) {
	*k = *other
}
