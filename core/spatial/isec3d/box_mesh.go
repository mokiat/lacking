package isec3d

import (
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// CheckBoxMesh reports whether the box intersects the mesh through any of its
// triangles.
//
// Each triangle is tested with [CheckBoxTriangle], so the same two-sided
// convention applies: the result is true as soon as the box intersects one
// triangle. A per-triangle bounding-sphere test is used to skip triangles that
// are too far from the box to possibly intersect it.
func CheckBoxMesh(box shape3d.Box, mesh shape3d.Mesh) bool {
	boundingSphere := box.BoundingSphere()
	for _, triangle := range mesh.Triangles {
		if !CheckSphereSphere(boundingSphere, triangle.BoundingSphere()) {
			continue
		}
		if CheckBoxTriangle(box, triangle) {
			return true
		}
	}
	return false
}

// ResolveBoxMesh yields the contact for the triangle the box penetrates most
// deeply, if it intersects the mesh at all.
//
// Every triangle is resolved with [ResolveBoxTriangle], skipping triangles whose
// bounding sphere the box does not reach, and the resulting contacts are reduced
// to the one with the greatest Depth using a [shape3d.DeepestContact]. The box
// Depth is a true penetration distance, so this selects the deepest overlap. The
// reported [shape3d.Contact] follows the same convention as [ResolveBoxTriangle],
// with the box as the source and the triangle as the target. No contact is
// yielded when the box does not intersect any triangle.
func ResolveBoxMesh(box shape3d.Box, mesh shape3d.Mesh, yield shape3d.ContactCallback) {
	boundingSphere := box.BoundingSphere()
	var deepestContact shape3d.DeepestContact
	for _, triangle := range mesh.Triangles {
		if !CheckSphereSphere(boundingSphere, triangle.BoundingSphere()) {
			continue
		}
		ResolveBoxTriangle(box, triangle, deepestContact.AddContact)
	}
	if contact, ok := deepestContact.Contact(); ok {
		yield(contact)
	}
}

// func ResolveBoxMesh(box shape3d.Box, mesh shape3d.Mesh, yield shape3d.ContactCallback) {
// boxPosition := box.Center
// boxRotation := box.Rotation

// maxX := dprec.Vec3Prod(boxRotation.BasisX, box.HalfWidth)
// minX := dprec.InverseVec3(maxX)
// maxY := dprec.Vec3Prod(boxRotation.BasisY, box.HalfHeight)
// minY := dprec.InverseVec3(maxY)
// maxZ := dprec.Vec3Prod(boxRotation.BasisZ, box.HalfLength)
// minZ := dprec.InverseVec3(maxZ)

// p1 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, minX), minZ), maxY)
// p2 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, minX), maxZ), maxY)
// p3 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, maxX), maxZ), maxY)
// p4 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, maxX), minZ), maxY)
// p5 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, minX), minZ), minY)
// p6 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, minX), maxZ), minY)
// p7 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, maxX), maxZ), minY)
// p8 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, maxX), minZ), minY)

// for _, triangle := range mesh.Triangles {
// 	ResolveSegmentTriangle(shape3d.NewSegment(p1, p2), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p2, p3), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p3, p4), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p4, p1), triangle, yield)

// 	ResolveSegmentTriangle(shape3d.NewSegment(p5, p6), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p6, p7), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p7, p8), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p8, p5), triangle, yield)

// 	ResolveSegmentTriangle(shape3d.NewSegment(p1, p5), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p2, p6), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p3, p7), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p4, p8), triangle, yield)

// 	// since segment intersections are unidirectional, check the opposite direction as well

// 	ResolveSegmentTriangle(shape3d.NewSegment(p2, p1), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p3, p2), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p4, p3), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p1, p4), triangle, yield)

// 	ResolveSegmentTriangle(shape3d.NewSegment(p6, p5), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p7, p6), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p8, p7), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p5, p8), triangle, yield)

// 	ResolveSegmentTriangle(shape3d.NewSegment(p5, p1), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p6, p2), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p7, p3), triangle, yield)
// 	ResolveSegmentTriangle(shape3d.NewSegment(p8, p4), triangle, yield)
// }
// }
