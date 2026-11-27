// Package query3d provides a 3D spatial query interface.
//
// The package provides two interchangeable structures that index items by
// their axis-aligned bounding box ([shape3d.AABB]): a [BVHTree], which is a
// dynamic bounding volume hierarchy without fixed bounds, and an [Octree],
// which is a loose octree of a fixed size. Both allow items to be searched by
// box, by segment and by frustum (for example [BVHTree.QueryAABB],
// [BVHTree.QuerySegment] and [BVHTree.QueryFrustum]).
//
// It is intended as a broad-phase (high-level) pass: queries are conservative
// and may yield false positives, so callers are expected to run their own
// narrow-phase tests on the returned items. It will never omit an item that
// truly matches the query.
//
// Every item is reduced to an axis-aligned bounding box, which means that
// orientation and concavity are not taken into account. Callers get the
// tightest results by passing the smallest box that still fully encompasses
// the item.
package query3d
