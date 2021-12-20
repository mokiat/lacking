package graphics

import (
	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/framework/opengl/game/graphics/internal"
	"github.com/mokiat/lacking/game/graphics"
)

var _ graphics.MeshTemplate = (*MeshTemplate)(nil)

type SubMeshTemplate struct {
	material         *Material
	primitive        uint32
	indexCount       int32
	indexOffsetBytes int
	indexType        uint32
}

type MeshTemplate struct {
	vertexBuffer *opengl.Buffer
	indexBuffer  *opengl.Buffer
	vertexArray  *opengl.VertexArray
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
