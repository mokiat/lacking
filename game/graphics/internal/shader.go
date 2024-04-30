package internal

import "github.com/mokiat/lacking/render"

type ShaderMeshInfo struct {
	MeshHasCoords       bool
	MeshHasNormals      bool
	MeshHasTangents     bool
	MeshHasTextureUVs   bool
	MeshHasVertexColors bool
	MeshHasArmature     bool
}

type ShaderProgramCodeInfo struct {
	ShaderMeshInfo
}

type Shader interface {
	CreateProgramCode(info ShaderProgramCodeInfo) render.ProgramCode
}
