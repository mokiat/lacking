package internal

import "github.com/mokiat/lacking/render"

type ShaderMeshInfo struct {
	VertexArray         render.VertexArray
	MeshHasCoords       bool
	MeshHasNormals      bool
	MeshHasTangents     bool
	MeshHasTextureUVs   bool
	MeshHasVertexColors bool
	MeshHasArmature     bool
}

type GeometryShaderProgramCodeInfo struct {
	ShaderMeshInfo
}

type GeometryShader interface {
	CreateProgramCode(info GeometryShaderProgramCodeInfo) render.ProgramCode
}

type ShadowShaderProgramCodeInfo struct {
	ShaderMeshInfo
}

type ShadowShader interface {
	CreateProgramCode(info ShadowShaderProgramCodeInfo) render.ProgramCode
}

type ForwardShaderProgramCodeInfo struct {
	ShaderMeshInfo
}

type ForwardShader interface {
	CreateProgramCode(info ForwardShaderProgramCodeInfo) render.ProgramCode
}
