// Package query2d provides a 2D spatial query interface.
//
// The package is built around a [Quadtree], a loose quadtree that indexes items
// by their spatial [Area] and allows them to be searched through
// [Quadtree.QueryAABB] and [Quadtree.QuerySegment].
//
// It is intended as a broad-phase (high-level) pass: queries are conservative
// and may yield false positives, so callers are expected to run their own
// narrow-phase tests on the returned items. It will never omit an item that
// truly matches the query.
//
// Every item is reduced to a center and a half-extent (an axis-aligned
// bounding box). As a result, non-square shapes are indexed by their bounding
// square, which is a deliberate trade-off in favor of speed and simplicity.
package query2d
