package shape_test

import (
	"testing"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape"
)

func BenchmarkSphereToSpherePositive(b *testing.B) {
	resultSet := shape.NewIntersectionResultSet(b.N)

	firstSphere := shape.NewPlacement(
		shape.NewTransform(
			dprec.NewVec3(0.0, 1.0, 0.0),
			dprec.IdentityQuat(),
		),
		shape.NewStaticSphere(1.0),
	)

	secondSphere := shape.NewPlacement(
		shape.NewTransform(
			dprec.NewVec3(0.5, 1.5, 0.5),
			dprec.IdentityQuat(),
		),
		shape.NewStaticSphere(1.0),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shape.CheckIntersectionSphereWithSphere(firstSphere, secondSphere, resultSet)
	}
}

func BenchmarkSphereToSphereNegative(b *testing.B) {
	resultSet := shape.NewIntersectionResultSet(0)

	firstSphere := shape.NewPlacement(
		shape.NewTransform(
			dprec.NewVec3(0.0, 1.0, 0.0),
			dprec.IdentityQuat(),
		),
		shape.NewStaticSphere(1.0),
	)

	secondSphere := shape.NewPlacement(
		shape.NewTransform(
			dprec.NewVec3(0.5, 5.5, 0.5),
			dprec.IdentityQuat(),
		),
		shape.NewStaticSphere(1.0),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shape.CheckIntersectionSphereWithSphere(firstSphere, secondSphere, resultSet)
	}
}
