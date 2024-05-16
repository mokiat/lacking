package game

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/util/async"
)

func (s *ResourceSet) convertMeshGeometry(assetGeometry asset.Geometry) async.Promise[*graphics.MeshGeometry] {
	meshFragmentsInfo := make([]graphics.MeshGeometryFragmentInfo, len(assetGeometry.Fragments))
	for j, assetFragment := range assetGeometry.Fragments {
		meshFragmentsInfo[j] = graphics.MeshGeometryFragmentInfo{
			Name:            assetFragment.Name,
			Topology:        s.resolveTopology(assetFragment.Topology),
			IndexByteOffset: assetFragment.IndexByteOffset,
			IndexCount:      assetFragment.IndexCount,
		}
	}

	meshGeometryInfo := graphics.MeshGeometryInfo{
		VertexBuffers: gog.Map(assetGeometry.VertexBuffers, func(buffer asset.VertexBuffer) graphics.MeshGeometryVertexBuffer {
			return graphics.MeshGeometryVertexBuffer{
				ByteStride: buffer.Stride,
				Data:       buffer.Data,
			}
		}),
		VertexFormat: s.resolveVertexFormat(assetGeometry.VertexLayout),
		IndexBuffer: graphics.MeshGeometryIndexBuffer{
			Data:   assetGeometry.IndexBuffer.Data,
			Format: s.resolveIndexFormat(assetGeometry.IndexBuffer.IndexLayout),
		},
		Fragments:            meshFragmentsInfo,
		BoundingSphereRadius: assetGeometry.BoundingSphereRadius,
		MinDistance:          opt.V(assetGeometry.MinDistance),
		MaxDistance:          opt.V(assetGeometry.MaxDistance),
	}

	promise := async.NewPromise[*graphics.MeshGeometry]()
	s.gfxWorker.Schedule(func() {
		gfxEngine := s.engine.Graphics()
		meshGeometry := gfxEngine.CreateMeshGeometry(meshGeometryInfo)
		promise.Deliver(meshGeometry)
	})
	return promise
}

func (s *ResourceSet) convertMeshDefinition(geometris []*graphics.MeshGeometry, materials []*graphics.Material, assetMeshDefinition asset.MeshDefinition) async.Promise[*graphics.MeshDefinition] {
	geometry := geometris[assetMeshDefinition.GeometryIndex]

	bindingMaterials := make([]*graphics.Material, geometry.FragmentCount())
	for _, assetBinding := range assetMeshDefinition.MaterialBindings {
		bindingMaterials[assetBinding.FragmentIndex] = materials[assetBinding.MaterialIndex]
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

func (s *ResourceSet) convertMeshInstance(assetMesh asset.Mesh) meshInstance {
	return meshInstance{
		NodeIndex:       int(assetMesh.NodeIndex),
		DefinitionIndex: int(assetMesh.MeshDefinitionIndex),
		ArmatureIndex:   int(assetMesh.ArmatureIndex),
	}
}
