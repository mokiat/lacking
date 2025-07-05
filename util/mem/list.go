package mem

import (
	"iter"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/ds"
)

func NilSparseID() SparseID {
	return SparseID{}
}

type SparseID struct {
	blockIndex  uint32
	blockOffset uint16
	revision    uint16
}

func (i SparseID) IsNil() bool {
	return i.revision == 0
}

func (i SparseID) IsBefore(other SparseID) bool {
	if i.blockIndex < other.blockIndex {
		return true
	}
	if i.blockIndex > other.blockIndex {
		return false
	}
	return i.blockOffset < other.blockOffset
}

func NewSparseList[T any](blockSize uint16) *SparseList[T] {
	return &SparseList[T]{
		freeIDs:   ds.NewStack[sparseItemID](0),
		blockSize: blockSize,
	}
}

type SparseList[T any] struct {
	freeIDs   *ds.Stack[sparseItemID]
	blockSize uint16
	blocks    [][]sparseListItem[T]
}

func (l *SparseList[T]) New() (SparseID, *T) {
	if l.freeIDs.IsEmpty() {
		l.grow()
	}
	itemID := l.freeIDs.Pop()
	item := l.itemRef(itemID)
	item.revision++
	item.inUse = true
	id := SparseID{
		blockIndex:  itemID.blockIndex,
		blockOffset: itemID.blockOffset,
		revision:    item.revision,
	}
	return id, &item.value
}

func (l *SparseList[T]) Get(id SparseID) *T {
	if id.revision == 0 {
		return nil
	}
	itemID := sparseItemID{
		blockIndex:  id.blockIndex,
		blockOffset: id.blockOffset,
	}
	item := l.itemRef(itemID)
	if !item.inUse || (item.revision != id.revision) {
		return nil
	}
	return &item.value
}

func (l SparseList[T]) Has(id SparseID) bool {
	if id.revision == 0 {
		return false
	}
	itemID := sparseItemID{
		blockIndex:  id.blockIndex,
		blockOffset: id.blockOffset,
	}
	item := l.itemRef(itemID)
	return item.inUse && (item.revision == id.revision)
}

func (l *SparseList[T]) Iter() iter.Seq2[SparseID, *T] {
	return func(yield func(SparseID, *T) bool) {
		for blockIndex := range l.blocks {
			for blockOffset := range l.blockSize {
				itemID := sparseItemID{
					blockIndex:  uint32(blockIndex),
					blockOffset: uint16(blockOffset),
				}
				item := l.itemRef(itemID)
				if !item.inUse {
					continue
				}
				id := SparseID{
					blockIndex:  uint32(blockIndex),
					blockOffset: uint16(blockOffset),
					revision:    item.revision,
				}
				if !yield(id, &item.value) {
					return
				}
			}
		}
	}
}

func (l *SparseList[T]) Delete(id SparseID) {
	if id.revision == 0 {
		return
	}
	itemID := sparseItemID{
		blockIndex:  id.blockIndex,
		blockOffset: id.blockOffset,
	}
	item := l.itemRef(itemID)
	if !item.inUse || (item.revision != id.revision) {
		return
	}
	item.value = gog.Zero[T]()
	item.revision++
	item.inUse = false
	l.freeIDs.Push(itemID)
}

func (l *SparseList[T]) itemRef(id sparseItemID) *sparseListItem[T] {
	return &l.blocks[id.blockIndex][id.blockOffset]
}

func (l *SparseList[T]) grow() {
	block := make([]sparseListItem[T], l.blockSize)
	blockIndex := uint32(len(l.blocks))
	for blockOffset := range l.blockSize {
		l.freeIDs.Push(sparseItemID{
			blockIndex:  blockIndex,
			blockOffset: blockOffset,
		})
	}
	l.blocks = append(l.blocks, block)
}

type sparseItemID struct {
	blockIndex  uint32
	blockOffset uint16
}

type sparseListItem[T any] struct {
	value    T
	revision uint16
	inUse    bool
}
