// Package placement3d provides a 3D scene in which objects, built from convex
// shapes, and static meshes can be placed and tested for intersection.
//
// Objects are dynamic entities that own one or more convex shapes (spheres and
// boxes). Meshes are static entities made of triangles. Both are indexed in
// separate octrees for efficient broad-phase queries, and narrow-phase
// intersection is resolved via GJK/EPA (see [github.com/mokiat/lacking/core/spatial/gjk3d]).
//
// Intersections are reported as [Contact] values through a [ContactCallback].
// A number of contact sinks (for example [DeepestContact] and [ContactList])
// are provided for common accumulation strategies.
package placement3d
