package shape3d

import "github.com/mokiat/gomath/dprec"

// Frustum represents a convex region in 3D space that is bounded by six
// clipping planes, such as the volume that a camera sees. Each plane is a
// [Surface] whose normal faces into the region; a point belongs to the
// region when it lies on the facing side of (or on) every one of the six
// surfaces.
//
// Unlike the other shapes in this package, which describe a solid through
// its outward facing surface, a frustum is a clip volume and follows the
// inward facing convention that clipping planes use. The region need not be
// a truncated pyramid: an orthographic projection yields a box (see also
// [FrustumFromAABB]) and repeating a surface yields a region that is
// effectively bounded by fewer planes.
type Frustum struct {
	// Surfaces holds the six bounding planes, with normals facing inwards.
	Surfaces [6]Surface
}

// NewFrustum creates a [Frustum] from the given six surfaces. The normal of
// each surface is expected to face into the region.
func NewFrustum(surfaces [6]Surface) Frustum {
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
func FrustumFromProjection(matrix dprec.Mat4) Frustum {
	row1 := matrix.Row1()
	row2 := matrix.Row2()
	row3 := matrix.Row3()
	row4 := matrix.Row4()
	return Frustum{
		// The surfaces are ordered so that typical queries get rejected as
		// early as possible.
		Surfaces: [6]Surface{
			surfaceFromClipVec(dprec.Vec4Sum(row4, row1)),  // left
			surfaceFromClipVec(dprec.Vec4Diff(row4, row1)), // right
			surfaceFromClipVec(dprec.Vec4Diff(row4, row3)), // far
			surfaceFromClipVec(dprec.Vec4Sum(row4, row2)),  // bottom
			surfaceFromClipVec(dprec.Vec4Diff(row4, row2)), // top
			surfaceFromClipVec(dprec.Vec4Sum(row4, row3)),  // near
		},
	}
}

// FrustumFromAABB returns a [Frustum] that covers exactly the volume of the
// specified axis-aligned bounding box.
func FrustumFromAABB(aabb AABB) Frustum {
	return Frustum{
		Surfaces: [6]Surface{
			{Normal: dprec.NewVec3(1.0, 0.0, 0.0), Distance: aabb.MinX},
			{Normal: dprec.NewVec3(-1.0, 0.0, 0.0), Distance: -aabb.MaxX},
			{Normal: dprec.NewVec3(0.0, 1.0, 0.0), Distance: aabb.MinY},
			{Normal: dprec.NewVec3(0.0, -1.0, 0.0), Distance: -aabb.MaxY},
			{Normal: dprec.NewVec3(0.0, 0.0, 1.0), Distance: aabb.MinZ},
			{Normal: dprec.NewVec3(0.0, 0.0, -1.0), Distance: -aabb.MaxZ},
		},
	}
}

// ContainsPoint returns whether the specified point lies within the region,
// which is the case when it is on the facing side of (or on) every one of the
// six surfaces.
func (h Frustum) ContainsPoint(point dprec.Vec3) bool {
	for i := range h.Surfaces {
		if h.Surfaces[i].SignedDistance(point) < 0.0 {
			return false
		}
	}
	return true
}

// surfaceFromClipVec converts a clip plane in the a*x + b*y + c*z + d = 0
// form, where a positive result marks the inside, into a normalized Surface.
func surfaceFromClipVec(v dprec.Vec4) Surface {
	normal := dprec.NewVec3(v.X, v.Y, v.Z)
	correction := 1.0 / normal.Length()
	return Surface{
		Normal:   dprec.Vec3Prod(normal, correction),
		Distance: -v.W * correction,
	}
}
