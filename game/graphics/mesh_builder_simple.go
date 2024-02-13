package graphics

import "github.com/mokiat/gomath/sprec"

// NewSimpleMeshBuilder creates a new simple mesh builder.
func NewSimpleMeshBuilder(material *MaterialDefinition) *SimpleMeshBuilder {
	return &SimpleMeshBuilder{
		material: material,
		meshBuilder: NewMeshBuilder(
			MeshBuilderWithCoords(),
		),
	}
}

// SimpleMeshBuilder is responsible for creating meshes made of a single
// material.
type SimpleMeshBuilder struct {
	material    *MaterialDefinition
	meshBuilder *MeshBuilder

	lineFragments     []indexRange
	triangleFragments []indexRange
}

// Wireframe creates a new fragment composed of lines.
func (mb *SimpleMeshBuilder) Wireframe() WireframeMeshBuilder {
	itemIndex := len(mb.lineFragments)
	mb.lineFragments = append(mb.lineFragments, indexRange{
		Offset: mb.meshBuilder.IndexOffset(),
		Count:  0,
	})
	return WireframeMeshBuilder{
		meshBuilder: mb.meshBuilder,
		indices:     &mb.lineFragments[itemIndex],
	}
}

// Solid creates a new fragment composed of solid triangles.
func (mb *SimpleMeshBuilder) Solid() SolidMeshBuilder {
	itemIndex := len(mb.triangleFragments)
	mb.triangleFragments = append(mb.triangleFragments, indexRange{
		Offset: mb.meshBuilder.IndexOffset(),
		Count:  0,
	})
	return SolidMeshBuilder{
		meshBuilder: mb.meshBuilder,
		indices:     &mb.triangleFragments[itemIndex],
	}
}

// BuildInfo returns the mesh definition info of the built mesh.
func (mb *SimpleMeshBuilder) BuildInfo() MeshDefinitionInfo {
	for _, indexRange := range mb.lineFragments {
		mb.meshBuilder.Fragment(PrimitiveLines, mb.material, indexRange.Offset, indexRange.Count)
	}
	for _, indexRange := range mb.triangleFragments {
		mb.meshBuilder.Fragment(PrimitiveTriangles, mb.material, indexRange.Offset, indexRange.Count)
	}
	return mb.meshBuilder.BuildInfo()
}

// WireframeMeshBuilder is responsible for creating solid mesh lines.
type WireframeMeshBuilder struct {
	meshBuilder *MeshBuilder
	indices     *indexRange
}

// Line creates a new line segment.
func (mb WireframeMeshBuilder) Line(from, to sprec.Vec3) WireframeMeshBuilder {
	vertexStart := mb.meshBuilder.VertexOffset()
	mb.meshBuilder.Vertex().Coord(from.X, from.Y, from.Z)
	mb.meshBuilder.Vertex().Coord(to.X, to.Y, to.Z)

	indexStart := mb.meshBuilder.IndexOffset()
	mb.meshBuilder.IndexLine(vertexStart, vertexStart+1)

	mb.indices.Count += mb.meshBuilder.IndexOffset() - indexStart
	return mb
}

// TODO: Add cuboid, sphere, cylinder, rectangle, etc. methods

// SolidMeshBuilder is responsible for creating solid mesh triangles.
type SolidMeshBuilder struct {
	meshBuilder *MeshBuilder
	indices     *indexRange
}

// Cuboid creates a new cuboid solid shape.
func (mb SolidMeshBuilder) Cuboid(position sprec.Vec3, rotation sprec.Quat, dimensions sprec.Vec3) SolidMeshBuilder {
	mb.meshBuilder.Transform(sprec.TRSMat4(
		position,
		rotation,
		sprec.NewVec3(dimensions.X*0.5, dimensions.Y*0.5, dimensions.Z*0.5),
	))
	defer mb.meshBuilder.Transform(sprec.IdentityMat4())

	vertexStart := mb.meshBuilder.VertexOffset()
	mb.meshBuilder.Vertex().Coord(-1.0, -1.0, 1.0)  // 0. front-bottom-left
	mb.meshBuilder.Vertex().Coord(1.0, -1.0, 1.0)   // 1. front-bottom-right
	mb.meshBuilder.Vertex().Coord(1.0, 1.0, 1.0)    // 2. front-top-right
	mb.meshBuilder.Vertex().Coord(-1.0, 1.0, 1.0)   // 3. front-top-left
	mb.meshBuilder.Vertex().Coord(-1.0, -1.0, -1.0) // 4. back-bottom-left
	mb.meshBuilder.Vertex().Coord(1.0, -1.0, -1.0)  // 5. back-bottom-right
	mb.meshBuilder.Vertex().Coord(1.0, 1.0, -1.0)   // 6. back-top-right
	mb.meshBuilder.Vertex().Coord(-1.0, 1.0, -1.0)  // 7. back-top-left

	indexStart := mb.meshBuilder.IndexOffset()
	// front face
	mb.meshBuilder.IndexTriangle(vertexStart+3, vertexStart+0, vertexStart+1)
	mb.meshBuilder.IndexTriangle(vertexStart+3, vertexStart+1, vertexStart+2)
	// back face
	mb.meshBuilder.IndexTriangle(vertexStart+5, vertexStart+4, vertexStart+7)
	mb.meshBuilder.IndexTriangle(vertexStart+6, vertexStart+5, vertexStart+7)
	// left face
	mb.meshBuilder.IndexTriangle(vertexStart+7, vertexStart+4, vertexStart+0)
	mb.meshBuilder.IndexTriangle(vertexStart+7, vertexStart+0, vertexStart+3)
	// right face
	mb.meshBuilder.IndexTriangle(vertexStart+1, vertexStart+5, vertexStart+6)
	mb.meshBuilder.IndexTriangle(vertexStart+1, vertexStart+6, vertexStart+2)
	// top face
	mb.meshBuilder.IndexTriangle(vertexStart+2, vertexStart+6, vertexStart+7)
	mb.meshBuilder.IndexTriangle(vertexStart+2, vertexStart+7, vertexStart+3)
	// bottom face
	mb.meshBuilder.IndexTriangle(vertexStart+4, vertexStart+5, vertexStart+1)
	mb.meshBuilder.IndexTriangle(vertexStart+4, vertexStart+1, vertexStart+0)

	mb.indices.Count += mb.meshBuilder.IndexOffset() - indexStart
	return mb
}

