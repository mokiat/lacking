package pack

import (
	"fmt"

	"github.com/mokiat/gomath/stod"
	"github.com/mokiat/lacking/game/asset"
)

type SaveLevelAssetAction struct {
	resource      asset.Resource
	levelProvider LevelProvider
}

func (a *SaveLevelAssetAction) Describe() string {
	return fmt.Sprintf("save_level_asset(%q)", a.resource.Name())
}

func (a *SaveLevelAssetAction) Run() error {
	level := a.levelProvider.Level()

	modelInstances := make([]asset.ModelInstance, len(level.StaticEntities))
	for i, staticEntity := range level.StaticEntities {
		t, r, s := stod.Mat4(staticEntity.Matrix).TRS()
		modelInstances[i] = asset.ModelInstance{
			ModelID:     staticEntity.Model,
			Translation: t,
			Rotation:    r,
			Scale:       s,
		}
	}

	levelAsset := &asset.Scene{
		ModelInstances: modelInstances,
	}
	if err := a.resource.WriteContent(levelAsset); err != nil {
		return fmt.Errorf("failed to write asset: %w", err)
	}
	return nil
}
