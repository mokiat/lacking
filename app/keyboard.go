package app

import (
	"fmt"
	"strings"
)

// KeyboardEvent is used to propagate events related to keyboard
// actions.
type KeyboardEvent struct {

	// Type specifies the keyboard event type.
	Type KeyboardEventType

	// Code returns the code of the keyboard key.
	Code KeyCode

	// Rune returns the character that was typed in case
	// of an KeyboardEventTypeType event.
	Rune rune

	// Modifiers contains a set of modifier keys that were
	// pressed during the event.
	Modifiers KeyModifierSet
}

const (
	// KeyboardEventTypeKeyDown indicates that a keyboard key
	// was pressed.
	KeyboardEventTypeKeyDown KeyboardEventType = 1 + iota

	// KeyboardEventTypeKeyUp indicates that a keyboard key was
	// released.
	KeyboardEventTypeKeyUp

	// keyboardEventTypeType indicates that a keyboard key is
	// being pressed continuously.
	KeyboardEventTypeRepeat

	// KeyboardEventTypeType indicates that a character is typed
	// with the keyboard.
	// Such events would be duplicates to KeyboardEventTypeDown
	// or KeyboardEventTypeRepeat but allow for the character
	// rune to be read after all modifiers have been applied so that
	// one does not have to implement that on their own.
	KeyboardEventTypeType
)

// KeyboardEventType is used to specify the type of keyboard
// action that occurred.
type KeyboardEventType int

// String returns a string representation of this event type,
func (t KeyboardEventType) String() string {
	switch t {
	case KeyboardEventTypeKeyDown:
		return "DOWN"
	case KeyboardEventTypeKeyUp:
		return "UP"
	case KeyboardEventTypeRepeat:
		return "REPEAT"
	case KeyboardEventTypeType:
		return "TYPE"
	default:
		return "UNKNOWN"
	}
}

const (
	KeyCodeEscape KeyCode = 1 + iota
	KeyCodeEnter
	KeyCodeSpace
	KeyCodeTab
	KeyCodeCaps
	KeyCodeLeftShift
	KeyCodeRightShift
	KeyCodeLeftControl
	KeyCodeRightControl
	KeyCodeLeftAlt
	KeyCodeRightAlt
	KeyCodeBackspace
	KeyCodeInsert
	KeyCodeDelete
	KeyCodeHome
	KeyCodeEnd
	KeyCodePageUp
	KeyCodePageDown
	KeyCodeArrowLeft
	KeyCodeArrowRight
	KeyCodeArrowUp
	KeyCodeArrowDown
	KeyCodeMinus
	KeyCodeEqual
	KeyCodeLeftBracket
	KeyCodeRightBracket
	KeyCodeSemicolon
	KeyCodeComma
	KeyCodePeriod
	KeyCodeSlash
	KeyCodeBackslash
	KeyCodeApostrophe
	KeyCodeGraveAccent
	KeyCodeA
	KeyCodeB
	KeyCodeC
	KeyCodeD
	KeyCodeE
	KeyCodeF
	KeyCodeG
	KeyCodeH
	KeyCodeI
	KeyCodeJ
	KeyCodeK
	KeyCodeL
	KeyCodeM
	KeyCodeN
	KeyCodeO
	KeyCodeP
	KeyCodeQ
	KeyCodeR
	KeyCodeS
	KeyCodeT
	KeyCodeU
	KeyCodeV
	KeyCodeW
	KeyCodeX
	KeyCodeY
	KeyCodeZ
	KeyCode0
	KeyCode1
	KeyCode2
	KeyCode3
	KeyCode4
	KeyCode5
	KeyCode6
	KeyCode7
	KeyCode8
	KeyCode9
	KeyCodeF1
	KeyCodeF2
	KeyCodeF3
	KeyCodeF4
	KeyCodeF5
	KeyCodeF6
	KeyCodeF7
	KeyCodeF8
	KeyCodeF9
	KeyCodeF10
	KeyCodeF11
	KeyCodeF12
)

// KeyCode represents a keyboard key.
type KeyCode int

