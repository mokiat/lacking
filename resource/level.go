package resource

import (
	"fmt"

	"github.com/mokiat/lacking/data/asset"
	gameasset "github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
)

const LevelTypeName = TypeName("level")

func InjectLevel(target **Level) func(value interface{}) {
	return func(value interface{}) {
		*target = value.(*Level)
	}
}

type Level struct {
	Name                     string
	SkyboxTexture            *CubeTexture
	AmbientReflectionTexture *CubeTexture
	AmbientRefractionTexture *CubeTexture
}

func NewLevelOperator(delegate gameasset.Registry, gfxEngine *graphics.Engine) *LevelOperator {
	return &LevelOperator{
		delegate:  delegate,
		gfxEngine: gfxEngine,
	}
}

type LevelOperator struct {
	delegate  gameasset.Registry
	gfxEngine *graphics.Engine
}

func (o *LevelOperator) Allocate(registry *Registry, id string) (interface{}, error) {
	levelAsset := new(asset.Level)
	resource := o.delegate.ResourceByID(id)
	if resource == nil {
		return nil, fmt.Errorf("cannot find asset %q", id)
	}
	if err := resource.ReadContent(levelAsset); err != nil {
		return nil, fmt.Errorf("failed to open level asset %q: %w", id, err)
	}

	level := &Level{
		Name: id,
	}

	if result := registry.LoadCubeTexture(levelAsset.SkyboxTexture).OnSuccess(InjectCubeTexture(&level.SkyboxTexture)).Wait(); result.Err != nil {
		return nil, result.Err
	}
	if result := registry.LoadCubeTexture(levelAsset.AmbientReflectionTexture).OnSuccess(InjectCubeTexture(&level.AmbientReflectionTexture)).Wait(); result.Err != nil {
		return nil, result.Err
	}
	if result := registry.LoadCubeTexture(levelAsset.AmbientRefractionTexture).OnSuccess(InjectCubeTexture(&level.AmbientRefractionTexture)).Wait(); result.Err != nil {
		return nil, result.Err
	}

	return level, nil
}

func (o *LevelOperator) Release(registry *Registry, res interface{}) error {
	level := res.(*Level)

	if result := registry.UnloadCubeTexture(level.SkyboxTexture).Wait(); result.Err != nil {
		return result.Err
	}
	if result := registry.UnloadCubeTexture(level.AmbientReflectionTexture).Wait(); result.Err != nil {
		return result.Err
	}
	if result := registry.UnloadCubeTexture(level.AmbientRefractionTexture).Wait(); result.Err != nil {
		return result.Err
	}

	return nil
}
