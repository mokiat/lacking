package state

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/util/algo"
)

// NewHistory creates a new History instance with the specified capacity,
// which determines the maximum amount of changes that will be tracked.
func NewHistory(capacity int) *History {
	return &History{
		undoStack: algo.NewClampStack[Change](capacity),
		redoStack: ds.NewStack[Change](capacity),
	}
}

// History tracks the sequence of changes and allows for undo and redo.
type History struct {
	undoStack *algo.ClampStack[Change]
	redoStack *ds.Stack[Change]
}

// Clear removes all undo and redo history tracked.
func (h *History) Clear() {
	h.undoStack.Clear()
	h.redoStack.Clear()
}

// LastChange returns the last applied change, if there is one, otherwise
// it returns nil.
func (h *History) LastChange() Change {
	if h.undoStack.IsEmpty() {
		return nil
	}
	return h.undoStack.Peek()
}

// Do applies the specified change and tracks it in the undo history.
func (h *History) Do(change Change) {
	if change == nil {
		return
	}
	if extChange, ok := h.extendableChange(); !ok || !extChange.Extend(change) {
		h.undoStack.Push(change)
	}
	h.redoStack.Clear()
	change.Apply()
}

// CanUndo returns whether there is a tracked change that can be undone.
func (h *History) CanUndo() bool {
	return !h.undoStack.IsEmpty()
}

// Undo reverts the last change and tracks it into the redo history.
func (h *History) Undo() {
	change := h.undoStack.Pop()
	h.redoStack.Push(change)
	change.Revert()
}

// CanRedo returns whether there is a tracked change that can be redone.
func (h *History) CanRedo() bool {
	return !h.redoStack.IsEmpty()
}

// Redo applies a change that was previously undone.
func (h *History) Redo() {
	change := h.redoStack.Pop()
	h.undoStack.Push(change)
	change.Apply()
}

func (h *History) extendableChange() (ExtendableChange, bool) {
	if h.undoStack.IsEmpty() {
		return nil, false
	}
	change := h.undoStack.Peek()
	extendable, ok := change.(ExtendableChange)
	if !ok {
		return nil, false
	}
	return extendable, true
}
