package shape3d_test

import (
	"math/rand/v2"
	"testing"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape3d"
)

////////////////// QUERY3D PACKAGE (OCTREE)

// goos: linux
// goarch: amd64
// pkg: github.com/mokiat/lacking/util/shape3d
// cpu: AMD Ryzen 7 3700X 8-Core Processor
// BenchmarkSceneInsert-16     	     622	   2240122 ns/op	 3447046 B/op	      44 allocs/op
// BenchmarkSceneSegment-16    	     538	   2185813 ns/op	       0 B/op	       0 allocs/op
// PASS
// ok  	github.com/mokiat/lacking/util/shape3d	2.580s

////////////////// QUERY3D PACKAGE (BVH TREE)

// goos: linux
// goarch: amd64
// pkg: github.com/mokiat/lacking/util/shape3d
// cpu: AMD Ryzen 7 3700X 8-Core Processor
// BenchmarkSceneInsert-16     	     224	   5346420 ns/op	 1908191 B/op	      38 allocs/op
// BenchmarkSceneSegment-16    	     370	   3042376 ns/op	       0 B/op	       0 allocs/op
// PASS
// ok  	github.com/mokiat/lacking/util/shape3d	2.334s

func BenchmarkSceneInsert(b *testing.B) {
	random := rand.New(rand.NewPCG(127, 63))

	const itemCount = 1024
	for b.Loop() {
		scene := shape3d.NewScene[struct{}, struct{}](shape3d.SceneSettings{
			Size:                opt.V(1024.0),
			MaxDepth:            opt.V[uint32](8),
			InitialNodeCapacity: opt.V[uint32](1024 * 64),
			InitialItemCapacity: opt.V[uint32](1024 * 16),
		})

		for range itemCount {
			objID := scene.CreateObject(shape3d.ObjectInfo[struct{}]{
				Position: opt.V(dprec.NewVec3(
					(random.Float64()*2.0-1.0)*511.0,
					(random.Float64()*2.0-1.0)*511.0,
					(random.Float64()*2.0-1.0)*511.0,
				)),
				Rotation: opt.V(dprec.RotationQuat(
					dprec.Degrees(random.Float64()*360.0),
					dprec.NewVec3(0.0, 1.0, 0.0),
				)),
				Static:   false,
				UserData: struct{}{},
			})
			scene.AttachSphere(objID, shape3d.SphereInfo[struct{}]{
				ShapeInfo: shape3d.ShapeInfo[struct{}]{},
				Sphere: shape3d.NewSphere(
					dprec.NewVec3(
						(random.Float64()*2.0-1.0)*4.0,
						(random.Float64()*2.0-1.0)*4.0,
						(random.Float64()*2.0-1.0)*4.0,
					),
					random.Float64()*2.0+0.1,
				),
			})
			scene.AttachBox(objID, shape3d.BoxInfo[struct{}]{
				ShapeInfo: shape3d.ShapeInfo[struct{}]{},
				Box: shape3d.NewBox(
					dprec.NewVec3(
						(random.Float64()*2.0-1.0)*4.0,
						(random.Float64()*2.0-1.0)*4.0,
						(random.Float64()*2.0-1.0)*4.0,
					),
					dprec.RotationQuat(
						dprec.Degrees(random.Float64()*360.0),
						dprec.NewVec3(0.0, 1.0, 0.0),
					),
					dprec.NewVec3(
						(random.Float64()*2.0-1.0)*2.0+0.1,
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

	scene := shape3d.NewScene[struct{}, struct{}](shape3d.SceneSettings{
		Size:                opt.V(1024.0),
		MaxDepth:            opt.V[uint32](8),
		InitialNodeCapacity: opt.V[uint32](1024 * 64),
		InitialItemCapacity: opt.V[uint32](1024 * 16),
	})

	const itemCount = 1024
	for i := range itemCount {
		objID := scene.CreateObject(shape3d.ObjectInfo[struct{}]{
			Position: opt.V(dprec.NewVec3(
				(random.Float64()*2.0-1.0)*511.0,
				(random.Float64()*2.0-1.0)*511.0,
				(random.Float64()*2.0-1.0)*511.0,
			)),
			Rotation: opt.V(dprec.RotationQuat(
				dprec.Degrees(random.Float64()*360.0),
				dprec.NewVec3(0.0, 1.0, 0.0),
			)),
			Static:   false,
			UserData: struct{}{},
		})
		if i%2 == 0 {
			scene.AttachSphere(objID, shape3d.SphereInfo[struct{}]{
				ShapeInfo: shape3d.ShapeInfo[struct{}]{},
				Sphere: shape3d.NewSphere(
					dprec.NewVec3(
						(random.Float64()*2.0-1.0)*4.0,
						(random.Float64()*2.0-1.0)*4.0,
						(random.Float64()*2.0-1.0)*4.0,
					),
					random.Float64()*64.0+0.1,
				),
			})
		} else {
			scene.AttachBox(objID, shape3d.BoxInfo[struct{}]{
				ShapeInfo: shape3d.ShapeInfo[struct{}]{},
				Box: shape3d.NewBox(
					dprec.NewVec3(
						(random.Float64()*2.0-1.0)*4.0,
						(random.Float64()*2.0-1.0)*4.0,
						(random.Float64()*2.0-1.0)*4.0,
					),
					dprec.RotationQuat(
						dprec.Degrees(random.Float64()*360.0),
						dprec.NewVec3(0.0, 1.0, 0.0),
					),
					dprec.NewVec3(
						(random.Float64()*2.0-1.0)*64.0+0.1,
						(random.Float64()*2.0-1.0)*64.0+0.1,
						(random.Float64()*2.0-1.0)*64.0+0.1,
					),
				),
			})
		}
	}

	var collection shape3d.SmallestObjectIntersection
	for b.Loop() {
		for range 1024 {
			segment := shape3d.NewSegment(
				dprec.NewVec3(
					(random.Float64()*2.0-1.0)*511.0,
					(random.Float64()*2.0-1.0)*511.0,
					(random.Float64()*2.0-1.0)*511.0,
				),
				dprec.NewVec3(
					(random.Float64()*2.0-1.0)*511.0,
					(random.Float64()*2.0-1.0)*511.0,
					(random.Float64()*2.0-1.0)*511.0,
				),
			)
			scene.CollectSegmentIntersections(segment, shape3d.Filter{}, &collection)
		}
	}
}
