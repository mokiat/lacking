package glfw

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/mokiat/lacking/ui"
)

var (
	keyMapping  map[glfw.Key]ui.KeyCode
	typeMapping map[glfw.Action]ui.KeyboardEventType
)

func init() {
	keyMapping = make(map[glfw.Key]ui.KeyCode)
	keyMapping[glfw.KeyEscape] = ui.KeyCodeEscape
	keyMapping[glfw.KeyEnter] = ui.KeyCodeEnter
	keyMapping[glfw.KeySpace] = ui.KeyCodeSpace
	keyMapping[glfw.KeyTab] = ui.KeyCodeTab
	keyMapping[glfw.KeyCapsLock] = ui.KeyCodeCaps
	keyMapping[glfw.KeyLeftShift] = ui.KeyCodeLeftShift
	keyMapping[glfw.KeyRightShift] = ui.KeyCodeRightShift
	keyMapping[glfw.KeyLeftControl] = ui.KeyCodeLeftControl
	keyMapping[glfw.KeyRightControl] = ui.KeyCodeRightControl
	keyMapping[glfw.KeyLeftAlt] = ui.KeyCodeLeftAlt
	keyMapping[glfw.KeyRightAlt] = ui.KeyCodeRightAlt
	keyMapping[glfw.KeyBackspace] = ui.KeyCodeBackspace
	keyMapping[glfw.KeyInsert] = ui.KeyCodeInsert
	keyMapping[glfw.KeyDelete] = ui.KeyCodeDelete
	keyMapping[glfw.KeyHome] = ui.KeyCodeHome
	keyMapping[glfw.KeyEnd] = ui.KeyCodeEnd
	keyMapping[glfw.KeyPageUp] = ui.KeyCodePageUp
	keyMapping[glfw.KeyPageDown] = ui.KeyCodePageDown
	keyMapping[glfw.KeyLeft] = ui.KeyCodeArrowLeft
	keyMapping[glfw.KeyRight] = ui.KeyCodeArrowRight
	keyMapping[glfw.KeyUp] = ui.KeyCodeArrowUp
	keyMapping[glfw.KeyDown] = ui.KeyCodeArrowDown
	keyMapping[glfw.KeyMinus] = ui.KeyCodeMinus
	keyMapping[glfw.KeyEqual] = ui.KeyCodeEqual
	keyMapping[glfw.KeyLeftBracket] = ui.KeyCodeLeftBracket
	keyMapping[glfw.KeyRightBracket] = ui.KeyCodeRightBracket
	keyMapping[glfw.KeySemicolon] = ui.KeyCodeSemicolon
	keyMapping[glfw.KeyComma] = ui.KeyCodeComma
	keyMapping[glfw.KeyPeriod] = ui.KeyCodePeriod
	keyMapping[glfw.KeySlash] = ui.KeyCodeSlash
	keyMapping[glfw.KeyBackslash] = ui.KeyCodeBackslash
	keyMapping[glfw.KeyApostrophe] = ui.KeyCodeApostrophe
	keyMapping[glfw.KeyGraveAccent] = ui.KeyCodeGraveAccent
	keyMapping[glfw.KeyA] = ui.KeyCodeA
	keyMapping[glfw.KeyB] = ui.KeyCodeB
	keyMapping[glfw.KeyC] = ui.KeyCodeC
	keyMapping[glfw.KeyD] = ui.KeyCodeD
	keyMapping[glfw.KeyE] = ui.KeyCodeE
	keyMapping[glfw.KeyF] = ui.KeyCodeF
	keyMapping[glfw.KeyG] = ui.KeyCodeG
	keyMapping[glfw.KeyH] = ui.KeyCodeH
	keyMapping[glfw.KeyI] = ui.KeyCodeI
	keyMapping[glfw.KeyJ] = ui.KeyCodeJ
	keyMapping[glfw.KeyK] = ui.KeyCodeK
	keyMapping[glfw.KeyL] = ui.KeyCodeL
	keyMapping[glfw.KeyM] = ui.KeyCodeM
	keyMapping[glfw.KeyN] = ui.KeyCodeN
	keyMapping[glfw.KeyO] = ui.KeyCodeO
	keyMapping[glfw.KeyP] = ui.KeyCodeP
	keyMapping[glfw.KeyQ] = ui.KeyCodeQ
	keyMapping[glfw.KeyR] = ui.KeyCodeR
	keyMapping[glfw.KeyS] = ui.KeyCodeS
	keyMapping[glfw.KeyT] = ui.KeyCodeT
	keyMapping[glfw.KeyU] = ui.KeyCodeU
	keyMapping[glfw.KeyV] = ui.KeyCodeV
	keyMapping[glfw.KeyW] = ui.KeyCodeW
	keyMapping[glfw.KeyX] = ui.KeyCodeX
	keyMapping[glfw.KeyY] = ui.KeyCodeY
	keyMapping[glfw.KeyZ] = ui.KeyCodeZ
	keyMapping[glfw.Key0] = ui.KeyCode0
	keyMapping[glfw.Key1] = ui.KeyCode1
	keyMapping[glfw.Key2] = ui.KeyCode2
	keyMapping[glfw.Key3] = ui.KeyCode3
	keyMapping[glfw.Key4] = ui.KeyCode4
	keyMapping[glfw.Key5] = ui.KeyCode5
	keyMapping[glfw.Key6] = ui.KeyCode6
	keyMapping[glfw.Key7] = ui.KeyCode7
	keyMapping[glfw.Key8] = ui.KeyCode8
	keyMapping[glfw.Key9] = ui.KeyCode9
	keyMapping[glfw.KeyF1] = ui.KeyCodeF1
	keyMapping[glfw.KeyF2] = ui.KeyCodeF2
	keyMapping[glfw.KeyF3] = ui.KeyCodeF3
	keyMapping[glfw.KeyF4] = ui.KeyCodeF4
	keyMapping[glfw.KeyF5] = ui.KeyCodeF5
	keyMapping[glfw.KeyF6] = ui.KeyCodeF6
	keyMapping[glfw.KeyF7] = ui.KeyCodeF7
	keyMapping[glfw.KeyF8] = ui.KeyCodeF8
	keyMapping[glfw.KeyF9] = ui.KeyCodeF9
	keyMapping[glfw.KeyF10] = ui.KeyCodeF10
	keyMapping[glfw.KeyF11] = ui.KeyCodeF11
	keyMapping[glfw.KeyF12] = ui.KeyCodeF12

	typeMapping = make(map[glfw.Action]ui.KeyboardEventType)
	typeMapping[glfw.Press] = ui.KeyboardEventTypeKeyDown
	typeMapping[glfw.Release] = ui.KeyboardEventTypeKeyUp
	typeMapping[glfw.Repeat] = ui.KeyboardEventTypeRepeat
}
