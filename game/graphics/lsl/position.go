package lsl

import "fmt"

// At creates a new Position with the given line and column numbers.
func At(line, column uint32) Position {
	return Position{
		Line:   line,
		Column: column,
	}
}

// Position represents a position of interest in the source code.
type Position struct {

	// Line is the line number, starting from 1.
	Line uint32

	// Column is the column number, starting from 1.
	Column uint32
}

// String returns a string representation of the position.
func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}
