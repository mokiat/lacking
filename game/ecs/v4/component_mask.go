package ecs

type typeIndex uint32

type componentMask [8]uint64

func emptyComponentMask() componentMask {
	return componentMask{}
}

func componentMaskFromType(tIndex typeIndex) componentMask {
	index := int(tIndex / 64)
	mask := uint64(1 << (tIndex % 64))
	var result componentMask
	result[index] = mask
	return result
}

func componentMaskFromTypes(tIndices ...typeIndex) componentMask {
	var result componentMask
	for _, tIndex := range tIndices {
		index := int(tIndex / 64)
		mask := uint64(1 << (tIndex % 64))
		result[index] |= mask
	}
	return result
}

func (m *componentMask) addType(tIndex typeIndex) {
	index := int(tIndex / 64)
	mask := uint64(1 << (tIndex % 64))
	m[index] |= mask
}

func (m *componentMask) removeType(tIndex typeIndex) {
	index := int(tIndex / 64)
	mask := uint64(1 << (tIndex % 64))
	m[index] &^= mask
}

func (m *componentMask) containsType(tIndex typeIndex) bool {
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
