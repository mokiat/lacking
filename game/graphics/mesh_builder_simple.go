package graphics

// import (
// 	"github.com/mokiat/gomath/sprec"
// 	"github.com/mokiat/lacking/render"
// )

// // NewSimpleMeshBuilder creates a new simple mesh builder.
// func NewSimpleMeshBuilder() *SimpleMeshBuilder {
// 	return &SimpleMeshBuilder{
// 		meshBuilder: NewMeshBuilder(
// 			MeshBuilderWithCoords(),
// 		),
// 	}
// }

// // SimpleMeshBuilder is responsible for creating meshes made of a single
// // material.
// type SimpleMeshBuilder struct {
// 	meshBuilder *MeshBuilder

// 	lineFragments     []simpleMeshFragment
// 	triangleFragments []simpleMeshFragment
// }

// // Wireframe creates a new fragment composed of lines.
// func (mb *SimpleMeshBuilder) Wireframe(material *MaterialDefinition) WireframeMeshBuilder {
// 	itemIndex := len(mb.lineFragments)
// 	mb.lineFragments = append(mb.lineFragments, simpleMeshFragment{
// 		Material:    material,
// 		IndexOffset: mb.meshBuilder.IndexOffset(),
// 		IndexCount:  0,
// 	})
// 	return WireframeMeshBuilder{
// 		meshBuilder: mb.meshBuilder,
// 		fragment:    &mb.lineFragments[itemIndex],
// 	}
// }

// // Solid creates a new fragment composed of solid triangles.
// func (mb *SimpleMeshBuilder) Solid(material *MaterialDefinition) SolidMeshBuilder {
// 	itemIndex := len(mb.triangleFragments)
// 	mb.triangleFragments = append(mb.triangleFragments, simpleMeshFragment{
// 		Material:    material,
// 		IndexOffset: mb.meshBuilder.IndexOffset(),
// 		IndexCount:  0,
// 	})
// 	return SolidMeshBuilder{
// 		meshBuilder: mb.meshBuilder,
// 		indices:     &mb.triangleFragments[itemIndex],
// 	}
// }

// // BuildInfo returns the mesh definition info of the built mesh.
// func (mb *SimpleMeshBuilder) BuildInfo() MeshDefinitionInfo {
// 	for _, fragment := range mb.lineFragments {
// 		mb.meshBuilder.Fragment(render.TopologyLineList, fragment.Material, fragment.IndexOffset, fragment.IndexCount)
// 	}
// 	for _, fragment := range mb.triangleFragments {
// 		mb.meshBuilder.Fragment(render.TopologyTriangleList, fragment.Material, fragment.IndexOffset, fragment.IndexCount)
// 	}
// 	return mb.meshBuilder.BuildInfo()
// }

// // WireframeMeshBuilder is responsible for creating solid mesh lines.
// type WireframeMeshBuilder struct {
// 	meshBuilder *MeshBuilder
// 	fragment    *simpleMeshFragment
// }

// // Line creates a new line segment.
// func (mb WireframeMeshBuilder) Line(from, to sprec.Vec3) WireframeMeshBuilder {
// 	vertexStart := mb.meshBuilder.VertexOffset()
// 	mb.meshBuilder.Vertex().Coord(from.X, from.Y, from.Z)
// 	mb.meshBuilder.Vertex().Coord(to.X, to.Y, to.Z)

// 	indexStart := mb.meshBuilder.IndexOffset()
// 	mb.meshBuilder.IndexLine(vertexStart, vertexStart+1)

// 	mb.fragment.IndexCount += mb.meshBuilder.IndexOffset() - indexStart
// 	return mb
// }

// func (mb WireframeMeshBuilder) Circle(position sprec.Vec3, rotation sprec.Quat, radius float32, segments int) WireframeMeshBuilder {
// 	mb.meshBuilder.Transform(sprec.TRSMat4(
// 		position,
// 		rotation,
// 		sprec.NewVec3(radius, radius, 1.0),
// 	))
// 	defer mb.meshBuilder.Transform(sprec.IdentityMat4())

// 	vertexStart := mb.meshBuilder.VertexOffset()
// 	for i := 0; i < segments; i++ {
// 		angle := sprec.Degrees(360.0 * (float32(i) / float32(segments)))
// 		mb.meshBuilder.Vertex().Coord(sprec.Cos(angle), sprec.Sin(angle), 0.0)
// 	}

