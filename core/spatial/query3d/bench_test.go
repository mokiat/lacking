package query3d_test

import (
	"math/rand/v2"
	"testing"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/core/spatial/query3d"
)

// This file compares the two spatial structures of this package on the same
// workloads.
//
// A note on the Insert benchmarks: an Octree defers almost all of the work of
// an insertion to the next query, through its internal refresh mechanism,
// whereas a BVHTree performs its work eagerly. A benchmark that only inserts
// therefore charges the BVHTree for work that the Octree still owes. The Build
// benchmarks add a single query and are the fair measure of how long it takes
// to reach a structure that can actually be queried.
//
// Results, recorded with:
//
//	go test -run='^$' -bench=. -benchmem -count=5 ./core/spatial/query3d
//
// reporting the median of the five runs:
//
//	goos: linux
//	goarch: amd64
//	pkg: github.com/mokiat/lacking/core/spatial/query3d
//	cpu: AMD Ryzen 7 3700X 8-Core Processor
//
//	BenchmarkOctreeInsert-16           7670684 ns/op   22102370 B/op   12 allocs/op
//	BenchmarkBVHTreeInsert-16         17764741 ns/op    1933459 B/op    5 allocs/op
//	BenchmarkOctreeBuild-16           12793078 ns/op   22102208 B/op   12 allocs/op
//	BenchmarkBVHTreeBuild-16          18994028 ns/op    1933463 B/op    5 allocs/op
//	BenchmarkOctreeInsertSorted-16    18272879 ns/op   29016275 B/op   13 allocs/op
//	BenchmarkBVHTreeInsertSorted-16   15555542 ns/op    1933457 B/op    5 allocs/op
//	BenchmarkOctreeUpdateSmall-16      1155916 ns/op          0 B/op    0 allocs/op
//	BenchmarkBVHTreeUpdateSmall-16       13136 ns/op          0 B/op    0 allocs/op
//	BenchmarkOctreeUpdateLarge-16      1152900 ns/op          0 B/op    0 allocs/op
//	BenchmarkBVHTreeUpdateLarge-16       13211 ns/op          0 B/op    0 allocs/op
//	BenchmarkOctreeRemove-16           2172561 ns/op     814832 B/op   39 allocs/op
//	BenchmarkBVHTreeRemove-16          2112105 ns/op     511728 B/op   37 allocs/op
//	BenchmarkOctreeQueryAABBSmall-16    142805 ns/op          0 B/op    0 allocs/op
//	BenchmarkBVHTreeQueryAABBSmall-16    70986 ns/op          0 B/op    0 allocs/op
//	BenchmarkOctreeQueryAABBLarge-16  98493071 ns/op          0 B/op    0 allocs/op
//	BenchmarkBVHTreeQueryAABBLarge-16 29547685 ns/op          0 B/op    0 allocs/op
//	BenchmarkOctreeQuerySegment-16      686385 ns/op          0 B/op    0 allocs/op
//	BenchmarkBVHTreeQuerySegment-16     619535 ns/op          0 B/op    0 allocs/op
//	BenchmarkOctreeBroadphase-16       4717283 ns/op          0 B/op    0 allocs/op
//	BenchmarkBVHTreeBroadphase-16      3848306 ns/op          0 B/op    0 allocs/op
//
// Speed of the BVHTree relative to the Octree, where a value above one means
// that the BVHTree is faster:
//
//	Insert           0.43x     QueryAABBSmall    2.01x
//	Build            0.67x     QueryAABBLarge    3.33x
//	InsertSorted     1.17x     QuerySegment      1.11x
//	UpdateSmall     88.00x     Broadphase        1.23x
//	UpdateLarge     87.27x     Remove            1.03x
//
// The BVHTree pays for its quality up front, during insertion, and earns it
// back on every query and every update afterwards. It also holds the same
// items in roughly a tenth of the memory.

