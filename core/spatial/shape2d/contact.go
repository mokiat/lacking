package shape2d

import "github.com/mokiat/gomath/dprec"

// Contact describes the intersection of a source shape with a target shape.
//
// Its fields are expressed relative to the target shape. The equivalent values
// for the source shape can be derived via EvalSourcePoint and EvalSourceNormal.
type Contact struct {

	// TargetPoint is the contact point on the surface of the target shape.
	TargetPoint dprec.Vec2

	// TargetNormal is the outward-facing surface normal of the target shape at
	// TargetPoint. It points away from the target and toward the source, and is
	// the direction along which the source shape must be moved by Depth to
	// resolve the intersection.
	TargetNormal dprec.Vec2

	// Depth is the penetration distance between the two shapes measured along
	// TargetNormal. It is always non-negative.
	Depth float64
}

// EvalSourcePoint returns the contact point on the surface of the source shape.
//
// It lies a distance of Depth from TargetPoint along the inverse of
// TargetNormal.
func (c Contact) EvalSourcePoint() dprec.Vec2 {
	return dprec.Vec2Diff(c.TargetPoint, dprec.Vec2Prod(c.TargetNormal, c.Depth))
}

// EvalSourceNormal returns the outward-facing surface normal of the source
// shape at its contact point.
//
// It is the inverse of TargetNormal and points in the direction along which the
// target shape must be moved by Depth to resolve the intersection.
func (c Contact) EvalSourceNormal() dprec.Vec2 {
	return dprec.InverseVec2(c.TargetNormal)
}

// Flipped returns a Contact with the source and target shapes swapped.
//
// The resulting contact describes the same intersection from the perspective of
// the opposite shape.
func (c Contact) Flipped() Contact {
	return Contact{
		TargetPoint:  c.EvalSourcePoint(),
		TargetNormal: c.EvalSourceNormal(),
		Depth:        c.Depth,
	}
}

// ContactCallback is invoked for each Contact discovered while testing shapes
// for intersection.
type ContactCallback func(contact Contact)
