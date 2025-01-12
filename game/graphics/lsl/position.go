package lsl

// Position represents a position of interest in the source code.
type Position struct {

	// Line is the line number, starting from 1.
	Line uint32

	// Column is the column number, starting from 1.
	Column uint32
}
