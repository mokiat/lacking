package resource

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/log"
)

type Mesh struct {
	Name string

	GFXMeshTemplate *graphics.MeshTemplate
}

func AllocateMesh(registry *Registry, gfxEngine *graphics.Engine, materials []*Material, meshAsset *asset.MeshDefinition) (*Mesh, error) {
	mesh := &Mesh{
		Name: meshAsset.Name,
	}

	subMeshDefinitions := make([]graphics.SubMeshTemplateDefinition, 0)
	for _, assetFragment := range meshAsset.Fragments {
		if matIndex := assetFragment.MaterialIndex; matIndex != asset.UnspecifiedMaterialIndex {
			subMeshDefinitions = append(subMeshDefinitions, graphics.SubMeshTemplateDefinition{
				Primitive:   assetToGraphicsPrimitive(assetFragment.Topology),
				IndexOffset: int(assetFragment.IndexOffset),
				IndexCount:  int(assetFragment.IndexCount),
				Material:    materials[matIndex].GFXMaterial,
			})
		} else {
			log.Warn("[resource] mesh fragment does not reference material")
		}
	}

	registry.ScheduleVoid(func() {
		definition := graphics.MeshTemplateDefinition{
			VertexData: meshAsset.VertexData,
			VertexFormat: graphics.VertexFormat{
				HasCoord:            meshAsset.VertexLayout.CoordOffset != asset.UnspecifiedOffset,
				CoordOffsetBytes:    int(meshAsset.VertexLayout.CoordOffset),
				CoordStrideBytes:    int(meshAsset.VertexLayout.CoordStride),
				HasNormal:           meshAsset.VertexLayout.NormalOffset != asset.UnspecifiedOffset,
				NormalOffsetBytes:   int(meshAsset.VertexLayout.NormalOffset),
				NormalStrideBytes:   int(meshAsset.VertexLayout.NormalStride),
				HasTangent:          meshAsset.VertexLayout.TangentOffset != asset.UnspecifiedOffset,
				TangentOffsetBytes:  int(meshAsset.VertexLayout.TangentOffset),
				TangentStrideBytes:  int(meshAsset.VertexLayout.TangentStride),
				HasTexCoord:         meshAsset.VertexLayout.TexCoordOffset != asset.UnspecifiedOffset,
				TexCoordOffsetBytes: int(meshAsset.VertexLayout.TexCoordOffset),
				TexCoordStrideBytes: int(meshAsset.VertexLayout.TexCoordStride),
				HasColor:            meshAsset.VertexLayout.ColorOffset != asset.UnspecifiedOffset,
				ColorOffsetBytes:    int(meshAsset.VertexLayout.ColorOffset),
				ColorStrideBytes:    int(meshAsset.VertexLayout.ColorStride),
				HasWeights:          meshAsset.VertexLayout.WeightsOffset != asset.UnspecifiedOffset,
				WeightsOffsetBytes:  int(meshAsset.VertexLayout.WeightsOffset),
				WeightsStrideBytes:  int(meshAsset.VertexLayout.WeightsStride),
				HasJoints:           meshAsset.VertexLayout.JointsOffset != asset.UnspecifiedOffset,
				JointsOffsetBytes:   int(meshAsset.VertexLayout.JointsOffset),
				JointsStrideBytes:   int(meshAsset.VertexLayout.JointsStride),
			},
			IndexData:   meshAsset.IndexData,
			IndexFormat: assetToGraphicsIndexFormat(meshAsset.IndexLayout),
			SubMeshes:   subMeshDefinitions,
		}
		mesh.GFXMeshTemplate = gfxEngine.CreateMeshTemplate(definition)
	}).Wait()

	return mesh, nil
}

func ReleaseMesh(registry *Registry, mesh *Mesh) error {
	registry.ScheduleVoid(func() {
		mesh.GFXMeshTemplate.Delete()
	}).Wait()

	mesh.GFXMeshTemplate = nil
	return nil
}

func assetToGraphicsIndexFormat(layout asset.IndexLayout) graphics.IndexFormat {
	switch layout {
	case asset.IndexLayoutUint16:
		return graphics.IndexFormatU16
	case asset.IndexLayoutUint32:
		return graphics.IndexFormatU32
	default:
		panic(fmt.Errorf("unsupported index layout: %d", layout))
	}
}

func assetToGraphicsPrimitive(primitive asset.MeshTopology) graphics.Primitive {
	switch primitive {
	case asset.MeshTopologyPoints:
		return graphics.PrimitivePoints
	case asset.MeshTopologyLines:
		return graphics.PrimitiveLines
	case asset.MeshTopologyLineStrip:
		return graphics.PrimitiveLineStrip
	case asset.MeshTopologyLineLoop:
		return graphics.PrimitiveLineLoop
	case asset.MeshTopologyTriangles:
		return graphics.PrimitiveTriangles
	case asset.MeshTopologyTriangleStrip:
		return graphics.PrimitiveTriangleStrip
	case asset.MeshTopologyTriangleFan:
		return graphics.PrimitiveTriangleFan
	default:
		panic(fmt.Errorf("unsupported primitive: %d", primitive))
	}
}
