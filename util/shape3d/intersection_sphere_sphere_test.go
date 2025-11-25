package shape3d_test

import (
	"math/rand/v2"
	"testing"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape3d"
)

func BenchmarkIsSphereSphere(b *testing.B) {
	random := rand.New(rand.NewPCG(13, 0xFFFFFFFF-17))

	spheres := make([]shape3d.Sphere, 1024)
	for i := range spheres {
		spheres[i] = shape3d.Sphere{
			Position: dprec.Vec3{
				X: 8.0 * (random.Float64()*2.0 - 1.0),
				Y: 0.0,
				Z: 8.0 * (random.Float64()*2.0 - 1.0),
			},
		}
	}

	testSphere := shape3d.Sphere{
		Position: dprec.NewVec3(-0.1, 0.1, -0.1),
		Radius:   1.0,
	}

	for b.Loop() {
		for _, sphere := range spheres {
			shape3d.IsSphereSphereIntersection(sphere, testSphere)
		}
	}
}

func BenchmarkSphereSphere(b *testing.B) {
	random := rand.New(rand.NewPCG(13, 0xFFFFFFFF-17))

	spheres := make([]shape3d.Sphere, 1024)
	for i := range spheres {
		spheres[i] = shape3d.Sphere{
			Position: dprec.Vec3{
				X: 8.0 * (random.Float64()*2.0 - 1.0),
				Y: 0.0,
				Z: 8.0 * (random.Float64()*2.0 - 1.0),
			},
		}
	}

	testSphere := shape3d.Sphere{
		Position: dprec.NewVec3(-0.1, 0.1, -0.1),
		Radius:   1.0,
	}

	for b.Loop() {
		for _, sphere := range spheres {
			shape3d.CheckSphereSphereIntersection(sphere, testSphere, func(shape3d.Intersection) {})
		}
	}
}
