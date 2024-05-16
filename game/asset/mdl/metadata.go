package mdl

import "strconv"

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

func (m Metadata) HasMinDistance() (float64, bool) {
	return m.FloatProperty("min-distance")
}

func (m Metadata) HasMaxDistance() (float64, bool) {
	return m.FloatProperty("max-distance")
}

func (m Metadata) IsSet(key string) bool {
	if m == nil {
		return false
	}
	_, ok := m[key]
	return ok
}

func (m Metadata) FloatProperty(key string) (float64, bool) {
	if m == nil {
		return 0.0, false
	}
	value, ok := m[key]
	if !ok {
		return 0.0, false
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0.0, false
	}
	return floatValue, true
}
