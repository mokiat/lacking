package ecs

import "iter"

type componentMask [8]uint64

func emptyComponentMask() componentMask {
	return componentMask{}
}

func componentMaskFromType(tIndex typeID) componentMask {
	index := int(tIndex / 64)
	mask := uint64(1 << (tIndex % 64))
	var result componentMask
	result[index] = mask
	return result
}

func componentMaskFromTypes(tIndices ...typeID) componentMask {
	var result componentMask
	for _, tIndex := range tIndices {
		index := int(tIndex / 64)
		mask := uint64(1 << (tIndex % 64))
		result[index] |= mask
	}
	return result
}

func (m *componentMask) addType(tIndex typeID) {
	index := int(tIndex / 64)
	mask := uint64(1 << (tIndex % 64))
	m[index] |= mask
}

func (m *componentMask) removeType(tIndex typeID) {
	index := int(tIndex / 64)
	mask := uint64(1 << (tIndex % 64))
	m[index] &^= mask
}

func (m *componentMask) containsType(tIndex typeID) bool {
	index := int(tIndex / 64)
	mask := uint64(1 << (tIndex % 64))
	return (m[index] & mask) != 0
}

func (m *componentMask) addMask(other componentMask) {
	for i := range m {
		m[i] |= other[i]
	}
}

func (m *componentMask) intersectsMask(other componentMask) bool {
	for i := range m {
		if (m[i] & other[i]) != 0 {
			return true
		}
	}
	return false
}

func (m *componentMask) containsMask(other componentMask) bool {
	for i := range m {
		if (m[i] & other[i]) != other[i] {
			return false
		}
	}
	return true
}

func (m *componentMask) typeIndicesIter() iter.Seq[typeID] {
	return func(yield func(typeID) bool) {
		for i := range uint32(512) {
			if m.containsType(typeID(i)) {
				if !yield(typeID(i)) {
					return
				}
			}
		}
	}
}
