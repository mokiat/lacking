package graphics

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
	"github.com/mokiat/lacking/util/spatial"
)

// MeshDefinitionInfo contains everything needed to create a new MeshDefinition.
type MeshDefinitionInfo struct {
	VertexData           []byte
	VertexFormat         VertexFormat
	IndexData            []byte
	IndexFormat          IndexFormat
	Fragments            []MeshFragmentDefinitionInfo
	BoundingSphereRadius float64
}

// NeedsArmature returns whether the mesh described by this info object will
// require an Armature to be visualized.
func (i *MeshDefinitionInfo) NeedsArmature() bool {
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
	vertexBuffer         render.Buffer
	indexBuffer          render.Buffer
	vertexArray          render.VertexArray
	fragments            []meshFragmentDefinition
	boundingSphereRadius float64
	hasVertexColors      bool
	needsArmature        bool
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

type meshFragmentDefinition struct {
	id               int
	mesh             *MeshDefinition
	topology         render.Topology
	indexCount       int
	indexOffsetBytes int
	material         *Material
}

func (d *meshFragmentDefinition) rebuildPipelines() {
	d.deletePipelines()
	d.createPipelines()
}

func (d *meshFragmentDefinition) deletePipelines() {
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

func (d *meshFragmentDefinition) createPipelines() {
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
	Definition *MeshDefinition
	Armature   *Armature
}

func newMesh(scene *Scene, info MeshInfo) *Mesh {
	definition := info.Definition

	mesh := scene.meshPool.Fetch()
	mesh.Node = newNode()
	mesh.scene = scene
	mesh.item = scene.meshOctree.CreateItem(mesh)
	mesh.item.SetRadius(definition.boundingSphereRadius)
	mesh.definition = definition
	mesh.armature = info.Armature
	return mesh
}

// Mesh represents an instance of a 3D mesh.
type Mesh struct {
	Node

	scene *Scene
	item  *spatial.OctreeItem[*Mesh]

	definition *MeshDefinition
	armature   *Armature
}

func (m *Mesh) SetMatrix(matrix dprec.Mat4) {
	m.Node.SetMatrix(matrix)
	m.item.SetPosition(matrix.Translation())
}

// Delete removes this mesh from the scene.
func (m *Mesh) Delete() {
	if m.scene == nil {
		panic(fmt.Errorf("mesh already deleted"))
	}
	m.item.Delete()
	m.item = nil
	m.scene.meshPool.Restore(m)
	m.scene = nil
}

type StaticMeshInfo struct {
	Definition *MeshDefinition
	Matrix     dprec.Mat4
}

func createStaticMesh(scene *Scene, info StaticMeshInfo) {
	position := info.Matrix.Translation()
	scale := info.Matrix.Scale()
	maxScale := dprec.Max(scale.X, dprec.Max(scale.Y, scale.Z))
	radius := info.Definition.boundingSphereRadius * maxScale

	meshIndex := uint32(len(scene.staticMeshes))
	scene.staticMeshes = append(scene.staticMeshes, StaticMesh{})
	scene.staticMeshOctree.Insert(position, radius, meshIndex)

	staticMesh := &scene.staticMeshes[meshIndex]
	staticMesh.definition = info.Definition
	staticMesh.matrixData = make([]byte, 16*4)

	matrix := dtos.Mat4(info.Matrix)
	plotter := blob.NewPlotter(staticMesh.matrixData)
	plotter.PlotFloat32(matrix.M11)
	plotter.PlotFloat32(matrix.M21)
	plotter.PlotFloat32(matrix.M31)
	plotter.PlotFloat32(matrix.M41)
	plotter.PlotFloat32(matrix.M12)
	plotter.PlotFloat32(matrix.M22)
	plotter.PlotFloat32(matrix.M32)
	plotter.PlotFloat32(matrix.M42)
	plotter.PlotFloat32(matrix.M13)
	plotter.PlotFloat32(matrix.M23)
	plotter.PlotFloat32(matrix.M33)
	plotter.PlotFloat32(matrix.M43)
	plotter.PlotFloat32(matrix.M14)
	plotter.PlotFloat32(matrix.M24)
	plotter.PlotFloat32(matrix.M34)
	plotter.PlotFloat32(matrix.M44)
}

type StaticMesh struct {
	matrixData []byte
	definition *MeshDefinition
}
