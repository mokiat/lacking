package shape2d_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/shape2d"
)

var _ = Describe("Scene", func() {
	var (
		scene *shape2d.Scene[string, string]

		firstObjID   shape2d.ObjectID
		firstShapeID shape2d.ShapeID

		secondObjID   shape2d.ObjectID
		secondShapeID shape2d.ShapeID
	)

	BeforeEach(func() {
		scene = shape2d.NewScene[string, string](shape2d.SceneSettings{
			Size:     opt.V(float32(128.0)),
			MaxDepth: opt.V[uint32](3),
		})

		firstObjID = scene.CreateObject(shape2d.ObjectInfo[string]{
			Position: opt.V(dprec.NewVec2(16.0, 16.0)),
			Rotation: opt.V(dprec.Degrees(0.0)),
			Static:   false,
			UserData: "First",
		})
		firstShapeID = scene.AttachCircle(firstObjID, shape2d.CircleInfo[string]{
			ShapeInfo: shape2d.ShapeInfo[string]{
				UserData: "Circle",
			},
			Circle: shape2d.NewCircle(
				dprec.NewVec2(4.0, 4.0),
				2.0,
			),
		})

		secondObjID = scene.CreateObject(shape2d.ObjectInfo[string]{
			Position: opt.V(dprec.NewVec2(-16.0, -48.0)),
			Rotation: opt.V(dprec.Degrees(0.0)),
			Static:   false,
			UserData: "Second",
		})
		secondShapeID = scene.AttachRectangle(secondObjID, shape2d.RectangleInfo[string]{
			ShapeInfo: shape2d.ShapeInfo[string]{
				UserData: "Rectangle",
			},
			Rectangle: shape2d.NewRectangle(
				dprec.NewVec2(2.0, 1.0),
				dprec.Degrees(45.0),
				dprec.NewVec2(4.0, 2.0),
			),
		})
	})

	It("detects segment-circle intersection", func() {
		var bucket shape2d.ObjectIntersectionBucket
		segment := shape2d.NewSegment(
			dprec.NewVec2(14.0, 20.0),
			dprec.NewVec2(26.0, 20.0),
		)
		scene.CollectSegmentIntersections(segment, shape2d.Filter{}, &bucket)
		intersections := bucket.Intersections()
		Expect(intersections).To(HaveLen(1))
		intersection := intersections[0]
		Expect(intersection.TargetObjectID).To(Equal(firstObjID))
		Expect(intersection.TargetShapeID).To(Equal(firstShapeID))
	})

	It("detects segment-rectangle intersection", func() {
		var bucket shape2d.ObjectIntersectionBucket
		segment := shape2d.NewSegment(
			dprec.NewVec2(-20.0, -47.0),
			dprec.NewVec2(-8.0, -47.0),
		)
		scene.CollectSegmentIntersections(segment, shape2d.Filter{}, &bucket)
		intersections := bucket.Intersections()
		Expect(intersections).To(HaveLen(1))
		intersection := intersections[0]
		Expect(intersection.TargetObjectID).To(Equal(secondObjID))
		Expect(intersection.TargetShapeID).To(Equal(secondShapeID))
	})
})