const (
	benchItemCount   = 10000
	benchSceneRadius = 1024.0
	benchQueryCount  = 256
	benchUpdateCount = 1000

	// benchSmallStep stays well within the default margin of a BVHTree, which
	// is the case that a physics engine hits on most of its frames.
	benchSmallStep = 0.05

	// benchLargeStep is far beyond any reasonable margin and forces a BVHTree
	// to relocate the item.
	benchLargeStep = 50.0
)

func benchRand() *rand.Rand {
	return rand.New(rand.NewPCG(0x1234ABCD, 0x5678EF01))
}

func newBenchOctree() *query3d.Octree[int] {
	return query3d.NewOctree[int](query3d.OctreeSettings{
		Size:                opt.V(4.0 * benchSceneRadius),
		MaxDepth:            opt.V[uint32](10),
		InitialNodeCapacity: opt.V[uint32](benchItemCount),
		InitialItemCapacity: opt.V[uint32](benchItemCount),
	})
}

func newBenchBVHTree() *query3d.BVHTree[int] {
	return query3d.NewBVHTree[int](query3d.BVHTreeSettings{
		AABBMargin:          opt.V(0.5),
		InitialNodeCapacity: opt.V[uint32](2 * benchItemCount),
		InitialItemCapacity: opt.V[uint32](benchItemCount),
	})
}

// benchSphere is the source form of a benchmark item. An [query3d.Area] hides
// its center and radius, so they are tracked separately in order to be able to
// derive displaced variants.
type benchSphere struct {
	center dprec.Vec3
	radius float64
}

func (s benchSphere) area() query3d.Area {
	return areaFromSphere(s.center.X, s.center.Y, s.center.Z, s.radius)
}

// benchSpheres returns randomly scattered items.
func benchSpheres() []benchSphere {
	random := benchRand()
	spheres := make([]benchSphere, benchItemCount)
	for i := range spheres {
		spheres[i] = benchSphere{
			center: dprec.NewVec3(
				random.Float64()*2.0*benchSceneRadius-benchSceneRadius,
				random.Float64()*2.0*benchSceneRadius-benchSceneRadius,
				random.Float64()*2.0*benchSceneRadius-benchSceneRadius,
			),
			radius: random.Float64()*8.0 + 0.5,
		}
	}
	return spheres
}

// benchAreas returns randomly scattered areas.
func benchAreas() []query3d.Area {
	spheres := benchSpheres()
	areas := make([]query3d.Area, len(spheres))
	for i, sphere := range spheres {
		areas[i] = sphere.area()
	}
	return areas
}

// benchSortedAreas returns areas along a scan order grid, which is the worst
// case for a hierarchy that is built through insertion without rebalancing.
func benchSortedAreas() []query3d.Area {
	const side = 22 // 22*22*22 is just above benchItemCount
	step := 2.0 * benchSceneRadius / float64(side)

	areas := make([]query3d.Area, 0, benchItemCount)
	for x := range side {
		for y := range side {
			for z := range side {
				if len(areas) == benchItemCount {
					return areas
				}
				areas = append(areas, areaFromSphere(
					float64(x)*step-benchSceneRadius,
					float64(y)*step-benchSceneRadius,
					float64(z)*step-benchSceneRadius,
					1.0,
				))
			}
		}
	}
	return areas
}

// benchMovedAreas returns, for each of the first benchUpdateCount items, a
// variant that is displaced by the specified distance in a random direction.
func benchMovedAreas(spheres []benchSphere, step float64) []query3d.Area {
	random := rand.New(rand.NewPCG(0x0FF5E7, 0x0FF5E7))
	moved := make([]query3d.Area, benchUpdateCount)
	for i := range moved {
		offset := dprec.ResizedVec3(dprec.NewVec3(
			random.Float64()*2.0-1.0,
			random.Float64()*2.0-1.0,
			random.Float64()*2.0-1.0,
		), step)
		sphere := spheres[i]
		sphere.center = dprec.Vec3Sum(sphere.center, offset)
		moved[i] = sphere.area()
	}
	return moved
}

