// Package query3d provides a 3D spatial query interface.
//
// The package offers two interchangeable structures that index items by their
// spatial [Area] and allow them to be searched through a QueryAABB and a
// QuerySegment method: an [Octree] and a [BVHTree]. Both expose the same set of
// methods and return the same items for a given query, so switching from one to
// the other is a matter of changing the constructor and the settings type.
//
//   - An [Octree] is a loose octree with shrinking bounding boxes. It needs to
//     know the size of the scene up front and it defers most of the work of a
//     modification to the next query. It is the better choice for content that
//     is mostly static and evenly spread over a known volume.
//
//   - A [BVHTree] is a dynamic bounding volume hierarchy. It has neither fixed
//     bounds nor a maximum depth and it reorganizes itself as items are added,
//     which makes it the better choice for scenes that are large, sparse or
//     unbounded, and for scenes in which items move frequently. It also uses
//     substantially less memory.
//
// The package is intended as a broad-phase (high-level) pass: queries are
// conservative and may yield false positives, so callers are expected to run
// their own narrow-phase tests on the returned items. Neither structure will
// ever omit an item that truly matches the query.
//
// Every item is reduced to a center and a half-extent (an axis-aligned
// bounding box). As a result, non-cubic shapes are indexed by their bounding
// cube, which is a deliberate trade-off in favor of speed and simplicity.
package query3d
