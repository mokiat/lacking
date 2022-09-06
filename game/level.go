package game

import (
	"fmt"

	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/game/asset"
)

type SceneDefinition struct {
	skyboxTexture     *CubeTexture
	reflectionTexture *CubeTexture
	refractionTexture *CubeTexture
	model             *ModelDefinition
}

func (r *ResourceSet) allocateScene(resource asset.Resource) (*SceneDefinition, error) {
	levelAsset := new(asset.Scene)

	ioTask := func() error {
		return resource.ReadContent(levelAsset)
	}
	if err := r.ioWorker.Schedule(ioTask).Wait(); err != nil {
		return nil, fmt.Errorf("failed to read asset: %w", err)
	}

	var skyboxTexture *CubeTexture
	if texID := levelAsset.SkyboxTexture; texID != "" {
		var promise async.Promise[*CubeTexture]
		r.gfxWorker.ScheduleVoid(func() {
			promise = r.OpenCubeTexture(texID)
		}).Wait()
		texture, err := promise.Wait()
		if err != nil {
			return nil, fmt.Errorf("error loading skybox texture: %w", err)
		}
		skyboxTexture = texture
	}

	var reflectionTexture *CubeTexture
	if texID := levelAsset.AmbientReflectionTexture; texID != "" {
		var promise async.Promise[*CubeTexture]
		r.gfxWorker.ScheduleVoid(func() {
			promise = r.OpenCubeTexture(texID)
		}).Wait()
		texture, err := promise.Wait()
		if err != nil {
			return nil, fmt.Errorf("error loading reflection texture: %w", err)
		}
		reflectionTexture = texture
	}

	var refractionTexture *CubeTexture
	if texID := levelAsset.AmbientRefractionTexture; texID != "" {
		var promise async.Promise[*CubeTexture]
		r.gfxWorker.ScheduleVoid(func() {
			promise = r.OpenCubeTexture(texID)
		}).Wait()
		texture, err := promise.Wait()
		if err != nil {
			return nil, fmt.Errorf("error loading refraction texture: %w", err)
		}
		refractionTexture = texture
	}

	// TODO: Allocate the model as well

	return &SceneDefinition{
		skyboxTexture:     skyboxTexture,
		reflectionTexture: reflectionTexture,
		refractionTexture: refractionTexture,
	}, nil
}

func (r *ResourceSet) releaseScene(model *SceneDefinition) {
	// TODO
}
