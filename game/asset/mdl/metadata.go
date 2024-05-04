package mdl

type Metadata map[string]string

func (m Metadata) HasCollision() bool {
	return m.IsSet("collidable")
}

func (m Metadata) HasSkipCollision() bool {
	return m.IsSet("non-collidable")
}

func (m Metadata) IsInvisible() bool {
	return m.IsSet("invisible")
}

func (m Metadata) IsSet(key string) bool {
	if m == nil {
		return false
	}
	_, ok := m[key]
	return ok
}
