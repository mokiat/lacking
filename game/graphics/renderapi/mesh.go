package graphics

import (
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/renderapi/internal"
	"github.com/mokiat/lacking/render"
)

var _ graphics.MeshTemplate = (*MeshTemplate)(nil)

type SubMeshTemplate struct {
	material         *Material
	topology         render.Topology
	indexCount       int
	indexOffsetBytes int
}

type MeshTemplate struct {
	vertexBuffer render.Buffer
	indexBuffer  render.Buffer
	vertexArray  render.VertexArray
	subMeshes    []SubMeshTemplate
}

func (t *MeshTemplate) Delete() {
	t.vertexArray.Release()
	t.indexBuffer.Release()
	t.vertexBuffer.Release()
	t.subMeshes = nil
}

var _ graphics.Mesh = (*Mesh)(nil)

type Mesh struct {
	internal.Node

	scene *Scene
	prev  *Mesh
	next  *Mesh

	template *MeshTemplate
}

func (m *Mesh) Delete() {
	m.scene.detachMesh(m)
	m.scene.cacheMesh(m)
	m.scene = nil
}
