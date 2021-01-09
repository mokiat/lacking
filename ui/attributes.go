package ui

// AttributeSet represents a set of attributes that can
// be applied to elements and control.
type AttributeSet interface {
	BoolAttribute(name string) (bool, bool)
	IntAttribute(name string) (int, bool)
	FloatAttribute(name string) (float32, bool)
	StringAttribute(name string) (string, bool)
}

// HierarchicalAttributeSet builds an attribute set that
// when fetching values tries from the last one towards
// the first one before giving up.
func HierarchicalAttributeSet(sets ...AttributeSet) AttributeSet {
	var result *hierarchicalAttributeSet
	for _, set := range sets {
		result = &hierarchicalAttributeSet{
			parent: result,
			set:    set,
		}
	}
	return result
}

type hierarchicalAttributeSet struct {
	parent *hierarchicalAttributeSet
	set    AttributeSet
}

func (s *hierarchicalAttributeSet) BoolAttribute(name string) (bool, bool) {
	if s == nil {
		return false, false
	}
	if value, ok := s.set.BoolAttribute(name); ok {
		return value, true
	}
	return s.parent.BoolAttribute(name)
}

func (s *hierarchicalAttributeSet) IntAttribute(name string) (int, bool) {
	if s == nil {
		return 0, false
	}
	if value, ok := s.set.IntAttribute(name); ok {
		return value, true
	}
	return s.parent.IntAttribute(name)
}

func (s *hierarchicalAttributeSet) FloatAttribute(name string) (float32, bool) {
	if s == nil {
		return 0.0, false
	}
	if value, ok := s.set.FloatAttribute(name); ok {
		return value, true
	}
	return s.parent.FloatAttribute(name)
}

func (s *hierarchicalAttributeSet) StringAttribute(name string) (string, bool) {
	if s == nil {
		return "", false
	}
	if value, ok := s.set.StringAttribute(name); ok {
		return value, true
	}
	return s.parent.StringAttribute(name)
}
