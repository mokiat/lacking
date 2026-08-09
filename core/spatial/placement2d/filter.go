package placement2d

import "github.com/mokiat/gog/opt"

// Mask is a bitmask over the layers that a shape can occupy. Queries use it to
// narrow down the shapes that they consider.
//
// A shape is considered by a query when at least one bit is set in both the
// mask of the query and the [FilterInfo.SourceMask] of the shape. Note that
// this means that the zero value matches no shape at all. Use [FullMask] to
// consider every shape in the scene.
type Mask = uint32

// FullMask is a [Mask] with all layer bits set. A query that uses it considers
// every shape in the scene, regardless of the layers that the shape occupies.
const FullMask Mask = 0xFFFFFFFF

// FilterInfo holds the collision-filtering metadata common to every shape that
// can be placed in a scene, whether an object shape (see [CircleInfo] and
// [RectangleInfo]) or a terrain shape (see [MeshInfo]).
//
// Its fields determine which shapes are tested against one another during
// intersection queries.
type FilterInfo struct {

	// RejectGroup becomes active if a value larger than zero is specified.
	// Shapes that share the same reject group are not checked for
	// intersection.
	RejectGroup uint32

	// SourceMask specifies the layers in which this shape is positioned.
	//
	// Defaults to the first layer only.
	SourceMask opt.T[uint32]

	// TargetMask specifies the layers with which this shape can intersect.
	//
	// Defaults to the first layer only.
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

// satisfiesMask reports whether this shape occupies at least one of the layers
// covered by the specified query mask.
func (s *filterRepresentation) satisfiesMask(mask Mask) bool {
	return (s.sourceMask & mask) != 0
}

// canInteractWith reports whether this shape and the specified one are allowed
// to be checked for intersection.
func (s *filterRepresentation) canInteractWith(other *filterRepresentation) bool {
	if s.rejectGroup != 0 && (s.rejectGroup == other.rejectGroup) {
		return false
	}
	if ((s.sourceMask & other.targetMask) == 0) && ((s.targetMask & other.sourceMask) == 0) {
		return false
	}
	return true
}
