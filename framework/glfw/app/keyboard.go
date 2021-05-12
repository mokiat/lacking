package app

import (
	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/mokiat/lacking/app"
)

var (
	keyboardEventTypeMapping map[glfw.Action]app.KeyboardEventType
	keyboardKeyMapping       map[glfw.Key]app.KeyCode
)

func init() {
	keyboardEventTypeMapping = make(map[glfw.Action]app.KeyboardEventType)
	keyboardEventTypeMapping[glfw.Press] = app.KeyboardEventTypeKeyDown
	keyboardEventTypeMapping[glfw.Release] = app.KeyboardEventTypeKeyUp
	keyboardEventTypeMapping[glfw.Repeat] = app.KeyboardEventTypeRepeat

	keyboardKeyMapping = make(map[glfw.Key]app.KeyCode)
	keyboardKeyMapping[glfw.KeyEscape] = app.KeyCodeEscape
	keyboardKeyMapping[glfw.KeyEnter] = app.KeyCodeEnter
	keyboardKeyMapping[glfw.KeySpace] = app.KeyCodeSpace
	keyboardKeyMapping[glfw.KeyTab] = app.KeyCodeTab
	keyboardKeyMapping[glfw.KeyCapsLock] = app.KeyCodeCaps
	keyboardKeyMapping[glfw.KeyLeftShift] = app.KeyCodeLeftShift
	keyboardKeyMapping[glfw.KeyRightShift] = app.KeyCodeRightShift
	keyboardKeyMapping[glfw.KeyLeftControl] = app.KeyCodeLeftControl
	keyboardKeyMapping[glfw.KeyRightControl] = app.KeyCodeRightControl
	keyboardKeyMapping[glfw.KeyLeftAlt] = app.KeyCodeLeftAlt
	keyboardKeyMapping[glfw.KeyRightAlt] = app.KeyCodeRightAlt
	keyboardKeyMapping[glfw.KeyBackspace] = app.KeyCodeBackspace
	keyboardKeyMapping[glfw.KeyInsert] = app.KeyCodeInsert
	keyboardKeyMapping[glfw.KeyDelete] = app.KeyCodeDelete
	keyboardKeyMapping[glfw.KeyHome] = app.KeyCodeHome
	keyboardKeyMapping[glfw.KeyEnd] = app.KeyCodeEnd
	keyboardKeyMapping[glfw.KeyPageUp] = app.KeyCodePageUp
	keyboardKeyMapping[glfw.KeyPageDown] = app.KeyCodePageDown
	keyboardKeyMapping[glfw.KeyLeft] = app.KeyCodeArrowLeft
	keyboardKeyMapping[glfw.KeyRight] = app.KeyCodeArrowRight
	keyboardKeyMapping[glfw.KeyUp] = app.KeyCodeArrowUp
	keyboardKeyMapping[glfw.KeyDown] = app.KeyCodeArrowDown
	keyboardKeyMapping[glfw.KeyMinus] = app.KeyCodeMinus
	keyboardKeyMapping[glfw.KeyEqual] = app.KeyCodeEqual
	keyboardKeyMapping[glfw.KeyLeftBracket] = app.KeyCodeLeftBracket
	keyboardKeyMapping[glfw.KeyRightBracket] = app.KeyCodeRightBracket
	keyboardKeyMapping[glfw.KeySemicolon] = app.KeyCodeSemicolon
	keyboardKeyMapping[glfw.KeyComma] = app.KeyCodeComma
	keyboardKeyMapping[glfw.KeyPeriod] = app.KeyCodePeriod
	keyboardKeyMapping[glfw.KeySlash] = app.KeyCodeSlash
	keyboardKeyMapping[glfw.KeyBackslash] = app.KeyCodeBackslash
	keyboardKeyMapping[glfw.KeyApostrophe] = app.KeyCodeApostrophe
	keyboardKeyMapping[glfw.KeyGraveAccent] = app.KeyCodeGraveAccent
	keyboardKeyMapping[glfw.KeyA] = app.KeyCodeA
	keyboardKeyMapping[glfw.KeyB] = app.KeyCodeB
	keyboardKeyMapping[glfw.KeyC] = app.KeyCodeC
	keyboardKeyMapping[glfw.KeyD] = app.KeyCodeD
	keyboardKeyMapping[glfw.KeyE] = app.KeyCodeE
	keyboardKeyMapping[glfw.KeyF] = app.KeyCodeF
	keyboardKeyMapping[glfw.KeyG] = app.KeyCodeG
	keyboardKeyMapping[glfw.KeyH] = app.KeyCodeH
	keyboardKeyMapping[glfw.KeyI] = app.KeyCodeI
	keyboardKeyMapping[glfw.KeyJ] = app.KeyCodeJ
	keyboardKeyMapping[glfw.KeyK] = app.KeyCodeK
	keyboardKeyMapping[glfw.KeyL] = app.KeyCodeL
	keyboardKeyMapping[glfw.KeyM] = app.KeyCodeM
	keyboardKeyMapping[glfw.KeyN] = app.KeyCodeN
	keyboardKeyMapping[glfw.KeyO] = app.KeyCodeO
	keyboardKeyMapping[glfw.KeyP] = app.KeyCodeP
	keyboardKeyMapping[glfw.KeyQ] = app.KeyCodeQ
	keyboardKeyMapping[glfw.KeyR] = app.KeyCodeR
	keyboardKeyMapping[glfw.KeyS] = app.KeyCodeS
	keyboardKeyMapping[glfw.KeyT] = app.KeyCodeT
	keyboardKeyMapping[glfw.KeyU] = app.KeyCodeU
	keyboardKeyMapping[glfw.KeyV] = app.KeyCodeV
	keyboardKeyMapping[glfw.KeyW] = app.KeyCodeW
	keyboardKeyMapping[glfw.KeyX] = app.KeyCodeX
	keyboardKeyMapping[glfw.KeyY] = app.KeyCodeY
	keyboardKeyMapping[glfw.KeyZ] = app.KeyCodeZ
	keyboardKeyMapping[glfw.Key0] = app.KeyCode0
	keyboardKeyMapping[glfw.Key1] = app.KeyCode1
	keyboardKeyMapping[glfw.Key2] = app.KeyCode2
	keyboardKeyMapping[glfw.Key3] = app.KeyCode3
	keyboardKeyMapping[glfw.Key4] = app.KeyCode4
	keyboardKeyMapping[glfw.Key5] = app.KeyCode5
	keyboardKeyMapping[glfw.Key6] = app.KeyCode6
	keyboardKeyMapping[glfw.Key7] = app.KeyCode7
	keyboardKeyMapping[glfw.Key8] = app.KeyCode8
	keyboardKeyMapping[glfw.Key9] = app.KeyCode9
	keyboardKeyMapping[glfw.KeyF1] = app.KeyCodeF1
	keyboardKeyMapping[glfw.KeyF2] = app.KeyCodeF2
	keyboardKeyMapping[glfw.KeyF3] = app.KeyCodeF3
	keyboardKeyMapping[glfw.KeyF4] = app.KeyCodeF4
	keyboardKeyMapping[glfw.KeyF5] = app.KeyCodeF5
	keyboardKeyMapping[glfw.KeyF6] = app.KeyCodeF6
	keyboardKeyMapping[glfw.KeyF7] = app.KeyCodeF7
	keyboardKeyMapping[glfw.KeyF8] = app.KeyCodeF8
	keyboardKeyMapping[glfw.KeyF9] = app.KeyCodeF9
	keyboardKeyMapping[glfw.KeyF10] = app.KeyCodeF10
	keyboardKeyMapping[glfw.KeyF11] = app.KeyCodeF11
	keyboardKeyMapping[glfw.KeyF12] = app.KeyCodeF12
}
