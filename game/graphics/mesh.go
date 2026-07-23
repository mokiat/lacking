package graphics

import (
	"fmt"
	"time"

	"github.com/mokiat/gblob"
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
	mesh.maxCascade = definition.geometry.maxCascade
	mesh.armature = info.Armature
	mesh.active = true
	mesh.SetCustom0Value(0.0)
	mesh.SetCustom1Value(0.0)
	mesh.SetCustom2Value(0.0)
	mesh.setSpawnTime(scene.gameTime)
	return mesh
}

// Mesh represents an instance of a 3D mesh.
type Mesh struct {
	Node

	scene        *Scene
	itemID       spatial.DynamicSetItemID
	definition   *MeshDefinition
	armature     *Armature
	maxCascade   uint8
	instanceData [4 * 4]byte // 1x vec4
	active       bool
}

func (m *Mesh) Active() bool {
	return m.active
}

func (m *Mesh) SetActive(active bool) {
	if active != m.active {
		m.setSpawnTime(m.scene.gameTime)
		m.active = active
	}
}

func (m *Mesh) SetMatrix(matrix dprec.Mat4) {
	m.Node.SetMatrix(matrix)
	position := matrix.Translation()
	radius := m.definition.geometry.boundingSphereRadius
	m.scene.dynamicMeshSet.Update(m.itemID, position, radius)
}

func (m *Mesh) SetCustom0Value(value float32) {
	block := gblob.LittleEndianBlock(m.instanceData[:])
	block.SetFloat32(1*4, value)
}

func (m *Mesh) SetCustom1Value(value float32) {
	block := gblob.LittleEndianBlock(m.instanceData[:])
	block.SetFloat32(2*4, value)
}

func (m *Mesh) SetCustom2Value(value float32) {
	block := gblob.LittleEndianBlock(m.instanceData[:])
	block.SetFloat32(3*4, value)
}

func (m *Mesh) setSpawnTime(spawnTime time.Duration) {
	block := gblob.LittleEndianBlock(m.instanceData[:])
	block.SetFloat32(0*4, float32(spawnTime.Seconds()))
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
	Armature   *Armature
	Matrix     dprec.Mat4
}

func createStaticMesh(scene *Scene, info StaticMeshInfo) {
	position := info.Matrix.Translation()
	scale := info.Matrix.Scale()
	maxScale := max(scale.X, scale.Y, scale.Z)
	radius := info.Definition.geometry.boundingSphereRadius * maxScale

	meshIndex := uint32(len(scene.staticMeshes))
	scene.staticMeshes = append(scene.staticMeshes, StaticMesh{})
	scene.staticMeshOctree.Insert(position, radius, meshIndex)

	staticMesh := &scene.staticMeshes[meshIndex]
	staticMesh.scene = scene
	staticMesh.position = position
	staticMesh.minDistance = info.Definition.geometry.minDistance
	staticMesh.maxDistance = info.Definition.geometry.maxDistance
	staticMesh.maxCascade = info.Definition.geometry.maxCascade
	staticMesh.definition = info.Definition
	staticMesh.matrixData = make([]byte, 16*4)
	staticMesh.armature = info.Armature
	staticMesh.active = true
	staticMesh.SetCustom0Value(0.0)
	staticMesh.SetCustom1Value(0.0)
	staticMesh.SetCustom2Value(0.0)
	staticMesh.setSpawnTime(scene.gameTime)

	matrix := dtos.Mat4(info.Matrix)
	plotter := blob.NewPlotter(staticMesh.matrixData)
	plotter.PlotSPMat4(matrix)
}

type StaticMesh struct {
	scene        *Scene
	position     dprec.Vec3
	minDistance  float64
	maxDistance  float64
	maxCascade   uint8
	matrixData   []byte
	definition   *MeshDefinition
	armature     *Armature
	instanceData [4 * 4]byte // 1x vec4
	active       bool
}

func (m *StaticMesh) Active() bool {
	return m.active
}

func (m *StaticMesh) SetActive(active bool) {
	if active != m.active {
		m.setSpawnTime(m.scene.gameTime)
		m.active = active
	}
}

func (m *StaticMesh) SetCustom0Value(value float32) {
	block := gblob.LittleEndianBlock(m.instanceData[:])
	block.SetFloat32(1*4, value)
}

func (m *StaticMesh) SetCustom1Value(value float32) {
	block := gblob.LittleEndianBlock(m.instanceData[:])
	block.SetFloat32(2*4, value)
}

func (m *StaticMesh) SetCustom2Value(value float32) {
	block := gblob.LittleEndianBlock(m.instanceData[:])
	block.SetFloat32(3*4, value)
}

func (m *StaticMesh) setSpawnTime(spawnTime time.Duration) {
	block := gblob.LittleEndianBlock(m.instanceData[:])
	block.SetFloat32(0*4, float32(spawnTime.Seconds()))
}