// String returns a string representation of this key code.
func (c KeyCode) String() string {
	switch c {
	case KeyCodeEscape:
		return "ESCAPE"
	case KeyCodeEnter:
		return "ENTER"
	case KeyCodeSpace:
		return "SPACE"
	case KeyCodeTab:
		return "TAB"
	case KeyCodeCaps:
		return "CAPS"
	case KeyCodeLeftShift:
		return "LSHIFT"
	case KeyCodeRightShift:
		return "RSHIFT"
	case KeyCodeLeftControl:
		return "LCTRL"
	case KeyCodeRightControl:
		return "RCTRL"
	case KeyCodeLeftAlt:
		return "LALT"
	case KeyCodeRightAlt:
		return "RALT"
	case KeyCodeBackspace:
		return "BACKSPACE"
	case KeyCodeInsert:
		return "INSERT"
	case KeyCodeDelete:
		return "DELETE"
	case KeyCodeHome:
		return "HOME"
	case KeyCodeEnd:
		return "END"
	case KeyCodePageUp:
		return "PGUP"
	case KeyCodePageDown:
		return "PGDOWN"
	case KeyCodeArrowLeft:
		return "LEFT"
	case KeyCodeArrowRight:
		return "RIGHT"
	case KeyCodeArrowUp:
		return "UP"
	case KeyCodeArrowDown:
		return "DOWN"
	case KeyCodeMinus:
		return "-"
	case KeyCodeEqual:
		return "="
	case KeyCodeLeftBracket:
		return "["
	case KeyCodeRightBracket:
		return "]"
	case KeyCodeSemicolon:
		return ";"
	case KeyCodeComma:
		return ","
	case KeyCodePeriod:
		return "."
	case KeyCodeSlash:
		return "/"
	case KeyCodeBackslash:
		return "\\"
	case KeyCodeApostrophe:
		return "'"
	case KeyCodeGraveAccent:
		return "`"
	case KeyCodeA:
		return "A"
	case KeyCodeB:
		return "B"
	case KeyCodeC:
		return "C"
	case KeyCodeD:
		return "D"
	case KeyCodeE:
		return "E"
	case KeyCodeF:
		return "F"
	case KeyCodeG:
		return "G"
	case KeyCodeH:
		return "H"
	case KeyCodeI:
		return "I"
	case KeyCodeJ:
		return "J"
	case KeyCodeK:
		return "K"
	case KeyCodeL:
		return "L"
	case KeyCodeM:
		return "M"
	case KeyCodeN:
		return "N"
	case KeyCodeO:
		return "O"
	case KeyCodeP:
		return "P"
	case KeyCodeQ:
		return "Q"
	case KeyCodeR:
		return "R"
	case KeyCodeS:
		return "S"
	case KeyCodeT:
		return "T"
	case KeyCodeU:
		return "U"
	case KeyCodeV:
		return "V"
	case KeyCodeW:
		return "W"
	case KeyCodeX:
		return "X"
	case KeyCodeY:
		return "Y"
	case KeyCodeZ:
		return "Z"
	case KeyCode0:
		return "0"
	case KeyCode1:
		return "1"
	case KeyCode2:
		return "2"
	case KeyCode3:
		return "3"
	case KeyCode4:
		return "4"
	case KeyCode5:
		return "5"
	case KeyCode6:
		return "6"
	case KeyCode7:
		return "7"
	case KeyCode8:
		return "8"
	case KeyCode9:
		return "9"
	case KeyCodeF1:
		return "F1"
	case KeyCodeF2:
		return "F2"
	case KeyCodeF3:
		return "F3"
	case KeyCodeF4:
		return "F4"
	case KeyCodeF5:
		return "F5"
	case KeyCodeF6:
		return "F6"
	case KeyCodeF7:
		return "F7"
	case KeyCodeF8:
		return "F8"
	case KeyCodeF9:
		return "F9"
	case KeyCodeF10:
		return "F10"
	case KeyCodeF11:
		return "F11"
	case KeyCodeF12:
		return "F12"
	default:
		return "UNKNOWN"
	}
}

const (
	KeyModifierControl KeyModifier = 1 << (iota + 1)
	KeyModifierShift
	KeyModifierAlt
	KeyModifierCapsLock
)

// KeyModifier represents a modifier key.
type KeyModifier int

// String returns a string representation of this key modifier.
func (m KeyModifier) String() string {
	switch m {
	case KeyModifierControl:
		return "CONTROL"
	case KeyModifierShift:
		return "SHIFT"
	case KeyModifierAlt:
		return "ALT"
	case KeyModifierCapsLock:
		return "CAPS"
	default:
		return "UNKNOWN"
	}
}

// KeyModifierSet is used to indicate which modifier
// keys were active at the event occurrence.
type KeyModifierSet int

// Contains returns whether the set contains the
// specified modifier
func (s KeyModifierSet) Contains(modifier KeyModifier) bool {
	return (int(s) & int(modifier)) == int(modifier)
}

// String returns a string representation of this key modifier set.
func (s KeyModifierSet) String() string {
	var descriptions []string
	if s.Contains(KeyModifierControl) {
		descriptions = append(descriptions, KeyModifierControl.String())
	}
	if s.Contains(KeyModifierShift) {
		descriptions = append(descriptions, KeyModifierShift.String())
	}
	if s.Contains(KeyModifierAlt) {
		descriptions = append(descriptions, KeyModifierAlt.String())
	}
	if s.Contains(KeyModifierCapsLock) {
		descriptions = append(descriptions, KeyModifierCapsLock.String())
	}
	return fmt.Sprintf("(%s)", strings.Join(descriptions, ","))
}
