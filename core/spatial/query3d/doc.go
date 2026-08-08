// Package query3d provides a 3D spatial query interface.
//
// The package is built around an [Octree], a loose octree that indexes items
// by their axis-aligned bounding box ([shape3d.AABB]) and allows them to be
// searched through [Octree.QueryAABB] and [Octree.QuerySegment].
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
