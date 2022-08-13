package graphics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/spatial"
)

// MeshDefinitionInfo contains everything needed to create a new MeshDefinition.
type MeshDefinitionInfo struct {
	VertexData   []byte
	VertexFormat VertexFormat
	IndexData    []byte
	IndexFormat  IndexFormat
	Fragments    []MeshFragmentDefinitionInfo
}

// HasArmature returns whether the mesh described by this info object will
// require an Armature to be visualized.
func (i *MeshDefinitionInfo) HasArmature() bool {
	return i.VertexFormat.HasWeights && i.VertexFormat.HasJoints
}

// VertexFormat describes the data that is contained in a single vertex.
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

// IndexFormat specifies the data type that is used to represent individual
// indices.
type IndexFormat int

// MeshFragmentDefinitionInfo represents the definition of a portion of a mesh
// that is drawn with a specific material and primitives.
type MeshFragmentDefinitionInfo struct {
	Primitive   Primitive
	IndexOffset int
	IndexCount  int
	Material    *MaterialDefinition
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

// Primitive represents the basic shape unit that is described by indices
// (for example: triangles).
type Primitive int

// MeshDefinition represents the definition of a mesh.
// Multiple mesh instances can be created off of one template
// reusing resources.
type MeshDefinition struct {
	vertexBuffer render.Buffer
	indexBuffer  render.Buffer
	vertexArray  render.VertexArray
	fragments    []MeshFragmentDefinition
	hasArmature  bool
}

// Delete releases any resources owned by this MeshDefinition.
func (t *MeshDefinition) Delete() {
	t.vertexArray.Release()
	t.indexBuffer.Release()
	t.vertexBuffer.Release()
	for _, fragment := range t.fragments {
		fragment.deletePipelines()
	}
	t.fragments = nil
}

type MeshFragmentDefinition struct {
	id               int
	mesh             *MeshDefinition
	topology         render.Topology
	indexCount       int
	indexOffsetBytes int
	material         *Material
}

func (d *MeshFragmentDefinition) rebuildPipelines() {
	d.deletePipelines()
	d.createPipelines()
}

func (d *MeshFragmentDefinition) deletePipelines() {
	if d.material.shadowPipeline != nil {
		d.material.shadowPipeline.Release()
		d.material.shadowPipeline = nil
	}
	if d.material.geometryPipeline != nil {
		d.material.geometryPipeline.Release()
		d.material.geometryPipeline = nil
	}
	if d.material.emissivePipeline != nil {
		d.material.emissivePipeline.Release()
		d.material.emissivePipeline = nil
	}
	if d.material.forwardPipeline != nil {
		d.material.forwardPipeline.Release()
		d.material.forwardPipeline = nil
	}
}

func (d *MeshFragmentDefinition) createPipelines() {
	// TODO: Consider moving to Material object instead
	material := d.material
	materialDef := material.definition
	material.definitionRevision = materialDef.revision
	material.shadowPipeline = materialDef.shading.ShadowPipeline(d.mesh, d)
	material.geometryPipeline = materialDef.shading.GeometryPipeline(d.mesh, d)
	material.emissivePipeline = materialDef.shading.EmissivePipeline(d.mesh, d)
	material.forwardPipeline = materialDef.shading.ForwardPipeline(d.mesh, d)
}

type MeshInfo struct {
	Template *MeshDefinition
	Armature *Armature
}

// Mesh represents an instance of a 3D mesh.
type Mesh struct {
	Node

	item  *spatial.OctreeItem[*Mesh]
	scene *Scene
	prev  *Mesh
	next  *Mesh

	definition *MeshDefinition
	armature   *Armature
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
