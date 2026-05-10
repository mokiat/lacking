package internal

import (
	"math/bits"
)

// MaxComponentTypes is the maximum number of component types that can be
// registered in the ECS.
const MaxComponentTypes = 1 << 8

// TypeID is a unique identifier for a component type.
type TypeID uint8

// EmptyTypeMask returns an empty TypeMask with no component types.
func EmptyTypeMask() TypeMask {
	return TypeMask{}
}

// TypeMaskFromType returns a TypeMask containing only the component type with
// the given ID.
func TypeMaskFromType(id TypeID) TypeMask {
	var mask TypeMask
	mask.AddType(id)
	return mask
}

// TypeMaskFromTypes returns a TypeMask containing the component types with the
// given IDs.
func TypeMaskFromTypes(ids ...TypeID) TypeMask {
	var mask TypeMask
	for _, id := range ids {
		mask.AddType(id)
	}
	return mask
}

// TypeMask is a bitmask representing a set of component types. Each bit in the
// mask corresponds to a component type.
type TypeMask [4]uint64

// Clear removes all component types from the TypeMask, resulting in an empty
// TypeMask.
func (m *TypeMask) Clear() {
	for i := range m {
		m[i] = 0
	}
}

// AddType adds the component type with the given ID to the TypeMask.
func (m *TypeMask) AddType(id TypeID) {
	index := int(id >> 6)
	mask := uint64(1 << (id & 0x3F))
	m[index] |= mask
}

// RemoveType removes the component type with the given ID from the TypeMask.
func (m *TypeMask) RemoveType(id TypeID) {
	index := int(id >> 6)
	mask := uint64(1 << (id & 0x3F))
	m[index] &^= mask
}

// HasType checks if the TypeMask contains the component type with the given ID.
func (m *TypeMask) HasType(id TypeID) bool {
	index := int(id >> 6)
	mask := uint64(1 << (id & 0x3F))
	return (m[index] & mask) != 0
}

// Inverted returns a new TypeMask that contains all component types not present
// in the original TypeMask and does not contain any component types present in
// the original TypeMask.
func (m *TypeMask) Inverted() TypeMask {
	var result TypeMask
	for i := range m {
		result[i] = ^m[i]
	}
	return result
}

// Combine adds all component types from the other TypeMask to the current
// TypeMask.
func (m *TypeMask) Combine(other TypeMask) {
	for i := range m {
		m[i] |= other[i]
	}
}

// Intersects checks if the TypeMask shares any component types with the other
// TypeMask.
func (m *TypeMask) Intersects(other TypeMask) bool {
	for i := range m {
		if (m[i] & other[i]) != 0 {
			return true
		}
	}
	return false
}

// Contains checks if the TypeMask contains all component types in the other
// TypeMask.
func (m *TypeMask) Contains(other TypeMask) bool {
	for i := range m {
		if (m[i] & other[i]) != other[i] {
			return false
		}
	}
	return true
}

// EachType iterates over all component types in the TypeMask and calls the
// provided function with the ID of each component type.
func (m *TypeMask) EachType(f func(TypeID)) {
	for i, mask := range m {
		for mask != 0 {
			bitIndex := bits.TrailingZeros64(mask)
			id := TypeID(i*64 + bitIndex)
			f(id)
			mask &= (mask - 1)
		}
	}
}

// TypeLookup is a mapping from component type IDs to their corresponding
// indices in an external array (e.g. the components array in a component
// archetype).
type TypeLookup [MaxComponentTypes]uint8
