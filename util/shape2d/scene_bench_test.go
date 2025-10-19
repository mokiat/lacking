package shape2d_test

import (
	"math/rand/v2"
	"testing"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape2d"
)

// ---Default implementation:
// BenchmarkSceneInsert-16     	     444	   2276661 ns/op	13439830 B/op	      49 allocs/op
// BenchmarkSceneInsert-16     	     514	   2259523 ns/op	13439795 B/op	      49 allocs/op
// BenchmarkSceneInsert-16     	     513	   2218677 ns/op	13439805 B/op	      49 allocs/op
// BenchmarkSceneSegment-16    	     181	   6374593 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSceneSegment-16    	     181	   6368127 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSceneSegment-16    	     182	   6335083 ns/op	       0 B/op	       0 allocs/op

// ---Single quadtree:
// BenchmarkSceneInsert-16     	     678	   1805383 ns/op	 5050413 B/op	      33 allocs/op
// BenchmarkSceneInsert-16     	     660	   2162179 ns/op	 5050402 B/op	      33 allocs/op
// BenchmarkSceneInsert-16     	     750	   1868836 ns/op	 5050397 B/op	      33 allocs/op
// BenchmarkSceneSegment-16    	     157	   7372762 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSceneSegment-16    	     158	   7346010 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSceneSegment-16    	     166	   7123698 ns/op	       0 B/op	       0 allocs/op

// ---New sorting:
// BenchmarkSceneInsert-16     	     727	   1794449 ns/op	 5050425 B/op	      33 allocs/op
// BenchmarkSceneInsert-16     	     686	   1942115 ns/op	 5050420 B/op	      33 allocs/op
// BenchmarkSceneInsert-16     	     726	   1644388 ns/op	 5050413 B/op	      33 allocs/op
// BenchmarkSceneSegment-16    	     178	   6446904 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSceneSegment-16    	     190	   6209953 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSceneSegment-16    	     177	   6496010 ns/op	       0 B/op	       0 allocs/op

// ---Latest:
// BenchmarkSceneInsert-16     	     775	   1488028 ns/op	 5574694 B/op	      33 allocs/op
// BenchmarkSceneInsert-16     	     591	   1989371 ns/op	 5574708 B/op	      33 allocs/op
// BenchmarkSceneInsert-16     	     669	   1943611 ns/op	 5574699 B/op	      33 allocs/op
// BenchmarkSceneSegment-16    	     180	   6431796 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSceneSegment-16    	     186	   6397599 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSceneSegment-16    	     184	   6414405 ns/op	       0 B/op	       0 allocs/op

func BenchmarkSceneInsert(b *testing.B) {
	random := rand.New(rand.NewPCG(127, 63))

	const itemCount = 1024
	for b.Loop() {
		scene := shape2d.NewScene[struct{}, struct{}](shape2d.SceneSettings{
			Size:                opt.V(1024.0),
			MaxDepth:            opt.V[uint32](8),
			InitialNodeCapacity: opt.V[uint32](1024 * 64),
			InitialItemCapacity: opt.V[uint32](1024 * 16),
		})

		for range itemCount {
			objID := scene.CreateObject(shape2d.ObjectInfo[struct{}]{
				Position: opt.V(dprec.NewVec2(
					(random.Float64()*2.0-1.0)*511.0,
					(random.Float64()*2.0-1.0)*511.0,
				)),
				Rotation: opt.V(dprec.Degrees(random.Float64() * 360.0)),
				Static:   false,
				UserData: struct{}{},
			})
			scene.AttachCircle(objID, shape2d.CircleInfo[struct{}]{
				ShapeInfo: shape2d.ShapeInfo[struct{}]{},
				Circle: shape2d.NewCircle(
					dprec.NewVec2(
						(random.Float64()*2.0-1.0)*4.0,
						(random.Float64()*2.0-1.0)*4.0,
					),
					random.Float64()*2.0+0.1,
				),
			})
			scene.AttachRectangle(objID, shape2d.RectangleInfo[struct{}]{
				ShapeInfo: shape2d.ShapeInfo[struct{}]{},
				Rectangle: shape2d.NewRectangle(
					dprec.NewVec2(
						(random.Float64()*2.0-1.0)*4.0,
						(random.Float64()*2.0-1.0)*4.0,
					),
					dprec.Degrees(random.Float64()*360.0),
					dprec.NewVec2(
						(random.Float64()*2.0-1.0)*2.0+0.1,
						(random.Float64()*2.0-1.0)*2.0+0.1,
					),
				),
			})
		}
	}
}

func BenchmarkSceneSegment(b *testing.B) {
	random := rand.New(rand.NewPCG(127, 63))

	scene := shape2d.NewScene[struct{}, struct{}](shape2d.SceneSettings{
		Size:                opt.V(1024.0),
		MaxDepth:            opt.V[uint32](8),
		InitialNodeCapacity: opt.V[uint32](1024 * 64),
		InitialItemCapacity: opt.V[uint32](1024 * 16),
	})

	const itemCount = 1024
	for i := range itemCount {
		objID := scene.CreateObject(shape2d.ObjectInfo[struct{}]{
			Position: opt.V(dprec.NewVec2(
				(random.Float64()*2.0-1.0)*511.0,
				(random.Float64()*2.0-1.0)*511.0,
			)),
			Rotation: opt.V(dprec.Degrees(random.Float64() * 360.0)),
			Static:   false,
			UserData: struct{}{},
		})
		if i%2 == 0 {
			scene.AttachCircle(objID, shape2d.CircleInfo[struct{}]{
				ShapeInfo: shape2d.ShapeInfo[struct{}]{},
				Circle: shape2d.NewCircle(
					dprec.NewVec2(
						(random.Float64()*2.0-1.0)*4.0,
						(random.Float64()*2.0-1.0)*4.0,
					),
					random.Float64()*64.0+0.1,
				),
			})
		} else {
			scene.AttachRectangle(objID, shape2d.RectangleInfo[struct{}]{
				ShapeInfo: shape2d.ShapeInfo[struct{}]{},
				Rectangle: shape2d.NewRectangle(
					dprec.NewVec2(
						(random.Float64()*2.0-1.0)*4.0,
						(random.Float64()*2.0-1.0)*4.0,
					),
					dprec.Degrees(random.Float64()*360.0),
					dprec.NewVec2(
						(random.Float64()*2.0-1.0)*64.0+0.1,
						(random.Float64()*2.0-1.0)*64.0+0.1,
					),
				),
			})
		}
	}

	var collection shape2d.NearestObjectIntersection
	for b.Loop() {
		for range 1024 {
			segment := shape2d.NewSegment(
				dprec.NewVec2(
					(random.Float64()*2.0-1.0)*511.0,
					(random.Float64()*2.0-1.0)*511.0,
				),
				dprec.NewVec2(
					(random.Float64()*2.0-1.0)*511.0,
					(random.Float64()*2.0-1.0)*511.0,
				),
			)
			scene.CollectSegmentIntersections(segment, 0xFFFFFFFF, &collection)
		}
	}
}
