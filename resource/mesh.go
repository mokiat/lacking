package resource

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/graphics"
)

const MeshTypeName = TypeName("mesh")

func InjectMesh(target **Mesh) func(value interface{}) {
	return func(value interface{}) {
		*target = value.(*Mesh)
	}
}

type Mesh struct {
	Name           string
	GFXVertexArray *graphics.VertexArray
	SubMeshes      []SubMesh
}

type SubMesh struct {
	Primitive        graphics.RenderPrimitive
	IndexOffset      int
	IndexCount       int32
	BackfaceCulling  bool
	Program          *Program
	Metalness        float32
	MetalnessTexture *TwoDTexture
	Roughness        float32
	RoughnessTexture *TwoDTexture
	AlbedoColor      sprec.Vec4
	AlbedoTexture    *TwoDTexture
	NormalScale      float32
	NormalTexture    *TwoDTexture
}

func NewMeshOperator(locator Locator, gfxWorker *graphics.Worker) *MeshOperator {
	return &MeshOperator{
		locator:   locator,
		gfxWorker: gfxWorker,
	}
}

type MeshOperator struct {
	locator   Locator
	gfxWorker *graphics.Worker
}

func (o *MeshOperator) Allocate(registry *Registry, name string) (interface{}, error) {
	in, err := o.locator.Open("assets", "meshes", name)
	if err != nil {
		return nil, fmt.Errorf("failed to open mesh asset %q: %w", name, err)
	}
	defer in.Close()

	meshAsset := new(asset.Mesh)
	if err := asset.DecodeMesh(in, meshAsset); err != nil {
		return nil, fmt.Errorf("failed to decode mesh asset %q: %w", name, err)
	}

	return AllocateMesh(registry, name, o.gfxWorker, meshAsset)
}

func (o *MeshOperator) Release(registry *Registry, resource interface{}) error {
	mesh := resource.(*Mesh)
	return ReleaseMesh(registry, o.gfxWorker, mesh)
}

func AllocateMesh(registry *Registry, name string, gfxWorker *graphics.Worker, meshAsset *asset.Mesh) (*Mesh, error) {
	mesh := &Mesh{
		Name:           name,
		GFXVertexArray: &graphics.VertexArray{},
	}

	gfxTask := gfxWorker.Schedule(func() error {
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
	})
	if err := gfxTask.Wait(); err != nil {
		return nil, fmt.Errorf("failed to allocate gfx vertex array: %w", err)
	}

	mesh.SubMeshes = make([]SubMesh, len(meshAsset.SubMeshes))
	for i, subMeshAsset := range meshAsset.SubMeshes {
		subMesh := SubMesh{
			Primitive:       assetToGraphicsPrimitive(subMeshAsset.Primitive),
			IndexOffset:     int(subMeshAsset.IndexOffset),
			IndexCount:      int32(subMeshAsset.IndexCount),
			BackfaceCulling: subMeshAsset.BackfaceCulling,
			Metalness:       subMeshAsset.Metalness,
			Roughness:       subMeshAsset.Roughness,
			AlbedoColor: sprec.NewVec4(
				subMeshAsset.Color[0],
				subMeshAsset.Color[1],
				subMeshAsset.Color[2],
				subMeshAsset.Color[3],
			),
			NormalScale: subMeshAsset.NormalScale,
		}
		if subMeshAsset.MetalnessTexture != "" {
			result := registry.LoadTwoDTexture(subMeshAsset.MetalnessTexture).
				OnSuccess(InjectTwoDTexture(&subMesh.MetalnessTexture)).
				Wait()
			if err := result.Err; err != nil {
				return nil, fmt.Errorf("failed to load metalness texture: %w", err)
			}
		}
		if subMeshAsset.RoughnessTexture != "" {
			result := registry.LoadTwoDTexture(subMeshAsset.RoughnessTexture).
				OnSuccess(InjectTwoDTexture(&subMesh.RoughnessTexture)).
				Wait()
			if err := result.Err; err != nil {
				return nil, fmt.Errorf("failed to load roughness texture: %w", err)
			}
		}
		if subMeshAsset.ColorTexture != "" {
			result := registry.LoadTwoDTexture(subMeshAsset.ColorTexture).
				OnSuccess(InjectTwoDTexture(&subMesh.AlbedoTexture)).
				Wait()
			if err := result.Err; err != nil {
				return nil, fmt.Errorf("failed to load albedo texture: %w", err)
			}
		}
		if subMeshAsset.NormalTexture != "" {
			result := registry.LoadTwoDTexture(subMeshAsset.NormalTexture).
				OnSuccess(InjectTwoDTexture(&subMesh.NormalTexture)).
				Wait()
			if err := result.Err; err != nil {
				return nil, fmt.Errorf("failed to load normal texture: %w", err)
			}
		}
		mesh.SubMeshes[i] = subMesh
	}
	return mesh, nil
}

func ReleaseMesh(registry *Registry, gfxWorker *graphics.Worker, mesh *Mesh) error {
	for _, subMesh := range mesh.SubMeshes {
		if subMesh.MetalnessTexture != nil {
			if result := registry.UnloadTwoDTexture(subMesh.MetalnessTexture).Wait(); result.Err != nil {
				return result.Err
			}
		}
		if subMesh.RoughnessTexture != nil {
			if result := registry.UnloadTwoDTexture(subMesh.RoughnessTexture).Wait(); result.Err != nil {
				return result.Err
			}
		}
		if subMesh.AlbedoTexture != nil {
			if result := registry.UnloadTwoDTexture(subMesh.AlbedoTexture).Wait(); result.Err != nil {
				return result.Err
			}
		}
		if subMesh.NormalTexture != nil {
			if result := registry.UnloadTwoDTexture(subMesh.NormalTexture).Wait(); result.Err != nil {
				return result.Err
			}
		}
	}

	gfxTask := gfxWorker.Schedule(func() error {
		return mesh.GFXVertexArray.Release()
	})
	if err := gfxTask.Wait(); err != nil {
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
