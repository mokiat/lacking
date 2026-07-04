package placement3d

import "github.com/mokiat/lacking/core/spatial/query3d"

type baseShape[S any] struct {
	objectIndex int32
	nextShape   int32

	spatialID query3d.TreeItemID
	static    bool

	rejectGroup uint32
	sourceMask  uint32
	targetMask  uint32

	userData S
}

func (s *baseShape[S]) matchesFilter(filter Filter) bool {
	if s.static && filter.SkipStatic {
		return false
	}
	if !s.static && filter.SkipDynamic {
		return false
	}
	if mask, ok := filter.Mask.Unwrap(); ok {
		if (s.sourceMask & mask) == 0 {
			return false
		}
	}
	return true
}

func shapesCanIntersect[S any](a, b *baseShape[S]) bool {
	if a.objectIndex == b.objectIndex {
		return false
	}
	if !a.static && !b.static && a.objectIndex >= b.objectIndex {
		return false // prevent double checks for dynamic shapes
	}
	if a.rejectGroup != 0 && (a.rejectGroup == b.rejectGroup) {
		return false
	}
	if ((a.sourceMask & b.targetMask) == 0) && ((a.targetMask & b.sourceMask) == 0) {
		return false
	}
	return true
}
