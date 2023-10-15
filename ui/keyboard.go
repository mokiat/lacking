package ui

import (
	"fmt"
	"strings"

	"github.com/mokiat/lacking/app"
)

// KeyboardEvent is used to propagate events related to keyboard
// actions.
type KeyboardEvent struct {

	// Action specifies the keyboard event type.
	Action KeyboardAction

	// Code returns the code of the keyboard key.
	Code KeyCode

	// Rune returns the character that was typed in case
	// of an KeyboardEventTypeType event.
	Rune rune

	// Modifiers contains a set of modifier keys that were
	// pressed during the event.
	Modifiers KeyModifierSet
}

// String returns a string representation for this keyboard event.
func (e KeyboardEvent) String() string {
	return fmt.Sprintf("(%s,%d,%c,%s)",
		e.Action,
		e.Code,
		e.Rune,
		e.Modifiers,
	)
}

const (
	KeyboardActionDown   KeyboardAction = KeyboardAction(app.KeyboardActionDown)
	KeyboardActionUp     KeyboardAction = KeyboardAction(app.KeyboardActionUp)
	KeyboardActionRepeat KeyboardAction = KeyboardAction(app.KeyboardActionRepeat)
	KeyboardActionType   KeyboardAction = KeyboardAction(app.KeyboardActionType)
)

// KeyboardEventType is used to specify the type of keyboard
// action that occurred.
type KeyboardAction int

// String returns a string representation of this event type,
func (a KeyboardAction) String() string {
	switch a {
	case KeyboardActionDown:
		return "DOWN"
	case KeyboardActionUp:
		return "UP"
	case KeyboardActionRepeat:
		return "REPEAT"
	case KeyboardActionType:
		return "TYPE"
	default:
		return "UNKNOWN"
	}
}

