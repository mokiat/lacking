package game

import (
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) convertMeshDefinition(geometris IdentifiableList[*graphics.MeshGeometry], materials IdentifiableList[*graphics.Material], assetMeshDefinition dto.MeshDefinition) async.Promise[*graphics.MeshDefinition] {
	geometry := geometris.GetByID(assetMeshDefinition.GeometryID)

	bindingMaterials := make([]*graphics.Material, geometry.FragmentCount())
	for _, assetBinding := range assetMeshDefinition.MaterialBindings {
		bindingMaterials[assetBinding.FragmentIndex] = materials.GetByID(assetBinding.MaterialID)
	}

	meshDefinitionInfo := graphics.MeshDefinitionInfo{
		Geometry:  geometry,
		Materials: bindingMaterials,
	}

	promise := async.NewPromise[*graphics.MeshDefinition]()
	s.gfxWorker.Schedule(func() {
		gfxEngine := s.engine.Graphics()
		meshDefinition := gfxEngine.CreateMeshDefinition(meshDefinitionInfo)
		promise.Deliver(meshDefinition)
	})
	return promise
}

func (s *ResourceSet) convertMeshInstance(meshDefinitionIndices map[uint32]int, assetMesh dto.Mesh) meshInstance {
	return meshInstance{
		NodeID:          assetMesh.NodeID,
		DefinitionIndex: meshDefinitionIndices[assetMesh.MeshDefinitionID],
		ArmatureID:      assetMesh.ArmatureID,
	}
}
