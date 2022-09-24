package pack

import (
	"fmt"
	"log"
	"time"

	"github.com/mokiat/lacking/game/asset"
)

type Action interface {
	Run() error
}

type Described interface {
	Describe() string
}

func newPipeline(id int, registry asset.Registry, resourceLocator ResourceLocator) *Pipeline {
	return &Pipeline{
		id:              id,
		registry:        registry,
		resourceLocator: resourceLocator,
	}
}

type Pipeline struct {
	id              int
	registry        asset.Registry
	resourceLocator ResourceLocator
	actions         []Action
}

func (p *Pipeline) OpenImageResource(uri string) *OpenImageResourceAction {
	action := &OpenImageResourceAction{
		locator: p.resourceLocator,
		uri:     uri,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) SaveTwoDTextureAsset(resource asset.Resource, image ImageProvider) *SaveTwoDTextureAssetAction {
	action := &SaveTwoDTextureAssetAction{
		resource:      resource,
		imageProvider: image,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) SaveCubeTextureAsset(resource asset.Resource, image CubeImageProvider, opts ...SaveCubeTextureOption) *SaveCubeTextureAction {
	action := &SaveCubeTextureAction{
		resource:      resource,
		imageProvider: image,
		format:        asset.TexelFormatRGBA8,
	}
	for _, opt := range opts {
		opt(action)
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) BuildCubeSideFromEquirectangular(side CubeSide, image ImageProvider) *BuildCubeSideFromEquirectangularAction {
	action := &BuildCubeSideFromEquirectangularAction{
		side:          side,
		imageProvider: image,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) BuildCubeImage(opts ...BuildCubeImageOption) *BuildCubeImageAction {
	action := &BuildCubeImageAction{}
	for _, opt := range opts {
		opt(action)
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) ScaleCubeImage(image CubeImageProvider, dimension int) *ScaleCubeImageAction {
	action := &ScaleCubeImageAction{
		imageProvider: image,
		dimension:     dimension,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) BuildIrradianceCubeImage(image CubeImageProvider, opts ...BuildIrradianceCubeImageOption) *BuildIrradianceCubeImageAction {
	action := &BuildIrradianceCubeImageAction{
		imageProvider: image,
		sampleCount:   10,
	}
	for _, opt := range opts {
		opt(action)
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) OpenGLTFResource(uri string) *OpenGLTFResourceAction {
	action := &OpenGLTFResourceAction{
		locator: p.resourceLocator,
		uri:     uri,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) SaveModelAsset(resource asset.Resource, model ModelProvider, opts ...SaveModelAssetOption) *SaveModelAssetAction {
	action := &SaveModelAssetAction{
		resource:      resource,
		modelProvider: model,
	}
	for _, opt := range opts {
		opt(action)
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) OpenLevelResource(uri string) *OpenLevelResourceAction {
	action := &OpenLevelResourceAction{
		locator: p.resourceLocator,
		uri:     uri,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) SaveLevelAsset(resource asset.Resource, level LevelProvider) *SaveLevelAssetAction {
	action := &SaveLevelAssetAction{
		resource:      resource,
		levelProvider: level,
	}
	p.scheduleAction(action)
	return action
}

func (p *Pipeline) scheduleAction(action Action) {
	p.actions = append(p.actions, action)
}

func (p *Pipeline) execute() error {
	for _, action := range p.actions {
		described, isDescribed := action.(Described)

		startTime := time.Now()
		if err := action.Run(); err != nil {
			if isDescribed {
				return fmt.Errorf("action %q failed: %w", described.Describe(), err)
			}
			return fmt.Errorf("an action failed: %w", err)
		}
		elapsedTime := time.Since(startTime)

		if isDescribed {
			log.Printf("pipeline %d, action %s, complete in %s", p.id, described.Describe(), elapsedTime)
		}
	}
	return nil
}
