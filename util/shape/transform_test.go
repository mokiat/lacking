package shape_test

import (
	"testing"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape"
)

func BenchmarkTransformTransformedStandard(b *testing.B) {
	first := shape.NewTransform(
		dprec.NewVec3(1.0, 2.0, 3.0),
		dprec.RotationQuat(dprec.Degrees(35), dprec.BasisXVec3()),
	)
	second := shape.NewTransform(
		dprec.NewVec3(3.0, 2.0, 1.0),
		dprec.RotationQuat(dprec.Degrees(35), dprec.BasisYVec3()),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		second.Transformed(first)
	}
}

func BenchmarkTransformTransformedParentIdentity(b *testing.B) {
	first := shape.IdentityTransform()
	second := shape.NewTransform(
		dprec.NewVec3(3.0, 2.0, 1.0),
		dprec.RotationQuat(dprec.Degrees(35), dprec.BasisYVec3()),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		second.Transformed(first)
	}
}

func BenchmarkTransformTransformedChildIdentity(b *testing.B) {
	first := shape.NewTransform(
		dprec.NewVec3(1.0, 2.0, 3.0),
		dprec.RotationQuat(dprec.Degrees(35), dprec.BasisXVec3()),
	)
	second := shape.IdentityTransform()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		second.Transformed(first)
	}
}
