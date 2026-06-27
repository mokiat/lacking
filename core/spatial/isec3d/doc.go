// Package isec3d provides 3D intersection tests and contact resolution between
// geometric primitives.
//
// Two contact conventions are used, depending on the shapes involved:
//
//   - Volume-versus-volume resolves (for example ResolveSphereSphere) report a
//     mutual-penetration contact: the normal is the axis of least penetration
//     and the depth is how far the two shapes overlap along it.
//
//   - Segment resolves treat the segment as a directed, face-culled probe
//     from A to B. They report the point at which the segment first enters the
//     shape through a front-facing surface, the outward surface normal there,
//     and a depth equal to how far the far endpoint B has travelled past that
//     surface. The matching segment checks are oriented the same way, so a
//     segment that starts inside the shape or that reaches it only through a
//     back-facing surface is not considered to intersect it.
package isec3d
