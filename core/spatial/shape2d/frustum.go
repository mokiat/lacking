package shape2d

import "github.com/mokiat/gomath/dprec"

// Frustum represents a convex region in 2D space that is bounded by four
// clipping lines, such as the area that a camera sees. Each line is a
// [Surface] whose normal faces into the region; a point belongs to the region
// when it lies on the facing side of (or on) every one of the four surfaces.
//
// It is the 2D analogue of the Frustum type in the shape3d package. Unlike the
// other shapes in this package, which describe a solid through its outward
// facing surface, a frustum is a clip region and follows the inward facing
// convention that clipping lines use. The region need not be a trapezoid: an
// orthographic projection yields a rectangle (see also [FrustumFromAABB]) and
// repeating a surface yields a region that is effectively bounded by fewer
// lines.
type Frustum struct {
	// Surfaces holds the four bounding lines, with normals facing inwards.
	Surfaces [4]Surface
}

// NewFrustum creates a [Frustum] from the given four surfaces. The normal of
// each surface is expected to face into the region.
func NewFrustum(surfaces [4]Surface) Frustum {
	return Frustum{
		Surfaces: surfaces,
	}
}

// FrustumFromProjection returns the view frustum described by the specified
// projection matrix. The matrix may be a combined projection and view matrix
// (e.g. projection * view), in which case the frustum is expressed in the
// space that the view matrix maps from (usually world space).
//
// The resulting surfaces have unit normals.
func FrustumFromProjection(matrix dprec.Mat3) Frustum {
	row1 := matrix.Row1()
	row2 := matrix.Row2()
	row3 := matrix.Row3()
	return Frustum{
		// The surfaces are ordered so that typical queries get rejected as
		// early as possible.
		Surfaces: [4]Surface{
			surfaceFromClipVec(dprec.Vec3Sum(row3, row1)),  // left
			surfaceFromClipVec(dprec.Vec3Diff(row3, row1)), // right
			surfaceFromClipVec(dprec.Vec3Sum(row3, row2)),  // bottom
			surfaceFromClipVec(dprec.Vec3Diff(row3, row2)), // top
		},
	}
}

// FrustumFromAABB returns a [Frustum] that covers exactly the area of the
// specified axis-aligned bounding box.
func FrustumFromAABB(aabb AABB) Frustum {
	return Frustum{
		Surfaces: [4]Surface{
			{Normal: dprec.NewVec2(1.0, 0.0), Distance: aabb.MinX},
			{Normal: dprec.NewVec2(-1.0, 0.0), Distance: -aabb.MaxX},
			{Normal: dprec.NewVec2(0.0, 1.0), Distance: aabb.MinY},
			{Normal: dprec.NewVec2(0.0, -1.0), Distance: -aabb.MaxY},
		},
	}
}

// ContainsPoint returns whether the specified point lies within the region,
// which is the case when it is on the facing side of (or on) every one of the
// four surfaces.
func (f Frustum) ContainsPoint(point dprec.Vec2) bool {
	for i := range f.Surfaces {
		if f.Surfaces[i].SignedDistance(point) < 0.0 {
			return false
		}
	}
	return true
}

// surfaceFromClipVec converts a clip line in the a*x + b*y + c = 0 form, where
// a positive result marks the inside, into a normalized Surface.
func surfaceFromClipVec(v dprec.Vec3) Surface {
	normal := dprec.NewVec2(v.X, v.Y)
	correction := 1.0 / normal.Length()
	return Surface{
		Normal:   dprec.Vec2Prod(normal, correction),
		Distance: -v.Z * correction,
	}
}
