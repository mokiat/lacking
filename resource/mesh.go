package resource

import (
	"fmt"

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
	IndexOffset   int
	IndexCount    int32
	AlbedoTexture *TwoDTexture
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
	for i := range mesh.SubMeshes {
		subMeshAsset := meshAsset.SubMeshes[i]
		subMesh := SubMesh{
			IndexOffset: int(subMeshAsset.IndexOffset),
			IndexCount:  int32(subMeshAsset.IndexCount),
		}
		if subMeshAsset.ColorTexture != "" {
			var albedoTexture *TwoDTexture
			if result := registry.LoadTwoDTexture(subMeshAsset.ColorTexture).OnSuccess(InjectTwoDTexture(&albedoTexture)).Wait(); result.Err != nil {
				return nil, result.Err
			}
			subMesh.AlbedoTexture = albedoTexture
		}
		mesh.SubMeshes[i] = subMesh
	}
	return mesh, nil
}

func ReleaseMesh(registry *Registry, gfxWorker *graphics.Worker, mesh *Mesh) error {
	for _, subMesh := range mesh.SubMeshes {
		if subMesh.AlbedoTexture != nil {
			if result := registry.UnloadTwoDTexture(subMesh.AlbedoTexture).Wait(); result.Err != nil {
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