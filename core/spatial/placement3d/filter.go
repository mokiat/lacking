package placement3d

import "github.com/mokiat/gog/opt"

// Filter represents a set of criteria to filter 3D shapes in a scene.
type Filter struct {

	// Mask is a bitmask used to filter shapes based on their assigned layers.
	Mask opt.T[uint32]

	// SkipDynamic indicates whether dynamic shapes should be excluded from the
	// results.
	SkipDynamic bool

	// SkipStatic indicates whether static shapes should be excluded from the
	// results.
	SkipStatic bool
}

type filterRepresentation struct {
	rejectGroup uint32
	sourceMask  uint32
	targetMask  uint32
}

func (s *filterRepresentation) matchesFilter(filter Filter) bool {
	if mask, ok := filter.Mask.Unwrap(); ok {
		if (s.sourceMask & mask) == 0 {
			return false
		}
	}
	return true
}

func (s *filterRepresentation) canInteractWith(other *filterRepresentation) bool {
	if s.rejectGroup != 0 && (s.rejectGroup == other.rejectGroup) {
		return false
	}
	if ((s.sourceMask & other.targetMask) == 0) && ((s.targetMask & other.sourceMask) == 0) {
		return false
	}
	return true
}
