package ui

import (
	"fmt"

	"github.com/mokiat/lacking/app"
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

// String returns a string representation for this keyboard event.
func (e KeyboardEvent) String() string {
	return fmt.Sprintf("(%s,%d,%c,%s)",
		e.Type,
		e.Code,
		e.Rune,
		e.Modifiers,
	)
}

const (
	KeyboardEventTypeKeyDown = app.KeyboardEventTypeKeyDown
	KeyboardEventTypeKeyUp   = app.KeyboardEventTypeKeyUp
	KeyboardEventTypeRepeat  = app.KeyboardEventTypeRepeat
	KeyboardEventTypeType    = app.KeyboardEventTypeType
)

// KeyboardEventType is used to specify the type of keyboard
// action that occurred.
type KeyboardEventType = app.KeyboardEventType

const (
	KeyCodeEscape       = app.KeyCodeEscape
	KeyCodeEnter        = app.KeyCodeEnter
	KeyCodeSpace        = app.KeyCodeSpace
	KeyCodeTab          = app.KeyCodeTab
	KeyCodeCaps         = app.KeyCodeCaps
	KeyCodeLeftShift    = app.KeyCodeLeftShift
	KeyCodeRightShift   = app.KeyCodeRightShift
	KeyCodeLeftControl  = app.KeyCodeLeftControl
	KeyCodeRightControl = app.KeyCodeRightControl
	KeyCodeLeftAlt      = app.KeyCodeLeftAlt
	KeyCodeRightAlt     = app.KeyCodeRightAlt
	KeyCodeBackspace    = app.KeyCodeBackspace
	KeyCodeInsert       = app.KeyCodeInsert
	KeyCodeDelete       = app.KeyCodeDelete
	KeyCodeHome         = app.KeyCodeHome
	KeyCodeEnd          = app.KeyCodeEnd
	KeyCodePageUp       = app.KeyCodePageUp
	KeyCodePageDown     = app.KeyCodePageDown
	KeyCodeArrowLeft    = app.KeyCodeArrowLeft
	KeyCodeArrowRight   = app.KeyCodeArrowRight
	KeyCodeArrowUp      = app.KeyCodeArrowUp
	KeyCodeArrowDown    = app.KeyCodeArrowDown
	KeyCodeMinus        = app.KeyCodeMinus
	KeyCodeEqual        = app.KeyCodeEqual
	KeyCodeLeftBracket  = app.KeyCodeLeftBracket
	KeyCodeRightBracket = app.KeyCodeRightBracket
	KeyCodeSemicolon    = app.KeyCodeSemicolon
	KeyCodeComma        = app.KeyCodeComma
	KeyCodePeriod       = app.KeyCodePeriod
	KeyCodeSlash        = app.KeyCodeSlash
	KeyCodeBackslash    = app.KeyCodeBackslash
	KeyCodeApostrophe   = app.KeyCodeApostrophe
	KeyCodeGraveAccent  = app.KeyCodeGraveAccent
	KeyCodeA            = app.KeyCodeA
	KeyCodeB            = app.KeyCodeB
	KeyCodeC            = app.KeyCodeC
	KeyCodeD            = app.KeyCodeD
	KeyCodeE            = app.KeyCodeE
	KeyCodeF            = app.KeyCodeF
	KeyCodeG            = app.KeyCodeG
	KeyCodeH            = app.KeyCodeH
	KeyCodeI            = app.KeyCodeI
	KeyCodeJ            = app.KeyCodeJ
	KeyCodeK            = app.KeyCodeK
	KeyCodeL            = app.KeyCodeL
	KeyCodeM            = app.KeyCodeM
	KeyCodeN            = app.KeyCodeN
	KeyCodeO            = app.KeyCodeO
	KeyCodeP            = app.KeyCodeP
	KeyCodeQ            = app.KeyCodeQ
	KeyCodeR            = app.KeyCodeR
	KeyCodeS            = app.KeyCodeS
	KeyCodeT            = app.KeyCodeT
	KeyCodeU            = app.KeyCodeU
	KeyCodeV            = app.KeyCodeV
	KeyCodeW            = app.KeyCodeW
	KeyCodeX            = app.KeyCodeX
	KeyCodeY            = app.KeyCodeY
	KeyCodeZ            = app.KeyCodeZ
	KeyCode0            = app.KeyCode0
	KeyCode1            = app.KeyCode1
	KeyCode2            = app.KeyCode2
	KeyCode3            = app.KeyCode3
	KeyCode4            = app.KeyCode4
	KeyCode5            = app.KeyCode5
	KeyCode6            = app.KeyCode6
	KeyCode7            = app.KeyCode7
	KeyCode8            = app.KeyCode8
	KeyCode9            = app.KeyCode9
	KeyCodeF1           = app.KeyCodeF1
	KeyCodeF2           = app.KeyCodeF2
	KeyCodeF3           = app.KeyCodeF3
	KeyCodeF4           = app.KeyCodeF4
	KeyCodeF5           = app.KeyCodeF5
	KeyCodeF6           = app.KeyCodeF6
	KeyCodeF7           = app.KeyCodeF7
	KeyCodeF8           = app.KeyCodeF8
	KeyCodeF9           = app.KeyCodeF9
	KeyCodeF10          = app.KeyCodeF10
	KeyCodeF11          = app.KeyCodeF11
	KeyCodeF12          = app.KeyCodeF12
)

// KeyCode represents a keyboard key.
type KeyCode = app.KeyCode

const (
	KeyModifierControl  = app.KeyModifierControl
	KeyModifierShift    = app.KeyModifierShift
	KeyModifierAlt      = app.KeyModifierAlt
	KeyModifierCapsLock = app.KeyModifierCapsLock
	KeyModifierSuper    = app.KeyModifierSuper
)

// KeyModifier represents a modifier key.
type KeyModifier = app.KeyModifier

// KeyModifierSet is used to indicate which modifier
// keys were active at the event occurrence.
type KeyModifierSet = app.KeyModifierSet
