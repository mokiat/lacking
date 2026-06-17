package shape2d_test

import (
	"math/rand/v2"
	"testing"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape2d"
)

////////////////// QUERY2D PACKAGE (QUADTREE)

// goos: linux
// goarch: amd64
// pkg: github.com/mokiat/lacking/util/shape2d
// cpu: AMD Ryzen 7 3700X 8-Core Processor
// BenchmarkSceneInsert-16     	     652	   1892388 ns/op	10293445 B/op	      37 allocs/op
// BenchmarkSceneSegment-16    	     195	   6067209 ns/op	       0 B/op	       0 allocs/op
// PASS
// ok  	github.com/mokiat/lacking/util/shape2d	2.432s

////////////////// QUERY2D PACKAGE (BVH TREE)

// goos: linux
// goarch: amd64
// pkg: github.com/mokiat/lacking/util/shape2d
// cpu: AMD Ryzen 7 3700X 8-Core Processor
// BenchmarkSceneInsert-16     	     578	   2176115 ns/op	 1348202 B/op	      37 allocs/op
// BenchmarkSceneSegment-16    	     160	   7427997 ns/op	       0 B/op	       0 allocs/op
// PASS
// ok  	github.com/mokiat/lacking/util/shape2d	2.457s

func BenchmarkSceneInsert(b *testing.B) {
	random := rand.New(rand.NewPCG(127, 63))

	const itemCount = 1024
	for b.Loop() {
		scene := shape2d.NewScene[struct{}, struct{}](shape2d.SceneSettings{
			Size:                opt.V(float32(1024.0)),
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
		Size:                opt.V(float32(1024.0)),
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

	var collection shape2d.SmallestObjectIntersection
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
			scene.CollectSegmentIntersections(segment, shape2d.Filter{}, &collection)
		}
	}
}
