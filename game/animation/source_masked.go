package animation

import "github.com/mokiat/gog/filter"

// NewMaskedSource creates a new animation source that picks specific bones
// from the specified source.
func NewMaskedSource(delegate Source, selection filter.Func[string]) *MaskedSource {
	return &MaskedSource{
		Source:    delegate,
		selection: selection,
	}
}

var _ Source = (*MaskedSource)(nil)

// MaskedSource is an animation source that picks specific bones
// from another animation source.
type MaskedSource struct {
	Source
	selection filter.Func[string]
}

// NodeTransform returns the transformation of the node with the specified
// name at the current time position.
func (s *MaskedSource) NodeTransform(name string) NodeTransform {
	if !s.selection(name) {
		return NodeTransform{}
	}
	return s.Source.NodeTransform(name)
}