// 	indexStart := mb.meshBuilder.IndexOffset()
// 	for i := uint32(0); i < uint32(segments); i++ {
// 		a := vertexStart + i
// 		b := vertexStart + ((i + 1) % uint32(segments))
// 		mb.meshBuilder.IndexLine(a, b)
// 	}

// 	mb.fragment.IndexCount += mb.meshBuilder.IndexOffset() - indexStart
// 	return mb
// }

// func (mb WireframeMeshBuilder) Arc(position sprec.Vec3, rotation sprec.Quat, radius float32, from, to sprec.Angle, segments int) WireframeMeshBuilder {
// 	mb.meshBuilder.Transform(sprec.TRSMat4(
// 		position,
// 		rotation,
// 		sprec.NewVec3(radius, radius, 1.0),
// 	))
// 	defer mb.meshBuilder.Transform(sprec.IdentityMat4())

// 	delta := to - from

// 	vertexStart := mb.meshBuilder.VertexOffset()
// 	for i := 0; i < segments; i++ {
// 		angle := from + sprec.Degrees(delta.Degrees()*(float32(i)/float32(segments-1)))
// 		mb.meshBuilder.Vertex().Coord(sprec.Cos(angle), sprec.Sin(angle), 0.0)
// 	}

// 	indexStart := mb.meshBuilder.IndexOffset()
// 	for i := uint32(0); i < uint32(segments)-1; i++ {
// 		a := vertexStart + i
// 		b := vertexStart + i + 1
// 		mb.meshBuilder.IndexLine(a, b)
// 	}

// 	mb.fragment.IndexCount += mb.meshBuilder.IndexOffset() - indexStart
// 	return mb
// }

// // SolidMeshBuilder is responsible for creating solid mesh triangles.
// type SolidMeshBuilder struct {
// 	meshBuilder *MeshBuilder
// 	indices     *simpleMeshFragment
// }

// // Cuboid creates a new cuboid solid shape.
// func (mb SolidMeshBuilder) Cuboid(position sprec.Vec3, rotation sprec.Quat, dimensions sprec.Vec3) SolidMeshBuilder {
// 	mb.meshBuilder.Transform(sprec.TRSMat4(
// 		position,
// 		rotation,
// 		sprec.NewVec3(dimensions.X*0.5, dimensions.Y*0.5, dimensions.Z*0.5),
// 	))
// 	defer mb.meshBuilder.Transform(sprec.IdentityMat4())

// 	vertexStart := mb.meshBuilder.VertexOffset()
// 	mb.meshBuilder.Vertex().Coord(-1.0, -1.0, 1.0)  // 0. front-bottom-left
// 	mb.meshBuilder.Vertex().Coord(1.0, -1.0, 1.0)   // 1. front-bottom-right
// 	mb.meshBuilder.Vertex().Coord(1.0, 1.0, 1.0)    // 2. front-top-right
// 	mb.meshBuilder.Vertex().Coord(-1.0, 1.0, 1.0)   // 3. front-top-left
// 	mb.meshBuilder.Vertex().Coord(-1.0, -1.0, -1.0) // 4. back-bottom-left
// 	mb.meshBuilder.Vertex().Coord(1.0, -1.0, -1.0)  // 5. back-bottom-right
// 	mb.meshBuilder.Vertex().Coord(1.0, 1.0, -1.0)   // 6. back-top-right
// 	mb.meshBuilder.Vertex().Coord(-1.0, 1.0, -1.0)  // 7. back-top-left

// 	indexStart := mb.meshBuilder.IndexOffset()
// 	// front face
// 	mb.meshBuilder.IndexTriangle(vertexStart+3, vertexStart+0, vertexStart+1)
// 	mb.meshBuilder.IndexTriangle(vertexStart+3, vertexStart+1, vertexStart+2)
// 	// back face
// 	mb.meshBuilder.IndexTriangle(vertexStart+5, vertexStart+4, vertexStart+7)
// 	mb.meshBuilder.IndexTriangle(vertexStart+6, vertexStart+5, vertexStart+7)
// 	// left face
// 	mb.meshBuilder.IndexTriangle(vertexStart+7, vertexStart+4, vertexStart+0)
// 	mb.meshBuilder.IndexTriangle(vertexStart+7, vertexStart+0, vertexStart+3)
// 	// right face
// 	mb.meshBuilder.IndexTriangle(vertexStart+1, vertexStart+5, vertexStart+6)
// 	mb.meshBuilder.IndexTriangle(vertexStart+1, vertexStart+6, vertexStart+2)
// 	// top face
// 	mb.meshBuilder.IndexTriangle(vertexStart+2, vertexStart+6, vertexStart+7)
// 	mb.meshBuilder.IndexTriangle(vertexStart+2, vertexStart+7, vertexStart+3)
// 	// bottom face
// 	mb.meshBuilder.IndexTriangle(vertexStart+4, vertexStart+5, vertexStart+1)
// 	mb.meshBuilder.IndexTriangle(vertexStart+4, vertexStart+1, vertexStart+0)

