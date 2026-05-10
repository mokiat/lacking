package app

import "fmt"

// KeyboardEvent is used to propagate events related to keyboard actions.
type KeyboardEvent struct {

	// Action specifies the keyboard event type.
	Action KeyboardAction

	// Code returns the code of the keyboard key.
	Code KeyCode

	// Character returns the character that was typed in case
	// of a KeyboardActionType event.
	Character rune
}

// String returns a string representation of this event.
func (e KeyboardEvent) String() string {
	return fmt.Sprintf("(%s,%s,%c)",
		e.Action,
		e.Code,
		e.Character,
	)
}

const (
	// KeyboardActionDown indicates that a keyboard key was pressed.
	KeyboardActionDown KeyboardAction = 1 + iota

	// KeyboardActionUp indicates that a keyboard key was released.
	KeyboardActionUp

	// KeyboardActionRepeat indicates that a keyboard key is being held pressed.
	KeyboardActionRepeat

	// KeyboardActionType indicates that a character is typed with the keyboard.
	//
	// Such actions would be duplicates of KeyboardActionDown and
	// KeyboardActionRepeat but allow for the character rune to be read which
	// might be the result of modifiers or special keys that would be hard
	// to reconstruct from just the key code.
	KeyboardActionType
)

// KeyboardAction is used to specify the type of keyboard action that occurred.
type KeyboardAction int

