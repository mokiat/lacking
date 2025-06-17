package ecs

import "iter"

// Entity represents an ECS entity. It an abstract handle to which a number
// of components can be attached and removed.
type Entity struct {
	scene    *Scene
	index    uint32
	revision uint32
}

// Exists returns whether this entity is still present in the Scene.
func (e Entity) Exists() bool {
	return e.scene.HasEntity(e)
}

// Delete marks this entity for deletion.
//
// Once the scene has been Purged, this handle will become invalid and Exists
// will start returning false. In the meantime, it will still be returned
// from queries, unless explicitly requested not to.
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

func (m *bitmask) Clear() {
	for i := range m.values {
		m.values[i] = 0x00
	}
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
