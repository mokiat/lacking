// Package isec2d provides 2D intersection tests and contact resolution between
// geometric primitives.
//
// Two contact conventions are used, depending on the shapes involved:
//
//   - Volume-versus-volume resolves (for example [ResolveCircleCircle]) report a
//     mutual-penetration contact: it is expressed relative to the second shape
//     as the target, with the normal pointing from the target toward the source
//     (the first shape) and the depth giving how far the two shapes overlap
//     along it.
//
//   - Segment resolves (for example [ResolveSegmentCircle]) treat the segment as
//     a directed probe from A to B. They report the point at which the segment
//     first enters the shape, the outward surface normal there, and a depth equal
//     to the fraction of the segment lying beyond that entry point: 1 when the
//     segment enters at A and 0 when it enters at B. Expressing the depth as a
//     fraction keeps it comparable across shapes, so that [shape2d.DeepestContact]
//     selects the earliest entry along the segment. The matching segment checks
//     are oriented the same way, so a segment that starts inside the shape is not
//     considered to intersect it.
package isec2d
