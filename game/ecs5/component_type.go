package ecs5

const MaxComponentCount = 64

type componentMask uint64

func (m componentMask) Contains(query componentMask) bool {
	return (m & query) != 0
}
