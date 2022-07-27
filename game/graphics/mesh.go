package graphics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/spatial"
)

// MeshTemplate represents the definition of a mesh.
// Multiple mesh instances can be created off of one template
// reusing resources.
type MeshTemplate struct {
	vertexBuffer render.Buffer
	indexBuffer  render.Buffer
	vertexArray  render.VertexArray
	subMeshes    []subMeshTemplate
}

// Delete releases any resources allocated by this
// template.
func (t *MeshTemplate) Delete() {
	t.vertexArray.Release()
	t.indexBuffer.Release()
	t.vertexBuffer.Release()
	t.subMeshes = nil
}

type subMeshTemplate struct {
	id               int
	material         *Material
	topology         render.Topology
	indexCount       int
	indexOffsetBytes int
	pipeline         render.Pipeline
}

// MeshTemplateDefinition contains everything needed to create
// a new MeshTemplate.
type MeshTemplateDefinition struct {
	VertexData   []byte
	VertexFormat VertexFormat
	IndexData    []byte
	IndexFormat  IndexFormat
	SubMeshes    []SubMeshTemplateDefinition
}

// SubMeshTemplateDefinition represents a portion of a mesh that
// is drawn with a specific material.
type SubMeshTemplateDefinition struct {
	Primitive   Primitive
	IndexOffset int
	IndexCount  int
	Material    *Material
}

const (
	PrimitivePoints Primitive = 1 + iota
	PrimitiveLines
	PrimitiveLineStrip
	PrimitiveLineLoop
	PrimitiveTriangles
	PrimitiveTriangleStrip
	PrimitiveTriangleFan
)

type Primitive int

type VertexFormat struct {
	HasCoord            bool
	CoordOffsetBytes    int
	CoordStrideBytes    int
	HasNormal           bool
	NormalOffsetBytes   int
	NormalStrideBytes   int
	HasTangent          bool
	TangentOffsetBytes  int
	TangentStrideBytes  int
	HasTexCoord         bool
	TexCoordOffsetBytes int
	TexCoordStrideBytes int
	HasColor            bool
	ColorOffsetBytes    int
	ColorStrideBytes    int
	HasWeights          bool
	WeightsOffsetBytes  int
	WeightsStrideBytes  int
	HasJoints           bool
	JointsOffsetBytes   int
	JointsStrideBytes   int
}

const (
	IndexFormatU16 IndexFormat = 1 + iota
	IndexFormatU32
)

type IndexFormat int

type MeshDefinition struct {
	Template *MeshTemplate
	Armature *Armature
}

// Mesh represents an instance of a 3D mesh.
type Mesh struct {
	Node

	item  *spatial.OctreeItem[*Mesh]
	scene *Scene
	prev  *Mesh
	next  *Mesh

	template *MeshTemplate
	armature *Armature
}

func (m *Mesh) SetMatrix(matrix dprec.Mat4) {
	m.Node.SetMatrix(matrix)
	m.item.SetPosition(matrix.Translation())
}

func (m *Mesh) SetArmature(armature *Armature) {
	m.armature = armature
}

func (m *Mesh) SetBoundingSphereRadius(radius float64) {
	m.item.SetRadius(radius)
}

// Delete removes this mesh from the scene.
func (m *Mesh) Delete() {
	m.item.Delete()
	m.scene.detachMesh(m)
	m.scene.cacheMesh(m)
	m.scene = nil
}
