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
