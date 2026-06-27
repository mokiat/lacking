package isec3d

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// CheckSegmentBox returns whether segment intersects box.
func CheckSegmentBox(segment shape3d.Segment, box shape3d.Box) bool {
	relativeStart := dprec.Vec3Diff(segment.A, box.Center)
	delta := dprec.Vec3Diff(segment.B, segment.A)

	startX := dprec.Vec3Dot(relativeStart, box.Rotation.BasisX)
	startY := dprec.Vec3Dot(relativeStart, box.Rotation.BasisY)
	startZ := dprec.Vec3Dot(relativeStart, box.Rotation.BasisZ)
	deltaX := dprec.Vec3Dot(delta, box.Rotation.BasisX)
	deltaY := dprec.Vec3Dot(delta, box.Rotation.BasisY)
	deltaZ := dprec.Vec3Dot(delta, box.Rotation.BasisZ)

	tCloseX, tFarX, okX := slabInterval(startX, deltaX, box.HalfWidth)
	if !okX {
		return false
	}
	tCloseY, tFarY, okY := slabInterval(startY, deltaY, box.HalfHeight)
	if !okY {
		return false
	}
	tCloseZ, tFarZ, okZ := slabInterval(startZ, deltaZ, box.HalfLength)
	if !okZ {
		return false
	}

	tClose := max(tCloseX, tCloseY, tCloseZ)
	tFar := min(tFarX, tFarY, tFarZ)
	return tClose <= tFar && tClose <= 1.0 && tFar >= 0.0
}

// ResolveSegmentBox calls yield for each contact point where segment
// penetrates box. The contact normal points outward from the box surface.
func ResolveSegmentBox(segment shape3d.Segment, box shape3d.Box, yield shape3d.ContactCallback) {
	relativeStart := dprec.Vec3Diff(segment.A, box.Center)
	delta := dprec.Vec3Diff(segment.B, segment.A)

	startX := dprec.Vec3Dot(relativeStart, box.Rotation.BasisX)
	startY := dprec.Vec3Dot(relativeStart, box.Rotation.BasisY)
	startZ := dprec.Vec3Dot(relativeStart, box.Rotation.BasisZ)
	deltaX := dprec.Vec3Dot(delta, box.Rotation.BasisX)
	deltaY := dprec.Vec3Dot(delta, box.Rotation.BasisY)
	deltaZ := dprec.Vec3Dot(delta, box.Rotation.BasisZ)

	tCloseX, tFarX, okX := slabInterval(startX, deltaX, box.HalfWidth)
	if !okX {
		return
	}
	tCloseY, tFarY, okY := slabInterval(startY, deltaY, box.HalfHeight)
	if !okY {
		return
	}
	tCloseZ, tFarZ, okZ := slabInterval(startZ, deltaZ, box.HalfLength)
	if !okZ {
		return
	}
	if max(tCloseX, tCloseY, tCloseZ) > min(tFarX, tFarY, tFarZ) ||
		max(tCloseX, tCloseY, tCloseZ) > 1.0 ||
		min(tFarX, tFarY, tFarZ) < 0.0 {
		return
	}

	midX := startX + deltaX*0.5
	midY := startY + deltaY*0.5
	midZ := startZ + deltaZ*0.5
	halfX := dprec.Abs(deltaX * 0.5)
	halfY := dprec.Abs(deltaY * 0.5)
	halfZ := dprec.Abs(deltaZ * 0.5)

	depth := box.HalfWidth + halfX - dprec.Abs(midX)
	localNormal := dprec.NewVec3(dprec.Sign(midX), 0.0, 0.0)
	if overlapY := box.HalfHeight + halfY - dprec.Abs(midY); overlapY < depth {
		depth = overlapY
		localNormal = dprec.NewVec3(0.0, dprec.Sign(midY), 0.0)
	}
	if overlapZ := box.HalfLength + halfZ - dprec.Abs(midZ); overlapZ < depth {
		depth = overlapZ
		localNormal = dprec.NewVec3(0.0, 0.0, dprec.Sign(midZ))
	}

	localPoint := dprec.NewVec3(
		signedExtent(localNormal.X, box.HalfWidth),
		signedExtent(localNormal.Y, box.HalfHeight),
		signedExtent(localNormal.Z, box.HalfLength),
	)
	yield(shape3d.Contact{
		TargetPoint:  dprec.Vec3Sum(box.Center, box.Rotation.Apply(localPoint)),
		TargetNormal: box.Rotation.Apply(localNormal),
		Depth:        depth,
	})
}

// slabInterval returns the parametric range [tClose, tFar] during which a ray,
// starting at start with the given delta, lies within the slab [-halfExtent,
// halfExtent]. When the ray is parallel to the slab (delta is zero) it reports
// the full [0, 1] range if the start is inside the slab, or ok=false otherwise.
func slabInterval(start, delta, halfExtent float64) (tClose, tFar float64, ok bool) {
	if delta == 0.0 {
		if start < -halfExtent || start > halfExtent {
			return 0.0, 0.0, false
		}
		return 0.0, 1.0, true
	}
	tLow := (-halfExtent - start) / delta
	tHigh := (halfExtent - start) / delta
	return min(tLow, tHigh), max(tLow, tHigh), true
}

// signedExtent returns the box half-extent with the sign of the given component,
// or zero when the component is zero. It selects the box's support coordinate
// along a single axis.
func signedExtent(component, halfExtent float64) float64 {
	switch {
	case component > 0.0:
		return halfExtent
	case component < 0.0:
		return -halfExtent
	default:
		return 0.0
	}
}
