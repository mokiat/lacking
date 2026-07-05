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

// FilterInfo holds the collision-filtering metadata common to every entity
// that can be placed in a scene, whether a shape (see [SphereInfo] and
// [BoxInfo]) or a mesh (see [MeshInfo]).
//
// Its fields determine which entities are tested against one another during
// intersection queries.
type FilterInfo struct {

	// RejectGroup becomes active if a value larger than zero is specified.
	// Entities that share the same reject group are not checked for
	// intersection.
	RejectGroup uint32

	// SourceMask specifies the layers in which this entity is positioned.
	SourceMask opt.T[uint32]

	// TargetMask specifies the layers with which this entity can intersect.
	TargetMask opt.T[uint32]
}

type filterRepresentation struct {
	rejectGroup uint32
	sourceMask  uint32
	targetMask  uint32
}

func newFilterRepresentation(info FilterInfo) filterRepresentation {
	return filterRepresentation{
		rejectGroup: info.RejectGroup,
		sourceMask:  info.SourceMask.ValueOrDefault(0b1),
		targetMask:  info.TargetMask.ValueOrDefault(0b1),
	}
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
