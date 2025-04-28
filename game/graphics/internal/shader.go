package internal

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
