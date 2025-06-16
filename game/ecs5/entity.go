package ecs5

import "iter"

// Entity represents a game entity. It can represent a number
// of things depending on the components that are attached
// to it.
type Entity struct {
	scene    *Scene
	index    uint32
	revision uint32
}

func (e Entity) Exists() bool {
	return e.scene.HasEntity(e)
}

func (e Entity) Delete() {
	e.scene.DeleteEntity(e)
}

type entityHandle struct {
	components        componentMask
	revision          uint32
	isPendingDeletion bool
}

func newBitmask() *bitmask {
	return new(bitmask)
}

type bitmask struct {
	values [16384]uint64
}

func (m *bitmask) Get(index uint32) bool {
	bucket := index / 64
	offset := index % 64
	query := uint64(1 << offset)
	return (m.values[bucket] & query) != 0
}

func (m *bitmask) Set(index uint32, active bool) {
	bucket := index / 64
	offset := index % 64
	query := uint64(1 << offset)
	if active {
		m.values[bucket] |= query
	} else {
		m.values[bucket] &= ^query
	}
}

func (m *bitmask) ActiveIter() iter.Seq[uint32] {
	return func(yield func(uint32) bool) {
		var index uint32
		for _, group := range m.values {
			if group == 0 { // skip whole group
				index += 64
				continue
			}
			for offset := range 64 {
				query := uint64(1 << offset)
				if (group & query) != 0 {
					if !yield(index) {
						return
					}
				}
				index++
			}
		}
	}
}
