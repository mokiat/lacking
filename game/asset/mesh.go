package asset

import newasset "github.com/mokiat/lacking/game/newasset"

type MeshDefinition struct {
	Name                 string
	VertexBuffers        []newasset.VertexBuffer
	VertexLayout         newasset.VertexLayout
	IndexBuffer          newasset.IndexBuffer
	Fragments            []MeshFragment
	BoundingSphereRadius float64
}

type MeshFragment struct {
	Topology      newasset.Topology
	IndexOffset   uint32
	IndexCount    uint32
	MaterialIndex int32
}
