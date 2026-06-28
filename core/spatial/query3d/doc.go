// Package query3d provides a 3D spatial query interface.
//
// The package is built around an [Octree], a loose octree that indexes items by
// their spatial [Area] and allows them to be searched through [Octree.QueryAABB]
// and [Octree.QuerySegment].
//
// It is intended as a broad-phase (high-level) pass: queries are conservative
// and may yield false positives, so callers are expected to run their own
// narrow-phase tests on the returned items. It will never omit an item that
// truly matches the query.
//
// Every item is reduced to a center and a half-extent (an axis-aligned
// bounding box). As a result, non-cubic shapes are indexed by their bounding
// cube, which is a deliberate trade-off in favor of speed and simplicity.
package query3d
