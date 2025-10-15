package shape2d_test

import (
	"math/rand/v2"
	"testing"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape2d"
)

// Default implementation:
// BenchmarkSceneInsert-16     	     444	   2276661 ns/op	13439830 B/op	      49 allocs/op
// BenchmarkSceneInsert-16     	     514	   2259523 ns/op	13439795 B/op	      49 allocs/op
// BenchmarkSceneInsert-16     	     513	   2218677 ns/op	13439805 B/op	      49 allocs/op
// BenchmarkSceneSegment-16    	     181	   6374593 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSceneSegment-16    	     181	   6368127 ns/op	       0 B/op	       0 allocs/op
// BenchmarkSceneSegment-16    	     182	   6335083 ns/op	       0 B/op	       0 allocs/op

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