func benchAABBs(radius float64) []query3d.AABB {
	random := benchRand()
	boxes := make([]query3d.AABB, benchQueryCount)
	for i := range boxes {
		boxes[i] = aabbFromSphere(
			random.Float64()*2.0*benchSceneRadius-benchSceneRadius,
			random.Float64()*2.0*benchSceneRadius-benchSceneRadius,
			random.Float64()*2.0*benchSceneRadius-benchSceneRadius,
			radius,
		)
	}
	return boxes
}

func benchSegments() []query3d.Segment {
	random := benchRand()
	point := func() dprec.Vec3 {
		return dprec.NewVec3(
			random.Float64()*2.0*benchSceneRadius-benchSceneRadius,
			random.Float64()*2.0*benchSceneRadius-benchSceneRadius,
			random.Float64()*2.0*benchSceneRadius-benchSceneRadius,
		)
	}
	segments := make([]query3d.Segment, benchQueryCount)
	for i := range segments {
		segments[i] = query3d.NewSegment(point(), point())
	}
	return segments
}

// benchItemAABBs returns the bounding box of every item, which is what a
// broadphase pass queries with.
func benchItemAABBs() []query3d.AABB {
	random := benchRand()
	boxes := make([]query3d.AABB, benchItemCount)
	for i := range boxes {
		boxes[i] = aabbFromSphere(
			random.Float64()*2.0*benchSceneRadius-benchSceneRadius,
			random.Float64()*2.0*benchSceneRadius-benchSceneRadius,
			random.Float64()*2.0*benchSceneRadius-benchSceneRadius,
			random.Float64()*8.0+0.5,
		)
	}
	return boxes
}

func BenchmarkOctreeInsert(b *testing.B) {
	areas := benchAreas()
	var tree *query3d.Octree[int]
	b.ReportAllocs()
	for b.Loop() {
		tree = newBenchOctree()
		for i, area := range areas {
			tree.Insert(area, i)
		}
	}
	if tree.Stats().ItemCount != benchItemCount {
		b.Fatal("unexpected item count")
	}
}

func BenchmarkBVHTreeInsert(b *testing.B) {
	areas := benchAreas()
	var tree *query3d.BVHTree[int]
	b.ReportAllocs()
	for b.Loop() {
		tree = newBenchBVHTree()
		for i, area := range areas {
			tree.Insert(area, i)
		}
	}
	if tree.Stats().ItemCount != benchItemCount {
		b.Fatal("unexpected item count")
	}
}

func BenchmarkOctreeBuild(b *testing.B) {
	areas := benchAreas()
	aabb := aabbFromSphere(0.0, 0.0, 0.0, benchSceneRadius*4.0)
	var total int
	visit := func(item int) bool {
		total++
		return true
	}
	b.ReportAllocs()
	for b.Loop() {
		tree := newBenchOctree()
		for i, area := range areas {
			tree.Insert(area, i)
		}
		tree.QueryAABB(aabb, visit)
	}
	if total == 0 {
		b.Fatal("no items were visited")
	}
}

func BenchmarkBVHTreeBuild(b *testing.B) {
	areas := benchAreas()
	aabb := aabbFromSphere(0.0, 0.0, 0.0, benchSceneRadius*4.0)
	var total int
	visit := func(item int) bool {
		total++
		return true
	}
	b.ReportAllocs()
	for b.Loop() {
		tree := newBenchBVHTree()
		for i, area := range areas {
			tree.Insert(area, i)
		}
		tree.QueryAABB(aabb, visit)
	}
	if total == 0 {
		b.Fatal("no items were visited")
	}
}

