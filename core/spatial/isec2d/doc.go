// Package isec2d provides 2D intersection tests and contact resolution between
// geometric primitives.
//
// Volume-versus-volume resolves (for example [ResolveCircleCircle]) report a
// mutual-penetration contact: it is expressed relative to the second shape as
// the target, with the normal pointing from the target toward the source (the
// first shape) and the depth giving how far the two shapes overlap along it.
package isec2d