// 	mb.indices.IndexCount += mb.meshBuilder.IndexOffset() - indexStart
// 	return mb
// }

// // Cylinder creates a new cylinder solid shape.
// func (mb SolidMeshBuilder) Cylinder(position sprec.Vec3, rotation sprec.Quat, radius float32, height float32, segments int) SolidMeshBuilder {
// 	if segments < 3 {
// 		panic("segments must be at least 3")
// 	}

// 	mb.meshBuilder.Transform(sprec.TRSMat4(
// 		position,
// 		rotation,
// 		sprec.NewVec3(radius, height*0.5, radius),
// 	))
// 	defer mb.meshBuilder.Transform(sprec.IdentityMat4())

// 	vertexStart := mb.meshBuilder.VertexOffset()
// 	mb.meshBuilder.Vertex().Coord(0.0, 1.0, 0.0)  // 0. top center
// 	mb.meshBuilder.Vertex().Coord(0.0, -1.0, 0.0) // 1. bottom center
// 	for i := 0; i < segments; i++ {               // top circle from 2 to 2+segments-1
// 		angle := sprec.Degrees(360.0 * (float32(i) / float32(segments)))
// 		mb.meshBuilder.Vertex().Coord(sprec.Cos(angle), 1.0, sprec.Sin(angle))
// 	}
// 	for i := 0; i < segments; i++ { // bottom circle from 2+segments to 2+segments*2-1
// 		angle := sprec.Degrees(360.0 * (float32(i) / float32(segments)))
// 		mb.meshBuilder.Vertex().Coord(sprec.Cos(angle), -1.0, sprec.Sin(angle))
// 	}

// 	indexStart := mb.meshBuilder.IndexOffset()
// 	// top circle
// 	for i := uint32(0); i < uint32(segments); i++ {
// 		a := vertexStart + 0 // center
// 		b := vertexStart + 2 + i
// 		c := vertexStart + 2 + ((i + 1) % uint32(segments))
// 		mb.meshBuilder.IndexTriangle(a, c, b)
// 	}
// 	// bottom circle
// 	for i := uint32(0); i < uint32(segments); i++ {
// 		a := vertexStart + 1 // center
// 		b := vertexStart + 2 + uint32(segments) + i
// 		c := vertexStart + 2 + uint32(segments) + ((i + 1) % uint32(segments))
// 		mb.meshBuilder.IndexTriangle(a, b, c)
// 	}
// 	// side
// 	for i := uint32(0); i < uint32(segments); i++ {
// 		a := vertexStart + 2 + i
// 		b := vertexStart + 2 + ((i + 1) % uint32(segments))
// 		c := vertexStart + 2 + uint32(segments) + i
// 		d := vertexStart + 2 + uint32(segments) + ((i + 1) % uint32(segments))
// 		mb.meshBuilder.IndexQuad(b, d, c, a)
// 	}

// 	mb.indices.IndexCount += mb.meshBuilder.IndexOffset() - indexStart
// 	return mb
// }

// // Cone creates a new cone solid shape.
// func (mb SolidMeshBuilder) Cone(position sprec.Vec3, rotation sprec.Quat, radius float32, height float32, segments int) SolidMeshBuilder {
// 	if segments < 3 {
// 		panic("segments must be at least 3")
// 	}

// 	mb.meshBuilder.Transform(sprec.TRSMat4(
// 		position,
// 		rotation,
// 		sprec.NewVec3(radius, height*0.5, radius),
// 	))
// 	defer mb.meshBuilder.Transform(sprec.IdentityMat4())

// 	vertexStart := mb.meshBuilder.VertexOffset()
// 	mb.meshBuilder.Vertex().Coord(0.0, 1.0, 0.0)  // 0. top center
// 	mb.meshBuilder.Vertex().Coord(0.0, -1.0, 0.0) // 1. bottom center
// 	for i := 0; i < segments; i++ {               // bottom circle from 2 to 2+segments-1
// 		angle := sprec.Degrees(360.0 * (float32(i) / float32(segments)))
// 		mb.meshBuilder.Vertex().Coord(sprec.Cos(angle), -1.0, sprec.Sin(angle))
// 	}