func BenchmarkOctreeInsertSorted(b *testing.B) {
	areas := benchSortedAreas()
	aabb := aabbFromSphere(0.0, 0.0, 0.0, benchSceneRadius*4.0)
	var total int
	visit := func(item int) bool {
		total++
		return true
	}
	b.ReportAllocs()
	for b.Loop() {
		tree := newBenchOctree()
		for i, area := range areas {
			tree.Insert(area, i)
		}
		tree.QueryAABB(aabb, visit)
	}
	if total == 0 {
		b.Fatal("no items were visited")
	}
}

func BenchmarkBVHTreeInsertSorted(b *testing.B) {
	areas := benchSortedAreas()
	aabb := aabbFromSphere(0.0, 0.0, 0.0, benchSceneRadius*4.0)
	var total int
	visit := func(item int) bool {
		total++
		return true
	}
	b.ReportAllocs()
	for b.Loop() {
		tree := newBenchBVHTree()
		for i, area := range areas {
			tree.Insert(area, i)
		}
		tree.QueryAABB(aabb, visit)
	}
	if total == 0 {
		b.Fatal("no items were visited")
	}
}

func benchmarkOctreeUpdate(b *testing.B, step float64) {
	spheres := benchSpheres()
	areas := benchAreas()
	moved := benchMovedAreas(spheres, step)
	aabb := aabbFromSphere(0.0, 0.0, 0.0, 64.0)
	var total int
	visit := func(item int) bool {
		total++
		return true
	}

	tree := newBenchOctree()
	ids := make([]query3d.TreeItemID, len(areas))
	for i, area := range areas {
		ids[i] = tree.Insert(area, i)
	}

	b.ReportAllocs()
	for b.Loop() {
		for i, area := range moved {
			tree.Update(ids[i], area)
		}
		// A query is what forces an Octree to pay for the updates above.
		tree.QueryAABB(aabb, visit)
	}
	if tree.Stats().ItemCount != benchItemCount {
		b.Fatal("unexpected item count")
	}
}

func benchmarkBVHTreeUpdate(b *testing.B, step float64) {
	spheres := benchSpheres()
	areas := benchAreas()
	moved := benchMovedAreas(spheres, step)
	aabb := aabbFromSphere(0.0, 0.0, 0.0, 64.0)
	var total int
	visit := func(item int) bool {
		total++
		return true
	}

	tree := newBenchBVHTree()
	ids := make([]query3d.TreeItemID, len(areas))
	for i, area := range areas {
		ids[i] = tree.Insert(area, i)
	}

	b.ReportAllocs()
	for b.Loop() {
		for i, area := range moved {
			tree.Update(ids[i], area)
		}
		tree.QueryAABB(aabb, visit)
	}
	if tree.Stats().ItemCount != benchItemCount {
		b.Fatal("unexpected item count")
	}
}

func BenchmarkOctreeUpdateSmall(b *testing.B) {
	benchmarkOctreeUpdate(b, benchSmallStep)
}

func BenchmarkBVHTreeUpdateSmall(b *testing.B) {
	benchmarkBVHTreeUpdate(b, benchSmallStep)
}

func BenchmarkOctreeUpdateLarge(b *testing.B) {
	benchmarkOctreeUpdate(b, benchLargeStep)
}

func BenchmarkBVHTreeUpdateLarge(b *testing.B) {
	benchmarkBVHTreeUpdate(b, benchLargeStep)
}

func BenchmarkOctreeRemove(b *testing.B) {
	areas := benchAreas()
	ids := make([]query3d.TreeItemID, len(areas))

	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		tree := newBenchOctree()
		for i, area := range areas {
			ids[i] = tree.Insert(area, i)
		}
		b.StartTimer()

		for _, id := range ids {
			tree.Remove(id)
		}
	}
}

func BenchmarkBVHTreeRemove(b *testing.B) {
	areas := benchAreas()
	ids := make([]query3d.TreeItemID, len(areas))

	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		tree := newBenchBVHTree()
		for i, area := range areas {
			ids[i] = tree.Insert(area, i)
		}
		b.StartTimer()

		for _, id := range ids {
			tree.Remove(id)
		}
	}
}

