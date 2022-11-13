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

func BenchmarkSphereToBoxPositiveCorner(b *testing.B) {
	resultSet := shape.NewIntersectionResultSet(b.N)

	sphere := shape.NewPlacement(
		shape.NewTransform(
			dprec.NewVec3(2.2, 4.4, 1.1),
			dprec.IdentityQuat(),
		),
		shape.NewStaticSphere(1.0),
	)

	box := shape.NewPlacement(
		shape.NewTransform(
			dprec.NewVec3(0.0, 3.0, 0.0),
			dprec.IdentityQuat(),
		),
		shape.NewStaticBox(4.0, 2.4, 1.8),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shape.CheckIntersectionSphereWithBox(sphere, box, resultSet)
	}
}

func BenchmarkSphereToBoxPositiveSide(b *testing.B) {
	resultSet := shape.NewIntersectionResultSet(b.N)

	sphere := shape.NewPlacement(
		shape.NewTransform(
			dprec.NewVec3(0.0, 4.7, 0.0),
			dprec.IdentityQuat(),
		),
		shape.NewStaticSphere(1.0),
	)

	box := shape.NewPlacement(
		shape.NewTransform(
			dprec.NewVec3(0.0, 3.0, 0.0),
			dprec.IdentityQuat(),
		),
		shape.NewStaticBox(4.0, 2.4, 1.8),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shape.CheckIntersectionSphereWithBox(sphere, box, resultSet)
	}
}

func BenchmarkSphereToBoxNegative(b *testing.B) {
	resultSet := shape.NewIntersectionResultSet(0)

	sphere := shape.NewPlacement(
		shape.NewTransform(
			dprec.NewVec3(2.5, 5.0, 1.4),
			dprec.IdentityQuat(),
		),
		shape.NewStaticSphere(1.0),
	)

	box := shape.NewPlacement(
		shape.NewTransform(
			dprec.NewVec3(0.0, 3.0, 0.0),
			dprec.IdentityQuat(),
		),
		shape.NewStaticBox(4.0, 2.4, 1.8),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shape.CheckIntersectionSphereWithBox(sphere, box, resultSet)
	}
}

func BenchmarkSphereToMeshNegative(b *testing.B) {
	resultSet := shape.NewIntersectionResultSet(0)

	sphere := shape.NewPlacement(
		shape.NewTransform(
			dprec.NewVec3(0.1, 5.0, 0.1),
			dprec.IdentityQuat(),
		),
		shape.NewStaticSphere(1.0),
	)

	mesh := shape.NewPlacement(
		shape.NewTransform(
			dprec.NewVec3(0.0, 3.0, 0.0),
			dprec.IdentityQuat(),
		),
		shape.NewStaticMesh([]shape.StaticTriangle{
			shape.NewStaticTriangle(
				shape.Point(dprec.NewVec3(0.0, 0.0, -2.0)),
				shape.Point(dprec.NewVec3(-2.0, 0.0, 1.4)),
				shape.Point(dprec.NewVec3(2.0, 0.0, 1.4)),
			),
		}),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shape.CheckIntersectionSphereWithMesh(sphere, mesh, resultSet)
	}
}

func BenchmarkSphereToMeshNegativeBroad(b *testing.B) {
	resultSet := shape.NewIntersectionResultSet(0)

	sphere := shape.NewPlacement(
		shape.NewTransform(
			dprec.NewVec3(4.0, 8.0, 0.0),
			dprec.IdentityQuat(),
		),
		shape.NewStaticSphere(1.0),
	)

	mesh := shape.NewPlacement(
		shape.NewTransform(
			dprec.NewVec3(0.0, 3.0, 0.0),
			dprec.IdentityQuat(),
		),
		shape.NewStaticMesh([]shape.StaticTriangle{
			shape.NewStaticTriangle(
				shape.Point(dprec.NewVec3(0.0, 5.0, -2.0)),
				shape.Point(dprec.NewVec3(-2.0, 5.0, 1.4)),
				shape.Point(dprec.NewVec3(2.0, 5.0, 1.4)),
			),
		}),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		shape.CheckIntersectionSphereWithMesh(sphere, mesh, resultSet)
	}
}
