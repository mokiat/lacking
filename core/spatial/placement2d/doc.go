// Package placement2d provides a 2D scene in which objects and terrains can be
// placed and tested for intersection.
//
// An object is a movable entity that owns one or more convex shapes (see
// [Scene.AttachCircle] and [Scene.AttachRectangle]). Moving the object through
// [Scene.SetObjectTransform] moves all of its shapes along with it.
//
// A terrain is an immovable entity that owns one or more concave shapes (see
// [Scene.AttachMesh]). Terrain shapes are specified directly in world space and
// cannot be relocated once attached.
//
// Object shapes and terrain shapes are indexed in separate quadtrees for
// efficient broad-phase queries. Narrow-phase intersection is resolved via
// GJK/EPA (see [github.com/mokiat/lacking/core/spatial/gjk2d]), with concave
// shapes being decomposed into convex pieces beforehand.
//
// Intersections come in two flavors, which are reported through separate
// methods so that callers never need to branch on the kind of the target. An
// intersection with an object shape is reported as an [ObjectContact] through
// an [ObjectContactCallback], whereas an intersection with a terrain shape is
// reported as a [TerrainContact] through a [TerrainContactCallback]. A number
// of contact sinks (for example [DeepestObjectContact] and
// [TerrainContactList]) are provided for common accumulation strategies.
//
// Since terrains cannot move, terrain shapes are never tested against one
// another. The source of a [TerrainContact] is always either an object shape
// or a query primitive.
package placement2d