func BenchmarkOctreeQueryAABBSmall(b *testing.B) {
	benchmarkOctreeQueryAABB(b, benchAABBs(32.0))
}

func BenchmarkBVHTreeQueryAABBSmall(b *testing.B) {
	benchmarkBVHTreeQueryAABB(b, benchAABBs(32.0))
}

func BenchmarkOctreeQueryAABBLarge(b *testing.B) {
	benchmarkOctreeQueryAABB(b, benchAABBs(benchSceneRadius))
}

func BenchmarkBVHTreeQueryAABBLarge(b *testing.B) {
	benchmarkBVHTreeQueryAABB(b, benchAABBs(benchSceneRadius))
}

func benchmarkOctreeQueryAABB(b *testing.B, boxes []query3d.AABB) {
	tree := newBenchOctree()
	for i, area := range benchAreas() {
		tree.Insert(area, i)
	}
	tree.Stats() // settle the tree before measuring

	var total int
	visit := func(item int) bool {
		total++
		return true
	}

	b.ReportAllocs()
	for b.Loop() {
		for _, aabb := range boxes {
			tree.QueryAABB(aabb, visit)
		}
	}
	if total == 0 {
		b.Fatal("no items were visited")
	}
}

func benchmarkBVHTreeQueryAABB(b *testing.B, boxes []query3d.AABB) {
	tree := newBenchBVHTree()
	for i, area := range benchAreas() {
		tree.Insert(area, i)
	}

	var total int
	visit := func(item int) bool {
		total++
		return true
	}

	b.ReportAllocs()
	for b.Loop() {
		for _, aabb := range boxes {
			tree.QueryAABB(aabb, visit)
		}
	}
	if total == 0 {
		b.Fatal("no items were visited")
	}
}

func BenchmarkOctreeQuerySegment(b *testing.B) {
	segments := benchSegments()
	tree := newBenchOctree()
	for i, area := range benchAreas() {
		tree.Insert(area, i)
	}
	tree.Stats() // settle the tree before measuring

	var total int
	visit := func(item int) bool {
		total++
		return true
	}

	b.ReportAllocs()
	for b.Loop() {
		for _, segment := range segments {
			tree.QuerySegment(segment, visit)
		}
	}
	if total == 0 {
		b.Fatal("no items were visited")
	}
}

func BenchmarkBVHTreeQuerySegment(b *testing.B) {
	segments := benchSegments()
	tree := newBenchBVHTree()
	for i, area := range benchAreas() {
		tree.Insert(area, i)
	}

	var total int
	visit := func(item int) bool {
		total++
		return true
	}

	b.ReportAllocs()
	for b.Loop() {
		for _, segment := range segments {
			tree.QuerySegment(segment, visit)
		}
	}
	if total == 0 {
		b.Fatal("no items were visited")
	}
}

func BenchmarkOctreeBroadphase(b *testing.B) {
	boxes := benchItemAABBs()
	tree := newBenchOctree()
	for i, area := range benchAreas() {
		tree.Insert(area, i)
	}
	tree.Stats() // settle the tree before measuring

	var total int
	visit := func(item int) bool {
		total++
		return true
	}

	b.ReportAllocs()
	for b.Loop() {
		for _, aabb := range boxes {
			tree.QueryAABB(aabb, visit)
		}
	}
	if total == 0 {
		b.Fatal("no items were visited")
	}
}

func BenchmarkBVHTreeBroadphase(b *testing.B) {
	boxes := benchItemAABBs()
	tree := newBenchBVHTree()
	for i, area := range benchAreas() {
		tree.Insert(area, i)
	}

	var total int
	visit := func(item int) bool {
		total++
		return true
	}

	b.ReportAllocs()
	for b.Loop() {
		for _, aabb := range boxes {
			tree.QueryAABB(aabb, visit)
		}
	}
	if total == 0 {
		b.Fatal("no items were visited")
	}
}