// String returns a string representation of this event type.
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
	// KeyCodeEscape indicates the Escape key.
	KeyCodeEscape KeyCode = 1 + iota
	// KeyCodeEnter indicates the Enter key.
	KeyCodeEnter
	// KeyCodeSpace indicates the Space key.
	KeyCodeSpace
	// KeyCodeTab indicates the Tab key.
	KeyCodeTab
	// KeyCodeCaps indicates the Caps Lock key.
	KeyCodeCaps
	// KeyCodeLeftShift indicates the left Shift key.
	KeyCodeLeftShift
	// KeyCodeRightShift indicates the right Shift key.
	KeyCodeRightShift
	// KeyCodeLeftControl indicates the left Control key.
	KeyCodeLeftControl
	// KeyCodeRightControl indicates the right Control key.
	KeyCodeRightControl
	// KeyCodeLeftAlt indicates the left Alt key.
	KeyCodeLeftAlt
	// KeyCodeRightAlt indicates the right Alt key.
	KeyCodeRightAlt
	// KeyCodeLeftSuper indicates the left Super (Win/Cmd) key.
	KeyCodeLeftSuper
	// KeyCodeRightSuper indicates the right Super (Win/Cmd) key.
	KeyCodeRightSuper
	// KeyCodeBackspace indicates the Backspace key.
	KeyCodeBackspace
	// KeyCodeInsert indicates the Insert key.
	KeyCodeInsert
	// KeyCodeDelete indicates the Delete key.
	KeyCodeDelete
	// KeyCodeHome indicates the Home key.
	KeyCodeHome
	// KeyCodeEnd indicates the End key.
	KeyCodeEnd
	// KeyCodePageUp indicates the Page Up key.
	KeyCodePageUp
	// KeyCodePageDown indicates the Page Down key.
	KeyCodePageDown
	// KeyCodeArrowLeft indicates the left arrow key.
	KeyCodeArrowLeft
	// KeyCodeArrowRight indicates the right arrow key.
	KeyCodeArrowRight
	// KeyCodeArrowUp indicates the up arrow key.
	KeyCodeArrowUp
	// KeyCodeArrowDown indicates the down arrow key.
	KeyCodeArrowDown
	// KeyCodeMinus indicates the minus/hyphen key.
	KeyCodeMinus
	// KeyCodeEqual indicates the equal sign key.
	KeyCodeEqual
	// KeyCodeLeftBracket indicates the left square bracket key.
	KeyCodeLeftBracket
	// KeyCodeRightBracket indicates the right square bracket key.
	KeyCodeRightBracket
	// KeyCodeSemicolon indicates the semicolon key.
	KeyCodeSemicolon
	// KeyCodeComma indicates the comma key.
	KeyCodeComma
	// KeyCodePeriod indicates the period/full stop key.
	KeyCodePeriod
	// KeyCodeSlash indicates the forward slash key.
	KeyCodeSlash
	// KeyCodeBackslash indicates the backslash key.
	KeyCodeBackslash
	// KeyCodeApostrophe indicates the apostrophe/single-quote key.
	KeyCodeApostrophe
	// KeyCodeGraveAccent indicates the grave accent/backtick key.
	KeyCodeGraveAccent
	// KeyCodeA indicates the A key.
	KeyCodeA
	// KeyCodeB indicates the B key.
	KeyCodeB
	// KeyCodeC indicates the C key.
	KeyCodeC
	// KeyCodeD indicates the D key.
	KeyCodeD
	// KeyCodeE indicates the E key.
	KeyCodeE
	// KeyCodeF indicates the F key.
	KeyCodeF
	// KeyCodeG indicates the G key.
	KeyCodeG
	// KeyCodeH indicates the H key.
	KeyCodeH
	// KeyCodeI indicates the I key.
	KeyCodeI
	// KeyCodeJ indicates the J key.
	KeyCodeJ
	// KeyCodeK indicates the K key.
	KeyCodeK
	// KeyCodeL indicates the L key.
	KeyCodeL
	// KeyCodeM indicates the M key.
	KeyCodeM
	// KeyCodeN indicates the N key.
	KeyCodeN
	// KeyCodeO indicates the O key.
	KeyCodeO
	// KeyCodeP indicates the P key.
	KeyCodeP
	// KeyCodeQ indicates the Q key.
	KeyCodeQ
	// KeyCodeR indicates the R key.
	KeyCodeR
	// KeyCodeS indicates the S key.
	KeyCodeS
	// KeyCodeT indicates the T key.
	KeyCodeT
	// KeyCodeU indicates the U key.
	KeyCodeU
	// KeyCodeV indicates the V key.
	KeyCodeV
	// KeyCodeW indicates the W key.
	KeyCodeW
	// KeyCodeX indicates the X key.
	KeyCodeX
	// KeyCodeY indicates the Y key.
	KeyCodeY
	// KeyCodeZ indicates the Z key.
	KeyCodeZ
	// KeyCode0 indicates the 0 key.
	KeyCode0
	// KeyCode1 indicates the 1 key.
	KeyCode1
	// KeyCode2 indicates the 2 key.
	KeyCode2
	// KeyCode3 indicates the 3 key.
	KeyCode3
	// KeyCode4 indicates the 4 key.
	KeyCode4
	// KeyCode5 indicates the 5 key.
	KeyCode5
	// KeyCode6 indicates the 6 key.
	KeyCode6
	// KeyCode7 indicates the 7 key.
	KeyCode7
	// KeyCode8 indicates the 8 key.
	KeyCode8
	// KeyCode9 indicates the 9 key.
	KeyCode9
	// KeyCodeF1 indicates the F1 function key.
	KeyCodeF1
	// KeyCodeF2 indicates the F2 function key.
	KeyCodeF2
	// KeyCodeF3 indicates the F3 function key.
	KeyCodeF3
	// KeyCodeF4 indicates the F4 function key.
	KeyCodeF4
	// KeyCodeF5 indicates the F5 function key.
	KeyCodeF5
	// KeyCodeF6 indicates the F6 function key.
	KeyCodeF6
	// KeyCodeF7 indicates the F7 function key.
	KeyCodeF7
	// KeyCodeF8 indicates the F8 function key.
	KeyCodeF8
	// KeyCodeF9 indicates the F9 function key.
	KeyCodeF9
	// KeyCodeF10 indicates the F10 function key.
	KeyCodeF10
	// KeyCodeF11 indicates the F11 function key.
	KeyCodeF11
	// KeyCodeF12 indicates the F12 function key.
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
	case KeyCodeLeftSuper:
		return "LSUPER"
	case KeyCodeRightSuper:
		return "RSUPER"
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
