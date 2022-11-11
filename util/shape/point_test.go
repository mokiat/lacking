package shape_test

import (
	"testing"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape"
)

func BenchmarkPointTransformedStandard(b *testing.B) {
	transform := shape.NewTransform(
		dprec.NewVec3(1.0, 2.0, 3.0),
		dprec.RotationQuat(dprec.Degrees(35), dprec.BasisXVec3()),
	)
	point := shape.Point(
		dprec.NewVec3(3.0, 2.0, 1.0),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		point.Transformed(transform)
	}
}

func BenchmarkPointTransformedIdentity(b *testing.B) {
	transform := shape.IdentityTransform()
	point := shape.Point(
		dprec.NewVec3(3.0, 2.0, 1.0),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		point.Transformed(transform)
	}
}