const (
	KeyCodeEscape       KeyCode = KeyCode(app.KeyCodeEscape)
	KeyCodeEnter        KeyCode = KeyCode(app.KeyCodeEnter)
	KeyCodeSpace        KeyCode = KeyCode(app.KeyCodeSpace)
	KeyCodeTab          KeyCode = KeyCode(app.KeyCodeTab)
	KeyCodeCaps         KeyCode = KeyCode(app.KeyCodeCaps)
	KeyCodeLeftShift    KeyCode = KeyCode(app.KeyCodeLeftShift)
	KeyCodeRightShift   KeyCode = KeyCode(app.KeyCodeRightShift)
	KeyCodeLeftControl  KeyCode = KeyCode(app.KeyCodeLeftControl)
	KeyCodeRightControl KeyCode = KeyCode(app.KeyCodeRightControl)
	KeyCodeLeftAlt      KeyCode = KeyCode(app.KeyCodeLeftAlt)
	KeyCodeRightAlt     KeyCode = KeyCode(app.KeyCodeRightAlt)
	KeyCodeLeftSuper    KeyCode = KeyCode(app.KeyCodeLeftSuper)
	KeyCodeRightSuper   KeyCode = KeyCode(app.KeyCodeRightSuper)
	KeyCodeBackspace    KeyCode = KeyCode(app.KeyCodeBackspace)
	KeyCodeInsert       KeyCode = KeyCode(app.KeyCodeInsert)
	KeyCodeDelete       KeyCode = KeyCode(app.KeyCodeDelete)
	KeyCodeHome         KeyCode = KeyCode(app.KeyCodeHome)
	KeyCodeEnd          KeyCode = KeyCode(app.KeyCodeEnd)
	KeyCodePageUp       KeyCode = KeyCode(app.KeyCodePageUp)
	KeyCodePageDown     KeyCode = KeyCode(app.KeyCodePageDown)
	KeyCodeArrowLeft    KeyCode = KeyCode(app.KeyCodeArrowLeft)
	KeyCodeArrowRight   KeyCode = KeyCode(app.KeyCodeArrowRight)
	KeyCodeArrowUp      KeyCode = KeyCode(app.KeyCodeArrowUp)
	KeyCodeArrowDown    KeyCode = KeyCode(app.KeyCodeArrowDown)
	KeyCodeMinus        KeyCode = KeyCode(app.KeyCodeMinus)
	KeyCodeEqual        KeyCode = KeyCode(app.KeyCodeEqual)
	KeyCodeLeftBracket  KeyCode = KeyCode(app.KeyCodeLeftBracket)
	KeyCodeRightBracket KeyCode = KeyCode(app.KeyCodeRightBracket)
	KeyCodeSemicolon    KeyCode = KeyCode(app.KeyCodeSemicolon)
	KeyCodeComma        KeyCode = KeyCode(app.KeyCodeComma)
	KeyCodePeriod       KeyCode = KeyCode(app.KeyCodePeriod)
	KeyCodeSlash        KeyCode = KeyCode(app.KeyCodeSlash)
	KeyCodeBackslash    KeyCode = KeyCode(app.KeyCodeBackslash)
	KeyCodeApostrophe   KeyCode = KeyCode(app.KeyCodeApostrophe)
	KeyCodeGraveAccent  KeyCode = KeyCode(app.KeyCodeGraveAccent)
	KeyCodeA            KeyCode = KeyCode(app.KeyCodeA)
	KeyCodeB            KeyCode = KeyCode(app.KeyCodeB)
	KeyCodeC            KeyCode = KeyCode(app.KeyCodeC)
	KeyCodeD            KeyCode = KeyCode(app.KeyCodeD)
	KeyCodeE            KeyCode = KeyCode(app.KeyCodeE)
	KeyCodeF            KeyCode = KeyCode(app.KeyCodeF)
	KeyCodeG            KeyCode = KeyCode(app.KeyCodeG)
	KeyCodeH            KeyCode = KeyCode(app.KeyCodeH)
	KeyCodeI            KeyCode = KeyCode(app.KeyCodeI)
	KeyCodeJ            KeyCode = KeyCode(app.KeyCodeJ)
	KeyCodeK            KeyCode = KeyCode(app.KeyCodeK)
	KeyCodeL            KeyCode = KeyCode(app.KeyCodeL)
	KeyCodeM            KeyCode = KeyCode(app.KeyCodeM)
	KeyCodeN            KeyCode = KeyCode(app.KeyCodeN)
	KeyCodeO            KeyCode = KeyCode(app.KeyCodeO)
	KeyCodeP            KeyCode = KeyCode(app.KeyCodeP)
	KeyCodeQ            KeyCode = KeyCode(app.KeyCodeQ)
	KeyCodeR            KeyCode = KeyCode(app.KeyCodeR)
	KeyCodeS            KeyCode = KeyCode(app.KeyCodeS)
	KeyCodeT            KeyCode = KeyCode(app.KeyCodeT)
	KeyCodeU            KeyCode = KeyCode(app.KeyCodeU)
	KeyCodeV            KeyCode = KeyCode(app.KeyCodeV)
	KeyCodeW            KeyCode = KeyCode(app.KeyCodeW)
	KeyCodeX            KeyCode = KeyCode(app.KeyCodeX)
	KeyCodeY            KeyCode = KeyCode(app.KeyCodeY)
	KeyCodeZ            KeyCode = KeyCode(app.KeyCodeZ)
	KeyCode0            KeyCode = KeyCode(app.KeyCode0)
	KeyCode1            KeyCode = KeyCode(app.KeyCode1)
	KeyCode2            KeyCode = KeyCode(app.KeyCode2)
	KeyCode3            KeyCode = KeyCode(app.KeyCode3)
	KeyCode4            KeyCode = KeyCode(app.KeyCode4)
	KeyCode5            KeyCode = KeyCode(app.KeyCode5)
	KeyCode6            KeyCode = KeyCode(app.KeyCode6)
	KeyCode7            KeyCode = KeyCode(app.KeyCode7)
	KeyCode8            KeyCode = KeyCode(app.KeyCode8)
	KeyCode9            KeyCode = KeyCode(app.KeyCode9)
	KeyCodeF1           KeyCode = KeyCode(app.KeyCodeF1)
	KeyCodeF2           KeyCode = KeyCode(app.KeyCodeF2)
	KeyCodeF3           KeyCode = KeyCode(app.KeyCodeF3)
	KeyCodeF4           KeyCode = KeyCode(app.KeyCodeF4)
	KeyCodeF5           KeyCode = KeyCode(app.KeyCodeF5)
	KeyCodeF6           KeyCode = KeyCode(app.KeyCodeF6)
	KeyCodeF7           KeyCode = KeyCode(app.KeyCodeF7)
	KeyCodeF8           KeyCode = KeyCode(app.KeyCodeF8)
	KeyCodeF9           KeyCode = KeyCode(app.KeyCodeF9)
	KeyCodeF10          KeyCode = KeyCode(app.KeyCodeF10)
	KeyCodeF11          KeyCode = KeyCode(app.KeyCodeF11)
	KeyCodeF12          KeyCode = KeyCode(app.KeyCodeF12)
)

// KeyCode represents a keyboard key.
type KeyCode int

// String returns a string representation of this key code.
func (c KeyCode) String() string {
	return app.KeyCode(c).String()
}

const (
	KeyModifierControl KeyModifier = 1 << (iota + 1)
	KeyModifierShift
	KeyModifierAlt
	KeyModifierCapsLock
	KeyModifierSuper
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
	case KeyModifierSuper:
		return "SUPER"
	default:
		return "UNKNOWN"
	}
}

// KeyModifiers constructs a KeyModifierSet by combining the specified
// modifier entries.
func KeyModifiers(entries ...KeyModifier) KeyModifierSet {
	var result KeyModifierSet
	for _, entry := range entries {
		result |= KeyModifierSet(entry)
	}
	return result
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
	if s.Contains(KeyModifierSuper) {
		descriptions = append(descriptions, KeyModifierSuper.String())
	}
	return fmt.Sprintf("(%s)", strings.Join(descriptions, ","))
}
