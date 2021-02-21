package resource

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/graphics"
)

type Mesh struct {
	GFXVertexArray *graphics.VertexArray
	SubMeshes      []SubMesh
}

type SubMesh struct {
	Primitive   graphics.RenderPrimitive
	IndexOffset int
	IndexCount  int32
	Material    *Material
}

type Material struct {
	Shader           *Shader
	BackfaceCulling  bool
	Metalness        float32
	MetalnessTexture *TwoDTexture
	Roughness        float32
	RoughnessTexture *TwoDTexture
	AlbedoColor      sprec.Vec4
	AlbedoTexture    *TwoDTexture
	NormalScale      float32
	NormalTexture    *TwoDTexture
}

func AllocateMesh(set *Set, gfxWorker *async.Worker, meshAsset *asset.Mesh) (*Mesh, error) {
	mesh := &Mesh{
		GFXVertexArray: graphics.NewVertexArray(),
	}

	gfxTask := gfxWorker.Schedule(async.VoidTask(func() error {
		return mesh.GFXVertexArray.Allocate(graphics.VertexArrayData{
			VertexData: meshAsset.VertexData,
			Layout: graphics.VertexArrayLayout{
				HasCoord:       meshAsset.VertexLayout.CoordOffset != asset.UnspecifiedOffset,
				CoordOffset:    int(meshAsset.VertexLayout.CoordOffset),
				CoordStride:    int32(meshAsset.VertexLayout.CoordStride),
				HasNormal:      meshAsset.VertexLayout.NormalOffset != asset.UnspecifiedOffset,
				NormalOffset:   int(meshAsset.VertexLayout.NormalOffset),
				NormalStride:   int32(meshAsset.VertexLayout.NormalStride),
				HasTangent:     meshAsset.VertexLayout.TangentOffset != asset.UnspecifiedOffset,
				TangentOffset:  int(meshAsset.VertexLayout.TangentOffset),
				TangentStride:  int32(meshAsset.VertexLayout.TangentStride),
				HasTexCoord:    meshAsset.VertexLayout.TexCoordOffset != asset.UnspecifiedOffset,
				TexCoordOffset: int(meshAsset.VertexLayout.TexCoordOffset),
				TexCoordStride: int32(meshAsset.VertexLayout.TexCoordStride),
				HasColor:       meshAsset.VertexLayout.ColorOffset != asset.UnspecifiedOffset,
				ColorOffset:    int(meshAsset.VertexLayout.ColorOffset),
				ColorStride:    int32(meshAsset.VertexLayout.ColorStride),
			},
			IndexData: meshAsset.IndexData,
		})
	}))
	if err := gfxTask.Wait().Err; err != nil {
		return nil, fmt.Errorf("failed to allocate gfx vertex array: %w", err)
	}

	mesh.SubMeshes = make([]SubMesh, len(meshAsset.SubMeshes))
	for i, subMeshAsset := range meshAsset.SubMeshes {
		subMesh := SubMesh{
			Primitive:   assetToGraphicsPrimitive(subMeshAsset.Primitive),
			IndexOffset: int(subMeshAsset.IndexOffset),
			IndexCount:  int32(subMeshAsset.IndexCount),
			Material: &Material{
				BackfaceCulling: subMeshAsset.Material.BackfaceCulling,
				Metalness:       subMeshAsset.Material.Metalness,
				Roughness:       subMeshAsset.Material.Roughness,
				AlbedoColor: sprec.NewVec4(
					subMeshAsset.Material.Color[0],
					subMeshAsset.Material.Color[1],
					subMeshAsset.Material.Color[2],
					subMeshAsset.Material.Color[3],
				),
				NormalScale: subMeshAsset.Material.NormalScale,
			},
		}
		if subMeshAsset.Material.MetalnessTexture != "" {
			uri := fmt.Sprintf("assets/textures/twod/%s.dat", subMeshAsset.Material.MetalnessTexture)
			if err := set.OpenTwoDTexture(uri, &subMesh.Material.MetalnessTexture).Wait(); err != nil {
				return nil, fmt.Errorf("failed to load metalness texture: %w", err)
			}
		}
		if subMeshAsset.Material.RoughnessTexture != "" {
			uri := fmt.Sprintf("assets/textures/twod/%s.dat", subMeshAsset.Material.RoughnessTexture)
			if err := set.OpenTwoDTexture(uri, &subMesh.Material.RoughnessTexture).Wait(); err != nil {
				return nil, fmt.Errorf("failed to load roughness texture: %w", err)
			}
		}
		if subMeshAsset.Material.ColorTexture != "" {
			uri := fmt.Sprintf("assets/textures/twod/%s.dat", subMeshAsset.Material.ColorTexture)
			if err := set.OpenTwoDTexture(uri, &subMesh.Material.AlbedoTexture).Wait(); err != nil {
				return nil, fmt.Errorf("failed to load albedo texture: %w", err)
			}
		}
		if subMeshAsset.Material.NormalTexture != "" {
			uri := fmt.Sprintf("assets/textures/twod/%s.dat", subMeshAsset.Material.NormalTexture)
			if err := set.OpenTwoDTexture(uri, &subMesh.Material.NormalTexture).Wait(); err != nil {
				return nil, fmt.Errorf("failed to load normal texture: %w", err)
			}
		}
		shaderInfo := ShaderInfo{
			Type:                subMeshAsset.Material.Type,
			HasMetalnessTexture: subMeshAsset.Material.MetalnessTexture != "",
			HasRoughnessTexture: subMeshAsset.Material.RoughnessTexture != "",
			HasAlbedoTexture:    subMeshAsset.Material.ColorTexture != "",
			HasNormalTexture:    subMeshAsset.Material.NormalTexture != "",
		}
		if err := set.CreateShader(shaderInfo, &subMesh.Material.Shader).Wait(); err != nil {
			return nil, fmt.Errorf("failed to create material shader: %w", err)
		}
		mesh.SubMeshes[i] = subMesh
	}
	return mesh, nil
}

func ReleaseMesh(gfxWorker *async.Worker, mesh *Mesh) error {
	gfxTask := gfxWorker.Schedule(async.VoidTask(func() error {
		return mesh.GFXVertexArray.Release()
	}))
	if err := gfxTask.Wait().Err; err != nil {
		return fmt.Errorf("failed to release gfx vertex array: %w", err)
	}
	return nil
}

func assetToGraphicsPrimitive(primitive asset.Primitive) graphics.RenderPrimitive {
	switch primitive {
	case asset.PrimitivePoints:
		return graphics.RenderPrimitivePoints
	case asset.PrimitiveLines:
		return graphics.RenderPrimitiveLines
	case asset.PrimitiveLineStrip:
		return graphics.RenderPrimitiveLineStrip
	case asset.PrimitiveLineLoop:
		return graphics.RenderPrimitiveLineStrip
	case asset.PrimitiveTriangles:
		return graphics.RenderPrimitiveTriangles
	case asset.PrimitiveTriangleStrip:
		return graphics.RenderPrimitiveTriangleStrip
	case asset.PrimitiveTriangleFan:
		return graphics.RenderPrimitiveTriangleFan
	default:
		panic(fmt.Errorf("unsupported primitive: %d", primitive))
	}
}