// 	indexStart := mb.meshBuilder.IndexOffset()
// 	// bottom circle
// 	for i := uint32(0); i < uint32(segments); i++ {
// 		a := vertexStart + 1 // center
// 		b := vertexStart + 2 + i
// 		c := vertexStart + 2 + ((i + 1) % uint32(segments))
// 		mb.meshBuilder.IndexTriangle(a, b, c)
// 	}
// 	// side
// 	for i := uint32(0); i < uint32(segments); i++ {
// 		a := vertexStart + 0
// 		b := vertexStart + 2 + i
// 		c := vertexStart + 2 + ((i + 1) % uint32(segments))
// 		mb.meshBuilder.IndexTriangle(a, c, b)
// 	}

// 	mb.indices.IndexCount += mb.meshBuilder.IndexOffset() - indexStart
// 	return mb
// }

// // Sphere creates a new sphere solid shape.
// func (mb SolidMeshBuilder) Sphere(position sprec.Vec3, radius float32, segments int) SolidMeshBuilder {
// 	if segments < 3 {
// 		panic("segments must be at least 3")
// 	}

// 	mb.meshBuilder.Transform(sprec.TRSMat4(
// 		position,
// 		sprec.IdentityQuat(),
// 		sprec.NewVec3(radius, radius, radius),
// 	))
// 	defer mb.meshBuilder.Transform(sprec.IdentityMat4())

// 	hAngleCount := (segments * 3) / 2
// 	vAngleCount := segments

// 	vertexStart := mb.meshBuilder.VertexOffset()
// 	mb.meshBuilder.Vertex().Coord(0.0, 1.0, 0.0)  // 0. top center
// 	mb.meshBuilder.Vertex().Coord(0.0, -1.0, 0.0) // 1. bottom center
// 	for x := 0; x < hAngleCount; x++ {
// 		hAngle := sprec.Degrees(360.0 * (float32(x) / float32(hAngleCount)))
// 		hCos := sprec.Cos(hAngle)
// 		hSin := sprec.Sin(hAngle)
// 		for y := 1; y < vAngleCount-1; y++ {
// 			vAngle := sprec.Degrees(90.0 - 180.0*(float32(y)/float32(vAngleCount-1)))
// 			vCos := sprec.Cos(vAngle)
// 			vSin := sprec.Sin(vAngle)
// 			mb.meshBuilder.Vertex().Coord(hCos*vCos, vSin, hSin*vCos)
// 		}
// 	}

// 	indexStart := mb.meshBuilder.IndexOffset()
// 	for x := 0; x < hAngleCount; x++ {
// 		left := x % hAngleCount
// 		right := (x + 1) % hAngleCount
// 		leftOffset := uint32(left * (vAngleCount - 2))
// 		rightOffset := uint32(right * (vAngleCount - 2))

// 		upperLeft := uint32(2 + leftOffset)
// 		upperRight := uint32(2 + rightOffset)

// 		mb.meshBuilder.IndexTriangle(
// 			vertexStart+0,
// 			vertexStart+upperLeft,
// 			vertexStart+upperRight,
// 		)

// 		for y := 1; y < vAngleCount-2; y++ {
// 			lowerLeft := upperLeft + 1
// 			lowerRight := upperRight + 1

// 			mb.meshBuilder.IndexTriangle(
// 				vertexStart+upperLeft,
// 				vertexStart+lowerLeft,
// 				vertexStart+lowerRight,
// 			)
// 			mb.meshBuilder.IndexTriangle(
// 				vertexStart+upperLeft,
// 				vertexStart+lowerRight,
// 				vertexStart+upperRight,
// 			)

// 			upperLeft++
// 			upperRight++
// 		}

// 		mb.meshBuilder.IndexTriangle(
// 			vertexStart+upperLeft,
// 			vertexStart+1,
// 			vertexStart+upperRight,
// 		)
// 	}

// 	mb.indices.IndexCount += mb.meshBuilder.IndexOffset() - indexStart
// 	return mb
// }

// type simpleMeshFragment struct {
// 	Material    *MaterialDefinition
// 	IndexOffset uint32
// 	IndexCount  uint32
// }
