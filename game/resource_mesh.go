package game

import (
	"github.com/mokiat/lacking/game/asset/dto"
)

func (s *ResourceSet) convertMeshInstance(assetMesh dto.Mesh) meshInstance {
	return meshInstance{
		NodeID:       assetMesh.NodeID,
		DefinitionID: assetMesh.MeshDefinitionID,
		ArmatureID:   assetMesh.ArmatureID,
	}
}
