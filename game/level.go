package game

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/util/async"
)

type SceneDefinition struct {
	modelDefinitions []*ModelDefinition
	modelInstances   []ModelInfo
}

func (r *ResourceSet) allocateScene(resource asset.Resource) (*SceneDefinition, error) {
	levelAsset := new(asset.Scene)

	ioTask := func() error {
		return resource.ReadContent(levelAsset)
	}
	if err := r.ioWorker.Schedule(ioTask).Wait(); err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}

	var (
		modelDefinitions = make([]*ModelDefinition, len(levelAsset.ModelDefinitions))
	)
	for i, definitionAsset := range levelAsset.ModelDefinitions {
		modelDef, err := r.allocateModel(&definitionAsset)
		if err != nil {
			return nil, fmt.Errorf("error allocating model: %w", err)
		}
		modelDefinitions[i] = modelDef
	}

	var (
		modelInstances = make([]ModelInfo, len(levelAsset.ModelInstances))
	)
	for i, instanceAsset := range levelAsset.ModelInstances {
		var modelDef *ModelDefinition
		if instanceAsset.ModelID == "" {
			modelDef = modelDefinitions[instanceAsset.ModelIndex]
		} else {
			var promise async.Promise[*ModelDefinition]
			r.gfxWorker.ScheduleVoid(func() {
				promise = r.OpenModel(instanceAsset.ModelID)
			}).Wait()
			var err error
			modelDef, err = promise.Wait()
			if err != nil {
				return nil, fmt.Errorf("error loading model %q: %w", instanceAsset.ModelID, err)
			}
		}
		modelInstances[i] = ModelInfo{
			Name:              instanceAsset.Name,
			Definition:        modelDef,
			Position:          instanceAsset.Translation,
			Rotation:          instanceAsset.Rotation,
			Scale:             instanceAsset.Scale,
			IsDynamic:         false, // TODO
			PrepareAnimations: true,  // TODO
		}
	}

	return &SceneDefinition{
		modelDefinitions: modelDefinitions,
		modelInstances:   modelInstances,
	}, nil
}

func (r *ResourceSet) releaseScene(scene *SceneDefinition) {
	for _, def := range scene.modelDefinitions {
		r.releaseModel(def)
	}
}
