package graphics

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/lacking/util/blob"
	"github.com/mokiat/lacking/util/spatial"
)

type MeshInfo struct {
	Definition *MeshDefinition
	Armature   *Armature
}

func newMesh(scene *Scene, info MeshInfo) *Mesh {
	definition := info.Definition
	mesh := scene.dynamicMeshPool.Fetch()
	mesh.Node = newNode()
	mesh.scene = scene
	mesh.itemID = scene.dynamicMeshSet.Insert(dprec.ZeroVec3(), definition.geometry.boundingSphereRadius, mesh)
	mesh.definition = definition
	mesh.armature = info.Armature
	return mesh
}

// Mesh represents an instance of a 3D mesh.
type Mesh struct {
	Node

	scene      *Scene
	itemID     spatial.DynamicSetItemID
	definition *MeshDefinition
	armature   *Armature
}

func (m *Mesh) SetMatrix(matrix dprec.Mat4) {
	m.Node.SetMatrix(matrix)
	position := matrix.Translation()
	radius := m.definition.geometry.boundingSphereRadius
	m.scene.dynamicMeshSet.Update(m.itemID, position, radius)
}

// Delete removes this mesh from the scene.
func (m *Mesh) Delete() {
	if m.scene == nil {
		panic(fmt.Errorf("mesh already deleted"))
	}
	m.scene.dynamicMeshSet.Remove(m.itemID)
	m.scene.dynamicMeshPool.Restore(m)
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
	radius := info.Definition.geometry.boundingSphereRadius * maxScale

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