// Cylinder creates a new cylinder solid shape.
func (mb SolidMeshBuilder) Cylinder(position sprec.Vec3, rotation sprec.Quat, radius float32, height float32, segments int) SolidMeshBuilder {
	if segments < 3 {
		panic("segments must be at least 3")
	}

	mb.meshBuilder.Transform(sprec.TRSMat4(
		position,
		rotation,
		sprec.NewVec3(radius, height*0.5, radius),
	))
	defer mb.meshBuilder.Transform(sprec.IdentityMat4())

	vertexStart := mb.meshBuilder.VertexOffset()
	mb.meshBuilder.Vertex().Coord(0.0, 1.0, 0.0)  // 0. top center
	mb.meshBuilder.Vertex().Coord(0.0, -1.0, 0.0) // 1. bottom center
	for i := 0; i < segments; i++ {               // top circle from 2 to 2+segments-1
		angle := sprec.Degrees(360.0 * (float32(i) / float32(segments)))
		mb.meshBuilder.Vertex().Coord(sprec.Cos(angle), 1.0, sprec.Sin(angle))
	}
	for i := 0; i < segments; i++ { // bottom circle from 2+segments to 2+segments*2-1
		angle := sprec.Degrees(360.0 * (float32(i) / float32(segments)))
		mb.meshBuilder.Vertex().Coord(sprec.Cos(angle), -1.0, sprec.Sin(angle))
	}

	indexStart := mb.meshBuilder.IndexOffset()
	// top circle
	for i := uint32(0); i < uint32(segments); i++ {
		a := vertexStart + 0 // center
		b := vertexStart + 2 + i
		c := vertexStart + 2 + ((i + 1) % uint32(segments))
		mb.meshBuilder.IndexTriangle(a, c, b)
	}
	// bottom circle
	for i := uint32(0); i < uint32(segments); i++ {
		a := vertexStart + 1 // center
		b := vertexStart + 2 + uint32(segments) + i
		c := vertexStart + 2 + uint32(segments) + ((i + 1) % uint32(segments))
		mb.meshBuilder.IndexTriangle(a, b, c)
	}
	// side
	for i := uint32(0); i < uint32(segments); i++ {
		a := vertexStart + 2 + i
		b := vertexStart + 2 + ((i + 1) % uint32(segments))
		c := vertexStart + 2 + uint32(segments) + i
		d := vertexStart + 2 + uint32(segments) + ((i + 1) % uint32(segments))
		mb.meshBuilder.IndexQuad(b, d, c, a)
	}

	mb.indices.Count += mb.meshBuilder.IndexOffset() - indexStart
	return mb
}

// Cone creates a new cone solid shape.
func (mb SolidMeshBuilder) Cone(position sprec.Vec3, rotation sprec.Quat, radius float32, height float32, segments int) SolidMeshBuilder {
	if segments < 3 {
		panic("segments must be at least 3")
	}

	mb.meshBuilder.Transform(sprec.TRSMat4(
		position,
		rotation,
		sprec.NewVec3(radius, height*0.5, radius),
	))
	defer mb.meshBuilder.Transform(sprec.IdentityMat4())

	vertexStart := mb.meshBuilder.VertexOffset()
	mb.meshBuilder.Vertex().Coord(0.0, 1.0, 0.0)  // 0. top center
	mb.meshBuilder.Vertex().Coord(0.0, -1.0, 0.0) // 1. bottom center
	for i := 0; i < segments; i++ {               // bottom circle from 2 to 2+segments-1
		angle := sprec.Degrees(360.0 * (float32(i) / float32(segments)))
		mb.meshBuilder.Vertex().Coord(sprec.Cos(angle), -1.0, sprec.Sin(angle))
	}

	indexStart := mb.meshBuilder.IndexOffset()
	// bottom circle
	for i := uint32(0); i < uint32(segments); i++ {
		a := vertexStart + 1 // center
		b := vertexStart + 2 + i
		c := vertexStart + 2 + ((i + 1) % uint32(segments))
		mb.meshBuilder.IndexTriangle(a, b, c)
	}
	// side
	for i := uint32(0); i < uint32(segments); i++ {
		a := vertexStart + 0
		b := vertexStart + 2 + i
		c := vertexStart + 2 + ((i + 1) % uint32(segments))
		mb.meshBuilder.IndexTriangle(a, c, b)
	}

	mb.indices.Count += mb.meshBuilder.IndexOffset() - indexStart
	return mb
}

// TODO: Add sphere

type indexRange struct {
	Offset uint32
	Count  uint32
}
